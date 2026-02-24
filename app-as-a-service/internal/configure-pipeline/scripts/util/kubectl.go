package util

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func RunKubectl(args ...string) []byte {
	fmt.Printf("Executing runKubectl: %v", args)
	out, err := RunKubectlOutput(args...)
	if err != nil {
		log.Fatalf("kubectl %v failed: %v", args, err)
	}
	return out
}

func RunKubectlOutput(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func KubectlJSONPath(namespace, resource, path string) (string, error) {
	out, err := RunKubectlOutput("get", resource, "--namespace="+namespace, "-o", "jsonpath="+path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func FirstField(value string) string {
	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
