package output

import (
	"io"
	"os"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	// Format converts data to the appropriate output format
	Format(data interface{}) (string, error)

	// FormatError formats an error for output
	FormatError(err error) string
}

// FormatType represents the output format type
type FormatType string

const (
	FormatJSON  FormatType = "json"
	FormatTable FormatType = "table"
	FormatPlain FormatType = "plain"
)

// Options holds formatting options
type Options struct {
	Format  FormatType
	NoColor bool
	Quiet   bool
	Verbose bool
	Writer  io.Writer
}

// TableData defines the interface for data that can be rendered as a table
type TableData interface {
	Headers() []string
	Rows() [][]string
}

// NewFormatter creates a formatter based on options
func NewFormatter(opts Options) Formatter {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	switch opts.Format {
	case FormatJSON:
		return NewJSONFormatter(opts)
	case FormatPlain:
		return NewPlainFormatter(opts)
	default:
		return NewTableFormatter(opts)
	}
}

// Print outputs the formatted data to the writer
func Print(f Formatter, data interface{}, w io.Writer) error {
	output, err := f.Format(data)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, output+"\n")
	return err
}

// PrintError outputs the formatted error to stderr
func PrintError(f Formatter, err error) {
	output := f.FormatError(err)
	io.WriteString(os.Stderr, output+"\n")
}
