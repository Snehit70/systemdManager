package ui

import "github.com/charmbracelet/lipgloss"

var (
	ActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#73F59F"))

	FailedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	InactiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	DegradedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))
)

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
