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

Note: Use cobra for command line parsing.

## Part B

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


## Part C

Implement the `pathman add`, `pathman remove` and `pathman list` commands. This
is effectively CRUD for our managed folder.

- The `pathman add` command takes optional arguments `--front` and `--back` to  
  determine which of the two subfolders the symlink is inserted into. 
- `pathman add` ensures that if the name already exists then it is not overwritten
  unless the `--force` option is supplied. When `--force` is used, any pre-existing
  link is deleted.
- `pathman add` will also check whether the added symlink will mask an executable
  on the rest of the $PATH or is masked by another executable. In either case
  it will give a warning and will decline to add the symlink without the 
  `--force` option.
- `pathman add` expands the link to an absolute path.
- If the priority is omitted, it defaults to `back`.

- `pathman remove|rm` will remove a symlink from either the front or back subfolders
  (or both, in case some manual editing has been done).

- `pathname list|ls` will list all the commands in a simple list but the `--long` or `-l` will
  add the priority (front/back) and the file it is linked to. The commands are
  listed in alphabetical order. By adding `--front` or `--back` options the
  list is restricted by priority - both options are not allowed.


## Part D

Implement the `pathman path` subcommand. This is to support this use-case:
`export PATH=$(pathman path)` into our `.bash_profile` (o.n.o.)

- `pathman path`: Checks $PATH to see if the managed subfolders are already on
  there, removing them if they are, then adds the front subfolder to the front
  of $PATH and the back subfolder to the end of $PATH. And finally prints the
  adjusted $PATH.

## Part E

Implement the `pathman rename|mv OLD NEW` command and the `pathman priority NAME`
and `pathman priority NAME=VALUE` commands.

## Part F

All normal interactive commands should run the permissions safety check. 

