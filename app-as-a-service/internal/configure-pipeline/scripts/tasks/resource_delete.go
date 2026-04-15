package tasks

import (
	"fmt"
	"os"
	"strings"

	kratix "github.com/syntasso/kratix-go"

	"github.com/syntasso/kratix-marketplace/app/util"
)

func ResourceDelete(sdk *kratix.KratixSDK) error {
	fmt.Println("Executing ResourceDelete...")
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	dbDriver := util.MustStringOrEmpty(util.Get(res, "spec.dbDriver"))
	resourceName := util.MustString(util.Get(res, "metadata.name"))
	appName := util.MustString(util.Get(res, "spec.name"))
	namespace := util.MustString(util.Get(res, "metadata.namespace"))
	vaultEnabled := util.HasLabelTrue(res, vaultOptInLabelKey)
	fmt.Println("resource-delete inputs:", "dbDriver="+dbDriver, "name="+resourceName, "namespace="+namespace, fmt.Sprintf("%s=%t", vaultOptInLabelKey, vaultEnabled))

	if dbDriver == "" || dbDriver == "none" {
		fmt.Println("resource-delete skipped: no database requested")
		return nil
	}
	if dbDriver != "postgresql" {
		return fmt.Errorf("unsupported db driver %q. supported: postgresql", dbDriver)
	}

	pgIdentity := derivePostgresIdentity(resourceName)
	if _, err := util.RunKubectlOutput(
		"delete",
		"postgresql/"+pgIdentity.requestName,
		"--namespace="+namespace,
		"--ignore-not-found=true",
		"--wait=false",
	); err != nil {
		return fmt.Errorf("delete postgresql/%s: %w", pgIdentity.requestName, err)
	}

	if !vaultEnabled {
		fmt.Println("resource-delete completed: database request removed (vault was not enabled)")
		return nil
	}

	vaultAddr := strings.TrimSpace(os.Getenv("VAULT_ADDR"))
	vaultToken := strings.TrimSpace(os.Getenv("VAULT_TOKEN"))
	if vaultAddr == "" || vaultToken == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN are required for delete cleanup")
	}

	vaultNames := deriveVaultArtifacts(namespace, appName, pgIdentity)

	revokePayload := fmt.Sprintf(`{"prefix":"database/creds/%s"}`, vaultNames.dbRoleName)
	if err := util.VaultRevokePrefixIgnoreMissing(vaultAddr, vaultToken, revokePayload); err != nil {
		return err
	}

	paths := []string{
		fmt.Sprintf("database/roles/%s", vaultNames.dbRoleName),
		fmt.Sprintf("database/config/%s", vaultNames.dbConfigName),
		fmt.Sprintf("auth/%s/role/%s", vaultNames.authPath, vaultNames.authRoleName),
		fmt.Sprintf("sys/policies/acl/%s", vaultNames.policyName),
	}

	for _, path := range paths {
		if err := util.VaultDeleteIgnoreMissing(vaultAddr, vaultToken, path); err != nil {
			return err
		}
	}

	fmt.Println("Finished executing ResourceDelete.")
	return nil
}
