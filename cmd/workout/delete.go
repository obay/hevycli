package workout

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
	Use:   "delete <workout-id>",
	Short: "Delete a workout",
	Long: `Delete a workout by ID.

By default, you will be prompted to confirm the deletion.
Use --force to skip the confirmation prompt.

Examples:
  hevycli workout delete <id>           # Delete with confirmation
  hevycli workout delete <id> --force   # Delete without confirmation`,
	Args: cmdutil.RequireArgs(1, "<workout-id>"),
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
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

	var workoutID string
	if len(args) > 0 {
		workoutID = args[0]
	} else {
		// Interactive mode - let user select from recent workouts
		selected, err := prompt.SearchSelect(prompt.SearchSelectConfig{
			Title:       "Select Workout to Delete",
			Placeholder: "Search workouts...",
			Help:        "Type to filter by workout title",
			LoadFunc: func() ([]prompt.SelectOption, error) {
				workouts, err := client.GetWorkouts(1, 20)
				if err != nil {
					return nil, err
				}
				options := make([]prompt.SelectOption, len(workouts.Workouts))
				for i, w := range workouts.Workouts {
					options[i] = prompt.SelectOption{
						ID:          w.ID,
						Title:       w.Title,
						Description: w.StartTime.Format("Jan 2, 2006") + " • " + fmt.Sprintf("%d exercises", len(w.Exercises)),
					}
				}
				return options, nil
			},
		})
		if err != nil {
			return err
		}
		workoutID = selected.ID
	}

	// Get workout details first to show what we're deleting
	workout, err := client.GetWorkout(workoutID)
	if err != nil {
		return fmt.Errorf("failed to fetch workout: %w", err)
	}

	// Confirm deletion unless --force is used
	if !deleteForce {
		fmt.Printf("Are you sure you want to delete workout '%s' (%s)?\n", workout.Title, workout.ID)
		fmt.Printf("This action cannot be undone.\n")
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			return fmt.Errorf("deletion cancelled")
		}
	}

	// Delete the workout
	if err := client.DeleteWorkout(workoutID); err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}

	fmt.Printf("Workout '%s' deleted successfully.\n", workout.Title)
	return nil
}
