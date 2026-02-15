# Systemd TUI Manager - Comprehensive Design Document

> **Version**: 2.0 (Redesign)  
> **Last Updated**: February 2026  
> **Status**: Planning Phase

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Project Vision & Scope](#2-project-vision--scope)
3. [Design Philosophy](#3-design-philosophy)
4. [User Interface Architecture](#4-user-interface-architecture)
5. [Detailed UI Specifications](#5-detailed-ui-specifications)
6. [Keybinding System](#6-keybinding-system)
7. [User Workflow Analysis](#7-user-workflow-analysis)
8. [State Machine & Interactions](#8-state-machine--interactions)
9. [Edge Cases & Error Handling](#9-edge-cases--error-handling)
10. [Performance & Rendering](#10-performance--rendering)
11. [Typography & Visual Design](#11-typography--visual-design)
12. [Color System](#12-color-system)
13. [Animation & Transitions](#13-animation--transitions)
14. [Logging & Debugging](#14-logging--debugging)
15. [Configuration System](#15-configuration-system)
16. [Accessibility Considerations](#16-accessibility-considerations)
17. [Future Roadmap](#17-future-roadmap)

---

## 1. Executive Summary

### Current State

The systemd-tui project is a functional Terminal User Interface for managing systemd user services. Built with Go and the Bubble Tea framework, it provides:

- **Service List**: Scrollable, filterable list of user services
- **Detail Pane**: Shows logs and status information
- **Actions**: Start, Stop, Restart, Enable, Disable, Edit
- **Status Colors**: Green (active), Red (failed), Gray (inactive)
- **Auto-refresh**: 2-second polling interval

### Identified Gaps

| Area | Current State | Desired State |
|------|---------------|---------------|
| **Layout** | Basic split view | Lazygit-inspired multi-panel with tabs |
| **Navigation** | Single-mode | Mode-based (normal, filter, command) |
| **Feedback** | Status bar only | Rich notifications + status indicators |
| **Performance** | Basic polling | Intelligent diffing + selective updates |
| **Extensibility** | Hardcoded | Plugin system + config files |
| **Error Handling** | Basic messages | Structured error display with suggestions |

### Redesign Goals

1. **Intuitive UX**: Zero-learning-curve for vim/lazygit users
2. **Rich Feedback**: Every action has immediate, clear visual response
3. **Performance**: Smooth 60fps rendering, no flickering
4. **Extensibility**: Config-driven keybindings, themes, behaviors
5. **Robustness**: Handle all edge cases gracefully

---

## 2. Project Vision & Scope

### Vision Statement

> "A beautiful, fast, and intuitive terminal interface for managing services — whether systemd user services, local processes, or custom service managers. One tool, multiple backends."

### Scope Definition

#### In Scope (v2.0)

| Feature | Description | Priority |
|---------|-------------|----------|
| **Core Service Management** | Start, stop, restart, enable, disable services | P0 |
| **Multi-backend Support** | systemd, procfs (local processes), Docker | P1 |
| **Live Log Streaming** | Real-time log tail with filtering | P0 |
| **Service Groups** | Group services by project, type, or custom tags | P1 |
| **Configuration** | YAML config for keybindings, themes, behaviors | P1 |
| **Command Palette** | Fuzzy-searchable command execution | P2 |
| **Service Templates** | Create new services from templates | P2 |

#### Out of Scope (v2.0)

- System-level services (requires root, security implications)
- Remote service management (SSH integration)
- GUI/Web interface
- Service dependency visualization
- Metrics and alerting

### Target Users

| User Type | Description | Key Needs |
|-----------|-------------|-----------|
| **Developer** | Manages dev services (databases, queues, etc.) | Quick start/stop, logs, multiple services |
| **DevOps** | Manages local environments | Bulk operations, templates, configurations |
| **Power User** | Custom setups, scripts | CLI integration, extensibility, scripting |

---

## 3. Design Philosophy

### Core Principles

#### 1. **Keyboard-First**

Every action must be achievable via keyboard. Mouse support is optional.

```
Principle: "The keyboard is 10x faster than the mouse for repeated tasks."
Reference: Vim, Lazygit, Htop, Btop
```

#### 2. **Immediate Feedback**

Every user action must have visible feedback within 100ms.

```
Principle: "If the user can't see it happened, it didn't happen."
Implementation: Optimistic UI updates, loading states, success/error indicators
```

#### 3. **Progressive Disclosure**

Simple things simple, complex things possible.

```
Principle: "Don't show everything at once. Reveal complexity on demand."
Implementation: 
- Default view: Essential info only
- Detail view: Full status, logs, config
- Command mode: All actions available
```

#### 4. **Predictable Behavior**

Same action, same result. No surprises.

```
Principle: "Muscle memory is sacred. Never break it."
Implementation: Consistent keybindings, stable UI layout
```

#### 5. **Graceful Degradation**

Work with what's available.

```
Principle: "The tool should work everywhere, even with limitations."
Implementation:
- No truecolor? Fall back to 256 colors
- No journalctl? Show basic status
- Small terminal? Adaptive layout
```

### Inspiration Sources

| Application | What We Learn |
|-------------|---------------|
| **Lazygit** | Split-panel layout, staging workflow, contextual help |
| **Htop/Btop** | Real-time updates, process management, visual indicators |
| **K9s** | Resource navigation, command mode, contextual actions |
| **Vim/Neovim** | Modal editing, command palette, extensibility |
| **FZF** | Fuzzy finding, preview window |

---

## 4. User Interface Architecture

### Layout Overview

```
+-------------------------------------------------------------------------+
|  Systemd TUI Manager                                    [help:?] [quit:q]|
+-------------------------------------------------------------------------+
| +---------------------+ +---------------------------------------------+ |
| | Services (12)       | | Details: my-service.service                 | |
| | ------------------- | | ------------------------------------------ | |
| | * my-service    act | | Status: active (running)                    | |
| | o db-postgres   ina | | Since: 2026-02-15 10:30:00 UTC              | |
| | X redis-server  fai | | PID: 12345                                   | |
| | * nginx-proxy   act | | Memory: 128 MB                               | |
| | o app-worker    ina | | CPU: 2.3%                                    | |
| |                     | |                                              | |
| | [2 more below...]   | | +------------------------------------------+ | |
| |                     | | | Logs (last 50 lines)                     | | |
| |                     | | |------------------------------------------| | |
| |                     | | | Feb 15 10:30:01 Starting service...      | | |
| |                     | | | Feb 15 10:30:02 Listening on port 8080   | | |
| |                     | | | Feb 15 10:30:03 Connection established   | | |
| |                     | | | ...                                      | | |
| |                     | | +------------------------------------------+ | |
| +---------------------+ |                                              | |
|                         | [Tab: switch] [e: edit] [l: live logs]       | |
+-------------------------------------------------------------------------+
| * active  o inactive  X failed  @ activating  Status updated 2s ago   |
+-------------------------------------------------------------------------+
| s:start x:stop r:restart E:enable D:disable e:edit /:filter ?:help     |
+-------------------------------------------------------------------------+
```

### Panel Structure

| Panel | Width | Content | Focus |
|-------|-------|---------|-------|
| **Service List** | 30% | Service names, status icons, descriptions | Primary |
| **Detail Pane** | 70% | Status info, metrics, logs, actions | Secondary |

### Responsive Layout

| Terminal Width | Layout |
|----------------|--------|
| >= 120 cols | Split view (30/70) |
| 80-119 cols | Split view (40/60) |
| < 80 cols | Single panel (toggle with Tab) |
| < 80x24 | Warning message, refuse to start |

### View Modes

```
Mode: NORMAL (default)
+-- j/k: Navigate services
+-- s/x/r: Service actions
+-- /: Enter FILTER mode
+-- :: Enter COMMAND mode
+-- ?: Toggle HELP overlay

Mode: FILTER
+-- type: Filter services
+-- Enter: Apply filter, return to NORMAL
+-- Esc: Clear filter, return to NORMAL
+-- Tab: Switch to detail panel

Mode: COMMAND
+-- type: Enter command
+-- Tab: Autocomplete
+-- Enter: Execute command
+-- Esc: Return to NORMAL

Mode: CONFIRM
+-- y: Confirm action
+-- n/Esc: Cancel action
+-- (locked until response)
```

---

## 5. Detailed UI Specifications

### 5.1 Service List Panel

#### Visual Structure

```
+-------------------------+
| Services (12) [filter]  |  <- Header with count and filter indicator
+-------------------------+
| * my-service      active|  <- Status icon + name + state
|   Long description...   |  <- Truncated description (gray)
|                         |
| o db-postgres    inactive|
|   PostgreSQL database   |
|                         |
| X redis-server     failed|  <- Failed service (red highlight)
|   Redis cache server    |
|                         |
|   ... more services     |
+-------------------------+
```

#### Status Icons

| Icon | State | Color | Description |
|------|-------|-------|-------------|
| `*` | active | Green (#73F59F) | Service is running |
| `o` | inactive | Gray (#666666) | Service is stopped |
| `X` | failed | Red (#FF6B6B) | Service failed to start/run |
| `@` | activating | Orange (#FFA500) | Service is starting |
| `-` | deactivating | Orange (#FFA500) | Service is stopping |
| `R` | reloading | Orange (#FFA500) | Service is reloading config |

#### List Item Rendering

```go
type ListItem struct {
    Unit        string
    Description string
    Active      string
    Sub         string
    Load        string
    
    // Rendering metadata
    IsSelected  bool
    IsFiltered  bool
    MatchScore  float64  // For fuzzy matching
}
```

#### Truncation Rules

| Field | Max Width | Overflow Handling |
|-------|-----------|-------------------|
| Unit name | 20 chars | Truncate left: `...ice.service` |
| Description | 40 chars | Truncate right with ellipsis |
| Combined | Panel width - 4 | Adaptive |

### 5.2 Detail Pane

#### Information Hierarchy

```
+---------------------------------------------+
| my-service.service                          | <- Title (bold)
+---------------------------------------------+
| Status: active (running)    Since: 10:30:00 | <- Primary info
| PID: 12345  Memory: 128MB  CPU: 2.3%       | <- Metrics
+---------------------------------------------+
| [Status] [Logs] [Config] [Dependencies]     | <- Tab navigation
+---------------------------------------------+
|                                             |
| (Tab content here)                          |
|                                             |
+---------------------------------------------+
```

#### Tabs

| Tab | Content | Refresh |
|-----|---------|---------|
| **Status** | Full systemctl status output | On demand |
| **Logs** | Journal logs (live or snapshot) | Auto (live mode) |
| **Config** | Service file content | On load |
| **Dependencies** | Before/After/Wants relationships | On demand |

#### Log View

```
+-----------------------------------------------+
| Logs (last 50 lines) [live: off] [filter: -]  |
+-----------------------------------------------+
| 10:30:01 [INFO] Starting service...           |
| 10:30:02 [INFO] Listening on port 8080        |
| 10:30:03 [WARN] High memory usage detected    |
| 10:30:04 [ERROR] Connection timeout           |
| ...                                           |
+-----------------------------------------------+
| [l: live] [f: filter] [c: copy] [/: search]   |
+-----------------------------------------------+
```

**Log Features:**
- Timestamp highlighting
- Log level colors (INFO=blue, WARN=yellow, ERROR=red)
- Fuzzy search within logs
- Copy to clipboard
- Live tail mode (follow new logs)

### 5.3 Status Bar

#### Structure

```
+-----------------------------------------------------------------+
| * 12 active  o 3 inactive  X 1 failed  | Last refresh: 2s ago  |
+-----------------------------------------------------------------+
```

**Left Section:** Quick stats (counts by state)  
**Right Section:** Last refresh time, connection status

### 5.4 Help Bar

#### Structure

```
+-----------------------------------------------------------------+
| s:start x:stop r:restart E:enable D:disable e:edit /:filter ?:help|
+-----------------------------------------------------------------+
```

**Behavior:**
- Shows most relevant shortcuts for current context
- Updates based on active panel/mode
- Full help overlay on `?`

### 5.5 Notification System

#### Toast Notifications

```
+--------------------------------------+
| OK Service started successfully       |
+--------------------------------------+
```

**Types:**

| Type | Icon | Duration | Color |
|------|------|----------|-------|
| Success | OK | 3s | Green |
| Error | X | 5s | Red |
| Warning | ! | 4s | Yellow |
| Info | i | 2s | Blue |

**Position:** Top-right corner, stacked (max 3 visible)
