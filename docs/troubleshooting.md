# Troubleshooting

This guide covers common issues you might encounter when using Pathman and how to resolve them.

## PATH Not Updated After Running `pathman init`

**Symptom**: You run `pathman add` successfully, but when you try to run the executable, you get "command not found".

**Cause**: Your current shell session hasn't reloaded the updated configuration.

**Solution**: Reload your shell configuration or restart your terminal.

For bash:
```bash
source ~/.profile
# or
source ~/.bash_profile
```

For zsh:
```bash
source ~/.zshrc
```

For fish:
```fish
source ~/.config/fish/config.fish
```

Alternatively, just open a new terminal window.

---

## Executables Not Found Even After Reloading

**Symptom**: After reloading your shell config, pathman-managed executables still aren't found.

**Diagnosis**: Check if pathman's folders are actually in your PATH:
```bash
echo $PATH | tr ':' '\n' | grep pathman
```

You should see entries like:
```
/home/username/.local/bin/pathman-links/front
/home/username/.local/bin/pathman-links/back
```

**Solution**: If they're missing, the shell configuration wasn't applied correctly. Verify:

1. Check that the pathman configuration exists in your shell profile:
   ```bash
   grep -A 10 "BEGIN PATHMAN CONFIG" ~/.profile
   ```

2. Make sure your shell actually sources that file. For bash, check that `.bash_profile` or `.bashrc` sources `.profile`:
   ```bash
   cat ~/.bash_profile
   ```

3. Re-run `pathman init` and accept the automatic configuration.

---

## Permission Denied When Adding Executable

**Symptom**: Running `pathman add` fails with a permission error:
```
Error: failed to create symlink: permission denied
```

**Cause**: The managed folder has incorrect permissions or ownership.

**Solution**: Fix the permissions:
```bash
chmod 755 ~/.local/bin/pathman-links
chmod 755 ~/.local/bin/pathman-links/front
chmod 755 ~/.local/bin/pathman-links/back
```

If you're still having issues, check ownership:
```bash
ls -la ~/.local/bin/pathman-links
```

If it's owned by a different user (e.g., root), take ownership back:
```bash
sudo chown -R $USER:$USER ~/.local/bin/pathman-links
```

---

## Warning About Insecure Permissions

**Symptom**: `pathman init` warns:
```
WARNING: Folder has insecure permissions: 0777
Group or others have write permission. This is a security risk.
```

**Cause**: The folder permissions are too permissive, allowing anyone to modify your executables.

**Why It Matters**: If others can write to your PATH folders, they could replace your executables with malicious ones.

**Solution**: Fix the permissions:
```bash
chmod 755 ~/.local/bin/pathman-links
chmod 755 ~/.local/bin/pathman-links/front
chmod 755 ~/.local/bin/pathman-links/back
```

This sets permissions to `rwxr-xr-x` (owner can write, all can read/execute).

---

## Symlink Points to Wrong Location

**Symptom**: A symlink was created but points to the wrong target, or becomes broken after moving the original executable.

**Cause**: The original executable was moved or deleted, or you specified the wrong path when adding it.

**Solution**: Remove and re-add the symlink with the correct path:
```bash
pathman remove myapp
pathman add /correct/path/to/myapp
```

---

## Name Clash: Command Exists in Multiple Locations

**Symptom**: `pathman add` shows a warning:
```
Warning: 'python' in back subfolder will be masked by /usr/bin/python
```

**Meaning**: The executable you're adding will never be used because something earlier in your PATH has the same name.

**Solution Options**:

1. **Add to front instead** to override the system version:
   ```bash
   pathman add /path/to/python --priority=front
   ```

2. **Use a different name** to avoid conflicts:
   ```bash
   pathman add /path/to/python --name python3.11
   ```

3. **Accept the masking** if you want it as a fallback (use `--force`):
   ```bash
   pathman add /path/to/python --priority=back --force
   ```

---

## Configuration File Corrupted

**Symptom**: Pathman fails with errors about invalid configuration:
```
Error: failed to load config: invalid character '}' looking for beginning of value
```

