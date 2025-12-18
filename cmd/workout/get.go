package workout

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var getCmd = &cobra.Command{
	Use:   "get <workout-id>",
	Short: "Get workout details",
	Long: `Get detailed information about a specific workout.

Examples:
  hevycli workout get abc123-def456    # Get workout by ID
  hevycli workout get abc123 -o json   # Output as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	workoutID := args[0]

	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

	workout, err := client.GetWorkout(workoutID)
	if err != nil {
		return fmt.Errorf("failed to fetch workout: %w", err)
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
		out, err := formatter.Format(workout)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		// Formatted output
		printWorkoutDetails(workout, cfg, formatter)
	}

	return nil
}

func printWorkoutDetails(w *api.Workout, cfg *config.Config, formatter output.Formatter) {
	fmt.Printf("Workout: %s\n", w.Title)
	fmt.Printf("ID: %s\n", w.ID)
	fmt.Printf("Date: %s\n", w.StartTime.Format(cfg.Display.DateFormat+" "+cfg.Display.TimeFormat))
	fmt.Printf("Duration: %s\n", formatDuration(w.Duration()))

	if w.Description != "" {
		fmt.Printf("Description: %s\n", w.Description)
	}

	fmt.Printf("\nExercises (%d):\n", len(w.Exercises))
	fmt.Println(strings.Repeat("-", 60))

	for i, ex := range w.Exercises {
		fmt.Printf("\n%d. %s\n", i+1, ex.Title)

		if ex.Notes != "" {
			fmt.Printf("   Notes: %s\n", ex.Notes)
		}

		// Print sets in a table
		table := output.NewSimpleTable([]string{"Set", "Type", "Weight", "Reps", "RPE"})

		for _, set := range ex.Sets {
			setNum := fmt.Sprintf("%d", set.Index+1)
			setType := string(set.SetType)

			weight := "-"
			if set.WeightKg != nil {
				if cfg.Display.Units == "imperial" {
					weight = fmt.Sprintf("%.1f lbs", *set.WeightKg*2.20462)
				} else {
					weight = fmt.Sprintf("%.1f kg", *set.WeightKg)
				}
			}

			reps := "-"
			if set.Reps != nil {
				reps = fmt.Sprintf("%d", *set.Reps)
			}

			rpe := "-"
			if set.RPE != nil {
				rpe = fmt.Sprintf("%.1f", *set.RPE)
			}

			table.AddRow(setNum, setType, weight, reps, rpe)
		}

		out, _ := formatter.Format(table)
		// Indent the table
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			fmt.Printf("   %s\n", line)
		}
	}
}
