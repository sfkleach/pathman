# Initial development

## Goal

Implement the _structure_ of the pathman console-application in Golang. Use
a full workspace structure with `cmd` and `pkg` folders.

## Part 1

It should be able to process the command-line options but the actions should be
stubbed out. Without any arguments, it should report on the folder for symlinks.
If the folder cannot be found it should report that it is missing and inform
the user that the folder can be created with the command `pathman init`.

## Part 2

To be completed.

## Background

Pathman is a command-line tool that helps you manage the list of applications
accessible by $PATH. With `pathman`, you can add, remove, and list executables
in a single local folder. By adding the local folder to your $PATH, you can
ensure your $PATH is alwys up to date with the executables you need.

Commands are:

- `pathman add <executable>` [--name NAME]: Adds an executable to the managed folder.
  Optionally, you can specify a different name for the symlink using the `--name` flag.

- `pathman remove <name>`: Removes the symlink with the specified name from the managed folder.

- `pathman list`: Lists all executables currently managed by `pathman`.

- `pathman folder`: Displays the path to the managed folder. The folder
   argument is optional.

- `pathman folder --set <path>`: Sets the managed folder to the specified path,
    creating it if it doesn't exist and saving the configuration in
    `$XDG_CONFIG_HOME/pathman/config.json`, falling back to `$HOME/.config/pathman/config.json`.

- `pathman init`: Creates the managed folder if it does not yet exist.
