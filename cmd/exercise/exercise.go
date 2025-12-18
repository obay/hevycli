package exercise

import "github.com/spf13/cobra"

// Cmd is the exercise command
var Cmd = &cobra.Command{
	Use:   "exercise",
	Short: "Browse and manage exercise templates",
	Long: `Browse, search, and create exercise templates.

Examples:
  hevycli exercise list             # List exercise templates
  hevycli exercise get <id>         # Get exercise details
  hevycli exercise search "bench"   # Search for exercises
  hevycli exercise create --title "My Exercise" --muscle chest --type weight_reps
  hevycli exercise interactive      # Interactive exercise browser`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(searchCmd)
	Cmd.AddCommand(createCmd)
	// interactiveCmd is added in interactive.go's init()
}
