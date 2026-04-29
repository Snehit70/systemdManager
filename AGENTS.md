# Systemd TUI Manager — Agent Notes

Operational notes for AI coding agents (and humans) working on this repo. The user-facing README lives at [`README.md`](./README.md); deeper design docs live under [`docs/`](./docs/).

## What this is

A Go + Bubble Tea TUI for `systemd --user`. Single binary, no daemon, no privileged ops — every action shells out to `systemctl --user` / `journalctl --user`. The tool is read-mostly with explicit confirmation for destructive actions.

## Layout

```
cmd/systemd-tui/main.go          entry point; mounts the bubbletea program with mouse + alt-screen
internal/client/                  ServiceClient interface, Service / ServiceTemplate types, error helpers
internal/service/                 systemctl-backed implementation of ServiceClient
internal/config/                  YAML config loader; Mocha / Dark / Light / High-Contrast palettes
internal/ui/
  model.go                        MainModel: state, keybindings, Update loop
  view.go                         View(), modal, status bar, help bar, detail header, helpers
  delegate.go                     Custom list row + group header rendering
  styles.go                       Theme-aware lipgloss styles applied at startup
  log_styling.go                  Heuristic recoloring of log lines (error/warn/info/debug)
  commands.go                     Async tea.Cmd builders (fetch, follow, actions)
docs/                             Design notes (some predate the current UI; treat as historical)
```

## Architecture invariants

- **Interface-first**. `ServiceClient` (in `internal/client`) is the only contract the UI depends on. Compile-time assertion lives next to the implementation. Adding a new backend means satisfying that interface, nothing else.
- **Context propagation**. Every `ServiceClient` method takes `context.Context` as its first argument. The UI threads its own context through and cancels follow-mode loops via the cancel function returned from `FollowLogs`.
- **Bubble Tea purity**. Update is the only place state mutates. Commands return `tea.Msg`s; views never trigger I/O. The model is passed by value so every Update returns a new copy.
- **No package-level state in UI** apart from the lipgloss style cache populated by `ApplyTheme`. Themes are applied once at startup based on the loaded config.

## UI contract

- Two panes: list (left) + detail (right). Active pane wears the accent border.
- Detail pane has three modes cycled by `t`: logs (default), full `systemctl status`, unit config.
- Follow mode (`f`) streams `journalctl -f` output. The buffer is bounded by `general.max_follow_lines` and reports trimming when it overflows.
- The terminal background shows through the panel borders by design. Only the modal, status bar, and help bar paint backgrounds (using `Surface` / `SurfaceDeep` from the theme).
- Mouse is enabled. Left-click on a pane focuses it; clicks/scroll inside the list and viewport are forwarded to the bubbles components.

## Themes

Defined in `internal/config/config.go` as `ThemeColors` records. Add a new theme by appending a record and extending `Config.ThemeColors()`. `Background` is intentionally empty for transparent panels — only the modal/bars use `SurfaceDeep`/`Surface`.

The default is `mocha` (Catppuccin Mocha). Older themes (`dark`, `light`, `high-contrast`) are kept for parity but the polish work has only been validated against Mocha.

## Common tasks

| Task | Where |
|---|---|
| Add a key | `keys` var in `internal/ui/model.go`; wire in the `tea.KeyMsg` switch and surface in the help bar bindings |
| Add a detail-view mode | `detailViewMode` enum in `model.go`; handle in `fetchDetailContent` and the logMsg path |
| Add a service action | Add method on `ServiceClient`, implement in `service.systemdClient`, add `tea.Cmd` builder in `commands.go`, wire into Update |
| Add a theme | Append `ThemeColors` in `config.go`, branch in `ThemeColors()` |
| Restyle a section | All styles flow from `ApplyTheme(theme)` — extend `internal/ui/styles.go` first, then reference the package-level style |

## Build, test, run

```bash
go build -o systemd-tui ./cmd/systemd-tui
go test ./...
go vet ./...
go run ./cmd/systemd-tui
```

## Known gaps / future work

- Golden-file tests for the View output (visual regressions are currently caught by eye)
- Bulk actions across selected services
- Status pane and config pane don't yet use the new styled detail header (only the logs view does)
- Light theme has only had a smoke check since the polish pass; needs a visual review
- `[` / `]` panel resize doesn't persist between runs

## Dependencies

```
github.com/charmbracelet/bubbles    v0.21.0
github.com/charmbracelet/bubbletea  v1.3.10
github.com/charmbracelet/lipgloss   v1.1.0
gopkg.in/yaml.v3                    v3.0.1
```

## Edge cases handled

| Case | Handling |
|---|---|
| `systemctl` not in PATH | Wrapped error surfaced in the status bar |
| No user services | Empty list with NoItems styling |
| Service name overflow | Suffix-preserving truncation (`truncateServiceName`) |
| Group header item selected | `getSelectedService()` returns nil; actions become no-ops |
| Filter active during refresh | Tick refresh skipped to avoid resetting the filter |
| Follow buffer overflow | Buffer trimmed to `max_follow_lines`, notice rendered |
| Editor exits non-zero | Edit failure surfaced; daemon reload skipped |
