package prompt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/obay/hevycli/internal/tui/common"
)

// TextInput prompts the user for text input with a beautiful TUI
func TextInput(title, placeholder, help string) (string, error) {
	model := newTextInputModel(title, placeholder, help)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := finalModel.(textInputModel); ok {
		if m.cancelled {
			return "", fmt.Errorf("cancelled")
		}
		return m.value, nil
	}

	return "", fmt.Errorf("unexpected model type")
}

// textInputModel is the Bubble Tea model for text input
type textInputModel struct {
	title       string
	help        string
	textInput   textinput.Model
	value       string
	cancelled   bool
	width       int
	height      int
}

func newTextInputModel(title, placeholder, help string) textInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.PromptStyle = common.FocusedStyle
	ti.TextStyle = common.NormalItemStyle
	ti.Cursor.Style = common.CursorStyle

	return textInputModel{
		title:     title,
		help:      help,
		textInput: ti,
	}
}

func (m textInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			m.value = m.textInput.Value()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = min(msg.Width-10, 60)
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m textInputModel) View() string {
	var b strings.Builder

	// Title box
	titleBox := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.PrimaryColor).
		MarginBottom(1).
		Render(m.title)

	b.WriteString(titleBox)
	b.WriteString("\n\n")

	// Input field
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Help text
	helpText := common.HelpStyle.Render(m.help + " • esc cancel")
	b.WriteString(helpText)

	// Wrap in a box
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.PrimaryColor).
		Padding(1, 2).
		Render(b.String())
}

// SelectOption represents an option in a selection list
type SelectOption struct {
	ID          string
	Title       string
	Description string
}

func (o SelectOption) FilterValue() string { return o.Title }

// selectItem wraps SelectOption for the list component
type selectItem struct {
	option SelectOption
}

func (i selectItem) Title() string       { return i.option.Title }
func (i selectItem) Description() string { return i.option.Description }
func (i selectItem) FilterValue() string { return i.option.Title }

// Select prompts the user to select from a list of options
func Select(title string, options []SelectOption, help string) (*SelectOption, error) {
	if len(options) == 0 {
		return nil, fmt.Errorf("no options provided")
	}

	model := newSelectModel(title, options, help)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := finalModel.(selectModel); ok {
		if m.cancelled {
			return nil, fmt.Errorf("cancelled")
		}
		return m.selected, nil
	}

	return nil, fmt.Errorf("unexpected model type")
}

// selectModel is the Bubble Tea model for selection
type selectModel struct {
	title     string
	help      string
	list      list.Model
	options   []SelectOption
	selected  *SelectOption
	cancelled bool
	width     int
	height    int
}

func newSelectModel(title string, options []SelectOption, help string) selectModel {
	items := make([]list.Item, len(options))
	for i, opt := range options {
		items[i] = selectItem{option: opt}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = common.SelectedItemStyle
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(common.SecondaryColor)
	delegate.Styles.NormalTitle = common.NormalItemStyle
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Foreground(common.MutedColor)

	l := list.New(items, delegate, 60, min(len(options)*3+4, 20))
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = common.TitleStyle
	l.Styles.FilterPrompt = common.FocusedStyle
	l.Styles.FilterCursor = common.CursorStyle

	return selectModel{
		title:   title,
		help:    help,
		list:    l,
		options: options,
	}
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't handle keys if filtering
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(selectItem); ok {
				m.selected = &item.option
			}
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-6)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectModel) View() string {
	var b strings.Builder

	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(common.HelpStyle.Render(m.help + " • / filter • esc cancel"))

	return b.String()
}

// SearchSelect prompts the user to search and select from a dynamically loaded list
type SearchSelectConfig struct {
	Title       string
	Placeholder string
	Help        string
	LoadFunc    func() ([]SelectOption, error)
}

// SearchSelect prompts with search functionality and async loading
func SearchSelect(config SearchSelectConfig) (*SelectOption, error) {
	model := newSearchSelectModel(config)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := finalModel.(searchSelectModel); ok {
		if m.cancelled {
			return nil, fmt.Errorf("cancelled")
		}
		return m.selected, nil
	}

	return nil, fmt.Errorf("unexpected model type")
}

// searchSelectModel is the model for search and select
type searchSelectModel struct {
	config      SearchSelectConfig
	textInput   textinput.Model
	list        list.Model
	allOptions  []SelectOption
	filtered    []SelectOption
	loading     bool
	err         error
	selected    *SelectOption
	cancelled   bool
	width       int
	height      int
}

func newSearchSelectModel(config SearchSelectConfig) searchSelectModel {
	ti := textinput.New()
	ti.Placeholder = config.Placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50
	ti.PromptStyle = common.FocusedStyle
	ti.TextStyle = common.NormalItemStyle

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = common.SelectedItemStyle
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(common.SecondaryColor)

	l := list.New([]list.Item{}, delegate, 60, 15)
	l.Title = ""
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)

	return searchSelectModel{
		config:    config,
		textInput: ti,
		list:      l,
		loading:   true,
	}
}

type optionsLoadedMsg struct {
	options []SelectOption
}

type loadErrMsg struct {
	err error
}

func loadOptions(loadFunc func() ([]SelectOption, error)) tea.Cmd {
	return func() tea.Msg {
		options, err := loadFunc()
		if err != nil {
			return loadErrMsg{err: err}
		}
		return optionsLoadedMsg{options: options}
	}
}

func (m searchSelectModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		loadOptions(m.config.LoadFunc),
	)
}

func (m searchSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if !m.loading && len(m.filtered) > 0 {
				if item, ok := m.list.SelectedItem().(selectItem); ok {
					m.selected = &item.option
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
		m.list.SetSize(msg.Width-4, msg.Height-12)

	case optionsLoadedMsg:
		m.allOptions = msg.options
		m.filtered = msg.options
		m.loading = false
		m.updateList()

	case loadErrMsg:
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
		m.filterOptions()
	}

	return m, tea.Batch(cmds...)
}

func (m *searchSelectModel) filterOptions() {
	query := strings.ToLower(m.textInput.Value())
	if query == "" {
		m.filtered = m.allOptions
	} else {
		m.filtered = nil
		for _, opt := range m.allOptions {
			if strings.Contains(strings.ToLower(opt.Title), query) ||
				strings.Contains(strings.ToLower(opt.Description), query) {
				m.filtered = append(m.filtered, opt)
			}
		}
	}
	m.updateList()
}

func (m *searchSelectModel) updateList() {
	items := make([]list.Item, len(m.filtered))
	for i, opt := range m.filtered {
		items[i] = selectItem{option: opt}
	}
	m.list.SetItems(items)
}

func (m searchSelectModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(common.TitleStyle.Render(m.config.Title))
	b.WriteString("\n\n")

	// Search input
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Loading or error state
	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(common.MutedColor).Render("Loading..."))
		b.WriteString("\n")
	} else if m.err != nil {
		b.WriteString(common.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	} else {
		// Results count
		b.WriteString(lipgloss.NewStyle().Foreground(common.MutedColor).Render(
			fmt.Sprintf("%d item(s) found", len(m.filtered))))
		b.WriteString("\n\n")

		// List
		b.WriteString(m.list.View())
	}

	// Help
	b.WriteString("\n")
	b.WriteString(common.HelpStyle.Render(m.config.Help + " • ↑/↓ navigate • enter select • esc cancel"))

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
