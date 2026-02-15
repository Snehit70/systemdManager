package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"systemd-tui/internal/client"
	"systemd-tui/internal/config"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	svc client.Service
}

func (i item) Title() string       { return i.svc.Name }
func (i item) Description() string { return i.svc.Description }
func (i item) FilterValue() string { return i.svc.Name }

type keyMap struct {
	Up           key.Binding
	Down         key.Binding
	Start        key.Binding
	Stop         key.Binding
	Restart      key.Binding
	Edit         key.Binding
	Enable       key.Binding
	Disable      key.Binding
	Filter       key.Binding
	SwitchFocus  key.Binding
	ToggleFollow key.Binding
	ToggleGroup  key.Binding
	ToggleSource key.Binding
	Create       key.Binding
	Quit         key.Binding
	Help         key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.SwitchFocus, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Filter, k.SwitchFocus},
		{k.Start, k.Stop, k.Restart, k.Edit},
		{k.Enable, k.Disable, k.ToggleFollow, k.ToggleGroup},
		{k.Create, k.ToggleSource, k.Quit, k.Help},
	}
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
	Enable: key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp("E", "enable"),
	),
	Disable: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "disable"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	SwitchFocus: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch view"),
	),
	ToggleFollow: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "follow logs"),
	),
	ToggleGroup: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "group by status"),
	),
	ToggleSource: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "filter source"),
	),
	Create: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "create service"),
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

type groupMode int

const (
	groupNone groupMode = iota
	groupByStatus
	groupByLoad
)

func (g groupMode) String() string {
	switch g {
	case groupByStatus:
		return "status"
	case groupByLoad:
		return "load"
	default:
		return "none"
	}
}

func (g groupMode) Next() groupMode {
	return (g + 1) % 3
}

type filterMode int

const (
	filterAll filterMode = iota
	filterUserOnly
	filterHideSystem
)

func (f filterMode) String() string {
	switch f {
	case filterUserOnly:
		return "my services"
	case filterHideSystem:
		return "hide system"
	default:
		return "all"
	}
}

func (f filterMode) Next() filterMode {
	return (f + 1) % 3
}

type createModal struct {
	focusIndex   int
	nameInput    textinput.Model
	execInput    textinput.Model
	descInput    textinput.Model
	workdirInput textinput.Model
	serviceType  string
	restart      string
}

func newCreateModal() createModal {
	name := textinput.New()
	name.Placeholder = "my-service"
	name.Focus()
	name.CharLimit = 100

	exec := textinput.New()
	exec.Placeholder = "/path/to/command --args"
	exec.CharLimit = 500

	desc := textinput.New()
	desc.Placeholder = "Service description"
	desc.CharLimit = 200

	workdir := textinput.New()
	workdir.Placeholder = "~"
	workdir.CharLimit = 200

	return createModal{
		nameInput:    name,
		execInput:    exec,
		descInput:    desc,
		workdirInput: workdir,
		serviceType:  "simple",
		restart:      "on-failure",
	}
}

type MainModel struct {
	list          list.Model
	viewport      viewport.Model
	help          help.Model
	services      []client.Service
	client        client.ServiceClient
	config        *config.Config
	theme         config.ThemeColors
	width         int
	height        int
	statusMessage string
	selectedSvc   string
	activeView    int

	confirmingAction string
	confirmingUnit   string

	following      bool
	followCancel   context.CancelFunc
	followLogLines []string

	groupMode  groupMode
	filterMode filterMode

	createModal createModal
	showCreate  bool

	activeBorder   lipgloss.Style
	inactiveBorder lipgloss.Style
	detailStyle    lipgloss.Style
}

