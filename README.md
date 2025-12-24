# Pathman - a $PATH manager

Pathman is a command-line tool that helps you manage the list of applications
accessible by $PATH. With `pathman`, you can add individual executables to the
front or back of your $PATH, remove them, list them and detect path clashes.

## Features

- **Two-priority system**: Manage executables at front (override system tools) or back (fallback) of $PATH
- **Automatic conflict detection**: Warns when adding executables that mask or are masked by other PATH entries
- **Interactive TUI**: Clean up broken symlinks and missing directories with a modern terminal interface
- **Directory management**: Add entire directories like `~/.cargo/bin` without symlinking individual files
- **Safe profile updates**: Automatically adds PATH configuration with safety checks for bash users
- **Health monitoring**: Review the state of your managed PATH and detect issues at any time

## Example 1: Adding a Directory

Rust's `cargo` is great - but you need to add the `~/.cargo/bin` folder to your $PATH, 
which usually means editing your .profile (or was it .bash_profile on
this system?). Pathman can manage this directory for you:

```sh
pathman add ~/.cargo/bin
```

**Note**: When you add a directory (not a file), pathman adds it to its configuration
and includes it when generating your PATH via `pathman path`. All executables in
that directory become available.

Pathman will also check whether or not adding ~/.cargo/bin to your $PATH will
cause any system executables to be masked or, equally unfortunately, whether
the ~/.cargo/bin executables will be masked by something else.

## Example 2: Adding a Single Executable

For individual executables, pathman creates symlinks in its managed folders.
Let's suppose you download a zip file `foozle.zip` for the imaginary command-line
application `foozle`. When you unpack it you find it has a fairly typical structure
with a single executable like this:

```
❯ tree foozle/
foozle/
├── bin
│   └── foozle
└── share
    ├── data1
    ├── data2
    └── data3
```

You find somewhere to put it, maybe `~/.local/foozle` and now you want to add
the binary to your $PATH. You could, of course, manually add a symlink to 
`~/.local/bin`. Except then you discover that isn't on your $PATH either!

With pathman, you can add just the binary:

```sh
pathman add ~/.local/foozle/bin/foozle
```

This creates a symlink in pathman's managed folder, making `foozle` available
on your PATH without modifying your shell configuration. You can review the state
of your managed path at any point - and when you get rid of the foozle app,
`pathman clean` can easily clean up broken symlinks or even missing folders.

## How It Works

Pathman uses two mechanisms to manage your PATH:

1. **Symlinks for executables**: Individual files are symlinked into managed subfolders
   - Front subfolder: `~/.local/bin/pathman-links/front` (high priority)
   - Back subfolder: `~/.local/bin/pathman-links/back` (low priority)

2. **Directory references for folders**: Entire directories like `~/.cargo/bin` are
   stored in configuration (`~/.config/pathman/config.json`) and included when
   generating your PATH.

Your shell configuration contains one line that calls `pathman path`, which generates
the complete PATH by combining:
- Front subfolder
- Front-priority directories
- Your existing PATH
- Back-priority directories
- Back subfolder

## Commands

- `pathman init`: Creates the managed folder and both subfolders if they don't exist. Checks if the subfolders are on your $PATH and offers to add them to your shell configuration using an interactive interface (for bash users).

- `pathman add <path>` [--name NAME] [--priority=PRIORITY] [--force]: Adds an executable or directory to pathman.
  - For **files**: Creates a symlink in the managed subfolder
  - For **directories**: Adds to configuration; all executables in the directory become available
  - Use `--priority=front` for high priority or `--priority=back` (default) for low priority
  - Use `--name` to customize the symlink name (files only)
  - If a symlink with the same name exists in the other subfolder, it will be moved
  - Use `--force` to overwrite existing symlinks and ignore PATH masking warnings

- `pathman remove <name>` (alias: `rm`): Removes the symlink with the specified name from whichever subfolder contains it (searches both).

- `pathman rename <old-name> <new-name>` (alias: `mv`): Renames a symlink in whichever subfolder contains it.

- `pathman list` (alias: `ls`) [--priority=PRIORITY]: Lists all executables. Without `--priority`, lists from both subfolders. Use `--priority=front` or `--priority=back` to filter. Use `--long` or `-l` to show priority and symlink targets.

- `pathman get <name>`: Shows which subfolder (front or back) a symlink is in.

