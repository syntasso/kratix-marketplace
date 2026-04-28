package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type PostgresReadinessSnapshot struct {
	Host                 string
	Reconciled           string
	WorksSucceeded       string
	VaultAuthPath        string
	VaultRole            string
	VaultCredentialsPath string
}

func ReadPostgresReadinessSnapshot(namespace, requestName string) (PostgresReadinessSnapshot, error) {
	resource := fmt.Sprintf("postgresql/%s", requestName)

	host, err := kubectlJSONPath(namespace, resource, "{.status.connectionDetails.host}")
	if err != nil {
		return PostgresReadinessSnapshot{}, err
	}

	reconciled, err := kubectlJSONPath(namespace, resource, `{.status.conditions[?(@.type=="Reconciled")].status}`)
	if err != nil {
		return PostgresReadinessSnapshot{}, err
	}

	worksSucceeded, err := kubectlJSONPath(namespace, resource, `{.status.conditions[?(@.type=="WorksSucceeded")].status}`)
	if err != nil {
		return PostgresReadinessSnapshot{}, err
	}

	vaultAuthPath, err := kubectlJSONPath(namespace, resource, "{.status.connectionDetails.vaultAuthPath}")
	if err != nil {
		return PostgresReadinessSnapshot{}, err
	}

	vaultRole, err := kubectlJSONPath(namespace, resource, "{.status.connectionDetails.vaultRole}")
	if err != nil {
		return PostgresReadinessSnapshot{}, err
	}

	vaultCredentialsPath, err := kubectlJSONPath(namespace, resource, "{.status.connectionDetails.vaultCredentialsPath}")
	if err != nil {
		return PostgresReadinessSnapshot{}, err
	}

	return PostgresReadinessSnapshot{
		Host:                 firstField(host),
		Reconciled:           firstField(reconciled),
		WorksSucceeded:       firstField(worksSucceeded),
		VaultAuthPath:        firstField(vaultAuthPath),
		VaultRole:            firstField(vaultRole),
		VaultCredentialsPath: firstField(vaultCredentialsPath),
	}, nil
}

func kubectlJSONPath(namespace, resource, path string) (string, error) {
	out, err := runKubectlOutput("get", resource, "--namespace="+namespace, "-o", "jsonpath="+path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func firstField(value string) string {
	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) == 0 {
		return ""
	}

	return fields[0]
}

func DeriveVaultResourceNames(namespace, teamID, dbName string) (string, string) {
	safeIdentifier := sanitizeIdentifier(fmt.Sprintf("%s-%s-%s", namespace, teamID, dbName))
	if safeIdentifier == "" {
		safeIdentifier = "postgresql"
	}

	baseName := "postgresql-" + safeIdentifier
	vaultAuthRole := shortenIdentifier(baseName+"-auth", 48)
	vaultDBRole := shortenIdentifier(baseName+"-db", 64)
	return vaultAuthRole, vaultDBRole
}

func IsVaultRoleConfigured(snapshot PostgresReadinessSnapshot) (bool, error) {
	vaultAddr := strings.TrimSpace(os.Getenv("VAULT_ADDR"))
	vaultToken := strings.TrimSpace(os.Getenv("VAULT_TOKEN"))
	if vaultAddr == "" || vaultToken == "" {
		return false, fmt.Errorf("vault enabled but VAULT_ADDR/VAULT_TOKEN not configured for wait-db-ready")
	}

	authMount := strings.TrimPrefix(snapshot.VaultAuthPath, "auth/")
	authMount = strings.TrimSuffix(authMount, "/login")
	if authMount == "" {
		return false, fmt.Errorf("invalid vault auth path %q", snapshot.VaultAuthPath)
	}

	dbRole := strings.TrimPrefix(snapshot.VaultCredentialsPath, "database/creds/")
	if dbRole == "" {
		return false, fmt.Errorf("invalid vault credentials path %q", snapshot.VaultCredentialsPath)
	}

	authRoleExists, err := vaultPathExists(vaultAddr, vaultToken, fmt.Sprintf("auth/%s/role/%s", authMount, snapshot.VaultRole))
	if err != nil {
		return false, err
	}
	if !authRoleExists {
		return false, nil
	}

	dbRoleExists, err := vaultPathExists(vaultAddr, vaultToken, fmt.Sprintf("database/roles/%s", dbRole))
	if err != nil {
		return false, err
	}

	return dbRoleExists, nil
}

func vaultPathExists(vaultAddr, vaultToken, path string) (bool, error) {
	url := fmt.Sprintf("%s/v1/%s", strings.TrimRight(vaultAddr, "/"), path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("X-Vault-Token", vaultToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return false, fmt.Errorf("vault path %q check failed with %d: %s", path, resp.StatusCode, strings.TrimSpace(string(body)))
}
