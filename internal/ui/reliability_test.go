package ui

import (
	"errors"
	"strings"
	"testing"

	"systemd-tui/internal/client"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestStaleServiceRefreshResultIsIgnored(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{{Name: "current.service", Status: client.StatusActive}})
	model.refreshRequestID = 5
	model.refreshPending = true

	updated, _ := model.Update(serviceListMsg{
		id:       4,
		services: []client.Service{{Name: "stale.service", Status: client.StatusInactive}},
	})
	model = updated.(MainModel)

	if len(model.services) != 1 || model.services[0].Name != "current.service" {
		t.Fatalf("stale refresh overwrote current data: %#v", model.services)
	}
	if !model.refreshPending {
		t.Fatal("stale refresh cleared the current request's pending state")
	}
}

func TestLatestServiceRefreshAppliesAndClearsPending(t *testing.T) {
	model := newTestModel()
	model.refreshRequestID = 5
	model.refreshPending = true

	updated, _ := model.Update(serviceListMsg{
		id:       5,
		services: []client.Service{{Name: "latest.service", Status: client.StatusActive}},
	})
	model = updated.(MainModel)

	if model.refreshPending {
		t.Fatal("latest refresh did not clear pending state")
	}
	if len(model.services) != 1 || model.services[0].Name != "latest.service" {
		t.Fatalf("latest refresh was not applied: %#v", model.services)
	}
}

func TestServiceRefreshErrorPreservesCurrentList(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{{Name: "current.service", Status: client.StatusActive}})
	model.refreshRequestID = 3
	model.refreshPending = true

	updated, _ := model.Update(serviceListMsg{id: 3, err: errors.New("bus unavailable")})
	model = updated.(MainModel)

	if model.refreshPending {
		t.Fatal("failed current refresh left pending set")
	}
	if len(model.services) != 1 || model.services[0].Name != "current.service" {
		t.Fatalf("failed refresh destroyed the current list: %#v", model.services)
	}
	if !strings.Contains(model.statusMessage, "bus unavailable") {
		t.Fatalf("refresh error was not surfaced: %q", model.statusMessage)
	}
}

func TestPeriodicRefreshDoesNotOverlapButForcedRefreshSupersedes(t *testing.T) {
	model := newTestModel()
	model.refreshRequestID = 8
	model.refreshPending = true

	if cmd := model.requestServices(false); cmd != nil {
		t.Fatal("periodic refresh overlapped an in-flight request")
	}
	cmd := model.requestServices(true)
	if cmd == nil {
		t.Fatal("forced post-action refresh did not supersede the in-flight request")
	}
	if model.refreshRequestID != 9 || !model.refreshPending {
		t.Fatalf("forced refresh state = id %d pending %v", model.refreshRequestID, model.refreshPending)
	}
	msg, ok := cmd().(serviceListMsg)
	if !ok || msg.id != 9 {
		t.Fatalf("forced refresh returned %#v", msg)
	}
}

func TestStaleFollowStartDoesNotClearCurrentPending(t *testing.T) {
	model := newTestModel()
	model.followSessionID = 2
	model.followPending = true
	cancelled := false
	ch := make(chan client.LogEvent)

	updated, _ := model.Update(followStartedMsg{
		id:     1,
		unit:   "old.service",
		ch:     ch,
		cancel: func() { cancelled = true },
	})
	model = updated.(MainModel)

	if !cancelled {
		t.Fatal("stale started stream was not cancelled")
	}
	if !model.followPending {
		t.Fatal("stale started message cleared the current follow request")
	}
	if model.following {
		t.Fatal("stale follow session became active")
	}
}

func TestFollowStartFailureIsScopedToSession(t *testing.T) {
	model := newTestModel()
	model.followSessionID = 4
	model.followPending = true
	model.statusMessage = "starting current"

	updated, _ := model.Update(followStartFailedMsg{id: 3, unit: "old.service", err: errors.New("old failure")})
	model = updated.(MainModel)
	if !model.followPending || model.statusMessage != "starting current" {
		t.Fatalf("stale failure changed current follow state: pending=%v status=%q", model.followPending, model.statusMessage)
	}

	updated, _ = model.Update(followStartFailedMsg{id: 4, unit: "current.service", err: errors.New("start failed")})
	model = updated.(MainModel)
	if model.followPending {
		t.Fatal("current follow-start failure did not clear pending state")
	}
	if !strings.Contains(model.statusMessage, "current.service") || !strings.Contains(model.statusMessage, "start failed") {
		t.Fatalf("current follow-start failure was not surfaced: %q", model.statusMessage)
	}
}

func TestFollowStreamFailureStopsCurrentSession(t *testing.T) {
	model := newTestModel()
	model = applyServices(t, model, []client.Service{{Name: "demo.service", Status: client.StatusActive}})
	model.followSessionID = 7
	model.following = true
	cancelled := false
	model.followCancel = func() { cancelled = true }
	model.followLogChan = make(chan client.LogEvent)

	updated, _ := model.Update(followStreamFailedMsg{id: 7, unit: "demo.service", err: errors.New("journal died")})
	model = updated.(MainModel)

	if model.following {
		t.Fatal("failed follow stream remained active")
	}
	if !cancelled {
		t.Fatal("failed follow stream was not cleaned up")
	}
	if !strings.Contains(model.statusMessage, "journal died") {
		t.Fatalf("follow failure was mislabeled as a normal stop: %q", model.statusMessage)
	}
}

func TestFilteringReservesAndReleasesTerminalRow(t *testing.T) {
	model := newTestModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updated.(MainModel)
	baseHeight := model.viewport.Height

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model = updated.(MainModel)
	if model.list.FilterState() != list.Filtering {
		t.Fatalf("filter did not enter input state: %v", model.list.FilterState())
	}
	if model.viewport.Height != baseHeight-1 {
		t.Fatalf("viewport height while filtering = %d, want %d", model.viewport.Height, baseHeight-1)
	}
	if got := lipgloss.Height(model.View()); got > model.height {
		t.Fatalf("filter UI renders %d rows into a %d-row terminal (list=%d viewport=%d filter=%d status=%d help=%d)",
			got, model.height,
			lipgloss.Height(model.list.View()),
			lipgloss.Height(model.viewport.View()),
			lipgloss.Height(model.renderFilterBar(model.width)),
			lipgloss.Height(model.renderStatusBar(model.width)),
			lipgloss.Height(model.renderHelpView(model.width)))
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(MainModel)
	if model.viewport.Height != baseHeight {
		t.Fatalf("viewport height after closing filter = %d, want %d", model.viewport.Height, baseHeight)
	}
}

func TestExpandedHelpAndFilterFitTerminal(t *testing.T) {
	model := newTestModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updated.(MainModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	model = updated.(MainModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model = updated.(MainModel)

	if got := lipgloss.Height(model.View()); got > model.height {
		t.Fatalf("expanded help + filter render %d rows into a %d-row terminal", got, model.height)
	}
}
