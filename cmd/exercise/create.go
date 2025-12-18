package exercise

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var (
	exerciseCreateFile      string
	exerciseCreateTitle     string
	exerciseCreateType      string
	exerciseCreateMuscle    string
	exerciseCreateEquipment string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom exercise",
	Long: `Create a new custom exercise template.

You can create from a JSON file or using command-line flags.

Exercise Types:
  weight_reps, reps_only, bodyweight_reps, bodyweight_assisted_reps,
  duration, weight_duration, distance_duration, short_distance_weight

Muscle Groups:
  chest, shoulders, biceps, triceps, forearms, lats, upper_back, traps,
  lower_back, abdominals, quadriceps, hamstrings, glutes, calves,
  abductors, adductors, cardio, neck, full_body, other

Equipment:
  none, barbell, dumbbell, kettlebell, machine, plate,
  resistance_band, suspension, other

JSON file format:
{
  "exercise": {
    "title": "My Custom Exercise",
    "exercise_type": "weight_reps",
    "muscle_group": "chest",
    "equipment_category": "barbell",
    "other_muscles": ["triceps", "shoulders"]
  }
}

Examples:
  hevycli exercise create --file exercise.json
  hevycli exercise create --title "Cable Fly" --type weight_reps --muscle chest --equipment machine
  hevycli exercise create --title "Plank" --type duration --muscle abdominals --equipment none`,
	RunE: runExerciseCreate,
}

func init() {
	createCmd.Flags().StringVarP(&exerciseCreateFile, "file", "f", "", "JSON file with exercise data")
	createCmd.Flags().StringVar(&exerciseCreateTitle, "title", "", "Exercise title")
	createCmd.Flags().StringVar(&exerciseCreateType, "type", "weight_reps", "Exercise type")
	createCmd.Flags().StringVar(&exerciseCreateMuscle, "muscle", "", "Primary muscle group")
	createCmd.Flags().StringVar(&exerciseCreateEquipment, "equipment", "none", "Equipment category")
}

func runExerciseCreate(cmd *cobra.Command, args []string) error {
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

	var req api.CreateCustomExerciseRequest

	if exerciseCreateFile != "" {
		// Read from file
		data, err := os.ReadFile(exerciseCreateFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else {
		// Build from flags
		if exerciseCreateTitle == "" {
			return fmt.Errorf("--title is required when not using --file")
		}
		if exerciseCreateMuscle == "" {
			return fmt.Errorf("--muscle is required when not using --file")
		}

		req = api.CreateCustomExerciseRequest{
			Exercise: api.CreateCustomExerciseData{
				Title:             exerciseCreateTitle,
				ExerciseType:      api.ExerciseType(exerciseCreateType),
				MuscleGroup:       api.MuscleGroup(exerciseCreateMuscle),
				EquipmentCategory: api.EquipmentCategory(exerciseCreateEquipment),
			},
		}
	}

	// Override with flags if provided
	if cmd.Flags().Changed("title") && exerciseCreateTitle != "" {
		req.Exercise.Title = exerciseCreateTitle
	}
	if cmd.Flags().Changed("type") {
		req.Exercise.ExerciseType = api.ExerciseType(exerciseCreateType)
	}
	if cmd.Flags().Changed("muscle") {
		req.Exercise.MuscleGroup = api.MuscleGroup(exerciseCreateMuscle)
	}
	if cmd.Flags().Changed("equipment") {
		req.Exercise.EquipmentCategory = api.EquipmentCategory(exerciseCreateEquipment)
	}

	// Create the exercise
	exercise, err := client.CreateCustomExercise(&req)
	if err != nil {
		return fmt.Errorf("failed to create exercise: %w", err)
	}

	// Format output
	if outputFmt == "json" {
		out, err := formatter.Format(exercise)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		fmt.Println("Custom exercise created successfully!")
		fmt.Printf("ID: %s\n", exercise.ID)
		fmt.Printf("Title: %s\n", exercise.Title)
		fmt.Printf("Type: %s\n", exercise.Type)
		fmt.Printf("Primary Muscle: %s\n", exercise.PrimaryMuscleGroup)
		fmt.Printf("Equipment: %s\n", exercise.Equipment)
	}

	return nil
}
