package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")
	assert.NotNil(t, client)
}

func TestNewClientWithOptions(t *testing.T) {
	client := NewClient("test-api-key",
		WithBaseURL("https://custom.api.com"),
	)
	assert.NotNil(t, client)
	assert.Equal(t, "https://custom.api.com", client.baseURL)
}

func TestValidateAuth_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key header
		assert.Equal(t, "test-api-key", r.Header.Get("api-key"))
		assert.Equal(t, "/workouts", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WorkoutsResponse{
			Page:      1,
			PageCount: 0,
			Workouts:  []Workout{},
		})
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))
	err := client.ValidateAuth()
	require.NoError(t, err)
}

func TestValidateAuth_InvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient("invalid-key", WithBaseURL(server.URL))
	err := client.ValidateAuth()
	require.Error(t, err)

	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, "INVALID_API_KEY", apiErr.ErrorCode)
}

func TestValidateAuth_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))
	err := client.ValidateAuth()
	require.Error(t, err)

	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", apiErr.ErrorCode)
}

func TestValidateAuth_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))
	err := client.ValidateAuth()
	require.Error(t, err)

	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, "RATE_LIMITED", apiErr.ErrorCode)
}

func TestGetWorkouts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/workouts", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("pageSize"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WorkoutsResponse{
			Page:      1,
			PageCount: 5,
			Workouts: []Workout{
				{ID: "workout-1", Title: "Push Day"},
				{ID: "workout-2", Title: "Pull Day"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))
	resp, err := client.GetWorkouts(1, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 5, resp.PageCount)
	assert.Len(t, resp.Workouts, 2)
	assert.Equal(t, "Push Day", resp.Workouts[0].Title)
}

func TestGetWorkout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/workouts/test-id-123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Workout{
			ID:    "test-id-123",
			Title: "Leg Day",
		})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))
	workout, err := client.GetWorkout("test-id-123")
	require.NoError(t, err)
	assert.Equal(t, "test-id-123", workout.ID)
	assert.Equal(t, "Leg Day", workout.Title)
}

func TestGetWorkout_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))
	_, err := client.GetWorkout("nonexistent")
	require.Error(t, err)

	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", apiErr.ErrorCode)
}

func TestGetWorkoutCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/workouts/count", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WorkoutCountResponse{WorkoutCount: 127})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))
	count, err := client.GetWorkoutCount()
	require.NoError(t, err)
	assert.Equal(t, 127, count)
}

func TestGetRoutines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/routines", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RoutinesResponse{
			Page:      1,
			PageCount: 1,
			Routines: []Routine{
				{ID: "routine-1", Title: "Upper Body"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))
	resp, err := client.GetRoutines(1, 10)
	require.NoError(t, err)
	assert.Len(t, resp.Routines, 1)
	assert.Equal(t, "Upper Body", resp.Routines[0].Title)
}

func TestGetExerciseTemplates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/exercise_templates", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ExerciseTemplatesResponse{
			Page:      1,
			PageCount: 10,
			ExerciseTemplates: []ExerciseTemplate{
				{ID: "ex-1", Title: "Bench Press", PrimaryMuscleGroup: "Chest"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))
	resp, err := client.GetExerciseTemplates(1, 10)
	require.NoError(t, err)
	assert.Len(t, resp.ExerciseTemplates, 1)
	assert.Equal(t, "Bench Press", resp.ExerciseTemplates[0].Title)
}
