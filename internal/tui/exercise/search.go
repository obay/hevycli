package exercise

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/tui/common"
)

// ExerciseItem represents an exercise in the list
type ExerciseItem struct {
	exercise api.ExerciseTemplate
}

func (i ExerciseItem) Title() string       { return i.exercise.Title }
func (i ExerciseItem) Description() string { return i.exercise.PrimaryMuscleGroup + " • " + i.exercise.Equipment }
func (i ExerciseItem) FilterValue() string { return i.exercise.Title }

// SearchModel is the interactive exercise search model
type SearchModel struct {
	client     *api.Client
	textInput  textinput.Model
	list       list.Model
	exercises  []api.ExerciseTemplate
	filtered   []api.ExerciseTemplate
	loading    bool
	err        error
	selected   *api.ExerciseTemplate
	quitting   bool
	width      int
	height     int
}

// NewSearchModel creates a new exercise search model
func NewSearchModel(client *api.Client) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search exercises..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40
	ti.PromptStyle = common.FocusedStyle
	ti.TextStyle = common.NormalItemStyle

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = common.SelectedItemStyle
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(common.SecondaryColor)

	l := list.New([]list.Item{}, delegate, 60, 15)
	l.Title = "Exercise Templates"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = common.TitleStyle

	return SearchModel{
		client:    client,
		textInput: ti,
		list:      l,
		loading:   true,
	}
}

// Messages
type exercisesLoadedMsg struct {
	exercises []api.ExerciseTemplate
}

type errMsg struct {
	err error
}

// loadExercises fetches exercises from the API
func loadExercises(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		var allExercises []api.ExerciseTemplate
		page := 1
		for {
			resp, err := client.GetExerciseTemplates(page, 10)
			if err != nil {
				return errMsg{err: err}
			}
			allExercises = append(allExercises, resp.ExerciseTemplates...)
			if page >= resp.PageCount || resp.PageCount == 0 {
				break
			}
			page++
			// Limit to first 500 exercises for responsiveness
			if len(allExercises) > 500 {
				break
			}
		}
		return exercisesLoadedMsg{exercises: allExercises}
	}
}

// Init initializes the model
func (m SearchModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		loadExercises(m.client),
	)
}

// Update handles messages
func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if !m.loading && len(m.filtered) > 0 {
				selected := m.list.SelectedItem()
				if item, ok := selected.(ExerciseItem); ok {
					m.selected = &item.exercise
					m.quitting = true
					return m, tea.Quit
				}
			}

		case "up", "down":
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-10)

	case exercisesLoadedMsg:
		m.exercises = msg.exercises
		m.filtered = msg.exercises
		m.loading = false
		m.updateList()

	case errMsg:
		m.err = msg.err
		m.loading = false
	}

	// Update text input
	var tiCmd tea.Cmd
	prevValue := m.textInput.Value()
	m.textInput, tiCmd = m.textInput.Update(msg)
	cmds = append(cmds, tiCmd)

	// Filter if search changed
	if m.textInput.Value() != prevValue {
		m.filterExercises()
	}

	return m, tea.Batch(cmds...)
}

// filterExercises filters the exercise list based on search input
func (m *SearchModel) filterExercises() {
	query := strings.ToLower(m.textInput.Value())
	if query == "" {
		m.filtered = m.exercises
	} else {
		m.filtered = nil
		for _, ex := range m.exercises {
			if strings.Contains(strings.ToLower(ex.Title), query) ||
				strings.Contains(strings.ToLower(ex.PrimaryMuscleGroup), query) ||
				strings.Contains(strings.ToLower(ex.Equipment), query) {
				m.filtered = append(m.filtered, ex)
			}
		}
	}
	m.updateList()
}

// updateList updates the list items
func (m *SearchModel) updateList() {
	items := make([]list.Item, len(m.filtered))
	for i, ex := range m.filtered {
		items[i] = ExerciseItem{exercise: ex}
	}
	m.list.SetItems(items)
}

// View renders the model
func (m SearchModel) View() string {
	if m.quitting {
		if m.selected != nil {
			return common.SuccessStyle.Render(fmt.Sprintf("Selected: %s\n", m.selected.Title))
		}
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(common.TitleStyle.Render("Exercise Search"))
	b.WriteString("\n\n")

	// Search input
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
		// Results count
		b.WriteString(lipgloss.NewStyle().Foreground(common.MutedColor).Render(
			fmt.Sprintf("%d exercises found", len(m.filtered))))
		b.WriteString("\n\n")

		// List
		b.WriteString(m.list.View())
	}

	// Help
	b.WriteString("\n")
	b.WriteString(common.HelpStyle.Render("↑/↓ navigate • enter select • esc quit"))

	return b.String()
}

// Selected returns the selected exercise, if any
func (m SearchModel) Selected() *api.ExerciseTemplate {
	return m.selected
}

// Run starts the interactive exercise search
func Run(client *api.Client) (*api.ExerciseTemplate, error) {
	model := NewSearchModel(client)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := finalModel.(SearchModel); ok {
		return m.Selected(), nil
	}

	return nil, nil
}
