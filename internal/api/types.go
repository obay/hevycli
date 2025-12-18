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
	Page          int            `json:"page"`
	PageCount     int            `json:"page_count"`
	WorkoutEvents []WorkoutEvent `json:"workout_events"`
}

// ---- Request types for POST/PUT operations ----

// CreateWorkoutRequest represents the request body for POST /workouts
type CreateWorkoutRequest struct {
	Workout CreateWorkoutData `json:"workout"`
}

// CreateWorkoutData represents the workout data for creation
type CreateWorkoutData struct {
	Title       string                  `json:"title"`
	Description *string                 `json:"description,omitempty"`
	StartTime   string                  `json:"start_time"`
	EndTime     string                  `json:"end_time"`
	IsPrivate   bool                    `json:"is_private,omitempty"`
	Exercises   []CreateWorkoutExercise `json:"exercises"`
}

// CreateWorkoutExercise represents an exercise in a workout creation request
type CreateWorkoutExercise struct {
	ExerciseTemplateID string             `json:"exercise_template_id"`
	SupersetID         *int               `json:"superset_id,omitempty"`
	Notes              *string            `json:"notes,omitempty"`
	Sets               []CreateWorkoutSet `json:"sets"`
}

// CreateWorkoutSet represents a set in a workout creation request
type CreateWorkoutSet struct {
	Type            SetType  `json:"type"`
	WeightKg        *float64 `json:"weight_kg,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	DistanceMeters  *int     `json:"distance_meters,omitempty"`
	DurationSeconds *int     `json:"duration_seconds,omitempty"`
	CustomMetric    *float64 `json:"custom_metric,omitempty"`
	RPE             *float64 `json:"rpe,omitempty"`
}

// UpdateWorkoutRequest represents the request body for PUT /workouts/{id}
type UpdateWorkoutRequest struct {
	Workout UpdateWorkoutData `json:"workout"`
}

// UpdateWorkoutData represents the workout data for update
type UpdateWorkoutData struct {
	Title       string                  `json:"title"`
	Description *string                 `json:"description,omitempty"`
	StartTime   string                  `json:"start_time"`
	EndTime     string                  `json:"end_time"`
	Exercises   []CreateWorkoutExercise `json:"exercises"`
}

// CreateRoutineRequest represents the request body for POST /routines
type CreateRoutineRequest struct {
	Routine CreateRoutineData `json:"routine"`
}

// CreateRoutineData represents the routine data for creation
type CreateRoutineData struct {
	Title     string                  `json:"title"`
	FolderID  *int                    `json:"folder_id,omitempty"`
	Notes     *string                 `json:"notes,omitempty"`
	Exercises []CreateRoutineExercise `json:"exercises"`
}

// CreateRoutineExercise represents an exercise in a routine creation request
type CreateRoutineExercise struct {
	ExerciseTemplateID string             `json:"exercise_template_id"`
	SupersetID         *int               `json:"superset_id,omitempty"`
	RestSeconds        *int               `json:"rest_seconds,omitempty"`
	Notes              *string            `json:"notes,omitempty"`
	Sets               []CreateRoutineSet `json:"sets"`
}

// CreateRoutineSet represents a set in a routine creation request
type CreateRoutineSet struct {
	Type            SetType   `json:"type"`
	WeightKg        *float64  `json:"weight_kg,omitempty"`
	Reps            *int      `json:"reps,omitempty"`
	DistanceMeters  *int      `json:"distance_meters,omitempty"`
	DurationSeconds *int      `json:"duration_seconds,omitempty"`
	CustomMetric    *float64  `json:"custom_metric,omitempty"`
	RepRange        *RepRange `json:"rep_range,omitempty"`
}

// RepRange represents a range of reps for a set
type RepRange struct {
	Start *int `json:"start,omitempty"`
	End   *int `json:"end,omitempty"`
}

// UpdateRoutineRequest represents the request body for PUT /routines/{id}
type UpdateRoutineRequest struct {
	Routine UpdateRoutineData `json:"routine"`
}

// UpdateRoutineData represents the routine data for update
type UpdateRoutineData struct {
	Title     string                  `json:"title"`
	Notes     *string                 `json:"notes,omitempty"`
	Exercises []CreateRoutineExercise `json:"exercises"`
}

// CreateRoutineFolderRequest represents the request body for POST /routine_folders
type CreateRoutineFolderRequest struct {
	RoutineFolder CreateRoutineFolderData `json:"routine_folder"`
}

// CreateRoutineFolderData represents the folder data for creation
type CreateRoutineFolderData struct {
	Title string `json:"title"`
}

// WorkoutResponse represents the response from POST/PUT /workouts
type WorkoutResponse struct {
	Workout Workout `json:"workout"`
}

// RoutineResponse represents the response from POST/PUT /routines
type RoutineResponse struct {
	Routine Routine `json:"routine"`
}

// RoutineFolderResponse represents the response from POST /routine_folders
type RoutineFolderResponse struct {
	RoutineFolder RoutineFolder `json:"routine_folder"`
}

// ---- Custom Exercise Types ----

// ExerciseType represents the type of exercise tracking
type ExerciseType string

const (
	ExerciseTypeWeightReps           ExerciseType = "weight_reps"
	ExerciseTypeRepsOnly             ExerciseType = "reps_only"
	ExerciseTypeBodyweightReps       ExerciseType = "bodyweight_reps"
	ExerciseTypeBodyweightAssisted   ExerciseType = "bodyweight_assisted_reps"
	ExerciseTypeDuration             ExerciseType = "duration"
	ExerciseTypeWeightDuration       ExerciseType = "weight_duration"
	ExerciseTypeDistanceDuration     ExerciseType = "distance_duration"
	ExerciseTypeShortDistanceWeight  ExerciseType = "short_distance_weight"
)

// MuscleGroup represents muscle groups for exercises
type MuscleGroup string

const (
	MuscleGroupAbdominals MuscleGroup = "abdominals"
	MuscleGroupShoulders  MuscleGroup = "shoulders"
	MuscleGroupBiceps     MuscleGroup = "biceps"
	MuscleGroupTriceps    MuscleGroup = "triceps"
	MuscleGroupForearms   MuscleGroup = "forearms"
	MuscleGroupQuadriceps MuscleGroup = "quadriceps"
	MuscleGroupHamstrings MuscleGroup = "hamstrings"
	MuscleGroupCalves     MuscleGroup = "calves"
	MuscleGroupGlutes     MuscleGroup = "glutes"
	MuscleGroupAbductors  MuscleGroup = "abductors"
	MuscleGroupAdductors  MuscleGroup = "adductors"
	MuscleGroupLats       MuscleGroup = "lats"
	MuscleGroupUpperBack  MuscleGroup = "upper_back"
	MuscleGroupTraps      MuscleGroup = "traps"
	MuscleGroupLowerBack  MuscleGroup = "lower_back"
	MuscleGroupChest      MuscleGroup = "chest"
	MuscleGroupCardio     MuscleGroup = "cardio"
	MuscleGroupNeck       MuscleGroup = "neck"
	MuscleGroupFullBody   MuscleGroup = "full_body"
	MuscleGroupOther      MuscleGroup = "other"
)

// EquipmentCategory represents equipment types
type EquipmentCategory string

const (
	EquipmentNone           EquipmentCategory = "none"
	EquipmentBarbell        EquipmentCategory = "barbell"
	EquipmentDumbbell       EquipmentCategory = "dumbbell"
	EquipmentKettlebell     EquipmentCategory = "kettlebell"
	EquipmentMachine        EquipmentCategory = "machine"
	EquipmentPlate          EquipmentCategory = "plate"
	EquipmentResistanceBand EquipmentCategory = "resistance_band"
	EquipmentSuspension     EquipmentCategory = "suspension"
	EquipmentOther          EquipmentCategory = "other"
)

// CreateCustomExerciseRequest represents the request body for POST /exercise_templates
type CreateCustomExerciseRequest struct {
	Exercise CreateCustomExerciseData `json:"exercise"`
}

// CreateCustomExerciseData represents the custom exercise data for creation
type CreateCustomExerciseData struct {
	Title             string            `json:"title"`
	ExerciseType      ExerciseType      `json:"exercise_type"`
	EquipmentCategory EquipmentCategory `json:"equipment_category"`
	MuscleGroup       MuscleGroup       `json:"muscle_group"`
	OtherMuscles      []MuscleGroup     `json:"other_muscles,omitempty"`
}

// ExerciseTemplateResponse represents the response from POST /exercise_templates
type ExerciseTemplateResponse struct {
	ExerciseTemplate ExerciseTemplate `json:"exercise_template"`
}
