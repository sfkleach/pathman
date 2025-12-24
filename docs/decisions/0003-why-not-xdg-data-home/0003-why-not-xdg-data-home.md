# 0003 - Why Not Use $XDG_DATA_HOME for Managed Links, 2025-12-24

## Issue

Decide where to place pathman's managed symlink folders: follow XDG Base Directory specification or use a more traditional location.

## Factors

- User expectations and discoverability
- Standards compliance
- Ease of browsing with standard tools
- Semantic appropriateness
- Proximity to traditional executable locations

## Options and Outcome

I considered using `~/.local/share/pathman/links` (following XDG) versus `~/.local/bin/pathman-links` (closer to traditional bin directories). I chose the latter because these are executables that users may want to browse directly, not application data.

## Pros and Cons of Options

### Option 1: XDG_DATA_HOME (~/.local/share/pathman)

**Pros:**
- Follows XDG Base Directory specification
- Standard location for application data
- Keeps all pathman data in one place
- Respects XDG_DATA_HOME environment variable

**Cons:**
- Semantically wrong - executables aren't "application data"
- Users wouldn't expect to find executables in `.local/share`
- Harder to discover with `ls` when troubleshooting
- Further from the traditional `~/.local/bin` location users know
- Makes it less obvious what pathman is doing

### Option 2: ~/.local/bin/pathman-links (Selected)

**Pros:**
- Semantically correct - these are executable links, not data files
- Users expect executables near `~/.local/bin`
- Easy to discover and browse with standard tools (`ls`, file managers)
- Closer to traditional locations users understand
- Makes pathman's purpose immediately clear
- Users can `cd ~/.local/bin/pathman-links/front` to inspect

**Cons:**
- Doesn't follow XDG Base Directory specification
- Slightly longer path than if using XDG defaults
- Mixes "managed" location with "unmanaged" parent

## Additional Notes

This decision came down to semantic appropriateness and user expectations.

The XDG Base Directory specification defines `XDG_DATA_HOME` (defaulting to `~/.local/share`) for "user-specific data files". But pathman's symlinks aren't data files - they're executable links that belong on $PATH.

Key considerations:

1. **Discoverability**: When users have a problem, `ls ~/.local/bin` is one of the first places they'll look. Having pathman-links as a subdirectory there makes sense.

2. **Mental model**: Users think of `~/.local/bin` as "where my executables go". Pathman managing a subfolder there aligns with that mental model better than hiding executables in `.local/share`.

3. **Browse-ability**: Users may want to `ls ~/.local/bin/pathman-links/front` to see what's there. That's natural in a bin directory, weird in a data directory.

4. **Proximity to alternatives**: If pathman didn't exist, users would put symlinks directly in `~/.local/bin`. Pathman just organizes them into `pathman-links/front` and `pathman-links/back` subfolders. It feels like a natural extension rather than a completely different location.

5. **Semantic correctness**: The XDG spec even says "should be analogous to /usr/local/share". System-wide executables go in `/usr/local/bin`, not `/usr/local/share`.

The configuration file (`~/.config/pathman/config.json`) does follow XDG specs by using `XDG_CONFIG_HOME`. That's appropriate because it's actual configuration data.

So pathman follows a hybrid approach:
- Configuration: `~/.config/pathman/` (follows XDG)
- Executable symlinks: `~/.local/bin/pathman-links/` (semantic correctness over XDG)

This gives users the best of both worlds - config where they expect config, executables where they expect executables.
