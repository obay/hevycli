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
  hevycli workout count             # Get total workout count
  hevycli workout create --file w.json  # Create from JSON
  hevycli workout update <id> --file w.json  # Update workout
  hevycli workout delete <id>       # Delete workout
  hevycli workout start             # Start interactive session
  hevycli workout events --since 2024-01-01  # Get change events`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(countCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
	// startCmd is added in start.go's init()
}
