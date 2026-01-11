package folder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sfkleach/pathman/pkg/config"
)

// GetShellIntegrationScript returns the shell script lines for PATH integration.
// The script checks for pathman on PATH first, then falls back to ~/.local/pathman/bin/pathman.
func GetShellIntegrationScript() []string {
	return []string{
		"if command -v pathman >/dev/null 2>&1; then",
		"  PATHMAN_CMD=pathman",
		"elif [ -x \"$HOME/.local/pathman/bin/pathman\" ]; then",
		"  PATHMAN_CMD=\"$HOME/.local/pathman/bin/pathman\"",
		"fi",
		"",
		"if [ -n \"$PATHMAN_CMD\" ]; then",
		"  # Calculate a new $PATH from the old one and pathman's configuration.",
		"  NEW_PATH=$(\"$PATHMAN_CMD\" path 2>/dev/null)",
		"  if [ $? -eq 0 ] && [ -n \"$NEW_PATH\" ]; then",
		"    export PATH=\"$NEW_PATH\"",
		"  elif [ -n \"$PS1\" ]; then",
		"    # PS1 is only set in interactive shells - safe to show errors here.",
		"    echo \"Warning: pathman failed to update PATH\" >&2",
		"  fi",
		"elif [ -n \"$PS1\" ]; then",
		"  # PS1 is only set in interactive shells - safe to show errors here.",
		"  echo \"Warning: pathman not found, PATH not updated\" >&2",
		"fi",
	}
}

// GetManagedFolder returns the path to the managed folder.
func GetManagedFolder() (string, error) {
	return config.GetDefaultManagedFolder()
}

// GetFrontFolder returns the path to the front subfolder.
func GetFrontFolder() (string, error) {
	base, err := GetManagedFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "front"), nil
}

// GetBackFolder returns the path to the back subfolder.
func GetBackFolder() (string, error) {
	base, err := GetManagedFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "back"), nil
}

// GetBothSubfolders returns both front and back subfolder paths.
func GetBothSubfolders() (front string, back string, err error) {
	front, err = GetFrontFolder()
	if err != nil {
		return "", "", err
	}
	back, err = GetBackFolder()
	if err != nil {
		return "", "", err
	}
	return front, back, nil
}

// GetStandardPathmanLocation returns the standard location where pathman should be installed.
func GetStandardPathmanLocation() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".local", "pathman", "bin", "pathman"), nil
}

// IsInStandardLocation checks if the given path is the standard pathman location.
func IsInStandardLocation(currentPath string) (bool, error) {
	standardPath, err := GetStandardPathmanLocation()
	if err != nil {
		return false, err
	}

	// Resolve both paths to handle symlinks.
	resolvedCurrent, err := filepath.EvalSymlinks(currentPath)
	if err != nil {
		// If we can't resolve, just compare directly.
		resolvedCurrent = currentPath
	}

	resolvedStandard, err := filepath.EvalSymlinks(standardPath)
	if err != nil {
		// If standard location doesn't exist yet, just compare directly.
		resolvedStandard = standardPath
	}

	return resolvedCurrent == resolvedStandard, nil
}

