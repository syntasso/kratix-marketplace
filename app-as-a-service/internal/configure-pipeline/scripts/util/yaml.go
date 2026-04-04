package util

import (
	"fmt"
	"os"

	kratix "github.com/syntasso/kratix-go"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

func WriteYAMLObject(sdk *kratix.KratixSDK, filename string, obj any) error {
	b, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	return sdk.WriteOutput(filename, b)
}

func WriteYAMLMap(sdk *kratix.KratixSDK, filename string, m map[string]any) error {
	b, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return sdk.WriteOutput(filename, b)
}

func ReadDeployment(path string) (appsv1.Deployment, error) {
	var d appsv1.Deployment
	b, err := os.ReadFile(path)
	if err != nil {
		return d, err
	}
	if err := yaml.Unmarshal(b, &d); err != nil {
		return d, err
	}
	if d.APIVersion == "" {
		d.APIVersion = "apps/v1"
	}
	if d.Kind == "" {
		d.Kind = "Deployment"
	}
	return d, nil
}

func WriteDestinationSelectors(ds []kratix.DestinationSelector) error {
	data, err := yaml.Marshal(ds)
	if err != nil {
		return fmt.Errorf("marshal destination selectors: %w", err)
	}
	return os.WriteFile("/kratix/metadata/destination-selectors.yaml", data, 0o644)
}
