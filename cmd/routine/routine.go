package routine

import "github.com/spf13/cobra"

// Cmd is the routine command
var Cmd = &cobra.Command{
	Use:   "routine",
	Short: "Manage workout routines",
	Long: `View and manage your Hevy workout routines (templates).

Examples:
  hevycli routine list           # List all routines
  hevycli routine get <id>       # Get routine details`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
}
