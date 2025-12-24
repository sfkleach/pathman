# Architecture

## Overview

Pathman manages $PATH by combining two mechanisms:
1. Symlinks for individual executables
2. Configuration references for entire directories

This dual approach provides flexibility: use symlinks for single executables from applications, and use directory references for tool collections like `~/.cargo/bin` or `~/go/bin`.

## Directory Structure

```
~/.local/bin/pathman-links/           # Base managed folder
├── front/                            # High priority symlinks
└── back/                             # Low priority symlinks

~/.config/pathman/config.json         # Managed directories
```

The managed folder contains two subfolders, each holding symlinks to executables. The configuration file tracks entire directories that should be included in the PATH.

## Why Two Subfolders?

The front/back subfolder approach solves PATH's fundamental limitation: you can't easily control priority ordering.

- **Front subfolder**: Override system executables (e.g., use a newer version of python or node)
- **Back subfolder**: Fallback executables (e.g., custom scripts that shouldn't shadow system tools)

This gives users fine-grained control over executable precedence without editing shell configuration files directly.

## Why Symlinks AND Configuration?

Using two different mechanisms provides the best of both worlds:

### Symlinks (for individual executables)
- **Portable**: Work regardless of the source directory structure
- **Visible**: Users can `ls` the managed folder to see what's available
- **Simple**: One symlink = one executable
- **Inspectable**: `readlink` shows where the executable lives

### Configuration (for directories)
- **Efficient**: Avoids creating hundreds of symlinks for large tool directories
- **Dynamic**: New executables in the directory become available immediately
- **Clean**: Removing the directory reference is one operation
- **Natural**: Matches how users think about tool installations (cargo, go, etc.)

## PATH Generation

When your shell calls `pathman path`, it generates a complete PATH in this order:

```
front-subfolder : front-dirs : $PATH : back-dirs : back-subfolder
```

This ensures:
1. User's high-priority symlinked tools come first
2. User's high-priority directories come next
3. System tools and existing PATH remain in the middle
4. User's low-priority directories come after system tools
5. User's low-priority symlinked fallbacks come last

Before adding pathman's components, any existing occurrences of pathman-managed items are removed from $PATH to prevent duplicates.

## Code Organization

```
cmd/pathman/            # Main entry point
pkg/
├── commands/           # Cobra command definitions
│   ├── commands.go     # Root and core commands
│   ├── clean.go        # Interactive cleanup TUI
│   └── init.go         # Interactive initialization TUI
├── config/             # Configuration file management
│   ├── config.go       # Load/save config.json
│   └── config_test.go  # Configuration tests
└── folder/             # Core folder operations
    ├── folder.go       # Add/remove/list operations
    ├── clean.go        # Cleanup detection logic
    └── folder_test.go  # Folder operation tests
```

### Key Components

- **commands package**: Handles user interaction via Cobra CLI framework and Bubbletea TUI
- **config package**: Manages persistent storage of managed directories in JSON format
- **folder package**: Core business logic for symlink and directory management

## Interactive Features

Pathman uses [Bubbletea](https://github.com/charmbracelet/bubbletea) for interactive TUI features:

- **`pathman init`**: Interactive prompt for adding PATH configuration to shell profile
- **`pathman clean`**: Visual selection interface for removing broken symlinks and missing directories

Both commands use consistent keyboard controls:
- Arrow keys or k/j to navigate
- Space to toggle selection
- Enter to confirm
- q to quit

## Security Considerations

### File Permissions

- Managed folders use 0755 permissions (owner write, all read/execute) because:
  - PATH directories must be readable and traversable by all users
  - Multiple users may need to execute the same tools
  - Only the owner should be able to modify the symlinks

- Configuration file uses 0644 permissions (owner write, all read) because:
  - Configuration doesn't need to be executable
  - Other users may need to read which directories are managed
  - Only the owner should modify the configuration

### Path Validation

- Symlink targets are validated before creation
- Relative paths are converted to absolute paths
- PATH masking is detected and warned about
- Users must explicitly use `--force` to override safety checks

### Shell Configuration Safety

The generated shell configuration includes multiple safety checks:
- Verifies pathman is available before calling it
- Checks exit code of `pathman path` before using output
- Only shows warnings in interactive shells (checks $PS1)
- Provides fallback behavior if pathman fails

This prevents users from being locked out of their system if pathman becomes unavailable.

## Testing Strategy

Tests use real filesystem operations via `t.TempDir()` to ensure pathman works correctly with actual files:

- **Config tests**: Verify JSON persistence, default path resolution, directory creation
- **Folder tests**: Test symlink creation, directory management, PATH manipulation, name clash detection

Function variables (e.g., `GetConfigPath`, `GetDefaultManagedFolder`) allow test overrides without mocking the entire filesystem.

## Future Extensibility

The architecture supports future enhancements:

- **Multiple shell support**: Currently bash-focused for auto-configuration, but manual instructions work for all shells
- **Additional priority levels**: The two-level system (front/back) could be extended to numeric priorities
- **Remote executables**: Could support downloading executables from URLs
- **Version management**: Could track multiple versions of the same tool
- **Profile management**: Could support multiple profiles (work vs personal)
