# Keybinding System

> Part of [DESIGN.md](./DESIGN.md)

---

## 6. Keybinding System

### 6.1 Design Philosophy

1. **Vim-compatible**: j/k navigation, : for commands
2. **Discoverable**: Always show relevant bindings
3. **Configurable**: User can remap everything
4. **Context-aware**: Bindings change based on active panel

### 6.2 Global Keybindings

| Key | Action | Description |
|-----|--------|-------------|
| `q` | Quit | Exit application (with confirmation if action pending) |
| `Ctrl+C` | Force Quit | Immediate exit, no confirmation |
| `?` | Help | Toggle full help overlay |
| `Tab` | Switch Panel | Move focus between panels |
| `Shift+Tab` | Reverse Switch | Move focus in opposite direction |
| `:` | Command Mode | Enter command palette |
| `Esc` | Cancel/Normal | Return to normal mode, cancel current action |
| `R` | Refresh | Force refresh all data |
| `1-4` | Switch Tabs | Jump to tab in detail pane |

### 6.3 Service List Keybindings

| Key | Action | Confirmation | Description |
|-----|--------|--------------|-------------|
| `j` / `Down` | Down | No | Move selection down |
| `k` / `Up` | Up | No | Move selection up |
| `g` | Top | No | Jump to first service |
| `G` | Bottom | No | Jump to last service |
| `s` | Start | No | Start selected service |
| `x` | Stop | Yes | Stop selected service |
| `r` | Restart | Yes | Restart selected service |
| `e` | Edit | No | Open service file in editor |
| `E` | Enable | Yes | Enable service autostart |
| `D` | Disable | Yes | Disable service autostart |
| `/` | Filter | No | Enter filter mode |
| `Enter` | Open | No | Focus detail pane with full info |
| `y` | Yank | No | Copy service name to clipboard |

### 6.4 Detail Pane Keybindings

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `Down` | Scroll Down | Scroll content down |
| `k` / `Up` | Scroll Up | Scroll content up |
| `l` | Live Toggle | Toggle live log tailing |
| `f` | Filter Logs | Filter log content |
| `c` | Copy | Copy visible content to clipboard |
| `/` | Search | Search within content |
| `n` | Next Match | Jump to next search match |
| `N` | Prev Match | Jump to previous search match |

### 6.5 Filter Mode Keybindings

| Key | Action | Description |
|-----|--------|-------------|
| `type` | Filter | Fuzzy filter services |
| `Enter` | Apply | Apply filter, return to normal |
| `Esc` | Clear | Clear filter, return to normal |
| `Ctrl+U` | Clear Line | Clear filter input |
| `Backspace` | Delete | Remove last character |

### 6.6 Command Mode Keybindings

| Key | Action | Description |
|-----|--------|-------------|
| `type` | Command | Enter command |
| `Tab` | Complete | Autocomplete command |
| `Enter` | Execute | Run command |
| `Esc` | Cancel | Return to normal mode |
| `Ctrl+P` | History Prev | Previous command |
| `Ctrl+N` | History Next | Next command |

### 6.7 Command Palette Commands

| Command | Description | Example |
|---------|-------------|---------|
| `start <service>` | Start service | `start my-service` |
| `stop <service>` | Stop service | `stop nginx` |
| `restart <service>` | Restart service | `restart redis` |
| `enable <service>` | Enable service | `enable docker` |
| `disable <service>` | Disable service | `disable apache` |
| `edit <service>` | Edit service file | `edit my-service` |
| `logs <service>` | View logs | `logs postgres` |
| `status <service>` | View status | `status nginx` |
| `refresh` | Refresh all data | `refresh` |
| `theme <name>` | Change theme | `theme dark` |
| `config` | Open config file | `config` |
| `help` | Show help | `help` |
| `quit` / `q` | Exit application | `quit` |

### 6.8 Confirmation Actions

Actions requiring confirmation (to prevent accidental data loss):

| Action | Key | Confirmation Message |
|--------|-----|---------------------|
| Stop | `x` | `Stop <service>? (y/n)` |
| Restart | `r` | `Restart <service>? (y/n)` |
| Enable | `E` | `Enable <service>? (y/n)` |
| Disable | `D` | `Disable <service>? (y/n)` |

**Why Start doesn't need confirmation:**
- Starting a service is a non-destructive action
- Services can always be stopped if started accidentally
- Reduces friction for common operations

---

## 7. User Workflow Analysis

### 7.1 Primary Workflows

#### Workflow 1: Check Service Status

```
User Goal: See if a service is running
Steps:
1. Launch app -> Services auto-load
2. Scan list for service name (visual search)
3. OR Press `/` -> Type name -> Enter (fuzzy search)
4. View status icon (*/o/X)
5. View details in right panel

Time: < 5 seconds
Cognitive Load: Low
```

#### Workflow 2: Start a Service

```
User Goal: Start an inactive service
Steps:
1. Navigate to service (j/k or filter)
2. Press `s` -> Immediate feedback in status bar
3. Watch status icon change (o -> @ -> *)
4. Optional: Press `l` for live logs

Time: < 3 seconds
Cognitive Load: Low
Confirmation: None (safe action)
```

#### Workflow 3: Stop a Service

```
User Goal: Stop a running service
Steps:
1. Navigate to service
2. Press `x` -> Confirmation prompt appears
3. Press `y` to confirm OR `n` to cancel
4. Watch status icon change (* -> - -> o)

Time: < 5 seconds
Cognitive Load: Low
Confirmation: Required (destructive action)
```

#### Workflow 4: Debug a Failed Service

```
User Goal: Understand why a service failed
Steps:
1. Spot failed service (X icon, red highlight)
2. Navigate to it
3. View error in detail panel (auto-selected)
4. Press `l` for live logs OR `4` for dependencies
5. Press `e` to edit service file
6. Fix issue, save -> Daemon auto-reload
7. Press `r` to restart
8. Watch status change

Time: Variable
Cognitive Load: High (debugging)
Tools Used: Logs, Edit, Status
```

#### Workflow 5: Bulk Operations

```
User Goal: Start multiple services at once
Steps:
1. Press `/` -> Filter by pattern (e.g., "dev-*")
2. View filtered list
3. Press `:` -> Command mode
4. Type `start-all` -> Enter
5. Watch progress in status bar

Time: < 10 seconds
Cognitive Load: Medium
Feature: Bulk commands (future)
```

### 7.2 Edge Case Workflows

#### No Services Found

```
Trigger: systemctl returns empty list
Display:
+-------------------------------------------------------------+
|                                                             |
|                    No services found                        |
|                                                             |
|   Create a user service:                                    |
|   $ systemctl --user edit --full my-service                 |
|                                                             |
|   [Press ? for help]                                        |
|                                                             |
+-------------------------------------------------------------+
```

#### Systemctl Not Available

```
Trigger: systemctl command not found
Display:
+-------------------------------------------------------------+
| X Error: systemctl not found                                |
|                                                             |
|   This application requires systemd.                        |
|                                                             |
|   Are you running on a systemd-based system?                |
|                                                             |
|   [Press q to exit]                                         |
+-------------------------------------------------------------+
```

#### Permission Denied

```
Trigger: Action fails due to permissions
Display:
+-------------------------------------------------------------+
| ! Warning: Permission denied                                |
|                                                             |
|   Cannot stop 'system-service.service'                      |
|   This is a system service requiring elevated privileges.   |
|                                                             |
|   Try: sudo systemctl stop system-service                   |
|                                                             |
|   [Press Esc to dismiss]                                    |
+-------------------------------------------------------------+
```
