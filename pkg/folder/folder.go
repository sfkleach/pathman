package folder

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sfkleach/pathman/pkg/config"
)

// GetManagedFolder returns the path to the managed folder.
// It first checks the configuration file, then falls back to the default.
// atFront determines which folder to return (true for front, false for back).
func GetManagedFolder(atFront bool) (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	if cfg != nil {
		if atFront && cfg.ManagedFolderFront != "" {
			return cfg.ManagedFolderFront, nil
		}
		if !atFront && cfg.ManagedFolderBack != "" {
			return cfg.ManagedFolderBack, nil
		}
	}

	if atFront {
		return config.GetDefaultManagedFolderFront()
	}
	return config.GetDefaultManagedFolderBack()
}

// GetBothManagedFolders returns both front and back folder paths.
func GetBothManagedFolders() (front string, back string, err error) {
	front, err = GetManagedFolder(true)
	if err != nil {
		return "", "", err
	}
	back, err = GetManagedFolder(false)
	if err != nil {
		return "", "", err
	}
	return front, back, nil
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
	return os.MkdirAll(folderPath, 0755)
}

// SetManagedFolder sets the managed folder path in the configuration.
// atFront determines which folder to set (true for front, false for back).
func SetManagedFolder(folderPath string, atFront bool) error {
	// Load existing config or create new one.
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	// Set the appropriate folder.
	if atFront {
		cfg.ManagedFolderFront = folderPath
	} else {
		cfg.ManagedFolderBack = folderPath
	}

	return cfg.Save()
}

// PrintSummary prints a summary of both managed folders and checks for name clashes.
func PrintSummary() error {
	frontPath, backPath, err := GetBothManagedFolders()
	if err != nil {
		return fmt.Errorf("failed to get managed folder paths: %w", err)
	}

	fmt.Println("Pathman Managed Folders:")
	fmt.Println()

	// Front folder status.
	fmt.Printf("  Front folder: %s\n", frontPath)
	if Exists(frontPath) {
		fmt.Println("    Status: exists")
	} else {
		fmt.Println("    Status: does not exist")
	}

	// Back folder status.
	fmt.Printf("  Back folder:  %s\n", backPath)
	if Exists(backPath) {
		fmt.Println("    Status: exists")
	} else {
		fmt.Println("    Status: does not exist")
	}

	// Check if they exist before checking clashes.
	if !Exists(frontPath) && !Exists(backPath) {
		fmt.Println()
		fmt.Println("Run 'pathman init' to create the managed folders.")
		return nil
	}

	// Check for name clashes.
	clashes, err := CheckNameClashes()
	if err != nil {
		return fmt.Errorf("failed to check name clashes: %w", err)
	}

	if len(clashes) > 0 {
		fmt.Println()
		fmt.Println("Name clashes detected:")
		for _, clash := range clashes {
			fmt.Printf("  %s\n", clash)
		}
	}

	return nil
}

// CheckNameClashes checks for executables with the same name in both folders.
func CheckNameClashes() ([]string, error) {
	frontPath, backPath, err := GetBothManagedFolders()
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

// Init initializes both managed folders.
// If the folders don't exist, it creates them with appropriate permissions.
// If the folders exist, it checks permissions and warns if insecure.
// It also checks if the folders are on $PATH and offers to add them for bash users.
func Init() error {
	frontPath, backPath, err := GetBothManagedFolders()
	if err != nil {
		return fmt.Errorf("failed to get managed folder paths: %w", err)
	}

	// Initialize front folder.
	frontCreated, err := initFolder(frontPath, "front")
	if err != nil {
		return err
	}

	// Initialize back folder.
	backCreated, err := initFolder(backPath, "back")
	if err != nil {
		return err
	}

	// Check if folders are on $PATH.
	frontOnPath := IsOnPath(frontPath)
	backOnPath := IsOnPath(backPath)

	if !frontOnPath || !backOnPath {
		fmt.Println()
		fmt.Println("The managed folders are not properly configured in your $PATH.")
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
				fmt.Printf("\nTo add it manually, add this line to your ~/%s:\n", profileName)
				fmt.Printf("  export PATH=$(pathman path)\n")
			}
		} else {
			fmt.Println("\nTo add it to your PATH, add this line to your shell configuration:")
			fmt.Println("  export PATH=$(pathman path)")
		}
	} else if frontCreated || backCreated {
		fmt.Println()
		fmt.Println("The managed folders are already properly configured in your $PATH.")
	}

	return nil
}

