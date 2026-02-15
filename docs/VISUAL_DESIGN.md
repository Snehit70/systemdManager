# Typography, Colors & Visual Design

> Part of [DESIGN.md](./DESIGN.md)

---

## 11. Typography & Visual Design

### 11.1 Font Recommendations

**Primary Font:** Monospace font with good Unicode support

| Font | Why | Use Case |
|------|-----|----------|
| **JetBrains Mono** | Excellent readability, clear zero/O | Default recommendation |
| **Fira Code** | Ligatures, good Unicode | For code-heavy content |
| **Nerd Fonts** | Icon support | Extended icons |
| **IBM Plex Mono** | Corporate, clean | Minimal setups |

**Fallback Chain:**
```
JetBrains Mono -> Fira Code -> DejaVu Sans Mono -> Monospace
```

### 11.2 Font Sizes

| Element | Relative Size | Actual (at 12pt base) |
|---------|---------------|----------------------|
| Service name | 1.0em | 12pt |
| Description | 0.9em | 11pt |
| Status indicators | 1.0em | 12pt |
| Log content | 0.9em | 11pt |
| Help text | 0.85em | 10pt |

**Note:** Terminal fonts are typically not resizable via the application. Users set their terminal font size.

### 11.3 Text Styling

| Style | Usage | Example |
|-------|-------|---------|
| **Bold** | Selected item, headers | `**my-service**` |
| **Dim** | Inactive services, timestamps | `inactive` |
| **Underline** | Links, hotkeys in help | `<u>s</u>tart` |
| **Reverse** | Selection in lists | `[selected]` |

### 11.4 Icon System

**Status Icons (Unicode):**

| Icon | Unicode | Meaning | Alternative (ASCII) |
|------|---------|---------|---------------------|
| * | U+25CF | Active | * |
| o | U+25CB | Inactive | o |
| X | U+2717 | Failed | X |
| @ | U+27F3 | Activating | @ |
| - | U+23F8 | Deactivating | - |
| R | U+21BB | Reloading | R |

**Action Icons:**

| Icon | Unicode | Meaning |
|------|---------|---------|
| > | U+25B6 | Start |
| # | U+25A0 | Stop |
| ~ | U+21BA | Restart |
| OK | U+2713 | Success |
| X | U+2717 | Error |
| ! | U+26A0 | Warning |
| i | U+2139 | Info |

---

## 12. Color System

### 12.1 Primary Palette

```
STATUS COLORS:
  Active:    #73F59F  (Green)   - Running services
  Failed:    #FF6B6B  (Red)     - Failed services
  Inactive:  #666666  (Gray)    - Stopped services
  Degraded:  #FFA500  (Orange)  - Transitioning states

UI COLORS:
  Primary:   #5B9BF3  (Blue)    - Selection, links
  Secondary: #8B8B8B  (Gray)    - Borders, dim text
  Background:#1A1A2E  (Dark)    - Background (dark theme)
  Surface:   #16213E  (Dark)    - Cards, panels
  Text:      #E8E8E8  (Light)   - Primary text
  Muted:     #6B6B6B  (Gray)    - Secondary text
```

### 12.2 Color Accessibility

| Combination | Contrast Ratio | WCAG AA | WCAG AAA |
|-------------|----------------|---------|----------|
| Active on Dark | 8.2:1 | Pass | Pass |
| Failed on Dark | 4.8:1 | Pass | Fail |
| Inactive on Dark | 5.1:1 | Pass | Fail |
| Primary on Dark | 6.9:1 | Pass | Pass |
| Text on Dark | 12.5:1 | Pass | Pass |

### 12.3 Theme Variants

#### Dark Theme (Default)

```
Background:    #1A1A2E
Surface:       #16213E
Border:        #2D3A5C
BorderActive:  #5B9BF3
Text:          #E8E8E8
TextMuted:     #6B6B6B
```

#### Light Theme

```
Background:    #FAFAFA
Surface:       #FFFFFF
Border:        #E0E0E0
BorderActive:  #1976D2
Text:          #1A1A1A
TextMuted:     #757575
```

#### High Contrast Theme

```
Background:    #000000
Surface:       #1A1A1A
Border:        #FFFFFF
BorderActive:  #00FF00
Text:          #FFFFFF
TextMuted:     #AAAAAA
Active:        #00FF00
Failed:        #FF0000
```

---

## 13. Animation & Transitions

### 13.1 Animation Principles

1. **Purposeful**: Every animation has meaning
2. **Fast**: Animations should be 100-300ms
3. **Subtle**: Don't distract from content
4. **Consistent**: Same transitions for same actions

### 13.2 Animation Types

| Animation | Trigger | Duration | Easing |
|-----------|---------|----------|--------|
| **Fade In** | Notification appear | 150ms | ease-out |
| **Fade Out** | Notification dismiss | 100ms | ease-in |
| **Slide** | Panel switch | 200ms | ease-in-out |
| **Flash** | Status change | 300ms | linear |
| **Pulse** | Loading state | 1s loop | ease-in-out |

### 13.3 Status Change Animation

```
Time    0ms    100ms   200ms   300ms
        +------+-------+-------+
Icon    o  ->  @  ->  @  ->  *
Color   Gray   Orange  Orange  Green

Transition shows: Inactive -> Activating -> Active
```

