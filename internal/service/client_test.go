package service

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"testing"

	"systemd-tui/internal/client"
)

func TestUnitJSONParsing(t *testing.T) {
	jsonData := `{
		"unit": "test.service",
		"load": "loaded",
		"active": "active",
		"sub": "running",
		"description": "Test Service"
	}`

	var unit Unit
	err := json.Unmarshal([]byte(jsonData), &unit)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if unit.Unit != "test.service" {
		t.Errorf("Expected unit 'test.service', got '%s'", unit.Unit)
	}
	if unit.Load != "loaded" {
		t.Errorf("Expected load 'loaded', got '%s'", unit.Load)
	}
	if unit.Active != "active" {
		t.Errorf("Expected active 'active', got '%s'", unit.Active)
	}
	if unit.Sub != "running" {
		t.Errorf("Expected sub 'running', got '%s'", unit.Sub)
	}
	if unit.Description != "Test Service" {
		t.Errorf("Expected description 'Test Service', got '%s'", unit.Description)
	}
}

func TestMultipleUnitsJSONParsing(t *testing.T) {
	jsonData := `[
		{
			"unit": "service1.service",
			"load": "loaded",
			"active": "active",
			"sub": "running",
			"description": "Service 1"
		},
		{
			"unit": "service2.service",
			"load": "loaded",
			"active": "failed",
			"sub": "dead",
			"description": "Service 2"
		},
		{
			"unit": "service3.service",
			"load": "loaded",
			"active": "inactive",
			"sub": "dead",
			"description": "Service 3"
		}
	]`

	var units []Unit
	err := json.Unmarshal([]byte(jsonData), &units)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(units) != 3 {
		t.Fatalf("Expected 3 units, got %d", len(units))
	}

	if units[0].Unit != "service1.service" {
		t.Errorf("Expected first unit 'service1.service', got '%s'", units[0].Unit)
	}
	if units[0].Active != "active" {
		t.Errorf("Expected first unit active 'active', got '%s'", units[0].Active)
	}

	if units[1].Unit != "service2.service" {
		t.Errorf("Expected second unit 'service2.service', got '%s'", units[1].Unit)
	}
	if units[1].Active != "failed" {
		t.Errorf("Expected second unit active 'failed', got '%s'", units[1].Active)
	}

	if units[2].Unit != "service3.service" {
		t.Errorf("Expected third unit 'service3.service', got '%s'", units[2].Unit)
	}
	if units[2].Active != "inactive" {
		t.Errorf("Expected third unit active 'inactive', got '%s'", units[2].Active)
	}
}

func TestUnitJSONParsingWithMissingFields(t *testing.T) {
	jsonData := `{
		"unit": "test.service",
		"load": "loaded"
	}`

	var unit Unit
	err := json.Unmarshal([]byte(jsonData), &unit)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if unit.Unit != "test.service" {
		t.Errorf("Expected unit 'test.service', got '%s'", unit.Unit)
	}
	if unit.Active != "" {
		t.Errorf("Expected empty active field, got '%s'", unit.Active)
	}
	if unit.Description != "" {
		t.Errorf("Expected empty description, got '%s'", unit.Description)
	}
}