- `pathman set <name> --priority=PRIORITY`: Moves a symlink between front and back subfolders.

- `pathman path`: Outputs an adjusted $PATH with managed subfolders and directories properly positioned. Removes any existing occurrences of pathman-managed items before adding them in the correct order. Only useful in shell configuration.

- `pathman summary`: Shows a summary of the managed folder, both subfolders with symlink counts, and any naming conflicts (folder clashes or PATH clashes).

- `pathman clean`: Interactively detect and remove broken symlinks and missing directories. Uses an interactive terminal UI to let you review and select items to clean up.

Note that `pathman` with no arguments is the same as `pathman summary`.

## Implementation

Pathman manages a base folder `~/.local/bin/pathman-links` with two subfolders of symlinks:
- **Front subfolder**: `~/.local/bin/pathman-links/front` - Added to the front of $PATH (highest precedence)
- **Back subfolder**: `~/.local/bin/pathman-links/back` - Added to the back of $PATH (lowest precedence)

It works by adding and removing symlinks from these two folders, and by tracking managed directories in its configuration file.

## Configuration

Pathman stores its configuration in `~/.config/pathman/config.json`. This file tracks:
- Managed directories and their priorities
- (Symlinks are not stored in config - they exist as actual files in the managed folders)

You normally don't need to edit this file directly - use `pathman add` and `pathman remove`.

## Get Started

First, initialize the managed folder:

```bash
pathman init
```

This will create the managed folder with both subfolders and offer to add them to your $PATH (for bash users).

Alternatively, you can manually add this configuration to your shell profile.
**Note**: This replaces your PATH with a pathman-managed version that includes
both managed subfolders and directories in the correct positions.

Bash users can add the following to their `~/.profile` or `~/.bash_profile`:

```bash
# ============ BEGIN PATHMAN CONFIG ============
# Added by pathman
if command -v pathman >/dev/null 2>&1; then
  # Calculate a new $PATH from the old one and pathman's configuration.
  NEW_PATH=$(pathman path 2>/dev/null)
  if [ $? -eq 0 ] && [ -n "$NEW_PATH" ]; then
    export PATH="$NEW_PATH"
  elif [ -n "$PS1" ]; then
    # PS1 is only set in interactive shells - safe to show errors here.
    echo "Warning: pathman failed to update PATH" >&2
  fi
elif [ -n "$PS1" ]; then
  # PS1 is only set in interactive shells - safe to show errors here.
  echo "Warning: pathman not found, PATH not updated" >&2
fi
# ============= END PATHMAN CONFIG =============
```

Zsh users can add similar configuration to their `.zshrc`:

```zsh
# Added by pathman
if command -v pathman >/dev/null 2>&1; then
  NEW_PATH=$(pathman path 2>/dev/null)
  if [ $? -eq 0 ] && [ -n "$NEW_PATH" ]; then
    export PATH="$NEW_PATH"
  fi
fi
```

Fish users can use:

```fish
# Added by pathman
if command -v pathman >/dev/null 2>&1
  set NEW_PATH (pathman path 2>/dev/null)
  if test -n "$NEW_PATH"
    set -gx PATH (string split : $NEW_PATH)
  end
end
``` 

After updating your shell configuration, restart your terminal or source the configuration file to apply the changes.

## Usage Examples

Add an executable to the back subfolder (default):

```bash
pathman add /usr/local/bin/myapp
```

Add an executable to the front subfolder (high precedence):

```bash
pathman add /path/to/executable --priority=front
```

Add an executable with a custom name:

```bash
pathman add /path/to/executable --name mycommand
```

List all executables from both subfolders:

```bash
pathman list
# or use the short alias
pathman ls
```

List executables from the front subfolder only:

```bash
pathman list --priority=front
```

List with priority and symlink targets:

```bash
pathman list --long
# or
pathman ls -l
```

Show which subfolder a symlink is in:

```bash
pathman get mycommand
```

Move a symlink to the front subfolder:

```bash
pathman set mycommand --priority=front
```

Rename a managed executable:

```bash
pathman rename oldname newname
# or use the short alias
pathman mv oldname newname
```

Remove a managed executable:

```bash
pathman remove mycommand
# or use the short alias
pathman rm mycommand
```

Check the managed folder status:

```bash
pathman
# or
pathman summary
```
