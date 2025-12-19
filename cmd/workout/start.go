package workout

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	tuiWorkout "github.com/obay/hevycli/internal/tui/workout"
)

var (
	startFromRoutine string
	startNoSave      bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an interactive workout session",
	Long: `Start an interactive workout session in the terminal.

You can start a blank workout or use an existing routine as a template.
When you finish the workout, it will be saved to Hevy automatically.

Examples:
  hevycli workout start                           # Start blank workout
  hevycli workout start --routine <routine-id>    # Start from routine
  hevycli workout start --no-save                 # Don't save to Hevy`,
	RunE: runStart,
}

func init() {
	startCmd.Flags().StringVar(&startFromRoutine, "routine", "", "Start from routine ID")
	startCmd.Flags().BoolVar(&startNoSave, "no-save", false, "Don't save workout to Hevy when finished")
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

	result, err := tuiWorkout.RunSession(title, exercises)
	if err != nil {
		return fmt.Errorf("error running workout session: %w", err)
	}

	if result == nil {
		fmt.Println("\nWorkout cancelled.")
		return nil
	}

	if result.Finished {
		fmt.Println("\nWorkout completed!")
		fmt.Printf("Exercises: %d\n", len(result.Exercises))

		totalSets := 0
		completedSets := 0
		for _, ex := range result.Exercises {
			totalSets += len(ex.Sets)
			for _, s := range ex.Sets {
				if s.Complete {
					completedSets++
				}
			}
		}
		fmt.Printf("Sets completed: %d/%d\n", completedSets, totalSets)

		// Save workout to Hevy unless --no-save is used
		if !startNoSave && completedSets > 0 {
			fmt.Print("\nSave workout to Hevy? (yes/no): ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response == "yes" || response == "y" {
				workout, err := saveWorkoutToHevy(client, result)
				if err != nil {
					return fmt.Errorf("failed to save workout: %w", err)
				}
				fmt.Printf("\nWorkout saved successfully!\n")
				fmt.Printf("ID: %s\n", workout.ID)
			} else {
				fmt.Println("\nWorkout not saved.")
			}
		} else if startNoSave {
			fmt.Println("\nWorkout not saved (--no-save flag used).")
		} else {
			fmt.Println("\nNo completed sets to save.")
		}
	} else {
		fmt.Println("\nWorkout cancelled.")
	}

	return nil
}

// saveWorkoutToHevy converts the session result to an API request and saves it
func saveWorkoutToHevy(client *api.Client, result *tuiWorkout.SessionResult) (*api.Workout, error) {
	// Build the exercises for the API request
	apiExercises := make([]api.CreateWorkoutExercise, 0)

	for _, ex := range result.Exercises {
		// Only include exercises with at least one completed set
		completedSets := make([]api.CreateWorkoutSet, 0)
		for _, s := range ex.Sets {
			if s.Complete {
				weight := s.Weight
				reps := s.Reps
				completedSets = append(completedSets, api.CreateWorkoutSet{
					Type:     s.SetType,
					WeightKg: &weight,
					Reps:     &reps,
				})
			}
		}

		if len(completedSets) > 0 {
			var notes *string
			if ex.Notes != "" {
				notes = &ex.Notes
			}

			apiExercises = append(apiExercises, api.CreateWorkoutExercise{
				ExerciseTemplateID: ex.Template.ID,
				Notes:              notes,
				Sets:               completedSets,
			})
		}
	}

	if len(apiExercises) == 0 {
		return nil, fmt.Errorf("no completed exercises to save")
	}

	req := &api.CreateWorkoutRequest{
		Workout: api.CreateWorkoutData{
			Title:     result.Title,
			StartTime: result.StartTime.Format(time.RFC3339),
			EndTime:   result.EndTime.Format(time.RFC3339),
			Exercises: apiExercises,
		},
	}

	return client.CreateWorkout(req)
}
