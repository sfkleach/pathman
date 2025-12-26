# Change Log for Pathman

Following the style in https://keepachangelog.com/en/1.0.0/

## v0.1.0, 2025/12/25

### Added

- Core PATH management with two-priority system (front/back subfolders)
- `pathman init` command with interactive TUI for shell configuration setup
- `pathman init --no` non-interactive mode for scripted setups (creates folders without modifying shell profiles)
- Self-install feature: `pathman init` installs itself to standard location (`~/.local/pathman/bin/pathman`)
- `pathman version` command and `--version` flag for checking installed version
- `pathman add` command supporting both individual executables and entire directories
- `pathman remove` command (with `rm` alias) for removing managed items
- `pathman list` command (with `ls` alias) for viewing managed executables and directories
- `pathman rename` command (with `mv` alias) for renaming symlinks
- `pathman get` command to show symlink priority (front/back)
- `pathman set` command to move symlinks between front and back subfolders
- `pathman path` command for generating adjusted PATH with managed items
- `pathman summary` command (default when no arguments) showing health and conflicts
- `pathman clean` command with interactive TUI for removing broken symlinks and missing directories
- Directory management: track and include entire directories in PATH via configuration
- Automatic PATH masking detection with warnings when adding executables
- Name clash detection between front/back subfolders and managed directories
- Bubbletea-based interactive interfaces for `init` and `clean` commands
- Cross-platform support with appropriate file permissions (0755 for directories, 0644 for files)
- Configuration file at `~/.config/pathman/config.json` for tracking managed directories
- Comprehensive unit tests for config and folder packages
- Shell integration support for bash, zsh, and fish with safe error handling


