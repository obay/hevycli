package folder

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var createCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new routine folder",
	Long: `Create a new folder for organizing routines.

Examples:
  hevycli folder create "Push Pull"        # Create a folder
  hevycli folder create "My Routines" -o json   # Output as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runFolderCreate,
}

func init() {
	Cmd.AddCommand(createCmd)
}

func runFolderCreate(cmd *cobra.Command, args []string) error {
	title := args[0]

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

	// Create the folder
	req := &api.CreateRoutineFolderRequest{
		RoutineFolder: api.CreateRoutineFolderData{
			Title: title,
		},
	}

	folder, err := client.CreateRoutineFolder(req)
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	// Format output
	if outputFmt == "json" {
		out, err := formatter.Format(folder)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		fmt.Println("Folder created successfully!")
		fmt.Printf("ID: %s\n", folder.ID)
		fmt.Printf("Title: %s\n", folder.Title)
	}

	return nil
}