func NewMainModel(client client.ServiceClient, cfg *config.Config) MainModel {
	theme := cfg.ThemeColors()
	ApplyTheme(theme)

	l := list.New(nil, itemDelegate{}, 0, 0)
	l.Title = "User Services"
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)

	listKeys := list.DefaultKeyMap()
	listKeys.Quit = key.NewBinding(key.WithKeys("ctrl+c"))
	l.KeyMap = listKeys

	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().PaddingLeft(1)

	active := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.BorderActive)).
		MarginRight(1)

	inactive := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Border)).
		MarginRight(1)

	return MainModel{
		list:           l,
		viewport:       vp,
		help:           help.New(),
		client:         client,
		config:         cfg,
		theme:          theme,
		activeView:     listView,
		activeBorder:   active,
		inactiveBorder: inactive,
		detailStyle:    lipgloss.NewStyle().PaddingLeft(1),
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(m.fetchServices, m.tick())
}

func (m MainModel) tick() tea.Cmd {
	interval := m.config.General.RefreshInterval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		helpHeight := 2
		statusBarHeight := 1
		mainHeight := m.height - helpHeight - statusBarHeight

		listWidth := m.width / 3
		detailWidth := m.width - listWidth - 4

		m.list.SetSize(listWidth-2, mainHeight-2)
		m.viewport.Width = detailWidth - 2
		m.viewport.Height = mainHeight - 2
		m.help.Width = m.width

	case list.FilterMatchesMsg:
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.showCreate {
			return m.handleCreateModal(msg)
		}

		if m.confirmingAction != "" {
			switch msg.String() {
			case "y", "Y":
				var actionCmd tea.Cmd
				switch m.confirmingAction {
				case "stop":
					m.statusMessage = "Stopping " + m.confirmingUnit + "..."
					actionCmd = m.stopService(m.confirmingUnit)
				case "restart":
					m.statusMessage = "Restarting " + m.confirmingUnit + "..."
					actionCmd = m.restartService(m.confirmingUnit)
				case "enable":
					m.statusMessage = "Enabling " + m.confirmingUnit + "..."
					actionCmd = m.enableService(m.confirmingUnit)
				case "disable":
					m.statusMessage = "Disabling " + m.confirmingUnit + "..."
					actionCmd = m.disableService(m.confirmingUnit)
				}
				m.confirmingAction = ""
				m.confirmingUnit = ""
				return m, actionCmd
			case "n", "N", "esc":
				m.statusMessage = "Cancelled"
				m.confirmingAction = ""
				m.confirmingUnit = ""
				return m, nil
			default:
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, keys.Quit):
			if m.following {
				m.stopFollow()
			}
			return m, tea.Quit
		case key.Matches(msg, keys.SwitchFocus):
			if m.activeView == listView {
				m.activeView = detailView
			} else {
				m.activeView = listView
			}
			return m, nil
		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, keys.ToggleFollow):
			if m.following {
				m.stopFollow()
				m.statusMessage = "Stopped following logs"
			} else {
				if selected := m.list.SelectedItem(); selected != nil {
					svc := selected.(item).svc
					cmd = m.startFollow(svc.Name)
					m.statusMessage = "Following logs for " + svc.Name
					return m, cmd
				}
			}
			return m, nil
		case key.Matches(msg, keys.ToggleGroup):
			m.groupMode = m.groupMode.Next()
			m.statusMessage = fmt.Sprintf("Group by: %s", m.groupMode)
			cmds = append(cmds, m.updateListItems())

		case key.Matches(msg, keys.ToggleSource):
			m.filterMode = m.filterMode.Next()
			m.statusMessage = fmt.Sprintf("Filter: %s", m.filterMode)
			cmds = append(cmds, m.updateListItems())

		case key.Matches(msg, keys.Create):
			m.showCreate = true
			m.createModal = newCreateModal()
			return m, nil
		}

		if m.activeView == listView {
			switch {
			case key.Matches(msg, keys.Restart):
				if selected := m.list.SelectedItem(); selected != nil {
					svc := selected.(item).svc
					m.confirmingAction = "restart"
					m.confirmingUnit = svc.Name
					m.statusMessage = fmt.Sprintf("Restart %s? (y/n)", svc.Name)
					return m, nil
				}
			case key.Matches(msg, keys.Start):
				if selected := m.list.SelectedItem(); selected != nil {
					svc := selected.(item).svc
					m.statusMessage = "Starting " + svc.Name + "..."
					return m, m.startService(svc.Name)
				}
			case key.Matches(msg, keys.Stop):
				if selected := m.list.SelectedItem(); selected != nil {
					svc := selected.(item).svc
					m.confirmingAction = "stop"
					m.confirmingUnit = svc.Name
					m.statusMessage = fmt.Sprintf("Stop %s? (y/n)", svc.Name)
					return m, nil
				}
			case key.Matches(msg, keys.Edit):
				if selected := m.list.SelectedItem(); selected != nil {
					svc := selected.(item).svc
					return m, m.editService(svc.Name)
				}
			case key.Matches(msg, keys.Enable):
				if selected := m.list.SelectedItem(); selected != nil {
					svc := selected.(item).svc
					m.confirmingAction = "enable"
					m.confirmingUnit = svc.Name
					m.statusMessage = fmt.Sprintf("Enable %s? (y/n)", svc.Name)
					return m, nil
				}
			case key.Matches(msg, keys.Disable):
				if selected := m.list.SelectedItem(); selected != nil {
					svc := selected.(item).svc
					m.confirmingAction = "disable"
					m.confirmingUnit = svc.Name
					m.statusMessage = fmt.Sprintf("Disable %s? (y/n)", svc.Name)
					return m, nil
				}
			default:
				var prevItem list.Item
				if m.list.SelectedItem() != nil {
					prevItem = m.list.SelectedItem()
				}

				m.list, cmd = m.list.Update(msg)
				cmds = append(cmds, cmd)

				if m.list.SelectedItem() != nil {
					currItem := m.list.SelectedItem()
					if prevItem == nil || currItem.FilterValue() != prevItem.FilterValue() {
						svc := currItem.(item).svc
						m.selectedSvc = svc.Name
						cmds = append(cmds, m.fetchLogs(svc.Name))
					}
				}
			}
		} else {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case []client.Service:
		m.services = msg
		userCount := 0
		for _, svc := range m.services {
			if svc.Source == client.SourceUser {
				userCount++
			}
		}
		m.statusMessage = fmt.Sprintf("Loaded %d services (%d user-created)", len(msg), userCount)

		cmds = append(cmds, m.updateListItems())

		if m.list.SelectedItem() != nil {
			svc := m.list.SelectedItem().(item).svc
			cmds = append(cmds, m.fetchLogs(svc.Name))
		}

	case actionResultMsg:
		m.statusMessage = msg.message
		if msg.err != nil {
			m.statusMessage = "Error: " + msg.err.Error()
		} else {
			cmds = append(cmds, m.fetchServices)
		}

	case editorFinishedMsg:
		if msg.err != nil {
			m.statusMessage = "Edit failed: " + msg.err.Error()
		} else {
			if err := m.client.ReloadDaemon(); err != nil {
				m.statusMessage = "Edit saved, but reload failed: " + err.Error()
			} else {
				m.statusMessage = "Edit saved. Reloaded daemon."
				cmds = append(cmds, m.fetchServices)
			}
		}
		return m, tea.Batch(cmds...)

	case logMsg:
		if msg.unit == m.selectedSvc && !m.following {
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				svc := selectedItem.(item).svc
				lines := m.config.General.LogLines
				if lines <= 0 {
					lines = 50
				}
				var content string
				if msg.err != nil {
					content = fmt.Sprintf(
						"Service: %s\nStatus: %s (%s)\nDescription: %s\n\n[Log Error]\n%s",
						svc.Name, svc.Status, svc.Sub, svc.Description, msg.err.Error(),
					)
				} else {
					content = fmt.Sprintf(
						"Service: %s\nStatus: %s (%s)\nDescription: %s\n\n[Last %d Lines of Log]\n%s",
						svc.Name, svc.Status, svc.Sub, svc.Description, lines, msg.logs,
					)
				}
				m.viewport.SetContent(content)
			}
		}

	case logLineMsg:
		if m.following {
			m.followLogLines = append(m.followLogLines, msg.line)
			maxLines := 1000
			if len(m.followLogLines) > maxLines {
				m.followLogLines = m.followLogLines[len(m.followLogLines)-maxLines:]
			}
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				svc := selectedItem.(item).svc
				content := fmt.Sprintf(
					"Service: %s\nStatus: %s (%s)\nDescription: %s\n\n[FOLLOWING - Press f to stop]\n%s",
					svc.Name, svc.Status, svc.Sub, svc.Description,
					strings.Join(m.followLogLines, "\n"),
				)
				m.viewport.SetContent(content)
				m.viewport.GotoBottom()
			}
		}

	case errMsg:
		m.statusMessage = fmt.Sprintf("Error: %s - %s", msg.context, msg.err.Error())

	case tickMsg:
		// Don't refresh while filtering - it would reset the filter
		if m.list.FilterState() == list.Filtering {
			return m, m.tick()
		}
		return m, tea.Batch(m.fetchServices, m.tick())
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

	statusBar := ""
	if m.statusMessage != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.TextMuted)).
			PaddingLeft(1).
			Width(fullWidth)
		statusBar = statusStyle.Render(m.statusMessage)
	}

	filterBar := ""
	if m.list.FilterState() == list.Filtering {
		filterStyle := lipgloss.NewStyle().
			PaddingLeft(1).
			Width(fullWidth)
		filterBar = filterStyle.Render(m.list.FilterInput.View())
	}

	helpView := renderHelpBar(keys, fullWidth, m.theme)

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

