package ui

import (
	"fmt"
	"io"
	"strings"

	"systemd-tui/internal/client"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// itemDelegate implements list.DefaultDelegate for custom service item rendering.
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 1 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	if header, ok := listItem.(groupHeaderItem); ok {
		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Bold(true)
		fmt.Fprint(w, headerStyle.Render(runewidth.Truncate(header.title, m.Width(), "…")))
		return
	}

	i, ok := listItem.(item)
	if !ok {
		return
	}

	statusStyle := GetStatusStyle(string(i.svc.Status))
	sourceIndicator := sourceIndicator(i.svc.Source)
	sourceStyling := sourceStyle(i.svc.Source)
	styledIndicator := sourceStyling.Render(sourceIndicator)

	rowWidth := m.Width()
	if rowWidth <= 0 {
		rowWidth = 40
	}

	prefixWidth := runewidth.StringWidth("> ")
	indicatorWidth := runewidth.StringWidth(sourceIndicator + " ")
	name := truncateServiceName(i.svc.Name, rowWidth-prefixWidth-indicatorWidth)

	var str string
	if index == m.Index() {
		str = SelectedStyle.Render("> " + styledIndicator + " " + statusStyle.Render(name))
	} else {
		str = "  " + styledIndicator + " " + statusStyle.Render(name)
	}

	desc := truncateDescription(i.svc.Description, rowWidth-2)
	descStyle := DescriptionStyle

	if i.svc.Status == "failed" {
		descStyle = FailedStyle
	}

	str += "\n  " + descStyle.Render(desc)

	fmt.Fprint(w, str)
}

func truncateServiceName(name string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(name) <= maxWidth {
		return name
	}

	serviceSuffix := ".service"
	suffixWidth := runewidth.StringWidth(serviceSuffix)

	if strings.HasSuffix(name, serviceSuffix) && maxWidth > suffixWidth+1 {
		base := strings.TrimSuffix(name, serviceSuffix)
		truncatedBase := runewidth.Truncate(base, maxWidth-suffixWidth, "…")
		return truncatedBase + serviceSuffix
	}

	return runewidth.Truncate(name, maxWidth, "…")
}

func truncateDescription(desc string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if runewidth.StringWidth(desc) <= maxLen {
		return desc
	}
	return runewidth.Truncate(desc, maxLen, "…")
}

// sourceIndicator returns a visual indicator for the service source type.
func sourceIndicator(source client.ServiceSource) string {
	switch source {
	case client.SourceUser:
		return "●"
	case client.SourceSystem:
		return "○"
	case client.SourceTransient:
		return "◌"
	case client.SourceGenerated:
		return "◆"
	case client.SourceStatic:
		return "◇"
	default:
		return "·"
	}
}

func sourceStyle(source client.ServiceSource) lipgloss.Style {
	switch source {
	case client.SourceUser:
		return SourceUserStyle
	case client.SourceSystem:
		return SourceSystemStyle
	case client.SourceTransient:
		return SourceTransientStyle
	case client.SourceGenerated:
		return SourceGeneratedStyle
	case client.SourceStatic:
		return SourceStaticStyle
	default:
		return SourceOtherStyle
	}
}
