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
	rowWidth := m.Width()
	if rowWidth <= 0 {
		rowWidth = 40
	}

	if header, ok := listItem.(groupHeaderItem); ok {
		label := GroupHeaderStyle.Render(strings.ToUpper(header.title))
		labelWidth := runewidth.StringWidth(strings.ToUpper(header.title))
		ruleLen := rowWidth - labelWidth - 1
		if ruleLen < 0 {
			ruleLen = 0
		}
		rule := GroupHeaderRuleStyle.Render(strings.Repeat("─", ruleLen))
		fmt.Fprint(w, label+" "+rule)
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

	prefixWidth := runewidth.StringWidth("▎ ")
	indicatorWidth := runewidth.StringWidth(sourceIndicator + " ")
	name := truncateServiceName(i.svc.Name, rowWidth-prefixWidth-indicatorWidth)

	selected := index == m.Index()
	var prefix string
	if selected {
		prefix = SelectedStyle.Render("▎")
	} else {
		prefix = " "
	}

	nameRender := statusStyle.Render(name)
	if selected {
		nameRender = lipgloss.NewStyle().Bold(true).Inherit(statusStyle).Render(name)
	}

	titleRow := prefix + " " + styledIndicator + " " + nameRender

	desc := truncateDescription(i.svc.Description, rowWidth-3)
	descStyle := DescriptionStyle
	if i.svc.Status == "failed" {
		descStyle = FailedStyle
	}
	descRow := "   " + descStyle.Render(desc)

	fmt.Fprint(w, titleRow+"\n"+descRow)
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
