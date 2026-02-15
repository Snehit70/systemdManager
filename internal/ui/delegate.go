package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

	var str string
	if index == m.Index() {
		str = SelectedStyle.Render("> " + statusStyle.Render(i.svc.Name))
	} else {
		str = "  " + statusStyle.Render(i.svc.Name)
	}

	desc := i.svc.Description
	if len(desc) > 50 {
		desc = desc[:47] + "..."
	}
	descStyle := DescriptionStyle

	if i.svc.Status == "failed" {
		descStyle = FailedStyle
	}

	str += "\n  " + descStyle.Render(desc)

	fmt.Fprint(w, str)
}
