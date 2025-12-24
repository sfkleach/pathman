# Add the ability to manage directories too

**Goal**: As a user, I can manage folders with `pathman` such as `~/.cargo/bin`,
`~/go/bin`, `~/.dotnet/tools` etc.

Implementation note: the folders are added to or removed from the pathman configuration file.

## Features

- `pathman add <directory>`, the `add` subcommand now supports adding a 
  directory. 
  - It can be combined with the `--priority` option. 
  - The directory path is always added as an absolute path.

- `pathman remove|rm <directory>`, the `remove` subcommand now supports 
  removing a directory. The directory path is made absolute and must exactly
  match an existing entry.

- `pathman list|ls`, the `list` subcommand now also list directories that 
  are managed by `pathman`. When listed their absolute paths are listed.

- Subcommands `get` and `set` are unchanged.

- `pathman path`, the subcommand `path` will remove any of the directories
  that are already on the $PATH (as well as removing its managed subfolders)
  and then add them back into the $PATH on the front or back depending on
  the priority. The managed subfolders are always first and last, the managed
  directories follow and precede them.

- `pathman [summary]`, the subcommand `summary` will list the configured
  directories. In addition it will perform a health check on those directories
  (do they exist? do they have good permissions?) including finding if any
  of their executable files mask or are masked by other path members, reporting
  any clashes, as before.
