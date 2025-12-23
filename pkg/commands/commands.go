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
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all managed executables",
		Long:    `List all symlinks currently managed by pathman.`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			symlinks, err := folder.List()
			if err != nil {
				return err
			}

			for _, name := range symlinks {
				fmt.Println(name)
			}
			return nil
		},
	}

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
