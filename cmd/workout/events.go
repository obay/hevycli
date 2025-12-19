package workout

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var (
	eventsSince string
	eventsType  string
	eventsLimit int
	eventsPage  int
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Get workout change events",
	Long: `Get workout events (updates and deletes) since a given date.

This is useful for syncing workout data or detecting changes.

Examples:
  hevycli workout events --since 2024-01-01
  hevycli workout events --since 2024-01-01 --type updated
  hevycli workout events --since 2024-01-01 -o json`,
	RunE: runEvents,
}

func init() {
	eventsCmd.Flags().StringVar(&eventsSince, "since", "", "Get events since date (YYYY-MM-DD)")
	eventsCmd.Flags().StringVar(&eventsType, "type", "", "Filter by event type: updated, deleted")
	eventsCmd.Flags().IntVar(&eventsLimit, "limit", 10, "Number of events to fetch")
	eventsCmd.Flags().IntVar(&eventsPage, "page", 1, "Page number")
	eventsCmd.MarkFlagRequired("since")
	Cmd.AddCommand(eventsCmd)
}

func runEvents(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
	}

	client := api.NewClient(apiKey)

	// Parse the since date
	sinceTime, err := time.Parse("2006-01-02", eventsSince)
	if err != nil {
		return fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
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

	// Fetch events
	resp, err := client.GetWorkoutEvents(sinceTime, eventsPage, eventsLimit)
	if err != nil {
		return fmt.Errorf("failed to fetch workout events: %w", err)
	}

	// Filter by type if specified
	var filteredEvents []api.WorkoutEvent
	if eventsType != "" {
		for _, event := range resp.WorkoutEvents {
			if string(event.Type) == eventsType {
				filteredEvents = append(filteredEvents, event)
			}
		}
	} else {
		filteredEvents = resp.WorkoutEvents
	}

	// Output based on format
	if outputFmt == "json" {
		result := map[string]interface{}{
			"events": filteredEvents,
			"pagination": map[string]interface{}{
				"page":       resp.Page,
				"page_count": resp.PageCount,
			},
		}
		out, err := formatter.Format(result)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else if outputFmt == "plain" {
		for _, event := range filteredEvents {
			fmt.Printf("%s|%s|%s|%s\n",
				event.ID,
				event.WorkoutID,
				event.Type,
				event.Timestamp.Format(time.RFC3339))
		}
	} else {
		// Table format
		if len(filteredEvents) == 0 {
			fmt.Println("No workout events found since", eventsSince)
			return nil
		}

		rows := make([][]string, len(filteredEvents))
		for i, event := range filteredEvents {
			typeStr := string(event.Type)
			if event.Type == api.EventTypeUpdated {
				typeStr = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("updated")
			} else if event.Type == api.EventTypeDeleted {
				typeStr = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("deleted")
			}

			rows[i] = []string{
				event.ID,
				event.WorkoutID,
				typeStr,
				event.Timestamp.Format("2006-01-02 15:04:05"),
			}
		}

		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
			Headers("Event ID", "Workout ID", "Type", "Timestamp").
			Rows(rows...)

		fmt.Println(t)
		fmt.Printf("\nPage %d of %d\n", resp.Page, resp.PageCount)
	}

	return nil
}
