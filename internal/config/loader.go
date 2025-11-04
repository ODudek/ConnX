package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LegacyConfig represents the old configuration format for backward compatibility
type LegacyConfig struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Backends []string `yaml:"backends"`
	HealthCheck struct {
		Interval int `yaml:"interval"`
		Timeout  int `yaml:"timeout"`
	} `yaml:"healthCheck"`
}

// LoadWithLegacySupport loads config with support for both old and new formats
func LoadWithLegacySupport(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try loading as new format first
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err == nil {
		// Check if it's using new format (has BackendConfig objects)
		if len(config.Backends) > 0 {
			return config, nil
		}
	}

	// Try loading as legacy format
	legacyConfig := &LegacyConfig{}
	if err := yaml.Unmarshal(data, legacyConfig); err != nil {
		return nil, err
	}

	// Convert legacy to new format
	config = &Config{}
	config.Server.Port = legacyConfig.Server.Port
	config.Server.Host = legacyConfig.Server.Host
	config.HealthCheck.Interval = legacyConfig.HealthCheck.Interval
	config.HealthCheck.Timeout = legacyConfig.HealthCheck.Timeout

	// Convert string backends to BackendConfig
	for _, backendURL := range legacyConfig.Backends {
		config.Backends = append(config.Backends, BackendConfig{
			URL:    backendURL,
			Weight: 1,
		})
	}

	return config, nil
}
