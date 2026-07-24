package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// ByteSize is an integer size in bytes that can be parsed from a human-readable
// TOML string like "10MB", "500KB", or a plain integer like "10485760".
type ByteSize int64

// UnmarshalText implements encoding.TextUnmarshaler for TOML parsing.
func (b *ByteSize) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		*b = 0
		return nil
	}

	// Try plain integer first.
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		*b = ByteSize(n)
		return nil
	}

	// Parse suffixed form: "10MB", "500KB", "1GB".
	s = strings.ToUpper(s)
	var multiplier int64 = 1
	switch {
	case strings.HasSuffix(s, "GB"):
		multiplier = 1 << 30
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1 << 20
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		multiplier = 1 << 10
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "B"):
		multiplier = 1
		s = strings.TrimSuffix(s, "B")
	}

	n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return fmt.Errorf("invalid size: %q", string(text))
	}
	*b = ByteSize(n * float64(multiplier))
	return nil
}

// Bytes returns the size as an int64 number of bytes.
func (b ByteSize) Bytes() int64 { return int64(b) }

// Config holds all configuration for the application.
type Config struct {
	Server  ServerConfig      `toml:"server"`
	Resend  ResendConfig      `toml:"resend"`
	Senders map[string]string `toml:"senders"`
}

// ServerConfig holds the HTTP server configuration.
type ServerConfig struct {
	Address           string   `toml:"address"`
	MaxAttachmentSize ByteSize `toml:"max_attachment_size"`
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
