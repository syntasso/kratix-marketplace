package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultCredentialsFile = "/vault/secrets/pg-db.env"
	defaultPIDFile         = "/vault/run/app.pid"
	defaultWaitTimeout     = 120 * time.Second
)

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: vault-env-launcher <command> [args...]")
	}

	credentialsFile := getenvOrDefault("DB_CREDENTIALS_FILE", defaultCredentialsFile)
	pidFile := getenvOrDefault("APP_PID_FILE", defaultPIDFile)

	if err := waitForFile(credentialsFile, defaultWaitTimeout); err != nil {
		fatalf("wait for credentials file: %v", err)
	}
	if err := loadCredentials(credentialsFile); err != nil {
		fatalf("load credentials: %v", err)
	}
	if err := writePID(pidFile); err != nil {
		fatalf("write pid file: %v", err)
	}

	cmd := append([]string(nil), os.Args[1:]...)
	resolved, err := exec.LookPath(cmd[0])
	if err == nil {
		cmd[0] = resolved
	}

	if err := syscall.Exec(cmd[0], cmd, os.Environ()); err != nil {
		fatalf("exec wrapped command: %v", err)
	}
}

func waitForFile(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		info, err := os.Stat(path)
		if err == nil && info.Size() > 0 {
			return nil
		}
		if !errors.Is(err, os.ErrNotExist) && err != nil {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for %s", path)
		}
		time.Sleep(1 * time.Second)
	}
}

func loadCredentials(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid env line %q", line)
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return fmt.Errorf("empty env key in line %q", line)
		}
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"`)
		value = strings.Trim(value, `'`)
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func writePID(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0o644)
}

func getenvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
