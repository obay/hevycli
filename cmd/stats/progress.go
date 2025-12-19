package stats

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/cmdutil"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
	"github.com/obay/hevycli/internal/tui/prompt"
)

var (
	progressMetric string
	progressPeriod string
)

var progressCmd = &cobra.Command{
	Use:   "progress <exercise-name>",
	Short: "Track progress on a specific exercise",
	Long: `Analyze your progress on a specific exercise over time.

Metrics:
  weight   - Max weight used (default)
  volume   - Total volume (weight × reps)
  reps     - Max reps at heaviest weight
  1rm      - Estimated one-rep max (Brzycki formula)

Examples:
  hevycli stats progress "Bench Press"
  hevycli stats progress "Squat" --metric 1rm
  hevycli stats progress "Deadlift" --metric volume --period year`,
	Args: cmdutil.RequireArgs(1, "<exercise-name>"),
	RunE: runProgress,
}

func init() {
	progressCmd.Flags().StringVar(&progressMetric, "metric", "weight",
		"metric to track: weight, volume, reps, 1rm")
	progressCmd.Flags().StringVar(&progressPeriod, "period", "all",
		"time period: week, month, year, all")
}

// ProgressData holds exercise progress data
type ProgressData struct {
	Exercise   string           `json:"exercise"`
	Metric     string           `json:"metric"`
	Unit       string           `json:"unit"`
	DataPoints []ProgressPoint  `json:"data_points"`
	Analysis   ProgressAnalysis `json:"analysis"`
}

// ProgressPoint is a single data point
type ProgressPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// ProgressAnalysis contains trend analysis
type ProgressAnalysis struct {
	StartingValue  float64 `json:"starting_value"`
	CurrentValue   float64 `json:"current_value"`
	AbsoluteChange float64 `json:"absolute_change"`
	PercentChange  float64 `json:"percent_change"`
	Trend          string  `json:"trend"`
}

