package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/sfkleach/pathman/pkg/folder"
)

// NewInitCmd creates the init command.
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create the managed folder",
		Long: `Create the managed folder with appropriate permissions.
If the folder already exists, check its permissions and warn if insecure.`,
		Args: cobra.NoArgs,
		RunE: runInit,
	}
}

// initModel represents the state of the init UI.
type initModel struct {
	stage              string // "setup", "prompt", "processing", "done"
	message            []string
	cursor             int
	choices            []string
	selected           int // -1 for no selection
	err                error
	shouldAddToProfile bool
}

func initialInitModel() initModel {
	return initModel{
		stage:    "setup",
		message:  []string{},
		choices:  []string{"Yes, add to profile", "No, I'll do it manually"},
		selected: -1,
	}
}

func (m initModel) Init() tea.Cmd {
	return performSetup
}

type setupCompleteMsg struct {
	message        []string
	needsPathSetup bool
	isBashor       bool
	profilePath    string
	err            error
}

func performSetup() tea.Msg {
	var messages []string

	basePath, err := folder.GetManagedFolder()
	if err != nil {
		return setupCompleteMsg{err: fmt.Errorf("failed to get managed folder path: %w", err)}
	}

	frontPath, backPath, err := folder.GetBothSubfolders()
	if err != nil {
		return setupCompleteMsg{err: fmt.Errorf("failed to get subfolder paths: %w", err)}
	}

	// Check/create base folder.
	baseCreated := false
	if folder.Exists(basePath) {
		info, err := os.Stat(basePath)
		if err != nil {
			return setupCompleteMsg{err: fmt.Errorf("failed to stat folder: %w", err)}
		}

		perm := info.Mode().Perm()
		if perm&0022 != 0 {
			messages = append(messages,
				fmt.Sprintf("Managed folder already exists: %s", basePath),
				fmt.Sprintf("WARNING: Folder has insecure permissions: %04o", perm),
				"Group or others have write permission. This is a security risk.",
				"Recommended permissions: 0755 (owner read/write/execute, all read/execute)",
			)
		} else {
			messages = append(messages,
				fmt.Sprintf("Managed folder already exists: %s", basePath),
				fmt.Sprintf("Permissions are correct: %04o", perm),
			)
		}
	} else {
		// #nosec G301 -- 0755 permissions are appropriate for PATH directories that need to be accessible by different users
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return setupCompleteMsg{err: fmt.Errorf("failed to create folder: %w", err)}
		}
		messages = append(messages,
			fmt.Sprintf("Created managed folder: %s", basePath),
			"Permissions set to: 0755 (owner read/write/execute, all read/execute)",
		)
		baseCreated = true
	}

	// Create front subfolder.
	frontCreated := false
	if !folder.Exists(frontPath) {
		// #nosec G301 -- 0755 permissions are appropriate for PATH directories that need to be accessible by different users
		if err := os.MkdirAll(frontPath, 0755); err != nil {
			return setupCompleteMsg{err: fmt.Errorf("failed to create front subfolder: %w", err)}
		}
		messages = append(messages, fmt.Sprintf("Created front subfolder: %s", frontPath))
		frontCreated = true
	}

	// Create back subfolder.
	backCreated := false
	if !folder.Exists(backPath) {
		// #nosec G301 -- 0755 permissions are appropriate for PATH directories that need to be accessible by different users
		if err := os.MkdirAll(backPath, 0755); err != nil {
			return setupCompleteMsg{err: fmt.Errorf("failed to create back subfolder: %w", err)}
		}
		messages = append(messages, fmt.Sprintf("Created back subfolder: %s", backPath))
		backCreated = true
	}

	// Check if subfolders are on $PATH.
	frontOnPath := folder.IsOnPath(frontPath)
	backOnPath := folder.IsOnPath(backPath)

	if !frontOnPath || !backOnPath {
		messages = append(messages,
			"",
			"The managed subfolders are not properly configured in your $PATH.",
			"To use executables in these folders, you need to add them to your $PATH.",
		)

		// Check if the user is using bash.
		shell := os.Getenv("SHELL")
		if strings.Contains(shell, "bash") {
			profilePath, err := folder.GetBashProfilePath()
			if err != nil {
				return setupCompleteMsg{err: fmt.Errorf("failed to get profile path: %w", err)}
			}

			profileName := filepath.Base(profilePath)
			messages = append(messages,
				"",
				fmt.Sprintf("Since you're using bash, this is normally done by adding a line to your ~/%s file.", profileName),
			)

			return setupCompleteMsg{
				message:        messages,
				needsPathSetup: true,
				isBashor:       true,
				profilePath:    profilePath,
			}
		}

		// Non-bash shell - just show instructions.
		messages = append(messages,
			"",
			"To add it to your PATH, add these lines to your shell configuration:",
			"",
			"# ============ BEGIN PATHMAN CONFIG ============",
			"# Added by pathman",
			"if command -v pathman >/dev/null 2>&1; then",
			"  # Calculate a new $PATH from the old one and pathman's configuration.",
			"  NEW_PATH=$(pathman path 2>/dev/null)",
			"  if [ $? -eq 0 ] && [ -n \"$NEW_PATH\" ]; then",
			"    export PATH=\"$NEW_PATH\"",
			"  elif [ -n \"$PS1\" ]; then",
			"    # PS1 is only set in interactive shells - safe to show errors here.",
			"    echo \"Warning: pathman failed to update PATH\" >&2",
			"  fi",
			"elif [ -n \"$PS1\" ]; then",
			"  # PS1 is only set in interactive shells - safe to show errors here.",
			"  echo \"Warning: pathman not found, PATH not updated\" >&2",
			"fi",
			"# ============= END PATHMAN CONFIG =============",
		)

		return setupCompleteMsg{
			message:        messages,
			needsPathSetup: false,
		}
	} else if baseCreated || frontCreated || backCreated {
		messages = append(messages,
			"",
			"The managed folder is already properly configured in your $PATH.",
		)
	}

	return setupCompleteMsg{
		message:        messages,
		needsPathSetup: false,
	}
}

