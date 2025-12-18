package output

import (
	"fmt"
	"strings"
)

// PlainFormatter formats output as plain pipe-delimited text
type PlainFormatter struct {
	opts      Options
	separator string
}

// NewPlainFormatter creates a new plain text formatter
func NewPlainFormatter(opts Options) *PlainFormatter {
	return &PlainFormatter{
		opts:      opts,
		separator: "|",
	}
}

// Format converts data to plain text format
func (f *PlainFormatter) Format(data interface{}) (string, error) {
	// Check if data implements TableData interface
	td, ok := data.(TableData)
	if !ok {
		// For non-table data, just return string representation
		return fmt.Sprintf("%v", data), nil
	}

	var lines []string
	for _, row := range td.Rows() {
		lines = append(lines, strings.Join(row, f.separator))
	}
	return strings.Join(lines, "\n"), nil
}

// FormatError formats an error as plain text
func (f *PlainFormatter) FormatError(err error) string {
	return fmt.Sprintf("error: %s", err.Error())
}
