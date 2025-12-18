package api

import "fmt"

// Exit codes for different error types
const (
	ExitSuccess         = 0
	ExitGeneralError    = 1
	ExitInvalidArgs     = 2
	ExitAuthError       = 3
	ExitRateLimited     = 4
	ExitNetworkError    = 5
	ExitNotFound        = 6
	ExitValidationError = 7
)

// APIError represents an error from the Hevy API
type APIError struct {
	ErrorCode    string `json:"code"`
	ErrorMessage string `json:"message"`
	ErrorDetails string `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.ErrorDetails != "" {
		return fmt.Sprintf("%s: %s (%s)", e.ErrorCode, e.ErrorMessage, e.ErrorDetails)
	}
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.ErrorMessage)
}

// Code returns the error code
func (e *APIError) Code() string {
	return e.ErrorCode
}

// ExitCode returns the appropriate exit code for this error
func (e *APIError) ExitCode() int {
	switch e.ErrorCode {
	case "INVALID_API_KEY", "UNAUTHORIZED":
		return ExitAuthError
	case "RATE_LIMITED":
		return ExitRateLimited
	case "NETWORK_ERROR":
		return ExitNetworkError
	case "NOT_FOUND":
		return ExitNotFound
	case "VALIDATION_ERROR":
		return ExitValidationError
	default:
		return ExitGeneralError
	}
}

// NewAPIError creates a new API error
func NewAPIError(code, message string) *APIError {
	return &APIError{
		ErrorCode:    code,
		ErrorMessage: message,
	}
}

// NewAPIErrorWithDetails creates a new API error with additional details
func NewAPIErrorWithDetails(code, message, details string) *APIError {
	return &APIError{
		ErrorCode:    code,
		ErrorMessage: message,
		ErrorDetails: details,
	}
}

// Common API errors
var (
	ErrInvalidAPIKey = NewAPIErrorWithDetails(
		"INVALID_API_KEY",
		"The provided API key is invalid or expired",
		"Please verify your API key at https://hevy.com/settings?developer",
	)

	ErrForbidden = NewAPIErrorWithDetails(
		"FORBIDDEN",
		"Access forbidden",
		"Hevy Pro subscription required for API access",
	)

	ErrNotFound = NewAPIError(
		"NOT_FOUND",
		"Resource not found",
	)

	ErrRateLimited = NewAPIError(
		"RATE_LIMITED",
		"Rate limit exceeded. Please try again later",
	)

	ErrNetworkError = NewAPIError(
		"NETWORK_ERROR",
		"Failed to connect to Hevy API",
	)
)
