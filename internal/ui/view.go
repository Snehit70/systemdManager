package ui

import (
	"fmt"
	"strings"

	"systemd-tui/internal/config"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func (m MainModel) View() string {
	var listStyle, detailStyle lipgloss.Style

	if m.activeView == listView {
		listStyle = m.activeBorder
		detailStyle = m.inactiveBorder
	} else {
		listStyle = m.inactiveBorder
		detailStyle = m.activeBorder
	}

	listView := listStyle.Render(m.list.View())
	detailView := detailStyle.Render(m.viewport.View())

	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listView,
		detailView,
	)

	fullWidth := m.width
	if fullWidth < 1 {
		fullWidth = lipgloss.Width(mainView)
	}

	statusBar := m.renderStatusBar(fullWidth)

	filterBar := ""
	if m.list.FilterState() == list.Filtering {
		filterStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			Width(fullWidth)
		filterBar = filterStyle.Render(m.list.FilterInput.View())
	}

	var helpView string
	if m.help.ShowAll {
		helpView = lipgloss.NewStyle().
			Width(fullWidth).
			PaddingLeft(1).
			Background(lipgloss.Color(m.theme.Surface)).
			Render(m.help.View(keys))
	} else {
		helpView = renderHelpBar(keys, fullWidth, m.theme)
	}

	baseView := lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		filterBar,
		statusBar,
		helpView,
	)

	if m.showCreate {
		return m.renderCreateModal(baseView)
	}

	return baseView
}

func (m MainModel) renderModalLabel(label string, index int) string {
	if m.createModal.focusIndex == index {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.BorderActive)).
			Bold(true).
			Render(label)
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Text)).
		Bold(true).
		Render(label)
}

func (m MainModel) renderCreateModal(baseView string) string {
	modalWidth := 60
	modalHeight := 18

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.TextMuted))

	var inputs []string

	inputs = append(inputs, fmt.Sprintf("%s %s", m.renderModalLabel("Name:", 0), m.createModal.nameInput.View()))
	inputs = append(inputs, fmt.Sprintf("%s %s", m.renderModalLabel("Command:", 1), m.createModal.execInput.View()))
	inputs = append(inputs, fmt.Sprintf("%s %s", m.renderModalLabel("Description:", 2), m.createModal.descInput.View()))
	inputs = append(inputs, fmt.Sprintf("%s %s", m.renderModalLabel("Workdir:", 3), m.createModal.workdirInput.View()))
	inputs = append(inputs, fmt.Sprintf("%s %s", m.renderModalLabel("Type:", 4), inputStyle.Render(m.createModal.serviceType+" (t to toggle)")))
	inputs = append(inputs, fmt.Sprintf("%s %s", m.renderModalLabel("Restart:", 5), inputStyle.Render(m.createModal.restart+" (r to cycle)")))

	footer := "Tab: next • Shift+Tab: prev • Enter: create • Esc: cancel"
	if m.creating {
		footer = "Creating service..."
	}

	content := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render("Create New Service"),
				"",
				strings.Join(inputs, "\n"),
				"",
				inputStyle.Render(footer),
			),
		)

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.BorderActive)).
		Background(lipgloss.Color(m.theme.Surface)).
		Render(content)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(lipgloss.Color(m.theme.Background)),
	)
}

func (m *MainModel) viewportWithScrollInfo(content string) string {
	totalLines := strings.Count(content, "\n") + 1
	viewportHeight := m.viewport.Height
	showScroll := totalLines > viewportHeight && viewportHeight > 0

	if !showScroll {
		return content
	}

	scrollPos := m.viewport.YOffset
	return fmt.Sprintf("[%d/%d]\n%s", scrollPos, totalLines, content)
}

func (m MainModel) renderStatusBar(width int) string {
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Border))
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.BorderActive)).
		Bold(true)
	valStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.TextMuted))

	sep := sepStyle.Render(" │ ")

	// Left segments: counts, filter, group, follow, detail mode
	var segments []string

	// Service counts
	filtered := m.filteredServices()
	if len(filtered) != len(m.services) {
		segments = append(segments, valStyle.Render(fmt.Sprintf("%d/%d services", len(filtered), len(m.services))))
	} else {
		segments = append(segments, valStyle.Render(fmt.Sprintf("%d services", len(m.services))))
	}

	// Filter mode
	if m.filterMode != filterAll {
		segments = append(segments, labelStyle.Render("F:")+valStyle.Render(m.filterMode.String()))
	}

	// Group mode
	if m.groupMode != groupNone {
		segments = append(segments, labelStyle.Render("G:")+valStyle.Render(m.groupMode.String()))
	}

	// Follow state
	if m.following {
		segments = append(segments, labelStyle.Render("⦿ FOLLOWING"))
	}

	// Detail view mode
	if m.detailViewMode != detailViewLogs {
		segments = append(segments, labelStyle.Render("View:")+valStyle.Render(m.detailViewMode.String()))
	}

	left := strings.Join(segments, sep)

	// Right segment: last action/status message
	right := ""
	if m.statusMessage != "" {
		right = valStyle.Render(m.statusMessage)
	}

	// Calculate available space
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := 2 // left padding

	bar := ""
	if width > 0 && leftWidth+rightWidth+len(sep)+padding < width {
		gap := width - leftWidth - rightWidth - padding
		if gap < 1 {
			gap = 1
		}
		bar = left + strings.Repeat(" ", gap) + right
	} else if width > 0 && leftWidth+padding < width {
		bar = left
	} else {
		bar = left
	}

	return lipgloss.NewStyle().
		PaddingLeft(1).
		Width(width).
		Background(lipgloss.Color(m.theme.Surface)).
		Render(bar)
}

func renderHelpBar(k keyMap, width int, theme config.ThemeColors) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.BorderActive)).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextMuted))
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Border))

	bindings := []key.Binding{
		k.Up, k.Start, k.Enable, k.ToggleFollow, k.Quit,
		k.Down, k.Stop, k.Disable, k.ToggleGroup, k.Help,
		k.Filter, k.Restart, k.Edit, k.SwitchFocus, k.Create, k.ToggleSource,
	}

	var parts []string
	for _, b := range bindings {
		help := b.Help()
		part := keyStyle.Render(help.Key) + " " + descStyle.Render(help.Desc)
		parts = append(parts, part)
	}

	sep := sepStyle.Render(" • ")
	line := strings.Join(parts, sep)

	return lipgloss.NewStyle().
		Width(width).
		PaddingLeft(1).
		Background(lipgloss.Color(theme.Surface)).
		Render(line)
}
