package tasks

import (
	"fmt"
	"os"
	"path/filepath"

	kratix "github.com/syntasso/kratix-go"
)

func PromiseConfigure(sdk *kratix.KratixSDK) error {
	fmt.Println("Executing PromiseConfigure...")

	entries, err := os.ReadDir("/tmp/transfer/dependencies")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("promise-configure skipped: no dependencies bundled")
			return nil
		}
		return fmt.Errorf("read dependencies dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Clean(filepath.Join("/tmp/transfer/dependencies", entry.Name()))
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read dependency %s: %w", entry.Name(), err)
		}
		if err := sdk.WriteOutput(entry.Name(), content); err != nil {
			return fmt.Errorf("write dependency %s: %w", entry.Name(), err)
		}
	}

	fmt.Println("Finished executing PromiseConfigure.")
	return nil
}
