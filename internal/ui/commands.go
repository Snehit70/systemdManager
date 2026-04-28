package ui

import (
	"fmt"
	"time"

	"systemd-tui/internal/client"

	tea "github.com/charmbracelet/bubbletea"
)

// Message types for the Update loop.

type errMsg struct {
	op  string
	err error
}

func (e errMsg) Error() string { return e.err.Error() }

type actionResultMsg struct {
	message string
	err     error
}

type createServiceResultMsg struct {
	name string
	err  error
}

type editorFinishedMsg struct {
	err error
}

type logMsg struct {
	unit string
	logs string
	err  error
}

type detailContentMsg struct {
	unit    string
	content string
}

type logLineMsg struct {
	line string
}

type tickMsg time.Time

// Async command builders.

func (m MainModel) tick() tea.Cmd {
	interval := m.config.General.RefreshInterval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m MainModel) fetchServices() tea.Msg {
	services, err := m.client.ListServices(m.ctx)
	if err != nil {
		return errMsg{op: "Failed to list services", err: err}
	}
	return services
}

func (m MainModel) fetchLogs(unit string) tea.Cmd {
	return func() tea.Msg {
		lines := m.config.General.LogLines
		if lines <= 0 {
			lines = 50
		}
		logs, err := m.client.GetLogs(m.ctx, unit, client.LogOptions{Lines: lines})
		return logMsg{unit: unit, logs: logs, err: err}
	}
}

func (m MainModel) fetchDetailContent(unit string) tea.Cmd {
	return func() tea.Msg {
		var content string
		var err error

		switch m.detailViewMode {
		case detailViewStatus:
			content, err = m.client.GetStatusDetails(m.ctx, unit)
			if err != nil {
				content = "Error fetching status: " + err.Error()
			}
		case detailViewConfig:
			content, err = m.client.GetConfig(m.ctx, unit)
			if err != nil {
				content = "Error fetching config: " + err.Error()
			}
		default:
			lines := m.config.General.LogLines
			if lines <= 0 {
				lines = 50
			}
			content, err = m.client.GetLogs(m.ctx, unit, client.LogOptions{Lines: lines})
			if err != nil {
				content = "Error fetching logs: " + err.Error()
			} else {
				svc := m.selectedServiceByName(unit)
				if svc != nil {
					content = fmt.Sprintf(
						"Service: %s\nStatus: %s (%s)\nDescription: %s\n\n[Last %d Lines of Log]\n%s",
						svc.Name, svc.Status, svc.Sub, svc.Description, lines, content,
					)
				}
			}
		}
		return detailContentMsg{unit: unit, content: content}
	}
}

func (m MainModel) selectedServiceByName(name string) *client.Service {
	for i := range m.services {
		if m.services[i].Name == name {
			return &m.services[i]
		}
	}
	return nil
}

// Service action commands.

func (m MainModel) startService(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.StartService(m.ctx, unit)
		return actionResultMsg{message: "Started " + unit, err: err}
	}
}

func (m MainModel) stopService(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.StopService(m.ctx, unit)
		return actionResultMsg{message: "Stopped " + unit, err: err}
	}
}

func (m MainModel) restartService(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.RestartService(m.ctx, unit)
		return actionResultMsg{message: "Restarted " + unit, err: err}
	}
}

func (m MainModel) enableService(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.EnableService(m.ctx, unit)
		return actionResultMsg{message: "Enabled " + unit, err: err}
	}
}

func (m MainModel) disableService(unit string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.DisableService(m.ctx, unit)
		return actionResultMsg{message: "Disabled " + unit, err: err}
	}
}

func (m MainModel) createService(tmpl client.ServiceTemplate) tea.Cmd {
	return func() tea.Msg {
		err := m.client.CreateService(m.ctx, tmpl)
		return createServiceResultMsg{name: tmpl.Name, err: err}
	}
}

func (m MainModel) editService(unit string) tea.Cmd {
	cmd, err := m.client.EditService(m.ctx, unit)
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

// Follow mode lifecycle.

func (m *MainModel) startFollow(unit string) tea.Cmd {
	lines := m.config.General.LogLines
	if lines <= 0 {
		lines = 50
	}
	logChan, cancel, err := m.client.FollowLogs(m.ctx, unit, client.LogOptions{Lines: lines})
	if err != nil {
		return func() tea.Msg {
			return errMsg{op: "Failed to follow logs", err: err}
		}
	}

	m.following = true
	m.followCancel = cancel
	m.followLogChan = logChan
	m.followLogLines = nil
	m.followTrimmed = false

	return func() tea.Msg {
		select {
		case line, ok := <-logChan:
			if !ok {
				m.stopFollow()
				return nil
			}
			return logLineMsg{line: line}
		case <-m.ctx.Done():
			m.stopFollow()
			return nil
		}
	}
}

func (m *MainModel) stopFollow() {
	if m.followCancel != nil {
		m.followCancel()
	}
	m.following = false
	m.followCancel = nil
	m.followLogChan = nil
	m.followLogLines = nil
	m.followTrimmed = false
}

func (m *MainModel) continueFollow(unit string) tea.Cmd {
	if !m.following || m.followLogChan == nil {
		return nil
	}

	return func() tea.Msg {
		select {
		case line, ok := <-m.followLogChan:
			if !ok {
				m.stopFollow()
				return nil
			}
			return logLineMsg{line: line}
		case <-m.ctx.Done():
			m.stopFollow()
			return nil
		}
	}
}
