package commands

import (
	"fmt"

	"github.com/sfkleach/pathman/pkg/folder"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for pathman.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pathman",
		Short: "Pathman manages executables on your $PATH",
		Long: `Pathman is a command-line tool that helps you manage the list of applications
accessible by $PATH. With pathman, you can add, remove, and list executables
in two managed folders (front and back of $PATH).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default behavior: show folder summary.
			return folder.PrintSummary()
		},
	}

	// Add subcommands.
	cmd.AddCommand(NewAddCmd())
	cmd.AddCommand(NewRemoveCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewFolderCmd())
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewPathCmd())
	cmd.AddCommand(NewRenameCmd())

	return cmd
}

// NewAddCmd creates the add command.
func NewAddCmd() *cobra.Command {
	var name string
	var front bool
	var back bool

	cmd := &cobra.Command{
		Use:   "add <executable>",
		Short: "Add an executable to the managed folder",
		Long: `Add a symlink to an executable in the managed folder.
The executable path can be relative or absolute. If --name is not specified,
the basename of the executable will be used as the symlink name.
Use --front to add to the front folder or --back to add to the back folder (default).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if front && back {
				return fmt.Errorf("cannot specify both --front and --back")
			}

			// Default to back if neither specified.
			atFront := front

			executable := args[0]
			return folder.Add(executable, name, atFront)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Custom name for the symlink")
	cmd.Flags().BoolVar(&front, "front", false, "Add to front folder")
	cmd.Flags().BoolVar(&back, "back", false, "Add to back folder (default)")

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
	var front bool
	var back bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all managed executables",
		Long:    `List all symlinks currently managed by pathman.
Use --front to list from the front folder or --back to list from the back folder (default).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if front && back {
				return fmt.Errorf("cannot specify both --front and --back")
			}

			// Default to back if neither specified.
			atFront := front

			if long {
				symlinks, err := folder.ListLong(atFront)
				if err != nil {
					return err
				}

				for _, info := range symlinks {
					fmt.Printf("%s -> %s\n", info.Name, info.Target)
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

	cmd.Flags().BoolVarP(&long, "long", "l", false, "Show symlink targets")
	cmd.Flags().BoolVar(&front, "front", false, "List from front folder")
	cmd.Flags().BoolVar(&back, "back", false, "List from back folder (default)")

	return cmd
}

// NewFolderCmd creates the folder command.
func NewFolderCmd() *cobra.Command {
	var setPath string
	var front bool
	var back bool

	cmd := &cobra.Command{
		Use:   "folder",
		Short: "Display or configure the managed folders",
		Long: `Display the paths of both managed folders, or set a new path for one of them.
With --set, you must also specify either --front or --back.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if setPath != "" {
				// Setting folder path requires --front or --back.
				if !front && !back {
					return fmt.Errorf("--set requires either --front or --back")
				}
				if front && back {
					return fmt.Errorf("cannot specify both --front and --back")
				}

				atFront := front
				if err := folder.SetManagedFolder(setPath, atFront); err != nil {
					return err
				}

				folderLabel := map[bool]string{true: "front", false: "back"}[atFront]
				fmt.Printf("Set %s folder to: %s\n", folderLabel, setPath)
				return nil
			}

			// Default: show folder summary.
			return folder.PrintSummary()
		},
	}

	cmd.Flags().StringVar(&setPath, "set", "", "Set the managed folder path")
	cmd.Flags().BoolVar(&front, "front", false, "Operate on front folder")
	cmd.Flags().BoolVar(&back, "back", false, "Operate on back folder")

	return cmd
}

// NewInitCmd creates the init command.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the managed folder",
		Long: `Create the managed folder with appropriate permissions.
If the folder already exists, check its permissions and warn if insecure.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return folder.Init()
		},
	}

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
