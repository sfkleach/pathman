package folder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sfkleach/pathman/pkg/config"
)

// TestGetManagedFolder verifies managed folder path construction.
func TestGetManagedFolder(t *testing.T) {
	folder, err := GetManagedFolder()
	if err != nil {
		t.Fatalf("GetManagedFolder failed: %v", err)
	}

	if !filepath.IsAbs(folder) {
		t.Errorf("Expected absolute path, got %s", folder)
	}

	if filepath.Base(folder) != "pathman-links" {
		t.Errorf("Expected 'pathman-links' as final component, got %s", filepath.Base(folder))
	}
}

// TestGetFrontBackFolders verifies subfolder path construction.
func TestGetFrontBackFolders(t *testing.T) {
	frontFolder, err := GetFrontFolder()
	if err != nil {
		t.Fatalf("GetFrontFolder failed: %v", err)
	}

	backFolder, err := GetBackFolder()
	if err != nil {
		t.Fatalf("GetBackFolder failed: %v", err)
	}

	if !filepath.IsAbs(frontFolder) || !filepath.IsAbs(backFolder) {
		t.Error("Expected absolute paths")
	}

	if filepath.Base(frontFolder) != "front" {
		t.Errorf("Expected 'front' subfolder, got %s", filepath.Base(frontFolder))
	}

	if filepath.Base(backFolder) != "back" {
		t.Errorf("Expected 'back' subfolder, got %s", filepath.Base(backFolder))
	}
}

// TestExistsCreate tests folder existence checking and creation.
func TestExistsCreate(t *testing.T) {
	tmpDir := t.TempDir()
	testFolder := filepath.Join(tmpDir, "test-folder")

	// Should not exist initially.
	if Exists(testFolder) {
		t.Error("Folder should not exist initially")
	}

	// Create it.
	if err := Create(testFolder); err != nil {
		t.Fatalf("Failed to create folder: %v", err)
	}

	// Should exist now.
	if !Exists(testFolder) {
		t.Error("Folder should exist after creation")
	}

	// Creating again should not fail.
	if err := Create(testFolder); err != nil {
		t.Errorf("Creating existing folder should not fail: %v", err)
	}
}

// TestAddSymlink tests adding a symlink to managed folder.
func TestAddSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	frontDir := filepath.Join(tmpDir, "front")
	backDir := filepath.Join(tmpDir, "back")

	// Create directories.
	if err := os.MkdirAll(frontDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}
	if err := os.MkdirAll(backDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Create a test executable.
	testExec := filepath.Join(tmpDir, "test-exec")
	if err := os.WriteFile(testExec, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create test executable: %v", err)
	}

	// Override config for testing.
	origGetConfigPath := config.GetConfigPath
	config.GetConfigPath = func() (string, error) {
		return filepath.Join(tmpDir, "config.json"), nil
	}
	defer func() { config.GetConfigPath = origGetConfigPath }()

	// Mock GetFrontFolder and GetBackFolder.
	origGetDefaultManagedFolder := config.GetDefaultManagedFolder
	config.GetDefaultManagedFolder = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { config.GetDefaultManagedFolder = origGetDefaultManagedFolder }()

	// Test adding to back folder.
	if err := Add(testExec, "mytest", false, false); err != nil {
		t.Fatalf("Failed to add symlink: %v", err)
	}

	// Verify symlink exists.
	linkPath := filepath.Join(backDir, "mytest")
	if _, err := os.Lstat(linkPath); err != nil {
		t.Errorf("Symlink not created: %v", err)
	}

	// Verify it points to absolute path.
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	if !filepath.IsAbs(target) {
		t.Errorf("Expected absolute target path, got %s", target)
	}
}

