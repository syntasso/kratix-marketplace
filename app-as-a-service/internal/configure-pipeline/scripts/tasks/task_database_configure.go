package tasks

import (
	"fmt"
	"os"
	"path/filepath"

	kratix "github.com/syntasso/kratix-go"
	"github.com/syntasso/kratix-marketplace/app/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ConfigureDatabase(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing configureDatabase...")

	ctx, err := util.ReadAppRequestContext(sdk)
	if err != nil {
		return err
	}

	fmt.Println("database-configure inputs:", "dbDriver="+ctx.DBDriver, "name="+ctx.Name, "namespace="+ctx.Namespace)

	if ctx.DBDriver == "" || ctx.DBDriver == "none" {
		st := kratix.NewStatus()
		_ = st.Set("database", nil) // -> database: null
		return sdk.WriteStatus(st)
	}
	if ctx.DBDriver != "postgresql" {
		return fmt.Errorf("unsupported db driver %q. supported: postgresql", ctx.DBDriver)
	}

	pgIdentity := util.DerivePostgresIdentity(ctx.Name)

	// Update deployment with non-vault database wiring via operator-managed
	// secret if a deployment manifest already exists in this pipeline run.
	deploy, err := util.ReadDeployment("/kratix/output/deployment.yaml")
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read existing deployment: %w", err)
		}
	} else {
		if err := util.ApplyNonVaultDatabaseWiring(&deploy, pgIdentity, ctx.Namespace); err != nil {
			return fmt.Errorf("wire non-vault database in deployment: %w", err)
		}
		if err := util.WriteYAMLObject(sdk, "deployment.yaml", &deploy); err != nil {
			return fmt.Errorf("write updated deployment: %w", err)
		}
	}

	// Write base postgresql CR. Vault-specific fields are added in
	// vault-configure.
	if err := os.MkdirAll(filepath.Clean("/kratix/output/platform"), 0o755); err != nil {
		return fmt.Errorf("mkdir platform: %w", err)
	}

	pg := &unstructured.Unstructured{Object: map[string]any{}}
	pg.SetAPIVersion("marketplace.kratix.io/v1alpha2")
	pg.SetKind("postgresql")
	pg.SetName(pgIdentity.RequestName)
	pg.SetNamespace(ctx.Namespace)
	_ = unstructured.SetNestedField(pg.Object, ctx.Namespace, "spec", "namespace")
	_ = unstructured.SetNestedField(pg.Object, pgIdentity.TeamID, "spec", "teamId")
	_ = unstructured.SetNestedField(pg.Object, pgIdentity.DBName, "spec", "dbName")

	if ctx.VaultEnabled {
		vaultAuthPath := "kubernetes"
		appServiceAccount := util.AppendKubeNameSuffix(ctx.Name, "-app-sa", 63)
		vaultAuthRole, vaultDBRole := util.DeriveVaultResourceNames(ctx.Namespace, pgIdentity.TeamID, pgIdentity.DBName)
		vaultCredentialsPath := fmt.Sprintf("database/creds/%s", vaultDBRole)

		_ = unstructured.SetNestedField(pg.Object, true, "spec", "vault", "enabled")
		_ = unstructured.SetNestedField(pg.Object, appServiceAccount, "spec", "vault", "appServiceAccount")
		_ = unstructured.SetNestedField(pg.Object, ctx.Namespace, "spec", "vault", "appServiceAccountNamespace")
		_ = unstructured.SetNestedField(pg.Object, vaultAuthPath, "spec", "vault", "kubernetesAuthPath")

		_ = st.Set("database.vaultEnabled", true)
		_ = st.Set("database.serviceAccount", appServiceAccount)
		_ = st.Set("database.vaultAuthPath", fmt.Sprintf("auth/%s/login", vaultAuthPath))
		_ = st.Set("database.vaultRole", vaultAuthRole)
		_ = st.Set("database.vaultCredentialsPath", vaultCredentialsPath)
		_ = st.Set("database.credentialsFile", util.VaultCredentialsFilePath)
	} else {
		_ = st.Set("database.vaultEnabled", false)
		_ = st.Set("database.serviceAccount", nil)
		_ = st.Set("database.vaultAuthPath", nil)
		_ = st.Set("database.vaultRole", nil)
		_ = st.Set("database.vaultCredentialsPath", nil)
		_ = st.Set("database.credentialsFile", nil)
	}

	if err := util.WriteYAMLMap(sdk, "platform/postgresql.yaml", pg.Object); err != nil {
		return fmt.Errorf("write postgresql request: %w", err)
	}

	ds := []kratix.DestinationSelector{
		{Directory: "platform", MatchLabels: map[string]string{"environment": "platform"}},
	}
	if err := util.WriteDestinationSelectors(ds); err != nil {
		return fmt.Errorf("write destination selectors: %w", err)
	}

	_ = st.Set("database.teamId", pgIdentity.TeamID)
	_ = st.Set("database.dbName", pgIdentity.DBName)
	_ = st.Set("database.requestName", pgIdentity.RequestName)
	_ = st.Set("database.instanceName", pgIdentity.InstanceName)
	_ = st.Set("database.credentialsSecret", pgIdentity.CredentialsSecretName)

	fmt.Println("Finished executing configureDatabase.")
	return sdk.WriteStatus(st)
}
