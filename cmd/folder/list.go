package folder

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var (
	listPage  int
	listLimit int
	listAll   bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List routine folders",
	Long: `List your Hevy routine folders.

Examples:
  hevycli folder list           # List folders
  hevycli folder list -o json   # Output as JSON`,
	RunE: runList,
}

func init() {
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number for pagination")
	listCmd.Flags().IntVar(&listLimit, "limit", 10, "Number of folders to fetch")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Fetch all folders")
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

	var allFolders []api.RoutineFolder

	if listAll {
		page := 1
		for {
			resp, err := client.GetRoutineFolders(page, 10)
			if err != nil {
				return fmt.Errorf("failed to fetch folders: %w", err)
			}
			allFolders = append(allFolders, resp.RoutineFolders...)

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

		resp, err := client.GetRoutineFolders(listPage, pageSize)
		if err != nil {
			return fmt.Errorf("failed to fetch folders: %w", err)
		}
		allFolders = resp.RoutineFolders
	}

	// Format output
	if outputFmt == "json" {
		result := map[string]interface{}{
			"routine_folders": allFolders,
			"count":           len(allFolders),
		}
		out, err := formatter.Format(result)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		if len(allFolders) == 0 {
			fmt.Println("No routine folders found.")
			return nil
		}

		table := output.NewSimpleTable([]string{"ID", "Title", "Index", "Updated"})

		for _, f := range allFolders {
			updated := f.UpdatedAt.Format(cfg.Display.DateFormat)
			table.AddRow(
				f.ID,
				f.Title,
				fmt.Sprintf("%d", f.Index),
				updated,
			)
		}

		out, err := formatter.Format(table)
		if err != nil {
			return err
		}
		fmt.Println(out)

		if outputFmt == "table" {
			fmt.Printf("\nShowing %d folder(s)\n", len(allFolders))
		}
	}

	return nil
}
