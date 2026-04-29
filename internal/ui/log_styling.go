package ui

import (
	"regexp"
	"strings"

	"systemd-tui/internal/config"

	"github.com/charmbracelet/lipgloss"
)

var (
	errorRegex = regexp.MustCompile(`(?i)\b(error|err|fatal|panic|critical)\b`)
	warnRegex  = regexp.MustCompile(`(?i)\b(warn|warning)\b`)
	infoRegex  = regexp.MustCompile(`(?i)\b(info)\b`)
	debugRegex = regexp.MustCompile(`(?i)\b(debug|trace)\b`)
)

func styleLogContent(content string, theme config.ThemeColors) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = styleLogLine(line, theme)
	}
	return strings.Join(lines, "\n")
}

func styleLogLine(line string, theme config.ThemeColors) string {
	if errorRegex.MatchString(line) {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.LogError)).
			Render(line)
	}
	if warnRegex.MatchString(line) {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.LogWarn)).
			Render(line)
	}
	if debugRegex.MatchString(line) {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.LogDebug)).
			Render(line)
	}
	if infoRegex.MatchString(line) {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.LogInfo)).
			Render(line)
	}
	return line
}
