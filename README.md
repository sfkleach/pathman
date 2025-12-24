# Pathman - a $PATH manager

Pathman is a command-line tool that helps you manage the list of applications accessible by $PATH. With `pathman`, you can add, remove, and list executables in two managed folders (front and back). By positioning these folders at the front and back of your $PATH, you have fine-grained control over command precedence.

## Two-Folder Architecture

Pathman manages two folders:
- **Front folder**: `~/.local/bin/pathman-links-front` - Added to the front of $PATH (highest precedence)
- **Back folder**: `~/.local/bin/pathman-links-back` - Added to the back of $PATH (lowest precedence)

By default, commands operate on the **back folder** unless `--front` is specified.

## Commands

- `pathman init`: Creates both managed folders if they don't exist. Checks if the folders are on your $PATH and offers to add them to your shell configuration (for bash users).

- `pathman add <executable>` [--name NAME] [--front|--back]: Adds a symlink to an executable in the managed folder. The executable path can be relative or absolute. Use `--front` to add to the front folder, or `--back` to add to the back folder (default). If a symlink with the same name exists in the other folder, it will be moved.

- `pathman remove <name>` (alias: `rm`): Removes the symlink with the specified name from whichever folder contains it (searches both).

- `pathman rename <old-name> <new-name>`: Renames a symlink in whichever folder contains it.

- `pathman list` (alias: `ls`) [--front|--back]: Lists all executables in the specified folder. By default lists the back folder. Use `--long` or `-l` to show symlink targets.

- `pathman folder` [--front|--back]: Displays the paths of both managed folders and their status. With `--set <path>`, changes the path of the specified folder (requires either `--front` or `--back`).

- `pathman path`: Outputs an adjusted $PATH with both managed folders properly positioned (front folder at front, back folder at back). Removes any existing occurrences of the folders. Use this in shell configuration: `export PATH=$(pathman path)`

- `pathman` (no arguments): Shows a summary of both managed folders, their status, and any name clashes.

## Get Started

First, initialize the managed folders:

```bash
pathman init
```

This will create both managed folders and offer to add them to your $PATH (for bash users).

Alternatively, you can manually add the folders to your PATH. Bash users can add the following line to their `.bashrc` or `.bash_profile`:

```bash
export PATH=$(pathman path)
```

Zsh users can add the following line to their `.zshrc`:

```zsh
export PATH=$(pathman path)
```

Fish users can use:

```fish
set -gx PATH (pathman path | string split :)
``` 

After updating your shell configuration, restart your terminal or source the configuration file to apply the changes.

## Usage Examples

Add an executable to the back folder (default):

```bash
pathman add /usr/local/bin/myapp
```

Add an executable to the front folder (high precedence):

```bash
pathman add --front /path/to/executable
```

Add an executable with a custom name:

```bash
pathman add /path/to/executable --name mycommand
```

List executables in the back folder:

```bash
pathman list
# or use the short alias
pathman ls
```

List executables in the front folder:

```bash
pathman list --front
```

List with symlink targets:

```bash
pathman list --long
# or
pathman ls -l --front
```

Rename a managed executable:

```bash
pathman rename oldname newname
```

Remove a managed executable:

```bash
pathman remove mycommand
# or use the short alias
pathman rm mycommand
```

Check the managed folders status:

```bash
pathman
# or
pathman folder
```
