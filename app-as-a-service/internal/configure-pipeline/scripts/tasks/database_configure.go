package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kratix "github.com/syntasso/kratix-go"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/syntasso/kratix-marketplace/app/util"
)

func DatabaseConfigure(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing DatabaseConfigure...")
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	dbDriver := util.MustStringOrEmpty(util.Get(res, "spec.dbDriver"))
	name := util.MustString(util.Get(res, "metadata.name"))
	namespace := util.MustString(util.Get(res, "metadata.namespace"))
	vaultEnabled := util.HasLabelTrue(res, vaultOptInLabelKey)
	fmt.Println("database-configure inputs:", "dbDriver="+dbDriver, "name="+name, "namespace="+namespace)

	if dbDriver == "" || dbDriver == "none" {
		emptyStatus := kratix.NewStatus()
		_ = emptyStatus.Set("database", nil)
		return sdk.WriteStatus(emptyStatus)
	}
	if dbDriver != "postgresql" {
		return fmt.Errorf("unsupported db driver %q. supported: postgresql", dbDriver)
	}

	pgIdentity := derivePostgresIdentity(name)

	deploy, err := util.ReadDeployment("/kratix/output/deployment.yaml")
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read existing deployment: %w", err)
		}
	} else {
		if err := applyNonVaultDatabaseWiring(&deploy, pgIdentity, namespace); err != nil {
			return fmt.Errorf("wire non-vault database in deployment: %w", err)
		}
		if err := util.WriteYAMLObject(sdk, "deployment.yaml", &deploy); err != nil {
			return fmt.Errorf("write updated deployment: %w", err)
		}
	}

	if err := os.MkdirAll(filepath.Clean("/kratix/output/platform"), 0o755); err != nil {
		return fmt.Errorf("mkdir platform: %w", err)
	}

	pg := &unstructured.Unstructured{Object: map[string]any{}}
	pg.SetAPIVersion("marketplace.kratix.io/v1alpha2")
	pg.SetKind("postgresql")
	pg.SetName(pgIdentity.requestName)
	pg.SetNamespace(namespace)

	resourceUID := strings.TrimSpace(util.MustString(util.Get(res, "metadata.uid")))
	if resourceUID != "" {
		_ = unstructured.SetNestedSlice(pg.Object, []any{map[string]any{
			"apiVersion":         "marketplace.kratix.io/v1",
			"kind":               "app",
			"name":               name,
			"uid":                resourceUID,
			"blockOwnerDeletion": false,
			"controller":         false,
		}}, "metadata", "ownerReferences")
	}

	_ = unstructured.SetNestedField(pg.Object, namespace, "spec", "namespace")
	_ = unstructured.SetNestedField(pg.Object, pgIdentity.teamID, "spec", "teamId")
	_ = unstructured.SetNestedField(pg.Object, pgIdentity.dbName, "spec", "dbName")

	if err := util.WriteYAMLMap(sdk, "platform/postgresql.yaml", pg.Object); err != nil {
		return fmt.Errorf("write postgresql request: %w", err)
	}

	ds := []kratix.DestinationSelector{{Directory: "platform", MatchLabels: map[string]string{"environment": "platform"}}}
	if err := util.WriteDestinationSelectors(ds); err != nil {
		return fmt.Errorf("write destination selectors: %w", err)
	}

	_ = st.Set("database.teamId", pgIdentity.teamID)
	_ = st.Set("database.dbName", pgIdentity.dbName)
	_ = st.Set("database.requestName", pgIdentity.requestName)
	_ = st.Set("database.instanceName", pgIdentity.instanceName)
	_ = st.Set("database.credentialsSecret", pgIdentity.credentialsSecretName)
	_ = st.Set("database.vaultEnabled", vaultEnabled)
	fmt.Println("Finished executing DatabaseConfigure.")
	return sdk.WriteStatus(st)
}