**Cause**: The `~/.config/pathman/config.json` file was manually edited and is now invalid JSON.

**Solution**: 

1. **Backup the corrupted file**:
   ```bash
   cp ~/.config/pathman/config.json ~/.config/pathman/config.json.backup
   ```

2. **Reset to empty configuration**:
   ```bash
   echo '{"managed_directories":[]}' > ~/.config/pathman/config.json
   ```

3. **Re-add your directories**:
   ```bash
   pathman add ~/.cargo/bin --priority=back
   pathman add ~/go/bin --priority=back
   ```

---

## Broken Symlinks After System Upgrade

**Symptom**: After upgrading your system or reinstalling applications, some executables don't work.

**Cause**: The original executables were moved or removed during the upgrade.

**Solution**: Use `pathman clean` to detect and remove broken symlinks:
```bash
pathman clean
```

This will show you all broken symlinks and let you select which ones to remove.

---

## `pathman path` Returns Nothing

**Symptom**: Running `pathman path` produces no output or an error.

**Diagnosis**: Check if pathman is working at all:
```bash
pathman summary
```

**Solution**: If pathman itself works but `pathman path` fails:

1. Check if your current PATH is set:
   ```bash
   echo $PATH
   ```

2. Try running pathman path manually to see the error:
   ```bash
   pathman path
   echo $?  # Check exit code
   ```

3. If it's failing due to missing folders, create them:
   ```bash
   pathman init
   ```

---

## Can't Remove Executable

**Symptom**: `pathman remove myapp` says "not found" but you can see it in the folder.

**Cause**: The symlink might be in the managed folder but with a different name.

**Solution**: List all managed executables to find the actual name:
```bash
pathman list --long
```

Then remove using the correct name:
```bash
pathman remove actual-name
```

---

## Fish Shell: PATH Not Updated

**Symptom**: On fish shell, pathman-managed executables aren't found.

**Cause**: Fish uses a different PATH format than bash/zsh.

**Solution**: Ensure your `~/.config/fish/config.fish` contains:
```fish
# Added by pathman
if command -v pathman >/dev/null 2>&1
  set NEW_PATH (pathman path 2>/dev/null)
  if test -n "$NEW_PATH"
    set -gx PATH (string split : $NEW_PATH)
  end
end
```

Then reload fish:
```fish
source ~/.config/fish/config.fish
```

---

## Directory Not Added to PATH

**Symptom**: You ran `pathman add ~/.cargo/bin` but the executables in that directory aren't found.

**Diagnosis**: Check if the directory is in the config:
```bash
cat ~/.config/pathman/config.json
```

Check if it's in the generated PATH:
```bash
pathman path | tr ':' '\n' | grep cargo
```

**Solution**: If it's not in config, the add might have failed. Try again:
```bash
pathman add ~/.cargo/bin --priority=back
```

If it's in config but not in the generated PATH, reload your shell configuration.

---

## Executables Work in Terminal but Not in Scripts

**Symptom**: Pathman-managed executables work when you type commands interactively, but fail in shell scripts or cron jobs.

**Cause**: Scripts and cron jobs often run in non-interactive shells that don't source your `.profile` or `.bashrc`.

**Solution**: Add pathman's PATH generation to your script:
```bash
#!/bin/bash
export PATH=$(pathman path)

# Now pathman-managed executables are available
myapp --do-something
```

For cron jobs, set PATH at the top of your crontab:
```cron
PATH=/usr/bin:/bin
SHELL=/bin/bash

# Before your cron commands, add:
0 * * * * export PATH=$(pathman path) && myapp --do-something
```

---

## Still Having Issues?

If none of these solutions work:

1. Check the [architecture documentation](docs/architecture.md) to understand how pathman works
2. Run `pathman summary` to see the current state
3. Check the [shell integration guide](docs/shell-integration.md) for shell-specific setup
4. Open an issue on GitHub with:
   - Your operating system and version
   - Your shell and version (`echo $SHELL` and `$SHELL --version`)
   - Output of `pathman summary`
   - The exact error message
   - Steps to reproduce
