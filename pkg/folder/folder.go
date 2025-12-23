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
func GetManagedFolder() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	if cfg != nil && cfg.ManagedFolder != "" {
		return cfg.ManagedFolder, nil
	}

	return config.GetDefaultManagedFolder()
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
func SetManagedFolder(folderPath string) error {
	cfg := &config.Config{
		ManagedFolder: folderPath,
	}
	return cfg.Save()
}

// PrintStatus prints the status of the managed folder.
func PrintStatus() error {
	folderPath, err := GetManagedFolder()
	if err != nil {
		return fmt.Errorf("failed to get managed folder path: %w", err)
	}

	if Exists(folderPath) {
		fmt.Printf("Managed folder: %s\n", folderPath)
		return nil
	}

	fmt.Printf("Managed folder does not exist: %s\n", folderPath)
	fmt.Println("You can create it with: pathman init")
	return nil
}

// Init initializes the managed folder.
// If the folder doesn't exist, it creates it with appropriate permissions (chmod a+r,u+w).
// If the folder exists, it checks permissions and warns if anyone except the user has write permission.
// It also checks if the folder is on $PATH and offers to add it for bash users.
func Init() error {
	folderPath, err := GetManagedFolder()
	if err != nil {
		return fmt.Errorf("failed to get managed folder path: %w", err)
	}

	folderCreated := false

	if Exists(folderPath) {
		// Folder exists, check permissions.
		info, err := os.Stat(folderPath)
		if err != nil {
			return fmt.Errorf("failed to stat folder: %w", err)
		}

		mode := info.Mode()
		perm := mode.Perm()

		// Check if group or others have write permission (bits 1 or 4 in octal).
		// We want only user to have write (0200).
		if perm&0022 != 0 {
			fmt.Printf("Managed folder already exists: %s\n", folderPath)
			fmt.Printf("WARNING: Folder has insecure permissions: %04o\n", perm)
			fmt.Println("Group or others have write permission. This is a security risk.")
			fmt.Println("Recommended permissions: 0755 (owner read/write/execute, all read/execute)")
		} else {
			fmt.Printf("Managed folder already exists: %s\n", folderPath)
			fmt.Printf("Permissions are correct: %04o\n", perm)
		}
	} else {
		// Create the folder with permissions: owner read+write+execute, all read+execute (0755).
		if err := os.MkdirAll(folderPath, 0755); err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}

		fmt.Printf("Created managed folder: %s\n", folderPath)
		fmt.Printf("Permissions set to: 0755 (owner read/write/execute, all read/execute)\n")
		folderCreated = true
	}

	// Check if the folder is on $PATH.
	if !IsOnPath(folderPath) {
		fmt.Println()
		fmt.Printf("The managed folder is not on your $PATH.\n")
		fmt.Println("To use executables in this folder, you need to add it to your $PATH.")

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

			if answer, err := PromptUser("Would you like me to add it for you?"); err != nil {
				return fmt.Errorf("failed to read user input: %w", err)
			} else if answer {
				if err := AddToProfile(folderPath); err != nil {
					return fmt.Errorf("failed to add to profile: %w", err)
				}
			} else {
				fmt.Printf("\nTo add it manually, add this line to your ~/%s:\n", profileName)
				fmt.Printf("  export PATH=\"%s:$PATH\"\n", folderPath)
			}
		} else {
			fmt.Printf("\nTo add it to your PATH, add this line to your shell configuration:\n")
			fmt.Printf("  export PATH=\"%s:$PATH\"\n", folderPath)
		}
	} else if folderCreated {
		fmt.Println()
		fmt.Println("The managed folder is already on your $PATH.")
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
func AddToProfile(folderPath string) error {
	profilePath, err := GetBashProfilePath()
	if err != nil {
		return fmt.Errorf("failed to get profile path: %w", err)
	}

	// Check if the export line already exists.
	if hasPathExport, err := profileHasPathExport(profilePath, folderPath); err != nil {
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

	// Add the export line.
	exportLine := fmt.Sprintf("\n# Added by pathman\nexport PATH=\"%s:$PATH\"\n", folderPath)
	if _, err := f.WriteString(exportLine); err != nil {
		return fmt.Errorf("failed to write to profile: %w", err)
	}

	fmt.Printf("Added PATH export to %s\n", profilePath)
	fmt.Println("Please restart your shell or run: source", profilePath)
	return nil
}

// profileHasPathExport checks if the profile already has an export for the folder path.
func profileHasPathExport(profilePath, folderPath string) (bool, error) {
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
		// Check if the line exports PATH and contains our folder.
		if strings.Contains(line, "export") && strings.Contains(line, "PATH") && strings.Contains(line, folderPath) {
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
func List() ([]string, error) {
	folderPath, err := GetManagedFolder()
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

// Add creates a symlink to the executable in the managed folder.
func Add(executablePath, name string) error {
	folderPath, err := GetManagedFolder()
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

	// Check if symlink already exists.
	if _, err := os.Lstat(symlinkPath); err == nil {
		return fmt.Errorf("symlink already exists: %s", symlinkName)
	}

	// Create the symlink.
	if err := os.Symlink(absExecutablePath, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	fmt.Printf("Added '%s' -> '%s'\n", symlinkName, absExecutablePath)
	return nil
}

// Remove removes a symlink from the managed folder.
func Remove(name string) error {
	folderPath, err := GetManagedFolder()
	if err != nil {
		return fmt.Errorf("failed to get managed folder path: %w", err)
	}

	if !Exists(folderPath) {
		return fmt.Errorf("managed folder does not exist: %s", folderPath)
	}

	symlinkPath := filepath.Join(folderPath, name)

	// Check if the symlink exists.
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("symlink does not exist: %s", name)
		}
		return fmt.Errorf("failed to stat symlink: %w", err)
	}

	// Make sure it's a symlink.
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("'%s' is not a symlink", name)
	}

	// Remove the symlink.
	if err := os.Remove(symlinkPath); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	fmt.Printf("Removed '%s'\n", name)
	return nil
}