func TestUnitJSONParsingInvalidJSON(t *testing.T) {
	jsonData := `{invalid json}`

	var unit Unit
	err := json.Unmarshal([]byte(jsonData), &unit)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// Interface compliance test
var _ client.ServiceClient = (*systemdClient)(nil)

// Mock implementations for testing
type mockSystemdClient struct{}

func (m *mockSystemdClient) ListServices() ([]client.Service, error) {
	return []client.Service{
		{
			Name:        "test.service",
			Description: "Test Service",
			Status:      "active",
			Sub:         "running",
			Enabled:     true,
			Load:        "loaded",
		},
	}, nil
}

func (m *mockSystemdClient) StartService(name string) error {
	if name == "error.service" {
		return errors.New("failed to start")
	}
	return nil
}

func (m *mockSystemdClient) StopService(name string) error {
	if name == "error.service" {
		return errors.New("failed to stop")
	}
	return nil
}

func (m *mockSystemdClient) RestartService(name string) error {
	if name == "error.service" {
		return errors.New("failed to restart")
	}
	return nil
}

func (m *mockSystemdClient) EnableService(name string) error {
	if name == "error.service" {
		return errors.New("failed to enable")
	}
	return nil
}

func (m *mockSystemdClient) DisableService(name string) error {
	if name == "error.service" {
		return errors.New("failed to disable")
	}
	return nil
}

func (m *mockSystemdClient) GetStatus(name string) (client.ServiceStatus, error) {
	if name == "error.service" {
		return "", errors.New("failed to get status")
	}
	return "active", nil
}

func (m *mockSystemdClient) GetLogs(name string, opts client.LogOptions) (string, error) {
	if name == "error.service" {
		return "", errors.New("failed to get logs")
	}
	return "Mock log output", nil
}

func (m *mockSystemdClient) GetConfig(name string) (string, error) {
	if name == "error.service" {
		return "", errors.New("failed to get config")
	}
	return "[Service]\nExecStart=/bin/true", nil
}

func (m *mockSystemdClient) EditService(name string) (*exec.Cmd, error) {
	if name == "error.service" {
		return nil, errors.New("failed to edit")
	}
	// Return a command that does nothing for tests
	cmd := exec.Command("true")
	return cmd, nil
}

func (m *mockSystemdClient) ReloadDaemon() error {
	if os.Getenv("SYSTEMD_TUI_TEST_FAIL_RELOAD") == "1" {
		return errors.New("failed to reload daemon")
	}
	return nil
}

// Tests for systemdClient implementation
func TestSystemdClient_ListServices(t *testing.T) {
	// This test would need systemctl available - skip in environments without systemd
	// For now, we trust the interface compliance and manual testing
	t.Skip("Requires systemd environment")
}

func TestSystemdClient_GetStatus(t *testing.T) {
	t.Skip("Requires systemd environment")
}

func TestSystemdClient_GetLogs(t *testing.T) {
	t.Skip("Requires systemd environment")
}

func TestSystemdClient_GetConfig(t *testing.T) {
	t.Skip("Requires systemd environment")
}

func TestSystemdClient_ReloadDaemon(t *testing.T) {
	t.Skip("Requires systemd environment")
}

// Mock tests - these verify the interface can be implemented
func TestMockSystemdClient_ListServices(t *testing.T) {
	mock := &mockSystemdClient{}
	services, err := mock.ListServices()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
	if services[0].Name != "test.service" {
		t.Errorf("Expected service name 'test.service', got '%s'", services[0].Name)
	}
}

func TestMockSystemdClient_Actions(t *testing.T) {
	mock := &mockSystemdClient{}

	// Test successful actions
	for _, action := range []func(string) error{
		mock.StartService,
		mock.StopService,
		mock.RestartService,
		mock.EnableService,
		mock.DisableService,
	} {
		if err := action("test.service"); err != nil {
			t.Errorf("Action %T failed: %v", action, err)
		}
	}

	// Test error case
	if err := mock.StartService("error.service"); err == nil {
		t.Error("Expected error for error.service")
	}
}

func TestMockSystemdClient_GetStatus(t *testing.T) {
	mock := &mockSystemdClient{}

	status, err := mock.GetStatus("test.service")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if status != "active" {
		t.Errorf("Expected status 'active', got '%s'", status)
	}

	_, err = mock.GetStatus("error.service")
	if err == nil {
		t.Error("Expected error for error.service")
	}
}

func TestMockSystemdClient_GetLogs(t *testing.T) {
	mock := &mockSystemdClient{}

	logs, err := mock.GetLogs("test.service", client.LogOptions{Lines: 50})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if logs != "Mock log output" {
		t.Errorf("Unexpected logs: %s", logs)
	}

	_, err = mock.GetLogs("error.service", client.LogOptions{})
	if err == nil {
		t.Error("Expected error for error.service")
	}
}

func TestMockSystemdClient_GetConfig(t *testing.T) {
	mock := &mockSystemdClient{}

	config, err := mock.GetConfig("test.service")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(config) == 0 {
		t.Error("Expected non-empty config")
	}

	_, err = mock.GetConfig("error.service")
	if err == nil {
		t.Error("Expected error for error.service")
	}
}

func TestMockSystemdClient_EditService(t *testing.T) {
	mock := &mockSystemdClient{}

	cmd, err := mock.EditService("test.service")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd == nil {
		t.Error("Expected non-nil command")
	}

	_, err = mock.EditService("error.service")
	if err == nil {
		t.Error("Expected error for error.service")
	}
}

func TestMockSystemdClient_ReloadDaemon(t *testing.T) {
	mock := &mockSystemdClient{}

	if err := mock.ReloadDaemon(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	os.Setenv("SYSTEMD_TUI_TEST_FAIL_RELOAD", "1")
	defer os.Unsetenv("SYSTEMD_TUI_TEST_FAIL_RELOAD")

	if err := mock.ReloadDaemon(); err == nil {
		t.Error("Expected error when fail reload is set")
	}
}
