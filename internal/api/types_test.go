package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkout_Duration(t *testing.T) {
	workout := Workout{
		StartTime: time.Date(2024, 12, 17, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 12, 17, 11, 15, 0, 0, time.UTC),
	}

	duration := workout.Duration()
	assert.Equal(t, 75*time.Minute, duration)
}

func TestWorkout_ExerciseCount(t *testing.T) {
	workout := Workout{
		Exercises: []Exercise{
			{Title: "Bench Press"},
			{Title: "Squat"},
			{Title: "Deadlift"},
		},
	}

	assert.Equal(t, 3, workout.ExerciseCount())
}

func TestWorkout_TotalSets(t *testing.T) {
	workout := Workout{
		Exercises: []Exercise{
			{
				Title: "Bench Press",
				Sets: []Set{
					{Index: 0, SetType: SetTypeWarmup},
					{Index: 1, SetType: SetTypeNormal},
					{Index: 2, SetType: SetTypeNormal},
				},
			},
			{
				Title: "Squat",
				Sets: []Set{
					{Index: 0, SetType: SetTypeNormal},
					{Index: 1, SetType: SetTypeNormal},
				},
			},
		},
	}

	assert.Equal(t, 5, workout.TotalSets())
}

func TestSetType_Values(t *testing.T) {
	assert.Equal(t, SetType("normal"), SetTypeNormal)
	assert.Equal(t, SetType("warmup"), SetTypeWarmup)
	assert.Equal(t, SetType("dropset"), SetTypeDropset)
	assert.Equal(t, SetType("failure"), SetTypeFailure)
}

func TestEventType_Values(t *testing.T) {
	assert.Equal(t, EventType("updated"), EventTypeUpdated)
	assert.Equal(t, EventType("deleted"), EventTypeDeleted)
}
