# Initial development

## Goal

Implement the _structure_ of the pathman console-application in Golang. Use
a full workspace structure with `cmd` and `pkg` folders.

## Part 1

It should be able to process the command-line options but the actions should be
stubbed out. Without any arguments, it should report on the folder for symlinks.
If the folder cannot be found it should report that it is missing and inform
the user that the folder can be created with the command `pathman init`.

Note: Use cobra for command line parsing.

## Part 2

Implement the `pathman init` command. If the folder does not exist, create it,
ensure the permissions are `chmod a+r,u+w` only, and log the action. If the 
folder already exists, check the permission and complain if anyone except the
user has write permission. Report the action back to the console.

## Part 3 - extending `pathman init`

- Check the $PATH variable to determine if the managed folder is on the $PATH.
- If it is not, print out a message explaining that it should be added to your $PATH.
- If the SHELL is `bash` explain that this is normally put in your `.profile` or `.bash_profile` and offer to add a suitable command at the end of the relevant file.
- If the user accepts this, then detect which of `.profile` or `.bash_profile` need to be editing, make the relevant change, and inform the user of what was done.

## Part 4

Implement the `pathman add`, `pathman remove` and `pathman list` commands. This
is effectively CRUD for our managed folder.


## Part N

All normal interactive commands should run the safety check. 


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
