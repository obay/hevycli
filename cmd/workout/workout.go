package workout

import "github.com/spf13/cobra"

// Cmd is the workout command
var Cmd = &cobra.Command{
	Use:   "workout",
	Short: "Manage workouts",
	Long: `View and manage your Hevy workouts.

Examples:
  hevycli workout list              # List recent workouts
  hevycli workout list --all        # List all workouts
  hevycli workout get <id>          # Get workout details
  hevycli workout count             # Get total workout count`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(countCmd)
}
