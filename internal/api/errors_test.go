package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "without details",
			err: &APIError{
				ErrorCode:    "TEST_ERROR",
				ErrorMessage: "Test message",
			},
			expected: "TEST_ERROR: Test message",
		},
		{
			name: "with details",
			err: &APIError{
				ErrorCode:    "TEST_ERROR",
				ErrorMessage: "Test message",
				ErrorDetails: "Additional details here",
			},
			expected: "TEST_ERROR: Test message (Additional details here)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestAPIError_Code(t *testing.T) {
	err := &APIError{
		ErrorCode:    "MY_CODE",
		ErrorMessage: "Some message",
	}
	assert.Equal(t, "MY_CODE", err.Code())
}

func TestAPIError_ExitCode(t *testing.T) {
	tests := []struct {
		code     string
		expected int
	}{
		{"INVALID_API_KEY", ExitAuthError},
		{"UNAUTHORIZED", ExitAuthError},
		{"RATE_LIMITED", ExitRateLimited},
		{"NETWORK_ERROR", ExitNetworkError},
		{"NOT_FOUND", ExitNotFound},
		{"VALIDATION_ERROR", ExitValidationError},
		{"UNKNOWN_ERROR", ExitGeneralError},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			err := &APIError{ErrorCode: tt.code}
			assert.Equal(t, tt.expected, err.ExitCode())
		})
	}
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError("CODE", "message")
	assert.Equal(t, "CODE", err.ErrorCode)
	assert.Equal(t, "message", err.ErrorMessage)
	assert.Empty(t, err.ErrorDetails)
}

func TestNewAPIErrorWithDetails(t *testing.T) {
	err := NewAPIErrorWithDetails("CODE", "message", "details")
	assert.Equal(t, "CODE", err.ErrorCode)
	assert.Equal(t, "message", err.ErrorMessage)
	assert.Equal(t, "details", err.ErrorDetails)
}

func TestPredefinedErrors(t *testing.T) {
	assert.Equal(t, "INVALID_API_KEY", ErrInvalidAPIKey.ErrorCode)
	assert.Equal(t, "FORBIDDEN", ErrForbidden.ErrorCode)
	assert.Equal(t, "NOT_FOUND", ErrNotFound.ErrorCode)
	assert.Equal(t, "RATE_LIMITED", ErrRateLimited.ErrorCode)
	assert.Equal(t, "NETWORK_ERROR", ErrNetworkError.ErrorCode)
}
