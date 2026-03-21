package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds the parsed watchdog.toml configuration.
type Config struct {
	NatsURL      string `toml:"nats_url"`
	LogPath      string `toml:"log_path"`
	NatsUser     string `toml:"nats_user"`
	NatsPassword string `toml:"nats_password"`
}

// Load reads and validates the TOML configuration file at path.
// Returns an error if the file is missing, malformed, or missing required fields.
func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	if cfg.NatsURL == "" {
		return nil, fmt.Errorf("required field 'nats_url' is missing or empty in %s", path)
	}

	return &cfg, nil
}
