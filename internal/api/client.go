package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// DefaultBaseURL is the default Hevy API base URL
	DefaultBaseURL = "https://api.hevyapp.com/v1"

	// DefaultTimeout is the default request timeout
	DefaultTimeout = 30 * time.Second

	// UserAgent is the user agent string for API requests
	UserAgent = "hevycli/1.0"
)

// Client represents the Hevy API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *resty.Client
}

// ClientOption is a function that configures the client
type ClientOption func(*Client)

// NewClient creates a new Hevy API client
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		apiKey:  apiKey,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.httpClient = resty.New().
		SetBaseURL(c.baseURL).
		SetHeader("api-key", c.apiKey).
		SetHeader("User-Agent", UserAgent).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetTimeout(DefaultTimeout)

	return c
}

// WithBaseURL sets a custom base URL
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithTimeout sets a custom timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if c.httpClient != nil {
			c.httpClient.SetTimeout(timeout)
		}
	}
}

// ValidateAuth tests if the API key is valid by calling /workouts endpoint
func (c *Client) ValidateAuth() error {
	var result WorkoutsResponse
	resp, err := c.httpClient.R().
		SetQueryParams(map[string]string{
			"page":     "1",
			"pageSize": "1",
		}).
		SetResult(&result).
		Get("/workouts")

	if err != nil {
		return &APIError{
			ErrorCode:    "NETWORK_ERROR",
			ErrorMessage: fmt.Sprintf("failed to connect to API: %v", err),
		}
	}

	return c.handleResponse(resp)
}

// GetWorkouts fetches workouts with pagination
func (c *Client) GetWorkouts(page, pageSize int) (*WorkoutsResponse, error) {
	var result WorkoutsResponse
	resp, err := c.httpClient.R().
		SetQueryParams(map[string]string{
			"page":     fmt.Sprintf("%d", page),
			"pageSize": fmt.Sprintf("%d", pageSize),
		}).
		SetResult(&result).
		Get("/workouts")

	if err != nil {
		return nil, &APIError{
			ErrorCode:    "NETWORK_ERROR",
			ErrorMessage: fmt.Sprintf("failed to fetch workouts: %v", err),
		}
	}

	if err := c.handleResponse(resp); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWorkout fetches a single workout by ID
func (c *Client) GetWorkout(id string) (*Workout, error) {
	var result Workout
	resp, err := c.httpClient.R().
		SetResult(&result).
		Get("/workouts/" + id)

	if err != nil {
		return nil, &APIError{
			ErrorCode:    "NETWORK_ERROR",
			ErrorMessage: fmt.Sprintf("failed to fetch workout: %v", err),
		}
	}

	if err := c.handleResponse(resp); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWorkoutCount fetches the total number of workouts
func (c *Client) GetWorkoutCount() (int, error) {
	var result WorkoutCountResponse
	resp, err := c.httpClient.R().
		SetResult(&result).
		Get("/workouts/count")

	if err != nil {
		return 0, &APIError{
			ErrorCode:    "NETWORK_ERROR",
			ErrorMessage: fmt.Sprintf("failed to fetch workout count: %v", err),
		}
	}

	if err := c.handleResponse(resp); err != nil {
		return 0, err
	}

	return result.WorkoutCount, nil
}

// GetRoutines fetches all routines
func (c *Client) GetRoutines(page, pageSize int) (*RoutinesResponse, error) {
	var result RoutinesResponse
	resp, err := c.httpClient.R().
		SetQueryParams(map[string]string{
			"page":     fmt.Sprintf("%d", page),
			"pageSize": fmt.Sprintf("%d", pageSize),
		}).
		SetResult(&result).
		Get("/routines")

	if err != nil {
		return nil, &APIError{
			ErrorCode:    "NETWORK_ERROR",
			ErrorMessage: fmt.Sprintf("failed to fetch routines: %v", err),
		}
	}

	if err := c.handleResponse(resp); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRoutine fetches a single routine by ID
func (c *Client) GetRoutine(id string) (*Routine, error) {
	var result struct {
		Routine Routine `json:"routine"`
	}
	resp, err := c.httpClient.R().
		SetResult(&result).
		Get("/routines/" + id)

	if err != nil {
		return nil, &APIError{
			ErrorCode:    "NETWORK_ERROR",
			ErrorMessage: fmt.Sprintf("failed to fetch routine: %v", err),
		}
	}

	if err := c.handleResponse(resp); err != nil {
		return nil, err
	}

	return &result.Routine, nil
}

// GetExerciseTemplates fetches exercise templates with pagination
func (c *Client) GetExerciseTemplates(page, pageSize int) (*ExerciseTemplatesResponse, error) {
	var result ExerciseTemplatesResponse
	resp, err := c.httpClient.R().
		SetQueryParams(map[string]string{
			"page":     fmt.Sprintf("%d", page),
			"pageSize": fmt.Sprintf("%d", pageSize),
		}).
		SetResult(&result).
		Get("/exercise_templates")

	if err != nil {
		return nil, &APIError{
			ErrorCode:    "NETWORK_ERROR",
			ErrorMessage: fmt.Sprintf("failed to fetch exercise templates: %v", err),
		}
	}

	if err := c.handleResponse(resp); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRoutineFolders fetches routine folders
func (c *Client) GetRoutineFolders(page, pageSize int) (*RoutineFoldersResponse, error) {
	var result RoutineFoldersResponse
	resp, err := c.httpClient.R().
		SetQueryParams(map[string]string{
			"page":     fmt.Sprintf("%d", page),
			"pageSize": fmt.Sprintf("%d", pageSize),
		}).
		SetResult(&result).
		Get("/routine_folders")

	if err != nil {
		return nil, &APIError{
			ErrorCode:    "NETWORK_ERROR",
			ErrorMessage: fmt.Sprintf("failed to fetch routine folders: %v", err),
		}
	}

	if err := c.handleResponse(resp); err != nil {
		return nil, err
	}

	return &result, nil
}

// handleResponse checks for API errors based on HTTP status code
func (c *Client) handleResponse(resp *resty.Response) error {
	switch resp.StatusCode() {
	case http.StatusOK, http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		return ErrInvalidAPIKey
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		if resp.StatusCode() >= 400 {
			return &APIError{
				ErrorCode:    "API_ERROR",
				ErrorMessage: fmt.Sprintf("API returned status %d: %s", resp.StatusCode(), resp.String()),
			}
		}
		return nil
	}
}
