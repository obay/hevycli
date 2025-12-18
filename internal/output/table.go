package output

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// TableFormatter formats output as a styled table
type TableFormatter struct {
	opts Options
}

// NewTableFormatter creates a new table formatter
func NewTableFormatter(opts Options) *TableFormatter {
	return &TableFormatter{opts: opts}
}

// Format converts data to a styled table format
func (f *TableFormatter) Format(data interface{}) (string, error) {
	// Check if data implements TableData interface
	td, ok := data.(TableData)
	if !ok {
		// Fall back to JSON for non-table data
		return NewJSONFormatter(f.opts).Format(data)
	}

	headers := td.Headers()
	rows := td.Rows()

	// Create the table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(f.borderStyle()).
		Headers(headers...).
		Rows(rows...)

	// Apply styling
	if !f.opts.NoColor {
		t = t.StyleFunc(f.styleFunc)
	}

	return t.Render(), nil
}

// FormatError formats an error for table output
func (f *TableFormatter) FormatError(err error) string {
	if f.opts.NoColor {
		return fmt.Sprintf("Error: %s", err.Error())
	}

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	return errorStyle.Render("Error: ") + err.Error()
}

// borderStyle returns the border style based on color settings
func (f *TableFormatter) borderStyle() lipgloss.Style {
	if f.opts.NoColor {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
}

// styleFunc returns a function for styling table cells
func (f *TableFormatter) styleFunc(row, col int) lipgloss.Style {
	if row == -1 {
		// Header row
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			Padding(0, 1)
	}

	// Alternate row colors for better readability
	if row%2 == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Padding(0, 1)
}

// SimpleTable is a helper struct for simple tabular data
type SimpleTable struct {
	headers []string
	rows    [][]string
}

// NewSimpleTable creates a new simple table
func NewSimpleTable(headers []string) *SimpleTable {
	return &SimpleTable{
		headers: headers,
		rows:    make([][]string, 0),
	}
}

// AddRow adds a row to the table
func (t *SimpleTable) AddRow(row ...string) {
	t.rows = append(t.rows, row)
}

// Headers returns the table headers
func (t *SimpleTable) Headers() []string {
	return t.headers
}

// Rows returns all table rows
func (t *SimpleTable) Rows() [][]string {
	return t.rows
}
