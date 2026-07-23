package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds all configuration for the application.
type Config struct {
	Server  ServerConfig   `toml:"server"`
	Resend  ResendConfig   `toml:"resend"`
	Senders map[string]string `toml:"senders"`
}

// ServerConfig holds the HTTP server configuration.
type ServerConfig struct {
	Address string `toml:"address"`
}

// ResendConfig holds the Resend API configuration.
type ResendConfig struct {
	APIKey string `toml:"api_key"`
}

// Load reads and parses the TOML configuration file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validate checks that the configuration values are sane.
func (c *Config) validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server.address is required")
	}
	if c.Resend.APIKey == "" || c.Resend.APIKey == "re_placeholder" {
		return fmt.Errorf("resend.api_key must be set to a valid Resend API key")
	}
	if len(c.Senders) == 0 {
		return fmt.Errorf("at least one sender identity is required in [senders]")
	}
	return nil
}
