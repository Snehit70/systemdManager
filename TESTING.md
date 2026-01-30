# Manual Testing Guide

This document describes manual testing scenarios for the systemd TUI manager.

## Prerequisites

- Linux system with systemd
- At least one user service available (`systemctl --user list-units --type=service`)
- Terminal with 80x24 minimum size

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

### 3. Filtering

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Enter filter | Press `/` | Filter input appears |
| Filter services | Type partial name | List filters to matching services |
| Clear filter | Press `Esc` in filter mode | Filter clears, full list restored |
| No matches | Filter with non-existent name | "No matching services" or empty list |

### 4. Service Actions

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Start service | Select inactive service, press `s` | Status message shows "Starting...", service starts |
| Stop service | Select active service, press `x` | Confirmation prompt appears |
| Stop confirm | Press `y` at confirmation | Service stops, status updates |
| Stop cancel | Press `n` or `Esc` at confirmation | "Cancelled" message, no action |
| Restart service | Select service, press `r` | Confirmation, then service restarts |
| Enable service | Select service, press `E` | Confirmation, then service enabled |
| Disable service | Select service, press `D` | Confirmation, then service disabled |

### 5. Edit Functionality

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Edit service | Select service, press `e` | Editor opens with service file |
| Save and exit | Edit, save, exit editor | "Edit saved. Reloaded daemon." message |
| Cancel edit | Exit editor without saving | No daemon reload |
| EDITOR variable | Set `EDITOR=nano`, press `e` | Opens in nano instead of vim |

### 6. Visual Indicators

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Active service | View active service in list | Name displayed in green |
| Failed service | View failed service in list | Name displayed in red |
| Inactive service | View inactive service in list | Name displayed in gray |
| Selected item | Navigate to item | Shows `>` prefix with bold styling |

### 7. Auto-refresh

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Polling | Wait 2+ seconds after action | List refreshes automatically |
| External change | Change service state externally | TUI reflects change within 2 seconds |

### 8. Error Handling

| Test | Steps | Expected Result |
|------|-------|-----------------|
| Start failure | Try starting already-active service | Error shown in status bar |
| Stop failure | Try stopping already-stopped service | Error shown in status bar |
| Log fetch error | Select service with no logs | "Log Error" message in detail pane |
| Permission denied | Try action without permission | Specific error message displayed |

### 9. Edge Cases

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

## Known Limitations

1. **Live log tailing**: Not yet implemented (logs are snapshot, not streaming)
2. **System services**: Only user services (`--user`) are shown
3. **Service grouping**: Not yet implemented (Phase 9)
4. **Clipboard support**: Not yet implemented (Phase 9)
5. **Config file**: Not yet implemented (Phase 9)
