# Pathman - a $PATH manager

Pathman is a command-line tool that helps you manage the list of applications accessible by $PATH. With `pathman`, you can add, remove, and list executables in a single local folder. By adding the local folder to your $PATH, you can ensure your $PATH is always up to date with the executables you need.

## Commands

- `pathman init`: Creates the managed folder if it does not exist. Checks if the folder is on your $PATH and offers to add it to your shell configuration (for bash users).

- `pathman add <executable>` [--name NAME]: Adds a symlink to an executable in the managed folder. The executable path can be relative or absolute. Optionally, you can specify a different name for the symlink using the `--name` flag.

- `pathman remove <name>` (alias: `rm`): Removes the symlink with the specified name from the managed folder.

- `pathman list` (alias: `ls`): Lists all executables currently managed by `pathman`. By default shows just the names. Use `--long` or `-l` to show symlink targets.

- `pathman folder`: Displays the path to the managed folder.

- `pathman folder --set <path>`: Sets the managed folder to the specified path, creating it if it doesn't exist and saving the configuration in `$XDG_CONFIG_HOME/pathman/config.json`, falling back to `$HOME/.config/pathman/config.json`.

## Get Started

First, initialize the managed folder:

```bash
pathman init
```

This will create the managed folder and offer to add it to your $PATH (for bash users).

Alternatively, you can manually add the folder to your PATH. Bash users can add the following line to their `.bashrc` or `.bash_profile`:

```bash
export PATH="$(pathman folder):$PATH"
```

Zsh users can add the following line to their `.zshrc`:

```zsh
export PATH="$(pathman folder):$PATH"
```

Fish users can add the following line to their `config.fish`:

```fish
set -gx PATH (pathman folder) $PATH
``` 

After updating your shell configuration, restart your terminal or source the configuration file to apply the changes.

## Usage Examples

Add an executable to your PATH:

```bash
pathman add /usr/local/bin/myapp
```

Add an executable with a custom name:

```bash
pathman add /path/to/executable --name mycommand
```

List managed executables:

```bash
pathman list
# or use the short alias
pathman ls
```

List with symlink targets:

```bash
pathman list --long
# or
pathman ls -l
```

Remove a managed executable:

```bash
pathman remove mycommand
# or use the short alias
pathman rm mycommand
```

Check the managed folder location:

```bash
pathman folder
```
