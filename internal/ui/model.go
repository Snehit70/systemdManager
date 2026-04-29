package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

func (m MainModel) getSelectedService() *client.Service {
	selected := m.list.SelectedItem()
	if selected == nil {
		return nil
	}
	if it, ok := selected.(item); ok {
		return &it.svc
	}
	return nil
}

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
	ToggleDetail key.Binding
	ShrinkPanel  key.Binding
	GrowPanel    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.SwitchFocus, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Filter, k.SwitchFocus},
		{k.Start, k.Stop, k.Restart, k.Edit},
		{k.Enable, k.Disable, k.ToggleFollow, k.ToggleGroup},
		{k.Create, k.ToggleSource, k.ToggleDetail, k.Quit},
		{k.ShrinkPanel, k.GrowPanel, k.Help},
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
	ToggleDetail: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle detail view"),
	),
	ShrinkPanel: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "shrink list"),
	),
	GrowPanel: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "grow list"),
	),
}

const (
	listView = iota
	detailView
)

type detailViewMode int

const (
	detailViewLogs detailViewMode = iota
	detailViewStatus
	detailViewConfig
)

func (d detailViewMode) String() string {
	switch d {
	case detailViewStatus:
		return "status"
	case detailViewConfig:
		return "config"
	default:
		return "logs"
	}
}

func (d detailViewMode) Next() detailViewMode {
	return (d + 1) % 3
}

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
	name.Prompt = ""
	name.Focus()
	name.CharLimit = 100

	exec := textinput.New()
	exec.Placeholder = "/path/to/command --args"
	exec.Prompt = ""
	exec.CharLimit = 500

	desc := textinput.New()
	desc.Placeholder = "Service description"
	desc.Prompt = ""
	desc.CharLimit = 200

	workdir := textinput.New()
	workdir.Placeholder = "~"
	workdir.Prompt = ""
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
	ctx           context.Context

	confirmingAction string
	confirmingUnit   string

	following       bool
	followPending   bool
	followSessionID uint64
	followCancel    context.CancelFunc
	followLogChan   <-chan string
	followLogLines  []string
	followTrimmed   bool

	groupMode      groupMode
	filterMode     filterMode
	detailViewMode detailViewMode

	createModal createModal
	showCreate  bool
	creating    bool

	activeBorder   lipgloss.Style
	inactiveBorder lipgloss.Style
	detailStyle    lipgloss.Style
	splitRatio     float64
}

func NewMainModel(client client.ServiceClient, cfg *config.Config) MainModel {
	theme := cfg.ThemeColors()
	ApplyTheme(theme)

	l := list.New(nil, itemDelegate{}, 0, 0)
	l.Title = "User Services"
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)

	// Override the default bubbles list title (purple bg, white fg) with a
	// theme-aware accent strip that fits the rest of the UI.
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.BorderActive)).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(theme.Border))
	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextDim)).
		Padding(0, 1)
	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextDim)).
		Padding(0, 1)

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
		ctx:            context.Background(),
		splitRatio:     0.33,
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(m.fetchServices, m.tick())
}

