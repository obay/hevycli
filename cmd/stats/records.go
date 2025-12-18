package stats

import (
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/config"
	"github.com/obay/hevycli/internal/output"
)

var (
	recordsExercise string
	recordsLimit    int
)

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "View personal records",
	Long: `Display your personal records across all exercises or for a specific exercise.

Record types:
  - Max weight lifted
  - Estimated 1RM (one-rep max)
  - Max volume in a single session

Examples:
  hevycli stats records                        # Top 10 PRs
  hevycli stats records --limit 20             # Top 20 PRs
  hevycli stats records --exercise "Bench"     # Bench press PRs only`,
	RunE: runRecords,
}

func init() {
	recordsCmd.Flags().StringVar(&recordsExercise, "exercise", "",
		"filter by exercise name")
	recordsCmd.Flags().IntVar(&recordsLimit, "limit", 10,
		"number of records to show")
}

// PersonalRecord represents a personal record
type PersonalRecord struct {
	Exercise   string  `json:"exercise"`
	RecordType string  `json:"record_type"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Reps       int     `json:"reps,omitempty"`
	Date       string  `json:"date"`
	WorkoutID  string  `json:"workout_id"`
}

// RecordsData holds all personal records
type RecordsData struct {
	PersonalRecords []PersonalRecord `json:"personal_records"`
}

func runRecords(cmd *cobra.Command, args []string) error {
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

	// Fetch all workouts
	fmt.Fprintln(os.Stderr, "Fetching workout data...")
	allWorkouts, err := client.GetAllWorkouts()
	if err != nil {
		return fmt.Errorf("failed to fetch workouts: %w", err)
	}

	// Compute records
	records := computeRecords(allWorkouts, recordsExercise, recordsLimit)

	if len(records.PersonalRecords) == 0 {
		fmt.Println("No personal records found.")
		return nil
	}

	// Format output
	if outputFmt == "json" {
		out, err := formatter.Format(records)
		if err != nil {
			return err
		}
		fmt.Println(out)
	} else {
		printRecordsTable(records)
	}

	return nil
}

type exerciseRecord struct {
	maxWeight     float64
	maxWeightReps int
	maxWeightDate string
	maxWeightWID  string

	max1RM     float64
	max1RMReps int
	max1RMDate string
	max1RMWID  string

	maxVolume     float64
	maxVolumeDate string
	maxVolumeWID  string
}

func computeRecords(workouts []api.Workout, exerciseFilter string, limit int) RecordsData {
	// Track best records per exercise
	exerciseRecords := make(map[string]*exerciseRecord)

	for _, w := range workouts {
		dateStr := w.StartTime.Format("2006-01-02")

		for _, ex := range w.Exercises {
			// Apply exercise filter if specified
			if exerciseFilter != "" {
				if !containsIgnoreCase(ex.Title, exerciseFilter) {
					continue
				}
			}

			if exerciseRecords[ex.Title] == nil {
				exerciseRecords[ex.Title] = &exerciseRecord{}
			}
			rec := exerciseRecords[ex.Title]

			// Calculate session metrics
			var sessionVolume float64
			var sessionMaxWeight float64
			var sessionMaxWeightReps int
			var session1RM float64
			var session1RMReps int

			for _, set := range ex.Sets {
				if set.WeightKg == nil {
					continue
				}
				weight := *set.WeightKg
				reps := 0
				if set.Reps != nil {
					reps = *set.Reps
				}

				// Max weight
				if weight > sessionMaxWeight {
					sessionMaxWeight = weight
					sessionMaxWeightReps = reps
				}

				// Volume
				if reps > 0 {
					sessionVolume += weight * float64(reps)
				}

				// Estimated 1RM (Brzycki formula)
				if reps > 0 && reps <= 10 {
					estimated := weight * (36.0 / (37.0 - float64(reps)))
					if estimated > session1RM {
						session1RM = estimated
						session1RMReps = reps
					}
				}
			}

			// Update records
			if sessionMaxWeight > rec.maxWeight {
				rec.maxWeight = sessionMaxWeight
				rec.maxWeightReps = sessionMaxWeightReps
				rec.maxWeightDate = dateStr
				rec.maxWeightWID = w.ID
			}

			if session1RM > rec.max1RM {
				rec.max1RM = session1RM
				rec.max1RMReps = session1RMReps
				rec.max1RMDate = dateStr
				rec.max1RMWID = w.ID
			}

			if sessionVolume > rec.maxVolume {
				rec.maxVolume = sessionVolume
				rec.maxVolumeDate = dateStr
				rec.maxVolumeWID = w.ID
			}
		}
	}

	// Convert to PersonalRecord slice
	var allRecords []PersonalRecord

	for exercise, rec := range exerciseRecords {
		if rec.maxWeight > 0 {
			allRecords = append(allRecords, PersonalRecord{
				Exercise:   exercise,
				RecordType: "weight",
				Value:      math.Round(rec.maxWeight*10) / 10,
				Unit:       "kg",
				Reps:       rec.maxWeightReps,
				Date:       rec.maxWeightDate,
				WorkoutID:  rec.maxWeightWID,
			})
		}

		if rec.max1RM > 0 {
			allRecords = append(allRecords, PersonalRecord{
				Exercise:   exercise,
				RecordType: "estimated_1rm",
				Value:      math.Round(rec.max1RM*10) / 10,
				Unit:       "kg",
				Reps:       rec.max1RMReps,
				Date:       rec.max1RMDate,
				WorkoutID:  rec.max1RMWID,
			})
		}
	}

	// Sort by value descending
	sort.Slice(allRecords, func(i, j int) bool {
		return allRecords[i].Value > allRecords[j].Value
	})

	// Apply limit
	if len(allRecords) > limit {
		allRecords = allRecords[:limit]
	}

	return RecordsData{PersonalRecords: allRecords}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(substr) == 0 ||
			(len(s) > 0 && containsLower(toLower(s), toLower(substr))))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func printRecordsTable(data RecordsData) {
	fmt.Println("\n🏆 Personal Records")

	fmt.Println("   Exercise                          Type          Value        Date")
	fmt.Println("   ─────────────────────────────────────────────────────────────────────")

	for _, pr := range data.PersonalRecords {
		exercise := pr.Exercise
		if len(exercise) > 32 {
			exercise = exercise[:29] + "..."
		}

		recordType := pr.RecordType
		if recordType == "estimated_1rm" {
			recordType = "est. 1RM"
		}

		valueStr := fmt.Sprintf("%.1f %s", pr.Value, pr.Unit)
		if pr.Reps > 0 && pr.RecordType == "weight" {
			valueStr = fmt.Sprintf("%.1f %s × %d", pr.Value, pr.Unit, pr.Reps)
		}

		fmt.Printf("   %-32s  %-12s  %-12s  %s\n",
			exercise, recordType, valueStr, pr.Date)
	}
	fmt.Println()
}
