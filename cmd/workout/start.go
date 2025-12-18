package workout

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	tuiWorkout "github.com/obay/hevycli/internal/tui/workout"
)

var (
	startFromRoutine string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an interactive workout session",
	Long: `Start an interactive workout session in the terminal.

You can start a blank workout or use an existing routine as a template.

Examples:
  hevycli workout start                           # Start blank workout
  hevycli workout start --routine <routine-id>    # Start from routine`,
	RunE: runStart,
}

func init() {
	startCmd.Flags().StringVar(&startFromRoutine, "routine", "", "Start from routine ID")
	Cmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

	var exercises []tuiWorkout.ExerciseData
	var title string

	if startFromRoutine != "" {
		// Load routine
		routine, err := client.GetRoutine(startFromRoutine)
		if err != nil {
			return fmt.Errorf("failed to load routine: %w", err)
		}

		title = routine.Title
		exercises = make([]tuiWorkout.ExerciseData, len(routine.Exercises))
		for i, ex := range routine.Exercises {
			sets := make([]tuiWorkout.SetData, len(ex.Sets))
			for j, s := range ex.Sets {
				sets[j] = tuiWorkout.SetData{
					SetType: s.SetType,
				}
			}
			exercises[i] = tuiWorkout.ExerciseData{
				Template: api.ExerciseTemplate{
					ID:    ex.ExerciseTemplateID,
					Title: ex.Title,
				},
				Sets:  sets,
				Notes: ex.Notes,
			}
		}
	} else {
		// Blank workout with example structure
		title = "New Workout"
		exercises = []tuiWorkout.ExerciseData{
			{
				Template: api.ExerciseTemplate{Title: "Exercise 1"},
				Sets: []tuiWorkout.SetData{
					{SetType: api.SetTypeNormal},
					{SetType: api.SetTypeNormal},
					{SetType: api.SetTypeNormal},
				},
			},
		}
	}

	resultExercises, finished, err := tuiWorkout.RunSession(title, exercises)
	if err != nil {
		return fmt.Errorf("error running workout session: %w", err)
	}

	if finished {
		fmt.Println("\nWorkout completed!")
		fmt.Printf("Exercises: %d\n", len(resultExercises))

		totalSets := 0
		completedSets := 0
		for _, ex := range resultExercises {
			totalSets += len(ex.Sets)
			for _, s := range ex.Sets {
				if s.Complete {
					completedSets++
				}
			}
		}
		fmt.Printf("Sets completed: %d/%d\n", completedSets, totalSets)

		// TODO: In Phase 2+, we would save this workout to the API
		fmt.Println("\nNote: Workout saving to Hevy API is not yet implemented.")
	} else {
		fmt.Println("\nWorkout cancelled.")
	}

	return nil
}
