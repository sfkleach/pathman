# Initial development

## Goal

Implement the first draft the pathman console-application in Golang. Use
a full workspace structure with `cmd` and `pkg` folders.

## Part A

It should be able to process the command-line options but the actions should be
stubbed out. Without any arguments, it should summarise the managed folder and
subfolders. 

If the folder cannot be found it should report that it is missing and inform
the user that the folder can be created with the command `pathman init`.

The set of subcommands are:

```
Available Commands:
  add         Add an executable to the managed folder
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Create the managed folder
  ls|list     List managed executables and directories
  path        Output PATH with managed folders included
  rm|remove   Remove a symlink from the managed folder
  mv|rename   Rename a symlink in the managed folders
  summary     Display a summary of both managed folders
  version     Print version information
```

The managed folder is determined by JSON configuration in `./.config/pathman.conf`.
By default it will be `~/.local/bin/pathman-links/` with subfolders `front` and 
`back`.

Note: Use cobra for command line parsing.

## Part B

Implement the `pathman path` subcommand. This is to support this use-case:
`NEWPATH=$(pathman path)` into our `.bash_profile` (o.n.o.)

- `pathman path`: Checks $PATH to see if the managed subfolders are already on
  there, removing them if they are, then adds the front subfolder to the front
  of $PATH and the back subfolder to the end of $PATH. And finally prints the
  adjusted $PATH.

- `pathman path` will also remove any duplicate folders that appear on the $PATH,
  not just its own managed folders, discarding any occurences after the first.
  This means it cleans up existing $PATH errors.


## Part C (20 mins)

Implement the `pathman init` command. If the managed folder does not exist,
create it, ensure the permissions are `chmod a=rx,u+w`, and log the action. If
the folder already exists, check the permission and complain if anyone except
the user has write permission. Report the action back to the console.

`pathname init` should:

- Check the $PATH variable to determine if the front-and-back subfolders are on
  the $PATH.
- If it is not, print out a message explaining that the pathman folders should
  be added to your $PATH.
- If the SHELL is `bash` explain that this is normally put in your `.profile` or
  `.bash_profile` and offer to add a suitable command at the end of the relevant
  file.
- If the user accepts this, then detect which of `.profile` or `.bash_profile`
  need to be editing, make the relevant change, and inform the user of what was
  done.
    - Ensure that the modification to the profile file (.profile or
      .bash_profile) uses `pathname path` rather than hard-codes the managed
      folder name.

Path substiution is managed by the `path` subcommand. In principle this allows
the modification to $PATH to be written as 

```
export PATH=$(pathman path 2>/dev/null)
```

However we have to be careful not to render the shell unusable if `pathman` 
is missing or has an error. Furthermore, it should not generate output in 
non-interactive contexts. So it should look something like:

```sh
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

## Part D (10 mins)

Implement the `pathman add`, `pathman remove` and `pathman list` commands. This
is effectively CRUD for our managed folder.

- The `pathman add` command takes an optional argument `--name=NAME` which 
  is overrides the default name of the symlink.
- The `pathman add` command takes an optional argument `--priority=front|back`
  determine which of the two subfolders the symlink is inserted into. 
- `pathman add` ensures that if the name already exists then it is not overwritten
  unless the `--force` option is supplied. When `--force` is used, any pre-existing
  link is deleted.
- `pathman add` will also check whether the added symlink will mask an executable
  on the rest of the $PATH or is masked by another executable. In either case
  it will give a warning and will decline to add the symlink without the 
  `--force` option.
- `pathman add` expands the link to an absolute path.
- If the priority is omitted, it defaults to `front`.

- `pathman remove|rm` will remove a symlink from either the front or back subfolders
  (or both, in case some manual editing has been done).

- `pathname list|ls` will list all the commands in a simple list but the `--long` or `-l` will
  add the priority (front/back) and the file it is linked to. The commands are
  listed in alphabetical order. By adding `--priority=front|back` option the
  list is restricted by priority. 


## Part E (10 mins)

Implement the `pathman rename|mv OLD NEW` command, which renames a managed
symlink. It checks whether or not the renaming would create problems such as
a duplicate name on the path or the target symlink already exists (and would need
unlinking). The `--force` option overrides. 

If a pre-existing symlink has to be deleted, the original link printed out as
part of the 'work-done' message. This gives a brief window of regret, so to
speak, to the user.


