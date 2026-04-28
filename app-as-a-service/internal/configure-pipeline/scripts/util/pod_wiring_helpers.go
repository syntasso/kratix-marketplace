package util

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func ResolveImageCommand(image string) ([]string, error) {
	ref, err := name.ParseReference(image, name.WeakValidation)
	if err != nil {
		return nil, fmt.Errorf("parse image reference: %w", err)
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, fmt.Errorf("fetch image config: %w", err)
	}

	cfgFile, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("read image config: %w", err)
	}

	resolved := make([]string, 0, len(cfgFile.Config.Entrypoint)+len(cfgFile.Config.Cmd))
	resolved = append(resolved, cfgFile.Config.Entrypoint...)
	resolved = append(resolved, cfgFile.Config.Cmd...)
	if len(resolved) == 0 {
		return nil, fmt.Errorf("image has no entrypoint/cmd to wrap")
	}

	return resolved, nil
}

func ApplyNonVaultDatabaseWiring(deploy *appsv1.Deployment, pgIdentity PostgresIdentity, namespace string) error {
	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("deployment has no containers")
	}

	appContainer := &deploy.Spec.Template.Spec.Containers[0]
	appContainer.Env = []corev1.EnvVar{
		{Name: "PGHOST", Value: fmt.Sprintf("%s.%s.svc.cluster.local", pgIdentity.InstanceName, namespace)},
		{Name: "DBNAME", Value: pgIdentity.DBName},
		{
			Name: "PGUSER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.CredentialsSecretName},
					Key:                  "username",
				},
			},
		},
		{
			Name: "PGPASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.CredentialsSecretName},
					Key:                  "password",
				},
			},
		},
		{
			Name: "DB_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.CredentialsSecretName},
					Key:                  "username",
				},
			},
		},
		{
			Name: "DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.CredentialsSecretName},
					Key:                  "password",
				},
			},
		},
	}

	UpsertInitContainer(&deploy.Spec.Template.Spec, BuildPostgresWaitInitContainer(pgIdentity.InstanceName, namespace))
	return nil
}

func BuildLauncherInitContainer() corev1.Container {
	copyCommand := "cp /app/vault-env-launcher /vault/bin/vault-env-launcher && chmod 0755 /vault/bin/vault-env-launcher"

	return corev1.Container{
		Name:    launcherInitContainerName,
		Image:   GetenvOrDefault(launcherImageEnv, defaultLauncherImage),
		Command: []string{"/bin/sh", "-ec"},
		Args:    []string{copyCommand},
		VolumeMounts: []corev1.VolumeMount{
			{Name: LauncherVolumeName, MountPath: "/vault/bin"},
		},
	}
}

func BuildPostgresWaitInitContainer(instanceName, namespace string) corev1.Container {
	waitScript := `set -eu
HOST="${PGHOST:?missing PGHOST}"
PORT="${PGPORT:-5432}"
TIMEOUT="${PG_WAIT_TIMEOUT_SECONDS:-900}"
START="$(date +%s)"

while ! nc -z "$HOST" "$PORT" >/dev/null 2>&1; do
  NOW="$(date +%s)"
  if [ $((NOW - START)) -ge "$TIMEOUT" ]; then
    echo "timed out waiting for postgres at ${HOST}:${PORT}" >&2
    exit 1
  fi
  sleep 2
done`

	return corev1.Container{
		Name:    postgresWaitInitName,
		Image:   GetenvOrDefault(postgresWaitImageEnv, defaultPostgresWaitImage),
		Command: []string{"/bin/sh", "-ec"},
		Args:    []string{waitScript},
		Env: []corev1.EnvVar{
			{Name: "PGHOST", Value: fmt.Sprintf("%s.%s.svc.cluster.local", instanceName, namespace)},
			{Name: "PGPORT", Value: "5432"},
			{Name: "PG_WAIT_TIMEOUT_SECONDS", Value: "900"},
		},
	}
}