func (m MainModel) renderCreateModal(baseView string) string {
	modalWidth := 60
	modalHeight := 18

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Text)).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.TextMuted))

	focusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.BorderActive)).
		Bold(true)

	var inputs []string

	nameLabel := "Name:"
	if m.createModal.focusIndex == 0 {
		nameLabel = focusStyle.Render("Name:")
	} else {
		nameLabel = labelStyle.Render("Name:")
	}
	inputs = append(inputs, fmt.Sprintf("%s %s", nameLabel, m.createModal.nameInput.View()))

	execLabel := "Command:"
	if m.createModal.focusIndex == 1 {
		execLabel = focusStyle.Render("Command:")
	} else {
		execLabel = labelStyle.Render("Command:")
	}
	inputs = append(inputs, fmt.Sprintf("%s %s", execLabel, m.createModal.execInput.View()))

	descLabel := "Description:"
	if m.createModal.focusIndex == 2 {
		descLabel = focusStyle.Render("Description:")
	} else {
		descLabel = labelStyle.Render("Description:")
	}
	inputs = append(inputs, fmt.Sprintf("%s %s", descLabel, m.createModal.descInput.View()))

	workdirLabel := "Workdir:"
	if m.createModal.focusIndex == 3 {
		workdirLabel = focusStyle.Render("Workdir:")
	} else {
		workdirLabel = labelStyle.Render("Workdir:")
	}
	inputs = append(inputs, fmt.Sprintf("%s %s", workdirLabel, m.createModal.workdirInput.View()))

	typeLabel := "Type:"
	if m.createModal.focusIndex == 4 {
		typeLabel = focusStyle.Render("Type:")
	} else {
		typeLabel = labelStyle.Render("Type:")
	}
	inputs = append(inputs, fmt.Sprintf("%s %s", typeLabel, inputStyle.Render(m.createModal.serviceType+" (t to toggle)")))

	restartLabel := "Restart:"
	if m.createModal.focusIndex == 5 {
		restartLabel = focusStyle.Render("Restart:")
	} else {
		restartLabel = labelStyle.Render("Restart:")
	}
	inputs = append(inputs, fmt.Sprintf("%s %s", restartLabel, inputStyle.Render(m.createModal.restart+" (r to cycle)")))

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
				inputStyle.Render("Tab: next • Shift+Tab: prev • Enter: create • Esc: cancel"),
			),
		)

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.BorderActive)).
		Background(lipgloss.Color(m.theme.Surface)).
		Render(content)

	overlayStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Faint(true).Render(baseView[:min(len(baseView), 100)]),
		overlayStyle.Render(modal),
	)
}

