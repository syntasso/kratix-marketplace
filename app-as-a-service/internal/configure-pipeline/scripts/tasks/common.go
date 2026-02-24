package tasks

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/syntasso/kratix-marketplace/app/util"
)

const (
	vaultOptInLabelKey         = "app-as-a-service.marketplace.kratix.io/vault"
	vaultBootstrapImageEnv     = "VAULT_BOOTSTRAP_IMAGE"
	defaultVaultBootstrapImage = "ghcr.io/syntasso/kratix-marketplace/app-as-a-service-configure-pipeline:v0.1.5"
	defaultVaultAuthPath       = "kubernetes"
)

type postgresReadinessSnapshot struct {
	Host           string
	Reconciled     string
	WorksSucceeded string
}

type postgresIdentity struct {
	teamID                string
	requestName           string
	dbName                string
	instanceName          string
	credentialsSecretName string
}

type vaultArtifacts struct {
	authPath                 string
	policyName               string
	authRoleName             string
	dbRoleName               string
	dbConfigName             string
	authServiceAccount       string
	bootstrapServiceAccount  string
	bootstrapRoleName        string
	bootstrapRoleBindingName string
	bootstrapJobName         string
	syncedCredentialsName    string
	connectionName           string
	authName                 string
	dynamicSecretName        string
}

func readPostgresReadinessSnapshot(namespace, requestName string) (postgresReadinessSnapshot, error) {
	resource := fmt.Sprintf("postgresql/%s", requestName)

	host, err := util.KubectlJSONPath(namespace, resource, "{.status.connectionDetails.host}")
	if err != nil {
		return postgresReadinessSnapshot{}, err
	}
	reconciled, err := util.KubectlJSONPath(namespace, resource, `{.status.conditions[?(@.type=="Reconciled")].status}`)
	if err != nil {
		return postgresReadinessSnapshot{}, err
	}
	worksSucceeded, err := util.KubectlJSONPath(namespace, resource, `{.status.conditions[?(@.type=="WorksSucceeded")].status}`)
	if err != nil {
		return postgresReadinessSnapshot{}, err
	}

	return postgresReadinessSnapshot{
		Host:           util.FirstField(host),
		Reconciled:     util.FirstField(reconciled),
		WorksSucceeded: util.FirstField(worksSucceeded),
	}, nil
}

func applyNonVaultDatabaseWiring(deploy *appsv1.Deployment, pgIdentity postgresIdentity, namespace string) error {
	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("deployment has no containers")
	}

	appContainer := &deploy.Spec.Template.Spec.Containers[0]
	appContainer.Env = []corev1.EnvVar{
		{Name: "PGHOST", Value: fmt.Sprintf("%s.%s.svc.cluster.local", pgIdentity.instanceName, namespace)},
		{Name: "DBNAME", Value: pgIdentity.dbName},
		{
			Name: "PGUSER",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.credentialsSecretName},
				Key:                  "username",
			}},
		},
		{
			Name: "PGPASSWORD",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.credentialsSecretName},
				Key:                  "password",
			}},
		},
		{
			Name: "DB_USER",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.credentialsSecretName},
				Key:                  "username",
			}},
		},
		{
			Name: "DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: pgIdentity.credentialsSecretName},
				Key:                  "password",
			}},
		},
	}

	return nil
}

func applyVaultSyncedSecretWiring(deploy *appsv1.Deployment, pgIdentity postgresIdentity, namespace, syncedSecret string) error {
	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("deployment has no containers")
	}

	appContainer := &deploy.Spec.Template.Spec.Containers[0]
	appContainer.Env = []corev1.EnvVar{
		{Name: "PGHOST", Value: fmt.Sprintf("%s.%s.svc.cluster.local", pgIdentity.instanceName, namespace)},
		{Name: "DBNAME", Value: pgIdentity.dbName},
		{
			Name: "PGUSER",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: syncedSecret},
				Key:                  "username",
			}},
		},
		{
			Name: "PGPASSWORD",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: syncedSecret},
				Key:                  "password",
			}},
		},
		{
			Name: "DB_USER",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: syncedSecret},
				Key:                  "username",
			}},
		},
		{
			Name: "DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: syncedSecret},
				Key:                  "password",
			}},
		},
	}

	return nil
}

func derivePostgresIdentity(appName string) postgresIdentity {
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
	safe := util.SanitizeIdentifier(appName)
	if safe == "" {
		safe = "app"
	}
	return util.ShortenIdentifier(safe, maxBaseLen)
}

func deriveVaultArtifacts(namespace, appName string, pgIdentity postgresIdentity) vaultArtifacts {
	safeApp := util.SanitizeIdentifier(appName)
	if safeApp == "" {
		safeApp = "app"
	}

	authRole, dbRole := deriveVaultResourceNames(namespace, pgIdentity.teamID, pgIdentity.dbName)
	safeIdentifier := util.SanitizeIdentifier(fmt.Sprintf("%s-%s-%s", namespace, pgIdentity.teamID, pgIdentity.dbName))
	if safeIdentifier == "" {
		safeIdentifier = "postgresql"
	}

	policyName := util.ShortenIdentifier("app-"+safeIdentifier+"-vault-sync", 64)
	dbConfigName := util.ShortenIdentifier("app-"+safeIdentifier+"-db-config", 64)

	return vaultArtifacts{
		authPath:                 defaultVaultAuthPath,
		policyName:               policyName,
		authRoleName:             authRole,
		dbRoleName:               dbRole,
		dbConfigName:             dbConfigName,
		authServiceAccount:       util.AppendKubeNameSuffix(safeApp, "-vault-auth-sa", 63),
		bootstrapServiceAccount:  util.AppendKubeNameSuffix(safeApp, "-vault-bootstrap-sa", 63),
		bootstrapRoleName:        util.AppendKubeNameSuffix(safeApp, "-vault-bootstrap-role", 63),
		bootstrapRoleBindingName: util.AppendKubeNameSuffix(safeApp, "-vault-bootstrap-rb", 63),
		bootstrapJobName:         util.AppendKubeNameSuffix(safeApp, "-vault-bootstrap", 63),
		syncedCredentialsName:    util.AppendKubeNameSuffix(safeApp, "-vault-db-credentials", 63),
		connectionName:           util.AppendKubeNameSuffix(safeApp, "-vault-conn", 63),
		authName:                 util.AppendKubeNameSuffix(safeApp, "-vault-auth", 63),
		dynamicSecretName:        util.AppendKubeNameSuffix(safeApp, "-vault-dynamic", 63),
	}
}

func deriveVaultResourceNames(namespace, teamID, dbName string) (string, string) {
	safeIdentifier := util.SanitizeIdentifier(fmt.Sprintf("%s-%s-%s", namespace, teamID, dbName))
	if safeIdentifier == "" {
		safeIdentifier = "postgresql"
	}

	baseName := "postgresql-" + safeIdentifier
	vaultAuthRole := util.ShortenIdentifier(baseName+"-auth", 48)
	vaultDBRole := util.ShortenIdentifier(baseName+"-db", 64)
	return vaultAuthRole, vaultDBRole
}
