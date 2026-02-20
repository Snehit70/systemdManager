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
- `UnitFile`: Struct for unit file state information.
  - Fields: UnitFile, State, Preset (JSON tags for `systemctl list-unit-files --output=json`).
- `Service`: Struct representing a managed service.
  - Fields: Name, Description, Status, Sub, Enabled, Load, Source, PID, Memory, CPU, Since.
  - `Source`: Indicates where unit file originates (user/system/generated/transient/static).
- `ServiceClient`: Interface defining unit management operations.
  - All methods accept `context.Context` as first parameter for cancellation support.
  - Methods: `ListServices(ctx)`, `GetLogs(ctx, unit, opts)`, `FollowLogs(ctx, unit, opts)`, `StartService(ctx, unit)`, `StopService(ctx, unit)`, `RestartService(ctx, unit)`, `EnableService(ctx, unit)`, `DisableService(ctx, unit)`, `EditService(ctx, unit)`, `ReloadDaemon(ctx)`, `CreateService(ctx, template)`.
  - Implementations can be provided by multiple backends (systemd, Docker, procfs).
  - Compile-time interface verification: `var _ ServiceClient = (*systemdClient)(nil)`.
- The interface was introduced to decouple UI logic from the service manager, enabling easier testing and future backend support.

### 3.2 UI Models (`internal/ui`)

- `MainModel`: The top-level Bubble Tea model.
  - State: `activeView` (listView vs detailView), `selectedUnit`, `statusMessage`, `ctx` (context.Context).
  - Components: `list.Model`, `viewport.Model`, `help.Model`.
  - Helper: `getSelectedService()` safely returns service (handles group headers).
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

### Phase 4: Advanced Features - COMPLETE

- [x] Implement `Edit` function (`$EDITOR` integration via `tea.ExecProcess`).
- [x] Implement `daemon-reload` hook after edit.
- [x] Implement live log tailing (continuous updates).

## 5. Improvement Roadmap

### Phase 5: Bug Fixes (Priority: High) - COMPLETE

| Task | Description                                                           | Status   |
| ---- | --------------------------------------------------------------------- | -------- |
| 5.1  | Fix `fetchUnits()` returning `nil` on error - return `errMsg` instead | Complete |
| 5.2  | Add `errMsg` type for proper error handling pattern                   | Complete |
| 5.3  | Display `statusMessage` in View - add status bar                      | Complete |
| 5.4  | Handle `logMsg.err` - show error when log fetch fails                 | Complete |

### Phase 6: Status Color Indicators (Priority: Medium) - COMPLETE

| Task | Description                                                   | Status   |
| ---- | ------------------------------------------------------------- | -------- |
| 6.1  | Create `styles.go` with centralized Lip Gloss styles          | Complete |
| 6.2  | Define status colors: Green=active, Red=failed, Gray=inactive | Complete |
| 6.3  | Implement custom list delegate for colored service names      | Complete |

### Phase 7: Missing Features (Priority: Medium) - COMPLETE

| Task | Description                                              | Status   |
| ---- | -------------------------------------------------------- | -------- |
| 7.1  | Add `EnableUnit()` / `DisableUnit()` to SystemdClient    | Complete |
| 7.2  | Add `E`/`D` keybindings for enable/disable               | Complete |
| 7.3  | Add polling with `tea.Tick` (2s interval)                | Complete |
| 7.4  | Add confirmation prompts for stop/restart/enable/disable | Complete |

### Phase 8: Testing (Priority: Medium) - PARTIAL

| Task | Description                                            | Status   |
| ---- | ------------------------------------------------------ | -------- |
| 8.1  | Add unit tests for `SystemdClient` (mock JSON parsing) | Complete |
| 8.2  | Add unit tests for `MainModel` (message handling)      | Complete |
| 8.3  | Add golden file tests for View output                  | Pending  |

### Phase 9: Extras (Priority: Low)

| Task | Description                                               | Status   |
| ---- | --------------------------------------------------------- | -------- |
| 9.1  | Service grouping by load state                            | Complete |
| 9.2  | Copy to clipboard (unit name, logs)                       | Pending  |
| 9.3  | Config file support (`~/.config/systemd-tui/config.yaml`) | Complete |
| 9.4  | Status view toggle (logs vs full `systemctl status`)      | Pending  |
| 9.5  | Theme support (dark/light/high-contrast)                  | Complete |
| 9.6  | Live log tailing with follow mode                         | Complete |

### Phase 10: Service Creation & Filtering (Priority: High) - COMPLETE

