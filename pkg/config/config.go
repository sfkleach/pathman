package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ManagedDirectory represents a directory managed by pathman.
type ManagedDirectory struct {
	Path     string `json:"path"`
	Priority string `json:"priority"` // "front" or "back"
}

// Config represents the pathman configuration.
type Config struct {
	ManagedDirectories []ManagedDirectory `json:"managed_directories"`
}

// GetDefaultManagedFolder returns the default path for the managed folder.
// This is a variable to allow tests to override it.
var GetDefaultManagedFolder = func() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local", "bin", "pathman-links"), nil
}

// GetConfigPath returns the path to the configuration file.
// This is a variable to allow tests to override it.
var GetConfigPath = func() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "pathman", "config.json"), nil
}

// Load reads the configuration file and returns a Config struct.
// If the file doesn't exist, returns an empty Config.
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return empty config.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{ManagedDirectories: []ManagedDirectory{}}, nil
	}

	// #nosec G304 -- configPath comes from GetConfigPath which returns user's home directory path
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Initialize slice if nil.
	if config.ManagedDirectories == nil {
		config.ManagedDirectories = []ManagedDirectory{}
	}

	return &config, nil
}

// Save writes the configuration to the config file.
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist.
	configDir := filepath.Dir(configPath)
	// #nosec G301 -- 0755 permissions are standard for .config directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// #nosec G306 -- 0644 permissions are appropriate for config files with non-sensitive data
	return os.WriteFile(configPath, data, 0644)
}
