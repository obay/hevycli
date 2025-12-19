package workout

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/cmdutil"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
	"github.com/obay/hevycli/internal/tui/prompt"
)

var (
	updateFile  string
	updateTitle string
)

var updateCmd = &cobra.Command{
	Use:   "update <workout-id>",
	Short: "Update an existing workout",
	Long: `Update an existing workout from a JSON file.

The JSON file should contain the updated workout data in the following format:
{
  "workout": {
    "title": "Updated Workout Title",
    "description": "Updated description",
    "start_time": "2024-01-15T09:00:00Z",
    "end_time": "2024-01-15T10:30:00Z",
    "exercises": [...]
  }
}

Examples:
  hevycli workout update <id> --file workout.json           # Update from JSON file
  hevycli workout update <id> --file workout.json -o json   # Output as JSON`,
	Args: cmdutil.RequireArgs(1, "<workout-id>"),
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().StringVarP(&updateFile, "file", "f", "", "JSON file with workout data (required)")
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "Update workout title (overrides file)")
	updateCmd.MarkFlagRequired("file")
}

func runUpdate(cmd *cobra.Command, args []string) error {
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
			Title:       "Select Workout to Update",
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

	// Read workout data from file
	data, err := os.ReadFile(updateFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var req api.UpdateWorkoutRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Override title if provided
	if updateTitle != "" {
		req.Workout.Title = updateTitle
	}

	// Update the workout
	workout, err := client.UpdateWorkout(workoutID, &req)
	if err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}

	// Format output
	if outputFmt == "json" {
		out, err := formatter.Format(workout)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		fmt.Println("Workout updated successfully!")
		fmt.Printf("ID: %s\n", workout.ID)
		fmt.Printf("Title: %s\n", workout.Title)
		fmt.Printf("Start: %s\n", workout.StartTime.Format(time.RFC3339))
		fmt.Printf("End: %s\n", workout.EndTime.Format(time.RFC3339))
		fmt.Printf("Exercises: %d\n", len(workout.Exercises))
	}

	return nil
}
