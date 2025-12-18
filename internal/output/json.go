package output

import (
	"encoding/json"
	"time"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	opts   Options
	indent bool
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(opts Options) *JSONFormatter {
	return &JSONFormatter{
		opts:   opts,
		indent: true,
	}
}

// Format converts data to JSON format
func (f *JSONFormatter) Format(data interface{}) (string, error) {
	var b []byte
	var err error

	if f.indent {
		b, err = json.MarshalIndent(data, "", "  ")
	} else {
		b, err = json.Marshal(data)
	}

	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FormatError formats an error as JSON
func (f *JSONFormatter) FormatError(err error) string {
	errObj := ErrorResponse{
		Error: ErrorDetail{
			Message:   err.Error(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Try to extract code if it's an APIError
	if apiErr, ok := err.(interface{ Code() string }); ok {
		errObj.Error.Code = apiErr.Code()
	}

	b, _ := json.MarshalIndent(errObj, "", "  ")
	return string(b)
}

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code      string `json:"code,omitempty"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Timestamp string `json:"timestamp"`
}
