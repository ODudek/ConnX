package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port            int `yaml:"port"`
		Host            string `yaml:"host"`
		ShutdownTimeout int `yaml:"shutdownTimeout"` // seconds
	} `yaml:"server"`

	Backends []BackendConfig `yaml:"backends"`

	LoadBalancing struct {
		Algorithm string `yaml:"algorithm"` // round-robin, weighted-round-robin, least-connections
	} `yaml:"loadBalancing"`

	HealthCheck struct {
		Interval int `yaml:"interval"`
		Timeout  int `yaml:"timeout"`
	} `yaml:"healthCheck"`

	CircuitBreaker struct {
		Enabled         bool `yaml:"enabled"`
		MaxFailures     int  `yaml:"maxFailures"`
		Timeout         int  `yaml:"timeout"`         // seconds
		ResetTimeout    int  `yaml:"resetTimeout"`    // seconds
		HalfOpenSuccess int  `yaml:"halfOpenSuccess"`
	} `yaml:"circuitBreaker"`

	RateLimit struct {
		Enabled        bool `yaml:"enabled"`
		RequestsPerSec int  `yaml:"requestsPerSec"`
		Burst          int  `yaml:"burst"`
		PerIP          bool `yaml:"perIP"` // Enable per-IP rate limiting
	} `yaml:"rateLimit"`
}

// BackendConfig represents a backend server configuration
type BackendConfig struct {
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight"` // For weighted round-robin (default: 1)
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
	// Server defaults
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = 30
	}

	// Health check defaults
	if c.HealthCheck.Interval == 0 {
		c.HealthCheck.Interval = 30
	}
	if c.HealthCheck.Timeout == 0 {
		c.HealthCheck.Timeout = 5
	}

	// Load balancing defaults
	if c.LoadBalancing.Algorithm == "" {
		c.LoadBalancing.Algorithm = "round-robin"
	}

	// Circuit breaker defaults
	if c.CircuitBreaker.MaxFailures == 0 {
		c.CircuitBreaker.MaxFailures = 5
	}
	if c.CircuitBreaker.Timeout == 0 {
		c.CircuitBreaker.Timeout = 60
	}
	if c.CircuitBreaker.ResetTimeout == 0 {
		c.CircuitBreaker.ResetTimeout = 30
	}
	if c.CircuitBreaker.HalfOpenSuccess == 0 {
		c.CircuitBreaker.HalfOpenSuccess = 2
	}

	// Rate limit defaults
	if c.RateLimit.RequestsPerSec == 0 {
		c.RateLimit.RequestsPerSec = 100
	}
	if c.RateLimit.Burst == 0 {
		c.RateLimit.Burst = c.RateLimit.RequestsPerSec * 2
	}

	// Backend weight defaults
	for i := range c.Backends {
		if c.Backends[i].Weight == 0 {
			c.Backends[i].Weight = 1
		}
	}

	return nil
}
