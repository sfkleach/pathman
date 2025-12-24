# Pathman - a $PATH manager

Pathman is a command-line tool that helps you manage the list of applications
accessible by $PATH. With `pathman`, you can add individual executables to the
front or back of your $PATH, remove them, list them and detect path clashes.

## Example 1, ~/.cargo/bin

Rust's `cargo` is great - but you need to add the `~/.cargo/bin` folder to your $PATH, 
which usually means editing your .profile (or was it .bash_profile on
this system?). Pathman lets you add a folder to your path with a single command
without editing your profile files.

```sh
pathman add ~/.cargo/bin
```

Pathman will also check whether or not adding ~/.cargo/bin to your $PATH will
cause any system executables to be masked or, equally unfortunately, whether
the ~/.cargo/bin executables will be masked by something else.

## Example 2, random application

For example, let's suppose you download a zip file `foozle.zip` for the
imaginary command-line application `foozle`. When you unpack it you find it has
a fairly typical structure with a single executable like this:

```
❯ tree foozle/
foozle/
├── bin
│   └── foozle
└── share
    ├── data1
    ├── data2
    └── data3
```

You find somewhere to put it, maybe `~/.local/foozle` and now you want to add
the binary to your $PATH. You could, of course, manually add a symlink to 
`~/.local/bin`. Except then you discover that isn't on your $PATH either!

Alternatively you could use `pathman add ~/.local/foozle/bin/foozle`. This
symlinks the single executable into pathman's own managed folders. You can
review the state of health of your managed path at any point - and when you get
rid of the foozle app, `pathman clean` can easily clean up broken symlinks or
even missing folders.


## Commands

- `pathman init`: Creates the managed folder and both subfolders if they don't exist. Checks if the subfolders are on your $PATH and offers to add them to your shell configuration (for bash users).

- `pathman add <executable>` [--name NAME] [--priority=PRIORITY] [--force]: Adds a symlink to an executable in the managed folder. The executable path can be relative or absolute. Use `--priority=front` or `--priority=back` (default: back). If a symlink with the same name exists in the other subfolder, it will be moved. Use `--force` to overwrite existing symlinks and ignore PATH masking warnings.

- `pathman remove <name>` (alias: `rm`): Removes the symlink with the specified name from whichever subfolder contains it (searches both).

- `pathman rename <old-name> <new-name>` (alias: `mv`): Renames a symlink in whichever subfolder contains it.

- `pathman list` (alias: `ls`) [--priority=PRIORITY]: Lists all executables. Without `--priority`, lists from both subfolders. Use `--priority=front` or `--priority=back` to filter. Use `--long` or `-l` to show priority and symlink targets.

- `pathman get <name>`: Shows which subfolder (front or back) a symlink is in.

- `pathman set <name> --priority=PRIORITY`: Moves a symlink between front and back subfolders.

- `pathman path`: Outputs an adjusted $PATH with both managed subfolders properly positioned (front subfolder at front, back subfolder at back). Removes any existing occurrences of the subfolders. Use this in shell configuration: `export PATH=$(pathman path)`

- `pathman summary`: Shows a summary of the managed folder, both subfolders with symlink counts, and any naming conflicts (folder clashes or PATH clashes).

- `pathman` (no arguments): Shows the same summary as `pathman summary`.

- `pathman clean`: detects broken symlinks and missing directories and helps you to delete them interactively.

## Implementation

Pathman manages a base folder `~/.local/bin/pathman-links` with two subfolders of symlinks:
- **Front subfolder**: `~/.local/bin/pathman-links/front` - Added to the front of $PATH (highest precedence)
- **Back subfolder**: `~/.local/bin/pathman-links/back` - Added to the back of $PATH (lowest precedence)

It works by adding and removing symlinks from these two folders. 


## Get Started

First, initialize the managed folder:

```bash
pathman init
```

This will create the managed folder with both subfolders and offer to add them to your $PATH (for bash users).

Alternatively, you can manually add the subfolders to your PATH. Bash users can add the following line to their `.bashrc` or `.bash_profile`:

```bash
export PATH=$(pathman path)
```

Zsh users can add the following line to their `.zshrc`:

```zsh
export PATH=$(pathman path)
```

Fish users can use:

```fish
set -gx PATH (pathman path | string split :)
``` 

After updating your shell configuration, restart your terminal or source the configuration file to apply the changes.

## Usage Examples

Add an executable to the back subfolder (default):

```bash
pathman add /usr/local/bin/myapp
```

Add an executable to the front subfolder (high precedence):

```bash
pathman add /path/to/executable --priority=front
```

Add an executable with a custom name:

```bash
pathman add /path/to/executable --name mycommand
```

List all executables from both subfolders:

```bash
pathman list
# or use the short alias
pathman ls
```

List executables from the front subfolder only:

```bash
pathman list --priority=front
```

List with priority and symlink targets:

```bash
pathman list --long
# or
pathman ls -l
```

Show which subfolder a symlink is in:

```bash
pathman get mycommand
```

Move a symlink to the front subfolder:

```bash
pathman set mycommand --priority=front
```

Rename a managed executable:

```bash
pathman rename oldname newname
# or use the short alias
pathman mv oldname newname
```

Remove a managed executable:

```bash
pathman remove mycommand
# or use the short alias
pathman rm mycommand
```

Check the managed folder status:

```bash
pathman
# or
pathman summary
```
