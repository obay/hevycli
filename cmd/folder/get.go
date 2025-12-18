package folder

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var getCmd = &cobra.Command{
	Use:   "get <folder-id>",
	Short: "Get folder details",
	Long: `Get detailed information about a specific routine folder.

Examples:
  hevycli folder get abc123-def456    # Get folder by ID
  hevycli folder get abc123 -o json   # Output as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	folderID := args[0]

	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

	// Search for the folder in the list
	page := 1
	var folder *api.RoutineFolder

	for {
		resp, err := client.GetRoutineFolders(page, 10)
		if err != nil {
			return fmt.Errorf("failed to fetch folders: %w", err)
		}

		for _, f := range resp.RoutineFolders {
			if f.ID == folderID {
				folder = &f
				break
			}
		}

		if folder != nil || page >= resp.PageCount || resp.PageCount == 0 {
			break
		}
		page++
	}

	if folder == nil {
		return fmt.Errorf("folder not found: %s", folderID)
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
		out, err := formatter.Format(folder)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		printFolderDetails(folder, cfg)
	}

	return nil
}

func printFolderDetails(f *api.RoutineFolder, cfg *config.Config) {
	fmt.Printf("Folder: %s\n", f.Title)
	fmt.Printf("ID: %s\n", f.ID)
	fmt.Printf("Index: %d\n", f.Index)
	fmt.Printf("Created: %s\n", f.CreatedAt.Format(cfg.Display.DateFormat))
	fmt.Printf("Updated: %s\n", f.UpdatedAt.Format(cfg.Display.DateFormat))
}
