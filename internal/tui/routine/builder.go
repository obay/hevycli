package routine

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

// Mode represents the current interaction mode
type Mode int

const (
	ModeTitle Mode = iota
	ModeExerciseList
	ModeAddExercise
	ModeEditSets
	ModeConfirm
)

// RoutineExercise represents an exercise in the routine being built
type RoutineExercise struct {
	Template    api.ExerciseTemplate
	Sets        int
	RestSeconds int
	Notes       string
}

// exerciseListItem for the routine exercise list
type exerciseListItem struct {
	exercise RoutineExercise
	index    int
}

func (i exerciseListItem) Title() string {
	return fmt.Sprintf("%d. %s", i.index+1, i.exercise.Template.Title)
}

func (i exerciseListItem) Description() string {
	return fmt.Sprintf("%d sets • %ds rest", i.exercise.Sets, i.exercise.RestSeconds)
}

func (i exerciseListItem) FilterValue() string {
	return i.exercise.Template.Title
}

// templateListItem for the exercise template selection
type templateListItem struct {
	template api.ExerciseTemplate
}

func (i templateListItem) Title() string       { return i.template.Title }
func (i templateListItem) Description() string { return i.template.PrimaryMuscleGroup + " • " + i.template.Equipment }
func (i templateListItem) FilterValue() string { return i.template.Title }

// BuilderModel is the routine builder TUI model
type BuilderModel struct {
	client          *api.Client
	mode            Mode
	title           string
	titleInput      textinput.Model
	exercises       []RoutineExercise
	exerciseList    list.Model
	templateList    list.Model
	templateSearch  textinput.Model
	allTemplates    []api.ExerciseTemplate
	filteredTempls  []api.ExerciseTemplate
	setsInput       textinput.Model
	restInput       textinput.Model
	notesInput      textinput.Model
	editingIndex    int
	loading         bool
	err             error
	saved           bool
	createdRoutine  *api.Routine
	quitting        bool
	width           int
	height          int
}

// NewBuilderModel creates a new routine builder model
func NewBuilderModel(client *api.Client) BuilderModel {
	// Title input
	ti := textinput.New()
	ti.Placeholder = "Enter routine title..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40
	ti.PromptStyle = common.FocusedStyle
	ti.TextStyle = common.NormalItemStyle

	// Template search input
	ts := textinput.New()
	ts.Placeholder = "Search exercises..."
	ts.CharLimit = 50
	ts.Width = 40
	ts.PromptStyle = common.FocusedStyle

	// Sets input
	si := textinput.New()
	si.Placeholder = "3"
	si.CharLimit = 2
	si.Width = 10
	si.PromptStyle = common.FocusedStyle

	// Rest input
	ri := textinput.New()
	ri.Placeholder = "90"
	ri.CharLimit = 4
	ri.Width = 10
	ri.PromptStyle = common.FocusedStyle

	// Notes input
	ni := textinput.New()
	ni.Placeholder = "Optional notes..."
	ni.CharLimit = 200
	ni.Width = 40
	ni.PromptStyle = common.FocusedStyle

	// Exercise list for routine
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = common.SelectedItemStyle
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(common.SecondaryColor)

	el := list.New([]list.Item{}, delegate, 60, 10)
	el.Title = "Exercises"
	el.SetShowStatusBar(false)
	el.SetFilteringEnabled(false)
	el.Styles.Title = common.SubtitleStyle

	// Template list for adding exercises
	tl := list.New([]list.Item{}, delegate, 60, 12)
	tl.Title = "Select Exercise"
	tl.SetShowStatusBar(true)
	tl.SetFilteringEnabled(false)
	tl.Styles.Title = common.TitleStyle

	return BuilderModel{
		client:         client,
		mode:           ModeTitle,
		titleInput:     ti,
		exerciseList:   el,
		templateList:   tl,
		templateSearch: ts,
		setsInput:      si,
		restInput:      ri,
		notesInput:     ni,
		loading:        true,
	}
}

// Messages
type templatesLoadedMsg struct {
	templates []api.ExerciseTemplate
}

type routineCreatedMsg struct {
	routine *api.Routine
}

type errMsg struct {
	err error
}