func runProgress(cmd *cobra.Command, args []string) error {
	var exerciseName string
	if len(args) > 0 {
		exerciseName = args[0]
	} else {
		// Interactive mode - let user search and select an exercise
		cfg, err := config.Load("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		apiKey := cfg.GetAPIKey()
		if apiKey == "" {
			return fmt.Errorf("API key not configured. Run 'hevycli config init' to set up")
		}

		client := api.NewClient(apiKey)

		selected, err := prompt.SearchSelect(prompt.SearchSelectConfig{
			Title:       "Select an Exercise",
			Placeholder: "Search exercises...",
			Help:        "Type to filter by exercise name",
			LoadFunc: func() ([]prompt.SelectOption, error) {
				var allExercises []api.ExerciseTemplate
				page := 1
				for {
					resp, err := client.GetExerciseTemplates(page, 10)
					if err != nil {
						return nil, err
					}
					allExercises = append(allExercises, resp.ExerciseTemplates...)
					if page >= resp.PageCount || len(allExercises) > 500 {
						break
					}
					page++
				}
				options := make([]prompt.SelectOption, len(allExercises))
				for i, ex := range allExercises {
					options[i] = prompt.SelectOption{
						ID:          ex.Title, // Use title as ID since we need the name
						Title:       ex.Title,
						Description: ex.PrimaryMuscleGroup + " • " + ex.Equipment,
					}
				}
				return options, nil
			},
		})
		if err != nil {
			return err
		}
		exerciseName = selected.ID // ID is the title in this case
	}

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

	// Calculate date range
	now := time.Now()
	var startDate time.Time
	switch progressPeriod {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	case "all":
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		return fmt.Errorf("invalid period: %s", progressPeriod)
	}

	// Fetch all workouts
	fmt.Fprintln(os.Stderr, "Fetching workout data...")
	allWorkouts, err := client.GetAllWorkouts()
	if err != nil {
		return fmt.Errorf("failed to fetch workouts: %w", err)
	}

	// Find matching exercises and compute metric
	progressData := computeProgress(allWorkouts, exerciseName, progressMetric, startDate, now)

	if len(progressData.DataPoints) == 0 {
		return fmt.Errorf("no data found for exercise '%s'", exerciseName)
	}

	// Format output
	if outputFmt == "json" {
		out, err := formatter.Format(progressData)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		printProgressTable(progressData)
	}

	return nil
}

func computeProgress(workouts []api.Workout, exerciseName, metric string, start, end time.Time) ProgressData {
	var data ProgressData
	data.Metric = metric

	switch metric {
	case "1rm":
		data.Unit = "kg (estimated)"
	case "volume":
		data.Unit = "kg"
	case "reps":
		data.Unit = "reps"
	default:
		data.Unit = "kg"
	}

	// Collect data points by date
	dateValues := make(map[string]float64)
	var matchedExercise string

	for _, w := range workouts {
		if w.StartTime.Before(start) || w.StartTime.After(end) {
			continue
		}

		dateStr := w.StartTime.Format("2006-01-02")

		for _, ex := range w.Exercises {
			// Case-insensitive partial match
			if !strings.Contains(strings.ToLower(ex.Title), strings.ToLower(exerciseName)) {
				continue
			}

			if matchedExercise == "" {
				matchedExercise = ex.Title
			}

			var value float64
			switch metric {
			case "weight":
				value = computeMaxWeight(ex.Sets)
			case "volume":
				value = computeVolume(ex.Sets)
			case "reps":
				value = computeMaxReps(ex.Sets)
			case "1rm":
				value = computeEstimated1RM(ex.Sets)
			}

			// Keep the best value for each date
			if value > dateValues[dateStr] {
				dateValues[dateStr] = value
			}
		}
	}

	data.Exercise = matchedExercise
	if data.Exercise == "" {
		data.Exercise = exerciseName
	}

	// Convert to sorted data points
	var dates []string
	for date := range dateValues {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	for _, date := range dates {
		data.DataPoints = append(data.DataPoints, ProgressPoint{
			Date:  date,
			Value: math.Round(dateValues[date]*100) / 100,
		})
	}

	// Compute analysis
	if len(data.DataPoints) >= 2 {
		first := data.DataPoints[0].Value
		last := data.DataPoints[len(data.DataPoints)-1].Value

		data.Analysis.StartingValue = first
		data.Analysis.CurrentValue = last
		data.Analysis.AbsoluteChange = math.Round((last-first)*100) / 100

		if first > 0 {
			data.Analysis.PercentChange = math.Round((last-first)/first*10000) / 100
		}

		if last > first {
			data.Analysis.Trend = "increasing"
		} else if last < first {
			data.Analysis.Trend = "decreasing"
		} else {
			data.Analysis.Trend = "stable"
		}
	} else if len(data.DataPoints) == 1 {
		data.Analysis.StartingValue = data.DataPoints[0].Value
		data.Analysis.CurrentValue = data.DataPoints[0].Value
		data.Analysis.Trend = "insufficient_data"
	}

	return data
}

func computeMaxWeight(sets []api.Set) float64 {
	var maxWeight float64
	for _, set := range sets {
		if set.WeightKg != nil && *set.WeightKg > maxWeight {
			maxWeight = *set.WeightKg
		}
	}
	return maxWeight
}

func computeVolume(sets []api.Set) float64 {
	var volume float64
	for _, set := range sets {
		if set.WeightKg != nil && set.Reps != nil {
			volume += *set.WeightKg * float64(*set.Reps)
		}
	}
	return volume
}

func computeMaxReps(sets []api.Set) float64 {
	var maxReps int
	for _, set := range sets {
		if set.Reps != nil && *set.Reps > maxReps {
			maxReps = *set.Reps
		}
	}
	return float64(maxReps)
}

// computeEstimated1RM uses the Brzycki formula: 1RM = weight × (36 / (37 - reps))
func computeEstimated1RM(sets []api.Set) float64 {
	var max1RM float64
	for _, set := range sets {
		if set.WeightKg == nil || set.Reps == nil || *set.Reps == 0 {
			continue
		}
		weight := *set.WeightKg
		reps := *set.Reps

		// Brzycki formula works best for reps <= 10
		if reps > 10 {
			continue
		}

		estimated := weight * (36.0 / (37.0 - float64(reps)))
		if estimated > max1RM {
			max1RM = estimated
		}
	}
	return max1RM
}

func printProgressTable(data ProgressData) {
	fmt.Printf("\n📈 Progress: %s\n", data.Exercise)
	fmt.Printf("   Metric: %s (%s)\n\n", data.Metric, data.Unit)

	// Show data points
	fmt.Println("   Date         Value")
	fmt.Println("   ──────────────────────")
	for _, dp := range data.DataPoints {
		fmt.Printf("   %s    %.1f\n", dp.Date, dp.Value)
	}

	// Show analysis
	if data.Analysis.Trend != "" && data.Analysis.Trend != "insufficient_data" {
		fmt.Println("\n   📊 Analysis")
		fmt.Printf("   Starting:    %.1f %s\n", data.Analysis.StartingValue, data.Unit)
		fmt.Printf("   Current:     %.1f %s\n", data.Analysis.CurrentValue, data.Unit)

		changeSign := ""
		if data.Analysis.AbsoluteChange > 0 {
			changeSign = "+"
		}
		fmt.Printf("   Change:      %s%.1f (%s%.1f%%)\n",
			changeSign, data.Analysis.AbsoluteChange,
			changeSign, data.Analysis.PercentChange)

		trendEmoji := "➡️"
		switch data.Analysis.Trend {
		case "increasing":
			trendEmoji = "📈"
		case "decreasing":
			trendEmoji = "📉"
		}
		fmt.Printf("   Trend:       %s %s\n", trendEmoji, data.Analysis.Trend)
	}
	fmt.Println()
}
