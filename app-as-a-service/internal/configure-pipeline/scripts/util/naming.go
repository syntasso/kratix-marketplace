package util

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	nonAlnumDashPattern = regexp.MustCompile(`[^a-z0-9-]+`)
	multiDashPattern    = regexp.MustCompile(`-+`)
)

func SanitizeIdentifier(value string) string {
	safe := strings.ToLower(value)
	safe = nonAlnumDashPattern.ReplaceAllString(safe, "-")
	safe = multiDashPattern.ReplaceAllString(safe, "-")
	safe = strings.Trim(safe, "-")
	return safe
}

func ShortenIdentifier(value string, maxLen int) string {
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
