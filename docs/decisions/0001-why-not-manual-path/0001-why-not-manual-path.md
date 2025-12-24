# 0001 - Why Not Just Add Directories to $PATH, 2025-12-24

## Issue

Decide whether to build a tool for managing PATH or just document how to add directories manually.

## Factors

- User experience and ease of use
- Ability to track what has been added
- Clean removal of entries
- Priority management (front vs back of PATH)
- Cross-shell compatibility

## Options and Outcome

I considered whether pathman was even necessary, or if users should just manually add directories to their shell configuration. I decided to build pathman because manual PATH management is error-prone, difficult to track, and hard to maintain over time.

## Pros and Cons of Options

### Option 1: Manual PATH Management

**Pros:**
- No additional tool required
- Direct control over shell configuration
- Works immediately without installation
- Simple for single additions

**Cons:**
- Must edit shell config files directly (which file? .profile? .bashrc? .bash_profile?)
- Hard to track what you've added over time
- Difficult to remove entries cleanly (find and edit the right line)
- Can't easily reorder priorities without rewriting PATH
- Different syntax for different shells (bash vs zsh vs fish)
- Easy to make syntax errors that break your shell
- Duplicate entries accumulate over time
- No validation that directories exist or are safe

### Option 2: Pathman Tool (Selected)

**Pros:**
- Single command to add/remove executables or directories
- Tracks all managed items in one place
- Clean removal operations
- Easy priority management (front vs back)
- Single line in shell config works across all shells
- Validates paths before adding
- Detects and warns about PATH masking
- Can detect and remove broken symlinks
- Provides health summary of managed PATH

**Cons:**
- Requires installing an additional tool
- Adds one level of indirection (calls `pathman path`)
- Users must understand the pathman model

## Additional Notes

The decision came down to user experience. While manually adding to PATH works for a single directory, it breaks down at scale:

1. **Discoverability**: After months or years, users forget what they've added to PATH and why
2. **Clean removal**: Finding and removing the right line from shell config is tedious
3. **Conflict detection**: Manual additions can't warn when they'll be masked or mask other executables
4. **Priority management**: Moving something from back to front requires rewriting PATH entries

Pathman solves these problems by:
- Maintaining a single source of truth (managed folders + config.json)
- Providing simple commands for all operations
- Automatically checking for conflicts
- Making priority changes trivial (`pathman set`)

The single line in shell config (`export PATH=$(pathman path)`) is actually simpler than accumulating multiple PATH modifications over time. Users only need to understand "pathman manages my PATH" rather than "where should I add this in my PATH string?"

For pathman itself, this approach also enables features that would be impossible with manual management:
- `pathman summary` - see the health of your PATH at a glance
- `pathman clean` - interactively remove broken items
- `pathman list` - see what's managed
- Masking detection - warn before creating problems

The tool pays for its complexity by making the common operations trivial and the complex operations possible.
