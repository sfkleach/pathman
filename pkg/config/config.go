package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the pathman configuration.
type Config struct {
	ManagedFolder string `json:"managed_folder"`
}

// GetConfigPath returns the path to the configuration file.
// It uses $XDG_CONFIG_HOME/pathman/config.json if set,
// otherwise falls back to $HOME/.config/pathman/config.json.
func GetConfigPath() (string, error) {
	var configDir string

	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		configDir = filepath.Join(xdgConfigHome, "pathman")
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config", "pathman")
	}

	return filepath.Join(configDir, "config.json"), nil
}

// Load reads the configuration from disk.
// Returns nil if the config file doesn't exist.
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetDefaultManagedFolder returns the default path for the managed folder.
func GetDefaultManagedFolder() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local", "bin", "pathman-links"), nil
}
