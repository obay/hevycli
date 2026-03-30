package folder

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/cmdutil"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
	"github.com/obay/hevycli/internal/tui/prompt"
)

var getCmd = &cobra.Command{
	Use:   "get <folder-id>",
	Short: "Get folder details",
	Long: `Get detailed information about a specific routine folder.

Examples:
  hevycli folder get 123              # Get folder by ID
  hevycli folder get 123 -o json      # Output as JSON`,
	Args: cmdutil.RequireArgs(1, "<folder-id>"),
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

	var folderID int
	if len(args) > 0 {
		var err error
		folderID, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid folder ID: %s", args[0])
		}
	} else {
		// Interactive mode - let user select from folders
		selected, err := prompt.SearchSelect(prompt.SearchSelectConfig{
			Title:       "Select a Folder",
			Placeholder: "Search folders...",
			Help:        "Type to filter by folder title",
			LoadFunc: func() ([]prompt.SelectOption, error) {
				folders, err := client.GetRoutineFolders(1, 20)
				if err != nil {
					return nil, err
				}
				options := make([]prompt.SelectOption, len(folders.RoutineFolders))
				for i, f := range folders.RoutineFolders {
					options[i] = prompt.SelectOption{
						ID:          fmt.Sprintf("%d", f.ID),
						Title:       f.Title,
						Description: fmt.Sprintf("Index: %d", f.Index),
					}
				}
				return options, nil
			},
		})
		if err != nil {
			return err
		}
		folderID, err = strconv.Atoi(selected.ID)
		if err != nil {
			return fmt.Errorf("invalid folder ID selected: %s", selected.ID)
		}
	}

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
		return fmt.Errorf("folder not found: %d", folderID)
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
	fmt.Printf("ID: %d\n", f.ID)
	fmt.Printf("Index: %d\n", f.Index)
	fmt.Printf("Created: %s\n", f.CreatedAt.Format(cfg.Display.DateFormat))
	fmt.Printf("Updated: %s\n", f.UpdatedAt.Format(cfg.Display.DateFormat))
}
