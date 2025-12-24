# Add the ability to manage directories too

**Goal**: As a user, I can manage folders with `pathman` such as `~/.cargo/bin`,
`~/go/bin`, `~/.dotnet/tools` etc.

Implementation note: the folders are added to or removed from a configuration
file in JSON or YAML format that manages the items and their priorities. 

Question: Because this configuration file is edited directly by `pathman` I think it 
should be in the managed folder rather than be treated as a normal configuration
file i.e. more like data. What is the recommended approach?

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
