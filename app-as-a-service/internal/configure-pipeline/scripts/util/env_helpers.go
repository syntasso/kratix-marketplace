package util

import (
	"os"
	"strconv"
	"strings"
)

func GetenvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func GetenvIntOrDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	intValue, err := strconv.Atoi(value)
	if err != nil || intValue <= 0 {
		return fallback
	}
	return intValue
}
