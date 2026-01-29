# Systemd TUI Manager (Go) - Project Plan

## 1. Project Overview
A terminal user interface (TUI) for managing `systemd --user` services on a personal development machine. Built with Go and Bubble Tea.

**Target**: Linux Workstation (User Services)
**Stack**: Go, Bubble Tea, Lip Gloss, `os/exec`

## 2. Requirements
- **Scope**: User services only (`systemd --user`).
- **Layout**: Split view (Lazygit style).
    - **Left**: Service list (scrollable, filterable).
    - **Right**: Details pane (Logs, Status, Config).
- **Interactions**:
    - `Start`, `Stop`, `Restart`, `Enable`, `Disable`.
    - `Edit`: Shell out to `nvim`, trigger `daemon-reload` on exit.
- **Visuals**: Status indicators (Green=Active, Red=Failed, Gray=Inactive).

## 3. Architecture

### 3.1 Data Models (`internal/service`)
- `Unit`: Struct representing a systemd unit.
    - Fields: Name, ActiveState, LoadState, SubState, Description.
- `SystemdClient`: Wrapper around `systemctl --user` commands.
    - Methods: `ListUnits()`, `GetStatus(unit)`, `GetLogs(unit)`, `Restart(unit)`, `Edit(unit)`.

### 3.2 UI Models (`internal/ui`)
- `MainModel`: The top-level Bubble Tea model.
    - State: `Focus` (List vs Detail), `SelectedUnit`.
- `ListView`: Bubble Tea list component.
- `DetailView`: Viewport component for logs/status.

### 3.3 Main Loop
- **Tick Cmd**: Poll for status updates every 1-2 seconds (or use `systemd-dbus` events if feasible, starting with polling for simplicity).
- **KeyMsg**: Handle navigation and shortcuts.

## 4. Implementation Steps (Todos)

### Phase 1: Foundation
- [ ] Initialize Go module (`go mod init systemd-tui`).
- [ ] Set up project structure (`cmd/`, `internal/`).
- [ ] Implement `SystemdClient` to parse `systemctl list-units --user`.
- [ ] Verify `systemctl` connectivity.

### Phase 2: Core UI (Read-Only)
- [ ] specific Bubble Tea model setup.
- [ ] Implement Split Layout (Sidebar + Main Content).
- [ ] Render list of services in Sidebar.
- [ ] Render raw status output in Main Content.

### Phase 3: Interactive Features
- [ ] Add keybindings (`s` start, `x` stop, `r` restart).
- [ ] Implement command execution logic.
- [ ] Add visual feedback (loading states).

### Phase 4: Advanced Features
- [ ] Implement `Edit` function (`nvim` integration).
- [ ] Implement `daemon-reload` hook.
- [ ] Implement Log tailing (live updates in Detail view).

## 5. Keybindings
- `j/k`: Navigate list
- `Enter`: Focus details / Toggle
- `s`: Start
- `x`: Stop
- `r`: Restart
- `e`: Edit (Shell out to nvim)
- `q`: Quit
