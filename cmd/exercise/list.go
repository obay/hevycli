package exercise

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var (
	listPage  int
	listLimit int
	listAll   bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List exercise templates",
	Long: `List exercise templates from the Hevy database.

Examples:
  hevycli exercise list              # List exercises
  hevycli exercise list --all        # List all exercises (may be slow)
  hevycli exercise list -o json      # Output as JSON`,
	RunE: runList,
}

func init() {
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number for pagination")
	listCmd.Flags().IntVar(&listLimit, "limit", 10, "Number of exercises to fetch")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Fetch all exercises (warning: may be slow)")
}

func runList(cmd *cobra.Command, args []string) error {
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

	var allExercises []api.ExerciseTemplate

	if listAll {
		// Fetch all exercises with pagination
		page := 1
		for {
			resp, err := client.GetExerciseTemplates(page, 10)
			if err != nil {
				return fmt.Errorf("failed to fetch exercises: %w", err)
			}
			allExercises = append(allExercises, resp.ExerciseTemplates...)

			if page >= resp.PageCount || resp.PageCount == 0 {
				break
			}
			page++

			// Progress indicator
			if !cmd.Flags().Changed("output") || outputFmt == "table" {
				fmt.Printf("\rFetching exercises... page %d/%d", page, resp.PageCount)
			}
		}
		if !cmd.Flags().Changed("output") || outputFmt == "table" {
			fmt.Print("\r                                    \r")
		}
	} else {
		pageSize := listLimit
		if pageSize > 10 {
			pageSize = 10
		}

		resp, err := client.GetExerciseTemplates(listPage, pageSize)
		if err != nil {
			return fmt.Errorf("failed to fetch exercises: %w", err)
		}
		allExercises = resp.ExerciseTemplates
	}

	// Format output
	if outputFmt == "json" {
		result := map[string]interface{}{
			"exercise_templates": allExercises,
			"count":              len(allExercises),
		}
		out, err := formatter.Format(result)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		table := output.NewSimpleTable([]string{"ID", "Title", "Primary Muscle", "Equipment", "Custom"})

		for _, ex := range allExercises {
			custom := "No"
			if ex.IsCustom {
				custom = "Yes"
			}

			table.AddRow(
				ex.ID,
				truncateString(ex.Title, 30),
				ex.PrimaryMuscleGroup,
				ex.Equipment,
				custom,
			)
		}

		out, err := formatter.Format(table)
		if err != nil {
			return err
		}
		fmt.Println(out)

		if outputFmt == "table" {
			fmt.Printf("\nShowing %d exercise(s)\n", len(allExercises))
		}
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
