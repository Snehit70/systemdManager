# Systemd TUI Manager

A terminal user interface for managing `systemd --user` services, inspired by LazyGit's intuitive workflow.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## Features

- **Service Management**: Start, stop, restart, enable, disable user services
- **Service Creation**: Create new services with a simple modal form
- **Smart Filtering**: Filter by source (all/my services/hide system services)
- **Live Logs**: View and follow service logs in real-time
- **Source Indicators**: Visual distinction between user-created (●) and system services (○)
- **Vim-style Navigation**: `j/k` to navigate, familiar keybindings
- **Beautiful UI**: Split-pane layout with color-coded status indicators
- **Themes**: Dark, light, and high-contrast theme support

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/systemd-tui
cd systemd-tui

# Build
go build -o systemd-tui ./cmd/systemd-tui

# Install to PATH (optional)
sudo mv systemd-tui /usr/local/bin/
```

## Usage

```bash
# Run the TUI
./systemd-tui

# Or if installed to PATH
systemd-tui
```

## Keybindings

| Key | Action |
|-----|--------|
| `j/k` or `↑/↓` | Navigate list |
| `Tab` | Switch focus (list/detail) |
| `/` | Filter services |
| `s` | Start service |
| `x` | Stop service |
| `r` | Restart service |
| `e` | Edit service file |
| `E` | Enable service |
| `D` | Disable service |
| `c` | **Create new service** |
| `F` | **Cycle filter mode** (all/my services/hide system) |
| `g` | Toggle grouping |
| `f` | Toggle follow logs |
| `?` | Toggle help |
| `q` | Quit |

### Create Service Modal

When creating a service (`c`):

| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `t` | Toggle type (simple/oneshot) |
| `r` | Cycle restart policy |
| `Enter` | Create service |
| `Esc` | Cancel |

## Creating Services

Press `c` to open the service creation modal:

1. **Name**: Service name (e.g., `myapp` - `.service` appended automatically)
2. **Command**: Full path to executable with args (e.g., `/usr/bin/node /home/user/app.js`)
3. **Description**: Optional description
4. **Working Directory**: Optional working directory (defaults to `~`)
5. **Type**: `simple` (default) or `oneshot`
6. **Restart**: `on-failure` (default), `always`, or `no`

The service file is created in `~/.config/systemd/user/` and `daemon-reload` is run automatically.

## Filtering

Press `F` to cycle through filter modes:

- **all**: Show all services
- **my services**: Show only services you created in `~/.config/systemd/user/`
- **hide system**: Hide static/generated/transient services

## Configuration

Config file: `~/.config/systemd-tui/config.yaml`

```yaml
general:
  refresh_interval: 2s
  log_lines: 50
  editor: vim  # or $EDITOR

ui:
  theme: dark  # dark, light, high-contrast
  reduce_motion: false
  show_hidden: false
```

## Requirements

- Linux with systemd
- Go 1.25+ (for building)
- `systemctl` and `journalctl` in PATH

## Development

```bash
# Run tests
go test ./...

# Run with race detector
go run -race ./cmd/systemd-tui

# Build release binary
go build -ldflags="-s -w" -o systemd-tui ./cmd/systemd-tui
```

## Architecture

- **Bubble Tea**: TUI framework using The Elm Architecture
- **Split Layout**: List pane (services) + Detail pane (logs/status)
- **ServiceClient Interface**: Abstraction for service operations
- **Polling**: 2-second refresh interval

See [docs/](./docs/) for detailed design documentation.

## License

MIT
