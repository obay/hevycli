package exercise

import "github.com/spf13/cobra"

// Cmd is the exercise command
var Cmd = &cobra.Command{
	Use:   "exercise",
	Short: "Browse exercise templates",
	Long: `Browse and search the Hevy exercise template database.

Examples:
  hevycli exercise list             # List exercise templates
  hevycli exercise get <id>         # Get exercise details
  hevycli exercise search "bench"   # Search for exercises`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(searchCmd)
}