// loadTemplates fetches exercise templates from the API
func loadTemplates(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		var allTemplates []api.ExerciseTemplate
		page := 1
		for {
			resp, err := client.GetExerciseTemplates(page, 10)
			if err != nil {
				return errMsg{err: err}
			}
			allTemplates = append(allTemplates, resp.ExerciseTemplates...)
			if page >= resp.PageCount || resp.PageCount == 0 {
				break
			}
			page++
			if len(allTemplates) > 500 {
				break
			}
		}
		return templatesLoadedMsg{templates: allTemplates}
	}
}

// createRoutine saves the routine to the API
func createRoutine(client *api.Client, title string, exercises []RoutineExercise) tea.Cmd {
	return func() tea.Msg {
		// Build the request
		var apiExercises []api.CreateRoutineExercise
		for _, ex := range exercises {
			sets := make([]api.CreateRoutineSet, ex.Sets)
			for j := 0; j < ex.Sets; j++ {
				sets[j] = api.CreateRoutineSet{
					Type: api.SetTypeNormal,
				}
			}
			restSeconds := ex.RestSeconds
			var notes *string
			if ex.Notes != "" {
				notes = &ex.Notes
			}
			apiExercises = append(apiExercises, api.CreateRoutineExercise{
				ExerciseTemplateID: ex.Template.ID,
				SupersetID:         nil,
				RestSeconds:        &restSeconds,
				Notes:              notes,
				Sets:               sets,
			})
		}

		req := &api.CreateRoutineRequest{
			Routine: api.CreateRoutineData{
				Title:     title,
				Exercises: apiExercises,
			},
		}

		routine, err := client.CreateRoutine(req)
		if err != nil {
			return errMsg{err: err}
		}

		return routineCreatedMsg{routine: routine}
	}
}

// Init initializes the model
func (m BuilderModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		loadTemplates(m.client),
	)
}

// Update handles messages
func (m BuilderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.exerciseList.SetSize(msg.Width-4, 10)
		m.templateList.SetSize(msg.Width-4, msg.Height-15)

	case templatesLoadedMsg:
		m.allTemplates = msg.templates
		m.filteredTempls = msg.templates
		m.loading = false
		m.updateTemplateList()

	case routineCreatedMsg:
		m.createdRoutine = msg.routine
		m.saved = true
		m.quitting = true
		return m, tea.Quit

	case errMsg:
		m.err = msg.err
		m.loading = false
	}

	// Update active input based on mode
	var cmd tea.Cmd
	switch m.mode {
	case ModeTitle:
		m.titleInput, cmd = m.titleInput.Update(msg)
		cmds = append(cmds, cmd)
	case ModeExerciseList:
		m.exerciseList, cmd = m.exerciseList.Update(msg)
		cmds = append(cmds, cmd)
	case ModeAddExercise:
		prevSearch := m.templateSearch.Value()
		m.templateSearch, cmd = m.templateSearch.Update(msg)
		cmds = append(cmds, cmd)
		if m.templateSearch.Value() != prevSearch {
			m.filterTemplates()
		}
		m.templateList, cmd = m.templateList.Update(msg)
		cmds = append(cmds, cmd)
	case ModeEditSets:
		m.setsInput, cmd = m.setsInput.Update(msg)
		cmds = append(cmds, cmd)
		m.restInput, cmd = m.restInput.Update(msg)
		cmds = append(cmds, cmd)
		m.notesInput, cmd = m.notesInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyPress processes key events based on current mode
func (m BuilderModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		switch m.mode {
		case ModeTitle:
			m.quitting = true
			return m, tea.Quit
		case ModeExerciseList:
			m.mode = ModeTitle
			m.titleInput.Focus()
		case ModeAddExercise:
			m.mode = ModeExerciseList
			m.templateSearch.SetValue("")
			m.filterTemplates()
		case ModeEditSets:
			m.mode = ModeAddExercise
			m.templateSearch.Focus()
		case ModeConfirm:
			m.mode = ModeExerciseList
		}
		return m, nil

	case "enter":
		return m.handleEnter()

	case "a":
		if m.mode == ModeExerciseList {
			m.mode = ModeAddExercise
			m.templateSearch.Focus()
			return m, nil
		}

	case "d":
		if m.mode == ModeExerciseList && len(m.exercises) > 0 {
			selected := m.exerciseList.Index()
			if selected >= 0 && selected < len(m.exercises) {
				m.exercises = append(m.exercises[:selected], m.exercises[selected+1:]...)
				m.updateExerciseList()
			}
			return m, nil
		}

	case "s":
		if m.mode == ModeExerciseList && len(m.exercises) > 0 && m.title != "" {
			m.mode = ModeConfirm
			return m, nil
		}

	case "up", "down":
		if m.mode == ModeAddExercise {
			var cmd tea.Cmd
			m.templateList, cmd = m.templateList.Update(msg)
			return m, cmd
		}

	case "tab":
		if m.mode == ModeEditSets {
			if m.setsInput.Focused() {
				m.setsInput.Blur()
				m.restInput.Focus()
			} else if m.restInput.Focused() {
				m.restInput.Blur()
				m.notesInput.Focus()
			} else {
				m.notesInput.Blur()
				m.setsInput.Focus()
			}
			return m, nil
		}
	}

	return m, nil
}

