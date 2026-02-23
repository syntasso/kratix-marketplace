package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	kratix "github.com/syntasso/kratix-go"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const (
	defaultLauncherImage      = "ghcr.io/syntasso/kratix-marketplace/app-as-a-service-configure-pipeline:v0.1.0"
	defaultReloaderImage      = "alpine:3.20"
	defaultPostgresWaitImage  = "alpine:3.20"
	vaultOptInLabelKey        = "app-as-a-service.marketplace.kratix.io/vault"
	launcherImageEnv          = "VAULT_LAUNCHER_IMAGE"
	reloaderImageEnv          = "VAULT_RELOADER_IMAGE"
	postgresWaitImageEnv      = "POSTGRES_WAIT_IMAGE"
	launcherInitContainerName = "vault-launcher-init"
	postgresWaitInitName      = "postgres-ready-wait"
	rotationSidecarName       = "vault-rotation-reloader"
	launcherBinaryPath        = "/vault/bin/vault-env-launcher"
	vaultCredentialsFilePath  = "/vault/secrets/pg-db.env"
	appPIDFilePath            = "/vault/run/app.pid"
	launcherVolumeName        = "vault-launcher-bin"
	runtimeVolumeName         = "vault-runtime"
)

var (
	nonAlnumDashPattern = regexp.MustCompile(`[^a-z0-9-]+`)
	multiDashPattern    = regexp.MustCompile(`-+`)
)

func main() {
	sdk := kratix.New()
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatalf("usage: %s <pipeline-name>", os.Args[0])
	}
	arg := args[0]
	fmt.Println("Pipeline requested:", arg)
	st, err := sdk.ReadStatus()
	if err != nil {
		st = kratix.NewStatus() // start with empty status
	}
	switch arg {
	case "resource-configure":
		if err := runResource(sdk, st); err != nil {
			log.Fatalf("resource pipeline: %v", err)
		}
	case "database-configure":
		if err := runDatabase(sdk, st); err != nil {
			log.Fatalf("database pipeline: %v", err)
		}
	case "vault-configure":
		if err := runVaultConfigure(sdk, st); err != nil {
			log.Fatalf("vault pipeline: %v", err)
		}
	default:
		log.Fatalf("unknown pipeline %q", sdk.PipelineName())
	}
	fmt.Println("Finished executing main.")
}

