package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sfkleach/pathman/pkg/folder"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

// Source is set at build time via ldflags to the git repository URL.
var Source = "https://github.com/sfkleach/pathman"

// NewRootCmd creates the root command for pathman.
func NewRootCmd() *cobra.Command {
	var versionFlag bool

	cmd := &cobra.Command{
		Use:   "pathman",
		Short: "Pathman manages executables on your $PATH",
		Long: `Pathman is a command-line tool that helps you manage the list of applications
accessible by $PATH. With pathman, you can add, remove, and list executables
in two managed folders (front and back of $PATH).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				fmt.Printf("pathman version %s\n", Version)
				return nil
			}
			// Default behavior: show folder summary.
			return folder.PrintSummary()
		},
	}

	cmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version information")

	// Add subcommands.
	cmd.AddCommand(NewAddCmd())
	cmd.AddCommand(NewRemoveCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewPathCmd())
	cmd.AddCommand(NewRenameCmd())
	cmd.AddCommand(NewGetCmd())
	cmd.AddCommand(NewSetCmd())
	cmd.AddCommand(NewSummaryCmd())
	cmd.AddCommand(NewCleanCmd())
	cmd.AddCommand(NewVersionCmd())

	return cmd
}

// NewAddCmd creates the add command.
func NewAddCmd() *cobra.Command {
	var name string
	var priority string
	var force bool

	cmd := &cobra.Command{
		Use:   "add <executable>",
		Short: "Add an executable to the managed folder",
		Long: `Add a symlink to an executable in the managed folder.
The executable path can be relative or absolute. If --name is not specified,
the basename of the executable will be used as the symlink name.
Use --priority to specify 'front' or 'back' folder (default: front).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if priority != "" && priority != "front" && priority != "back" {
				return fmt.Errorf("--priority must be 'front' or 'back', got '%s'", priority)
			}

			// Default to back if not specified.
			atFront := priority == "front"

			executable := args[0]
			return folder.Add(executable, name, atFront, force)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Custom name for the symlink")
	cmd.Flags().StringVar(&priority, "priority", "front", "Priority: 'front' or 'back' (default: front)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing symlink and ignore masking warnings")

	return cmd
}

// NewRemoveCmd creates the remove command.
func NewRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove a symlink from the managed folder",
		Long:    `Remove a symlink by name from the managed folder.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			return folder.Remove(name)
		},
	}

	return cmd
}

// NewListCmd creates the list command.
func NewListCmd() *cobra.Command {
	var long bool
	var priority string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all managed executables and directories",
		Long: `List all symlinks and directories currently managed by pathman.
Use --priority to list only from 'front' or 'back' folder.
Without --priority, lists from both folders and all managed directories.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if priority != "" && priority != "front" && priority != "back" {
				return fmt.Errorf("--priority must be 'front' or 'back', got '%s'", priority)
			}

			// If priority not specified, list from both.
			if priority == "" {
				if long {
					symlinks, dirs, err := folder.ListLongBothWithDirs()
					if err != nil {
						return err
					}

					// Print symlinks.
					for _, info := range symlinks {
						fmt.Printf("%-5s  %s -> %s\n", info.Priority, info.Name, info.Target)
					}

					// Print directories.
					for _, dir := range dirs {
						fmt.Printf("%-5s  %s/\n", dir.Priority, dir.Path)
					}
				} else {
					symlinks, dirs, err := folder.ListBothWithDirs()
					if err != nil {
						return err
					}

					// Print symlinks.
					for _, name := range symlinks {
						fmt.Println(name)
					}

					// Print directories.
					for _, dir := range dirs {
						fmt.Printf("%s/\n", dir.Path)
					}
				}
				return nil
			}

			// List from specific folder (symlinks only, no directories filtered by priority here).
			atFront := priority == "front"

			if long {
				symlinks, err := folder.ListLong(atFront)
				if err != nil {
					return err
				}

				for _, info := range symlinks {
					fmt.Printf("%-5s  %s -> %s\n", info.Priority, info.Name, info.Target)
				}
			} else {
				symlinks, err := folder.List(atFront)
				if err != nil {
					return err
				}

				for _, name := range symlinks {
					fmt.Println(name)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&long, "long", "l", false, "Show symlink targets and priority")
	cmd.Flags().StringVar(&priority, "priority", "", "List only from 'front' or 'back' folder")

	return cmd
}

// NewPathCmd creates the path command.
func NewPathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Output PATH with managed folders included",
		Long: `Check if the managed folders are on $PATH and add them if not.
Removes any existing occurrences of the folders and adds the front folder
to the front of PATH and the back folder to the back of PATH.
Outputs the adjusted PATH for use in shell configuration.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			adjustedPath, err := folder.GetAdjustedPath()
			if err != nil {
				return err
			}

			fmt.Println(adjustedPath)
			return nil
		},
	}

	return cmd
}

// NewRenameCmd creates the rename command.
func NewRenameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename a symlink in the managed folders",
		Long:  `Rename a symlink in whichever managed folder contains it.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName := args[0]
			newName := args[1]
			return folder.Rename(oldName, newName)
		},
	}

	return cmd
}

// NewGetCmd creates the get command.
func NewGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Show the priority of a symlink",
		Long:  `Show which folder (front or back) a symlink is in.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			return folder.ShowPriority(name)
		},
	}

	return cmd
}

// NewSetCmd creates the set command.
func NewSetCmd() *cobra.Command {
	var priority string

	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "Change the priority of a symlink",
		Long:  `Move a symlink between front and back folders using --priority flag.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if priority == "" {
				return fmt.Errorf("--priority flag is required")
			}
			if priority != "front" && priority != "back" {
				return fmt.Errorf("--priority must be 'front' or 'back', got '%s'", priority)
			}
			return folder.SetPriority(name, priority == "front")
		},
	}

	cmd.Flags().StringVar(&priority, "priority", "", "Priority: 'front' or 'back' (required)")
	if err := cmd.MarkFlagRequired("priority"); err != nil {
		panic(fmt.Sprintf("failed to mark priority flag as required: %v", err))
	}

	return cmd
}

// NewSummaryCmd creates the summary command.
func NewSummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Display a summary of both managed folders",
		Long:  `Display the paths and status of both managed folders, including any name clashes.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return folder.PrintSummary()
		},
	}

	return cmd
}

// NewVersionCmd creates the version command.
func NewVersionCmd() *cobra.Command {
	var jsonFlag bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  `Print the version of pathman.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonFlag {
				return outputVersionJSON()
			}
			fmt.Printf("pathman version %s\n", Version)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output in JSON format")

	return cmd
}

// outputVersionJSON outputs version information in JSON format.
func outputVersionJSON() error {
	output := map[string]string{
		"version": Version,
		"source":  Source,
	}

	// Pretty-print JSON.
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
