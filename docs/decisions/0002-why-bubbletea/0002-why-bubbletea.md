# 0002 - Why Bubbletea for TUI, 2025-12-24

## Issue

Decide on the framework for implementing interactive terminal user interfaces in pathman.

## Factors

- Modern and well-maintained
- Clear state management model
- Consistent keyboard controls across features
- Easy to extend for future interactive features
- Good documentation and examples

## Options and Outcome

I considered several approaches for the interactive features (`pathman clean` and `pathman init`), including raw terminal manipulation, existing TUI libraries, and Bubbletea. I selected Bubbletea for its clean Elm-architecture design and excellent Go integration.

## Pros and Cons of Options

### Option 1: Raw Terminal Control (termbox-go, tcell)

**Pros:**
- Full control over terminal behavior
- Minimal dependencies
- Fast and lightweight
- Can do anything you want

**Cons:**
- Must manually manage state and rendering
- Complex keyboard handling code
- Easy to introduce bugs in state transitions
- Tedious to implement common patterns (selection, scrolling)
- Hard to maintain and extend

### Option 2: Traditional TUI Libraries (tview, gocui)

**Pros:**
- Widget-based approach
- Built-in components (lists, forms, etc.)
- Mature and stable

**Cons:**
- Callback-based state management can get messy
- Global state often required
- Harder to reason about complex interactions
- Less modern programming model

### Option 3: Bubbletea (Selected)

**Pros:**
- Elm-architecture (Model-Update-View) makes state management clean
- Pure functions for updates - easy to test and reason about
- Immutable state reduces bugs
- Excellent documentation and examples
- Active development and community
- Integrates well with Lip Gloss for styling
- Consistent patterns across all interactive features
- Easy to compose complex UIs from simple components

**Cons:**
- Relatively newer (less battle-tested than alternatives)
- Requires understanding the Elm architecture
- More conceptual overhead than imperative approaches

## Additional Notes

The Elm architecture that Bubbletea uses maps perfectly to the interactive features we needed:

**For `pathman clean`:**
- **Model**: List of cleanup items with selection state
- **Update**: Handle keyboard events to toggle selections
- **View**: Render the current state as text

**For `pathman init`:**
- **Model**: Setup progress, prompt state, choice selection
- **Update**: Handle setup completion, keyboard navigation
- **View**: Show progress and prompt appropriately

The functional approach makes testing straightforward - you can test the update logic without actually running a terminal. The immutable state means you can't accidentally corrupt the UI state through side effects.

The consistency across features was a major win. Both commands use the same keyboard conventions:
- Arrow keys or k/j for navigation
- Space for selection
- Enter to confirm
- q to quit

This consistency comes naturally from Bubbletea's model - all commands implement the same interface and handle the same message types.

For future interactive features (if we add them), Bubbletea provides a proven pattern to follow. The investment in learning the architecture pays off with maintainability and extensibility.