// SelfInstall installs the pathman binary to the standard location and creates a symlink.
func SelfInstall(currentPath string) error {
	standardPath, err := GetStandardPathmanLocation()
	if err != nil {
		return err
	}

	frontPath, err := GetFrontFolder()
	if err != nil {
		return err
	}

	// Create the standard location directory.
	standardDir := filepath.Dir(standardPath)
	// #nosec G301 -- 0755 permissions are appropriate for PATH directories that need to be accessible by different users
	if err := os.MkdirAll(standardDir, 0755); err != nil {
		return fmt.Errorf("failed to create standard location directory: %w", err)
	}

	// Check if there's already a pathman binary at the standard location.
	if _, err := os.Stat(standardPath); err == nil {
		// File exists - add write permission before attempting to remove it.
		// This is defensive: the file might not have write permission.
		// We ignore any error from Chmod since we might still be able to remove it.
		// #nosec G302 -- 0755 permissions are appropriate for executables
		_ = os.Chmod(standardPath, 0755)
		// Remove the existing binary.
		if err := os.Remove(standardPath); err != nil {
			return fmt.Errorf("failed to remove existing binary at %s: %w", standardPath, err)
		}
	}

	// Copy the binary to the standard location.
	if err := copyFile(currentPath, standardPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make the copied binary executable.
	// #nosec G302 -- 0755 permissions are appropriate for executables
	if err := os.Chmod(standardPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	// Create symlink in front subfolder if it doesn't already exist.
	symlinkPath := filepath.Join(frontPath, "pathman")
	if _, err := os.Lstat(symlinkPath); err == nil {
		// Symlink already exists - don't replace it.
		// This is appropriate: if the symlink exists, it's already managed.
		return nil
	} else if !os.IsNotExist(err) {
		// Some other error occurred while checking for the symlink.
		return fmt.Errorf("failed to check for existing symlink: %w", err)
	}

	// Symlink doesn't exist, create it.
	if err := os.Symlink(standardPath, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// RemoveOriginalBinary removes the original pathman binary after successful self-install.
func RemoveOriginalBinary(originalPath string) error {
	if err := os.Remove(originalPath); err != nil {
		return fmt.Errorf("failed to remove original executable at %s: %w", originalPath, err)
	}
	return nil
}

// copyFile copies a file from src to dst, preserving file mode.
func copyFile(src, dst string) error {
	// #nosec G304 -- src is validated by os.Executable and filepath.EvalSymlinks in SelfInstall caller
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// #nosec G304 -- dst is constructed from GetStandardPathmanLocation which uses os.UserHomeDir
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Get source file info to preserve permissions.
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// Exists checks if the managed folder exists.
func Exists(folderPath string) bool {
	info, err := os.Stat(folderPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Create creates the managed folder if it doesn't exist.
func Create(folderPath string) error {
	// #nosec G301 -- 0755 permissions are appropriate for PATH directories that need to be accessible by different users
	return os.MkdirAll(folderPath, 0755)
}

// checkPathMasking checks if adding a symlink will mask or be masked by other executables on PATH.
func checkPathMasking(symlinkName, targetFolder string, atFront bool) error {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil
	}

	pathDirs := filepath.SplitList(pathEnv)
	frontFolder, _ := GetFrontFolder()
	backFolder, _ := GetBackFolder()

	// Find where in PATH this symlink will be placed.
	var symlinkPosition int = -1
	for i, dir := range pathDirs {
		if (atFront && dir == frontFolder) || (!atFront && dir == backFolder) {
			symlinkPosition = i
			break
		}
	}

	// Check all PATH directories for the same executable name.
	for i, dir := range pathDirs {
		// Skip the managed folders themselves.
		if dir == frontFolder || dir == backFolder {
			continue
		}

		execPath := filepath.Join(dir, symlinkName)
		if _, err := os.Stat(execPath); err == nil {
			// Found executable with same name.
			if symlinkPosition == -1 {
				// Managed folder not in PATH, can't determine masking.
				fmt.Printf("Warning: executable '%s' exists at %s\n", symlinkName, execPath)
			} else if i < symlinkPosition {
				// Executable comes before our symlink - our symlink will be masked.
				return fmt.Errorf("symlink '%s' will be masked by existing executable at %s (use --force to add anyway)", symlinkName, execPath)
			} else {
				// Our symlink comes before executable - we will mask it.
				return fmt.Errorf("symlink '%s' will mask existing executable at %s (use --force to add anyway)", symlinkName, execPath)
			}
		}
	}

	return nil
}

// SetManagedFolder sets the managed folder path in the configuration.
// PrintSummary prints a summary of both managed folders and checks for name clashes.
func PrintSummary() error {
	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return fmt.Errorf("failed to get managed subfolder paths: %w", err)
	}

	basePath, err := GetManagedFolder()
	if err != nil {
		return fmt.Errorf("failed to get managed folder path: %w", err)
	}

	// Load managed directories.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Pathman Managed Folder:")
	fmt.Printf("  Base: %s", basePath)
	if !Exists(basePath) {
		fmt.Print(" (does not exist - run 'pathman init' to create)")
	}
	fmt.Println()

	// Count symlinks in front folder.
	frontCount := 0
	if Exists(frontPath) {
		frontLinks, err := List(true)
		if err == nil {
			frontCount = len(frontLinks)
		}
	}
	fmt.Printf("  Front subfolder: %s (%d symlinks)\n", frontPath, frontCount)

	// Count symlinks in back folder.
	backCount := 0
	if Exists(backPath) {
		backLinks, err := List(false)
		if err == nil {
			backCount = len(backLinks)
		}
	}
	fmt.Printf("  Back subfolder:  %s (%d symlinks)\n", backPath, backCount)

	// Show managed directories.
	fmt.Println()
	if len(cfg.ManagedDirectories) > 0 {
		fmt.Printf("Managed Directories (%d):\n", len(cfg.ManagedDirectories))
		for _, dir := range cfg.ManagedDirectories {
			fmt.Printf("  [%s] %s", dir.Priority, dir.Path)
			// Health check: does it exist?
			if info, err := os.Stat(dir.Path); err != nil {
				if os.IsNotExist(err) {
					fmt.Print(" (does not exist)")
				} else {
					fmt.Printf(" (error: %v)", err)
				}
			} else if !info.IsDir() {
				fmt.Print(" (not a directory)")
			}
			fmt.Println()
		}
	} else {
		fmt.Println("No managed directories.")
	}

	// Check for conflicts.
	fmt.Println()

	// Check for name clashes between front and back.
	clashes, err := CheckNameClashes()
	if err != nil {
		return fmt.Errorf("failed to check name clashes: %w", err)
	}

	// Check for PATH clashes (including managed directories).
	pathClashes, err := CheckPathClashesWithDirs()
	if err != nil {
		return fmt.Errorf("failed to check PATH clashes: %w", err)
	}

	// Report conflicts.
	if len(clashes) == 0 && len(pathClashes) == 0 {
		fmt.Println("No PATH clashes detected.")
	} else {
		if len(clashes) > 0 {
			fmt.Println("Name clashes detected (same name in both front and back):")
			for _, clash := range clashes {
				fmt.Printf("  %s\n", clash)
			}
			if len(pathClashes) > 0 {
				fmt.Println()
			}
		}

		if len(pathClashes) > 0 {
			fmt.Println("PATH clashes detected (masking or masked by other executables):")
			for _, clash := range pathClashes {
				fmt.Printf("  %s\n", clash)
			}
		}
	}

	return nil
}

// CheckNameClashes checks for executables with the same name in both subfolders.
func CheckNameClashes() ([]string, error) {
	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return nil, err
	}

	var clashes []string

	// Only check if both exist.
	if !Exists(frontPath) || !Exists(backPath) {
		return clashes, nil
	}

	// Get lists from both folders.
	frontLinks, err := List(true)
	if err != nil {
		return nil, err
	}

	backLinks, err := List(false)
	if err != nil {
		return nil, err
	}

	// Find common names.
	frontSet := make(map[string]bool)
	for _, name := range frontLinks {
		frontSet[name] = true
	}

	for _, name := range backLinks {
		if frontSet[name] {
			clashes = append(clashes, name)
		}
	}

	return clashes, nil
}

// CheckPathClashes checks if any managed symlinks mask or are masked by executables elsewhere on PATH.
func CheckPathClashes() ([]string, error) {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil, nil
	}

	pathDirs := filepath.SplitList(pathEnv)
	frontFolder, _ := GetFrontFolder()
	backFolder, _ := GetBackFolder()

	// Get all managed symlinks with their priorities.
	allSymlinks, err := ListLongBoth()
	if err != nil {
		return nil, err
	}

	var clashes []string

	for _, symlink := range allSymlinks {
		// Find where this symlink is in PATH.
		var symlinkPosition int = -1
		var symlinkFolder string

		if symlink.Priority == "front" {
			symlinkFolder = frontFolder
		} else {
			symlinkFolder = backFolder
		}

		for i, dir := range pathDirs {
			if dir == symlinkFolder {
				symlinkPosition = i
				break
			}
		}

		if symlinkPosition == -1 {
			// Managed folder not in PATH, skip checking.
			continue
		}

		// Check all PATH directories for the same executable name.
		for i, dir := range pathDirs {
			// Skip the managed folders themselves.
			if dir == frontFolder || dir == backFolder {
				continue
			}

			execPath := filepath.Join(dir, symlink.Name)
			if _, err := os.Stat(execPath); err == nil {
				// Found executable with same name.
				if i < symlinkPosition {
					// Executable comes before our symlink - our symlink is masked.
					clashes = append(clashes, fmt.Sprintf("%s (masked by %s)", symlink.Name, execPath))
				} else {
					// Our symlink comes before executable - we mask it.
					clashes = append(clashes, fmt.Sprintf("%s (masks %s)", symlink.Name, execPath))
				}
				break // Only report first clash per symlink.
			}
		}
	}

	return clashes, nil
}

// CheckPathClashesWithDirs checks if any managed symlinks or executables in managed directories
// mask or are masked by executables elsewhere on PATH.
func CheckPathClashesWithDirs() ([]string, error) {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil, nil
	}

	pathDirs := filepath.SplitList(pathEnv)
	frontFolder, _ := GetFrontFolder()
	backFolder, _ := GetBackFolder()

	// Load managed directories.
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Build set of all managed paths.
	managedPaths := make(map[string]bool)
	managedPaths[frontFolder] = true
	managedPaths[backFolder] = true
	for _, dir := range cfg.ManagedDirectories {
		managedPaths[dir.Path] = true
	}

	// Collect all executables from managed folders and directories.
	type ManagedExec struct {
		Name     string
		Path     string
		Priority string
	}
	var managedExecs []ManagedExec

	// Get symlinks from front and back.
	allSymlinks, err := ListLongBoth()
	if err != nil {
		return nil, err
	}
	for _, symlink := range allSymlinks {
		managedExecs = append(managedExecs, ManagedExec{
			Name:     symlink.Name,
			Path:     frontFolder,
			Priority: symlink.Priority,
		})
	}

	// Get executables from managed directories.
	for _, dir := range cfg.ManagedDirectories {
		if info, err := os.Stat(dir.Path); err == nil && info.IsDir() {
			entries, err := os.ReadDir(dir.Path)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						entryPath := filepath.Join(dir.Path, entry.Name())
						if info, err := os.Stat(entryPath); err == nil && info.Mode()&0111 != 0 {
							// File is executable.
							managedExecs = append(managedExecs, ManagedExec{
								Name:     entry.Name(),
								Path:     dir.Path,
								Priority: dir.Priority,
							})
						}
					}
				}
			}
		}
	}

	var clashes []string

	for _, exec := range managedExecs {
		// Find where this executable's directory is in PATH.
		var execPosition int = -1

		for i, dir := range pathDirs {
			if dir == exec.Path {
				execPosition = i
				break
			}
		}

		if execPosition == -1 {
			// Not in PATH, skip checking.
			continue
		}

		// Check all PATH directories for the same executable name.
		for i, dir := range pathDirs {
			// Skip managed paths.
			if managedPaths[dir] {
				continue
			}

			execPath := filepath.Join(dir, exec.Name)
			if _, err := os.Stat(execPath); err == nil {
				// Found executable with same name.
				if i < execPosition {
					// Executable comes before our managed one - ours is masked.
					clashes = append(clashes, fmt.Sprintf("%s (masked by %s)", exec.Name, execPath))
				} else {
					// Our managed executable comes before - we mask it.
					clashes = append(clashes, fmt.Sprintf("%s (masks %s)", exec.Name, execPath))
				}
				break // Only report first clash per executable.
			}
		}
	}

	return clashes, nil
}