func runResource(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing runResource...")
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	image := mustString(get(res, "spec.image"))
	name := mustString(get(res, "spec.name"))
	namespace := mustString(get(res, "metadata.namespace"))
	servicePort := mustString(get(res, "spec.service.port"))
	fmt.Println("resource-configure inputs:", "image="+image, "name="+name, "namespace="+namespace, "servicePort="+servicePort)

	deployYAML := runKubectl(
		"create", "deployment", name,
		"--namespace="+namespace,
		"--replicas=1",
		"--image="+image,
		"--dry-run=client", "-o", "yaml",
	)
	if err := sdk.WriteOutput("deployment.yaml", deployYAML); err != nil {
		return fmt.Errorf("write deployment: %w", err)
	}

	svcYAML := runKubectl(
		"create", "service", "nodeport", name,
		"--namespace="+namespace,
		"--tcp="+servicePort+":8080",
		"--dry-run=client", "-o", "yaml",
	)
	if err := sdk.WriteOutput("service.yaml", svcYAML); err != nil {
		return fmt.Errorf("write service: %w", err)
	}

	rule := fmt.Sprintf("%s.local.gd/*=%s:%s", name, name, servicePort)
	ingYAML := runKubectl(
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
	fmt.Println("Finished executing runResource.")
	return sdk.WriteStatus(st)
}

func runDatabase(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing runDatabase...")
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	dbDriver := mustStringOrEmpty(get(res, "spec.dbDriver"))
	name := mustString(get(res, "metadata.name"))
	namespace := mustString(get(res, "metadata.namespace"))
	fmt.Println("database-configure inputs:", "dbDriver="+dbDriver, "name="+name, "namespace="+namespace)

	if dbDriver == "" || dbDriver == "none" {
		st := kratix.NewStatus()
		_ = st.Set("database", nil) // -> database: null
		return sdk.WriteStatus(st)
	}
	if dbDriver != "postgresql" {
		return fmt.Errorf("unsupported db driver %q. supported: postgresql", dbDriver)
	}

	pgIdentity := derivePostgresIdentity(name)

	// update deployment with non-vault database wiring via operator-managed secret
	deploy, err := readDeployment("/kratix/output/deployment.yaml")
	if err != nil {
		return fmt.Errorf("read existing deployment: %w", err)
	}
	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("deployment has no containers")
	}

	appContainer := &deploy.Spec.Template.Spec.Containers[0]
	appContainer.Env = []corev1.EnvVar{
		{Name: "PGHOST", Value: fmt.Sprintf("%s.%s.svc.cluster.local", pgIdentity.instanceName, namespace)},
		{Name: "DBNAME", Value: pgIdentity.dbName},
		{
			Name: "PGUSER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.credentialsSecretName},
					Key:                  "username",
				},
			},
		},
		{
			Name: "PGPASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.credentialsSecretName},
					Key:                  "password",
				},
			},
		},
		{
			Name: "DB_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.credentialsSecretName},
					Key:                  "username",
				},
			},
		},
		{
			Name: "DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.credentialsSecretName},
					Key:                  "password",
				},
			},
		},
	}

	if err := writeYAMLObject(sdk, "deployment.yaml", &deploy); err != nil {
		return fmt.Errorf("write updated deployment: %w", err)
	}

	// write base postgresql CR. Vault-specific fields are added in vault-configure.
	if err := os.MkdirAll(filepath.Clean("/kratix/output/platform"), 0o755); err != nil {
		return fmt.Errorf("mkdir platform: %w", err)
	}
	pg := &unstructured.Unstructured{Object: map[string]any{}}
	pg.SetAPIVersion("marketplace.kratix.io/v1alpha2")
	pg.SetKind("postgresql")
	pg.SetName(pgIdentity.requestName)
	pg.SetNamespace(namespace)

	_ = unstructured.SetNestedField(pg.Object, namespace, "spec", "namespace")
	_ = unstructured.SetNestedField(pg.Object, pgIdentity.teamID, "spec", "teamId")
	_ = unstructured.SetNestedField(pg.Object, pgIdentity.dbName, "spec", "dbName")

	if err := writeYAMLMap(sdk, "platform/postgresql.yaml", pg.Object); err != nil {
		return fmt.Errorf("write postgresql request: %w", err)
	}

	ds := []kratix.DestinationSelector{
		{Directory: "platform", MatchLabels: map[string]string{"environment": "platform"}},
	}
	if err := writeDestinationSelectors(ds); err != nil {
		return fmt.Errorf("write destination selectors: %w", err)
	}

	_ = st.Set("database.teamId", pgIdentity.teamID)
	_ = st.Set("database.dbName", pgIdentity.dbName)
	_ = st.Set("database.requestName", pgIdentity.requestName)
	_ = st.Set("database.instanceName", pgIdentity.instanceName)
	_ = st.Set("database.credentialsSecret", pgIdentity.credentialsSecretName)
	_ = st.Set("database.vaultEnabled", false)
	_ = st.Set("database.serviceAccount", nil)
	_ = st.Set("database.vaultAuthPath", nil)
	_ = st.Set("database.vaultRole", nil)
	_ = st.Set("database.vaultCredentialsPath", nil)
	_ = st.Set("database.credentialsFile", nil)
	fmt.Println("Finished executing runDatabase.")
	return sdk.WriteStatus(st)
}

