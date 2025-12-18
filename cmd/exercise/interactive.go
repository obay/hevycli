package exercise

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	tuiExercise "github.com/obay/hevycli/internal/tui/exercise"
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Aliases: []string{"i", "browse"},
	Short:   "Interactive exercise search",
	Long: `Launch an interactive exercise search with fuzzy filtering.

Use arrow keys to navigate, type to filter, and press Enter to select.

Examples:
  hevycli exercise interactive
  hevycli exercise i
  hevycli exercise browse`,
	RunE: runInteractive,
}

func init() {
	Cmd.AddCommand(interactiveCmd)
}

func runInteractive(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

	selected, err := tuiExercise.Run(client)
	if err != nil {
		return fmt.Errorf("error running exercise search: %w", err)
	}

	if selected != nil {
		fmt.Printf("\nSelected exercise:\n")
		fmt.Printf("  ID: %s\n", selected.ID)
		fmt.Printf("  Title: %s\n", selected.Title)
		fmt.Printf("  Primary Muscle: %s\n", selected.PrimaryMuscleGroup)
		if selected.Equipment != "" {
			fmt.Printf("  Equipment: %s\n", selected.Equipment)
		}
	}

	return nil
}