func (m MainModel) fetchServices() tea.Msg {
	services, err := m.client.ListServices()
	if err != nil {
		return errMsg{context: "Failed to list services", err: err}
	}
	return services
}

type errMsg struct {
	context string
	err     error
}

func (e errMsg) Error() string { return e.err.Error() }

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

type logLineMsg struct {
	line string
}

type tickMsg time.Time

func (m MainModel) fetchLogs(unit string) tea.Cmd {
	return func() tea.Msg {
		lines := m.config.General.LogLines
		if lines <= 0 {
			lines = 50
		}
		logs, err := m.client.GetLogs(unit, client.LogOptions{Lines: lines})
		return logMsg{unit: unit, logs: logs, err: err}
	}
}

func (m MainModel) startService(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.StartService(unit)
		return actionResultMsg{message: "Started " + unit, err: err}
	}
}

func (m MainModel) stopService(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.StopService(unit)
		return actionResultMsg{message: "Stopped " + unit, err: err}
	}
}

func (m MainModel) restartService(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.RestartService(unit)
		return actionResultMsg{message: "Restarted " + unit, err: err}
	}
}

func (m MainModel) enableService(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.EnableService(unit)
		return actionResultMsg{message: "Enabled " + unit, err: err}
	}
}

func (m MainModel) disableService(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.DisableService(unit)
		return actionResultMsg{message: "Disabled " + unit, err: err}
	}
}