// TestAddDuplicateSymlink tests that adding duplicate fails without force.
func TestAddDuplicateSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	backDir := filepath.Join(tmpDir, "back")
	if err := os.MkdirAll(backDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	testExec := filepath.Join(tmpDir, "test-exec")
	if err := os.WriteFile(testExec, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Override config for testing.
	origGetDefaultManagedFolder := config.GetDefaultManagedFolder
	config.GetDefaultManagedFolder = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { config.GetDefaultManagedFolder = origGetDefaultManagedFolder }()

	origGetConfigPath := config.GetConfigPath
	config.GetConfigPath = func() (string, error) {
		return filepath.Join(tmpDir, "config.json"), nil
	}
	defer func() { config.GetConfigPath = origGetConfigPath }()

	// Add once - should succeed.
	if err := Add(testExec, "test", false, false); err != nil {
		t.Fatalf("First add should succeed: %v", err)
	}

	// Add again without force - should fail.
	if err := Add(testExec, "test", false, false); err == nil {
		t.Error("Second add should fail without --force")
	}

	// Add again with force - should succeed.
	if err := Add(testExec, "test", false, true); err != nil {
		t.Errorf("Add with --force should succeed: %v", err)
	}
}

// TestRemoveSymlink tests removing a symlink.
func TestRemoveSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	backDir := filepath.Join(tmpDir, "back")
	frontDir := filepath.Join(tmpDir, "front")
	if err := os.MkdirAll(backDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(frontDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Override config for testing.
	origGetDefaultManagedFolder := config.GetDefaultManagedFolder
	config.GetDefaultManagedFolder = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { config.GetDefaultManagedFolder = origGetDefaultManagedFolder }()

	origGetConfigPath := config.GetConfigPath
	config.GetConfigPath = func() (string, error) {
		return filepath.Join(tmpDir, "config.json"), nil
	}
	defer func() { config.GetConfigPath = origGetConfigPath }()

	// Create a test symlink.
	linkPath := filepath.Join(backDir, "testlink")
	if err := os.Symlink("/usr/bin/true", linkPath); err != nil {
		t.Fatalf("Failed to create test symlink: %v", err)
	}

	// Remove it.
	if err := Remove("testlink"); err != nil {
		t.Fatalf("Failed to remove symlink: %v", err)
	}

	// Verify it's gone.
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Error("Symlink should have been removed")
	}
}

// TestRename tests renaming a symlink.
func TestRename(t *testing.T) {
	tmpDir := t.TempDir()
	backDir := filepath.Join(tmpDir, "back")
	frontDir := filepath.Join(tmpDir, "front")
	if err := os.MkdirAll(backDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(frontDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Override config for testing.
	origGetDefaultManagedFolder := config.GetDefaultManagedFolder
	config.GetDefaultManagedFolder = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { config.GetDefaultManagedFolder = origGetDefaultManagedFolder }()

	// Create a test symlink.
	oldPath := filepath.Join(backDir, "oldname")
	targetPath := "/usr/bin/true"
	if err := os.Symlink(targetPath, oldPath); err != nil {
		t.Fatalf("Failed to create test symlink: %v", err)
	}

	// Rename it.
	if err := Rename("oldname", "newname"); err != nil {
		t.Fatalf("Failed to rename symlink: %v", err)
	}

	// Verify old name is gone and new name exists.
	if _, err := os.Lstat(oldPath); !os.IsNotExist(err) {
		t.Error("Old symlink should be removed")
	}

	newPath := filepath.Join(backDir, "newname")
	if _, err := os.Lstat(newPath); err != nil {
		t.Error("New symlink should exist")
	}

	// Verify target is preserved.
	newTarget, err := os.Readlink(newPath)
	if err != nil {
		t.Fatalf("Failed to read new symlink: %v", err)
	}

	if newTarget != targetPath {
		t.Errorf("Expected target %s, got %s", targetPath, newTarget)
	}
}

// TestList tests listing symlinks.
func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	backDir := filepath.Join(tmpDir, "back")
	if err := os.MkdirAll(backDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Override config for testing.
	origGetDefaultManagedFolder := config.GetDefaultManagedFolder
	config.GetDefaultManagedFolder = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { config.GetDefaultManagedFolder = origGetDefaultManagedFolder }()

	// Create test symlinks.
	if err := os.Symlink("/usr/bin/true", filepath.Join(backDir, "link2")); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}
	if err := os.Symlink("/usr/bin/false", filepath.Join(backDir, "link1")); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// List them.
	entries, err := List(false)
	if err != nil {
		t.Fatalf("Failed to list: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// Entries should be present (order not guaranteed by List).
	found := make(map[string]bool)
	for _, entry := range entries {
		found[entry] = true
	}

	if !found["link1"] || !found["link2"] {
		t.Error("Expected link1 and link2 in list")
	}
}

// TestGetAdjustedPath tests PATH manipulation.
func TestGetAdjustedPath(t *testing.T) {
	tmpDir := t.TempDir()
	frontDir := filepath.Join(tmpDir, "front")
	backDir := filepath.Join(tmpDir, "back")
	testDir := filepath.Join(tmpDir, "test-bin")
	if err := os.MkdirAll(frontDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(backDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Override config for testing.
	origGetDefaultManagedFolder := config.GetDefaultManagedFolder
	config.GetDefaultManagedFolder = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { config.GetDefaultManagedFolder = origGetDefaultManagedFolder }()

	origGetConfigPath := config.GetConfigPath
	config.GetConfigPath = func() (string, error) {
		return filepath.Join(tmpDir, "config.json"), nil
	}
	defer func() { config.GetConfigPath = origGetConfigPath }()

	// Create config with managed directory.
	cfg := &config.Config{
		ManagedDirectories: []config.ManagedDirectory{
			{Path: testDir, Priority: "front"},
		},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Set PATH environment.
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)
	os.Setenv("PATH", "/usr/bin:/bin")

	// Test PATH adjustment.
	newPath, err := GetAdjustedPath()
	if err != nil {
		t.Fatalf("GetAdjustedPath failed: %v", err)
	}

	// Verify order: front-subfolder : front-dirs : original : back-subfolder.
	parts := strings.Split(newPath, string(os.PathListSeparator))

	if parts[0] != frontDir {
		t.Errorf("Expected front subfolder first, got %s", parts[0])
	}

	if parts[len(parts)-1] != backDir {
		t.Errorf("Expected back subfolder last, got %s", parts[len(parts)-1])
	}

	// Verify managed directory is included.
	foundTestDir := false
	for _, part := range parts {
		if part == testDir {
			foundTestDir = true
			break
		}
	}

	if !foundTestDir {
		t.Error("Expected managed directory in PATH")
	}
}

// TestCheckNameClashes tests detection of name clashes between front and back.
func TestCheckNameClashes(t *testing.T) {
	tmpDir := t.TempDir()
	frontDir := filepath.Join(tmpDir, "front")
	backDir := filepath.Join(tmpDir, "back")
	if err := os.MkdirAll(frontDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(backDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Override config for testing.
	origGetDefaultManagedFolder := config.GetDefaultManagedFolder
	config.GetDefaultManagedFolder = func() (string, error) {
		return tmpDir, nil
	}
	defer func() { config.GetDefaultManagedFolder = origGetDefaultManagedFolder }()

	// Create symlinks with same name in both folders.
	if err := os.Symlink("/usr/bin/true", filepath.Join(frontDir, "samename")); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}
	if err := os.Symlink("/usr/bin/false", filepath.Join(backDir, "samename")); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}
	if err := os.Symlink("/usr/bin/ls", filepath.Join(frontDir, "unique")); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	clashes, err := CheckNameClashes()
	if err != nil {
		t.Fatalf("CheckNameClashes failed: %v", err)
	}

	if len(clashes) != 1 {
		t.Errorf("Expected 1 clash, got %d", len(clashes))
	}

	if len(clashes) > 0 && clashes[0] != "samename" {
		t.Errorf("Expected 'samename' clash, got %s", clashes[0])
	}
}
