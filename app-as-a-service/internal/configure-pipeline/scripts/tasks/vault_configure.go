package tasks

import (
	"fmt"
	"os"
	"strings"

	kratix "github.com/syntasso/kratix-go"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/syntasso/kratix-marketplace/app/util"
)

func VaultConfigure(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing VaultConfigure...")
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	dbDriver := util.MustStringOrEmpty(util.Get(res, "spec.dbDriver"))
	resourceName := util.MustString(util.Get(res, "metadata.name"))
	appName := util.MustString(util.Get(res, "spec.name"))
	namespace := util.MustString(util.Get(res, "metadata.namespace"))
	vaultEnabled := util.HasLabelTrue(res, vaultOptInLabelKey)
	fmt.Println("vault-configure inputs:", "dbDriver="+dbDriver, "name="+resourceName, "namespace="+namespace, fmt.Sprintf("%s=%t", vaultOptInLabelKey, vaultEnabled))

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

	vaultAddr := strings.TrimSpace(os.Getenv("VAULT_ADDR"))
	if vaultAddr == "" {
		return fmt.Errorf("VAULT_ADDR is required for vault-configure")
	}

	pgIdentity := derivePostgresIdentity(resourceName)
	vaultNames := deriveVaultArtifacts(namespace, appName, pgIdentity)

	deploy, err := util.ReadDeployment("/kratix/output/deployment.yaml")
	if err != nil {
		return fmt.Errorf("read existing deployment: %w", err)
	}

	if err := applyVaultSyncedSecretWiring(&deploy, pgIdentity, namespace, vaultNames.syncedCredentialsName); err != nil {
		return fmt.Errorf("wire vault-backed database secret in deployment: %w", err)
	}
	if err := util.WriteYAMLObject(sdk, "deployment.yaml", &deploy); err != nil {
		return fmt.Errorf("write updated deployment: %w", err)
	}

	bootstrapImage := util.GetenvOrDefault(vaultBootstrapImageEnv, defaultVaultBootstrapImage)

	authServiceAccount := buildVaultAuthServiceAccount(vaultNames.authServiceAccount, namespace)
	if err := util.WriteYAMLObject(sdk, "vault-auth-serviceaccount.yaml", authServiceAccount); err != nil {
		return fmt.Errorf("write vault auth serviceaccount: %w", err)
	}

	bootstrapServiceAccount := buildVaultBootstrapServiceAccount(vaultNames.bootstrapServiceAccount, namespace)
	if err := util.WriteYAMLObject(sdk, "vault-bootstrap-serviceaccount.yaml", bootstrapServiceAccount); err != nil {
		return fmt.Errorf("write vault bootstrap serviceaccount: %w", err)
	}

	bootstrapRole := buildVaultBootstrapRole(vaultNames.bootstrapRoleName, namespace, pgIdentity.credentialsSecretName)
	if err := util.WriteYAMLObject(sdk, "vault-bootstrap-role.yaml", bootstrapRole); err != nil {
		return fmt.Errorf("write vault bootstrap role: %w", err)
	}

	bootstrapRoleBinding := buildVaultBootstrapRoleBinding(vaultNames.bootstrapRoleBindingName, namespace, vaultNames.bootstrapRoleName, vaultNames.bootstrapServiceAccount)
	if err := util.WriteYAMLObject(sdk, "vault-bootstrap-rolebinding.yaml", bootstrapRoleBinding); err != nil {
		return fmt.Errorf("write vault bootstrap rolebinding: %w", err)
	}

	bootstrapJob := buildVaultBootstrapJob(bootstrapImage, namespace, pgIdentity, vaultNames)
	if err := util.WriteYAMLObject(sdk, "vault-bootstrap-job.yaml", bootstrapJob); err != nil {
		return fmt.Errorf("write vault bootstrap job: %w", err)
	}

	vaultConnection := buildVaultConnection(vaultNames.connectionName, namespace, vaultAddr)
	if err := util.WriteYAMLMap(sdk, "vault-connection.yaml", vaultConnection); err != nil {
		return fmt.Errorf("write vault connection: %w", err)
	}

	vaultAuth := buildVaultAuth(vaultNames.authName, namespace, vaultNames.connectionName, vaultNames.authPath, vaultNames.authRoleName, vaultNames.authServiceAccount)
	if err := util.WriteYAMLMap(sdk, "vault-auth.yaml", vaultAuth); err != nil {
		return fmt.Errorf("write vault auth: %w", err)
	}

	vaultDynamicSecret := buildVaultDynamicSecret(vaultNames.dynamicSecretName, namespace, vaultNames.authName, vaultNames.dbRoleName, vaultNames.syncedCredentialsName, appName)
	if err := util.WriteYAMLMap(sdk, "vault-dynamic-secret.yaml", vaultDynamicSecret); err != nil {
		return fmt.Errorf("write vault dynamic secret: %w", err)
	}

	_ = st.Set("database.vaultEnabled", true)
	_ = st.Set("database.vaultAuthPath", fmt.Sprintf("auth/%s/login", vaultNames.authPath))
	_ = st.Set("database.vaultRole", vaultNames.authRoleName)
	_ = st.Set("database.vaultCredentialsPath", fmt.Sprintf("database/creds/%s", vaultNames.dbRoleName))
	_ = st.Set("database.credentialsSecret", vaultNames.syncedCredentialsName)
	fmt.Println("Finished executing VaultConfigure.")
	return sdk.WriteStatus(st)
}

