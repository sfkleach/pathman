# Task: pathman init installs the pathman binary into its own managed subfolders

**Goal**: As a user, when I first run `pathman init`, it allows me to optionally
install itself into a standard location.

This task extends the `pathman init` interaction. After adding the set-up script
into .profile/.bash_profile (or skipping that step), it asks the user if it
should relocate the executable to the standard location and add to the managed
path?

- Standard location: I suggest this is ~/.local/pathman/bin/pathman
- Managed path: the front subfolder

If the user approves this suggestion, it creates the standard location and
attempts to copy the executable to it, creates the symlink in the front 
subfolder, and finally unlink the running executable.

Note: If the application is already in the standard location, this interaction
should not be offered.

## Implementation note on Windows

On windows we cannot unlink a running executable. On Windows we should therefore
launch a background task just before the application quits that attempts to 
delete it - a powershell script will do it.

## Motivation

This idea behind this change is to answer the question "where should pathman be
installed to minimise dependencies?" The simple answer is that pathman requires
the .local folder exists, so it can be put there, and the front subfolder 
must exist and that's the natural place for the symlink.
