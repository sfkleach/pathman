package config

import (
	"os"
	"path/filepath"
)

// GetDefaultManagedFolder returns the default path for the managed folder.
func GetDefaultManagedFolder() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local", "bin", "pathman-links"), nil
}
