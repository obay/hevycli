package workout

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/obay/hevycli/internal/api"
	"github.com/obay/hevycli/internal/tui/common"
)

// SetData represents a single set being logged
type SetData struct {
	Weight   float64
	Reps     int
	SetType  api.SetType
	Complete bool
}

// ExerciseData represents an exercise in the workout
type ExerciseData struct {
	Template api.ExerciseTemplate
	Sets     []SetData
	Notes    string
	Done     bool
}

// SessionModel is the interactive workout session model
type SessionModel struct {
	title           string
	startTime       time.Time
	exercises       []ExerciseData
	currentExercise int
	currentSet      int
	currentField    int // 0=weight, 1=reps
	weightInput     textinput.Model
	repsInput       textinput.Model
	setsTable       table.Model
	quitting        bool
	finished        bool
	width           int
	height          int
}

// NewSessionModel creates a new workout session
func NewSessionModel(title string, exercises []ExerciseData) SessionModel {
	wi := textinput.New()
	wi.Placeholder = "0"
	wi.Focus()
	wi.CharLimit = 6
	wi.Width = 8
	wi.PromptStyle = common.FocusedStyle

	ri := textinput.New()
	ri.Placeholder = "0"
	ri.CharLimit = 4
	ri.Width = 6

	// Create the sets table
	columns := []table.Column{
		{Title: "#", Width: 4},
		{Title: "Type", Width: 8},
		{Title: "Weight", Width: 10},
		{Title: "Reps", Width: 6},
		{Title: "", Width: 3},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(8),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(common.MutedColor).
		BorderBottom(true).
		Bold(true).
		Foreground(common.PrimaryColor)
	s.Selected = s.Selected.
		Foreground(common.PrimaryColor).
		Bold(true)
	t.SetStyles(s)

	m := SessionModel{
		title:       title,
		startTime:   time.Now(),
		exercises:   exercises,
		weightInput: wi,
		repsInput:   ri,
		setsTable:   t,
	}

	// Initialize table with first exercise's sets
	m.updateSetsTable()
	return m
}

// NewSessionFromRoutine creates a session from a routine
func NewSessionFromRoutine(routine *api.Routine) SessionModel {
	exercises := make([]ExerciseData, len(routine.Exercises))
	for i, ex := range routine.Exercises {
		sets := make([]SetData, len(ex.Sets))
		for j, s := range ex.Sets {
			sets[j] = SetData{
				SetType: s.SetType,
			}
		}
		exercises[i] = ExerciseData{
			Template: api.ExerciseTemplate{
				ID:    ex.ExerciseTemplateID,
				Title: ex.Title,
			},
			Sets:  sets,
			Notes: ex.Notes,
		}
	}
	return NewSessionModel(routine.Title, exercises)
}

// tickMsg for timer updates
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init initializes the model
func (m SessionModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tickCmd())
}

