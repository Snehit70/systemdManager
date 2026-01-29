package service

import (
	"encoding/json"
	"fmt"
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

func (c *SystemdClient) runAction(action, unit string) error {
	cmd := exec.Command("systemctl", "--user", action, unit)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to %s %s: %w", action, unit, err)
	}
	return nil
}