// handleEnter processes enter key based on current mode
func (m BuilderModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeTitle:
		if m.titleInput.Value() != "" {
			m.title = m.titleInput.Value()
			m.mode = ModeExerciseList
			m.titleInput.Blur()
		}

	case ModeExerciseList:
		// Edit selected exercise sets
		if len(m.exercises) > 0 {
			selected := m.exerciseList.Index()
			if selected >= 0 && selected < len(m.exercises) {
				m.editingIndex = selected
				ex := m.exercises[selected]
				m.setsInput.SetValue(fmt.Sprintf("%d", ex.Sets))
				m.restInput.SetValue(fmt.Sprintf("%d", ex.RestSeconds))
				m.notesInput.SetValue(ex.Notes)
				m.mode = ModeEditSets
				m.setsInput.Focus()
			}
		}

	case ModeAddExercise:
		// Select exercise template
		if len(m.filteredTempls) > 0 {
			selected := m.templateList.SelectedItem()
			if item, ok := selected.(templateListItem); ok {
				// Add exercise with defaults
				m.exercises = append(m.exercises, RoutineExercise{
					Template:    item.template,
					Sets:        3,
					RestSeconds: 90,
					Notes:       "",
				})
				m.updateExerciseList()
				m.mode = ModeExerciseList
				m.templateSearch.SetValue("")
				m.filterTemplates()
			}
		}

	case ModeEditSets:
		// Save exercise settings
		if m.editingIndex >= 0 && m.editingIndex < len(m.exercises) {
			sets := 3
			rest := 90
			fmt.Sscanf(m.setsInput.Value(), "%d", &sets)
			fmt.Sscanf(m.restInput.Value(), "%d", &rest)
			if sets < 1 {
				sets = 1
			}
			if sets > 20 {
				sets = 20
			}
			if rest < 0 {
				rest = 0
			}
			m.exercises[m.editingIndex].Sets = sets
			m.exercises[m.editingIndex].RestSeconds = rest
			m.exercises[m.editingIndex].Notes = m.notesInput.Value()
			m.updateExerciseList()
		}
		m.mode = ModeExerciseList
		m.setsInput.Blur()
		m.restInput.Blur()
		m.notesInput.Blur()

	case ModeConfirm:
		// Save routine
		m.loading = true
		return m, createRoutine(m.client, m.title, m.exercises)
	}

	return m, nil
}

// filterTemplates filters templates based on search input
func (m *BuilderModel) filterTemplates() {
	query := strings.ToLower(m.templateSearch.Value())
	if query == "" {
		m.filteredTempls = m.allTemplates
	} else {
		m.filteredTempls = nil
		for _, t := range m.allTemplates {
			if strings.Contains(strings.ToLower(t.Title), query) ||
				strings.Contains(strings.ToLower(t.PrimaryMuscleGroup), query) {
				m.filteredTempls = append(m.filteredTempls, t)
			}
		}
	}
	m.updateTemplateList()
}

// updateTemplateList updates the template list items
func (m *BuilderModel) updateTemplateList() {
	items := make([]list.Item, len(m.filteredTempls))
	for i, t := range m.filteredTempls {
		items[i] = templateListItem{template: t}
	}
	m.templateList.SetItems(items)
}

// updateExerciseList updates the exercise list items
func (m *BuilderModel) updateExerciseList() {
	items := make([]list.Item, len(m.exercises))
	for i, ex := range m.exercises {
		items[i] = exerciseListItem{exercise: ex, index: i}
	}
	m.exerciseList.SetItems(items)
}

