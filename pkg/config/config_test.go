package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGetDefaultManagedFolder verifies the default folder path construction.
func TestGetDefaultManagedFolder(t *testing.T) {
	folder, err := GetDefaultManagedFolder()
	if err != nil {
		t.Fatalf("GetDefaultManagedFolder failed: %v", err)
	}

	if !filepath.IsAbs(folder) {
		t.Errorf("Expected absolute path, got %s", folder)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	if !strings.HasPrefix(folder, homeDir) {
		t.Errorf("Expected folder in HOME directory, got %s", folder)
	}

	if filepath.Base(folder) != "pathman-links" {
		t.Errorf("Expected 'pathman-links' as final component, got %s", filepath.Base(folder))
	}
}

// TestConfigLoadSave tests configuration persistence.
func TestConfigLoadSave(t *testing.T) {
	// Create temporary config directory.
	tmpDir := t.TempDir()

	// Temporarily override GetConfigPath for testing.
	origGetConfigPath := GetConfigPath
	GetConfigPath = func() (string, error) {
		return filepath.Join(tmpDir, "config.json"), nil
	}
	defer func() { GetConfigPath = origGetConfigPath }()

	// Test saving.
	cfg := &Config{
		ManagedDirectories: []ManagedDirectory{
			{Path: "/test/path", Priority: "front"},
			{Path: "/another/path", Priority: "back"},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test loading.
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(loaded.ManagedDirectories) != 2 {
		t.Errorf("Expected 2 directories, got %d", len(loaded.ManagedDirectories))
	}

	if loaded.ManagedDirectories[0].Path != "/test/path" {
		t.Errorf("Expected /test/path, got %s", loaded.ManagedDirectories[0].Path)
	}

	if loaded.ManagedDirectories[0].Priority != "front" {
		t.Errorf("Expected 'front' priority, got %s", loaded.ManagedDirectories[0].Priority)
	}

	if loaded.ManagedDirectories[1].Path != "/another/path" {
		t.Errorf("Expected /another/path, got %s", loaded.ManagedDirectories[1].Path)
	}

	if loaded.ManagedDirectories[1].Priority != "back" {
		t.Errorf("Expected 'back' priority, got %s", loaded.ManagedDirectories[1].Priority)
	}
}

// TestLoadNonexistentConfig verifies behavior when config doesn't exist.
func TestLoadNonexistentConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Temporarily override GetConfigPath for testing.
	origGetConfigPath := GetConfigPath
	GetConfigPath = func() (string, error) {
		return filepath.Join(tmpDir, "nonexistent", "config.json"), nil
	}
	defer func() { GetConfigPath = origGetConfigPath }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error for missing config, got %v", err)
	}

	if len(cfg.ManagedDirectories) != 0 {
		t.Errorf("Expected empty config, got %d directories", len(cfg.ManagedDirectories))
	}
}

// TestConfigSaveCreatesDirectory verifies that Save creates the config directory.
func TestConfigSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "dir", "config.json")

	// Temporarily override GetConfigPath for testing.
	origGetConfigPath := GetConfigPath
	GetConfigPath = func() (string, error) {
		return configPath, nil
	}
	defer func() { GetConfigPath = origGetConfigPath }()

	cfg := &Config{
		ManagedDirectories: []ManagedDirectory{
			{Path: "/test", Priority: "front"},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify the directory was created.
	if _, err := os.Stat(filepath.Dir(configPath)); os.IsNotExist(err) {
		t.Error("Config directory should have been created")
	}

	// Verify the file was created.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should have been created")
	}
}

// TestEmptyConfigSaveLoad tests saving and loading an empty config.
func TestEmptyConfigSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Temporarily override GetConfigPath for testing.
	origGetConfigPath := GetConfigPath
	GetConfigPath = func() (string, error) {
		return filepath.Join(tmpDir, "config.json"), nil
	}
	defer func() { GetConfigPath = origGetConfigPath }()

	// Save empty config.
	cfg := &Config{
		ManagedDirectories: []ManagedDirectory{},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Failed to save empty config: %v", err)
	}

	// Load it back.
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Failed to load empty config: %v", err)
	}

	if len(loaded.ManagedDirectories) != 0 {
		t.Errorf("Expected empty directories list, got %d", len(loaded.ManagedDirectories))
	}
}
