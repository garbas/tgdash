package tui

import "github.com/charmbracelet/lipgloss"

var (
	StatusWaiting = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
	StatusRunning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true)
	StatusDone = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34"))
	StatusError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("63")).
			Padding(0, 1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57"))

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	AddStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34"))
	ChangeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))
	DestroyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	EstimateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
)
