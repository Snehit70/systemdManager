package ui

import (
	"fmt"
	"systemd-tui/internal/service"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	unit service.Unit
}

func (i item) Title() string       { return i.unit.Unit }
func (i item) Description() string { return i.unit.Description }
func (i item) FilterValue() string { return i.unit.Unit }

type MainModel struct {
	list          list.Model
	viewport      viewport.Model
	units         []service.Unit
	systemdClient *service.SystemdClient
	width         int
	height        int

	listStyle   lipgloss.Style
	detailStyle lipgloss.Style
}

func NewMainModel(client *service.SystemdClient) MainModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "User Services"
	l.SetShowHelp(false)

	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().PaddingLeft(1)

	return MainModel{
		list:          l,
		viewport:      vp,
		systemdClient: client,
		listStyle:     lipgloss.NewStyle().MarginRight(1).Border(lipgloss.NormalBorder(), false, true, false, false),
		detailStyle:   lipgloss.NewStyle().PaddingLeft(1),
	}
}

func (m MainModel) Init() tea.Cmd {
	return m.fetchUnits
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		listWidth := m.width / 3
		m.list.SetSize(listWidth, m.height-2)

		detailWidth := m.width - listWidth - 2
		m.viewport.Width = detailWidth
		m.viewport.Height = m.height - 2

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case []service.Unit:
		m.units = msg
		items := make([]list.Item, len(msg))
		for i, unit := range msg {
			items[i] = item{unit: unit}
		}
		cmd = m.list.SetItems(items)
		cmds = append(cmds, cmd)
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	selectedItem := m.list.SelectedItem()
	if selectedItem != nil {
		unit := selectedItem.(item).unit
		content := fmt.Sprintf(
			"Service: %s\nStatus: %s (%s)\nDescription: %s\n\n[Logs would go here...]",
			unit.Unit, unit.Active, unit.Sub, unit.Description,
		)
		m.viewport.SetContent(content)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.listStyle.Render(m.list.View()),
		m.viewport.View(),
	)
}

func (m MainModel) fetchUnits() tea.Msg {
	units, err := m.systemdClient.ListUnits()
	if err != nil {
		return nil
	}
	return units
}