// initFolder initializes a single folder (helper for Init).
func initFolder(folderPath, label string) (bool, error) {
	folderCreated := false

	if Exists(folderPath) {
		// Folder exists, check permissions.
		info, err := os.Stat(folderPath)
		if err != nil {
			return false, fmt.Errorf("failed to stat folder: %w", err)
		}

		mode := info.Mode()
		perm := mode.Perm()

		// Check if group or others have write permission.
		if perm&0022 != 0 {
			fmt.Printf("Managed folder (%s) already exists: %s\n", label, folderPath)
			fmt.Printf("WARNING: Folder has insecure permissions: %04o\n", perm)
			fmt.Println("Group or others have write permission. This is a security risk.")
			fmt.Println("Recommended permissions: 0755 (owner read/write/execute, all read/execute)")
		} else {
			fmt.Printf("Managed folder (%s) already exists: %s\n", label, folderPath)
			fmt.Printf("Permissions are correct: %04o\n", perm)
		}
	} else {
		// Create the folder with permissions: owner read+write+execute, all read+execute (0755).
		if err := os.MkdirAll(folderPath, 0755); err != nil {
			return false, fmt.Errorf("failed to create folder: %w", err)
		}

		fmt.Printf("Created managed folder (%s): %s\n", label, folderPath)
		fmt.Printf("Permissions set to: 0755 (owner read/write/execute, all read/execute)\n")
		folderCreated = true
	}

	return folderCreated, nil
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
	frontPath, backPath, err := GetBothManagedFolders()
	if err != nil {
		return "", fmt.Errorf("failed to get managed folder paths: %w", err)
	}

	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		// Empty PATH: just add both folders.
		return frontPath + string(os.PathListSeparator) + backPath, nil
	}

	// Remove any existing occurrences of both folders from PATH.
	pathParts := strings.Split(pathEnv, string(os.PathListSeparator))
	var cleanedParts []string
	for _, part := range pathParts {
		if part != frontPath && part != backPath {
			cleanedParts = append(cleanedParts, part)
		}
	}

	// Build new PATH: front folder + cleaned parts + back folder.
	var newPath string
	if len(cleanedParts) == 0 {
		newPath = frontPath + string(os.PathListSeparator) + backPath
	} else {
		newPath = frontPath + string(os.PathListSeparator) +
			strings.Join(cleanedParts, string(os.PathListSeparator)) +
			string(os.PathListSeparator) + backPath
	}

	return newPath, nil
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
	exportLine := "\n# Added by pathman\nexport PATH=$(pathman path)\n"
	if _, err := f.WriteString(exportLine); err != nil {
		return fmt.Errorf("failed to write to profile: %w", err)
	}

	fmt.Printf("Added PATH export to %s\n", profilePath)
	fmt.Println("Please restart your shell or run: source", profilePath)
	return nil
}

// profileHasPathmanExport checks if the profile already has a pathman export.
func profileHasPathmanExport(profilePath string) (bool, error) {
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
	folderPath, err := GetManagedFolder(atFront)
	if err != nil {
		return nil, fmt.Errorf("failed to get managed folder path: %w", err)
	}

	if !Exists(folderPath) {
		return nil, fmt.Errorf("managed folder does not exist: %s", folderPath)
	}

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read managed folder: %w", err)
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

// SymlinkInfo represents information about a symlink.
type SymlinkInfo struct {
	Name   string
	Target string
}

// ListLong returns detailed information about all symlinks in the managed folder.
func ListLong(atFront bool) ([]SymlinkInfo, error) {
	folderPath, err := GetManagedFolder(atFront)
	if err != nil {
		return nil, fmt.Errorf("failed to get managed folder path: %w", err)
	}

	if !Exists(folderPath) {
		return nil, fmt.Errorf("managed folder does not exist: %s", folderPath)
	}

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read managed folder: %w", err)
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
			symlinks = append(symlinks, SymlinkInfo{
				Name:   entry.Name(),
				Target: target,
			})
		}
	}

	return symlinks, nil
}

// Add creates a symlink to the executable in the managed folder.
// If a symlink with the same name exists in the other folder, it's moved to the specified folder.
func Add(executablePath, name string, atFront bool) error {
	folderPath, err := GetManagedFolder(atFront)
	if err != nil {
		return fmt.Errorf("failed to get managed folder path: %w", err)
	}

	if !Exists(folderPath) {
		return fmt.Errorf("managed folder does not exist: %s\nRun 'pathman init' to create it", folderPath)
	}

	// Get absolute path of the executable.
	absExecutablePath, err := filepath.Abs(executablePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if the executable exists.
	info, err := os.Stat(absExecutablePath)
	if err != nil {
		return fmt.Errorf("executable does not exist: %s", absExecutablePath)
	}

	// Check if it's a regular file or symlink (not a directory).
	if info.IsDir() {
		return fmt.Errorf("cannot add directory: %s", absExecutablePath)
	}

	// Determine the symlink name.
	symlinkName := name
	if symlinkName == "" {
		symlinkName = filepath.Base(absExecutablePath)
	}

	symlinkPath := filepath.Join(folderPath, symlinkName)

	// Check if symlink already exists in the target folder.
	if _, err := os.Lstat(symlinkPath); err == nil {
		return fmt.Errorf("symlink already exists: %s", symlinkName)
	}

	// Check if symlink exists in the other folder and remove it if so.
	otherFolderPath, err := GetManagedFolder(!atFront)
	if err == nil && Exists(otherFolderPath) {
		otherSymlinkPath := filepath.Join(otherFolderPath, symlinkName)
		if _, err := os.Lstat(otherSymlinkPath); err == nil {
			// Symlink exists in other folder, remove it.
			if err := os.Remove(otherSymlinkPath); err != nil {
				return fmt.Errorf("failed to remove symlink from other folder: %w", err)
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

// Remove removes a symlink from the managed folders (searches both front and back).
func Remove(name string) error {
	frontPath, backPath, err := GetBothManagedFolders()
	if err != nil {
		return fmt.Errorf("failed to get managed folder paths: %w", err)
	}

	// Try front folder first.
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

	// Try back folder.
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

// Rename renames a symlink in the managed folders (searches both front and back).
func Rename(oldName, newName string) error {
	frontPath, backPath, err := GetBothManagedFolders()
	if err != nil {
		return fmt.Errorf("failed to get managed folder paths: %w", err)
	}

	// Try front folder first.
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

	// Try back folder.
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
