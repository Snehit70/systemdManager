package service

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type SystemdClient struct{}

func NewSystemdClient() *SystemdClient {
	return &SystemdClient{}
}

func (c *SystemdClient) ListUnits() ([]Unit, error) {
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

func (c *SystemdClient) StartUnit(unit string) error {
	return c.runAction("start", unit)
}

func (c *SystemdClient) StopUnit(unit string) error {
	return c.runAction("stop", unit)
}

func (c *SystemdClient) RestartUnit(unit string) error {
	return c.runAction("restart", unit)
}

func (c *SystemdClient) EnableUnit(unit string) error {
	return c.runAction("enable", unit)
}

func (c *SystemdClient) DisableUnit(unit string) error {
	return c.runAction("disable", unit)
}

func (c *SystemdClient) EditCmd(unit string) *exec.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command("systemctl", "--user", "edit", "--full", unit)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "SYSTEMD_EDITOR="+editor)

	return cmd
}

func (c *SystemdClient) ReloadDaemon() error {
	cmd := exec.Command("systemctl", "--user", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to daemon-reload: %w", err)
	}
	return nil
}

func (c *SystemdClient) GetLogs(unit string) (string, error) {
	cmd := exec.Command("journalctl", "--user", "-u", unit, "-n", "50", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}
	return string(output), nil
}

func (c *SystemdClient) runAction(action, unit string) error {
	cmd := exec.Command("systemctl", "--user", action, unit)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to %s %s: %w", action, unit, err)
	}
	return nil
}
