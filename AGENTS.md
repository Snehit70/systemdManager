# Systemd TUI Manager (Go) - Project Plan

## 1. Project Overview
A terminal user interface (TUI) for managing `systemd --user` services on a personal development machine. Built with Go and Bubble Tea.

**Target**: Linux Workstation (User Services)
**Stack**: Go 1.25.5, Bubble Tea, Bubbles, Lip Gloss, `os/exec`

## 2. Requirements
- **Scope**: User services only (`systemd --user`).
- **Layout**: Split view (Lazygit style).
    - **Left**: Service list (scrollable, filterable).
    - **Right**: Details pane (Logs, Status, Config).
- **Interactions**:
    - `Start`, `Stop`, `Restart`, `Enable`, `Disable`.
    - `Edit`: Shell out to `$EDITOR` (default: vim), trigger `daemon-reload` on exit.
- **Visuals**: Status indicators (Green=Active, Red=Failed, Gray=Inactive) via colored service names.

## 3. Architecture

### 3.1 Data Models (`internal/service`)
- `Unit`: Struct representing a systemd unit.
    - Fields: Unit, Load, Active, Sub, Description (JSON tags for `systemctl --output=json`).
- `SystemdClient`: Wrapper around `systemctl --user` commands.
    - Methods: `ListUnits()`, `GetLogs(unit)`, `StartUnit(unit)`, `StopUnit(unit)`, `RestartUnit(unit)`, `EnableUnit(unit)`, `DisableUnit(unit)`, `EditCmd(unit)`, `ReloadDaemon()`.

### 3.2 UI Models (`internal/ui`)
- `MainModel`: The top-level Bubble Tea model.
    - State: `activeView` (listView vs detailView), `selectedUnit`, `statusMessage`.
    - Components: `list.Model`, `viewport.Model`, `help.Model`.
- `styles.go`: Centralized Lip Gloss styles for consistent theming.
- `delegate.go`: Custom list delegate for colored service names.

### 3.3 Main Loop
- **Tick Cmd**: Poll for status updates every 2 seconds using `tea.Tick`.
- **KeyMsg**: Handle navigation and shortcuts.
- **Error Handling**: `errMsg` type for proper error propagation to UI.

## 4. Implementation Status

### Phase 1: Foundation - COMPLETE
- [x] Initialize Go module (`go mod init systemd-tui`).
- [x] Set up project structure (`cmd/`, `internal/`).
- [x] Implement `SystemdClient` to parse `systemctl list-units --user --output=json`.
- [x] Verify `systemctl` connectivity.

### Phase 2: Core UI (Read-Only) - COMPLETE
- [x] Bubble Tea model setup with Init/Update/View pattern.
- [x] Implement Split Layout (List + Viewport).
- [x] Render list of services with filtering (`/` key).
- [x] Render logs and status in Detail viewport.

### Phase 3: Interactive Features - COMPLETE
- [x] Add keybindings (`s` start, `x` stop, `r` restart).
- [x] Implement command execution logic via `SystemdClient`.
- [x] Add visual feedback (status messages).
- [x] Focus switching between list and detail view (`Tab`).

### Phase 4: Advanced Features - PARTIAL
- [x] Implement `Edit` function (`$EDITOR` integration via `tea.ExecProcess`).
- [x] Implement `daemon-reload` hook after edit.
- [ ] Implement live log tailing (continuous updates).

## 5. Improvement Roadmap

### Phase 5: Bug Fixes (Priority: High)
| Task | Description | Status |
|------|-------------|--------|
| 5.1 | Fix `fetchUnits()` returning `nil` on error - return `errMsg` instead | Pending |
| 5.2 | Add `errMsg` type for proper error handling pattern | Pending |
| 5.3 | Display `statusMessage` in View - add status bar | Pending |
| 5.4 | Handle `logMsg.err` - show error when log fetch fails | Pending |

### Phase 6: Status Color Indicators (Priority: Medium)
| Task | Description | Status |
|------|-------------|--------|
| 6.1 | Create `styles.go` with centralized Lip Gloss styles | Pending |
| 6.2 | Define status colors: Green=active, Red=failed, Gray=inactive | Pending |
| 6.3 | Implement custom list delegate for colored service names | Pending |

### Phase 7: Missing Features (Priority: Medium)
| Task | Description | Status |
|------|-------------|--------|
| 7.1 | Add `EnableUnit()` / `DisableUnit()` to SystemdClient | Pending |
| 7.2 | Add `E`/`D` keybindings for enable/disable | Pending |
| 7.3 | Add polling with `tea.Tick` (2s interval) | Pending |
| 7.4 | Add confirmation prompts for stop/restart/enable/disable | Pending |

### Phase 8: Testing (Priority: Medium)
| Task | Description | Status |
|------|-------------|--------|
| 8.1 | Add unit tests for `SystemdClient` (mock JSON parsing) | Pending |
| 8.2 | Add unit tests for `MainModel` (message handling) | Pending |
| 8.3 | Add golden file tests for View output | Pending |

### Phase 9: Extras (Priority: Low)
| Task | Description | Status |
|------|-------------|--------|
| 9.1 | Service grouping by load state | Pending |
| 9.2 | Copy to clipboard (unit name, logs) | Pending |
| 9.3 | Config file support (`~/.config/systemd-tui/config.yaml`) | Pending |
| 9.4 | Status view toggle (logs vs full `systemctl status`) | Pending |

## 6. Keybindings

| Key | Action | Scope |
|-----|--------|-------|
| `j/k` or `Arrow` | Navigate list | List view |
| `Tab` | Switch focus (list/detail) | Global |
| `/` | Filter services | List view |
| `s` | Start service | List view |
| `x` | Stop service (with confirmation) | List view |
| `r` | Restart service (with confirmation) | List view |
| `e` | Edit service file | List view |
| `E` | Enable service (with confirmation) | List view |
| `D` | Disable service (with confirmation) | List view |
| `R` | Refresh service list | Global |
| `?` | Toggle help | Global |
| `q` | Quit | Global |

## 7. File Structure

```
systemdManager/
├── cmd/systemd-tui/
│   └── main.go                 # Entry point
├── internal/
│   ├── service/
│   │   ├── client.go           # SystemdClient wrapper
│   │   ├── client_test.go      # Unit tests
│   │   ├── unit.go             # Unit struct
│   │   └── unit_test.go        # Unit tests
│   └── ui/
│       ├── model.go            # MainModel (Bubble Tea)
│       ├── model_test.go       # Unit tests
│       ├── keys.go             # Keybinding definitions
│       ├── styles.go           # Lip Gloss styles
│       └── delegate.go         # Custom list delegate
├── go.mod
├── go.sum
├── .golangci.yml               # Linter configuration
└── AGENTS.md                   # This file
```

## 8. Edge Cases

| Edge Case | Handling |
|-----------|----------|
| No systemd on system | Graceful error: "systemctl not found" |
| No user services exist | Show "No services found" message |
| Service name very long | Truncate with ellipsis in list |
| journalctl not available | Fallback message in logs pane |
| Edit with no EDITOR set | Default to `vim` |
| Permission denied | Display specific permission error |
| Terminal too small | Minimum size check (80x24) |
| Filter with no results | "No matching services" message |
| Rapid key presses | Commands are queued by Bubble Tea |

## 9. Dependencies

```go
require (
    github.com/charmbracelet/bubbles v0.21.0
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
)
```

## 10. Build & Run

```bash
# Build
go build -o systemd-tui ./cmd/systemd-tui

# Run
./systemd-tui

# Run tests
go test ./...

# Run with race detector
go run -race ./cmd/systemd-tui
```
