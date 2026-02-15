package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"systemd-tui/internal/client"
)

type systemdClient struct {
	editor        string
	userConfigDir string
}

func NewSystemdClient(cfg interface{ GetEditor() string }) client.ServiceClient {
	editor := "vim"
	if cfg != nil {
		editor = cfg.GetEditor()
	}
	userConfigDir, _ := os.UserConfigDir()
	return &systemdClient{editor: editor, userConfigDir: userConfigDir}
}

func (c *systemdClient) ListServices() ([]client.Service, error) {
	units, err := c.listUnits()
	if err != nil {
		return nil, err
	}

	unitFiles, err := c.listUnitFiles()
	if err != nil {
		return nil, err
	}

	unitFileStates := make(map[string]string)
	for _, uf := range unitFiles {
		unitFileStates[uf.UnitFile] = uf.State
	}

	services := make([]client.Service, len(units))
	for i, u := range units {
		services[i] = c.unitToService(u, unitFileStates[u.Unit])
	}

	return services, nil
}

func (c *systemdClient) listUnits() ([]Unit, error) {
	cmd := exec.Command("systemctl", "--user", "list-units", "--type=service", "--all", "--output=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run systemctl: %w", err)
	}

	var units []Unit
	if err := json.Unmarshal(output, &units); err != nil {
		return nil, fmt.Errorf("failed to parse systemctl output: %w", err)
	}

	return units, nil
}

func (c *systemdClient) listUnitFiles() ([]UnitFile, error) {
	cmd := exec.Command("systemctl", "--user", "list-unit-files", "--type=service", "--output=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run systemctl list-unit-files: %w", err)
	}

	var unitFiles []UnitFile
	if err := json.Unmarshal(output, &unitFiles); err != nil {
		return nil, fmt.Errorf("failed to parse systemctl output: %w", err)
	}

	return unitFiles, nil
}

func (c *systemdClient) StartService(name string) error {
	return c.runAction("start", name)
}

func (c *systemdClient) StopService(name string) error {
	return c.runAction("stop", name)
}

func (c *systemdClient) RestartService(name string) error {
	return c.runAction("restart", name)
}

func (c *systemdClient) EnableService(name string) error {
	return c.runAction("enable", name)
}

func (c *systemdClient) DisableService(name string) error {
	return c.runAction("disable", name)
}

func (c *systemdClient) GetStatus(name string) (client.ServiceStatus, error) {
	cmd := exec.Command("systemctl", "--user", "show", name, "--property=ActiveState", "--value")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	state := client.ServiceStatus(strings.TrimSpace(string(output)))
	return state, nil
}

func (c *systemdClient) GetLogs(name string, opts client.LogOptions) (string, error) {
	lines := opts.Lines
	if lines <= 0 {
		lines = 50
	}

	args := []string{"--user", "-u", name, "-n", fmt.Sprintf("%d", lines), "--no-pager"}
	if opts.Follow {
		return "", fmt.Errorf("follow mode not yet supported")
	}

	cmd := exec.Command("journalctl", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	return string(output), nil
}

func (c *systemdClient) FollowLogs(name string, opts client.LogOptions) (<-chan string, context.CancelFunc, error) {
	lines := opts.Lines
	if lines <= 0 {
		lines = 50
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "journalctl", "--user", "-u", name, "-n", fmt.Sprintf("%d", lines), "-f", "--no-pager")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to start journalctl: %w", err)
	}

	logChan := make(chan string, 100)

	go func() {
		defer close(logChan)
		defer cmd.Wait()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case logChan <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	return logChan, cancel, nil
}

func (c *systemdClient) GetConfig(name string) (string, error) {
	cmd := exec.Command("systemctl", "--user", "cat", name)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}
	return string(output), nil
}

func (c *systemdClient) EditService(name string) (*exec.Cmd, error) {
	cmd := exec.Command("systemctl", "--user", "edit", "--full", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "SYSTEMD_EDITOR="+c.editor)

	return cmd, nil
}

func (c *systemdClient) ReloadDaemon() error {
	cmd := exec.Command("systemctl", "--user", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to daemon-reload: %w", err)
	}
	return nil
}

func (c *systemdClient) runAction(action, unit string) error {
	cmd := exec.Command("systemctl", "--user", action, unit)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to %s %s: %w", action, unit, err)
	}
	return nil
}

func (c *systemdClient) unitToService(u Unit, state string) client.Service {
	enabled := u.Load == "loaded" && (u.Active == "active" || strings.Contains(u.Sub, "enabled"))
	source := c.determineSource(u.Unit, state)

	return client.Service{
		Name:        u.Unit,
		Description: u.Description,
		Status:      client.ServiceStatus(u.Active),
		Sub:         u.Sub,
		Enabled:     enabled,
		Load:        u.Load,
		Source:      source,
	}
}

func (c *systemdClient) determineSource(unitName, state string) client.ServiceSource {
	if c.userConfigDir != "" {
		userServicePath := filepath.Join(c.userConfigDir, "systemd", "user", unitName)
		if _, err := os.Stat(userServicePath); err == nil {
			return client.SourceUser
		}
	}

	switch state {
	case "generated":
		return client.SourceGenerated
	case "transient":
		return client.SourceTransient
	case "static", "alias":
		return client.SourceStatic
	case "enabled", "disabled":
		return client.SourceSystem
	default:
		return client.SourceUnknown
	}
}

func (c *systemdClient) CreateService(tmpl client.ServiceTemplate) error {
	if !strings.HasSuffix(tmpl.Name, ".service") {
		tmpl.Name += ".service"
	}

	if c.userConfigDir == "" {
		return fmt.Errorf("cannot determine user config directory")
	}

	serviceDir := filepath.Join(c.userConfigDir, "systemd", "user")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create service directory: %w", err)
	}

	servicePath := filepath.Join(serviceDir, tmpl.Name)
	if _, err := os.Stat(servicePath); err == nil {
		return fmt.Errorf("service %s already exists", tmpl.Name)
	}

	content := c.generateServiceFile(tmpl)

	if err := os.WriteFile(servicePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	return c.ReloadDaemon()
}

func (c *systemdClient) generateServiceFile(tmpl client.ServiceTemplate) string {
	var sb strings.Builder

	sb.WriteString("[Unit]\n")
	sb.WriteString(fmt.Sprintf("Description=%s\n", tmpl.Description))
	sb.WriteString("After=network.target\n\n")

	sb.WriteString("[Service]\n")
	sb.WriteString(fmt.Sprintf("Type=%s\n", c.serviceType(tmpl.Type)))
	sb.WriteString(fmt.Sprintf("ExecStart=%s\n", tmpl.ExecStart))

	if tmpl.WorkingDirectory != "" {
		sb.WriteString(fmt.Sprintf("WorkingDirectory=%s\n", tmpl.WorkingDirectory))
	}

	sb.WriteString(fmt.Sprintf("Restart=%s\n", c.restartPolicy(tmpl.Restart)))
	sb.WriteString("RestartSec=5\n\n")

	sb.WriteString("[Install]\n")
	sb.WriteString("WantedBy=default.target\n")

	return sb.String()
}

func (c *systemdClient) serviceType(t string) string {
	switch t {
	case "oneshot", "forking", "notify", "dbus":
		return t
	default:
		return "simple"
	}
}

func (c *systemdClient) restartPolicy(r string) string {
	switch r {
	case "always", "on-success", "on-failure", "on-abnormal", "on-abort", "on-watchdog":
		return r
	case "no":
		return "no"
	default:
		return "on-failure"
	}
}