func buildVaultAuthServiceAccount(name, namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceAccount"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func buildVaultBootstrapServiceAccount(name, namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceAccount"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func buildVaultBootstrapRole(name, namespace, credentialsSecretName string) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "Role"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups:     []string{""},
			Resources:     []string{"secrets"},
			ResourceNames: []string{credentialsSecretName},
			Verbs:         []string{"get"},
		}},
	}
}

func buildVaultBootstrapRoleBinding(name, namespace, roleName, serviceAccount string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "RoleBinding"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      serviceAccount,
			Namespace: namespace,
		}},
	}
}

func buildVaultBootstrapJob(image, namespace string, pgIdentity postgresIdentity, vaultNames vaultArtifacts) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: "batch/v1", Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      vaultNames.bootstrapJobName,
			Namespace: namespace,
			Annotations: map[string]string{
				"kustomize.toolkit.fluxcd.io/force": "enabled",
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            util.Int32Ptr(6),
			TTLSecondsAfterFinished: util.Int32Ptr(600),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: vaultNames.bootstrapServiceAccount,
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{{
						Name:    "vault-bootstrap",
						Image:   image,
						Command: []string{"/bin/sh", "-ec"},
						Args:    []string{vaultBootstrapScript()},
						Env: []corev1.EnvVar{
							{
								Name: "VAULT_ADDR",
								ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql-vault"},
									Key:                  "address",
								}},
							},
							{
								Name: "VAULT_TOKEN",
								ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql-vault"},
									Key:                  "token",
								}},
							},
							{Name: "VAULT_AUTH_PATH", Value: vaultNames.authPath},
							{Name: "VAULT_POLICY_NAME", Value: vaultNames.policyName},
							{Name: "VAULT_AUTH_ROLE_NAME", Value: vaultNames.authRoleName},
							{Name: "VAULT_DB_ROLE_NAME", Value: vaultNames.dbRoleName},
							{Name: "VAULT_DB_CONFIG_NAME", Value: vaultNames.dbConfigName},
							{Name: "VAULT_DB_PLUGIN_NAME", Value: "postgresql-database-plugin"},
							{Name: "DB_NAME", Value: pgIdentity.dbName},
							{Name: "DB_HOST", Value: fmt.Sprintf("%s.%s.svc.cluster.local", pgIdentity.instanceName, namespace)},
							{Name: "APP_SERVICE_ACCOUNT", Value: vaultNames.authServiceAccount},
							{Name: "APP_SERVICE_ACCOUNT_NAMESPACE", Value: namespace},
							{Name: "CREDENTIALS_SECRET_NAME", Value: pgIdentity.credentialsSecretName},
							{Name: "CREDENTIALS_SECRET_NAMESPACE", Value: namespace},
							{Name: "ADMIN_SECRET_WAIT_TIMEOUT_SECONDS", Value: "900"},
							{Name: "CREDENTIAL_TTL", Value: "1h"},
							{Name: "CREDENTIAL_MAX_TTL", Value: "24h"},
							{Name: "VAULT_ADMIN_PORT", Value: "5432"},
							{Name: "VAULT_ADMIN_DATABASE", Value: "postgres"},
							{Name: "VAULT_CONNECTION_URL_PARAMETERS", Value: "sslmode=require"},
							{Name: "VAULT_DB_READY_TIMEOUT_SECONDS", Value: "900"},
							{
								Name: "ADMIN_USERNAME",
								ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.credentialsSecretName},
									Key:                  "username",
									Optional:             util.BoolPtr(true),
								}},
							},
							{
								Name: "ADMIN_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.credentialsSecretName},
									Key:                  "password",
									Optional:             util.BoolPtr(true),
								}},
							},
						},
					}},
				},
			},
		},
	}
}

