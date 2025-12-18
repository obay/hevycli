package workout

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
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
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
	workoutID := args[0]

	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

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
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Delete the workout
	if err := client.DeleteWorkout(workoutID); err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}

	fmt.Printf("Workout '%s' deleted successfully.\n", workout.Title)
	return nil
}
