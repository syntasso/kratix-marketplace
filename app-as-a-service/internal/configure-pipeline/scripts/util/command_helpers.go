package util

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func RunKubectl(args ...string) []byte {
	fmt.Printf("Executing runKubectl: %v", args)
	out, err := runKubectlOutput(args...)
	if err != nil {
		log.Fatalf("kubectl %v failed: %v", args, err)
	}
	return out
}

func runKubectlOutput(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}
