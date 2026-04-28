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

// Compile-time interface verification
var _ client.ServiceClient = (*systemdClient)(nil)

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

func (c *systemdClient) ListServices(ctx context.Context) ([]client.Service, error) {
	units, err := c.listUnits(ctx)
	if err != nil {
		return nil, err
	}

	unitFileStates := make(map[string]string)
	unitFiles, err := c.listUnitFiles(ctx)
	if err == nil {
		for _, uf := range unitFiles {
			unitFileStates[uf.UnitFile] = uf.State
		}
	}

	services := make([]client.Service, len(units))
	for i, u := range units {
		services[i] = c.unitToService(u, unitFileStates[u.Unit])
	}

	return services, nil
}

func (c *systemdClient) listUnits(ctx context.Context) ([]Unit, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", "list-units", "--type=service", "--all", "--output=json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, commandError("failed to run systemctl", err, output)
	}

	var units []Unit
	if err := json.Unmarshal(output, &units); err != nil {
		return nil, fmt.Errorf("failed to parse systemctl output: %w", err)
	}

	return units, nil
}

func (c *systemdClient) listUnitFiles(ctx context.Context) ([]UnitFile, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", "list-unit-files", "--type=service", "--output=json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, commandError("failed to run systemctl list-unit-files", err, output)
	}

	var unitFiles []UnitFile
	if err := json.Unmarshal(output, &unitFiles); err != nil {
		return nil, fmt.Errorf("failed to parse systemctl output: %w", err)
	}

	return unitFiles, nil
}

func (c *systemdClient) StartService(ctx context.Context, name string) error {
	return c.runAction(ctx, "start", name)
}

func (c *systemdClient) StopService(ctx context.Context, name string) error {
	return c.runAction(ctx, "stop", name)
}

func (c *systemdClient) RestartService(ctx context.Context, name string) error {
	return c.runAction(ctx, "restart", name)
}

func (c *systemdClient) EnableService(ctx context.Context, name string) error {
	return c.runAction(ctx, "enable", name)
}

func (c *systemdClient) DisableService(ctx context.Context, name string) error {
	return c.runAction(ctx, "disable", name)
}

func (c *systemdClient) GetStatus(ctx context.Context, name string) (client.ServiceStatus, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", "show", name, "--property=ActiveState", "--value")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", commandError(fmt.Sprintf("failed to get status for %s", name), err, output)
	}

	state := client.ServiceStatus(strings.TrimSpace(string(output)))
	return state, nil
}

func (c *systemdClient) GetLogs(ctx context.Context, name string, opts client.LogOptions) (string, error) {
	lines := opts.Lines
	if lines <= 0 {
		lines = 50
	}

	args := []string{"--user", "-u", name, "-n", fmt.Sprintf("%d", lines), "--no-pager"}
	if opts.Follow {
		return "", fmt.Errorf("follow mode not yet supported")
	}

	cmd := exec.CommandContext(ctx, "journalctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", commandError(fmt.Sprintf("failed to get logs for %s", name), err, output)
	}

	return string(output), nil
}

func (c *systemdClient) FollowLogs(ctx context.Context, name string, opts client.LogOptions) (<-chan string, context.CancelFunc, error) {
	lines := opts.Lines
	if lines <= 0 {
		lines = 50
	}

	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, "journalctl", "--user", "-u", name, "-n", fmt.Sprintf("%d", lines), "-f", "--no-pager")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to create pipe for %s: %w", name, err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to start journalctl for %s: %w", name, err)
	}

	logChan := make(chan string, 100)

	go func() {
		defer close(logChan)
		defer cmd.Wait()

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
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

func (c *systemdClient) GetConfig(ctx context.Context, name string) (string, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", "cat", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", commandError(fmt.Sprintf("failed to get config for %s", name), err, output)
	}
	return string(output), nil
}

func (c *systemdClient) GetStatusDetails(ctx context.Context, name string) (string, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", "status", name, "--no-pager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return string(output), nil
		}
		return "", commandError(fmt.Sprintf("failed to get status details for %s", name), err, output)
	}
	return string(output), nil
}

func (c *systemdClient) EditService(ctx context.Context, name string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", "edit", "--full", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "SYSTEMD_EDITOR="+c.editor)

	return cmd, nil
}

func (c *systemdClient) ReloadDaemon(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to daemon-reload: %w", err)
	}
	return nil
}

func (c *systemdClient) runAction(ctx context.Context, action, unit string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", action, unit)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return client.NewServiceError(action, unit, commandError("systemctl "+action, err, output))
	}
	return nil
}

func commandError(op string, err error, output []byte) error {
	details := strings.TrimSpace(string(output))
	if details == "" {
		return fmt.Errorf("%s: %w", op, err)
	}
	return fmt.Errorf("%s: %w: %s", op, err, details)
}

func (c *systemdClient) unitToService(u Unit, state string) client.Service {
	enabled := isUnitFileEnabled(state)
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

func isUnitFileEnabled(state string) bool {
	switch state {
	case "enabled", "enabled-runtime", "static", "indirect", "generated":
		return true
	default:
		return false
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

func (c *systemdClient) CreateService(ctx context.Context, tmpl client.ServiceTemplate) error {
	tmpl.Name = c.sanitizeServiceName(tmpl.Name)
	if tmpl.Name == "" {
		return fmt.Errorf("invalid service name")
	}

	if c.userConfigDir == "" {
		return fmt.Errorf("cannot determine user config directory")
	}

	serviceDir := filepath.Join(c.userConfigDir, "systemd", "user")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create service directory: %w", err)
	}

	servicePath := filepath.Join(serviceDir, tmpl.Name)

	if err := c.sanitizeTemplate(&tmpl); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	content := c.generateServiceFile(tmpl)

	file, err := os.OpenFile(servicePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("service %s already exists", tmpl.Name)
		}
		return fmt.Errorf("failed to create service file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	return c.ReloadDaemon(ctx)
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

func (c *systemdClient) sanitizeServiceName(name string) string {
	name = filepath.Base(name)
	if strings.Contains(name, "..") {
		return ""
	}

	if !strings.HasSuffix(name, ".service") {
		name += ".service"
	}

	return name
}

func (c *systemdClient) sanitizeTemplate(tmpl *client.ServiceTemplate) error {
	if strings.ContainsAny(tmpl.Description, "\n\r") {
		return fmt.Errorf("description contains newlines")
	}
	if strings.ContainsAny(tmpl.ExecStart, "\n\r") {
		return fmt.Errorf("exec start contains newlines")
	}
	if strings.ContainsAny(tmpl.WorkingDirectory, "\n\r") {
		return fmt.Errorf("working directory contains newlines")
	}
	if strings.ContainsAny(tmpl.Type, "\n\r") {
		return fmt.Errorf("type contains newlines")
	}
	if strings.ContainsAny(tmpl.Restart, "\n\r") {
		return fmt.Errorf("restart contains newlines")
	}

	tmpl.Description = strings.TrimSpace(tmpl.Description)
	tmpl.ExecStart = strings.TrimSpace(tmpl.ExecStart)
	tmpl.WorkingDirectory = strings.TrimSpace(tmpl.WorkingDirectory)
	tmpl.Type = strings.TrimSpace(tmpl.Type)
	tmpl.Restart = strings.TrimSpace(tmpl.Restart)

	return nil
}
