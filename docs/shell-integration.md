# Shell Integration

This guide covers how to integrate Pathman with different shells and shell configurations.

## Overview

Pathman requires a single line in your shell configuration that calls `pathman path` to generate your complete PATH. This approach keeps your shell config simple while letting pathman manage all the complexity.

The generated PATH combines:
1. Front subfolder (high-priority symlinks)
2. Front-priority directories
3. Your existing $PATH
4. Back-priority directories
5. Back subfolder (low-priority symlinks)

## Automatic Setup (Bash Only)

The easiest way to set up pathman with bash is:

```bash
pathman init
```

This will:
1. Create the managed folders
2. Detect that you're using bash
3. Offer to add the configuration automatically
4. Update your `~/.profile` or `~/.bash_profile`

Just accept the prompt and restart your terminal.

## Manual Setup

If you prefer manual setup or use a different shell, follow the instructions below.

---

## Bash

### Standard Setup

Add to `~/.profile` (or `~/.bash_profile` on macOS):

```bash
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

### Interactive Shells Only

If you only want pathman in interactive shells (not login scripts), add to `~/.bashrc`:

```bash
if [[ $- == *i* ]]; then
  # Only in interactive shells
  if command -v pathman >/dev/null 2>&1; then
    NEW_PATH=$(pathman path 2>/dev/null)
    if [ $? -eq 0 ] && [ -n "$NEW_PATH" ]; then
      export PATH="$NEW_PATH"
    fi
  fi
fi
```

### Bash on macOS

macOS Terminal.app typically uses login shells, so use `~/.bash_profile`:

```bash
# Source ~/.profile if it exists
if [ -f ~/.profile ]; then
  source ~/.profile
fi

# Add pathman configuration here (if not in ~/.profile)
```

---

## Zsh

### Standard Setup

Add to `~/.zshrc`:

```zsh
# Added by pathman
if command -v pathman >/dev/null 2>&1; then
  NEW_PATH=$(pathman path 2>/dev/null)
  if [ $? -eq 0 ] && [ -n "$NEW_PATH" ]; then
    export PATH="$NEW_PATH"
  fi
fi
```

### Zsh on macOS

Zsh on macOS works the same way. Add the configuration to `~/.zshrc`.

### Using Zsh Plugin Managers

If you use Oh My Zsh or another plugin manager, add the pathman configuration after all plugins are loaded, typically near the end of your `~/.zshrc`.

---

## Fish

### Standard Setup

Add to `~/.config/fish/config.fish`:

```fish
# Added by pathman
if command -v pathman >/dev/null 2>&1
  set NEW_PATH (pathman path 2>/dev/null)
  if test -n "$NEW_PATH"
    set -gx PATH (string split : $NEW_PATH)
  end
end
```

### Fish Universal Variables

Fish can use universal variables that persist across sessions. However, since pathman's PATH changes based on configuration, we recommend the dynamic approach above that regenerates PATH each time.

If you prefer a universal variable approach:

```fish
# Set once, manually
set -Ux PATH (pathman path)
```

**Note**: You'll need to re-run this command whenever you add or remove executables with pathman.

---

## Other Shells

### tcsh/csh

Add to `~/.tcshrc` or `~/.cshrc`:

```csh
if ( -X pathman ) then
  set path = ( `pathman path | sed 's/:/ /g'` )
endif
```

### ksh (Korn Shell)

Add to `~/.kshrc`:

```ksh
if command -v pathman >/dev/null 2>&1; then
  export PATH=$(pathman path)
fi
```

---

## Advanced Configurations

### Multiple Shell Profiles

If you use multiple shells, you can create a shared configuration:

1. Create `~/.pathmanrc`:
   ```bash
   # Pathman configuration - sourced by multiple shells
   if command -v pathman >/dev/null 2>&1; then
     export PATH=$(pathman path)
   fi
   ```

2. Source it from each shell's config:
   ```bash
   # In ~/.bashrc, ~/.zshrc, etc.
   source ~/.pathmanrc
   ```

### Conditional Pathman Activation

Only activate pathman on certain systems:

```bash
# Only use pathman on personal machines
if [ "$HOSTNAME" = "my-laptop" ]; then
  if command -v pathman >/dev/null 2>&1; then
    export PATH=$(pathman path)
  fi