type profileUpdateMsg struct {
	err error
}

func updateProfile() tea.Msg {
	if err := folder.AddToProfile(); err != nil {
		return profileUpdateMsg{err: err}
	}
	return profileUpdateMsg{}
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case setupCompleteMsg:
		if msg.err != nil {
			m.err = msg.err
			m.stage = "done"
			return m, tea.Quit
		}

		m.message = msg.message

		if msg.needsPathSetup && msg.isBashor {
			m.stage = "prompt"
		} else {
			m.stage = "done"
			return m, tea.Quit
		}

	case profileUpdateMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.message = append(m.message,
				"",
				"Successfully added pathman configuration to your profile.",
				"Please restart your shell or run 'source ~/.profile' to apply changes.",
			)
		}
		m.stage = "done"
		return m, tea.Quit

	case tea.KeyMsg:
		if m.stage == "prompt" {
			switch msg.String() {
			case "ctrl+c", "q":
				m.stage = "done"
				return m, tea.Quit

			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}

			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}

			case "enter", " ":
				m.selected = m.cursor
				m.shouldAddToProfile = (m.cursor == 0)
				m.stage = "processing"

				if m.shouldAddToProfile {
					return m, updateProfile
				}

				// User chose manual setup - show instructions.
				profilePath, _ := folder.GetBashProfilePath()
				profileName := filepath.Base(profilePath)
				m.message = append(m.message,
					"",
					fmt.Sprintf("To add it manually, add these lines to your ~/%s:", profileName),
					"",
					"# ============ BEGIN PATHMAN CONFIG ============",
					"# Added by pathman",
					"if command -v pathman >/dev/null 2>&1; then",
					"  # Calculate a new $PATH from the old one and pathman's configuration.",
					"  NEW_PATH=$(pathman path 2>/dev/null)",
					"  if [ $? -eq 0 ] && [ -n \"$NEW_PATH\" ]; then",
					"    export PATH=\"$NEW_PATH\"",
					"  elif [ -n \"$PS1\" ]; then",
					"    # PS1 is only set in interactive shells - safe to show errors here.",
					"    echo \"Warning: pathman failed to update PATH\" >&2",
					"  fi",
					"elif [ -n \"$PS1\" ]; then",
					"  # PS1 is only set in interactive shells - safe to show errors here.",
					"  echo \"Warning: pathman not found, PATH not updated\" >&2",
					"fi",
					"# ============= END PATHMAN CONFIG =============",
				)
				m.stage = "done"
				return m, tea.Quit
			}
		} else if m.stage == "done" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m initModel) View() string {
	var b strings.Builder

	// Show all messages.
	for _, msg := range m.message {
		b.WriteString(msg)
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString(fmt.Sprintf("\nError: %v\n", m.err))
		return b.String()
	}

	switch m.stage {
	case "prompt":
		b.WriteString("\nWould you like me to add the PATH configuration for you?\n\n")

		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			b.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
		}

		b.WriteString("\nControls: ↑/k, ↓/j to move, Enter/Space to select, q to quit\n")

	case "processing":
		b.WriteString("\nProcessing...\n")

	case "done":
		// Messages already printed above.
	}

	return b.String()
}

func runInit(cmd *cobra.Command, args []string) error {
	p := tea.NewProgram(initialInitModel())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running interactive UI: %w", err)
	}

	if m, ok := finalModel.(initModel); ok && m.err != nil {
		return m.err
	}

	return nil
}
