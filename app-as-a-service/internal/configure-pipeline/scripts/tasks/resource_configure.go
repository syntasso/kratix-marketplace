package tasks

import (
	"fmt"

	kratix "github.com/syntasso/kratix-go"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"

	"github.com/syntasso/kratix-marketplace/app/util"
)

func ResourceConfigure(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing ResourceConfigure...")
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	image := util.MustString(util.Get(res, "spec.image"))
	name := util.MustString(util.Get(res, "spec.name"))
	namespace := util.MustString(util.Get(res, "metadata.namespace"))
	resourceName := util.MustString(util.Get(res, "metadata.name"))
	dbDriver := util.MustStringOrEmpty(util.Get(res, "spec.dbDriver"))
	servicePort := util.MustString(util.Get(res, "spec.service.port"))
	fmt.Println("resource-configure inputs:", "image="+image, "name="+name, "namespace="+namespace, "servicePort="+servicePort)

	deployYAML := util.RunKubectl(
		"create", "deployment", name,
		"--namespace="+namespace,
		"--replicas=1",
		"--image="+image,
		"--dry-run=client", "-o", "yaml",
	)

	if dbDriver == "postgresql" {
		var deploy appsv1.Deployment
		if err := yaml.Unmarshal(deployYAML, &deploy); err != nil {
			return fmt.Errorf("parse generated deployment: %w", err)
		}
		if deploy.APIVersion == "" {
			deploy.APIVersion = "apps/v1"
		}
		if deploy.Kind == "" {
			deploy.Kind = "Deployment"
		}

		pgIdentity := derivePostgresIdentity(resourceName)
		if err := applyNonVaultDatabaseWiring(&deploy, pgIdentity, namespace); err != nil {
			return fmt.Errorf("wire non-vault database in deployment: %w", err)
		}

		patchedDeployYAML, err := yaml.Marshal(&deploy)
		if err != nil {
			return fmt.Errorf("marshal generated deployment: %w", err)
		}
		deployYAML = patchedDeployYAML
	}

	if err := sdk.WriteOutput("deployment.yaml", deployYAML); err != nil {
		return fmt.Errorf("write deployment: %w", err)
	}

	svcYAML := util.RunKubectl(
		"create", "service", "nodeport", name,
		"--namespace="+namespace,
		"--tcp="+servicePort+":8080",
		"--dry-run=client", "-o", "yaml",
	)
	if err := sdk.WriteOutput("service.yaml", svcYAML); err != nil {
		return fmt.Errorf("write service: %w", err)
	}

	rule := fmt.Sprintf("%s.local.gd/*=%s:%s", name, name, servicePort)
	ingYAML := util.RunKubectl(
		"create", "ingress", name,
		"--namespace="+namespace,
		"--class=nginx",
		"--rule="+rule,
		"--dry-run=client", "-o", "yaml",
	)
	if err := sdk.WriteOutput("ingress.yaml", ingYAML); err != nil {
		return fmt.Errorf("write ingress: %w", err)
	}

	endpoint := fmt.Sprintf("http://%s.local.gd:31338", name)
	_ = st.Set("message", "deployed to "+endpoint)
	_ = st.Set("endpoint", endpoint)
	_ = st.Set("replicas", int64(1))
	fmt.Println("Finished executing ResourceConfigure.")
	return sdk.WriteStatus(st)
}
