package ui

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"systemd-tui/internal/client"
	"systemd-tui/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

var _ client.ServiceClient = (*mockServiceClient)(nil)

type mockServiceClient struct {
	services    []client.Service
	createCalls int
	createErr   error
	followOpts  client.LogOptions
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
func (m *mockServiceClient) FollowLogs(ctx context.Context, name string, opts client.LogOptions) (<-chan client.LogEvent, context.CancelFunc, error) {
	m.followOpts = opts
	ch := make(chan client.LogEvent)
	cancel := func() { close(ch) }
	return ch, cancel, nil
}
func (m *mockServiceClient) CreateService(ctx context.Context, tmpl client.ServiceTemplate) error {
	m.createCalls++
	return m.createErr
}

func newTestModel() MainModel {
	return NewMainModel(&mockServiceClient{}, config.Default())
}

func applyServices(t *testing.T, model MainModel, services []client.Service) MainModel {
	t.Helper()
	updated, _ := model.Update(services)
	return updated.(MainModel)
}

func selectedServiceName(t *testing.T, model MainModel) string {
	t.Helper()
	service := model.getSelectedService()
	if service == nil {
		t.Fatal("expected a service to be selected")
	}
	if service.Name != model.selectedSvc {
		t.Fatalf("visible selection %q does not match selectedSvc %q", service.Name, model.selectedSvc)
	}
	return service.Name
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

func TestCreateServiceFromModalRunsAsync(t *testing.T) {
	mockClient := &mockServiceClient{}
	model := NewMainModel(mockClient, config.Default())
	model.showCreate = true
	model.createModal = newCreateModal()
	model.createModal.nameInput.SetValue("demo")
	model.createModal.execInput.SetValue("/bin/true")

	updatedModel, cmd := model.createServiceFromModal()
	m := updatedModel.(MainModel)

	if cmd == nil {
		t.Fatal("expected create command")
	}
	if mockClient.createCalls != 0 {
		t.Fatalf("expected create not to run synchronously, got %d calls", mockClient.createCalls)
	}
	if !m.creating {
		t.Fatal("expected creating state while command is pending")
	}

	msg := cmd()
	if mockClient.createCalls != 1 {
		t.Fatalf("expected create command to run once, got %d calls", mockClient.createCalls)
	}

	updatedModel, _ = m.Update(msg)
	m = updatedModel.(MainModel)

	if m.creating {
		t.Fatal("expected creating state to clear after result")
	}
	if m.showCreate {
		t.Fatal("expected modal to close after successful creation")
	}
	if m.statusMessage != "Created service: demo" {
		t.Fatalf("unexpected status message: %q", m.statusMessage)
	}
}

func TestStartFollowUsesConfiguredLogLines(t *testing.T) {
	mockClient := &mockServiceClient{}
	cfg := config.Default()
	cfg.General.LogLines = 123
	model := NewMainModel(mockClient, cfg)

	cmd := model.startFollow("demo.service", 1)

	if cmd == nil {
		t.Fatal("expected follow command")
	}
	if mockClient.followOpts.Lines != 0 {
		t.Fatalf("expected FollowLogs not to run synchronously, got %d", mockClient.followOpts.Lines)
	}

	msg := cmd()
	started, ok := msg.(followStartedMsg)
	if !ok {
		t.Fatalf("expected followStartedMsg, got %T", msg)
	}
	if started.unit != "demo.service" {
		t.Fatalf("expected unit demo.service, got %s", started.unit)
	}
	if started.ch == nil {
		t.Fatal("expected follow channel")
	}
	if started.cancel == nil {
		t.Fatal("expected follow cancel function")
	}
	if mockClient.followOpts.Lines != 123 {
		t.Fatalf("expected configured log lines 123, got %d", mockClient.followOpts.Lines)
	}
}

func TestGroupingPreservesSelectedService(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{
		{Name: "active.service", Status: client.StatusActive, Source: client.SourceUser},
		{Name: "inactive.service", Status: client.StatusInactive, Source: client.SourceUser},
		{Name: "failed.service", Status: client.StatusFailed, Source: client.SourceUser},
	})

	model.list.Select(1)
	model.selectedSvc = "inactive.service"
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	model = updated.(MainModel)

	if got := selectedServiceName(t, model); got != "inactive.service" {
		t.Fatalf("selected service after grouping = %q, want inactive.service", got)
	}
	if _, ok := model.list.SelectedItem().(groupHeaderItem); ok {
		t.Fatal("group header became the effective selection")
	}
}

func TestSourceFilterFallsBackAndStopsFollow(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{
		{Name: "system.service", Status: client.StatusActive, Source: client.SourceSystem},
		{Name: "user.service", Status: client.StatusActive, Source: client.SourceUser},
	})

	cancelled := false
	model.following = true
	model.followCancel = func() { cancelled = true }
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("F")})
	model = updated.(MainModel)

	if got := selectedServiceName(t, model); got != "user.service" {
		t.Fatalf("selected service after source filter = %q, want user.service", got)
	}
	if model.following {
		t.Fatal("follow mode remained active after the selected service changed")
	}
	if !cancelled {
		t.Fatal("active follow session was not cancelled")
	}
}

