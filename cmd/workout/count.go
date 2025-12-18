package workout

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Get total workout count",
	Long: `Get the total number of workouts in your Hevy account.

Examples:
  hevycli workout count           # Show workout count
  hevycli workout count -o json   # Output as JSON`,
	RunE: runCount,
}

func runCount(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

	count, err := client.GetWorkoutCount()
	if err != nil {
		return fmt.Errorf("failed to fetch workout count: %w", err)
	}

	// Determine output format
	outputFmt := cfg.Display.OutputFormat
	if cmd.Flags().Changed("output") {
		outputFmt, _ = cmd.Flags().GetString("output")
	}

	formatter := output.NewFormatter(output.Options{
		Format:  output.FormatType(outputFmt),
		NoColor: !cfg.Display.Color,
		Writer:  os.Stdout,
	})

	if outputFmt == "json" {
		result := map[string]interface{}{
			"workout_count": count,
		}
		out, err := formatter.Format(result)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else if outputFmt == "plain" {
		fmt.Println(count)
	} else {
		fmt.Printf("Total workouts: %d\n", count)
	}

	return nil
}
