package service

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"systemd-tui/internal/client"
)

// systemdClient implements ServiceClient for systemd user services
type systemdClient struct{}

func NewSystemdClient() client.ServiceClient {
	return &systemdClient{}
}

func (c *systemdClient) ListServices() ([]client.Service, error) {
	cmd := exec.Command("systemctl", "--user", "list-units", "--type=service", "--all", "--output=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run systemctl: %w", err)
	}

	var units []Unit
	if err := json.Unmarshal(output, &units); err != nil {
		return nil, fmt.Errorf("failed to parse systemctl output: %w", err)
	}

	services := make([]client.Service, len(units))
	for i, u := range units {
		services[i] = c.unitToService(u)
	}

	return services, nil
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

func (c *systemdClient) GetConfig(name string) (string, error) {
	cmd := exec.Command("systemctl", "--user", "cat", name)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}
	return string(output), nil
}

func (c *systemdClient) EditService(name string) (*exec.Cmd, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command("systemctl", "--user", "edit", "--full", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "SYSTEMD_EDITOR="+editor)

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

func (c *systemdClient) unitToService(u Unit) client.Service {
	enabled := u.Load == "loaded" && (u.Active == "active" || strings.Contains(u.Sub, "enabled"))
	return client.Service{
		Name:        u.Unit,
		Description: u.Description,
		Status:      client.ServiceStatus(u.Active),
		Sub:         u.Sub,
		Enabled:     enabled,
		Load:        u.Load,
	}
}