// Init initializes both managed folders.
// If the folders don't exist, it creates them with appropriate permissions.
// If the folders exist, it checks permissions and warns if insecure.
// It also checks if the folders are on $PATH and offers to add them for bash users.
func Init() error {
	basePath, err := GetManagedFolder()
	if err != nil {
		return fmt.Errorf("failed to get managed folder path: %w", err)
	}

	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return fmt.Errorf("failed to get subfolder paths: %w", err)
	}

	// Check/create base folder.
	baseCreated := false
	if Exists(basePath) {
		info, err := os.Stat(basePath)
		if err != nil {
			return fmt.Errorf("failed to stat folder: %w", err)
		}

		perm := info.Mode().Perm()
		if perm&0022 != 0 {
			fmt.Printf("Managed folder already exists: %s\n", basePath)
			fmt.Printf("WARNING: Folder has insecure permissions: %04o\n", perm)
			fmt.Println("Group or others have write permission. This is a security risk.")
			fmt.Println("Recommended permissions: 0755 (owner read/write/execute, all read/execute)")
		} else {
			fmt.Printf("Managed folder already exists: %s\n", basePath)
			fmt.Printf("Permissions are correct: %04o\n", perm)
		}
	} else {
		// #nosec G301 -- 0755 permissions are appropriate for PATH directories that need to be accessible by different users
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
		fmt.Printf("Created managed folder: %s\n", basePath)
		fmt.Printf("Permissions set to: 0755 (owner read/write/execute, all read/execute)\n")
		baseCreated = true
	}

	// Create front subfolder.
	frontCreated := false
	if !Exists(frontPath) {
		// #nosec G301 -- 0755 permissions are appropriate for PATH directories that need to be accessible by different users
		if err := os.MkdirAll(frontPath, 0755); err != nil {
			return fmt.Errorf("failed to create front subfolder: %w", err)
		}
		fmt.Printf("Created front subfolder: %s\n", frontPath)
		frontCreated = true
	}

	// Create back subfolder.
	backCreated := false
	if !Exists(backPath) {
		// #nosec G301 -- 0755 permissions are appropriate for PATH directories that need to be accessible by different users
		if err := os.MkdirAll(backPath, 0755); err != nil {
			return fmt.Errorf("failed to create back subfolder: %w", err)
		}
		fmt.Printf("Created back subfolder: %s\n", backPath)
		backCreated = true
	}

	// Check if subfolders are on $PATH.
	frontOnPath := IsOnPath(frontPath)
	backOnPath := IsOnPath(backPath)

	if !frontOnPath || !backOnPath {
		fmt.Println()
		fmt.Println("The managed subfolders are not properly configured in your $PATH.")
		fmt.Println("To use executables in these folders, you need to add them to your $PATH.")

		// Check if the user is using bash.
		shell := os.Getenv("SHELL")
		if strings.Contains(shell, "bash") {
			fmt.Println()
			profilePath, err := GetBashProfilePath()
			if err != nil {
				return fmt.Errorf("failed to get profile path: %w", err)
			}

			profileName := filepath.Base(profilePath)
			fmt.Printf("Since you're using bash, this is normally done by adding a line to your ~/%s file.\n", profileName)

			if answer, err := PromptUser("Would you like me to add the PATH configuration for you?"); err != nil {
				return fmt.Errorf("failed to read user input: %w", err)
			} else if answer {
				if err := AddToProfile(); err != nil {
					return fmt.Errorf("failed to add to profile: %w", err)
				}
			} else {
				fmt.Printf("\nTo add it manually, add these lines to your ~/%s:\n", profileName)
				fmt.Println("  # Added by pathman")
				for _, line := range GetShellIntegrationScript() {
					fmt.Printf("  %s\n", line)
				}
			}
		} else {
			fmt.Println("\nTo add it to your PATH, add these lines to your shell configuration:")
			fmt.Println("  # Added by pathman")
			for _, line := range GetShellIntegrationScript() {
				fmt.Printf("  %s\n", line)
			}
		}
	} else if baseCreated || frontCreated || backCreated {
		fmt.Println()
		fmt.Println("The managed folder is already properly configured in your $PATH.")
	}

	return nil
}

