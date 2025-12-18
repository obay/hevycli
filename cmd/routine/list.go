package routine

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var (
	listPage   int
	listLimit  int
	listAll    bool
	listFolder string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List workout routines",
	Long: `List your Hevy workout routines (templates).

Examples:
  hevycli routine list              # List routines
  hevycli routine list --all        # List all routines
  hevycli routine list -o json      # Output as JSON`,
	RunE: runList,
}

func init() {
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number for pagination")
	listCmd.Flags().IntVar(&listLimit, "limit", 10, "Number of routines to fetch")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Fetch all routines")
	listCmd.Flags().StringVar(&listFolder, "folder", "", "Filter by folder ID")
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

	var allRoutines []api.Routine

	if listAll {
		// Fetch all routines with pagination
		page := 1
		for {
			resp, err := client.GetRoutines(page, 10)
			if err != nil {
				return fmt.Errorf("failed to fetch routines: %w", err)
			}
			allRoutines = append(allRoutines, resp.Routines...)

			if page >= resp.PageCount || resp.PageCount == 0 {
				break
			}
			page++
		}
	} else {
		pageSize := listLimit
		if pageSize > 10 {
			pageSize = 10
		}

		resp, err := client.GetRoutines(listPage, pageSize)
		if err != nil {
			return fmt.Errorf("failed to fetch routines: %w", err)
		}
		allRoutines = resp.Routines
	}

	// Filter by folder if specified
	if listFolder != "" {
		var filtered []api.Routine
		for _, r := range allRoutines {
			if r.FolderID != nil && *r.FolderID == listFolder {
				filtered = append(filtered, r)
			}
		}
		allRoutines = filtered
	}

	// Format output
	if outputFmt == "json" {
		result := map[string]interface{}{
			"routines": allRoutines,
			"count":    len(allRoutines),
		}
		out, err := formatter.Format(result)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		table := output.NewSimpleTable([]string{"ID", "Title", "Exercises", "Folder", "Updated"})

		for _, r := range allRoutines {
			folder := "-"
			if r.FolderID != nil {
				folder = *r.FolderID
			}

			updated := r.UpdatedAt.Format(cfg.Display.DateFormat)

			table.AddRow(
				r.ID,
				truncateString(r.Title, 30),
				fmt.Sprintf("%d", len(r.Exercises)),
				truncateString(folder, 15),
				updated,
			)
		}

		out, err := formatter.Format(table)
		if err != nil {
			return err
		}
		fmt.Println(out)

		if outputFmt == "table" {
			fmt.Printf("\nShowing %d routine(s)\n", len(allRoutines))
		}
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