func runVaultConfigure(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing runVaultConfigure...")
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	dbDriver := mustStringOrEmpty(get(res, "spec.dbDriver"))
	name := mustString(get(res, "metadata.name"))
	namespace := mustString(get(res, "metadata.namespace"))
	vaultEnabled := hasLabelTrue(res, vaultOptInLabelKey)
	fmt.Println("vault-configure inputs:", "dbDriver="+dbDriver, "name="+name, "namespace="+namespace, fmt.Sprintf("%s=%t", vaultOptInLabelKey, vaultEnabled))

	if dbDriver == "" || dbDriver == "none" {
		fmt.Println("vault-configure skipped: no database requested")
		return nil
	}
	if dbDriver != "postgresql" {
		return fmt.Errorf("unsupported db driver %q. supported: postgresql", dbDriver)
	}
	if !vaultEnabled {
		fmt.Println("vault-configure skipped: vault label not enabled")
		return nil
	}

	pgIdentity := derivePostgresIdentity(name)
	vaultAuthPath := "kubernetes"
	appServiceAccount := appendKubeNameSuffix(name, "-app-sa", 63)
	vaultAuthRole, vaultDBRole := deriveVaultResourceNames(namespace, pgIdentity.teamID, pgIdentity.dbName)
	vaultCredentialsPath := fmt.Sprintf("database/creds/%s", vaultDBRole)

	if err := writeServiceAccount(sdk, appServiceAccount, namespace); err != nil {
		return fmt.Errorf("write service account: %w", err)
	}

	deploy, err := readDeployment("/kratix/output/deployment.yaml")
	if err != nil {
		return fmt.Errorf("read existing deployment: %w", err)
	}
	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("deployment has no containers")
	}

	appContainer := &deploy.Spec.Template.Spec.Containers[0]
	resolvedCommand, err := resolveImageCommand(appContainer.Image)
	if err != nil {
		return fmt.Errorf("resolve image command for %q: %w", appContainer.Image, err)
	}

	appContainer.Env = []corev1.EnvVar{
		{Name: "PGHOST", Value: fmt.Sprintf("%s.%s.svc.cluster.local", pgIdentity.instanceName, namespace)},
		{Name: "DBNAME", Value: pgIdentity.dbName},
		{Name: "DB_CREDENTIALS_FILE", Value: vaultCredentialsFilePath},
		{Name: "APP_PID_FILE", Value: appPIDFilePath},
	}
	appContainer.Command = []string{launcherBinaryPath}
	appContainer.Args = resolvedCommand
	upsertVolumeMount(appContainer, corev1.VolumeMount{Name: launcherVolumeName, MountPath: "/vault/bin", ReadOnly: true})
	upsertVolumeMount(appContainer, corev1.VolumeMount{Name: runtimeVolumeName, MountPath: "/vault/run"})

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = map[string]string{}
	}
	for key, value := range buildVaultAgentAnnotations(vaultAuthPath, vaultAuthRole, vaultCredentialsPath, appContainer.Name) {
		deploy.Spec.Template.Annotations[key] = value
	}

	trueValue := true
	deploy.Spec.Template.Spec.ShareProcessNamespace = &trueValue
	deploy.Spec.Template.Spec.ServiceAccountName = appServiceAccount
	upsertVolume(&deploy.Spec.Template.Spec, corev1.Volume{
		Name:         launcherVolumeName,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})
	upsertVolume(&deploy.Spec.Template.Spec, corev1.Volume{
		Name:         runtimeVolumeName,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})
	upsertInitContainer(&deploy.Spec.Template.Spec, buildPostgresWaitInitContainer(pgIdentity.instanceName, namespace))
	upsertInitContainer(&deploy.Spec.Template.Spec, buildLauncherInitContainer())
	upsertContainer(&deploy.Spec.Template.Spec, buildRotationSidecar())

	if err := writeYAMLObject(sdk, "deployment.yaml", &deploy); err != nil {
		return fmt.Errorf("write updated deployment: %w", err)
	}

	pgObject, err := readYAMLMap("/kratix/output/platform/postgresql.yaml")
	if err != nil {
		return fmt.Errorf("read postgresql request: %w", err)
	}
	_ = unstructured.SetNestedField(pgObject, true, "spec", "vault", "enabled")
	_ = unstructured.SetNestedField(pgObject, appServiceAccount, "spec", "vault", "appServiceAccount")
	_ = unstructured.SetNestedField(pgObject, namespace, "spec", "vault", "appServiceAccountNamespace")
	_ = unstructured.SetNestedField(pgObject, vaultAuthPath, "spec", "vault", "kubernetesAuthPath")

	if err := writeYAMLMap(sdk, "platform/postgresql.yaml", pgObject); err != nil {
		return fmt.Errorf("write postgresql request: %w", err)
	}

	_ = st.Set("database.serviceAccount", appServiceAccount)
	_ = st.Set("database.vaultEnabled", true)
	_ = st.Set("database.vaultAuthPath", fmt.Sprintf("auth/%s/login", vaultAuthPath))
	_ = st.Set("database.vaultRole", vaultAuthRole)
	_ = st.Set("database.vaultCredentialsPath", vaultCredentialsPath)
	_ = st.Set("database.credentialsFile", vaultCredentialsFilePath)
	fmt.Println("Finished executing runVaultConfigure.")
	return sdk.WriteStatus(st)
}

func runKubectl(args ...string) []byte {
	fmt.Printf("Executing runKubectl: %v", args)
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Fatalf("kubectl %v failed: %v\nstderr:\n%s", args, err, string(ee.Stderr))
		}
		log.Fatalf("kubectl %v failed: %v", args, err)
	}
	return out
}