func BuildRotationSidecar() corev1.Container {
	reloadScript := `set -eu
ENV_FILE="${DB_CREDENTIALS_FILE:-/vault/secrets/pg-db.env}"
PID_FILE="${APP_PID_FILE:-/vault/run/app.pid}"
INTERVAL="${ROTATION_CHECK_INTERVAL_SECONDS:-15}"

while [ ! -s "$ENV_FILE" ]; do
  sleep 2
done
last="$(sha256sum "$ENV_FILE" | awk '{print $1}')"

while true; do
  sleep "$INTERVAL"
  [ -s "$ENV_FILE" ] || continue
  current="$(sha256sum "$ENV_FILE" | awk '{print $1}')"
  if [ "$current" != "$last" ]; then
    if [ -f "$PID_FILE" ]; then
      pid="$(cat "$PID_FILE" 2>/dev/null || true)"
      if [ -n "$pid" ]; then
        kill -TERM "$pid" 2>/dev/null || true
      fi
    fi
    last="$current"
  fi
done`

	return corev1.Container{
		Name:    rotationSidecarName,
		Image:   GetenvOrDefault(reloaderImageEnv, defaultReloaderImage),
		Command: []string{"/bin/sh", "-ec"},
		Args:    []string{reloadScript},
		Env: []corev1.EnvVar{
			{Name: "DB_CREDENTIALS_FILE", Value: VaultCredentialsFilePath},
			{Name: "APP_PID_FILE", Value: AppPIDFilePath},
			{Name: "ROTATION_CHECK_INTERVAL_SECONDS", Value: "15"},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: RuntimeVolumeName, MountPath: "/vault/run"},
		},
	}
}

func BuildVaultAgentAnnotations(vaultAuthPath, vaultAuthRole, vaultCredentialsPath, appContainerName string) map[string]string {
	const secretAlias = "pg-creds"

	return map[string]string{
		"vault.hashicorp.com/agent-inject":                         "true",
		"vault.hashicorp.com/agent-pre-populate":                   "true",
		"vault.hashicorp.com/auth-path":                            fmt.Sprintf("auth/%s", vaultAuthPath),
		"vault.hashicorp.com/role":                                 vaultAuthRole,
		"vault.hashicorp.com/agent-inject-containers":              fmt.Sprintf("%s,%s", appContainerName, rotationSidecarName),
		"vault.hashicorp.com/agent-inject-secret-" + secretAlias:   vaultCredentialsPath,
		"vault.hashicorp.com/agent-inject-file-" + secretAlias:     "pg-db.env",
		"vault.hashicorp.com/agent-inject-template-" + secretAlias: vaultCredentialsTemplate(vaultCredentialsPath),
		"vault.hashicorp.com/agent-inject-command-" + secretAlias:  "chmod 0644 /vault/secrets/pg-db.env",
	}
}

func vaultCredentialsTemplate(vaultCredentialsPath string) string {
	return fmt.Sprintf(`{{- with secret %q -}}
PGUSER={{ .Data.username }}
PGPASSWORD={{ .Data.password }}
DB_USER={{ .Data.username }}
DB_PASSWORD={{ .Data.password }}
{{- end -}}`, vaultCredentialsPath)
}

func UpsertVolume(spec *corev1.PodSpec, volume corev1.Volume) {
	for i := range spec.Volumes {
		if spec.Volumes[i].Name == volume.Name {
			spec.Volumes[i] = volume
			return
		}
	}
	spec.Volumes = append(spec.Volumes, volume)
}

func UpsertVolumeMount(container *corev1.Container, mount corev1.VolumeMount) {
	for i := range container.VolumeMounts {
		if container.VolumeMounts[i].Name == mount.Name {
			container.VolumeMounts[i] = mount
			return
		}
	}
	container.VolumeMounts = append(container.VolumeMounts, mount)
}

func UpsertContainer(spec *corev1.PodSpec, container corev1.Container) {
	for i := range spec.Containers {
		if spec.Containers[i].Name == container.Name {
			spec.Containers[i] = container
			return
		}
	}
	spec.Containers = append(spec.Containers, container)
}

func UpsertInitContainer(spec *corev1.PodSpec, container corev1.Container) {
	for i := range spec.InitContainers {
		if spec.InitContainers[i].Name == container.Name {
			spec.InitContainers[i] = container
			return
		}
	}
	spec.InitContainers = append(spec.InitContainers, container)
}
