package stats

import "github.com/spf13/cobra"

// Cmd is the stats command
var Cmd = &cobra.Command{
	Use:   "stats",
	Short: "View workout analytics and statistics",
	Long: `Analyze your workout data with summary statistics, progress tracking, and personal records.

Examples:
  hevycli stats summary                    # Monthly workout summary
  hevycli stats summary --period week      # Weekly summary
  hevycli stats progress "Bench Press"     # Track bench press progress
  hevycli stats records                    # View personal records`,
}

func init() {
	Cmd.AddCommand(summaryCmd)
	Cmd.AddCommand(progressCmd)
	Cmd.AddCommand(recordsCmd)
}