| Task | Description                                                         | Status   |
| ---- | ------------------------------------------------------------------- | -------- |
| 10.1 | Add `Source` field to `Service` struct for categorization           | Complete |
| 10.2 | Add `CreateService` method to `ServiceClient` interface             | Complete |
| 10.3 | Implement source detection (user/system/generated/transient/static) | Complete |
| 10.4 | Add filter mode cycling (`F` key)                                   | Complete |
| 10.5 | Add service creation modal (`c` key)                                | Complete |
| 10.6 | Add source indicator in list (`●` user, `○` system)                 | Complete |

### Phase 11: Go Best Practices (Priority: High) - COMPLETE

| Task | Description                                                       | Status   |
| ---- | ----------------------------------------------------------------- | -------- |
| 11.1 | Add `context.Context` to all ServiceClient methods                | Complete |
| 11.2 | Add compile-time interface verification                           | Complete |
| 11.3 | Add documentation to exported types                               | Complete |
| 11.4 | Improve error messages with operation context                     | Complete |
| 11.5 | Add `getSelectedService()` helper for safe type assertion         | Complete |
| 11.6 | Fix panic when selecting group header items                       | Complete |

## 6. Keybindings

| Key              | Action                                          | Scope     |
| ---------------- | ----------------------------------------------- | --------- |
| `j/k` or `Arrow` | Navigate list                                   | List view |
| `Tab`            | Switch focus (list/detail)                      | Global    |
| `/`              | Filter services                                 | List view |
| `s`              | Start service                                   | List view |
| `x`              | Stop service (with confirmation)                | List view |
| `r`              | Restart service (with confirmation)             | List view |
| `e`              | Edit service file                               | List view |
| `E`              | Enable service (with confirmation)              | List view |
| `D`              | Disable service (with confirmation)             | List view |
| `c`              | Create new service                              | List view |
| `f`              | Toggle follow mode (live log tailing)           | Global    |
| `g`              | Toggle grouping (none/status/load)              | Global    |
| `F`              | Cycle filter mode (all/my services/hide system) | Global    |
| `R`              | Refresh service list                            | Global    |
| `?`              | Toggle help                                     | Global    |
| `q`              | Quit                                            | Global    |

### Create Service Modal Keys

| Key         | Action                                      |
| ----------- | ------------------------------------------- |
| `Tab`       | Next field                                  |
| `Shift+Tab` | Previous field                              |
| `t`         | Toggle service type (simple/oneshot)        |
| `r`         | Cycle restart policy (on-failure/always/no) |
| `Enter`     | Create service                              |
| `Esc`       | Cancel                                      |

## 7. File Structure

```
systemdManager/
├── cmd/systemd-tui/
│   └── main.go                 # Entry point
├── internal/
│   ├── client/
│   │   └── types.go            # ServiceClient interface, Service, ServiceTemplate structs
│   ├── config/
│   │   └── config.go           # YAML config loading, themes
│   ├── service/
│   │   ├── client.go           # SystemdClient implements ServiceClient
│   │   ├── client_test.go      # Unit tests using mock implementations
│   │   └── unit.go             # Unit, UnitFile structs
│   └── ui/
│       ├── model.go            # MainModel (Bubble Tea)
│       ├── model_test.go       # Unit tests
│       ├── styles.go           # Lip Gloss styles with theme support
│       └── delegate.go         # Custom list delegate
├── docs/
│   ├── DESIGN.md               # UI architecture, specifications
│   ├── KEYBINDINGS.md          # Complete keybinding reference
│   ├── STATE_MACHINE.md        # Modes, edge cases, performance
│   ├── VISUAL_DESIGN.md        # Typography, colors, accessibility
│   └── ARCHITECTURE.md         # Technical architecture
├── go.mod
├── go.sum
├── .golangci.yml               # Linter configuration
├── AGENTS.md                   # This file
└── TESTING.md                  # Manual testing guide
```

## 8. Edge Cases

| Edge Case                | Handling                              |
| ------------------------ | ------------------------------------- |
| No systemd on system     | Graceful error: "systemctl not found" |
| No user services exist   | Show "No services found" message      |
| Service name very long   | Truncate with ellipsis in list        |
| journalctl not available | Fallback message in logs pane         |
| Edit with no EDITOR set  | Default to `vim`                      |
| Permission denied        | Display specific permission error     |
| Terminal too small       | Minimum size check (80x24)            |
| Filter with no results   | "No matching services" message        |
| Rapid key presses        | Commands are queued by Bubble Tea     |
| Group header selected    | `getSelectedService()` returns nil safely |

## 9. Dependencies

```go
require (
    github.com/charmbracelet/bubbles v0.21.0
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
    gopkg.in/yaml.v3 v3.0.1
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