func resolveImageCommand(image string) ([]string, error) {
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

func buildLauncherInitContainer() corev1.Container {
	copyCommand := "cp /app/vault-env-launcher /vault/bin/vault-env-launcher && chmod 0755 /vault/bin/vault-env-launcher"

	return corev1.Container{
		Name:    launcherInitContainerName,
		Image:   getenvOrDefault(launcherImageEnv, defaultLauncherImage),
		Command: []string{"/bin/sh", "-ec"},
		Args:    []string{copyCommand},
		VolumeMounts: []corev1.VolumeMount{
			{Name: launcherVolumeName, MountPath: "/vault/bin"},
		},
	}
}

func buildPostgresWaitInitContainer(instanceName, namespace string) corev1.Container {
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
		Image:   getenvOrDefault(postgresWaitImageEnv, defaultPostgresWaitImage),
		Command: []string{"/bin/sh", "-ec"},
		Args:    []string{waitScript},
		Env: []corev1.EnvVar{
			{Name: "PGHOST", Value: fmt.Sprintf("%s.%s.svc.cluster.local", instanceName, namespace)},
			{Name: "PGPORT", Value: "5432"},
			{Name: "PG_WAIT_TIMEOUT_SECONDS", Value: "900"},
		},
	}
}

func buildRotationSidecar() corev1.Container {
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
		Image:   getenvOrDefault(reloaderImageEnv, defaultReloaderImage),
		Command: []string{"/bin/sh", "-ec"},
		Args:    []string{reloadScript},
		Env: []corev1.EnvVar{
			{Name: "DB_CREDENTIALS_FILE", Value: vaultCredentialsFilePath},
			{Name: "APP_PID_FILE", Value: appPIDFilePath},
			{Name: "ROTATION_CHECK_INTERVAL_SECONDS", Value: "15"},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: runtimeVolumeName, MountPath: "/vault/run"},
		},
	}
}

func buildVaultAgentAnnotations(vaultAuthPath, vaultAuthRole, vaultCredentialsPath, appContainerName string) map[string]string {
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

type postgresIdentity struct {
	teamID                string
	requestName           string
	dbName                string
	instanceName          string
	credentialsSecretName string
}

func derivePostgresIdentity(appName string) postgresIdentity {
	// Postgres operator appends a hash to the instance label value; keep the base
	// app token short enough so generated labels stay <= 63 chars.
	baseName := shortenPostgresResourceBase(appName)
	teamID := baseName + "team"
	requestName := baseName + "db"
	dbName := requestName
	instanceName := fmt.Sprintf("%s-%s-postgresql", teamID, requestName)
	credentialsSecretName := fmt.Sprintf("%s.%s.credentials.postgresql.acid.zalan.do", teamID, instanceName)

	return postgresIdentity{
		teamID:                teamID,
		requestName:           requestName,
		dbName:                dbName,
		instanceName:          instanceName,
		credentialsSecretName: credentialsSecretName,
	}
}

func shortenPostgresResourceBase(appName string) string {
	const maxBaseLen = 17

	safe := sanitizeIdentifier(appName)
	if safe == "" {
		safe = "app"
	}

	return shortenIdentifier(safe, maxBaseLen)
}

func deriveVaultResourceNames(namespace, teamID, dbName string) (string, string) {
	safeIdentifier := sanitizeIdentifier(fmt.Sprintf("%s-%s-%s", namespace, teamID, dbName))
	if safeIdentifier == "" {
		safeIdentifier = "postgresql"
	}

	baseName := "postgresql-" + safeIdentifier
	vaultAuthRole := shortenIdentifier(baseName+"-auth", 48)
	vaultDBRole := shortenIdentifier(baseName+"-db", 64)
	return vaultAuthRole, vaultDBRole
}

func sanitizeIdentifier(value string) string {
	safe := strings.ToLower(value)
	safe = nonAlnumDashPattern.ReplaceAllString(safe, "-")
	safe = multiDashPattern.ReplaceAllString(safe, "-")
	safe = strings.Trim(safe, "-")
	return safe
}

func shortenIdentifier(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}

	hashBytes := sha256.Sum256([]byte(value))
	hash := hex.EncodeToString(hashBytes[:])[:8]
	keepLen := maxLen - 9
	if keepLen < 1 {
		keepLen = 1
	}

	prefix := strings.TrimRight(value[:keepLen], "-")
	if prefix == "" {
		return hash
	}

	return prefix + "-" + hash
}

