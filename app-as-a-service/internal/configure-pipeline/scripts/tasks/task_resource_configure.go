package tasks

import (
	"fmt"

	kratix "github.com/syntasso/kratix-go"
	"github.com/syntasso/kratix-marketplace/app/util"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

func ConfigureResource(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing configureResource...")

	ctx, err := util.ReadResourceConfigureContext(sdk)
	if err != nil {
		return err
	}

	fmt.Println(
		"resource-configure inputs:",
		"image="+ctx.Image,
		"name="+ctx.Name,
		"namespace="+ctx.Namespace,
		"servicePort="+ctx.ServicePort,
	)

	deployYAML := util.RunKubectl(
		"create", "deployment", ctx.Name,
		"--namespace="+ctx.Namespace,
		"--replicas=1",
		"--image="+ctx.Image,
		"--dry-run=client", "-o", "yaml",
	)

	if ctx.DBDriver == "postgresql" {
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

		pgIdentity := util.DerivePostgresIdentity(ctx.ResourceName)
		if err := util.ApplyNonVaultDatabaseWiring(&deploy, pgIdentity, ctx.Namespace); err != nil {
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
		"create", "service", "nodeport", ctx.Name,
		"--namespace="+ctx.Namespace,
		"--tcp="+ctx.ServicePort+":8080",
		"--dry-run=client", "-o", "yaml",
	)
	if err := sdk.WriteOutput("service.yaml", svcYAML); err != nil {
		return fmt.Errorf("write service: %w", err)
	}

	rule := fmt.Sprintf("%s.local.gd/*=%s:%s", ctx.Name, ctx.Name, ctx.ServicePort)
	ingYAML := util.RunKubectl(
		"create", "ingress", ctx.Name,
		"--namespace="+ctx.Namespace,
		"--class=nginx",
		"--rule="+rule,
		"--dry-run=client", "-o", "yaml",
	)
	if err := sdk.WriteOutput("ingress.yaml", ingYAML); err != nil {
		return fmt.Errorf("write ingress: %w", err)
	}

	endpoint := fmt.Sprintf("http://%s.local.gd:31338", ctx.Name)
	_ = st.Set("message", "deployed to "+endpoint)
	_ = st.Set("endpoint", endpoint)
	_ = st.Set("replicas", int64(1))

	fmt.Println("Finished executing configureResource.")
	return sdk.WriteStatus(st)
}
