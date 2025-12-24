# Interactive clean-up

**Goal**: add a new subcommand `clean` that supports the interactive removal
of broken symlinks and missing folders.

**Criteria**: 
- It must be easy to choose individual items to clean up.
- It must be easy to choose all the items.
- It must be easy to unselect individual items to clean up.
- The user must confirm the changes.
- All the changes are summarised before being confirmed.

**Implementation suggestion**:
- Use bubbletea to provide an interactive display.
