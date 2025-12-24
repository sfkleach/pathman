package folder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sfkleach/pathman/pkg/config"
)

// CleanupItem represents an item that can be cleaned up.
type CleanupItem struct {
	Type        string // "symlink" or "directory"
	Name        string // Symlink name or directory path
	Path        string // Full path to the item
	Priority    string // "front", "back", or priority for directories
	Reason      string // Why it needs cleanup
	Selected    bool   // Whether it's selected for cleanup
	Description string // Human-readable description
}

// FindCleanupItems scans for broken symlinks and missing directories.
func FindCleanupItems() ([]CleanupItem, error) {
	var items []CleanupItem

	frontPath, backPath, err := GetBothSubfolders()
	if err != nil {
		return nil, fmt.Errorf("failed to get subfolder paths: %w", err)
	}

	// Check symlinks in front folder.
	if Exists(frontPath) {
		frontItems, err := findBrokenSymlinksInFolder(frontPath, "front")
		if err != nil {
			return nil, err
		}
		items = append(items, frontItems...)
	}

	// Check symlinks in back folder.
	if Exists(backPath) {
		backItems, err := findBrokenSymlinksInFolder(backPath, "back")
		if err != nil {
			return nil, err
		}
		items = append(items, backItems...)
	}

	// Check managed directories.
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	for _, dir := range cfg.ManagedDirectories {
		if _, err := os.Stat(dir.Path); os.IsNotExist(err) {
			items = append(items, CleanupItem{
				Type:        "directory",
				Name:        filepath.Base(dir.Path),
				Path:        dir.Path,
				Priority:    dir.Priority,
				Reason:      "Directory does not exist",
				Selected:    true, // Selected by default.
				Description: fmt.Sprintf("[%s] %s (missing)", dir.Priority, dir.Path),
			})
		} else if err != nil {
			// Check permission errors or other issues.
			items = append(items, CleanupItem{
				Type:        "directory",
				Name:        filepath.Base(dir.Path),
				Path:        dir.Path,
				Priority:    dir.Priority,
				Reason:      fmt.Sprintf("Cannot access: %v", err),
				Selected:    true,
				Description: fmt.Sprintf("[%s] %s (error: %v)", dir.Priority, dir.Path, err),
			})
		}
	}

	return items, nil
}

// findBrokenSymlinksInFolder scans a folder for broken symlinks.
func findBrokenSymlinksInFolder(folderPath, priority string) ([]CleanupItem, error) {
	var items []CleanupItem

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read folder: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(folderPath, entry.Name())
		info, err := os.Lstat(entryPath)
		if err != nil {
			continue
		}

		// Check if it's a symlink.
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(entryPath)
			if err != nil {
				items = append(items, CleanupItem{
					Type:        "symlink",
					Name:        entry.Name(),
					Path:        entryPath,
					Priority:    priority,
					Reason:      "Cannot read symlink target",
					Selected:    true,
					Description: fmt.Sprintf("[%s] %s (unreadable)", priority, entry.Name()),
				})
				continue
			}

			// Check if target exists.
			if _, err := os.Stat(target); os.IsNotExist(err) {
				items = append(items, CleanupItem{
					Type:        "symlink",
					Name:        entry.Name(),
					Path:        entryPath,
					Priority:    priority,
					Reason:      fmt.Sprintf("Target does not exist: %s", target),
					Selected:    true,
					Description: fmt.Sprintf("[%s] %s -> %s (broken)", priority, entry.Name(), target),
				})
			}
		}
	}

	return items, nil
}

// PerformCleanup removes the selected items.
func PerformCleanup(items []CleanupItem) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	configModified := false

	for _, item := range items {
		if !item.Selected {
			continue
		}

		if item.Type == "symlink" {
			// Remove symlink.
			if err := os.Remove(item.Path); err != nil {
				return fmt.Errorf("failed to remove symlink %s: %w", item.Name, err)
			}
			fmt.Printf("Removed symlink: %s\n", item.Description)
		} else if item.Type == "directory" {
			// Remove from config.
			for i, dir := range cfg.ManagedDirectories {
				if dir.Path == item.Path {
					cfg.ManagedDirectories = append(cfg.ManagedDirectories[:i], cfg.ManagedDirectories[i+1:]...)
					configModified = true
					break
				}
			}
			fmt.Printf("Removed from config: %s\n", item.Description)
		}
	}

	// Save config if modified.
	if configModified {
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	return nil
}
