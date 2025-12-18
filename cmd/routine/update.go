package routine

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
	routineUpdateFile  string
	routineUpdateTitle string
)

var updateCmd = &cobra.Command{
	Use:   "update <routine-id>",
	Short: "Update an existing routine",
	Long: `Update an existing routine from a JSON file.

The JSON file should contain the updated routine data in the following format:
{
  "routine": {
    "title": "Updated Routine Title",
    "notes": "Updated notes",
    "exercises": [...]
  }
}

Examples:
  hevycli routine update <id> --file routine.json           # Update from JSON file
  hevycli routine update <id> --file routine.json -o json   # Output as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runRoutineUpdate,
}

func init() {
	updateCmd.Flags().StringVarP(&routineUpdateFile, "file", "f", "", "JSON file with routine data (required)")
	updateCmd.Flags().StringVar(&routineUpdateTitle, "title", "", "Update routine title (overrides file)")
	updateCmd.MarkFlagRequired("file")
}

func runRoutineUpdate(cmd *cobra.Command, args []string) error {
	routineID := args[0]

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

	// Read routine data from file
	data, err := os.ReadFile(routineUpdateFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var req api.UpdateRoutineRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Override title if provided
	if routineUpdateTitle != "" {
		req.Routine.Title = routineUpdateTitle
	}

	// Update the routine
	routine, err := client.UpdateRoutine(routineID, &req)
	if err != nil {
		return fmt.Errorf("failed to update routine: %w", err)
	}

	// Format output
	if outputFmt == "json" {
		out, err := formatter.Format(routine)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		fmt.Println("Routine updated successfully!")
		fmt.Printf("ID: %s\n", routine.ID)
		fmt.Printf("Title: %s\n", routine.Title)
		fmt.Printf("Exercises: %d\n", len(routine.Exercises))
	}

	return nil
}
