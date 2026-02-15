package ui

import (
	"systemd-tui/internal/config"

	"github.com/charmbracelet/lipgloss"
)

var (
	ActiveStyle      lipgloss.Style
	FailedStyle      lipgloss.Style
	InactiveStyle    lipgloss.Style
	DegradedStyle    lipgloss.Style
	SelectedStyle    lipgloss.Style
	DescriptionStyle lipgloss.Style
)

func ApplyTheme(theme config.ThemeColors) {
	ActiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.StatusActive))

	FailedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.StatusFailed))

	InactiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.StatusInactive))

	DegradedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500"))

	SelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.BorderActive)).
		Bold(true)

	DescriptionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextMuted))
}

func init() {
	ApplyTheme(config.DarkTheme)
}

func GetStatusStyle(activeState string) lipgloss.Style {
	switch activeState {
	case "active":
		return ActiveStyle
	case "failed":
		return FailedStyle
	case "activating", "deactivating", "reloading":
		return DegradedStyle
	default:
		return InactiveStyle
	}
}
