package folder

import (
    "fmt"
    "os"

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
    fmt.Println("You can create it with: pathman folder --create")
    return nil
}
