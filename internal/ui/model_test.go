package ui

import (
	"errors"
	"systemd-tui/internal/service"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMainModelInit(t *testing.T) {
	client := service.NewSystemdClient()
	model := NewMainModel(client)

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
	client := service.NewSystemdClient()
	model := NewMainModel(client)

	model.confirmingAction = "stop"
	model.confirmingUnit = "test.service"
	model.statusMessage = "Stop test.service? (y/n)"

	if model.confirmingAction != "stop" {
		t.Errorf("Expected confirmingAction 'stop', got '%s'", model.confirmingAction)
	}

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	updatedModel, _ := model.Update(keyMsg)
	m := updatedModel.(MainModel)

	if m.confirmingAction != "" {
		t.Errorf("Expected confirmingAction to be cleared after 'n', got '%s'", m.confirmingAction)
	}

	if m.confirmingUnit != "" {
		t.Errorf("Expected confirmingUnit to be cleared after 'n', got '%s'", m.confirmingUnit)
	}

	if m.statusMessage != "Cancelled" {
		t.Errorf("Expected statusMessage 'Cancelled', got '%s'", m.statusMessage)
	}
}

func TestConfirmationCancelOnEscape(t *testing.T) {
	client := service.NewSystemdClient()
	model := NewMainModel(client)

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
	client := service.NewSystemdClient()
	model := NewMainModel(client)

	testErr := errors.New("test error occurred")
	errMsg := errMsg{
		context: "Test operation",
		err:     testErr,
	}

	updatedModel, _ := model.Update(errMsg)
	m := updatedModel.(MainModel)

	expectedMsg := "Test operation: test error occurred"
	if m.statusMessage != expectedMsg {
		t.Errorf("Expected statusMessage '%s', got '%s'", expectedMsg, m.statusMessage)
	}
}

func TestActionResultMessageHandling(t *testing.T) {
	client := service.NewSystemdClient()
	model := NewMainModel(client)

	actionMsg := actionResultMsg{
		message: "Started test.service",
		err:     nil,
	}

	updatedModel, _ := model.Update(actionMsg)
	m := updatedModel.(MainModel)

	if m.statusMessage != "Started test.service" {
		t.Errorf("Expected statusMessage 'Started test.service', got '%s'", m.statusMessage)
	}
}

func TestWindowSizeHandling(t *testing.T) {
	client := service.NewSystemdClient()
	model := NewMainModel(client)

	sizeMsg := tea.WindowSizeMsg{
		Width:  100,
		Height: 40,
	}

	updatedModel, _ := model.Update(sizeMsg)
	m := updatedModel.(MainModel)

	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}

	if m.height != 40 {
		t.Errorf("Expected height 40, got %d", m.height)
	}
}

func TestViewSwitching(t *testing.T) {
	client := service.NewSystemdClient()
	model := NewMainModel(client)

	if model.activeView != listView {
		t.Errorf("Expected initial view to be listView, got %d", model.activeView)
	}

	model.activeView = detailView
	if model.activeView != detailView {
		t.Errorf("Expected view to be detailView after switch, got %d", model.activeView)
	}

	model.activeView = listView
	if model.activeView != listView {
		t.Errorf("Expected view to be listView after switch back, got %d", model.activeView)
	}
}
