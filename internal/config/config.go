package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
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

func Load(path string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.Server.Port == 0 {
		c.Server.Port = 8080 // domyślny port
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0" // domyślny host
	}
	if c.HealthCheck.Interval == 0 {
		c.HealthCheck.Interval = 30 // domyślny interwał health check (sekundy)
	}
	if c.HealthCheck.Timeout == 0 {
		c.HealthCheck.Timeout = 5 // domyślny timeout (sekundy)
	}
	return nil
}
