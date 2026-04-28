package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type PostgresIdentity struct {
	TeamID                string
	RequestName           string
	DBName                string
	InstanceName          string
	CredentialsSecretName string
}

func DerivePostgresIdentity(appName string) PostgresIdentity {
	// Postgres operator appends a hash to the instance label value; keep the base
	// app token short enough so generated labels stay <= 63 chars.
	baseName := shortenPostgresResourceBase(appName)
	teamID := baseName + "team"
	requestName := baseName + "db"
	dbName := requestName
	instanceName := fmt.Sprintf("%s-%s-postgresql", teamID, requestName)
	credentialsSecretName := fmt.Sprintf("%s.%s.credentials.postgresql.acid.zalan.do", teamID, instanceName)

	return PostgresIdentity{
		TeamID:                teamID,
		RequestName:           requestName,
		DBName:                dbName,
		InstanceName:          instanceName,
		CredentialsSecretName: credentialsSecretName,
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

func AppendKubeNameSuffix(base, suffix string, maxLen int) string {
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
