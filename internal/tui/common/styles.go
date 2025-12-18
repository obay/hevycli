package common

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	PrimaryColor   = lipgloss.Color("212")
	SecondaryColor = lipgloss.Color("86")
	AccentColor    = lipgloss.Color("205")
	SuccessColor   = lipgloss.Color("82")
	WarningColor   = lipgloss.Color("214")
	ErrorColor     = lipgloss.Color("196")
	MutedColor     = lipgloss.Color("240")
	TextColor      = lipgloss.Color("252")
)

// Styles
var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			MarginBottom(1)

	// Input styles
	FocusedStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	BlurredStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	CursorStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor)

	// List item styles
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(TextColor)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginTop(1)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MutedColor).
			Padding(1, 2)

	FocusedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			Width(60)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
)

// RenderPrompt creates a styled prompt string
func RenderPrompt(prompt string, focused bool) string {
	if focused {
		return FocusedStyle.Render(prompt)
	}
	return BlurredStyle.Render(prompt)
}

// RenderHelp creates a help text
func RenderHelp(keys ...string) string {
	var items []string
	for i := 0; i < len(keys)-1; i += 2 {
		key := keys[i]
		desc := keys[i+1]
		items = append(items, lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Render(key)+" "+
			lipgloss.NewStyle().
				Foreground(MutedColor).
				Render(desc))
	}

	return HelpStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, items...))
}
