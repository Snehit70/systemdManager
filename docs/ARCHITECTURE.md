# Architecture & System Design

> Part of [DESIGN.md](./DESIGN.md)

---

## Architecture Overview

### Current Architecture

```
+-------------------+     +------------------+
|      main.go      |---->|   SystemdClient  |
+-------------------+     +------------------+
         |                        |
         v                        v
+-------------------+     +------------------+
|    MainModel      |<----|    Unit struct   |
|  (Bubble Tea)     |     +------------------+
+-------------------+
         |
         v
+-----------------------------------+
|         UI Components             |
+-----------------------------------+
| list.Model    | viewport.Model    |
| help.Model    | styles.go         |
| delegate.go                       |
+-----------------------------------+
```

### Proposed Architecture (v2.0)

```
+------------------+     +---------------------+     +------------------+
|     main.go      |---->|  ServiceClient      |<----|   SystemdClient  |
+------------------+     |  (interface)        |     +------------------+
                         +---------------------+              |
                                   ^                          |
                                   |                          v
                         +---------+---------+     +------------------+
                         |                   |     |  DockerClient    |
                         |                   |     +------------------+
                         |                   |
                         v                   v
+------------------+     +---------------------+     +------------------+
|  ConfigManager   |---->|      MainModel      |<----|   ProcfsClient   |
+------------------+     |    (Bubble Tea)     |     +------------------+
                         +---------------------+
                                   |
         +-------------------------+-------------------------+
         |                         |                         |
         v                         v                         v
+------------------+     +------------------+     +------------------+
|  ServiceList     |     |  DetailPane      |     |  NotificationMgr |
|  (Component)     |     |  (Component)     |     |  (Component)     |
+------------------+     +------------------+     +------------------+
```

---

## Component Design

### ServiceClient Interface

```go
// ServiceClient defines the interface for service management backends
type ServiceClient interface {
    // List all services
    ListServices() ([]Service, error)
    
    // Service actions
    StartService(name string) error
    StopService(name string) error
    RestartService(name string) error
    EnableService(name string) error
    DisableService(name string) error
    
    // Service information
    GetStatus(name string) (ServiceStatus, error)
    GetLogs(name string, opts LogOptions) (string, error)
    GetConfig(name string) (string, error)
    
    // Service file management
    EditService(name string) (*exec.Cmd, error)
    ReloadDaemon() error
}

// Service represents a managed service
type Service struct {
    Name        string
    Description string
    Status      ServiceStatus
    Enabled     bool
    PID         int
    Memory      int64
    CPU         float64
    Since       time.Time
}

// ServiceStatus represents the current state
type ServiceStatus string

const (
    StatusActive      ServiceStatus = "active"
    StatusInactive    ServiceStatus = "inactive"
    StatusFailed      ServiceStatus = "failed"
    StatusActivating  ServiceStatus = "activating"
    StatusDeactivating ServiceStatus = "deactivating"
)
```

### MainModel Structure

```go
type MainModel struct {
    // Core state
    services     []Service
    selected     int
    activeView   ViewMode
    
    // Components
    serviceList  *ServiceListComponent
    detailPane   *DetailPaneComponent
    statusBar    *StatusBarComponent
    helpOverlay  *HelpOverlayComponent
    notifications *NotificationManager
    
    // Clients
    client       ServiceClient
    config       *Config
    
    // State
    mode         Mode
    filter       string
    confirmAction *PendingAction
    lastRefresh  time.Time
}
```

---

## Data Flow

### Service List Refresh

```
+--------+     +------------+     +-------------+     +--------+
| Tick   |---->| ListUnits  |---->| UpdateList  |---->| Render |
| (2s)   |     | (Cmd)      |     | (Msg)       |     |        |
+--------+     +------------+     +-------------+     +--------+
                    |
                    v
              +-------------+
              | SystemdClient|
              | (or other)  |
              +-------------+
```

### Action Execution