// View renders the model
func (m BuilderModel) View() string {
	if m.quitting {
		if m.saved && m.createdRoutine != nil {
			return common.SuccessStyle.Render(fmt.Sprintf("✓ Routine '%s' created successfully!\n  ID: %s\n",
				m.createdRoutine.Title, m.createdRoutine.ID))
		}
		return ""
	}

	var b strings.Builder

	// Header
	b.WriteString(common.TitleStyle.Render("Create Routine"))
	b.WriteString("\n\n")

	// Error display
	if m.err != nil {
		b.WriteString(common.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	// Loading state
	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(common.MutedColor).Render("Loading..."))
		b.WriteString("\n")
		return b.String()
	}

	switch m.mode {
	case ModeTitle:
		b.WriteString(m.renderTitleMode())
	case ModeExerciseList:
		b.WriteString(m.renderExerciseListMode())
	case ModeAddExercise:
		b.WriteString(m.renderAddExerciseMode())
	case ModeEditSets:
		b.WriteString(m.renderEditSetsMode())
	case ModeConfirm:
		b.WriteString(m.renderConfirmMode())
	}

	return b.String()
}

func (m BuilderModel) renderTitleMode() string {
	var b strings.Builder
	b.WriteString("Title: ")
	b.WriteString(m.titleInput.View())
	b.WriteString("\n\n")
	b.WriteString(common.HelpStyle.Render("enter continue • esc cancel"))
	return b.String()
}

func (m BuilderModel) renderExerciseListMode() string {
	var b strings.Builder

	// Title display
	b.WriteString(lipgloss.NewStyle().Foreground(common.SecondaryColor).Render("Title: "))
	b.WriteString(m.title)
	b.WriteString("\n\n")

	// Exercise list
	if len(m.exercises) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(common.MutedColor).Render("No exercises added yet. Press 'a' to add exercises."))
		b.WriteString("\n")
	} else {
		b.WriteString(m.exerciseList.View())
	}

	b.WriteString("\n")

	// Help
	help := "[a] add exercise"
	if len(m.exercises) > 0 {
		help += " • [enter] edit • [d] delete • [s] save routine"
	}
	help += " • [esc] back"
	b.WriteString(common.HelpStyle.Render(help))

	return b.String()
}

func (m BuilderModel) renderAddExerciseMode() string {
	var b strings.Builder

	b.WriteString("Search: ")
	b.WriteString(m.templateSearch.View())
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(common.MutedColor).Render(
		fmt.Sprintf("%d exercises found", len(m.filteredTempls))))
	b.WriteString("\n\n")

	b.WriteString(m.templateList.View())
	b.WriteString("\n")

	b.WriteString(common.HelpStyle.Render("↑/↓ navigate • enter select • esc cancel"))

	return b.String()
}

func (m BuilderModel) renderEditSetsMode() string {
	var b strings.Builder

	if m.editingIndex >= 0 && m.editingIndex < len(m.exercises) {
		ex := m.exercises[m.editingIndex]
		b.WriteString(common.SubtitleStyle.Render(ex.Template.Title))
		b.WriteString("\n\n")
	}

	b.WriteString("Sets: ")
	b.WriteString(m.setsInput.View())
	b.WriteString("\n\n")

	b.WriteString("Rest (seconds): ")
	b.WriteString(m.restInput.View())
	b.WriteString("\n\n")

	b.WriteString("Notes: ")
	b.WriteString(m.notesInput.View())
	b.WriteString("\n\n")

	b.WriteString(common.HelpStyle.Render("tab next field • enter save • esc cancel"))

	return b.String()
}

func (m BuilderModel) renderConfirmMode() string {
	var b strings.Builder

	b.WriteString(common.SubtitleStyle.Render("Save Routine?"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Title: %s\n", m.title))
	b.WriteString(fmt.Sprintf("Exercises: %d\n\n", len(m.exercises)))

	for i, ex := range m.exercises {
		b.WriteString(fmt.Sprintf("  %d. %s (%d sets)\n", i+1, ex.Template.Title, ex.Sets))
	}

	b.WriteString("\n")
	b.WriteString(common.HelpStyle.Render("enter confirm • esc cancel"))

	return b.String()
}

// CreatedRoutine returns the created routine, if any
func (m BuilderModel) CreatedRoutine() *api.Routine {
	return m.createdRoutine
}

// Run starts the interactive routine builder
func Run(client *api.Client) (*api.Routine, error) {
	model := NewBuilderModel(client)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := finalModel.(BuilderModel); ok {
		return m.CreatedRoutine(), nil
	}

	return nil, nil
}
