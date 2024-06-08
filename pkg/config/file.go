package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func ReadConfigFromFile(configFilePath string) (*Config, error) {
	fileBytes, err := os.ReadFile(filepath.Clean(configFilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to ReadFile: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(fileBytes, config); err != nil {
		return nil, fmt.Errorf("failed to yaml.Unmarshal: %w", err)
	}

	return config, nil
}
