package routine

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
	Use:   "delete <routine-id>",
	Short: "Delete a routine",
	Long: `Delete a routine by ID.

By default, you will be prompted to confirm the deletion.
Use --force to skip the confirmation prompt.

Examples:
  hevycli routine delete <id>           # Delete with confirmation
  hevycli routine delete <id> --force   # Delete without confirmation`,
	Args: cmdutil.RequireArgs(1, "<routine-id>"),
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
	Cmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
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
			Title:       "Select Routine to Delete",
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

	// Get routine details first to show what we're deleting
	routine, err := client.GetRoutine(routineID)
	if err != nil {
		return fmt.Errorf("failed to fetch routine: %w", err)
	}

	// Confirm deletion unless --force is used
	if !deleteForce {
		fmt.Printf("Are you sure you want to delete routine '%s' (%s)?\n", routine.Title, routine.ID)
		fmt.Printf("This action cannot be undone.\n")
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

	// Delete the routine
	if err := client.DeleteRoutine(routineID); err != nil {
		return fmt.Errorf("failed to delete routine: %w", err)
	}

	fmt.Printf("Routine '%s' deleted successfully.\n", routine.Title)
	return nil
}
