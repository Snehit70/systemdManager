package ui

import (
	"fmt"
	"systemd-tui/internal/service"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Start   key.Binding
	Stop    key.Binding
	Restart key.Binding
	Edit    key.Binding
	Filter  key.Binding
	Quit    key.Binding
	Help    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Filter},
		{k.Start, k.Stop, k.Restart, k.Edit},
		{k.Quit, k.Help},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Start: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "start"),
	),
	Stop: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "stop"),
	),
	Restart: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "restart"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

type MainModel struct {
	list          list.Model
	viewport      viewport.Model
	help          help.Model
	units         []service.Unit
	systemdClient *service.SystemdClient
	width         int
	height        int
	statusMessage string
	selectedUnit  string

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
		help:          help.New(),
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

		helpHeight := 2
		mainHeight := m.height - helpHeight - 1

		listWidth := m.width / 3
		m.list.SetSize(listWidth, mainHeight)

		detailWidth := m.width - listWidth - 2
		m.viewport.Width = detailWidth
		m.viewport.Height = mainHeight
		m.help.Width = m.width

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, keys.Restart):
			if selected := m.list.SelectedItem(); selected != nil {
				unit := selected.(item).unit.Unit
				m.statusMessage = "Restarting " + unit + "..."
				return m, m.restartUnit(unit)
			}
		case key.Matches(msg, keys.Start):
			if selected := m.list.SelectedItem(); selected != nil {
				unit := selected.(item).unit.Unit
				m.statusMessage = "Starting " + unit + "..."
				return m, m.startUnit(unit)
			}
		case key.Matches(msg, keys.Stop):
			if selected := m.list.SelectedItem(); selected != nil {
				unit := selected.(item).unit.Unit
				m.statusMessage = "Stopping " + unit + "..."
				return m, m.stopUnit(unit)
			}
		case key.Matches(msg, keys.Edit):
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

		if m.list.SelectedItem() != nil {
			unit := m.list.SelectedItem().(item).unit.Unit
			cmds = append(cmds, m.fetchLogs(unit))
		}

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

	case logMsg:
		if msg.unit == m.selectedUnit {
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				unit := selectedItem.(item).unit
				content := fmt.Sprintf(
					"Service: %s\nStatus: %s (%s)\nDescription: %s\n\n[Last 50 Lines of Log]\n%s",
					unit.Unit, unit.Active, unit.Sub, unit.Description, msg.logs,
				)
				m.viewport.SetContent(content)
			}
		}
	}

	previousUnit := ""
	if m.list.SelectedItem() != nil {
		previousUnit = m.list.SelectedItem().(item).unit.Unit
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	currentUnit := ""
	if m.list.SelectedItem() != nil {
		currentUnit = m.list.SelectedItem().(item).unit.Unit
	}

	if currentUnit != "" && currentUnit != previousUnit {
		m.selectedUnit = currentUnit
		cmds = append(cmds, m.fetchLogs(currentUnit))
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.listStyle.Render(m.list.View()),
		m.viewport.View(),
	)

	helpView := m.help.View(keys)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		helpView,
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

type logMsg struct {
	unit string
	logs string
	err  error
}

func (m MainModel) fetchLogs(unit string) tea.Cmd {
	return func() tea.Msg {
		logs, err := m.systemdClient.GetLogs(unit)
		return logMsg{unit: unit, logs: logs, err: err}
	}
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
