package folder

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/cmdutil"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
	"github.com/obay/hevycli/internal/tui/prompt"
)

var (
	updateTitle string
)

var updateCmd = &cobra.Command{
	Use:   "update <folder-id>",
	Short: "Update a routine folder",
	Long: `Update a routine folder's title.

Examples:
  hevycli folder update <id> --title "New Name"
  hevycli folder update <id> --title "Push/Pull" -o json`,
	Args: cmdutil.RequireArgs(1, "<folder-id>"),
	RunE: runFolderUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New folder title")
	updateCmd.MarkFlagRequired("title")
	Cmd.AddCommand(updateCmd)
}

func runFolderUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

	var folderID string
	if len(args) > 0 {
		folderID = args[0]
	} else {
		// Interactive mode - let user select from folders
		selected, err := prompt.SearchSelect(prompt.SearchSelectConfig{
			Title:       "Select Folder to Update",
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
						ID:          f.ID,
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
		folderID = selected.ID
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

	// Update the folder
	req := &api.UpdateRoutineFolderRequest{
		RoutineFolder: api.UpdateRoutineFolderData{
			Title: updateTitle,
		},
	}

	folder, err := client.UpdateRoutineFolder(folderID, req)
	if err != nil {
		return fmt.Errorf("failed to update folder: %w", err)
	}

	// Format output
	if outputFmt == "json" {
		out, err := formatter.Format(folder)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		fmt.Println("Folder updated successfully!")
		fmt.Printf("ID: %s\n", folder.ID)
		fmt.Printf("Title: %s\n", folder.Title)
	}

	return nil
}
