package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"systemd-tui/internal/client"
)

func TestIsUnitFileEnabled(t *testing.T) {
	tests := []struct {
		state string
		want  bool
	}{
		{state: "enabled", want: true},
		{state: "enabled-runtime", want: true},
		{state: "disabled", want: false},
		{state: "static", want: false},
		{state: "indirect", want: false},
		{state: "generated", want: false},
		{state: "transient", want: false},
		{state: "alias", want: false},
		{state: "linked", want: false},
		{state: "linked-runtime", want: false},
		{state: "masked", want: false},
		{state: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if got := isUnitFileEnabled(tt.state); got != tt.want {
				t.Fatalf("isUnitFileEnabled(%q) = %t, want %t", tt.state, got, tt.want)
			}
		})
	}
}

func TestCreateServiceWritesFileAndReloadsDaemon(t *testing.T) {
	callLog := installFakeSystemctl(t, 0)
	configDir := t.TempDir()
	c := &systemdClient{userConfigDir: configDir}

	tmpl := client.ServiceTemplate{
		Name:        "demo",
		Description: "Demo service",
		ExecStart:   "/bin/true",
		Type:        "simple",
		Restart:     "on-failure",
	}

	if err := c.CreateService(context.Background(), tmpl); err != nil {
		t.Fatalf("CreateService() error = %v", err)
	}

	servicePath := filepath.Join(configDir, "systemd", "user", "demo.service")
	data, err := os.ReadFile(servicePath)
	if err != nil {
		t.Fatalf("read created service: %v", err)
	}
	if !strings.Contains(string(data), "ExecStart=/bin/true") {
		t.Fatalf("created service missing ExecStart: %q", data)
	}

	calls := readCallLog(t, callLog)
	if len(calls) != 1 || calls[0] != "--user daemon-reload" {
		t.Fatalf("systemctl calls = %q, want one daemon-reload", calls)
	}
}

func TestCreateServiceRollsBackWhenDaemonReloadFails(t *testing.T) {
	callLog := installFakeSystemctl(t, 1)
	configDir := t.TempDir()
	c := &systemdClient{userConfigDir: configDir}

	err := c.CreateService(context.Background(), client.ServiceTemplate{
		Name:      "demo",
		ExecStart: "/bin/true",
		Type:      "simple",
		Restart:   "on-failure",
	})
	if err == nil {
		t.Fatal("CreateService() error = nil, want daemon-reload failure")
	}

	servicePath := filepath.Join(configDir, "systemd", "user", "demo.service")
	if _, statErr := os.Stat(servicePath); !os.IsNotExist(statErr) {
		t.Fatalf("service file remains after failed reload; stat error = %v", statErr)
	}

	calls := readCallLog(t, callLog)
	if len(calls) != 2 {
		t.Fatalf("systemctl calls = %q, want initial reload and rollback reload", calls)
	}
}

func TestCreateServiceDoesNotOverwriteExistingFile(t *testing.T) {
	installFakeSystemctl(t, 0)
	configDir := t.TempDir()
	serviceDir := filepath.Join(configDir, "systemd", "user")
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		t.Fatalf("create service directory: %v", err)
	}

	servicePath := filepath.Join(serviceDir, "demo.service")
	const original = "original content\n"
	if err := os.WriteFile(servicePath, []byte(original), 0o644); err != nil {
		t.Fatalf("write existing service: %v", err)
	}

	c := &systemdClient{userConfigDir: configDir}
	err := c.CreateService(context.Background(), client.ServiceTemplate{
		Name:      "demo",
		ExecStart: "/bin/true",
	})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("CreateService() error = %v, want already exists", err)
	}

	data, readErr := os.ReadFile(servicePath)
	if readErr != nil {
		t.Fatalf("read existing service: %v", readErr)
	}
	if string(data) != original {
		t.Fatalf("existing service was modified: %q", data)
	}
}

func installFakeSystemctl(t *testing.T, exitCode int) string {
	t.Helper()

	binDir := t.TempDir()
	callLog := filepath.Join(t.TempDir(), "systemctl-calls.log")
	scriptPath := filepath.Join(binDir, "systemctl")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$*\" >> \"$SYSTEMD_TUI_TEST_LOG\"\nexit %d\n", exitCode)
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake systemctl: %v", err)
	}

	t.Setenv("SYSTEMD_TUI_TEST_LOG", callLog)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return callLog
}

func readCallLog(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read systemctl call log: %v", err)
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}
