# Pathman - a $PATH manager

Pathman is a command-line tool that helps you manage the list of applications accessible by $PATH. With `pathman`, you can add, remove, and list executables in a single local folder. By adding the local folder to your $PATH, you can ensure your $PATH is alwys up to date with the executables you need.

Commands are:

- `pathman add <executable>` [--name NAME]: Adds an executable to the managed folder.
  Optionally, you can specify a different name for the symlink using the `--name` flag.

- `pathman remove <name>`: Removes the symlink with the specified name from the managed folder.

- `pathman list`: Lists all executables currently managed by `pathman`.

- `pathman folder`: Displays the path to the managed folder.

- `pathman folder --set <path>`: Sets the managed folder to the specified path,
    creating it if it doesn't exist and saving the configuration in
    `$XDG_CONFIG_HOME/pathman/config.json`, falling back to `$HOME/.config/pathman/config.json`.

## Get Started

Install `pathman` using your preferred method (e.g., pip, cargo, etc.).

Bash users can add the following line to their `.bashrc` or `.bash_profile`:

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
