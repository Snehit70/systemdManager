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

var _ client.ServiceClient = (*mockServiceClient)(nil)

type mockServiceClient struct {
	services []client.Service
}

func (m *mockServiceClient) ListServices(ctx context.Context) ([]client.Service, error) {
	return m.services, nil
}

func (m *mockServiceClient) StartService(ctx context.Context, name string) error   { return nil }
func (m *mockServiceClient) StopService(ctx context.Context, name string) error    { return nil }
func (m *mockServiceClient) RestartService(ctx context.Context, name string) error { return nil }
func (m *mockServiceClient) EnableService(ctx context.Context, name string) error  { return nil }
func (m *mockServiceClient) DisableService(ctx context.Context, name string) error { return nil }
func (m *mockServiceClient) GetStatus(ctx context.Context, name string) (client.ServiceStatus, error) {
	return "active", nil
}
func (m *mockServiceClient) GetLogs(ctx context.Context, name string, opts client.LogOptions) (string, error) {
	return "Mock logs", nil
}
func (m *mockServiceClient) GetConfig(ctx context.Context, name string) (string, error) {
	return "[Service]", nil
}
func (m *mockServiceClient) GetStatusDetails(ctx context.Context, name string) (string, error) {
	return "● test.service - Mock Status", nil
}
func (m *mockServiceClient) EditService(ctx context.Context, name string) (*exec.Cmd, error) {
	cmd := exec.Command("true")
	return cmd, nil
}
func (m *mockServiceClient) ReloadDaemon(ctx context.Context) error { return nil }
func (m *mockServiceClient) FollowLogs(ctx context.Context, name string, opts client.LogOptions) (<-chan string, context.CancelFunc, error) {
	ch := make(chan string)
	cancel := func() { close(ch) }
	return ch, cancel, nil
}
func (m *mockServiceClient) CreateService(ctx context.Context, tmpl client.ServiceTemplate) error {
	return nil
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
	errMsg := errMsg{op: "Test", err: testErr}

	updatedModel, _ := model.Update(errMsg)
	m := updatedModel.(MainModel)

	expected := "Error: Test - test error"
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

func TestServiceLoadingPopulatesList(t *testing.T) {
	mockClient := &mockServiceClient{
		services: []client.Service{
			{Name: "test1.service", Description: "Test 1", Status: client.StatusActive, Source: client.SourceUser},
			{Name: "test2.service", Description: "Test 2", Status: client.StatusInactive, Source: client.SourceSystem},
		},
	}
	model := NewMainModel(mockClient, config.Default())

	// Simulate window size first (like real app)
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedModel, _ := model.Update(sizeMsg)
	m := updatedModel.(MainModel)

	// Simulate services being loaded
	servicesMsg := []client.Service{
		{Name: "test1.service", Description: "Test 1", Status: client.StatusActive, Source: client.SourceUser},
		{Name: "test2.service", Description: "Test 2", Status: client.StatusInactive, Source: client.SourceSystem},
	}
	updatedModel, _ = m.Update(servicesMsg)
	m = updatedModel.(MainModel)

	if len(m.services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(m.services))
	}

	items := m.list.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 list items, got %d", len(items))
	}

	// Verify status message
	if m.statusMessage != "Loaded 2 services (1 user-created)" {
		t.Errorf("Expected 'Loaded 2 services (1 user-created)', got '%s'", m.statusMessage)
	}
}
