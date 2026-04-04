package tasks

import (
	"fmt"

	kratix "github.com/syntasso/kratix-go"
	"github.com/syntasso/kratix-marketplace/app/util"
	corev1 "k8s.io/api/core/v1"
)

func ConfigureVault(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing configureVault...")

	ctx, err := util.ReadAppRequestContext(sdk)
	if err != nil {
		return err
	}

	fmt.Println(
		"vault-configure inputs:",
		"dbDriver="+ctx.DBDriver,
		"name="+ctx.Name,
		"namespace="+ctx.Namespace,
		fmt.Sprintf("%s=%t", util.VaultOptInLabelKey, ctx.VaultEnabled),
	)

	if ctx.DBDriver == "" || ctx.DBDriver == "none" {
		fmt.Println("vault-configure skipped: no database requested")
		return nil
	}
	if ctx.DBDriver != "postgresql" {
		return fmt.Errorf("unsupported db driver %q. supported: postgresql", ctx.DBDriver)
	}
	if !ctx.VaultEnabled {
		fmt.Println("vault-configure skipped: vault label not enabled")
		return nil
	}

	pgIdentity := util.DerivePostgresIdentity(ctx.Name)
	vaultAuthPath := "kubernetes"
	appServiceAccount := util.AppendKubeNameSuffix(ctx.Name, "-app-sa", 63)
	vaultAuthRole, vaultDBRole := util.DeriveVaultResourceNames(ctx.Namespace, pgIdentity.TeamID, pgIdentity.DBName)
	vaultCredentialsPath := fmt.Sprintf("database/creds/%s", vaultDBRole)

	if err := util.WriteServiceAccount(sdk, appServiceAccount, ctx.Namespace); err != nil {
		return fmt.Errorf("write service account: %w", err)
	}

	deploy, err := util.ReadDeployment("/kratix/output/deployment.yaml")
	if err != nil {
		return fmt.Errorf("read existing deployment: %w", err)
	}
	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("deployment has no containers")
	}

	appContainer := &deploy.Spec.Template.Spec.Containers[0]
	resolvedCommand, err := util.ResolveImageCommand(appContainer.Image)
	if err != nil {
		return fmt.Errorf("resolve image command for %q: %w", appContainer.Image, err)
	}

	appContainer.Env = []corev1.EnvVar{
		{Name: "PGHOST", Value: fmt.Sprintf("%s.%s.svc.cluster.local", pgIdentity.InstanceName, ctx.Namespace)},
		{Name: "DBNAME", Value: pgIdentity.DBName},
		{Name: "DB_CREDENTIALS_FILE", Value: util.VaultCredentialsFilePath},
		{Name: "APP_PID_FILE", Value: util.AppPIDFilePath},
	}
	appContainer.Command = []string{util.LauncherBinaryPath}
	appContainer.Args = resolvedCommand
	util.UpsertVolumeMount(appContainer, corev1.VolumeMount{Name: util.LauncherVolumeName, MountPath: "/vault/bin", ReadOnly: true})
	util.UpsertVolumeMount(appContainer, corev1.VolumeMount{Name: util.RuntimeVolumeName, MountPath: "/vault/run"})

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = map[string]string{}
	}
	for key, value := range util.BuildVaultAgentAnnotations(vaultAuthPath, vaultAuthRole, vaultCredentialsPath, appContainer.Name) {
		deploy.Spec.Template.Annotations[key] = value
	}

	trueValue := true
	deploy.Spec.Template.Spec.ShareProcessNamespace = &trueValue
	deploy.Spec.Template.Spec.ServiceAccountName = appServiceAccount
	util.UpsertVolume(&deploy.Spec.Template.Spec, corev1.Volume{
		Name:         util.LauncherVolumeName,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})
	util.UpsertVolume(&deploy.Spec.Template.Spec, corev1.Volume{
		Name:         util.RuntimeVolumeName,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})
	util.UpsertInitContainer(&deploy.Spec.Template.Spec, util.BuildPostgresWaitInitContainer(pgIdentity.InstanceName, ctx.Namespace))
	util.UpsertInitContainer(&deploy.Spec.Template.Spec, util.BuildLauncherInitContainer())
	util.UpsertContainer(&deploy.Spec.Template.Spec, util.BuildRotationSidecar())

	if err := util.WriteYAMLObject(sdk, "deployment.yaml", &deploy); err != nil {
		return fmt.Errorf("write updated deployment: %w", err)
	}

	_ = st.Set("database.serviceAccount", appServiceAccount)
	_ = st.Set("database.vaultEnabled", true)
	_ = st.Set("database.vaultAuthPath", fmt.Sprintf("auth/%s/login", vaultAuthPath))
	_ = st.Set("database.vaultRole", vaultAuthRole)
	_ = st.Set("database.vaultCredentialsPath", vaultCredentialsPath)
	_ = st.Set("database.credentialsFile", util.VaultCredentialsFilePath)

	fmt.Println("Finished executing configureVault.")
	return sdk.WriteStatus(st)
}