func (m MainModel) editService(unit string) tea.Cmd {
	cmd, err := m.client.EditService(unit)
	if err != nil {
		return func() tea.Msg {
			return editorFinishedMsg{err: err}
		}
	}
	return tea.ExecProcess(
		cmd,
		func(err error) tea.Msg {
			return editorFinishedMsg{err: err}
		},
	)
}

func (m *MainModel) startFollow(unit string) tea.Cmd {
	logChan, cancel, err := m.client.FollowLogs(unit, client.LogOptions{Lines: 50})
	if err != nil {
		return func() tea.Msg {
			return errMsg{context: "Failed to follow logs", err: err}
		}
	}

	m.following = true
	m.followCancel = cancel
	m.followLogLines = nil

	return func() tea.Msg {
		for line := range logChan {
			return logLineMsg{line: line}
		}
		return nil
	}
}

func (m *MainModel) stopFollow() {
	if m.followCancel != nil {
		m.followCancel()
	}
	m.following = false
	m.followCancel = nil
	m.followLogLines = nil
}

func (m MainModel) updateListItems() tea.Cmd {
	var items []list.Item

	switch m.groupMode {
	case groupByStatus:
		items = m.groupedByStatus()
	case groupByLoad:
		items = m.groupedByLoad()
	default:
		for _, svc := range m.filteredServices() {
			items = append(items, item{svc: svc})
		}
	}

	return m.list.SetItems(items)
}

type groupHeaderItem struct {
	title string
}

func (g groupHeaderItem) Title() string       { return g.title }
func (g groupHeaderItem) Description() string { return "" }
func (g groupHeaderItem) FilterValue() string { return "" }

func (m MainModel) groupedByStatus() []list.Item {
	active := make([]list.Item, 0)
	failed := make([]list.Item, 0)
	inactive := make([]list.Item, 0)

	for _, svc := range m.filteredServices() {
		switch svc.Status {
		case client.StatusActive:
			active = append(active, item{svc: svc})
		case client.StatusFailed:
			failed = append(failed, item{svc: svc})
		default:
			inactive = append(inactive, item{svc: svc})
		}
	}

	var items []list.Item
	if len(active) > 0 {
		items = append(items, groupHeaderItem{title: "── Active ──"})
		items = append(items, active...)
	}
	if len(failed) > 0 {
		items = append(items, groupHeaderItem{title: "── Failed ──"})
		items = append(items, failed...)
	}
	if len(inactive) > 0 {
		items = append(items, groupHeaderItem{title: "── Inactive ──"})
		items = append(items, inactive...)
	}

	return items
}

