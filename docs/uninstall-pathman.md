# Completely removing pathman

To completely remove `pathman` from your system, follow these instructions:

## Step 1: Remove shell integration

Edit your shell configuration file (`.profile`, `.bash_profile`, or `.bashrc`)
and remove the pathman integration block. Look for and delete lines similar to:

```bash
if command -v pathman >/dev/null 2>&1; then
  PATHMAN_CMD=pathman
elif [ -x "$HOME/.local/pathman/bin/pathman" ]; then
  PATHMAN_CMD="$HOME/.local/pathman/bin/pathman"
fi

if [ -n "$PATHMAN_CMD" ]; then
  NEW_PATH=$("$PATHMAN_CMD" path 2>/dev/null)
  if [ $? -eq 0 ] && [ -n "$NEW_PATH" ]; then
    export PATH="$NEW_PATH"
  fi
fi
```

## Step 2: Remove the managed folder

Remove the pathman-managed symlinks folder:

```bash
rm -rf ~/.local/bin/pathman-links
```

## Step 3: Remove configuration

Remove the pathman configuration directory:

```bash
rm -rf ~/.config/pathman
```

## Step 4: Remove the pathman executable

If you installed pathman using `execman`:

```bash
execman remove pathman
```

If you installed pathman manually or to a custom location, remove the binary:

```bash
# If installed to ~/.local/bin
rm ~/.local/bin/pathman

# If installed to ~/.local/pathman/bin
rm -rf ~/.local/pathman
```

## Step 5: Restart your shell

Start a new shell session or source your profile to apply the changes:

```bash
exec $SHELL
```

## Verification

To verify pathman has been completely removed:

```bash
# Should return "command not found"
which pathman

# Should not exist
ls ~/.local/bin/pathman-links
ls ~/.config/pathman
```