// Update handles messages
func (m SessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			// Confirm quit
			m.quitting = true
			return m, tea.Quit

		case "tab":
			// Switch between weight and reps input
			if m.currentField == 0 {
				m.currentField = 1
				m.weightInput.Blur()
				m.repsInput.Focus()
			} else {
				m.currentField = 0
				m.repsInput.Blur()
				m.weightInput.Focus()
			}

		case "enter":
			// Save current set and move to next
			m.saveCurrentSet()
			m.moveToNextSet()

		case "up":
			if m.currentSet > 0 {
				m.currentSet--
				m.loadCurrentSet()
			}

		case "down":
			if m.currentSet < len(m.exercises[m.currentExercise].Sets)-1 {
				m.currentSet++
				m.loadCurrentSet()
			}

		case "left", "shift+tab":
			// Previous exercise
			if m.currentExercise > 0 {
				m.currentExercise--
				m.currentSet = 0
				m.loadCurrentSet()
			}

		case "right":
			// Next exercise
			if m.currentExercise < len(m.exercises)-1 {
				m.currentExercise++
				m.currentSet = 0
				m.loadCurrentSet()
			}

		case "n":
			// Add new set
			m.addSet()

		case "f":
			// Finish workout
			m.finished = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		return m, tickCmd()
	}

	// Update text inputs
	if m.currentField == 0 {
		var cmd tea.Cmd
		m.weightInput, cmd = m.weightInput.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		m.repsInput, cmd = m.repsInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// saveCurrentSet saves the current input to the set data
func (m *SessionModel) saveCurrentSet() {
	if m.currentExercise >= len(m.exercises) {
		return
	}
	ex := &m.exercises[m.currentExercise]
	if m.currentSet >= len(ex.Sets) {
		return
	}

	var weight float64
	fmt.Sscanf(m.weightInput.Value(), "%f", &weight)
	var reps int
	fmt.Sscanf(m.repsInput.Value(), "%d", &reps)

	ex.Sets[m.currentSet].Weight = weight
	ex.Sets[m.currentSet].Reps = reps
	ex.Sets[m.currentSet].Complete = true
}

// loadCurrentSet loads the set data into inputs
func (m *SessionModel) loadCurrentSet() {
	if m.currentExercise >= len(m.exercises) {
		return
	}
	ex := m.exercises[m.currentExercise]
	if m.currentSet >= len(ex.Sets) {
		return
	}

	set := ex.Sets[m.currentSet]
	if set.Complete {
		m.weightInput.SetValue(fmt.Sprintf("%.1f", set.Weight))
		m.repsInput.SetValue(fmt.Sprintf("%d", set.Reps))
	} else {
		m.weightInput.SetValue("")
		m.repsInput.SetValue("")
	}

	m.currentField = 0
	m.repsInput.Blur()
	m.weightInput.Focus()

	// Update table
	m.updateSetsTable()
}

// updateSetsTable refreshes the table rows for the current exercise
func (m *SessionModel) updateSetsTable() {
	if m.currentExercise >= len(m.exercises) {
		return
	}
	ex := m.exercises[m.currentExercise]

	rows := make([]table.Row, len(ex.Sets))
	for i, set := range ex.Sets {
		var weight, reps, status string

		if set.Complete {
			weight = fmt.Sprintf("%.1f kg", set.Weight)
			reps = fmt.Sprintf("%d", set.Reps)
			status = "✓"
		} else {
			weight = "-"
			reps = "-"
			status = ""
		}

		rows[i] = table.Row{
			fmt.Sprintf("%d", i+1),
			string(set.SetType),
			weight,
			reps,
			status,
		}
	}

	m.setsTable.SetRows(rows)
	m.setsTable.SetCursor(m.currentSet)
}

// moveToNextSet moves to the next incomplete set or exercise
func (m *SessionModel) moveToNextSet() {
	ex := &m.exercises[m.currentExercise]

	// Try next set in current exercise
	if m.currentSet < len(ex.Sets)-1 {
		m.currentSet++
		m.loadCurrentSet()
		return
	}

	// Mark exercise as done if all sets complete
	allComplete := true
	for _, s := range ex.Sets {
		if !s.Complete {
			allComplete = false
			break
		}
	}
	if allComplete {
		ex.Done = true
	}

	// Try next exercise
	if m.currentExercise < len(m.exercises)-1 {
		m.currentExercise++
		m.currentSet = 0
		m.loadCurrentSet()
	}
}

// addSet adds a new set to the current exercise
func (m *SessionModel) addSet() {
	ex := &m.exercises[m.currentExercise]
	ex.Sets = append(ex.Sets, SetData{SetType: api.SetTypeNormal})
	m.currentSet = len(ex.Sets) - 1
	m.loadCurrentSet()
}

// View renders the model
func (m SessionModel) View() string {
	if m.quitting && !m.finished {
		return common.WarningStyle.Render("Workout cancelled.\n")
	}
	if m.finished {
		return m.renderSummary()
	}

	var b strings.Builder

	// Header
	duration := time.Since(m.startTime)
	header := fmt.Sprintf(" %s    Duration: %s    %d/%d exercises ",
		m.title,
		formatDuration(duration),
		m.countCompleteExercises(),
		len(m.exercises))

	b.WriteString(common.HeaderStyle.Width(m.width - 2).Render(header))
	b.WriteString("\n\n")

	// Exercise list (left panel)
	exerciseList := m.renderExerciseList()

	// Current exercise detail (right panel)
	exerciseDetail := m.renderExerciseDetail()

	// Layout
	leftPanel := lipgloss.NewStyle().Width(25).Render(exerciseList)
	rightPanel := lipgloss.NewStyle().Width(m.width - 30).Render(exerciseDetail)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel))

	// Help bar
	b.WriteString("\n\n")
	b.WriteString(common.HelpStyle.Render(
		"tab switch field • enter save set • ↑↓ sets • ←→ exercises • n new set • f finish • esc quit"))

	return b.String()
}