func (m *MainModel) recalcPanelSizes() {
	helpHeight := 2
	statusBarHeight := 1
	mainHeight := m.height - helpHeight - statusBarHeight

	listWidth := int(float64(m.width) * m.splitRatio)
	if listWidth < 20 {
		listWidth = 20
	}
	detailWidth := m.width - listWidth - 4
	if detailWidth < 20 {
		detailWidth = 20
	}

	m.list.SetSize(listWidth-2, mainHeight-2)
	m.viewport.Width = detailWidth - 2
	m.viewport.Height = mainHeight - 2
	m.help.Width = m.width
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcPanelSizes()

	case list.FilterMatchesMsg:
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case tea.MouseMsg:
		// Determine which pane the mouse is over and route the event.
		// Layout: list pane on the left of width listWidth, then detail pane.
		listWidth := int(float64(m.width) * m.splitRatio)
		if listWidth < 20 {
			listWidth = 20
		}
		// The list pane occupies columns [0, listWidth] roughly; clicks beyond
		// that fall in the detail pane.
		inDetailPane := msg.X >= listWidth

		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if inDetailPane {
				m.activeView = detailView
			} else {
				m.activeView = listView
			}
		}

		if inDetailPane {
			m.viewport, cmd = m.viewport.Update(msg)
		} else {
			var prevItem list.Item
			if m.list.SelectedItem() != nil {
				prevItem = m.list.SelectedItem()
			}
			m.list, cmd = m.list.Update(msg)
			if svc := m.getSelectedService(); svc != nil {
				currItem := m.list.SelectedItem()
				if prevItem == nil || currItem.FilterValue() != prevItem.FilterValue() {
					if m.following {
						m.stopFollow()
						m.statusMessage = "Stopped following (service changed)"
					} else if m.followPending {
						m.followSessionID++ // invalidate in-flight startFollow
						m.followPending = false
					}
					m.selectedSvc = svc.Name
					cmds = append(cmds, m.fetchDetailContent(svc.Name))
				}
			}
		}
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

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
		case key.Matches(msg, keys.ToggleDetail):
			m.detailViewMode = m.detailViewMode.Next()
			if m.following && m.detailViewMode != detailViewLogs {
				m.stopFollow()
			}
			m.statusMessage = fmt.Sprintf("Detail view: %s", m.detailViewMode)
			if svc := m.getSelectedService(); svc != nil {
				cmds = append(cmds, m.fetchDetailContent(svc.Name))
			}
			return m, tea.Batch(cmds...)
		case key.Matches(msg, keys.ToggleFollow):
			if m.following {
				m.stopFollow()
				m.statusMessage = "Stopped following logs"
			} else if !m.followPending {
				if svc := m.getSelectedService(); svc != nil {
					m.detailViewMode = detailViewLogs
					m.followSessionID++
					m.followPending = true
					cmd = m.startFollow(svc.Name, m.followSessionID)
					m.statusMessage = "Following logs for " + svc.Name
					return m, cmd
				}
			}
			return m, nil
		case key.Matches(msg, keys.ToggleGroup):
			m.groupMode = m.groupMode.Next()
			m.statusMessage = fmt.Sprintf("Group by: %s", m.groupMode)
			m.updateListTitle()
			cmds = append(cmds, m.list.SetItems(m.buildListItems()))

		case key.Matches(msg, keys.ToggleSource):
			m.filterMode = m.filterMode.Next()
			m.statusMessage = fmt.Sprintf("Filter: %s", m.filterMode)
			m.updateListTitle()
			cmds = append(cmds, m.list.SetItems(m.buildListItems()))

		case key.Matches(msg, keys.ShrinkPanel):
			m.splitRatio -= 0.05
			if m.splitRatio < 0.15 {
				m.splitRatio = 0.15
			}
			m.recalcPanelSizes()
			return m, nil

		case key.Matches(msg, keys.GrowPanel):
			m.splitRatio += 0.05
			if m.splitRatio > 0.70 {
				m.splitRatio = 0.70
			}
			m.recalcPanelSizes()
			return m, nil

		case key.Matches(msg, keys.Create):
			m.showCreate = true
			m.createModal = newCreateModal()
			return m, nil
		}

		if m.activeView == listView {
			switch {
			case key.Matches(msg, keys.Restart):
				if svc := m.getSelectedService(); svc != nil {
					m.confirmingAction = "restart"
					m.confirmingUnit = svc.Name
					m.statusMessage = fmt.Sprintf("Restart %s? (y/n)", svc.Name)
					return m, nil
				}
			case key.Matches(msg, keys.Start):
				if svc := m.getSelectedService(); svc != nil {
					m.statusMessage = "Starting " + svc.Name + "..."
					return m, m.startService(svc.Name)
				}
			case key.Matches(msg, keys.Stop):
				if svc := m.getSelectedService(); svc != nil {
					m.confirmingAction = "stop"
					m.confirmingUnit = svc.Name
					m.statusMessage = fmt.Sprintf("Stop %s? (y/n)", svc.Name)
					return m, nil
				}
			case key.Matches(msg, keys.Edit):
				if svc := m.getSelectedService(); svc != nil {
					return m, m.editService(svc.Name)
				}
			case key.Matches(msg, keys.Enable):
				if svc := m.getSelectedService(); svc != nil {
					m.confirmingAction = "enable"
					m.confirmingUnit = svc.Name
					m.statusMessage = fmt.Sprintf("Enable %s? (y/n)", svc.Name)
					return m, nil
				}
			case key.Matches(msg, keys.Disable):
				if svc := m.getSelectedService(); svc != nil {
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

				if svc := m.getSelectedService(); svc != nil {
					currItem := m.list.SelectedItem()
					if prevItem == nil || currItem.FilterValue() != prevItem.FilterValue() {
						if m.following {
							m.stopFollow()
							m.statusMessage = "Stopped following (service changed)"
						} else if m.followPending {
							m.followSessionID++ // invalidate in-flight startFollow
							m.followPending = false
						}
						m.selectedSvc = svc.Name
						cmds = append(cmds, m.fetchDetailContent(svc.Name))
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

		filtered := m.filteredServices()
		filteredCount := len(filtered)
		totalCount := len(m.services)
		if filteredCount != totalCount {
			m.list.Title = fmt.Sprintf("User Services (%d/%d)", filteredCount, totalCount)
		} else {
			m.list.Title = fmt.Sprintf("User Services (%d)", totalCount)
		}

		items := m.buildListItems()
		cmds = append(cmds, m.list.SetItems(items))

		if m.selectedSvc == "" && len(items) > 0 {
			for i, it := range items {
				if _, ok := it.(item); ok {
					m.list.Select(i)
					m.selectedSvc = it.FilterValue()
					break
				}
			}
		}

		if svc := m.getSelectedService(); svc != nil {
			m.selectedSvc = svc.Name
			cmds = append(cmds, m.fetchDetailContent(svc.Name))
		}

	case actionResultMsg:
		m.statusMessage = msg.message
		if msg.err != nil {
			m.statusMessage = "Error: " + renderUserError(msg.err)
		} else {
			cmds = append(cmds, m.fetchServices)
		}

	case createServiceResultMsg:
		m.creating = false
		if msg.err != nil {
			m.statusMessage = "Failed to create service: " + renderUserError(msg.err)
			m.showCreate = true
		} else {
			m.showCreate = false
			m.statusMessage = fmt.Sprintf("Created service: %s", msg.name)
			cmds = append(cmds, m.fetchServices)
		}

	case editorFinishedMsg:
		if msg.err != nil {
			m.statusMessage = "Edit failed: " + renderUserError(msg.err)
		} else {
			m.statusMessage = "Edit saved. Reloading daemon..."
			cmds = append(cmds, m.reloadDaemon())
		}
		return m, tea.Batch(cmds...)

	case reloadDaemonResultMsg:
		if msg.err != nil {
			m.statusMessage = "Edit saved, but reload failed: " + msg.err.Error()
		} else {
			m.statusMessage = "Edit saved. Reloaded daemon."
			cmds = append(cmds, m.fetchServices)
		}

	case logMsg:
		if msg.unit == m.selectedSvc && !m.following && m.detailViewMode == detailViewLogs {
			if svc := m.getSelectedService(); svc != nil {
				lines := m.config.General.LogLines
				if lines <= 0 {
					lines = 50
				}
				var body string
				banner := fmt.Sprintf("Last %d lines", lines)
				if msg.err != nil {
					banner = "Log error"
					body = lipgloss.NewStyle().
						Foreground(lipgloss.Color(m.theme.StatusFailed)).
						Render(renderUserError(msg.err))
				} else {
					body = styleLogContent(msg.logs, m.theme)
				}
				header := renderDetailHeader(m.theme,
					svc.Name, string(svc.Status), svc.Sub, string(svc.Source),
					svc.Description, banner, m.viewport.Width)
				content := header + "\n" + body
				m.viewport.SetContent(m.viewportWithScrollInfo(content))
			}
		}

	case detailContentMsg:
		if msg.unit == m.selectedSvc && msg.mode == m.detailViewMode {
			if m.following && m.detailViewMode == detailViewLogs {
				break
			}
			if msg.err != nil {
				m.statusMessage = fmt.Sprintf("Failed to load %s for %s: %s", msg.mode, msg.unit, renderUserError(msg.err))
				break
			}
			m.viewport.SetContent(m.viewportWithScrollInfo(msg.content))
		}

	case logLineMsg:
		if m.following && msg.id == m.followSessionID && m.selectedSvc != "" {
			line := msg.line
			maxLineBytes := 64 * 1024
			if len(line) > maxLineBytes {
				line = line[:maxLineBytes] + "... [truncated]"
				m.followTrimmed = true
			}

			m.followLogLines = append(m.followLogLines, line)
			maxLines := m.config.General.MaxFollowLines
			if maxLines <= 0 {
				maxLines = 1000
			}
			if len(m.followLogLines) > maxLines {
				m.followLogLines = m.followLogLines[len(m.followLogLines)-maxLines:]
				m.followTrimmed = true
			}
			if m.detailViewMode == detailViewLogs {
				svc := m.getSelectedService()
				if svc == nil {
					cmds = append(cmds, m.continueFollow(m.selectedSvc, msg.id))
					break
				}
				trimNotice := ""
				if m.followTrimmed {
					trimNotice = "\n" + lipgloss.NewStyle().
						Foreground(lipgloss.Color(m.theme.TextDim)).
						Italic(true).
						Render(fmt.Sprintf("buffer trimmed — showing last %d lines", maxLines)) + "\n"
				}
				styledFollowLogs := styleLogContent(strings.Join(m.followLogLines, "\n"), m.theme)

				// Header without the standard banner; followBanner replaces it.
				header := renderDetailHeader(m.theme,
					svc.Name, string(svc.Status), svc.Sub, string(svc.Source),
					svc.Description, "", m.viewport.Width)
				banner := followBanner(m.theme, m.viewport.Width)
				content := header + banner + "\n" + trimNotice + styledFollowLogs
				m.viewport.SetContent(content)
				m.viewport.GotoBottom()
			}
			cmds = append(cmds, m.continueFollow(m.selectedSvc, msg.id))
		}

	case followStartedMsg:
		m.followPending = false
		if msg.id != m.followSessionID {
			// Stale session (selection changed while follow was starting); cancel stream.
			msg.cancel()
			break
		}
		m.following = true
		m.followCancel = msg.cancel
		m.followLogChan = msg.ch
		m.followLogLines = nil
		m.followTrimmed = false
		m.selectedSvc = msg.unit
		cmds = append(cmds, m.continueFollow(msg.unit, msg.id))

	case followStoppedMsg:
		if m.following && msg.id == m.followSessionID && msg.unit == m.selectedSvc {
			m.stopFollow()
			m.statusMessage = "Stopped following logs"
			if m.selectedSvc != "" {
				cmds = append(cmds, m.fetchDetailContent(m.selectedSvc))
			}
		}

	case errMsg:
		m.followPending = false
		m.statusMessage = fmt.Sprintf("Error: %s - %s", msg.op, renderUserError(msg.err))

	case tickMsg:
		// Don't refresh while a filter is active or being entered — it would
		// reset either the in-progress filter input or the applied filter result.
		if m.list.FilterState() != list.Unfiltered {
			return m, m.tick()
		}
		return m, tea.Batch(m.fetchServices, m.tick())
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) buildListItems() []list.Item {
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

	return items
}

func (m *MainModel) updateListTitle() {
	filtered := m.filteredServices()
	filteredCount := len(filtered)
	totalCount := len(m.services)
	if filteredCount != totalCount {
		m.list.Title = fmt.Sprintf("User Services (%d/%d)", filteredCount, totalCount)
	} else {
		m.list.Title = fmt.Sprintf("User Services (%d)", totalCount)
	}
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
			if svc.Source != client.SourceStatic && svc.Source != client.SourceGenerated && svc.Source != client.SourceTransient && svc.Source != client.SourceSystem {
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
	if m.creating {
		return m, nil
	}

	name := strings.TrimSpace(m.createModal.nameInput.Value())
	exec := strings.TrimSpace(m.createModal.execInput.Value())

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

	m.creating = true
	m.statusMessage = fmt.Sprintf("Creating service: %s...", name)
	return m, m.createService(tmpl)
}

func renderUserError(err error) string {
	if err == nil {
		return ""
	}

	var svcErr *client.ServiceError
	if errors.As(err, &svcErr) {
		return client.UserErrorMessage(svcErr)
	}

	return err.Error()
}
