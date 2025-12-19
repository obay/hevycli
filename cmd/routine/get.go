package routine

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
	Use:   "get <routine-id>",
	Short: "Get routine details",
	Long: `Get detailed information about a specific workout routine.

Examples:
  hevycli routine get abc123-def456    # Get routine by ID
  hevycli routine get abc123 -o json   # Output as JSON`,
	Args: cmdutil.RequireArgs(1, "<routine-id>"),
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

	var routineID string
	if len(args) > 0 {
		routineID = args[0]
	} else {
		// Interactive mode - let user select from routines
		selected, err := prompt.SearchSelect(prompt.SearchSelectConfig{
			Title:       "Select a Routine",
			Placeholder: "Search routines...",
			Help:        "Type to filter by routine title",
			LoadFunc: func() ([]prompt.SelectOption, error) {
				routines, err := client.GetRoutines(1, 20)
				if err != nil {
					return nil, err
				}
				options := make([]prompt.SelectOption, len(routines.Routines))
				for i, r := range routines.Routines {
					options[i] = prompt.SelectOption{
						ID:          r.ID,
						Title:       r.Title,
						Description: fmt.Sprintf("%d exercises", len(r.Exercises)),
					}
				}
				return options, nil
			},
		})
		if err != nil {
			return err
		}
		routineID = selected.ID
	}

	routine, err := client.GetRoutine(routineID)
	if err != nil {
		return fmt.Errorf("failed to fetch routine: %w", err)
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
		out, err := formatter.Format(routine)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		printRoutineDetails(routine, cfg, formatter)
	}

	return nil
}

func printRoutineDetails(r *api.Routine, cfg *config.Config, formatter output.Formatter) {
	fmt.Printf("Routine: %s\n", r.Title)
	fmt.Printf("ID: %s\n", r.ID)

	if r.FolderID != nil {
		fmt.Printf("Folder: %s\n", *r.FolderID)
	}

	fmt.Printf("Created: %s\n", r.CreatedAt.Format(cfg.Display.DateFormat))
	fmt.Printf("Updated: %s\n", r.UpdatedAt.Format(cfg.Display.DateFormat))

	fmt.Printf("\nExercises (%d):\n", len(r.Exercises))
	fmt.Println(strings.Repeat("-", 60))

	for i, ex := range r.Exercises {
		fmt.Printf("\n%d. %s\n", i+1, ex.Title)

		if ex.Notes != "" {
			fmt.Printf("   Notes: %s\n", ex.Notes)
		}

		fmt.Printf("   Sets: %d\n", len(ex.Sets))

		// Show set details
		if len(ex.Sets) > 0 {
			table := output.NewSimpleTable([]string{"Set", "Type"})
			for _, set := range ex.Sets {
				setType := string(set.SetType)
				if setType == "" {
					setType = "normal"
				}
				table.AddRow(fmt.Sprintf("%d", set.Index+1), setType)
			}

			out, _ := formatter.Format(table)
			lines := strings.Split(out, "\n")
			for _, line := range lines {
				fmt.Printf("   %s\n", line)
			}
		}
	}
}
