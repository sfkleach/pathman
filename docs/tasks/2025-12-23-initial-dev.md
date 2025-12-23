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

## Part 4b

Add a `--long` (short option `-l`) to include the symlink info on `pathman list`.

## Part 4c

Adjust the README.md file to be consistent with the command-options developed
so far.

## Part 4d

- Ensure that `pathman add` always adds an absolute symlink to the managed folder.

## Part 5

Implement the `pathman path` subcommand, see below.

## Part 5b

- Ensure that the modification to the profile file (.profile or .bash_profile) 
  uses `pathname path` rather than hard-codes the managed folder name.
- And ensure the suggested modification to that file is consistent with this.

## Part 6

- Any command that can take the `--front` or `--back` options will default
  to using `--back`.
- Instead of a single managed folder, change to using TWO managed folders:
  - `pathman-links-front`
  - `pathman-links-back`
- To set these folders we use the commands: 
  - `pathman folder --front --set=PATH`
  - `pathman folder --back --set=PATH`
- The `pathman folder [--front] [--back]` command lists the relevant folder.
- The `pathname` command with no subcommand will now be a synonym for `pathman summary`
  and will do the following:
  - List both front and back folders in a human-friendly format
  - Check the $PATH for all name clashes.
- The `pathman init` command will set up both folders and, as the name
  suggests, put one on the front of the $PATH and one on the back of $PATH.
- The interactive commentary will be adjusted accordingly.
- Adjust the README.md accordingly.
- `pathname remove|rm` will remove from either folder.
- `pathname add` will check both folders and, if appropriate move the 
  symlink to the correct managed folder.


## Part X

All normal interactive commands should run the permissions safety check. 


## Background

Pathman is a command-line tool that helps you manage the list of applications
accessible by $PATH. With `pathman`, you can add, remove, and list executables
in a single local folder. By adding the local folder to your $PATH, you can
ensure your $PATH is alwys up to date with the executables you need.

Commands are:

- `pathman add <executable>` [--name NAME]: Adds an executable to a managed folder.
  Optionally, you can specify a different name for the symlink using the `--name` flag.
  Even if supplied in relative form the symlink must be to an absolute path.
  Optionally takes the `--front` or `--back` options, to determine which  managed
  folder to add it to.

- `pathman remove <name>`: Removes the symlink with the specified name from 
  the managed folders.
  Note that `rm` is a synonym for `remove`.

- `pathman list`: Lists all executables currently managed by `pathman`.
  Note that `ls` is a synonym for `list`. This an unadorned list by default.
  Use the `--long` option to include a link back to the symlinked file.
  `-l` also lists whether or not it masks another $PATH entry and which 
  entries it is masked by.

- `pathname rename OLD NEW`: renames a managed symlink.

- `pathman folder`: Displays the path to a managed folder. The folder
   argument is optional.

- `pathman folder --set <path>`: Sets the managed folder to the specified path,
    creating it if it doesn't exist and saving the configuration in
    `$XDG_CONFIG_HOME/pathman/config.json`, falling back to `$HOME/.config/pathman/config.json`.
    Note that even if supplied in relative form it must be expanded to an 
    absolute path.

- `pathman init`: Creates the managed folder if it does not yet exist.

- `pathman path [--front] [--back]`: Checks $PATH to see if the managed folder is already on there.
  If not it adds the managed folder to the front or back (defaulting to the back)
  Echos the adjusted path. This is to support this use-case:
   `export PATH=`pathman path` into our `.bash_profile` (o.n.o.)
