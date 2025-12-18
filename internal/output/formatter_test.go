package output

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name       string
		format     FormatType
		expectType string
	}{
		{
			name:       "json formatter",
			format:     FormatJSON,
			expectType: "*output.JSONFormatter",
		},
		{
			name:       "table formatter",
			format:     FormatTable,
			expectType: "*output.TableFormatter",
		},
		{
			name:       "plain formatter",
			format:     FormatPlain,
			expectType: "*output.PlainFormatter",
		},
		{
			name:       "default is table",
			format:     "",
			expectType: "*output.TableFormatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(Options{
				Format: tt.format,
				Writer: os.Stdout,
			})
			assert.NotNil(t, f)
		})
	}
}

// SimpleTableImpl implements TableData for testing
type testTableData struct {
	headers []string
	rows    [][]string
}

func (t *testTableData) Headers() []string {
	return t.headers
}

func (t *testTableData) Rows() [][]string {
	return t.rows
}

func TestJSONFormatter_Format(t *testing.T) {
	f := NewJSONFormatter(Options{})

	// Test with map
	data := map[string]interface{}{
		"name":  "Test Workout",
		"count": 5,
	}

	result, err := f.Format(data)
	assert.NoError(t, err)
	assert.Contains(t, result, `"name": "Test Workout"`)
	assert.Contains(t, result, `"count": 5`)
}

func TestJSONFormatter_FormatError(t *testing.T) {
	f := NewJSONFormatter(Options{})

	result := f.FormatError(assert.AnError)
	assert.Contains(t, result, `"error"`)
	assert.Contains(t, result, `"message"`)
}

func TestPlainFormatter_Format(t *testing.T) {
	f := NewPlainFormatter(Options{})

	// Test with TableData
	data := &testTableData{
		headers: []string{"ID", "Name"},
		rows: [][]string{
			{"1", "First"},
			{"2", "Second"},
		},
	}

	result, err := f.Format(data)
	assert.NoError(t, err)
	assert.Equal(t, "1|First\n2|Second", result)
}

func TestPlainFormatter_FormatNonTable(t *testing.T) {
	f := NewPlainFormatter(Options{})

	result, err := f.Format("simple string")
	assert.NoError(t, err)
	assert.Equal(t, "simple string", result)
}

func TestPlainFormatter_FormatError(t *testing.T) {
	f := NewPlainFormatter(Options{})

	result := f.FormatError(assert.AnError)
	assert.Contains(t, result, "error:")
}

func TestTableFormatter_Format(t *testing.T) {
	f := NewTableFormatter(Options{NoColor: true})

	data := &testTableData{
		headers: []string{"ID", "Name"},
		rows: [][]string{
			{"1", "First"},
			{"2", "Second"},
		},
	}

	result, err := f.Format(data)
	assert.NoError(t, err)
	assert.Contains(t, result, "ID")
	assert.Contains(t, result, "Name")
	assert.Contains(t, result, "First")
	assert.Contains(t, result, "Second")
}

func TestTableFormatter_FormatNonTable(t *testing.T) {
	f := NewTableFormatter(Options{NoColor: true})

	// Non-table data should fall back to JSON
	data := map[string]string{"key": "value"}
	result, err := f.Format(data)
	assert.NoError(t, err)
	assert.Contains(t, result, "key")
	assert.Contains(t, result, "value")
}

func TestTableFormatter_FormatError(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
	}{
		{name: "with color", noColor: false},
		{name: "without color", noColor: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewTableFormatter(Options{NoColor: tt.noColor})
			result := f.FormatError(assert.AnError)
			assert.Contains(t, result, "Error:")
		})
	}
}

func TestSimpleTable(t *testing.T) {
	table := NewSimpleTable([]string{"Col1", "Col2"})
	table.AddRow("a", "b")
	table.AddRow("c", "d")

	assert.Equal(t, []string{"Col1", "Col2"}, table.Headers())
	assert.Len(t, table.Rows(), 2)
	assert.Equal(t, []string{"a", "b"}, table.Rows()[0])
	assert.Equal(t, []string{"c", "d"}, table.Rows()[1])
}