// IsOnPath checks if the given folder path is on the $PATH.
func IsOnPath(folderPath string) bool {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return false
	}

	// Clean the folder path for comparison.
	cleanFolderPath := filepath.Clean(folderPath)

	// Split PATH by colon and check each entry.
	pathEntries := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, entry := range pathEntries {
		if filepath.Clean(entry) == cleanFolderPath {
			return true
		}
	}

	return false
}

// GetAdjustedPath returns the PATH with the managed folder added if not already present.
// If atFront is true, adds to the front; otherwise adds to the back.
func GetAdjustedPath() (string, error) {
	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return "", fmt.Errorf("failed to get subfolder paths: %w", err)
	}

	// Load managed directories from config.
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Separate directories by priority.
	var frontDirs []string
	var backDirs []string
	for _, dir := range cfg.ManagedDirectories {
		if dir.Priority == "front" {
			frontDirs = append(frontDirs, dir.Path)
		} else {
			backDirs = append(backDirs, dir.Path)
		}
	}

	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		// Empty PATH: add managed folders and directories.
		parts := []string{frontPath}
		parts = append(parts, frontDirs...)
		parts = append(parts, backDirs...)
		parts = append(parts, backPath)
		return strings.Join(parts, string(os.PathListSeparator)), nil
	}

	// Build set of all managed paths to remove.
	managedPaths := make(map[string]bool)
	managedPaths[frontPath] = true
	managedPaths[backPath] = true
	for _, dir := range cfg.ManagedDirectories {
		managedPaths[dir.Path] = true
	}

	// Remove any existing occurrences of managed paths from PATH.
	pathParts := strings.Split(pathEnv, string(os.PathListSeparator))
	var cleanedParts []string
	for _, part := range pathParts {
		if !managedPaths[part] {
			cleanedParts = append(cleanedParts, part)
		}
	}

	// Build new PATH: front subfolder + front dirs + cleaned parts + back dirs + back subfolder.
	var newPathParts []string
	newPathParts = append(newPathParts, frontPath)
	newPathParts = append(newPathParts, frontDirs...)
	if len(cleanedParts) > 0 {
		newPathParts = append(newPathParts, cleanedParts...)
	}
	newPathParts = append(newPathParts, backDirs...)
	newPathParts = append(newPathParts, backPath)

	return strings.Join(newPathParts, string(os.PathListSeparator)), nil
}

// GetBashProfilePath determines which bash profile file to use.
// Returns the path to .bash_profile if it exists, otherwise .profile.
func GetBashProfilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	bashProfile := filepath.Join(homeDir, ".bash_profile")
	if _, err := os.Stat(bashProfile); err == nil {
		return bashProfile, nil
	}

	return filepath.Join(homeDir, ".profile"), nil
}

// AddToProfile adds the managed folder to the user's bash profile.
func AddToProfile() error {
	profilePath, err := GetBashProfilePath()
	if err != nil {
		return fmt.Errorf("failed to get profile path: %w", err)
	}

	// Check if the export line already exists.
	if hasPathExport, err := profileHasPathmanExport(profilePath); err != nil {
		return err
	} else if hasPathExport {
		fmt.Printf("PATH export already exists in %s\n", profilePath)
		return nil
	}

	// Open the file for appending.
	// #nosec G302,G304 -- 0644 permissions are standard for shell profile files; profilePath comes from GetBashProfilePath which returns user's home directory paths
	f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open profile file: %w", err)
	}
	defer f.Close()

	// Add a newline if the file doesn't end with one.
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat profile file: %w", err)
	}

	if info.Size() > 0 {
		// Check if file ends with newline.
		// #nosec G304 -- profilePath comes from GetBashProfilePath which returns user's home directory paths
		content, err := os.ReadFile(profilePath)
		if err != nil {
			return fmt.Errorf("failed to read profile file: %w", err)
		}
		if len(content) > 0 && content[len(content)-1] != '\n' {
			if _, err := f.WriteString("\n"); err != nil {
				return fmt.Errorf("failed to write newline: %w", err)
			}
		}
	}

	// Add the export line using pathman path.
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n# Added by 'pathman init' on %s\n", timestamp))
	for _, line := range GetShellIntegrationScript() {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	exportLine := sb.String()
	if _, err := f.WriteString(exportLine); err != nil {
		return fmt.Errorf("failed to write to profile: %w", err)
	}

	fmt.Printf("Added PATH export to %s\n", profilePath)
	fmt.Println("Please restart your shell or run: source", profilePath)
	return nil
}

// profileHasPathmanExport checks if the profile already has a pathman export.
func profileHasPathmanExport(profilePath string) (bool, error) {
	// #nosec G304 -- profilePath comes from GetBashProfilePath which returns user's home directory paths
	f, err := os.Open(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Check if the line exports PATH and uses pathman path.
		if strings.Contains(line, "export") && strings.Contains(line, "PATH") && strings.Contains(line, "pathman path") {
			return true, nil
		}
	}

	return false, scanner.Err()
}

// PromptUser prompts the user with a yes/no question and returns true if they answer yes.
func PromptUser(question string) (bool, error) {
	fmt.Printf("%s (y/n): ", question)

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, nil
	}

	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}