// renderExerciseList renders the exercise sidebar
func (m SessionModel) renderExerciseList() string {
	var b strings.Builder
	b.WriteString(common.SubtitleStyle.Render("Exercises"))
	b.WriteString("\n")

	for i, ex := range m.exercises {
		prefix := "  "
		style := common.NormalItemStyle

		if i == m.currentExercise {
			prefix = "→ "
			style = common.SelectedItemStyle
		}

		status := "○"
		if ex.Done {
			status = common.SuccessStyle.Render("✓")
		} else if i == m.currentExercise {
			status = "●"
		}

		name := ex.Template.Title
		if len(name) > 18 {
			name = name[:15] + "..."
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", prefix, status, style.Render(name)))
	}

	return b.String()
}

// renderExerciseDetail renders the current exercise detail
func (m SessionModel) renderExerciseDetail() string {
	if m.currentExercise >= len(m.exercises) {
		return ""
	}

	ex := m.exercises[m.currentExercise]
	var b strings.Builder

	// Exercise title
	b.WriteString(common.TitleStyle.Render(ex.Template.Title))
	b.WriteString("\n")

	if ex.Notes != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(common.MutedColor).Render(ex.Notes))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Sets table using Bubbles table component
	b.WriteString(m.setsTable.View())
	b.WriteString("\n\n")

	// Input fields for the current set
	b.WriteString("  Weight: ")
	b.WriteString(m.weightInput.View())
	b.WriteString("  Reps: ")
	b.WriteString(m.repsInput.View())

	return b.String()
}

// renderSummary renders the workout summary
func (m SessionModel) renderSummary() string {
	var b strings.Builder

	duration := time.Since(m.startTime)

	b.WriteString(common.TitleStyle.Render("Workout Complete!"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Title: %s\n", m.title))
	b.WriteString(fmt.Sprintf("Duration: %s\n", formatDuration(duration)))
	b.WriteString(fmt.Sprintf("Exercises: %d\n", len(m.exercises)))
	b.WriteString(fmt.Sprintf("Total Sets: %d\n", m.countTotalSets()))
	b.WriteString("\n")

	// Exercise summary
	for _, ex := range m.exercises {
		status := common.SuccessStyle.Render("✓")
		if !ex.Done {
			status = common.WarningStyle.Render("○")
		}
		completeSets := 0
		for _, s := range ex.Sets {
			if s.Complete {
				completeSets++
			}
		}
		b.WriteString(fmt.Sprintf("%s %s (%d/%d sets)\n",
			status, ex.Template.Title, completeSets, len(ex.Sets)))
	}

	return b.String()
}

// countCompleteExercises counts done exercises
func (m SessionModel) countCompleteExercises() int {
	count := 0
	for _, ex := range m.exercises {
		if ex.Done {
			count++
		}
	}
	return count
}

// countTotalSets counts all sets
func (m SessionModel) countTotalSets() int {
	count := 0
	for _, ex := range m.exercises {
		count += len(ex.Sets)
	}
	return count
}

// formatDuration formats a duration as HH:MM:SS
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// GetExercises returns the exercise data
func (m SessionModel) GetExercises() []ExerciseData {
	return m.exercises
}

// IsFinished returns true if workout was completed
func (m SessionModel) IsFinished() bool {
	return m.finished
}

// Run starts the interactive workout session
func RunSession(title string, exercises []ExerciseData) ([]ExerciseData, bool, error) {
	model := NewSessionModel(title, exercises)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, false, err
	}

	if m, ok := finalModel.(SessionModel); ok {
		return m.GetExercises(), m.IsFinished(), nil
	}

	return nil, false, nil
}

// RunSessionFromRoutine starts a session from a routine
func RunSessionFromRoutine(routine *api.Routine) ([]ExerciseData, bool, error) {
	model := NewSessionFromRoutine(routine)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, false, err
	}

	if m, ok := finalModel.(SessionModel); ok {
		return m.GetExercises(), m.IsFinished(), nil
	}

	return nil, false, nil
}
