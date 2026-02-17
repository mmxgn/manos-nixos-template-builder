package tui

import "github.com/charmbracelet/lipgloss"

var (
	// TitleStyle for screen titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1).
			MarginTop(1)

	// SubtitleStyle for subtitles and descriptions
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginBottom(1)

	// SelectedItemStyle for selected list items
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	// CheckboxStyle for checked items
	CheckboxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	// UncheckedStyle for unchecked items
	UncheckedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	// HelpStyle for help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	// ErrorStyle for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	// SuccessStyle for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	// DisabledStyle for greyed-out / unsupported items
	DisabledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444"))

	// InfoStyle for informational notes
	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5A623")).
			Bold(true)

	// BorderStyle for boxes
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)
)