// List returns a list of all symlinks in the managed folder.
func List(atFront bool) ([]string, error) {
	var folderPath string
	var err error

	if atFront {
		folderPath, err = GetFrontFolder()
	} else {
		folderPath, err = GetBackFolder()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get subfolder path: %w", err)
	}

	if !Exists(folderPath) {
		return nil, fmt.Errorf("subfolder does not exist: %s", folderPath)
	}

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read subfolder: %w", err)
	}

	var symlinks []string
	for _, entry := range entries {
		entryPath := filepath.Join(folderPath, entry.Name())
		info, err := os.Lstat(entryPath)
		if err != nil {
			continue
		}
		// Only include symlinks.
		if info.Mode()&os.ModeSymlink != 0 {
			symlinks = append(symlinks, entry.Name())
		}
	}

	return symlinks, nil
}

// ListBoth returns all symlink names from both front and back folders.
func ListBoth() ([]string, error) {
	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return nil, fmt.Errorf("failed to get subfolder paths: %w", err)
	}

	seenNames := make(map[string]bool)
	var allSymlinks []string

	// List front folder first.
	if Exists(frontPath) {
		entries, err := os.ReadDir(frontPath)
		if err == nil {
			for _, entry := range entries {
				entryPath := filepath.Join(frontPath, entry.Name())
				info, err := os.Lstat(entryPath)
				if err == nil && info.Mode()&os.ModeSymlink != 0 {
					if !seenNames[entry.Name()] {
						allSymlinks = append(allSymlinks, entry.Name())
						seenNames[entry.Name()] = true
					}
				}
			}
		}
	}

	// List back folder.
	if Exists(backPath) {
		entries, err := os.ReadDir(backPath)
		if err == nil {
			for _, entry := range entries {
				entryPath := filepath.Join(backPath, entry.Name())
				info, err := os.Lstat(entryPath)
				if err == nil && info.Mode()&os.ModeSymlink != 0 {
					if !seenNames[entry.Name()] {
						allSymlinks = append(allSymlinks, entry.Name())
						seenNames[entry.Name()] = true
					}
				}
			}
		}
	}

	return allSymlinks, nil
}

// SymlinkInfo represents information about a symlink.
type SymlinkInfo struct {
	Name     string
	Target   string
	Priority string // "front" or "back"
}

// ListLong returns detailed information about all symlinks in the managed subfolder.
func ListLong(atFront bool) ([]SymlinkInfo, error) {
	var folderPath string
	var err error

	if atFront {
		folderPath, err = GetFrontFolder()
	} else {
		folderPath, err = GetBackFolder()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get subfolder path: %w", err)
	}

	if !Exists(folderPath) {
		return nil, fmt.Errorf("subfolder does not exist: %s", folderPath)
	}

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read subfolder: %w", err)
	}

	var symlinks []SymlinkInfo
	for _, entry := range entries {
		entryPath := filepath.Join(folderPath, entry.Name())
		info, err := os.Lstat(entryPath)
		if err != nil {
			continue
		}
		// Only include symlinks.
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(entryPath)
			if err != nil {
				target = "<error reading link>"
			}
			priority := "back"
			if atFront {
				priority = "front"
			}
			symlinks = append(symlinks, SymlinkInfo{
				Name:     entry.Name(),
				Target:   target,
				Priority: priority,
			})
		}
	}

	return symlinks, nil
}

// ListLongBoth returns detailed information about all symlinks from both front and back folders.
func ListLongBoth() ([]SymlinkInfo, error) {
	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return nil, fmt.Errorf("failed to get subfolder paths: %w", err)
	}

	var allSymlinks []SymlinkInfo

	// List front folder first.
	if Exists(frontPath) {
		entries, err := os.ReadDir(frontPath)
		if err == nil {
			for _, entry := range entries {
				entryPath := filepath.Join(frontPath, entry.Name())
				info, err := os.Lstat(entryPath)
				if err == nil && info.Mode()&os.ModeSymlink != 0 {
					target, err := os.Readlink(entryPath)
					if err != nil {
						target = "<error reading link>"
					}
					allSymlinks = append(allSymlinks, SymlinkInfo{
						Name:     entry.Name(),
						Target:   target,
						Priority: "front",
					})
				}
			}
		}
	}

	// List back folder.
	if Exists(backPath) {
		entries, err := os.ReadDir(backPath)
		if err == nil {
			for _, entry := range entries {
				entryPath := filepath.Join(backPath, entry.Name())
				info, err := os.Lstat(entryPath)
				if err == nil && info.Mode()&os.ModeSymlink != 0 {
					target, err := os.Readlink(entryPath)
					if err != nil {
						target = "<error reading link>"
					}
					allSymlinks = append(allSymlinks, SymlinkInfo{
						Name:     entry.Name(),
						Target:   target,
						Priority: "back",
					})
				}
			}
		}
	}

	return allSymlinks, nil
}

// DirInfo represents information about a managed directory.
type DirInfo struct {
	Path     string
	Priority string
}

