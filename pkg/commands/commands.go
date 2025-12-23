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
in a single local folder.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default behavior: show folder status.
			return folder.PrintStatus()
		},
	}

	// Add subcommands.
	cmd.AddCommand(NewAddCmd())
	cmd.AddCommand(NewRemoveCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewFolderCmd())
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewPathCmd())

	return cmd
}

// NewAddCmd creates the add command.
func NewAddCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "add <executable>",
		Short: "Add an executable to the managed folder",
		Long: `Add a symlink to an executable in the managed folder.
The executable path can be relative or absolute. If --name is not specified,
the basename of the executable will be used as the symlink name.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			executable := args[0]
			return folder.Add(executable, name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Custom name for the symlink")

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

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all managed executables",
		Long:    `List all symlinks currently managed by pathman.`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if long {
				symlinks, err := folder.ListLong()
				if err != nil {
					return err
				}

				for _, info := range symlinks {
					fmt.Printf("%s -> %s\n", info.Name, info.Target)
				}
			} else {
				symlinks, err := folder.List()
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

	return cmd
}

// NewFolderCmd creates the folder command.
func NewFolderCmd() *cobra.Command {
	var setPath string
	var create bool

	cmd := &cobra.Command{
		Use:   "folder",
		Short: "Display or configure the managed folder",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if setPath != "" {
				fmt.Printf("TODO: Set managed folder to '%s'\n", setPath)
				return nil
			}

			if create {
				fmt.Println("TODO: Create the managed folder")
				return nil
			}

			// Default: show folder status.
			return folder.PrintStatus()
		},
	}

	cmd.Flags().StringVar(&setPath, "set", "", "Set the managed folder path")
	cmd.Flags().BoolVar(&create, "create", false, "Create the managed folder")

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
	var front bool
	var back bool

	cmd := &cobra.Command{
		Use:   "path",
		Short: "Output PATH with managed folder included",
		Long: `Check if the managed folder is on $PATH and add it if not.
By default adds to the back. Use --front to add to the front instead.
Outputs the adjusted PATH for use in shell configuration.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if front && back {
				return fmt.Errorf("cannot specify both --front and --back")
			}

			// Default to back if neither specified.
			atFront := front

			adjustedPath, err := folder.GetAdjustedPath(atFront)
			if err != nil {
				return err
			}

			fmt.Println(adjustedPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&front, "front", false, "Add managed folder to front of PATH")
	cmd.Flags().BoolVar(&back, "back", false, "Add managed folder to back of PATH (default)")

	return cmd
}
