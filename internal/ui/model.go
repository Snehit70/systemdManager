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
	Up          key.Binding
	Down        key.Binding
	Start       key.Binding
	Stop        key.Binding
	Restart     key.Binding
	Edit        key.Binding
	Filter      key.Binding
	SwitchFocus key.Binding
	Quit        key.Binding
	Help        key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.SwitchFocus, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Filter, k.SwitchFocus},
		{k.Start, k.Stop, k.Restart, k.Edit},
		{k.Quit, k.Help},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move"),
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
	SwitchFocus: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch view"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

const (
	listView = iota
	detailView
)

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
	activeView    int

	activeBorder   lipgloss.Style
	inactiveBorder lipgloss.Style
	detailStyle    lipgloss.Style
}

func NewMainModel(client *service.SystemdClient) MainModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "User Services"
	l.SetShowHelp(false)

	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().PaddingLeft(1)

	active := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		MarginRight(1)

	inactive := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		MarginRight(1)

	return MainModel{
		list:           l,
		viewport:       vp,
		help:           help.New(),
		systemdClient:  client,
		activeView:     listView,
		activeBorder:   active,
		inactiveBorder: inactive,
		detailStyle:    lipgloss.NewStyle().PaddingLeft(1),
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

		helpHeight := 4
		mainHeight := m.height - helpHeight

		listWidth := m.width / 3

		m.list.SetSize(listWidth-2, mainHeight-2)

		detailWidth := m.width - listWidth - 4
		m.viewport.Width = detailWidth
		m.viewport.Height = mainHeight - 2
		m.help.Width = m.width

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch {

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

			var prevItem list.Item
			if m.list.SelectedItem() != nil {
				prevItem = m.list.SelectedItem()
			}
			
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			
			if m.list.SelectedItem() != nil {
				currItem := m.list.SelectedItem()
				if prevItem == nil || currItem.FilterValue() != prevItem.FilterValue() {
					unit := currItem.(item).unit.Unit
					m.selectedUnit = unit
					cmds = append(cmds, m.fetchLogs(unit))
				}
			}
			
		} else {


			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
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

	return m, tea.Batch(cmds...)
}


func (m MainModel) View() string {
	var listStyle, detailStyle lipgloss.Style

	if m.activeView == listView {
		listStyle = m.activeBorder
		detailStyle = m.inactiveBorder
	} else {
		listStyle = m.inactiveBorder
		detailStyle = m.activeBorder
	}

	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(m.list.View()),
		detailStyle.Render(m.viewport.View()),
	)

	helpView := lipgloss.NewStyle().
		PaddingTop(1).
		PaddingLeft(1).
		Render(m.help.View(keys))

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
