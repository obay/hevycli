package exercise

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/tui/common"
)

// TableResult represents the result of the table interaction
type TableResult struct {
	Selected  *api.ExerciseTemplate
	Cancelled bool
}

// TableModel is the interactive exercise table model
type TableModel struct {
	client       *api.Client
	table        table.Model
	textInput    textinput.Model
	allExercises []api.ExerciseTemplate
	filtered     []api.ExerciseTemplate
	loading      bool
	err          error
	quitting     bool
	selected     *api.ExerciseTemplate
	searchFocus  bool
	width        int
	height       int
}

// NewTableModel creates a new exercise table model
func NewTableModel(client *api.Client) TableModel {
	// Search input
	ti := textinput.New()
	ti.Placeholder = "Type to search exercises..."
	ti.CharLimit = 100
	ti.Width = 50
	ti.PromptStyle = common.FocusedStyle
	ti.TextStyle = common.NormalItemStyle

	// Table columns
	columns := []table.Column{
		{Title: "Title", Width: 40},
		{Title: "Primary Muscle", Width: 20},
		{Title: "Equipment", Width: 15},
		{Title: "Custom", Width: 6},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(common.MutedColor).
		BorderBottom(true).
		Bold(true).
		Foreground(common.PrimaryColor)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(common.PrimaryColor).
		Bold(true)
	t.SetStyles(s)

	return TableModel{
		client:      client,
		table:       t,
		textInput:   ti,
		loading:     true,
		searchFocus: false,
	}
}

// tableExercisesLoadedMsg is sent when exercises are loaded
type tableExercisesLoadedMsg struct {
	exercises []api.ExerciseTemplate
}

// tableLoadErrorMsg is sent when loading fails
type tableLoadErrorMsg struct {
	err error
}

// tableLoadExercises fetches all exercises from the API
func tableLoadExercises(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		var all []api.ExerciseTemplate
		page := 1
		for {
			resp, err := client.GetExerciseTemplates(page, 10)
			if err != nil {
				return tableLoadErrorMsg{err: err}
			}
			all = append(all, resp.ExerciseTemplates...)
			if page >= resp.PageCount || resp.PageCount == 0 {
				break
			}
			page++
		}
		return tableExercisesLoadedMsg{exercises: all}
	}
}

// Init initializes the model
func (m TableModel) Init() tea.Cmd {
	return tableLoadExercises(m.client)
}

// Update handles messages
func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchFocus {
			switch msg.String() {
			case "esc":
				m.searchFocus = false
				m.textInput.Blur()
				return m, nil
			case "enter":
				m.searchFocus = false
				m.textInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				prevValue := m.textInput.Value()
				m.textInput, cmd = m.textInput.Update(msg)
				if m.textInput.Value() != prevValue {
					m.filterExercises()
				}
				return m, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.textInput.Value() != "" {
				m.textInput.SetValue("")
				m.filterExercises()
			} else {
				m.quitting = true
				return m, tea.Quit
			}
		case "/":
			m.searchFocus = true
			m.textInput.Focus()
			return m, textinput.Blink
		case "enter":
			if !m.loading && len(m.filtered) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.filtered) {
					m.selected = &m.filtered[idx]
					return m, tea.Quit
				}
			}
		}

		// Pass navigation keys to table
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(msg.Height - 10)
		m.table.SetWidth(msg.Width - 4)
		m.textInput.Width = min(msg.Width-20, 60)

	case tableExercisesLoadedMsg:
		m.allExercises = msg.exercises
		m.filtered = msg.exercises
		m.loading = false
		m.updateTableRows()

	case tableLoadErrorMsg:
		m.err = msg.err
		m.loading = false
	}

	return m, tea.Batch(cmds...)
}

// filterExercises filters exercises based on search input
func (m *TableModel) filterExercises() {
	query := strings.ToLower(m.textInput.Value())
	if query == "" {
		m.filtered = m.allExercises
	} else {
		m.filtered = nil
		for _, ex := range m.allExercises {
			if strings.Contains(strings.ToLower(ex.Title), query) ||
				strings.Contains(strings.ToLower(ex.PrimaryMuscleGroup), query) ||
				strings.Contains(strings.ToLower(ex.Equipment), query) {
				m.filtered = append(m.filtered, ex)
			}
		}
	}
	m.updateTableRows()
}

// updateTableRows updates the table with filtered exercises
func (m *TableModel) updateTableRows() {
	rows := make([]table.Row, len(m.filtered))
	for i, ex := range m.filtered {
		custom := "No"
		if ex.IsCustom {
			custom = "Yes"
		}
		rows[i] = table.Row{
			truncate(ex.Title, 38),
			ex.PrimaryMuscleGroup,
			ex.Equipment,
			custom,
		}
	}
	m.table.SetRows(rows)
}

// View renders the model
func (m TableModel) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.PrimaryColor).
		MarginBottom(1)
	b.WriteString(titleStyle.Render("Exercise Templates"))
	b.WriteString("\n\n")

	// Search input
	searchLabel := "Search: "
	if m.searchFocus {
		searchLabel = lipgloss.NewStyle().Foreground(common.PrimaryColor).Render("Search: ")
	}
	b.WriteString(searchLabel)
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Loading or error state
	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(common.MutedColor).Render("Loading exercises..."))
		b.WriteString("\n")
	} else if m.err != nil {
		b.WriteString(common.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	} else {
		// Count
		countStyle := lipgloss.NewStyle().Foreground(common.MutedColor)
		b.WriteString(countStyle.Render(fmt.Sprintf("%d exercise(s)", len(m.filtered))))
		b.WriteString("\n\n")

		// Table
		b.WriteString(m.table.View())
	}

	// Help
	b.WriteString("\n")
	helpText := "↑/↓ navigate • enter select • / search • esc clear/quit • q quit"
	b.WriteString(common.HelpStyle.Render(helpText))

	return b.String()
}

// GetResult returns the table result
func (m TableModel) GetResult() TableResult {
	return TableResult{
		Selected:  m.selected,
		Cancelled: m.quitting || m.selected == nil,
	}
}

// RunTable runs the interactive table and returns the result
func RunTable(client *api.Client) (TableResult, error) {
	model := NewTableModel(client)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return TableResult{Cancelled: true}, err
	}

	if m, ok := finalModel.(TableModel); ok {
		return m.GetResult(), nil
	}

	return TableResult{Cancelled: true}, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
