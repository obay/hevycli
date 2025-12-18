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
	routineCreateFile   string
	routineCreateTitle  string
	routineCreateFolder int
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new routine",
	Long: `Create a new routine from a JSON file.

The JSON file should contain the routine data in the following format:
{
  "routine": {
    "title": "Push Day A",
    "folder_id": null,
    "notes": "Focus on form",
    "exercises": [
      {
        "exercise_template_id": "D04AC939",
        "rest_seconds": 90,
        "notes": "Stay controlled",
        "sets": [
          {"type": "warmup", "weight_kg": 60, "reps": 10},
          {"type": "normal", "weight_kg": 100, "reps": 8}
        ]
      }
    ]
  }
}

Examples:
  hevycli routine create --file routine.json           # Create from JSON file
  hevycli routine create --file routine.json -o json   # Output as JSON`,
	RunE: runRoutineCreate,
}

func init() {
	createCmd.Flags().StringVarP(&routineCreateFile, "file", "f", "", "JSON file with routine data (required)")
	createCmd.Flags().StringVar(&routineCreateTitle, "title", "", "Routine title (overrides file)")
	createCmd.Flags().IntVar(&routineCreateFolder, "folder", 0, "Folder ID to place routine in")
	createCmd.MarkFlagRequired("file")
}

func runRoutineCreate(cmd *cobra.Command, args []string) error {
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
	data, err := os.ReadFile(routineCreateFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var req api.CreateRoutineRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Override title if provided
	if routineCreateTitle != "" {
		req.Routine.Title = routineCreateTitle
	}

	// Override folder if provided
	if cmd.Flags().Changed("folder") {
		req.Routine.FolderID = &routineCreateFolder
	}

	// Create the routine
	routine, err := client.CreateRoutine(&req)
	if err != nil {
		return fmt.Errorf("failed to create routine: %w", err)
	}

	// Format output
	if outputFmt == "json" {
		out, err := formatter.Format(routine)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		fmt.Println("Routine created successfully!")
		fmt.Printf("ID: %s\n", routine.ID)
		fmt.Printf("Title: %s\n", routine.Title)
		fmt.Printf("Exercises: %d\n", len(routine.Exercises))
	}

	return nil
}
