package exercise

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
	Use:   "get <exercise-id>",
	Short: "Get exercise template details",
	Long: `Get detailed information about a specific exercise template.

Examples:
  hevycli exercise get ABC123       # Get exercise by ID
  hevycli exercise get ABC123 -o json  # Output as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	exerciseID := args[0]

	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

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
	}
}