// ListBothWithDirs returns symlink names and managed directories.
func ListBothWithDirs() ([]string, []DirInfo, error) {
	symlinks, err := ListBoth()
	if err != nil {
		return nil, nil, err
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	var dirs []DirInfo
	for _, dir := range cfg.ManagedDirectories {
		dirs = append(dirs, DirInfo{
			Path:     dir.Path,
			Priority: dir.Priority,
		})
	}

	return symlinks, dirs, nil
}

// ListLongBothWithDirs returns detailed symlink information and managed directories.
func ListLongBothWithDirs() ([]SymlinkInfo, []DirInfo, error) {
	symlinks, err := ListLongBoth()
	if err != nil {
		return nil, nil, err
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	var dirs []DirInfo
	for _, dir := range cfg.ManagedDirectories {
		dirs = append(dirs, DirInfo{
			Path:     dir.Path,
			Priority: dir.Priority,
		})
	}

	return symlinks, dirs, nil
}

// Add creates a symlink to the executable in the managed subfolder.
// If a symlink with the same name exists in the other subfolder, it's moved to the specified subfolder.
func Add(executablePath, name string, atFront bool, force bool) error {
	// Get absolute path first.
	absPath, err := filepath.Abs(executablePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if the path exists.
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// If it's a directory, add to config.
	if info.IsDir() {
		return addDirectory(absPath, atFront)
	}

	// Otherwise, add as symlink (existing behavior).
	return addSymlink(absPath, name, atFront, force)
}

// addDirectory adds a directory to the managed directories in config.
func addDirectory(absPath string, atFront bool) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	priority := "back"
	if atFront {
		priority = "front"
	}

	// Check if directory is already managed.
	for i, dir := range cfg.ManagedDirectories {
		if dir.Path == absPath {
			if dir.Priority == priority {
				fmt.Printf("Directory already managed with priority '%s': %s\n", priority, absPath)
				return nil
			}
			// Update priority.
			cfg.ManagedDirectories[i].Priority = priority
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("Updated directory priority to '%s': %s\n", priority, absPath)
			return nil
		}
	}

	// Add new directory.
	cfg.ManagedDirectories = append(cfg.ManagedDirectories, config.ManagedDirectory{
		Path:     absPath,
		Priority: priority,
	})

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Added directory (%s): %s\n", priority, absPath)
	return nil
}

// addSymlink adds a file as a symlink (original Add behavior).
func addSymlink(absExecutablePath, name string, atFront bool, force bool) error {
	var folderPath, otherFolderPath string
	var err error

	if atFront {
		folderPath, err = GetFrontFolder()
		if err != nil {
			return fmt.Errorf("failed to get front subfolder path: %w", err)
		}
		otherFolderPath, _ = GetBackFolder()
	} else {
		folderPath, err = GetBackFolder()
		if err != nil {
			return fmt.Errorf("failed to get back subfolder path: %w", err)
		}
		otherFolderPath, _ = GetFrontFolder()
	}

	if !Exists(folderPath) {
		return fmt.Errorf("subfolder does not exist: %s\nRun 'pathman init' to create it", folderPath)
	}

	// Determine the symlink name.
	symlinkName := name
	if symlinkName == "" {
		symlinkName = filepath.Base(absExecutablePath)
	}

	symlinkPath := filepath.Join(folderPath, symlinkName)

	// Check if symlink already exists in the target subfolder.
	if _, err := os.Lstat(symlinkPath); err == nil {
		if !force {
			return fmt.Errorf("symlink already exists: %s (use --force to overwrite)", symlinkName)
		}
		// Remove existing symlink when force is used.
		if err := os.Remove(symlinkPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	// Check for PATH masking issues (only if not forcing).
	if !force {
		if err := checkPathMasking(symlinkName, folderPath, atFront); err != nil {
			return err
		}
	}

	// Check if symlink exists in the other subfolder and remove it if so.
	if Exists(otherFolderPath) {
		otherSymlinkPath := filepath.Join(otherFolderPath, symlinkName)
		if _, err := os.Lstat(otherSymlinkPath); err == nil {
			// Symlink exists in other subfolder, remove it.
			if err := os.Remove(otherSymlinkPath); err != nil {
				return fmt.Errorf("failed to remove symlink from other subfolder: %w", err)
			}
			fromLabel := map[bool]string{true: "front", false: "back"}[!atFront]
			toLabel := map[bool]string{true: "front", false: "back"}[atFront]
			fmt.Printf("Moved '%s' from %s to %s\n", symlinkName, fromLabel, toLabel)
		}
	}

	// Create the symlink.
	if err := os.Symlink(absExecutablePath, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	folderLabel := map[bool]string{true: "front", false: "back"}[atFront]
	fmt.Printf("Added '%s' -> '%s' (%s)\n", symlinkName, absExecutablePath, folderLabel)
	return nil
}

// Remove removes a symlink from the managed subfolders (searches both front and back).
func Remove(name string) error {
	// First, try to remove as a symlink.
	if err := removeSymlink(name); err == nil {
		return nil
	}

	// If not found as symlink, try to remove as a managed directory.
	absPath, err := filepath.Abs(name)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	return removeDirectory(absPath)
}

// removeSymlink removes a symlink from the managed subfolders.
func removeSymlink(name string) error {
	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return fmt.Errorf("failed to get subfolder paths: %w", err)
	}

	// Try front subfolder first.
	if Exists(frontPath) {
		symlinkPath := filepath.Join(frontPath, name)
		if info, err := os.Lstat(symlinkPath); err == nil {
			// Make sure it's a symlink.
			if info.Mode()&os.ModeSymlink == 0 {
				return fmt.Errorf("'%s' is not a symlink", name)
			}
			// Remove the symlink.
			if err := os.Remove(symlinkPath); err != nil {
				return fmt.Errorf("failed to remove symlink: %w", err)
			}
			fmt.Printf("Removed '%s' (from front)\n", name)
			return nil
		}
	}

	// Try back subfolder.
	if Exists(backPath) {
		symlinkPath := filepath.Join(backPath, name)
		if info, err := os.Lstat(symlinkPath); err == nil {
			// Make sure it's a symlink.
			if info.Mode()&os.ModeSymlink == 0 {
				return fmt.Errorf("'%s' is not a symlink", name)
			}
			// Remove the symlink.
			if err := os.Remove(symlinkPath); err != nil {
				return fmt.Errorf("failed to remove symlink: %w", err)
			}
			fmt.Printf("Removed '%s' (from back)\n", name)
			return nil
		}
	}

	return fmt.Errorf("symlink does not exist: %s", name)
}

// removeDirectory removes a directory from the managed directories in config.
func removeDirectory(absPath string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find and remove the directory.
	for i, dir := range cfg.ManagedDirectories {
		if dir.Path == absPath {
			cfg.ManagedDirectories = append(cfg.ManagedDirectories[:i], cfg.ManagedDirectories[i+1:]...)
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("Removed directory: %s\n", absPath)
			return nil
		}
	}

	return fmt.Errorf("not found as symlink or managed directory: %s", absPath)
}

// Rename renames a symlink in the managed subfolders (searches both front and back).
func Rename(oldName, newName string) error {
	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return fmt.Errorf("failed to get subfolder paths: %w", err)
	}

	// Try front subfolder first.
	if Exists(frontPath) {
		oldSymlinkPath := filepath.Join(frontPath, oldName)
		if info, err := os.Lstat(oldSymlinkPath); err == nil {
			// Make sure it's a symlink.
			if info.Mode()&os.ModeSymlink == 0 {
				return fmt.Errorf("'%s' is not a symlink", oldName)
			}

			// Check if new name already exists.
			newSymlinkPath := filepath.Join(frontPath, newName)
			if _, err := os.Lstat(newSymlinkPath); err == nil {
				return fmt.Errorf("symlink already exists: %s", newName)
			}

			// Rename the symlink.
			if err := os.Rename(oldSymlinkPath, newSymlinkPath); err != nil {
				return fmt.Errorf("failed to rename symlink: %w", err)
			}
			fmt.Printf("Renamed '%s' to '%s' (in front)\n", oldName, newName)
			return nil
		}
	}

	// Try back subfolder.
	if Exists(backPath) {
		oldSymlinkPath := filepath.Join(backPath, oldName)
		if info, err := os.Lstat(oldSymlinkPath); err == nil {
			// Make sure it's a symlink.
			if info.Mode()&os.ModeSymlink == 0 {
				return fmt.Errorf("'%s' is not a symlink", oldName)
			}

			// Check if new name already exists.
			newSymlinkPath := filepath.Join(backPath, newName)
			if _, err := os.Lstat(newSymlinkPath); err == nil {
				return fmt.Errorf("symlink already exists: %s", newName)
			}

			// Rename the symlink.
			if err := os.Rename(oldSymlinkPath, newSymlinkPath); err != nil {
				return fmt.Errorf("failed to rename symlink: %w", err)
			}
			fmt.Printf("Renamed '%s' to '%s' (in back)\n", oldName, newName)
			return nil
		}
	}

	return fmt.Errorf("symlink does not exist: %s", oldName)
}

// ShowPriority displays which folder (front or back) a symlink is in.
func ShowPriority(name string) error {
	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return fmt.Errorf("failed to get subfolder paths: %w", err)
	}

	// Check front folder.
	if Exists(frontPath) {
		symlinkPath := filepath.Join(frontPath, name)
		if _, err := os.Lstat(symlinkPath); err == nil {
			fmt.Printf("%s: front\n", name)
			return nil
		}
	}

	// Check back folder.
	if Exists(backPath) {
		symlinkPath := filepath.Join(backPath, name)
		if _, err := os.Lstat(symlinkPath); err == nil {
			fmt.Printf("%s: back\n", name)
			return nil
		}
	}

	return fmt.Errorf("symlink '%s' not found in either folder", name)
}

// SetPriority moves a symlink between front and back folders.
func SetPriority(name string, toFront bool) error {
	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return fmt.Errorf("failed to get subfolder paths: %w", err)
	}

	var fromPath, toPath string
	var fromLabel, toLabel string

	if toFront {
		fromPath = backPath
		toPath = frontPath
		fromLabel = "back"
		toLabel = "front"
	} else {
		fromPath = frontPath
		toPath = backPath
		fromLabel = "front"
		toLabel = "back"
	}

	// Check if symlink exists in source folder.
	if !Exists(fromPath) {
		return fmt.Errorf("%s folder does not exist", fromLabel)
	}

	fromSymlinkPath := filepath.Join(fromPath, name)
	info, err := os.Lstat(fromSymlinkPath)
	if err != nil {
		return fmt.Errorf("symlink '%s' not found in %s folder", name, fromLabel)
	}

	// Verify it's a symlink.
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("'%s' is not a symlink", name)
	}

	// Read the target.
	target, err := os.Readlink(fromSymlinkPath)
	if err != nil {
		return fmt.Errorf("failed to read symlink target: %w", err)
	}

	// Create destination folder if it doesn't exist.
	if !Exists(toPath) {
		if err := Create(toPath); err != nil {
			return fmt.Errorf("failed to create %s folder: %w", toLabel, err)
		}
	}

	toSymlinkPath := filepath.Join(toPath, name)

	// Check if symlink already exists in destination.
	if _, err := os.Lstat(toSymlinkPath); err == nil {
		return fmt.Errorf("symlink '%s' already exists in %s folder", name, toLabel)
	}

	// Create new symlink in destination.
	if err := os.Symlink(target, toSymlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink in %s folder: %w", toLabel, err)
	}

	// Remove old symlink.
	if err := os.Remove(fromSymlinkPath); err != nil {
		// Try to clean up the new symlink.
		// #nosec G104 -- best-effort cleanup in error path, main error is more important
		os.Remove(toSymlinkPath)
		return fmt.Errorf("failed to remove symlink from %s folder: %w", fromLabel, err)
	}

	fmt.Printf("Moved '%s' from %s to %s\n", name, fromLabel, toLabel)
	return nil
}