func vaultBootstrapScript() string {
	return `set -eu

admin_username="${ADMIN_USERNAME:-}"
admin_password="${ADMIN_PASSWORD:-}"

secret_name="${CREDENTIALS_SECRET_NAME:-}"
secret_namespace="${CREDENTIALS_SECRET_NAMESPACE:-default}"
secret_wait_timeout="${ADMIN_SECRET_WAIT_TIMEOUT_SECONDS:-900}"
secret_wait_start="$(date +%s)"

while [ -z "$admin_username" ] || [ -z "$admin_password" ]; do
  if [ -n "$secret_name" ]; then
    username_b64="$(kubectl get secret "$secret_name" -n "$secret_namespace" -o jsonpath='{.data.username}' 2>/dev/null || true)"
    password_b64="$(kubectl get secret "$secret_name" -n "$secret_namespace" -o jsonpath='{.data.password}' 2>/dev/null || true)"

    if [ -n "$username_b64" ]; then
      admin_username="$(printf "%s" "$username_b64" | base64 -d 2>/dev/null || true)"
    fi
    if [ -n "$password_b64" ]; then
      admin_password="$(printf "%s" "$password_b64" | base64 -d 2>/dev/null || true)"
    fi
  fi

  if [ -n "$admin_username" ] && [ -n "$admin_password" ]; then
    break
  fi

  now="$(date +%s)"
  if [ $((now - secret_wait_start)) -ge "$secret_wait_timeout" ]; then
    echo "timed out waiting for admin credentials secret ${secret_namespace}/${secret_name}" >&2
    exit 1
  fi
  sleep 5
done

if [ -z "$admin_username" ] || [ -z "$admin_password" ]; then
  echo "missing admin credentials from postgres operator secret" >&2
  exit 1
fi

vault_url() {
  printf "%s/v1/%s" "${VAULT_ADDR%/}" "$1"
}

vault_write() {
  method="$1"
  path="$2"
  payload="$3"
  response_file="/tmp/vault-response.json"

  code="$(curl -sS -o "$response_file" -w '%{http_code}' -X "$method" \
    -H "X-Vault-Token: ${VAULT_TOKEN}" \
    -H "Content-Type: application/json" \
    --data "$payload" \
    "$(vault_url "$path")")"

  case "$code" in
    2*)
      return 0
      ;;
    *)
      echo "vault request failed: ${method} ${path} -> HTTP ${code}" >&2
      cat "$response_file" >&2 || true
      return 1
      ;;
  esac
}

connection_suffix=""
if [ -n "${VAULT_CONNECTION_URL_PARAMETERS}" ]; then
  connection_suffix="?${VAULT_CONNECTION_URL_PARAMETERS}"
fi

if [ -z "${DB_HOST:-}" ] || [ "${DB_HOST}" = "null" ]; then
  echo "missing DB_HOST for Vault bootstrap job" >&2
  exit 1
fi

admin_db="${VAULT_ADMIN_DATABASE:-postgres}"
db_ready_timeout="${VAULT_DB_READY_TIMEOUT_SECONDS:-900}"
db_ready_start="$(date +%s)"

while true; do
  if PGPASSWORD="$admin_password" psql \
    "host=${DB_HOST} port=${VAULT_ADMIN_PORT} dbname=${admin_db} user=$admin_username sslmode=require" \
    -tAc "SELECT 1" >/dev/null 2>/tmp/postgres-ready.err; then
    break
  fi

  now="$(date +%s)"
  if [ $((now - db_ready_start)) -ge "$db_ready_timeout" ]; then
    echo "timed out waiting for postgres at ${DB_HOST}:${VAULT_ADMIN_PORT}" >&2
    cat /tmp/postgres-ready.err >&2 || true
    exit 1
  fi
  sleep 5
done

db_exists="$(PGPASSWORD="$admin_password" psql \
  "host=${DB_HOST} port=${VAULT_ADMIN_PORT} dbname=${admin_db} user=$admin_username sslmode=require" \
  -tAc "SELECT 1 FROM pg_database WHERE datname = '${DB_NAME}'" 2>/tmp/postgres-db-check.err | tr -d '[:space:]' || true)"

if [ "$db_exists" != "1" ]; then
  PGPASSWORD="$admin_password" psql \
    "host=${DB_HOST} port=${VAULT_ADMIN_PORT} dbname=${admin_db} user=$admin_username sslmode=require" \
    -v ON_ERROR_STOP=1 \
    -c "CREATE DATABASE \"${DB_NAME}\" OWNER \"${admin_username}\";" \
    >/tmp/postgres-db-create.out 2>/tmp/postgres-db-create.err || {
      echo "failed creating database ${DB_NAME}" >&2
      cat /tmp/postgres-db-check.err >&2 || true
      cat /tmp/postgres-db-create.err >&2 || true
      exit 1
    }
fi

db_connection_url="postgresql://{{username}}:{{password}}@${DB_HOST}:${VAULT_ADMIN_PORT}/${DB_NAME}${connection_suffix}"

db_config_payload="$(jq -nc \
  --arg plugin_name "${VAULT_DB_PLUGIN_NAME}" \
  --arg allowed_roles "${VAULT_DB_ROLE_NAME}" \
  --arg connection_url "$db_connection_url" \
  --arg username "$admin_username" \
  --arg password "$admin_password" \
  '{plugin_name: $plugin_name, allowed_roles: $allowed_roles, connection_url: $connection_url, username: $username, password: $password, verify_connection: true}')"
vault_write POST "database/config/${VAULT_DB_CONFIG_NAME}" "$db_config_payload"

resource_policy="$(printf 'path "database/creds/%s" {\n  capabilities = ["read"]\n}\n' "${VAULT_DB_ROLE_NAME}")"
policy_payload="$(jq -nc --arg policy "$resource_policy" '{policy: $policy}')"
vault_write PUT "sys/policies/acl/${VAULT_POLICY_NAME}" "$policy_payload"

auth_payload="$(jq -nc \
  --arg sa "${APP_SERVICE_ACCOUNT}" \
  --arg sa_ns "${APP_SERVICE_ACCOUNT_NAMESPACE}" \
  --arg policy "${VAULT_POLICY_NAME}" \
  --arg ttl "${CREDENTIAL_TTL}" \
  --arg max_ttl "${CREDENTIAL_MAX_TTL}" \
  '{bound_service_account_names: [$sa], bound_service_account_namespaces: [$sa_ns], token_policies: [$policy], token_ttl: $ttl, token_max_ttl: $max_ttl}')"
vault_write POST "auth/${VAULT_AUTH_PATH}/role/${VAULT_AUTH_ROLE_NAME}" "$auth_payload"

creation_statements="$(cat <<EOF
CREATE ROLE "{{name}}" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';
GRANT CONNECT ON DATABASE "${DB_NAME}" TO "{{name}}";
GRANT USAGE, CREATE ON SCHEMA public TO "{{name}}";
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER ON ALL TABLES IN SCHEMA public TO "{{name}}";
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO "{{name}}";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER ON TABLES TO "{{name}}";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO "{{name}}";
EOF
)"
db_role_payload="$(jq -nc \
  --arg db_name "${VAULT_DB_CONFIG_NAME}" \
  --arg creation_statements "$creation_statements" \
  --arg default_ttl "${CREDENTIAL_TTL}" \
  --arg max_ttl "${CREDENTIAL_MAX_TTL}" \
  '{db_name: $db_name, creation_statements: $creation_statements, default_ttl: $default_ttl, max_ttl: $max_ttl}')"
vault_write POST "database/roles/${VAULT_DB_ROLE_NAME}" "$db_role_payload"`
}

