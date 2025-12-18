package routine

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	routineTUI "github.com/obay/hevycli/internal/tui/routine"
)

var builderCmd = &cobra.Command{
	Use:   "builder",
	Short: "Interactive routine builder",
	Long: `Create a new routine using an interactive terminal interface.

The builder allows you to:
  - Set a routine title
  - Search and add exercises from the template library
  - Configure sets and rest times for each exercise
  - Reorder or remove exercises
  - Save the routine to your Hevy account

Examples:
  hevycli routine builder     # Start the interactive builder`,
	RunE: runBuilder,
}

func init() {
	Cmd.AddCommand(builderCmd)
}

func runBuilder(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

	routine, err := routineTUI.Run(client)
	if err != nil {
		return fmt.Errorf("routine builder error: %w", err)
	}

	if routine == nil {
		fmt.Println("Routine creation cancelled.")
	}

	return nil
}
