package exercise

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
	Use:   "get <exercise-id>",
	Short: "Get exercise template details",
	Long: `Get detailed information about a specific exercise template.

Examples:
  hevycli exercise get ABC123       # Get exercise by ID
  hevycli exercise get ABC123 -o json  # Output as JSON`,
	Args: cmdutil.RequireArgs(1, "<exercise-id>"),
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

	var exerciseID string
	if len(args) > 0 {
		exerciseID = args[0]
	} else {
		// Interactive mode - let user search and select an exercise
		selected, err := prompt.SearchSelect(prompt.SearchSelectConfig{
			Title:       "Select an Exercise",
			Placeholder: "Search exercises...",
			Help:        "Type to filter by exercise name",
			LoadFunc: func() ([]prompt.SelectOption, error) {
				var allExercises []api.ExerciseTemplate
				page := 1
				for {
					resp, err := client.GetExerciseTemplates(page, 10)
					if err != nil {
						return nil, err
					}
					allExercises = append(allExercises, resp.ExerciseTemplates...)
					if page >= resp.PageCount || len(allExercises) > 500 {
						break
					}
					page++
				}
				options := make([]prompt.SelectOption, len(allExercises))
				for i, ex := range allExercises {
					options[i] = prompt.SelectOption{
						ID:          ex.ID,
						Title:       ex.Title,
						Description: ex.PrimaryMuscleGroup + " • " + ex.Equipment,
					}
				}
				return options, nil
			},
		})
		if err != nil {
			return err
		}
		exerciseID = selected.ID
	}

	// Get exercise template by fetching from the list
	// The API may not have a direct /exercise_templates/{id} endpoint that returns full data
	// So we search through the list
	page := 1
	var exercise *api.ExerciseTemplate

	for {
		resp, err := client.GetExerciseTemplates(page, 10)
		if err != nil {
			return fmt.Errorf("failed to fetch exercises: %w", err)
		}

		for _, ex := range resp.ExerciseTemplates {
			if ex.ID == exerciseID {
				exercise = &ex
				break
			}
		}

		if exercise != nil || page >= resp.PageCount || resp.PageCount == 0 {
			break
		}
		page++
	}

	if exercise == nil {
		return fmt.Errorf("exercise template not found: %s", exerciseID)
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
		out, err := formatter.Format(exercise)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		printExerciseDetails(exercise, cfg)
	}

	return nil
}

func printExerciseDetails(ex *api.ExerciseTemplate, cfg *config.Config) {
	fmt.Printf("Exercise: %s\n", ex.Title)
	fmt.Printf("ID: %s\n", ex.ID)
	fmt.Printf("Primary Muscle: %s\n", ex.PrimaryMuscleGroup)

	if len(ex.SecondaryMuscleGroups) > 0 {
		fmt.Printf("Secondary Muscles: %s\n", strings.Join(ex.SecondaryMuscleGroups, ", "))
	}

	if ex.Equipment != "" {
		fmt.Printf("Equipment: %s\n", ex.Equipment)
	}

	if ex.Type != "" {
		fmt.Printf("Type: %s\n", ex.Type)
	}

	if ex.IsCustom {
		fmt.Println("Custom: Yes")
	} else {
		fmt.Println("Custom: No")
	}
}
