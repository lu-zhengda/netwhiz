package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	DefaultPingHost  string `yaml:"default_ping_host"`
	DefaultPingCount int    `yaml:"default_ping_count"`
	DefaultDNS       string `yaml:"default_dns"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		DefaultPingHost:  "google.com",
		DefaultPingCount: 5,
		DefaultDNS:       "cloudflare",
	}
}

// DefaultPath returns the default config file path.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "netwhiz", "config.yaml")
}

// Load loads the configuration from the given path.
// If the file doesn't exist, it returns the default configuration.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save writes the configuration to the given path.
func (c *Config) Save(path string) error {
	if path == "" {
		path = DefaultPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