func TestSourceFilterInvalidatesPendingFollow(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{
		{Name: "system.service", Status: client.StatusActive, Source: client.SourceSystem},
		{Name: "user.service", Status: client.StatusActive, Source: client.SourceUser},
	})

	model.followPending = true
	model.followSessionID = 7
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("F")})
	model = updated.(MainModel)

	if model.followPending {
		t.Fatal("pending follow remained active after the selected service changed")
	}
	if model.followSessionID != 8 {
		t.Fatalf("follow session ID = %d, want 8", model.followSessionID)
	}
}

func TestRefreshPreservesSelectionWhenOrderChanges(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{
		{Name: "alpha.service", Status: client.StatusActive, Source: client.SourceUser},
		{Name: "beta.service", Status: client.StatusInactive, Source: client.SourceUser},
	})

	model.list.Select(1)
	model.selectedSvc = "beta.service"
	model = applyServices(t, model, []client.Service{
		{Name: "beta.service", Status: client.StatusActive, Source: client.SourceUser},
		{Name: "alpha.service", Status: client.StatusInactive, Source: client.SourceUser},
	})

	if got := selectedServiceName(t, model); got != "beta.service" {
		t.Fatalf("selected service after refresh = %q, want beta.service", got)
	}
	if model.list.Index() != 0 {
		t.Fatalf("selected index after reorder = %d, want 0", model.list.Index())
	}
}

func TestEmptyRefreshClearsSelectionAndDetail(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{
		{Name: "demo.service", Status: client.StatusActive, Source: client.SourceUser},
	})
	model.viewport.Width = 40
	model.viewport.Height = 5
	model.viewport.SetContent("stale detail")
	if !strings.Contains(model.viewport.View(), "stale detail") {
		t.Fatal("test setup did not render stale detail content")
	}

	model = applyServices(t, model, nil)

	if model.selectedSvc != "" {
		t.Fatalf("selectedSvc after empty refresh = %q, want empty", model.selectedSvc)
	}
	if model.getSelectedService() != nil {
		t.Fatal("a service remained selected after the list became empty")
	}
	if strings.Contains(model.viewport.View(), "stale detail") {
		t.Fatal("stale detail content remained after the list became empty")
	}
}

func TestInitialGroupedListSkipsHeader(t *testing.T) {
	model := newTestModel()
	model.groupMode = groupByStatus
	model = applyServices(t, model, []client.Service{
		{Name: "demo.service", Status: client.StatusActive, Source: client.SourceUser},
	})

	if got := selectedServiceName(t, model); got != "demo.service" {
		t.Fatalf("selected service = %q, want demo.service", got)
	}
	if model.list.Index() != 1 {
		t.Fatalf("selected index = %d, want 1 after the group header", model.list.Index())
	}
}
