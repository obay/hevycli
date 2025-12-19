package workout

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/cmdutil"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
	"github.com/obay/hevycli/internal/tui/prompt"
)

var getCmd = &cobra.Command{
	Use:   "get <workout-id>",
	Short: "Get workout details",
	Long: `Get detailed information about a specific workout.

Examples:
  hevycli workout get abc123-def456    # Get workout by ID
  hevycli workout get abc123 -o json   # Output as JSON`,
	Args: cmdutil.RequireArgs(1, "<workout-id>"),
	RunE: runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

	var workoutID string
	if len(args) > 0 {
		workoutID = args[0]
	} else {
		// Interactive mode - let user select from recent workouts
		selected, err := prompt.SearchSelect(prompt.SearchSelectConfig{
			Title:       "Select a Workout",
			Placeholder: "Search workouts...",
			Help:        "Type to filter by workout title",
			LoadFunc: func() ([]prompt.SelectOption, error) {
				workouts, err := client.GetWorkouts(1, 20)
				if err != nil {
					return nil, err
				}
				options := make([]prompt.SelectOption, len(workouts.Workouts))
				for i, w := range workouts.Workouts {
					options[i] = prompt.SelectOption{
						ID:          w.ID,
						Title:       w.Title,
						Description: w.StartTime.Format("Jan 2, 2006") + " • " + fmt.Sprintf("%d exercises", len(w.Exercises)),
					}
				}
				return options, nil
			},
		})
		if err != nil {
			return err
		}
		workoutID = selected.ID
	}

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
