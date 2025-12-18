package folder

import "github.com/spf13/cobra"

// Cmd is the folder command
var Cmd = &cobra.Command{
	Use:   "folder",
	Short: "Manage routine folders",
	Long: `View and manage your Hevy routine folders.

Examples:
  hevycli folder list          # List all folders
  hevycli folder get <id>      # Get folder details`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
}