fi
```

### Performance Optimization

If `pathman path` is too slow for your needs, you can cache the result:

```bash
# Cache pathman PATH for this session
if [ -z "$PATHMAN_PATH_CACHED" ]; then
  export PATH=$(pathman path)
  export PATHMAN_PATH_CACHED=1
fi
```

**Warning**: This means changes made with `pathman add` or `pathman remove` won't take effect until you restart your shell.

---

## Non-Interactive Shells

### Shell Scripts

For shell scripts that need pathman-managed executables:

```bash
#!/bin/bash
# Add at the top of your script
export PATH=$(pathman path)

# Now pathman executables are available
my-custom-tool --flag value
```

### Cron Jobs

Cron runs with a minimal environment. To use pathman-managed executables in cron:

**Option 1**: Set PATH in the script:
```bash
#!/bin/bash
export PATH=$(pathman path)
exec /path/to/actual/script.sh
```

**Option 2**: Use full paths:
```bash
# In crontab
0 * * * * $HOME/.local/bin/pathman-links/front/my-tool --do-something
```

**Option 3**: Set PATH in crontab:
```cron
SHELL=/bin/bash
PATH=/usr/bin:/bin

0 * * * * export PATH=$(pathman path) && my-tool --do-something
```

### systemd Services

For systemd services, set the full PATH in the service unit:

```ini
[Service]
ExecStart=/bin/bash -c 'export PATH=$(pathman path) && my-service'
```

Or use the full path to the executable:

```ini
[Service]
ExecStart=/home/user/.local/bin/pathman-links/front/my-service
```

---

## SSH Sessions

### Remote SSH

When SSHing to a remote machine, your shell typically sources the login profile:

```bash
ssh user@host
# ~/.profile or ~/.bash_profile will be sourced, including pathman config
```

### SSH Commands

When running a single command over SSH, the shell is often non-interactive and doesn't source profiles:

```bash
ssh user@host my-tool  # May not find pathman executables
```

**Solution**: Use a login shell:
```bash
ssh -t user@host 'bash -l -c "my-tool"'
```

Or explicitly set PATH:
```bash
ssh user@host 'export PATH=$(pathman path) && my-tool'
```

---

## IDE and Editor Integration

### VS Code

VS Code's integrated terminal uses your default shell configuration, so pathman will work automatically if configured in your shell profile.

For VS Code tasks and launch configurations, the PATH may differ. Set it explicitly:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Run Tool",
      "type": "shell",
      "command": "my-tool",
      "options": {
        "env": {
          "PATH": "${env:HOME}/.local/bin/pathman-links/front:${env:HOME}/.local/bin/pathman-links/back:${env:PATH}"
        }
      }
    }
  ]
}
```

### JetBrains IDEs (IntelliJ, PyCharm, etc.)

JetBrains IDEs may not inherit your shell's PATH. Set it in Settings:

1. Go to Settings → Tools → Terminal
2. Set "Shell path" to your shell with the login option:
   - Bash: `/bin/bash -l`
   - Zsh: `/bin/zsh -l`

---

## Troubleshooting

### PATH Not Updated

If executables still aren't found after setup:

1. Verify the configuration is in the right file for your shell
2. Check if that file is actually sourced (add a `echo "Profile loaded"` temporarily)
3. Ensure pathman itself is in your PATH (try `which pathman`)
4. Restart your terminal or run `source ~/.profile` (or equivalent)

### Slow Shell Startup

If pathman makes your shell start slowly:

1. Time it: `time pathman path` (should be <50ms)
2. Check for NFS or network-mounted home directories
3. Consider caching (see Performance Optimization above)

### Changes Not Taking Effect

After running `pathman add` or `pathman remove`, you need to:

1. Open a new terminal, or
2. Reload your shell configuration: `source ~/.profile`

The current shell session has the old PATH cached.

---

## Best Practices

1. **Use login profiles**: Put pathman config in `~/.profile` or `~/.bash_profile`, not `~/.bashrc`
2. **Check before using**: Always test `command -v pathman` before calling it
3. **Handle errors**: Suppress errors in non-interactive shells to avoid breaking scripts
4. **Keep it simple**: The one-line approach is easier to maintain than manual PATH editing
5. **Document it**: If you customize the setup, add comments explaining why

---

## See Also

- [Troubleshooting Guide](troubleshooting.md) for common issues
- [Architecture Documentation](architecture.md) for how pathman works internally
- [README](../README.md) for general usage information