func writeYAMLObject(sdk *kratix.KratixSDK, filename string, obj any) error {
	b, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	return sdk.WriteOutput(filename, b)
}

func writeYAMLMap(sdk *kratix.KratixSDK, filename string, m map[string]any) error {
	b, err := yaml.Marshal(m) // preserves apiVersion/kind/metadata/spec casing
	if err != nil {
		return err
	}
	return sdk.WriteOutput(filename, b)
}

func writeServiceAccount(sdk *kratix.KratixSDK, name, namespace string) error {
	sa := map[string]any{
		"apiVersion": "v1",
		"kind":       "ServiceAccount",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
		},
	}
	return writeYAMLMap(sdk, "serviceaccount.yaml", sa)
}

func readDeployment(path string) (appsv1.Deployment, error) {
	var d appsv1.Deployment
	b, err := os.ReadFile(path)
	if err != nil {
		return d, err
	}
	if err := yaml.Unmarshal(b, &d); err != nil {
		return d, err
	}
	// Ensure TypeMeta present when re-serializing
	if d.APIVersion == "" {
		d.APIVersion = "apps/v1"
	}
	if d.Kind == "" {
		d.Kind = "Deployment"
	}
	return d, nil
}

func readYAMLMap(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := yaml.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func hasLabelTrue(res kratix.Resource, label string) bool {
	v, err := res.GetValue("metadata.labels")
	if err != nil || v == nil {
		return false
	}

	switch labels := v.(type) {
	case map[string]any:
		return isTruthy(labels[label])
	case map[string]string:
		value, ok := labels[label]
		if !ok {
			return false
		}
		return strings.EqualFold(strings.TrimSpace(value), "true")
	default:
		return false
	}
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}

	switch value := v.(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(strings.TrimSpace(value), "true")
	default:
		return strings.EqualFold(strings.TrimSpace(fmt.Sprintf("%v", value)), "true")
	}
}

func get(res kratix.Resource, path string) any {
	v, err := res.GetValue(path)
	if err != nil {
		log.Fatalf("get %s: %v", path, err)
	}
	return v
}

func mustString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		log.Fatalf("want string got %T", v)
		return ""
	}
}

func mustStringOrEmpty(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return mustString(v)
	}
}

func appendKubeNameSuffix(base, suffix string, maxLen int) string {
	cleanBase := strings.Trim(base, "-")
	if cleanBase == "" {
		cleanBase = "app"
	}

	maxBaseLen := maxLen - len(suffix)
	if maxBaseLen < 1 {
		return strings.Trim(strings.TrimPrefix(suffix, "-"), "-")
	}

	if len(cleanBase) > maxBaseLen {
		cleanBase = strings.TrimRight(cleanBase[:maxBaseLen], "-")
	}
	if cleanBase == "" {
		cleanBase = "app"
	}

	return cleanBase + suffix
}

func upsertVolume(spec *corev1.PodSpec, volume corev1.Volume) {
	for i := range spec.Volumes {
		if spec.Volumes[i].Name == volume.Name {
			spec.Volumes[i] = volume
			return
		}
	}
	spec.Volumes = append(spec.Volumes, volume)
}

func upsertVolumeMount(container *corev1.Container, mount corev1.VolumeMount) {
	for i := range container.VolumeMounts {
		if container.VolumeMounts[i].Name == mount.Name {
			container.VolumeMounts[i] = mount
			return
		}
	}
	container.VolumeMounts = append(container.VolumeMounts, mount)
}

func upsertContainer(spec *corev1.PodSpec, container corev1.Container) {
	for i := range spec.Containers {
		if spec.Containers[i].Name == container.Name {
			spec.Containers[i] = container
			return
		}
	}
	spec.Containers = append(spec.Containers, container)
}

func upsertInitContainer(spec *corev1.PodSpec, container corev1.Container) {
	for i := range spec.InitContainers {
		if spec.InitContainers[i].Name == container.Name {
			spec.InitContainers[i] = container
			return
		}
	}
	spec.InitContainers = append(spec.InitContainers, container)
}

func getenvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func writeDestinationSelectors(ds []kratix.DestinationSelector) error {
	data, err := yaml.Marshal(ds)
	if err != nil {
		return fmt.Errorf("marshal destination selectors: %w", err)
	}
	return os.WriteFile("/kratix/metadata/destination-selectors.yaml", data, 0o644)
}