```
+--------+     +------------+     +-------------+
| Key s  |---->| StartUnit  |---->| ActionMsg   |
| (press)|     | (Cmd)      |     | (pending)   |
+--------+     +------------+     +-------------+
                                        |
                                        v
+--------+     +------------+     +-------------+
| Render |<----| Update     |<----| Result      |
| (status)|    | (status)   |     | (success/fail)|
+--------+     +------------+     +-------------+
```

---

## File Structure (Proposed)

```
systemdManager/
+-- cmd/
|   +-- systemd-tui/
|       +-- main.go              # Entry point
|
+-- internal/
|   +-- client/
|   |   +-- client.go            # ServiceClient interface
|   |   +-- systemd.go           # SystemdClient implementation
|   |   +-- docker.go            # DockerClient implementation
|   |   +-- procfs.go            # ProcfsClient implementation
|   |   +-- client_test.go       # Client tests
|   |
|   +-- config/
|   |   +-- config.go            # Config manager
|   |   +-- config_test.go       # Config tests
|   |
|   +-- ui/
|   |   +-- model.go             # MainModel
|   |   +-- components/
|   |   |   +-- service_list.go  # Service list component
|   |   |   +-- detail_pane.go   # Detail pane component
|   |   |   +-- status_bar.go    # Status bar component
|   |   |   +-- help_overlay.go  # Help overlay
|   |   |   +-- notification.go  # Toast notifications
|   |   |   +-- command_pal.go   # Command palette
|   |   |
|   |   +-- styles/
|   |   |   +-- colors.go        # Color definitions
|   |   |   +-- theme.go         # Theme manager
|   |   |   +-- icons.go         # Icon definitions
|   |   |
|   |   +-- keys/
|   |       +-- bindings.go      # Keybinding definitions
|   |       +-- handlers.go      # Key handlers
|   |
|   +-- service/
|       +-- unit.go              # Service data model
|       +-- group.go             # Service grouping
|
+-- configs/
|   +-- default.yaml             # Default configuration
|   +-- themes/
|       +-- dark.yaml
|       +-- light.yaml
|
+-- docs/
|   +-- DESIGN.md
|   +-- KEYBINDINGS.md
|   +-- STATE_MACHINE.md
|   +-- VISUAL_DESIGN.md
|   +-- ARCHITECTURE.md          # This file
|
+-- go.mod
+-- go.sum
+-- Makefile
+-- README.md
+-- TESTING.md
```

---

## Testing Strategy

### Unit Tests

| Component | Test Focus |
|-----------|------------|
| SystemdClient | Command generation, output parsing |
| ServiceClient | Interface compliance |
| MainModel | State transitions, message handling |
| Components | Rendering, key handling |
| Config | Loading, validation |

### Integration Tests

| Scenario | Test Focus |
|----------|------------|
| Service lifecycle | Start/stop/restart cycle |
| UI refresh | Data updates propagate |
| Error handling | Graceful degradation |
| Configuration | Hot reload |

### E2E Tests

| Scenario | Test Focus |
|----------|------------|
| User workflow | Complete task flows |
| Edge cases | Error recovery |
| Performance | Response times |

---

## Performance Considerations

### Startup Optimization

```
Cold Start (current):
  1. Initialize all components
  2. Fetch all services
  3. Render UI
  Time: ~500ms

Warm Start (proposed):
  1. Initialize UI shell
  2. Render placeholder
  3. Lazy load services
  Time: ~100ms to interactive
```

### Memory Management

```go
// Ring buffer for logs
type LogBuffer struct {
    data    []string
    size    int
    head    int
    count   int
}

// Max 10KB per service
const MaxLogSize = 10 * 1024
```

### Render Optimization

```go
// Differential rendering
func (m *MainModel) View() string {
    // Only recompute changed regions
    if m.lastView == nil || m.viewDirty {
        m.lastView = m.computeView()
        m.viewDirty = false
    }
    return m.lastView
}
```

---

## Security Considerations

### Service Actions

- User services only (no root required)
- Confirmation for destructive actions
- Audit log of all actions

### Configuration

- Validate all config values
- Sanitize file paths
- No arbitrary command execution

### External Integration

- Validate systemd output
- Timeout on all external commands
- Handle malformed responses gracefully
