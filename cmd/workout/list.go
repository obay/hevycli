package workout

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var (
	listLimit int
	listPage  int
	listAll   bool
	listSince string
	listUntil string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List workouts",
	Long: `List your Hevy workouts with pagination support.

Examples:
  hevycli workout list                    # List recent workouts (default: 10)
  hevycli workout list --limit 5          # List 5 workouts
  hevycli workout list --all              # List all workouts
  hevycli workout list --since 2024-01-01 # List workouts since date
  hevycli workout list -o json            # Output as JSON`,
	RunE: runList,
}

func init() {
	listCmd.Flags().IntVar(&listLimit, "limit", 10, "Number of workouts to fetch (max 10 per page)")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number for pagination")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Fetch all workouts (auto-pagination)")
	listCmd.Flags().StringVar(&listSince, "since", "", "Filter workouts since date (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listUntil, "until", "", "Filter workouts until date (YYYY-MM-DD)")
}

func runList(cmd *cobra.Command, args []string) error {
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

	var allWorkouts []api.Workout

	if listAll {
		// Fetch all workouts with pagination
		page := 1
		for {
			resp, err := client.GetWorkouts(page, 10)
			if err != nil {
				return fmt.Errorf("failed to fetch workouts: %w", err)
			}
			allWorkouts = append(allWorkouts, resp.Workouts...)

			if page >= resp.PageCount {
				break
			}
			page++
		}
	} else {
		// Fetch single page
		pageSize := listLimit
		if pageSize > 10 {
			pageSize = 10 // API max is 10
		}

		resp, err := client.GetWorkouts(listPage, pageSize)
		if err != nil {
			return fmt.Errorf("failed to fetch workouts: %w", err)
		}
		allWorkouts = resp.Workouts
	}

	// Filter by date if specified
	if listSince != "" || listUntil != "" {
		allWorkouts = filterByDate(allWorkouts, listSince, listUntil)
	}

	// Format output
	if outputFmt == "json" {
		result := map[string]interface{}{
			"workouts": allWorkouts,
			"count":    len(allWorkouts),
		}
		out, err := formatter.Format(result)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		// Table or plain output
		table := output.NewSimpleTable([]string{"ID", "Title", "Date", "Duration", "Exercises"})

		for _, w := range allWorkouts {
			duration := formatDuration(w.Duration())
			date := w.StartTime.Format(cfg.Display.DateFormat)
			table.AddRow(
				w.ID,
				truncateString(w.Title, 30),
				date,
				duration,
				fmt.Sprintf("%d", w.ExerciseCount()),
			)
		}

		out, err := formatter.Format(table)
		if err != nil {
			return err
		}
		fmt.Println(out)

		if outputFmt == "table" {
			fmt.Printf("\nShowing %d workout(s)\n", len(allWorkouts))
		}
	}

	return nil
}

func filterByDate(workouts []api.Workout, since, until string) []api.Workout {
	var filtered []api.Workout

	var sinceDate, untilDate time.Time
	var hasSince, hasUntil bool

	if since != "" {
		if t, err := time.Parse("2006-01-02", since); err == nil {
			sinceDate = t
			hasSince = true
		}
	}

	if until != "" {
		if t, err := time.Parse("2006-01-02", until); err == nil {
			untilDate = t.Add(24 * time.Hour) // Include the entire day
			hasUntil = true
		}
	}

	for _, w := range workouts {
		if hasSince && w.StartTime.Before(sinceDate) {
			continue
		}
		if hasUntil && w.StartTime.After(untilDate) {
			continue
		}
		filtered = append(filtered, w)
	}

	return filtered
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
