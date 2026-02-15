package ui

import (
	"context"
	"errors"
	"os/exec"
	"systemd-tui/internal/client"
	"systemd-tui/internal/config"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// mockServiceClient implements client.ServiceClient for testing
type mockServiceClient struct {
	services []client.Service
}

func (m *mockServiceClient) ListServices() ([]client.Service, error) {
	return m.services, nil
}

func (m *mockServiceClient) StartService(name string) error   { return nil }
func (m *mockServiceClient) StopService(name string) error    { return nil }
func (m *mockServiceClient) RestartService(name string) error { return nil }
func (m *mockServiceClient) EnableService(name string) error  { return nil }
func (m *mockServiceClient) DisableService(name string) error { return nil }
func (m *mockServiceClient) GetStatus(name string) (client.ServiceStatus, error) {
	return "active", nil
}
func (m *mockServiceClient) GetLogs(name string, opts client.LogOptions) (string, error) {
	return "Mock logs", nil
}
func (m *mockServiceClient) GetConfig(name string) (string, error) {
	return "[Service]", nil
}
func (m *mockServiceClient) EditService(name string) (*exec.Cmd, error) {
	cmd := exec.Command("true")
	return cmd, nil
}
func (m *mockServiceClient) ReloadDaemon() error { return nil }
func (m *mockServiceClient) FollowLogs(name string, opts client.LogOptions) (<-chan string, context.CancelFunc, error) {
	ch := make(chan string)
	cancel := func() { close(ch) }
	return ch, cancel, nil
}

func newTestModel() MainModel {
	return NewMainModel(&mockServiceClient{}, config.Default())
}

func TestMainModelInit(t *testing.T) {
	model := newTestModel()

	if model.activeView != listView {
		t.Errorf("Expected initial activeView to be listView (0), got %d", model.activeView)
	}
	if model.confirmingAction != "" {
		t.Errorf("Expected initial confirmingAction to be empty, got '%s'", model.confirmingAction)
	}
	if model.confirmingUnit != "" {
		t.Errorf("Expected initial confirmingUnit to be empty, got '%s'", model.confirmingUnit)
	}
}

func TestConfirmationStateTransitions(t *testing.T) {
	model := newTestModel()

	model.confirmingAction = "stop"
	model.confirmingUnit = "test.service"
	model.statusMessage = "Stop test.service? (y/n)"

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	updatedModel, _ := model.Update(keyMsg)
	m := updatedModel.(MainModel)

	if m.confirmingAction != "" {
		t.Errorf("Expected confirmingAction to be cleared, got '%s'", m.confirmingAction)
	}
	if m.confirmingUnit != "" {
		t.Errorf("Expected confirmingUnit to be cleared, got '%s'", m.confirmingUnit)
	}
	if m.statusMessage != "Cancelled" {
		t.Errorf("Expected statusMessage 'Cancelled', got '%s'", m.statusMessage)
	}
}

func TestConfirmationCancelOnEscape(t *testing.T) {
	model := newTestModel()

	model.confirmingAction = "restart"
	model.confirmingUnit = "test.service"

	keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := model.Update(keyMsg)
	m := updatedModel.(MainModel)

	if m.confirmingAction != "" {
		t.Errorf("Expected confirmingAction to be cleared on Esc, got '%s'", m.confirmingAction)
	}
	if m.statusMessage != "Cancelled" {
		t.Errorf("Expected statusMessage 'Cancelled', got '%s'", m.statusMessage)
	}
}

func TestErrorMessageHandling(t *testing.T) {
	model := newTestModel()

	testErr := errors.New("test error")
	errMsg := errMsg{context: "Test", err: testErr}

	updatedModel, _ := model.Update(errMsg)
	m := updatedModel.(MainModel)

	expected := "Test: test error"
	if m.statusMessage != expected {
		t.Errorf("Expected '%s', got '%s'", expected, m.statusMessage)
	}
}

func TestActionResultMessageHandling(t *testing.T) {
	model := newTestModel()

	actionMsg := actionResultMsg{message: "Started test.service", err: nil}
	updatedModel, _ := model.Update(actionMsg)
	m := updatedModel.(MainModel)

	if m.statusMessage != "Started test.service" {
		t.Errorf("Expected 'Started test.service', got '%s'", m.statusMessage)
	}

	// Test error case
	errorMsg := actionResultMsg{message: "", err: errors.New("failed")}
	updatedModel, _ = model.Update(errorMsg)
	m = updatedModel.(MainModel)

	if m.statusMessage != "Error: failed" {
		t.Errorf("Expected 'Error: failed', got '%s'", m.statusMessage)
	}
}

func TestWindowSizeHandling(t *testing.T) {
	model := newTestModel()

	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	m := updatedModel.(MainModel)

	if m.width != 100 || m.height != 40 {
		t.Errorf("Expected size 100x40, got %dx%d", m.width, m.height)
	}
}

func TestViewSwitching(t *testing.T) {
	model := newTestModel()

	if model.activeView != listView {
		t.Errorf("Expected listView initially")
	}

	model.activeView = detailView
	if model.activeView != detailView {
		t.Errorf("Expected detailView after switch")
	}

	model.activeView = listView
	if model.activeView != listView {
		t.Errorf("Expected listView after switch back")
	}
}
