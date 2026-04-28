package ui

import (
	"fmt"
	"io"

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
		fmt.Fprint(w, headerStyle.Render(header.title))
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

	name := truncateName(i.svc.Name, 35)

	var str string
	if index == m.Index() {
		str = SelectedStyle.Render("> " + styledIndicator + " " + statusStyle.Render(name))
	} else {
		str = "  " + styledIndicator + " " + statusStyle.Render(name)
	}

	desc := truncateDescription(i.svc.Description, 50)
	descStyle := DescriptionStyle

	if i.svc.Status == "failed" {
		descStyle = FailedStyle
	}

	str += "\n  " + descStyle.Render(desc)

	fmt.Fprint(w, str)
}

func truncateName(name string, maxWidth int) string {
	prefix := "  "
	indicator := " "
	availableWidth := maxWidth - runewidth.StringWidth(prefix) - runewidth.StringWidth(indicator)

	currentWidth := runewidth.StringWidth(name)
	if currentWidth <= availableWidth {
		return name
	}

	serviceSuffix := ".service"
	suffixWidth := runewidth.StringWidth(serviceSuffix)

	if suffixWidth < currentWidth && currentWidth-suffixWidth <= availableWidth {
		truncated := name[:len(name)-len(serviceSuffix)]
		for runewidth.StringWidth(truncated)+suffixWidth > availableWidth && len(truncated) > 0 {
			truncated = truncated[:len(truncated)-1]
		}
		return truncated + serviceSuffix
	}

	for runewidth.StringWidth(name) > availableWidth && len(name) > 0 {
		name = name[:len(name)-1]
	}
	return name + "…"
}

func truncateDescription(desc string, maxLen int) string {
	if runewidth.StringWidth(desc) <= maxLen {
		return desc
	}
	for runewidth.StringWidth(desc) > maxLen && len(desc) > 0 {
		desc = desc[:len(desc)-1]
	}
	return desc + "…"
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
