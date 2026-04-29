# Documentation

This directory holds design and reference docs for the systemd TUI manager.

## Index

| Document | Audience | Status |
|---|---|---|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Developers | Reference — package layout, data flow |
| [KEYBINDINGS.md](./KEYBINDINGS.md) | Users / developers | Reference — full keybinding catalogue |
| [STATE_MACHINE.md](./STATE_MACHINE.md) | Developers | Reference — modes, transitions, edge cases |
| [DESIGN.md](./DESIGN.md) | Historical | Original design doc; predates current UI polish pass |
| [VISUAL_DESIGN.md](./VISUAL_DESIGN.md) | Historical | Original visual specs; superseded by implementation (see below) |

For everyday usage, see the top-level [README.md](../README.md). For agent / contributor notes, see [AGENTS.md](../AGENTS.md).

---

## Current implementation reference

This section captures decisions made during the UI polish pass that diverge from or supersede the older design notes.

### Default theme — Catppuccin Mocha

`internal/config/config.go` defines four themes: `mocha` (default), `dark` (legacy), `light`, and `high-contrast`. The Mocha palette:

| Token | Hex | Role |
|---|---|---|
| Surface | `#313244` | Status bar, kbd chip background, list-row hover |
| SurfaceDeep | `#181825` | Modal & help bar background |
| Border | `#45475a` | Inactive panel borders, separators |
| BorderActive | `#cba6f7` | Active panel border, focused inputs, list title |
| Accent | `#89b4fa` | Section banners, links, secondary highlight |
| Text | `#cdd6f4` | Primary text |
| TextMuted | `#a6adc8` | Secondary text |
| TextDim | `#6c7086` | Tertiary / placeholder text |
| StatusActive | `#a6e3a1` | Running services |
| StatusFailed | `#f38ba8` | Failed services, error logs |
| StatusActivating | `#fab387` | Activating / deactivating / reloading |

### Transparent panels

The list and detail panels do not paint a background — the user's terminal background shows through. Only the modal, status bar, help bar, and filter bar paint backgrounds (using `SurfaceDeep` or `Surface`). This is intentional: it keeps the tool feeling terminal-native and avoids clashing with image / wallpaper backgrounds.

### Detail header

Replaces the old `Service: x / Status: y / Description: z` plain text. The new header shows:

- Service name (bold, full Text color)
- Right-aligned status pill (`● active`) and source pill (`◆ user`)
- Description (TextMuted)
- A section banner separating the header from the body (`──── Last 50 lines ────────`)

In follow mode the section banner becomes `⦿ FOLLOWING  press f to stop  ─────` in the active-status colour.

### Modal

Single `SurfaceDeep` background; per-line padding ensures no trailing dark band. Each text input has a thin underline border that highlights when focused. Choice fields (Type, Restart) use a kbd chip (`[t]`) and a "cycle" hint instead of inline `(press t)`.

### Status bar

Segments render as labelled pills (`filter all`, `group status`, `view config`, `● following`). The right side shows the most recent action message.

### Help bar

Two rows of `[key] description` chips, separated by `·`. Toggling `?` expands the bubbles `help.Model` full view.

### Mouse

Enabled via `tea.WithMouseCellMotion()`. Left-click on a pane sets it as the active view; clicks and the scroll wheel are forwarded to the underlying list / viewport components.

### Source indicators

Listed in the row delegate:

| Glyph | Source | Meaning |
|---|---|---|
| ● | user | Created in `~/.config/systemd/user/` |
| ○ | system | Provided by `/usr/lib/systemd/user/` |
| ◌ | transient | Runtime-created |
| ◆ | generated | Auto-generated (e.g. desktop autostart) |
| ◇ | static | Static / alias unit |
| · | unknown | Origin could not be determined |

---

## Updating these docs

When you change UI behaviour, prefer updating the **Current implementation reference** section above and the user-facing README. The historical design docs (`DESIGN.md`, `VISUAL_DESIGN.md`) can be left alone — they document the original intent and are useful when revisiting the rationale, but they are not authoritative for the current code.
