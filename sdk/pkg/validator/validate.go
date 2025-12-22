package validator

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ValidationManifest struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	} `yaml:"metadata"`
}

func ValidateApp(path string) error {
	manifestPath := filepath.Join(path, "prospect.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("prospect.yaml not found in %s", path)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var m ValidationManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("invalid YAML in prospect.yaml: %w", err)
	}

	if m.APIVersion != "prospect/v1alpha1" {
		return fmt.Errorf("unsupported apiVersion: %s", m.APIVersion)
	}

	if m.Kind != "App" {
		return fmt.Errorf("invalid kind: %s, expected 'App'", m.Kind)
	}

	if m.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	return nil
}
