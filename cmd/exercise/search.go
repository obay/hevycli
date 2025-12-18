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

var (
	searchMuscle    string
	searchEquipment string
	searchLimit     int
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search exercise templates",
	Long: `Search for exercises by name, muscle group, or equipment.

Examples:
  hevycli exercise search "bench"                    # Search by name
  hevycli exercise search "press" --muscle chest     # Filter by muscle
  hevycli exercise search "" --equipment barbell     # Filter by equipment only
  hevycli exercise search "curl" -o json             # Output as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringVar(&searchMuscle, "muscle", "", "Filter by muscle group")
	searchCmd.Flags().StringVar(&searchEquipment, "equipment", "", "Filter by equipment type")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 20, "Maximum number of results")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(args[0])

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

	// Search through all exercises
	var results []api.ExerciseTemplate
	page := 1

	for len(results) < searchLimit {
		resp, err := client.GetExerciseTemplates(page, 10)
		if err != nil {
			return fmt.Errorf("failed to fetch exercises: %w", err)
		}

		for _, ex := range resp.ExerciseTemplates {
			if matchesSearch(ex, query, searchMuscle, searchEquipment) {
				results = append(results, ex)
				if len(results) >= searchLimit {
					break
				}
			}
		}

		if page >= resp.PageCount || resp.PageCount == 0 {
			break
		}
		page++

		// Progress indicator
		if outputFmt == "table" && page%10 == 0 {
			fmt.Printf("\rSearching... page %d", page)
		}
	}

	if outputFmt == "table" {
		fmt.Print("\r                    \r")
	}

	// Format output
	if outputFmt == "json" {
		result := map[string]interface{}{
			"exercise_templates": results,
			"count":              len(results),
			"query":              query,
		}
		out, err := formatter.Format(result)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		if len(results) == 0 {
			fmt.Println("No exercises found matching your search criteria.")
			return nil
		}

		table := output.NewSimpleTable([]string{"ID", "Title", "Primary Muscle", "Equipment"})

		for _, ex := range results {
			table.AddRow(
				ex.ID,
				truncateString(ex.Title, 35),
				ex.PrimaryMuscleGroup,
				ex.Equipment,
			)
		}

		out, err := formatter.Format(table)
		if err != nil {
			return err
		}
		fmt.Println(out)

		if outputFmt == "table" {
			fmt.Printf("\nFound %d exercise(s)\n", len(results))
		}
	}

	return nil
}

func matchesSearch(ex api.ExerciseTemplate, query, muscle, equipment string) bool {
	// Check name match
	if query != "" && !strings.Contains(strings.ToLower(ex.Title), query) {
		return false
	}

	// Check muscle filter
	if muscle != "" {
		muscle = strings.ToLower(muscle)
		primaryMatch := strings.Contains(strings.ToLower(ex.PrimaryMuscleGroup), muscle)
		secondaryMatch := false
		for _, m := range ex.SecondaryMuscleGroups {
			if strings.Contains(strings.ToLower(m), muscle) {
				secondaryMatch = true
				break
			}
		}
		if !primaryMatch && !secondaryMatch {
			return false
		}
	}

	// Check equipment filter
	if equipment != "" {
		if !strings.Contains(strings.ToLower(ex.Equipment), strings.ToLower(equipment)) {
			return false
		}
	}

	return true
}
