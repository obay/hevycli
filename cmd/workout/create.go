package workout

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var (
	createFile      string
	createTitle     string
	createPrivate   bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new workout",
	Long: `Create a new workout from a JSON file or with basic parameters.

The JSON file should contain the workout data in the following format:
{
  "workout": {
    "title": "Morning Workout",
    "description": "Optional description",
    "start_time": "2024-01-15T09:00:00Z",
    "end_time": "2024-01-15T10:00:00Z",
    "is_private": false,
    "exercises": [
      {
        "exercise_template_id": "D04AC939",
        "notes": "Felt strong today",
        "sets": [
          {"type": "warmup", "weight_kg": 60, "reps": 10},
          {"type": "normal", "weight_kg": 100, "reps": 8}
        ]
      }
    ]
  }
}

Examples:
  hevycli workout create --file workout.json           # Create from JSON file
  hevycli workout create --file workout.json -o json   # Output as JSON`,
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "JSON file with workout data (required)")
	createCmd.Flags().StringVar(&createTitle, "title", "", "Workout title (overrides file)")
	createCmd.Flags().BoolVar(&createPrivate, "private", false, "Make workout private")
	createCmd.MarkFlagRequired("file")
}

func runCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

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
	data, err := os.ReadFile(createFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var req api.CreateWorkoutRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Override title if provided
	if createTitle != "" {
		req.Workout.Title = createTitle
	}

	// Override privacy if flag is set
	if cmd.Flags().Changed("private") {
		req.Workout.IsPrivate = createPrivate
	}

	// Create the workout
	workout, err := client.CreateWorkout(&req)
	if err != nil {
		return fmt.Errorf("failed to create workout: %w", err)
	}

	// Format output
	if outputFmt == "json" {
		out, err := formatter.Format(workout)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		fmt.Println("Workout created successfully!")
		fmt.Printf("ID: %s\n", workout.ID)
		fmt.Printf("Title: %s\n", workout.Title)
		fmt.Printf("Start: %s\n", workout.StartTime.Format(time.RFC3339))
		fmt.Printf("End: %s\n", workout.EndTime.Format(time.RFC3339))
		fmt.Printf("Exercises: %d\n", len(workout.Exercises))
	}

	return nil
}
