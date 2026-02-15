# State Machine & Interactions

> Part of [DESIGN.md](./DESIGN.md)

---

## 8. State Machine & Interactions

### 8.1 Application States

```
+-------------------------------------------------------------------------+
|                          APPLICATION STATE                               |
+-------------------------------------------------------------------------+
|                                                                         |
|  +---------+    +---------+    +---------+    +---------+             |
|  | LOADING |--->|  READY  |--->| REFRESH |--->|  READY  |             |
|  +---------+    +----+----+    +---------+    +---------+             |
|                      |                                                  |
|                      v                                                  |
|  +---------+    +---------+    +---------+    +---------+             |
|  |  ERROR  |<---|CONFIRM- |--->| EXECUTE |--->| SUCCESS |             |
|  +---------+    |  ING    |    +---------+    +---------+             |
|                 +---------+                                               |
|                                                                         |
+-------------------------------------------------------------------------+
```

### 8.2 Mode Transitions

```
+-------------------------------------------------------------------------+
|                            MODE TRANSITIONS                              |
+-------------------------------------------------------------------------+
|                                                                         |
|                         +----------+                                   |
|                    +--->|  NORMAL  |<---+                              |
|                    |    +----+-----+    |                              |
|                    |         |          |                              |
|              Esc   |    /    |    :     |   Esc                        |
|              +-----+----+    |    +-----+----+                        |
|              |          |    v    |          |                        |
|              |   +------+-------+------+   |                        |
|              |   |                     |   |                        |
|              |   |  +-------+  +-------+  |   |                        |
|              +---+  | FILTER|  |COMMAND|  +---+                        |
|                  |  +-------+  +-------+  |                           |
|                  |                     |                           |
|                  |   +-------+  +-------+  |                           |
|                  |   |CONFIRM|  |  EDIT |  |                           |
|                  |   +-------+  +-------+  |                           |
|                  +-----------------------+                           |
|                                                                         |
+-------------------------------------------------------------------------+
```

### 8.3 Interaction Flow

```
User Action -> Input Handler -> State Update -> View Render -> Display

                    |
                    v
            +---------------+
            | Is Valid Key? |
            +-------+-------+
                    |
           +--------+--------+
           |                 |
           v                 v
        [Yes]             [No]
           |                 |
           v                 v
    +--------------+   +--------------+
    | Execute      |   | Ignore/Beep  |
    | Action       |   | (if invalid) |
    +------+-------+   +--------------+
           |
           v
    +--------------+
    | Needs        |
    | Confirmation?|
    +------+-------+
           |
    +------+------+
    |             |
    v             v
 [Yes]          [No]
    |             |
    v             v
+---------+  +---------+
| Show    |  | Execute |
| Confirm |  | Immed.  |
| Prompt  |  |         |
+---------+  +---------+
```

### 8.4 Message Types

```go
// Core message types for Bubble Tea architecture

type tickMsg time.Time           // Periodic refresh trigger
type errMsg struct {             // Error propagation
    context string
    err     error
}
type actionResultMsg struct {    // Action completion
    message string
    err     error
}
type logMsg struct {             // Log fetch result
    unit string
    logs string
    err  error
}
type editorFinishedMsg struct {  // Editor completion
    err error
}
```

---

## 9. Edge Cases & Error Handling

### 9.1 Terminal Constraints

| Constraint | Detection | Handling |
|------------|-----------|----------|
| Too small (< 80x24) | On startup, on resize | Show error message with current size |
| No truecolor support | Query `$TERM` and terminfo | Fall back to 256-color palette |
| No mouse support | Query terminal capabilities | Disable mouse, show keyboard-only UI |

### 9.2 System Constraints

| Constraint | Detection | Handling |
|------------|-----------|----------|
| systemctl not found | `exec.LookPath("systemctl")` | Show error, suggest alternatives |
| journalctl not found | `exec.LookPath("journalctl")` | Disable log view, show status only |
| No user services | Empty list from systemctl | Show "No services" with creation hint |
| Service disappeared | 404 on action | Show error, refresh list |

### 9.3 Runtime Errors

