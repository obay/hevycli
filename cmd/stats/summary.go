package stats

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var (
	summaryPeriod string
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show workout summary statistics",
	Long: `Display summary statistics for your workouts over a specified period.

Periods:
  week    - Last 7 days
  month   - Last 30 days (default)
  year    - Last 365 days
  all     - All time

Examples:
  hevycli stats summary                # Last 30 days
  hevycli stats summary --period week  # Last 7 days
  hevycli stats summary --period all   # All time statistics`,
	RunE: runSummary,
}

func init() {
	summaryCmd.Flags().StringVar(&summaryPeriod, "period", "month",
		"time period: week, month, year, all")
}

// SummaryStats holds computed statistics
type SummaryStats struct {
	Period struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"period"`
	Workouts struct {
		Total                  int     `json:"total"`
		AverageDurationMinutes float64 `json:"average_duration_minutes"`
		TotalDurationHours     float64 `json:"total_duration_hours"`
	} `json:"workouts"`
	Volume struct {
		TotalKg            float64 `json:"total_kg"`
		AveragePerWorkout  float64 `json:"average_per_workout_kg"`
	} `json:"volume"`
	Exercises struct {
		UniqueCount   int                 `json:"unique_count"`
		TotalSets     int                 `json:"total_sets"`
		MostFrequent  []ExerciseFrequency `json:"most_frequent"`
	} `json:"exercises"`
	Consistency struct {
		WorkoutsPerWeek    float64 `json:"workouts_per_week"`
		LongestStreakDays  int     `json:"longest_streak_days"`
		CurrentStreakDays  int     `json:"current_streak_days"`
	} `json:"consistency"`
}

// ExerciseFrequency tracks exercise usage
type ExerciseFrequency struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func runSummary(cmd *cobra.Command, args []string) error {
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

	// Calculate date range based on period
	now := time.Now()
	var startDate time.Time
	switch summaryPeriod {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	case "all":
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		return fmt.Errorf("invalid period: %s (use week, month, year, or all)", summaryPeriod)
	}

	// Fetch all workouts
	fmt.Fprintln(os.Stderr, "Fetching workout data...")
	allWorkouts, err := client.GetAllWorkouts()
	if err != nil {
		return fmt.Errorf("failed to fetch workouts: %w", err)
	}

	// Filter workouts by date range
	var workouts []api.Workout
	for _, w := range allWorkouts {
		if w.StartTime.After(startDate) && w.StartTime.Before(now) {
			workouts = append(workouts, w)
		}
	}

	// Compute statistics
	stats := computeSummaryStats(workouts, startDate, now)

	// Format output
	if outputFmt == "json" {
		out, err := formatter.Format(stats)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		printSummaryTable(stats, formatter)
	}

	return nil
}

func computeSummaryStats(workouts []api.Workout, start, end time.Time) SummaryStats {
	var stats SummaryStats
	stats.Period.Start = start.Format("2006-01-02")
	stats.Period.End = end.Format("2006-01-02")
	stats.Workouts.Total = len(workouts)

	if len(workouts) == 0 {
		return stats
	}

	// Calculate durations and volume
	var totalDurationMinutes float64
	var totalVolumeKg float64
	exerciseCount := make(map[string]int)
	workoutDates := make(map[string]bool)

	for _, w := range workouts {
		// Calculate duration
		duration := w.EndTime.Sub(w.StartTime).Minutes()
		totalDurationMinutes += duration

		// Track workout dates for streak calculation
		workoutDates[w.StartTime.Format("2006-01-02")] = true

		// Process exercises
		for _, ex := range w.Exercises {
			exerciseCount[ex.Title]++
			stats.Exercises.TotalSets += len(ex.Sets)

			// Calculate volume (weight × reps)
			for _, set := range ex.Sets {
				if set.WeightKg != nil && set.Reps != nil {
					totalVolumeKg += *set.WeightKg * float64(*set.Reps)
				}
			}
		}
	}

	// Workout stats
	stats.Workouts.AverageDurationMinutes = totalDurationMinutes / float64(len(workouts))
	stats.Workouts.TotalDurationHours = totalDurationMinutes / 60

	// Volume stats
	stats.Volume.TotalKg = totalVolumeKg
	stats.Volume.AveragePerWorkout = totalVolumeKg / float64(len(workouts))

	// Exercise stats
	stats.Exercises.UniqueCount = len(exerciseCount)

	// Most frequent exercises
	type kv struct {
		Name  string
		Count int
	}
	var sorted []kv
	for name, count := range exerciseCount {
		sorted = append(sorted, kv{name, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})
	limit := 5
	if len(sorted) < limit {
		limit = len(sorted)
	}
	for i := 0; i < limit; i++ {
		stats.Exercises.MostFrequent = append(stats.Exercises.MostFrequent, ExerciseFrequency{
			Name:  sorted[i].Name,
			Count: sorted[i].Count,
		})
	}

	// Consistency stats
	days := end.Sub(start).Hours() / 24
	if days > 0 {
		weeks := days / 7
		if weeks > 0 {
			stats.Consistency.WorkoutsPerWeek = float64(len(workouts)) / weeks
		}
	}

	// Calculate streaks
	stats.Consistency.LongestStreakDays, stats.Consistency.CurrentStreakDays = calculateStreaks(workoutDates, end)

	return stats
}

func calculateStreaks(workoutDates map[string]bool, endDate time.Time) (longest, current int) {
	if len(workoutDates) == 0 {
		return 0, 0
	}

	// Convert to sorted slice of dates
	var dates []time.Time
	for dateStr := range workoutDates {
		t, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			dates = append(dates, t)
		}
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	// Calculate longest streak
	currentStreak := 1
	longestStreak := 1

	for i := 1; i < len(dates); i++ {
		diff := dates[i].Sub(dates[i-1]).Hours() / 24
		if diff <= 1 {
			currentStreak++
			if currentStreak > longestStreak {
				longestStreak = currentStreak
			}
		} else {
			currentStreak = 1
		}
	}

	// Calculate current streak (from today going back)
	today := endDate.Format("2006-01-02")
	yesterday := endDate.AddDate(0, 0, -1).Format("2006-01-02")

	if !workoutDates[today] && !workoutDates[yesterday] {
		return longestStreak, 0
	}

	currentActiveStreak := 0
	checkDate := endDate
	for {
		dateStr := checkDate.Format("2006-01-02")
		if workoutDates[dateStr] {
			currentActiveStreak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return longestStreak, currentActiveStreak
}

func printSummaryTable(stats SummaryStats, formatter output.Formatter) {
	fmt.Printf("\n📊 Workout Summary (%s to %s)\n\n", stats.Period.Start, stats.Period.End)

	// Workouts section
	fmt.Println("📅 Workouts")
	fmt.Printf("   Total workouts:      %d\n", stats.Workouts.Total)
	fmt.Printf("   Average duration:    %.0f min\n", stats.Workouts.AverageDurationMinutes)
	fmt.Printf("   Total time:          %.1f hours\n", stats.Workouts.TotalDurationHours)

	// Volume section
	fmt.Println("\n💪 Volume")
	fmt.Printf("   Total volume:        %.0f kg\n", stats.Volume.TotalKg)
	fmt.Printf("   Avg per workout:     %.0f kg\n", stats.Volume.AveragePerWorkout)

	// Exercises section
	fmt.Println("\n🏋️ Exercises")
	fmt.Printf("   Unique exercises:    %d\n", stats.Exercises.UniqueCount)
	fmt.Printf("   Total sets:          %d\n", stats.Exercises.TotalSets)
	if len(stats.Exercises.MostFrequent) > 0 {
		fmt.Println("   Most frequent:")
		for i, ex := range stats.Exercises.MostFrequent {
			fmt.Printf("     %d. %s (%d)\n", i+1, ex.Name, ex.Count)
		}
	}

	// Consistency section
	fmt.Println("\n📈 Consistency")
	fmt.Printf("   Workouts/week:       %.1f\n", stats.Consistency.WorkoutsPerWeek)
	fmt.Printf("   Longest streak:      %d days\n", stats.Consistency.LongestStreakDays)
	fmt.Printf("   Current streak:      %d days\n", stats.Consistency.CurrentStreakDays)
	fmt.Println()
}
