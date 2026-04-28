package ui

import (
	"systemd-tui/internal/config"

	"github.com/charmbracelet/lipgloss"
)

var (
	ActiveStyle          lipgloss.Style
	FailedStyle          lipgloss.Style
	InactiveStyle        lipgloss.Style
	DegradedStyle        lipgloss.Style
	SelectedStyle        lipgloss.Style
	DescriptionStyle     lipgloss.Style
	SourceUserStyle      lipgloss.Style
	SourceSystemStyle    lipgloss.Style
	SourceTransientStyle lipgloss.Style
	SourceGeneratedStyle lipgloss.Style
	SourceStaticStyle    lipgloss.Style
	SourceOtherStyle     lipgloss.Style
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

	SourceUserStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SourceUser))
	SourceSystemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SourceSystem))
	SourceTransientStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SourceTransient))
	SourceGeneratedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SourceGenerated))
	SourceStaticStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SourceStatic))
	SourceOtherStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SourceStatic))
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

func GetSourceStyle(source string) lipgloss.Style {
	switch source {
	case "user":
		return SourceUserStyle
	case "system":
		return SourceSystemStyle
	default:
		return SourceOtherStyle
	}
}
