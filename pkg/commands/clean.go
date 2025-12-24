package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/sfkleach/pathman/pkg/folder"
)

// NewCleanCmd creates the clean command.
func NewCleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Interactively remove broken symlinks and missing directories",
		Long: `Scans for broken symlinks in the front/back folders and missing managed directories.
Presents an interactive interface for selecting which items to remove.`,
		RunE: runClean,
	}
}

// cleanModel represents the state of the interactive clean UI.
type cleanModel struct {
	items   []folder.CleanupItem
	cursor  int
	done    bool
	confirm bool
	err     error
	width   int
	height  int
}

func initialModel(items []folder.CleanupItem) cleanModel {
	return cleanModel{
		items:   items,
		cursor:  0,
		done:    false,
		confirm: false,
	}
}

func (m cleanModel) Init() tea.Cmd {
	return nil
}

func (m cleanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.confirm {
			// In confirmation screen.
			switch msg.String() {
			case "y", "Y":
				// Perform cleanup.
				err := folder.PerformCleanup(m.items)
				if err != nil {
					m.err = err
				}
				m.done = true
				return m, tea.Quit
			case "n", "N", "q", "ctrl+c", "esc":
				// Cancel.
				m.done = true
				return m, tea.Quit
			}
		} else {
			// In selection screen.
			switch msg.String() {
			case "ctrl+c", "q":
				m.done = true
				return m, tea.Quit

			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}

			case "down", "j":
				if m.cursor < len(m.items)-1 {
					m.cursor++
				}

			case " ":
				// Toggle selection.
				if len(m.items) > 0 {
					m.items[m.cursor].Selected = !m.items[m.cursor].Selected
				}

			case "a", "A":
				// Select all.
				for i := range m.items {
					m.items[i].Selected = true
				}

			case "d", "D":
				// Deselect all.
				for i := range m.items {
					m.items[i].Selected = false
				}

			case "enter":
				// Show confirmation.
				m.confirm = true
			}
		}
	}

	return m, nil
}

func (m cleanModel) View() string {
	if m.done {
		if m.err != nil {
			return fmt.Sprintf("Error during cleanup: %v\n", m.err)
		}

		selectedCount := 0
		for _, item := range m.items {
			if item.Selected {
				selectedCount++
			}
		}

		if selectedCount == 0 {
			return "No items selected. Nothing to clean up.\n"
		}

		return fmt.Sprintf("Successfully cleaned up %d item(s).\n", selectedCount)
	}

	if m.confirm {
		return m.confirmView()
	}

	return m.selectionView()
}

func (m cleanModel) selectionView() string {
	var b strings.Builder

	b.WriteString("Pathman Clean - Select items to remove\n\n")

	if len(m.items) == 0 {
		b.WriteString("No cleanup items found. Your pathman installation is clean!\n\n")
		b.WriteString("Press q to quit.\n")
		return b.String()
	}

	// Show items.
	for i, item := range m.items {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if item.Selected {
			checked = "✓"
		}

		b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, checked, item.Description))
	}

	b.WriteString("\n")

	// Show summary.
	selectedCount := 0
	symlinkCount := 0
	dirCount := 0
	for _, item := range m.items {
		if item.Selected {
			selectedCount++
			if item.Type == "symlink" {
				symlinkCount++
			} else {
				dirCount++
			}
		}
	}

	if selectedCount > 0 {
		b.WriteString(fmt.Sprintf("Selected: %d item(s)", selectedCount))
		if symlinkCount > 0 && dirCount > 0 {
			b.WriteString(fmt.Sprintf(" (%d symlink(s), %d director(ies))", symlinkCount, dirCount))
		} else if symlinkCount > 0 {
			b.WriteString(fmt.Sprintf(" (%d symlink(s))", symlinkCount))
		} else if dirCount > 0 {
			b.WriteString(fmt.Sprintf(" (%d director(ies))", dirCount))
		}
		b.WriteString("\n\n")
	} else {
		b.WriteString("No items selected\n\n")
	}

	// Show controls.
	b.WriteString("Controls:\n")
	b.WriteString("  ↑/k, ↓/j: Move cursor\n")
	b.WriteString("  Space: Toggle selection\n")
	b.WriteString("  a: Select all\n")
	b.WriteString("  d: Deselect all\n")
	b.WriteString("  Enter: Confirm and clean up\n")
	b.WriteString("  q: Quit\n")

	return b.String()
}

func (m cleanModel) confirmView() string {
	var b strings.Builder

	b.WriteString("Confirm Cleanup\n\n")

	selectedItems := []folder.CleanupItem{}
	for _, item := range m.items {
		if item.Selected {
			selectedItems = append(selectedItems, item)
		}
	}

	if len(selectedItems) == 0 {
		b.WriteString("No items selected. Nothing to clean up.\n\n")
		b.WriteString("Press any key to return.\n")
		return b.String()
	}

	b.WriteString("The following items will be removed:\n\n")

	for _, item := range selectedItems {
		if item.Type == "symlink" {
			b.WriteString(fmt.Sprintf("  • Symlink: %s\n", item.Description))
		} else {
			b.WriteString(fmt.Sprintf("  • Directory (from config): %s\n", item.Description))
		}
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Total: %d item(s)\n\n", len(selectedItems)))
	b.WriteString("Are you sure you want to proceed? (y/n): ")

	return b.String()
}

func runClean(cmd *cobra.Command, args []string) error {
	// Find cleanup items.
	items, err := folder.FindCleanupItems()
	if err != nil {
		return fmt.Errorf("failed to find cleanup items: %w", err)
	}

	// Run interactive UI.
	p := tea.NewProgram(initialModel(items))
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running interactive UI: %w", err)
	}

	// Check for errors.
	if m, ok := finalModel.(cleanModel); ok && m.err != nil {
		return m.err
	}

	return nil
}
