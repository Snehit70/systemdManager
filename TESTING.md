# Manual Testing Guide

This document describes manual testing scenarios for the systemd TUI manager.

## Prerequisites

- Linux system with systemd
- At least one user service available (`systemctl --user list-units --type=service`)
- Terminal with 80x24 minimum size

## Build & Run

```bash
go build -o systemd-tui ./cmd/systemd-tui
./systemd-tui
```

## Test Scenarios

### 1. Application Launch

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Clean start | Run `./systemd-tui` | TUI displays with service list on left, detail pane on right |
| No services | (If no user services exist) | Shows empty list or "No services found" |
| systemctl unavailable | Run on non-systemd system | Graceful error message |

### 2. Navigation

| Test | Steps | Expected Result |
|------|-------|-----------------|
| List navigation | Press `j`/`k` or arrow keys | Selection moves up/down in service list |
| Focus switch | Press `Tab` | Border color changes, focus moves between list/detail |
| Detail scroll | Focus detail pane, use `j`/`k` | Viewport scrolls through log content |
| Help toggle | Press `?` | Help panel expands/collapses |
| Quit | Press `q` | Application exits cleanly |

### 3. Text Filtering

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Enter filter | Press `/` | Filter input appears |
| Filter services | Type partial name | List filters to matching services |
| Clear filter | Press `Esc` in filter mode | Filter clears, full list restored |
| No matches | Filter with non-existent name | "No matching services" or empty list |

### 4. Source Filtering (NEW)

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Cycle filter | Press `F` | Status bar shows filter mode: "all" → "my services" → "hide system" → "all" |
| My services | Filter mode "my services" | Only shows services created by user (● indicator) |
| Hide system | Filter mode "hide system" | Hides static, generated, transient, and system services |

### 5. Source Indicators (NEW)

| Test | Steps | Expected Result |
|------|-------|-----------------|
| User service | Look at service you created | Shows `●` (filled circle) before name |
| System service | Look at system-provided service | Shows `○` (empty circle) before name |
| Unknown source | Services without clear source | Shows `·` (dot) before name |

### 6. Service Creation (NEW)

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Open modal | Press `c` | Create service modal appears centered |
| Navigate fields | Press `Tab` / `Shift+Tab` | Focus moves between fields |
| Toggle type | Press `t` when Type field focused | Cycles: simple → oneshot → simple |
| Cycle restart | Press `r` when Restart field focused | Cycles: on-failure → always → no → on-failure |
| Cancel creation | Press `Esc` | Modal closes, no service created |
| Create minimal | Name: "test", Command: "/bin/echo test", Enter | Service created, status shows success |
| Create with all fields | Fill all fields, Enter | Service created with all settings |
| Validation error | Leave Name empty, press Enter | Shows "Error: name and command are required" |

### 7. Service Actions

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Start service | Select inactive service, press `s` | Status message shows "Starting...", service starts |
| Stop service | Select active service, press `x` | Confirmation prompt appears |
| Stop confirm | Press `y` at confirmation | Service stops, status updates |
| Stop cancel | Press `n` or `Esc` at confirmation | "Cancelled" message, no action |
| Restart service | Select service, press `r` | Confirmation, then service restarts |
| Enable service | Select service, press `E` | Confirmation, then service enabled |
| Disable service | Select service, press `D` | Confirmation, then service disabled |

### 8. Edit Functionality

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Edit service | Select service, press `e` | Editor opens with service file |
| Save and exit | Edit, save, exit editor | "Edit saved. Reloaded daemon." message |
| Cancel edit | Exit editor without saving | No daemon reload |
| EDITOR variable | Set `EDITOR=nano`, press `e` | Opens in nano instead of vim |

### 9. Visual Indicators

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Active service | View active service in list | Name displayed in green |
| Failed service | View failed service in list | Name displayed in red |
| Inactive service | View inactive service in list | Name displayed in gray |
| Selected item | Navigate to item | Shows `>` prefix with bold styling |

### 10. Log Viewing

| Test | Steps | Expected Result |
|------|-------|-----------------|
| View logs | Select any service | Detail pane shows last 50 lines of logs |
| Follow mode | Press `f` | Shows "[FOLLOWING - Press f to stop]", logs update live |
| Stop follow | Press `f` again | Stops following, shows static logs |

### 11. Grouping

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Cycle grouping | Press `g` | Groups: none → status → load → none |
| Status grouping | Group by status | Services grouped under "Active", "Failed", "Inactive" headers |
| Load grouping | Group by load | Services grouped under "Loaded", "Not Found", "Other" headers |

### 12. Auto-refresh

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Polling | Wait 2+ seconds after action | List refreshes automatically |
| External change | Change service state externally | TUI reflects change within 2 seconds |

### 13. Error Handling

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Start failure | Try starting already-active service | Error shown in status bar |
| Stop failure | Try stopping already-stopped service | Error shown in status bar |
| Log fetch error | Select service with no logs | "Log Error" message in detail pane |
| Permission denied | Try action without permission | Specific error message displayed |
| Duplicate service | Create service with existing name | Shows error "service already exists" |
| Invalid name | Create service with "../" in name | Shows error "invalid service name" |
| Newline injection | Create service with newline in description | Shows error "description contains newlines" |

### 14. Edge Cases

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Small terminal | Resize to < 80x24 | Layout adjusts or shows minimum size warning |
| Long service name | Service with 50+ char name | Name truncated with ellipsis |
| Long description | Service with long description | Description truncated at 47 chars + "..." |
| Rapid keypresses | Spam action keys quickly | Actions queued, no crash |

## Automated Tests

Run unit tests with:

```bash
go test ./...
```

Current test coverage:
- `internal/service`: JSON parsing, unit struct validation
- `internal/ui`: Model initialization, confirmation state, error handling, window sizing

## Testing Script

Quick automated verification:

```bash
# Build
go build -o systemd-tui ./cmd/systemd-tui

# Run unit tests
go test ./...

# Check services available
systemctl --user list-units --type=service --all | head -10

# Verify user service directory exists
ls ~/.config/systemd/user/

# Create a test service manually to verify creation works
cat > ~/.config/systemd/user/test-manual.service << 'EOF'
[Unit]
Description=Manual Test Service

[Service]
Type=simple
ExecStart=/bin/sleep infinity

[Install]
WantedBy=default.target
EOF

# Reload and verify
systemctl --user daemon-reload
systemctl --user list-unit-files test-manual.service

# Clean up
rm ~/.config/systemd/user/test-manual.service
systemctl --user daemon-reload
```

## Known Limitations

1. **System services**: Only user services (`--user`) are shown
2. **Terminal required**: Cannot run non-interactively (no `--help` flag)

## Test Results Log

### [Date: 2026-02-15]

**Build Status**: ✅ Pass
**Unit Tests**: ✅ Pass (internal/service, internal/ui)

**Manual Testing**: _Pending user verification_

Features to verify manually:
- [ ] Service creation modal opens and creates services
- [ ] Filter modes cycle correctly (F key)
- [ ] Source indicators display properly (●/○)
- [ ] Created services appear in "my services" filter mode
