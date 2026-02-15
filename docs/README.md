# Design Documentation Index

> **Systemd TUI Manager v2.0** - Comprehensive Design Documentation

---

## Documentation Structure

| Document | Description |
|----------|-------------|
| [DESIGN.md](./DESIGN.md) | Main design document with executive summary, vision, UI architecture |
| [KEYBINDINGS.md](./KEYBINDINGS.md) | Complete keybinding system and user workflows |
| [STATE_MACHINE.md](./STATE_MACHINE.md) | Application states, interactions, edge cases, performance |
| [VISUAL_DESIGN.md](./VISUAL_DESIGN.md) | Typography, colors, animations, accessibility |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Technical architecture, components, data flow |

---

## Quick Reference

### Design Principles

1. **Keyboard-First** - Every action via keyboard
2. **Immediate Feedback** - Response within 100ms
3. **Progressive Disclosure** - Simple default, powerful on demand
4. **Predictable Behavior** - Consistent keybindings
5. **Graceful Degradation** - Work everywhere

### Key Insights from Research

**Lazygit Patterns:**
- Split-panel layout with contextual focus
- Tab to switch panels
- Status bar with action hints
- Confirmation for destructive actions

**Htop/Btop Patterns:**
- Real-time updates with color indicators
- Header with system summary
- Footer with key hints
- Scrollable list with metadata columns

**K9s Patterns:**
- Command mode with `:`
- Contextual actions per resource type
- Tab-based navigation
- Live resource updates

**Common Keybinding Conventions:**
- `j/k` - Navigation (vim-style)
- `q` - Quit
- `?` - Help
- `/` - Search/Filter
- `:` - Command mode
- `Enter` - Select/Confirm
- `Esc` - Cancel/Return

---

## Implementation Priority

### P0 - Core Features
- [ ] Multi-panel layout with tabs
- [ ] Mode-based navigation (Normal/Filter/Command)
- [ ] Live log tailing
- [ ] ServiceClient interface abstraction

### P1 - Important Features
- [ ] Configuration system
- [ ] Theme support (dark/light/high-contrast)
- [ ] Service grouping
- [ ] Multi-backend support

### P2 - Nice to Have
- [ ] Command palette
- [ ] Service templates
- [ ] Bulk operations
- [ ] Plugin system

---

## Current State Analysis

### What's Working Well

| Aspect | Status |
|--------|--------|
| Basic UI Layout | ✓ Split view functional |
| Service Actions | ✓ Start/Stop/Restart/Enable/Disable |
| Status Colors | ✓ Green/Red/Gray indicators |
| Auto-refresh | ✓ 2-second polling |
| Confirmation Flow | ✓ Safe destructive actions |
| Editor Integration | ✓ Works with $EDITOR |

### What Needs Improvement

| Gap | Impact | Solution |
|-----|--------|----------|
| No interface abstraction | Testing, extensibility | ServiceClient interface |
| Limited error context | User confusion | Structured error display |
| No configuration | Hardcoded behavior | YAML config system |
| Single backend | Limited use | Multi-backend support |
| Static logs | No debugging | Live log tailing |
| No grouping | Cluttered list | Service groups/filters |

---

## Next Steps

1. **Review** this documentation for accuracy and completeness
2. **Prioritize** features based on user needs
3. **Prototype** the new UI layout
4. **Implement** the ServiceClient interface
5. **Test** with real-world workflows

---

## Questions to Resolve

1. Should we support system services (with sudo)?
2. What's the preferred config format? (YAML vs TOML)
3. Should there be a web/REST API for scripting?
4. How to handle services with long names?
5. Should groups be user-defined or auto-detected?
