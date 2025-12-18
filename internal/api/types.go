package api

import "time"

// Workout represents a workout session
type Workout struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Exercises   []Exercise `json:"exercises"`
}

// Duration returns the workout duration
func (w *Workout) Duration() time.Duration {
	return w.EndTime.Sub(w.StartTime)
}

// ExerciseCount returns the number of exercises in the workout
func (w *Workout) ExerciseCount() int {
	return len(w.Exercises)
}

// TotalSets returns the total number of sets across all exercises
func (w *Workout) TotalSets() int {
	total := 0
	for _, e := range w.Exercises {
		total += len(e.Sets)
	}
	return total
}

// Exercise represents an exercise within a workout
type Exercise struct {
	Index              int    `json:"index"`
	Title              string `json:"title"`
	Notes              string `json:"notes,omitempty"`
	ExerciseTemplateID string `json:"exercise_template_id"`
	SupersetID         *int   `json:"superset_id,omitempty"`
	Sets               []Set  `json:"sets"`
}

// Set represents a single set of an exercise
type Set struct {
	Index           int      `json:"index"`
	SetType         SetType  `json:"type"`
	WeightKg        *float64 `json:"weight_kg,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	DistanceMeters  *float64 `json:"distance_meters,omitempty"`
	DurationSeconds *int     `json:"duration_seconds,omitempty"`
	RPE             *float64 `json:"rpe,omitempty"`
}

// SetType represents the type of set
type SetType string

const (
	SetTypeNormal  SetType = "normal"
	SetTypeWarmup  SetType = "warmup"
	SetTypeDropset SetType = "dropset"
	SetTypeFailure SetType = "failure"
)

// Routine represents a workout routine template
type Routine struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	FolderID  *string    `json:"folder_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Exercises []Exercise `json:"exercises"`
}

// RoutineFolder represents a folder for organizing routines
type RoutineFolder struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Index     int       `json:"index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ExerciseTemplate represents an exercise definition from the Hevy database
type ExerciseTemplate struct {
	ID                    string   `json:"id"`
	Title                 string   `json:"title"`
	Type                  string   `json:"type,omitempty"`
	PrimaryMuscleGroup    string   `json:"primary_muscle_group"`
	SecondaryMuscleGroups []string `json:"secondary_muscle_groups,omitempty"`
	Equipment             string   `json:"equipment,omitempty"`
	IsCustom              bool     `json:"is_custom"`
}

// WorkoutEvent represents a change event for workout sync
type WorkoutEvent struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	WorkoutID string    `json:"workout_id"`
	Timestamp time.Time `json:"timestamp"`
}

// EventType represents the type of workout event
type EventType string

const (
	EventTypeUpdated EventType = "updated"
	EventTypeDeleted EventType = "deleted"
)

// API Response types

// WorkoutsResponse represents the /workouts endpoint response
type WorkoutsResponse struct {
	Page      int       `json:"page"`
	PageCount int       `json:"page_count"`
	Workouts  []Workout `json:"workouts"`
}

// WorkoutCountResponse represents the /workouts/count endpoint response
type WorkoutCountResponse struct {
	WorkoutCount int `json:"workout_count"`
}

// RoutinesResponse represents the /routines endpoint response
type RoutinesResponse struct {
	Page      int       `json:"page"`
	PageCount int       `json:"page_count"`
	Routines  []Routine `json:"routines"`
}

// RoutineFoldersResponse represents the /routine_folders endpoint response
type RoutineFoldersResponse struct {
	Page          int             `json:"page"`
	PageCount     int             `json:"page_count"`
	RoutineFolders []RoutineFolder `json:"routine_folders"`
}

// ExerciseTemplatesResponse represents the /exercise_templates endpoint response
type ExerciseTemplatesResponse struct {
	Page              int                `json:"page"`
	PageCount         int                `json:"page_count"`
	ExerciseTemplates []ExerciseTemplate `json:"exercise_templates"`
}

// WorkoutEventsResponse represents the /workouts/events endpoint response
type WorkoutEventsResponse struct {
	Page         int            `json:"page"`
	PageCount    int            `json:"page_count"`
	WorkoutEvents []WorkoutEvent `json:"workout_events"`
}
