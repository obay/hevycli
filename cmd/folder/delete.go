package folder

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/cmdutil"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/tui/prompt"
)

var (
	deleteForce bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <folder-id>",
	Short: "Delete a routine folder",
	Long: `Delete a routine folder by ID.

By default, you will be prompted to confirm the deletion.
Use --force to skip the confirmation prompt.

Note: Deleting a folder does not delete the routines inside it.
The routines will become unorganized (no folder).

Examples:
  hevycli folder delete <id>           # Delete with confirmation
  hevycli folder delete <id> --force   # Delete without confirmation`,
	Args: cmdutil.RequireArgs(1, "<folder-id>"),
	RunE: runFolderDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
	Cmd.AddCommand(deleteCmd)
}

func runFolderDelete(cmd *cobra.Command, args []string) error {
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
			Title:       "Select Folder to Delete",
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

	// Get folder details first to show what we're deleting
	folder, err := client.GetRoutineFolder(folderID)
	if err != nil {
		return fmt.Errorf("failed to fetch folder: %w", err)
	}

	// Confirm deletion unless --force is used
	if !deleteForce {
		fmt.Printf("Are you sure you want to delete folder '%s' (%s)?\n", folder.Title, folder.ID)
		fmt.Printf("This action cannot be undone. Routines inside will become unorganized.\n")
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Delete the folder
	if err := client.DeleteRoutineFolder(folderID); err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}

	fmt.Printf("Folder '%s' deleted successfully.\n", folder.Title)
	return nil
}