// ListEntry represents a single entry (file or directory) in the list output.
type ListEntry struct {
	Type     string // "file" or "directory"
	Name     string // For files: symlink name. For directories: empty (use Path instead).
	Path     string // For directories: full path. For files: empty.
	Symlink  string // For files: symlink target.
	Priority string // "front" or "back"
}

// GetAllEntries retrieves all files and directories, optionally filtered.
func GetAllEntries(priorityFilter, typeFilter, nameFilter string) ([]ListEntry, error) {
	var entries []ListEntry

	// Get symlinks from both folders.
	if typeFilter == "" || typeFilter == "file" {
		frontPath, backPath, err := GetBothSubfolders()
		if err != nil {
			return nil, fmt.Errorf("failed to get subfolder paths: %w", err)
		}

		// Process front folder if not filtered to back only.
		if priorityFilter == "" || priorityFilter == "front" {
			if Exists(frontPath) {
				frontEntries, err := os.ReadDir(frontPath)
				if err == nil {
					for _, entry := range frontEntries {
						entryPath := filepath.Join(frontPath, entry.Name())
						info, err := os.Lstat(entryPath)
						if err == nil && info.Mode()&os.ModeSymlink != 0 {
							// Apply name filter.
							if nameFilter != "" && entry.Name() != nameFilter {
								continue
							}

							target, err := os.Readlink(entryPath)
							if err != nil {
								target = "<error reading link>"
							}
							entries = append(entries, ListEntry{
								Type:     "file",
								Name:     entry.Name(),
								Symlink:  target,
								Priority: "front",
							})
						}
					}
				}
			}
		}

		// Process back folder if not filtered to front only.
		if priorityFilter == "" || priorityFilter == "back" {
			if Exists(backPath) {
				backEntries, err := os.ReadDir(backPath)
				if err == nil {
					for _, entry := range backEntries {
						entryPath := filepath.Join(backPath, entry.Name())
						info, err := os.Lstat(entryPath)
						if err == nil && info.Mode()&os.ModeSymlink != 0 {
							// Apply name filter.
							if nameFilter != "" && entry.Name() != nameFilter {
								continue
							}

							target, err := os.Readlink(entryPath)
							if err != nil {
								target = "<error reading link>"
							}
							entries = append(entries, ListEntry{
								Type:     "file",
								Name:     entry.Name(),
								Symlink:  target,
								Priority: "back",
							})
						}
					}
				}
			}
		}
	}

	// Get directories from config.
	if typeFilter == "" || typeFilter == "directory" {
		cfg, err := config.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}

		for _, dir := range cfg.ManagedDirectories {
			// Apply priority filter.
			if priorityFilter != "" && dir.Priority != priorityFilter {
				continue
			}
			// Apply name filter (match against base name of directory path).
			if nameFilter != "" && filepath.Base(dir.Path) != nameFilter {
				continue
			}

			entries = append(entries, ListEntry{
				Type:     "directory",
				Path:     dir.Path,
				Priority: dir.Priority,
			})
		}
	}

	return entries, nil
}

