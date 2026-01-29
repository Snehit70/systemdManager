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
	statusMessage string

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
		if m.list.FilterState() == list.Filtering {
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			if selected := m.list.SelectedItem(); selected != nil {
				unit := selected.(item).unit.Unit
				m.statusMessage = "Restarting " + unit + "..."
				return m, m.restartUnit(unit)
			}
		case "s":
			if selected := m.list.SelectedItem(); selected != nil {
				unit := selected.(item).unit.Unit
				m.statusMessage = "Starting " + unit + "..."
				return m, m.startUnit(unit)
			}
		case "x":
			if selected := m.list.SelectedItem(); selected != nil {
				unit := selected.(item).unit.Unit
				m.statusMessage = "Stopping " + unit + "..."
				return m, m.stopUnit(unit)
			}
		case "e":
			if selected := m.list.SelectedItem(); selected != nil {
				unit := selected.(item).unit.Unit
				return m, m.editUnit(unit)
			}
		}

	case []service.Unit:
		m.units = msg
		items := make([]list.Item, len(msg))
		for i, unit := range msg {
			items[i] = item{unit: unit}
		}
		cmd = m.list.SetItems(items)
		cmds = append(cmds, cmd)
		m.statusMessage = "Refreshed."

	case actionResultMsg:
		m.statusMessage = msg.message
		if msg.err != nil {
			m.statusMessage = "Error: " + msg.err.Error()
		} else {
			cmds = append(cmds, m.fetchUnits)
		}

	case editorFinishedMsg:
		if msg.err != nil {
			m.statusMessage = "Edit failed: " + msg.err.Error()
		} else {
			if err := m.systemdClient.ReloadDaemon(); err != nil {
				m.statusMessage = "Edit saved, but reload failed: " + err.Error()
			} else {
				m.statusMessage = "Edit saved. Reloaded daemon."
				cmds = append(cmds, m.fetchUnits)
			}
		}
		return m, tea.Batch(cmds...)
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	selectedItem := m.list.SelectedItem()
	if selectedItem != nil {
		unit := selectedItem.(item).unit
		content := fmt.Sprintf(
			"Service: %s\nStatus: %s (%s)\nDescription: %s\n\n[Action Log]\n%s",
			unit.Unit, unit.Active, unit.Sub, unit.Description, m.statusMessage,
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

type actionResultMsg struct {
	message string
	err     error
}

type editorFinishedMsg struct {
	err error
}

func (m MainModel) startUnit(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.systemdClient.StartUnit(unit)
		return actionResultMsg{message: "Started " + unit, err: err}
	}
}

func (m MainModel) stopUnit(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.systemdClient.StopUnit(unit)
		return actionResultMsg{message: "Stopped " + unit, err: err}
	}
}

func (m MainModel) restartUnit(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.systemdClient.RestartUnit(unit)
		return actionResultMsg{message: "Restarted " + unit, err: err}
	}
}

func (m MainModel) editUnit(unit string) tea.Cmd {
	return tea.ExecProcess(
		m.systemdClient.EditCmd(unit),
		func(err error) tea.Msg {
			return editorFinishedMsg{err: err}
		},
	)
}
