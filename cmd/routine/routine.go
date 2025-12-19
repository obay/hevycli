package routine

import "github.com/spf13/cobra"

// Cmd is the routine command
var Cmd = &cobra.Command{
	Use:   "routine",
	Short: "Manage workout routines",
	Long: `View and manage your Hevy workout routines (templates).

Examples:
  hevycli routine list                  # List all routines
  hevycli routine get <id>              # Get routine details
  hevycli routine create --file r.json  # Create from JSON
  hevycli routine update <id> --file r.json  # Update routine
  hevycli routine delete <id>           # Delete routine
  hevycli routine builder               # Interactive routine builder`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	// deleteCmd is added in delete.go's init()
}