func buildVaultConnection(name, namespace, address string) map[string]any {
	return map[string]any{
		"apiVersion": "secrets.hashicorp.com/v1beta1",
		"kind":       "VaultConnection",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]any{
			"address":       address,
			"skipTLSVerify": true,
		},
	}
}

func buildVaultAuth(name, namespace, connectionName, authPath, authRole, serviceAccount string) map[string]any {
	return map[string]any{
		"apiVersion": "secrets.hashicorp.com/v1beta1",
		"kind":       "VaultAuth",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]any{
			"method":             "kubernetes",
			"mount":              authPath,
			"vaultConnectionRef": connectionName,
			"kubernetes": map[string]any{
				"role":           authRole,
				"serviceAccount": serviceAccount,
			},
		},
	}
}

func buildVaultDynamicSecret(name, namespace, authName, dbRole, targetSecret, deploymentName string) map[string]any {
	return map[string]any{
		"apiVersion": "secrets.hashicorp.com/v1beta1",
		"kind":       "VaultDynamicSecret",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]any{
			"vaultAuthRef":   authName,
			"mount":          "database",
			"path":           fmt.Sprintf("creds/%s", dbRole),
			"renewalPercent": 67,
			"revoke":         true,
			"destination": map[string]any{
				"name":      targetSecret,
				"create":    true,
				"overwrite": true,
			},
			"rolloutRestartTargets": []map[string]any{{
				"kind": "Deployment",
				"name": deploymentName,
			}},
		},
	}
}
