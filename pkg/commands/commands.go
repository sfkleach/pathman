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

	return cmd
}

// NewAddCmd creates the add command.
func NewAddCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "add <executable>",
		Short: "Add an executable to the managed folder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			executable := args[0]
			fmt.Printf("TODO: Add executable '%s'", executable)
			if name != "" {
				fmt.Printf(" with name '%s'", name)
			}
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Custom name for the symlink")

	return cmd
}

// NewRemoveCmd creates the remove command.
func NewRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a symlink from the managed folder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			fmt.Printf("TODO: Remove symlink '%s'\n", name)
			return nil
		},
	}

	return cmd
}

// NewListCmd creates the list command.
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all managed executables",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("TODO: List all symlinks in the managed folder")
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