func (m MainModel) groupedByLoad() []list.Item {
	loaded := make([]list.Item, 0)
	notFound := make([]list.Item, 0)
	other := make([]list.Item, 0)

	for _, svc := range m.filteredServices() {
		switch svc.Load {
		case "loaded":
			loaded = append(loaded, item{svc: svc})
		case "not-found":
			notFound = append(notFound, item{svc: svc})
		default:
			other = append(other, item{svc: svc})
		}
	}

	var items []list.Item
	if len(loaded) > 0 {
		items = append(items, groupHeaderItem{title: "── Loaded ──"})
		items = append(items, loaded...)
	}
	if len(notFound) > 0 {
		items = append(items, groupHeaderItem{title: "── Not Found ──"})
		items = append(items, notFound...)
	}
	if len(other) > 0 {
		items = append(items, groupHeaderItem{title: "── Other ──"})
		items = append(items, other...)
	}

	return items
}

func (m MainModel) filteredServices() []client.Service {
	if m.filterMode == filterAll {
		return m.services
	}

	filtered := make([]client.Service, 0)
	for _, svc := range m.services {
		switch m.filterMode {
		case filterUserOnly:
			if svc.Source == client.SourceUser {
				filtered = append(filtered, svc)
			}
		case filterHideSystem:
			if svc.Source != client.SourceStatic && svc.Source != client.SourceGenerated && svc.Source != client.SourceTransient {
				filtered = append(filtered, svc)
			}
		}
	}
	return filtered
}

func (m MainModel) handleCreateModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.showCreate = false
		return m, nil

	case "tab", "shift+tab":
		if msg.String() == "shift+tab" {
			m.createModal.focusIndex--
			if m.createModal.focusIndex < 0 {
				m.createModal.focusIndex = 5
			}
		} else {
			m.createModal.focusIndex = (m.createModal.focusIndex + 1) % 6
		}
		m.focusCreateInput()
		return m, nil

	case "enter":
		return m.createServiceFromModal()

	case "t", "T":
		if m.createModal.focusIndex == 4 {
			if m.createModal.serviceType == "simple" {
				m.createModal.serviceType = "oneshot"
			} else {
				m.createModal.serviceType = "simple"
			}
			return m, nil
		}

	case "r", "R":
		if m.createModal.focusIndex == 5 {
			switch m.createModal.restart {
			case "on-failure":
				m.createModal.restart = "always"
			case "always":
				m.createModal.restart = "no"
			default:
				m.createModal.restart = "on-failure"
			}
			return m, nil
		}
	}

	switch m.createModal.focusIndex {
	case 0:
		m.createModal.nameInput, _ = m.createModal.nameInput.Update(msg)
	case 1:
		m.createModal.execInput, _ = m.createModal.execInput.Update(msg)
	case 2:
		m.createModal.descInput, _ = m.createModal.descInput.Update(msg)
	case 3:
		m.createModal.workdirInput, _ = m.createModal.workdirInput.Update(msg)
	}

	return m, nil
}

func (m *MainModel) focusCreateInput() {
	m.createModal.nameInput.Blur()
	m.createModal.execInput.Blur()
	m.createModal.descInput.Blur()
	m.createModal.workdirInput.Blur()

	switch m.createModal.focusIndex {
	case 0:
		m.createModal.nameInput.Focus()
	case 1:
		m.createModal.execInput.Focus()
	case 2:
		m.createModal.descInput.Focus()
	case 3:
		m.createModal.workdirInput.Focus()
	}
}

func (m MainModel) createServiceFromModal() (tea.Model, tea.Cmd) {
	name := m.createModal.nameInput.Value()
	exec := m.createModal.execInput.Value()

	if name == "" || exec == "" {
		m.statusMessage = "Error: name and command are required"
		return m, nil
	}

	tmpl := client.ServiceTemplate{
		Name:             name,
		Description:      m.createModal.descInput.Value(),
		ExecStart:        exec,
		WorkingDirectory: m.createModal.workdirInput.Value(),
		Type:             m.createModal.serviceType,
		Restart:          m.createModal.restart,
	}

	err := m.client.CreateService(tmpl)
	if err != nil {
		m.statusMessage = "Failed to create service: " + err.Error()
		return m, nil
	}

	m.showCreate = false
	m.statusMessage = fmt.Sprintf("Created service: %s", name)
	return m, m.fetchServices
}

type serviceCreatedMsg struct {
	name string
	err  error
}