// ListCompactFormat lists entries in compact format (names only).
func ListCompactFormat(priorityFilter, typeFilter, nameFilter string, byPriority bool) error {
	entries, err := GetAllEntries(priorityFilter, typeFilter, nameFilter)
	if err != nil {
		return err
	}

	if byPriority {
		// Sort by priority (front first), then alphabetically.
		var frontEntries []string
		var backEntries []string

		for _, entry := range entries {
			var name string
			if entry.Type == "file" {
				name = entry.Name
			} else {
				name = entry.Path
			}

			if entry.Priority == "front" {
				frontEntries = append(frontEntries, name)
			} else {
				backEntries = append(backEntries, name)
			}
		}

		// Sort each group alphabetically.
		sortStrings(frontEntries)
		sortStrings(backEntries)

		// Print front first, then back.
		for _, name := range frontEntries {
			fmt.Println(name)
		}
		for _, name := range backEntries {
			fmt.Println(name)
		}
	} else {
		// Partition by type and sort alphabetically within each partition.
		var files []string
		var directories []string

		for _, entry := range entries {
			if entry.Type == "file" {
				files = append(files, entry.Name)
			} else {
				directories = append(directories, entry.Path)
			}
		}

		// Sort each group alphabetically.
		sortStrings(files)
		sortStrings(directories)

		// Print files first, then directories.
		for _, name := range files {
			fmt.Println(name)
		}
		for _, path := range directories {
			fmt.Println(path)
		}
	}

	return nil
}

// ListLongFormat lists entries in long format with labels.
func ListLongFormat(priorityFilter, typeFilter, nameFilter string, byPriority bool) error {
	entries, err := GetAllEntries(priorityFilter, typeFilter, nameFilter)
	if err != nil {
		return err
	}

	// Sort entries based on flag.
	if byPriority {
		sortEntriesByPriority(entries)
	} else {
		sortEntriesByType(entries)
	}

	first := true
	for _, entry := range entries {
		// Add blank line between entries (but not before first entry).
		if !first {
			fmt.Println()
		}
		first = false

		if entry.Type == "file" {
			fmt.Printf("%-13s %s\n", "File:", entry.Name)
			fmt.Printf("%-13s %s\n", "Symlink:", entry.Symlink)
			fmt.Printf("%-13s %s\n", "Priority:", entry.Priority)
		} else {
			fmt.Printf("%-13s %s\n", "Directory:", entry.Path)
			fmt.Printf("%-13s %s\n", "Priority:", entry.Priority)
		}
	}

	return nil
}

// FileEntry represents a file entry for JSON output.
type FileEntry struct {
	File     string `json:"file"`
	Symlink  string `json:"symlink"`
	Priority string `json:"priority"`
}

// DirEntry represents a directory entry for JSON output.
type DirEntry struct {
	Directory string `json:"directory"`
	Priority  string `json:"priority"`
}

// ListJSON lists entries in JSON format.
func ListJSON(priorityFilter, typeFilter, nameFilter string) error {
	entries, err := GetAllEntries(priorityFilter, typeFilter, nameFilter)
	if err != nil {
		return err
	}

	// Separate files and directories.

	output := struct {
		Files       []FileEntry `json:"files"`
		Directories []DirEntry  `json:"directories"`
	}{
		Files:       []FileEntry{},
		Directories: []DirEntry{},
	}

	// Collect entries by type.
	var files []FileEntry
	var dirs []DirEntry

	for _, entry := range entries {
		if entry.Type == "file" {
			files = append(files, FileEntry{
				File:     entry.Name,
				Symlink:  entry.Symlink,
				Priority: entry.Priority,
			})
		} else {
			dirs = append(dirs, DirEntry{
				Directory: entry.Path,
				Priority:  entry.Priority,
			})
		}
	}

	// Sort files alphabetically by name.
	sortFileEntries(files)
	// Sort directories alphabetically by path.
	sortDirEntries(dirs)

	output.Files = files
	output.Directories = dirs

	// Pretty-print JSON.
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// sortStrings sorts a slice of strings alphabetically in-place.
func sortStrings(items []string) {
	sort.Strings(items)
}

// sortEntriesByType sorts entries by type (files first, then directories), then alphabetically within each type.
func sortEntriesByType(entries []ListEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		// Files before directories.
		if entries[i].Type != entries[j].Type {
			return entries[i].Type == "file"
		}

		// Within same type, sort alphabetically.
		if entries[i].Type == "file" {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].Path < entries[j].Path
	})
}

// sortEntriesByPriority sorts entries by priority (front first, then back), then alphabetically within each priority.
func sortEntriesByPriority(entries []ListEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		// Front before back.
		if entries[i].Priority != entries[j].Priority {
			return entries[i].Priority == "front"
		}

		// Within same priority, sort alphabetically.
		if entries[i].Type == "file" {
			if entries[j].Type == "file" {
				return entries[i].Name < entries[j].Name
			}
			// If i is file and j is directory, compare by name vs path.
			return entries[i].Name < entries[j].Path
		}
		if entries[j].Type == "file" {
			// If i is directory and j is file.
			return entries[i].Path < entries[j].Name
		}
		// Both are directories.
		return entries[i].Path < entries[j].Path
	})
}

// sortFileEntries sorts file entries alphabetically by name.
func sortFileEntries(files []FileEntry) {
	sort.SliceStable(files, func(i, j int) bool {
		return files[i].File < files[j].File
	})
}

// sortDirEntries sorts directory entries alphabetically by path.
func sortDirEntries(dirs []DirEntry) {
	sort.SliceStable(dirs, func(i, j int) bool {
		return dirs[i].Directory < dirs[j].Directory
	})
}