| Error | Detection | Recovery |
|-------|-----------|----------|
| Start failed | Non-zero exit code | Show error message, suggest checking logs |
| Stop failed | Non-zero exit code | Show error, ask if force-kill |
| Edit failed | Editor exit code | Show error, offer retry |
| Daemon-reload failed | Non-zero exit code | Show warning, continue |
| Log fetch failed | journalctl error | Show "Logs unavailable" message |

### 9.4 Network/External Errors

| Error | Detection | Recovery |
|-------|-----------|----------|
| D-Bus connection lost | systemctl timeout | Retry 3 times, then show error |
| Service file corrupted | Parse error | Show error, offer edit to fix |
| Permission denied | Exit status 126/127 | Show error with suggestion |

### 9.5 User Input Errors

| Error | Detection | Handling |
|-------|-----------|----------|
| Invalid key sequence | No binding found | Ignore silently |
| Invalid command | Unknown command | Show "Unknown command" with suggestions |
| Invalid service name | Service not found | Show "Service not found" |
| Empty filter | Filter input empty | Show all services |

### 9.6 Error Display Format

```
+-------------------------------------------------------------+
| X Error: Failed to start my-service                         |
|                                                             |
|   Reason: Service file has syntax error                     |
|   Details: Line 15: Unknown key 'ExecStartt'                |
|                                                             |
|   Suggested actions:                                         |
|   * Press 'e' to edit the service file                      |
|   * Run 'journalctl -u my-service' for more details         |
|                                                             |
|   [Press Esc to dismiss]                                    |
+-------------------------------------------------------------+
```

### 9.7 Error Categories & User Actions

| Category | Icon | Color | User Action |
|----------|------|-------|-------------|
| **Recoverable** | ! | Yellow | Retry / Fix |
| **Warning** | ! | Orange | Continue / Abort |
| **Fatal** | X | Red | Exit / Report |
| **Info** | i | Blue | Dismiss |

---

## 10. Performance & Rendering

### 10.1 Flickering Prevention

**Problem:** TUIs flicker when the entire screen is redrawn.

**Solution:** Differential rendering with double buffering.

```go
// Pseudocode
type Renderer struct {
    prevFrame []string
    currFrame []string
}

func (r *Renderer) Render(newContent []string) {
    r.currFrame = newContent
    diffs := r.computeDiffs(r.prevFrame, r.currFrame)
    for _, diff := range diffs {
        r.applyDiff(diff)  // Only update changed cells
    }
    r.prevFrame = r.currFrame
}
```

**Bubble Tea Approach:**
- Use `tea.Batch` for parallel operations
- Avoid blocking in `Update()`
- Use `tea.Tick` for polling, not goroutines + channels

### 10.2 Refresh Strategy

```
+-------------------------------------------------------------+
|                     REFRESH STRATEGY                        |
+-------------------------------------------------------------+
|                                                             |
|  Full Refresh (every 2s):                                   |
|  +-- Service list (all units)                               |
|  +-- Status counts                                          |
|                                                             |
|  Partial Refresh (on action):                               |
|  +-- Selected service status                                |
|  +-- Detail pane content                                    |
|                                                             |
|  Live Mode (when enabled):                                  |
|  +-- Log streaming (every 500ms)                            |
|                                                             |
|  Deferred Refresh:                                          |
|  +-- During filter mode (pause)                             |
|  +-- During confirm mode (pause)                            |
|                                                             |
+-------------------------------------------------------------+
```

### 10.3 Optimizations

| Optimization | Description | Impact |
|--------------|-------------|--------|
| **Diff-based updates** | Only redraw changed cells | -70% redraws |
| **Lazy loading** | Load logs on demand | Faster startup |
| **Caching** | Cache service descriptions | -50% systemctl calls |
| **Debouncing** | Debounce rapid keypresses | Prevents queue buildup |
| **Batching** | Batch multiple updates | Reduces render calls |

### 10.4 Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Startup time | < 500ms | Time to first render |
| Key response | < 16ms (60fps) | Time from keypress to render |
| Full refresh | < 200ms | Time to fetch and render all data |
| Log tail latency | < 100ms | Time for new log to appear |

### 10.5 Memory Management

| Scenario | Strategy |
|----------|----------|
| **Large log buffers** | Ring buffer with max 10KB per service |
| **Many services** | Lazy load details, cache visible items only |
| **Long-running sessions** | Periodic cleanup of stale cache entries |