### 13.4 Reducing Motion

For users sensitive to motion:

```yaml
# config.yaml
ui:
  reduce_motion: true  # Disables all animations
```

When enabled:
- No fade/slide animations
- Instant state changes
- Static loading indicators

---

## 14. Logging & Debugging

### 14.1 Log Levels

| Level | Use Case | Output |
|-------|----------|--------|
| DEBUG | Development, verbose | File only |
| INFO | Normal operations | File + Console |
| WARN | Recoverable issues | File + Console |
| ERROR | Failures | File + Console |
| FATAL | Application crash | All outputs |

### 14.2 Log Format

```
[2026-02-15 10:30:00] [INFO]  Service list refreshed: 12 units
[2026-02-15 10:30:01] [DEBUG] Fetching logs for: my-service
[2026-02-15 10:30:02] [WARN]  Slow refresh: 250ms (threshold: 200ms)
[2026-02-15 10:30:03] [ERROR] Failed to start redis: exit status 1
```

### 14.3 Debug Mode

Enable with flag: `./systemd-tui --debug`

Additional outputs:
- Key press events
- State transitions
- Render timing
- Memory usage

### 14.4 Log File Location

```
~/.local/share/systemd-tui/logs/
  |-- systemd-tui.log        (current session)
  |-- systemd-tui.1.log      (previous session)
  |-- systemd-tui.2.log      (older session)
```

Retention: Keep last 5 sessions, max 10MB total

---

## 15. Configuration System

### 15.1 Config File Location

```
~/.config/systemd-tui/config.yaml
```

### 15.2 Configuration Schema

```yaml
# General settings
general:
  refresh_interval: 2s        # Service list refresh rate
  log_lines: 50               # Default log lines to fetch
  editor: ""                  # Override $EDITOR (empty = use env)

# UI settings
ui:
  theme: dark                 # dark | light | high-contrast
  reduce_motion: false        # Disable animations
  show_hidden: false          # Show hidden services
  
# Keybinding overrides
keybindings:
  navigate_up:
    - "k"
    - "up"
  navigate_down:
    - "j"
    - "down"
  start_service: "s"
  stop_service: "x"
  restart_service: "r"
  
# Service groups (custom)
groups:
  development:
    - "postgres-*"
    - "redis-*"
    - "nginx-*"
  background:
    - "syncthing"
    - "dropbox"
```

### 15.3 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SYSTEMD_TUI_CONFIG` | Custom config path | `~/.config/systemd-tui/config.yaml` |
| `SYSTEMD_TUI_DEBUG` | Enable debug logging | `false` |
| `SYSTEMD_TUI_THEME` | Override theme | (from config) |
| `EDITOR` | Editor for service files | `vim` |

---

## 16. Accessibility Considerations

### 16.1 Keyboard Navigation

- All features accessible via keyboard
- Logical tab order
- Escape always returns to safe state
- Confirmation prompts for destructive actions

### 16.2 Visual Accessibility

| Feature | Implementation |
|---------|----------------|
| **High Contrast Mode** | Distinct color palette option |
| **Status Icons** | Always paired with text status |
| **Color Blindness** | Never rely on color alone |
| **Focus Indicators** | Clear visual focus on all elements |

### 16.3 Screen Reader Support

Terminal applications have limited screen reader support, but we can:

- Provide structured text output
- Use semantic ordering
- Document keyboard shortcuts clearly

### 16.4 Reduced Motion

```yaml
ui:
  reduce_motion: true
```

Disables:
- Fade animations
- Slide transitions
- Pulsing indicators

---

## 17. Future Roadmap

### Phase 1: Foundation (v2.0)

- [ ] Multi-panel layout with tabs
- [ ] Mode-based navigation
- [ ] Command palette
- [ ] Configuration system
- [ ] Theme support

### Phase 2: Enhancement (v2.1)

- [ ] Live log tailing
- [ ] Service grouping
- [ ] Bulk operations
- [ ] Service templates
- [ ] Search history

### Phase 3: Integration (v2.2)

- [ ] Multi-backend support (Docker, procfs)
- [ ] SSH remote management
- [ ] REST API for scripting
- [ ] Plugin system

### Phase 4: Advanced (v3.0)

- [ ] Dependency visualization
- [ ] Metrics dashboard
- [ ] Alert system
- [ ] Web interface
- [ ] Mobile companion app

---

## Appendix A: Current Implementation Analysis

### Strengths

1. **Clean Architecture**: Separation of service and UI layers
2. **Bubble Tea Pattern**: Proper use of Elm architecture
3. **Status Colors**: Clear visual indicators
4. **Confirmation Flow**: Safe handling of destructive actions
5. **Auto-refresh**: Keeps UI in sync with system state

### Weaknesses

1. **No Interface Abstraction**: Direct coupling to SystemdClient
2. **Limited Error Context**: Generic error messages
3. **Single Backend**: Only systemd user services
4. **No Configuration**: Hardcoded behavior
5. **Missing Tests**: Limited test coverage for edge cases

### Technical Debt

| Item | Impact | Effort |
|------|--------|--------|
| Add ServiceClient interface | High | Low |
| Improve error messages | Medium | Low |
| Add configuration support | High | Medium |
| Write integration tests | Medium | Medium |
| Add log pagination | Medium | Medium |
