# systemd-tui Revised Improvements Plan

**Branch**: `feat/bugfixes-ui-improvements`
**Revised**: 2026-03-08
**Status**: Reworked after code review

---

## Why This Revision Exists

The original plan had good UX ideas, but it needed reordering and tighter alignment with the current codebase.

Main corrections in this revision:

- Put correctness and backend/UI contract fixes ahead of UI polish.
- Remove conflicts with the current `ServiceClient` interface and config schema.
- Treat partially implemented items as follow-up work, not greenfield work.
- Mark backend-blocked features as blocked instead of pretending they are ready.
- Move test coverage earlier so refactors and UI work are safer.

---

## Planning Principles

1. Fix correctness before polish.
2. Add tests before large UI refactors.
3. Respect the existing theme and config system.
4. Do not change public interfaces when a new method is more appropriate.
5. Avoid speculative performance work unless profiling shows a real problem.
6. Every UI change must still work at 80x24 and after terminal resize.

---

## Priority Matrix

| Priority | Items | Focus |
|----------|-------|-------|
| **P0 - Foundation / Correctness** | F1, F2, F3, F4, 12 | Fix broken behavior, data correctness, and test drift |
| **P1 - High Value UX** | 14, 18, 1, 7, 5+13 | Safe refactor and most visible usability wins |
| **P2 - Product Improvements** | 3, 6, 8, 9, 10, 16, 19 | Useful enhancements with manageable risk |
| **P3 - Deferred / Nice to Have** | 4, 11, 15, 17, 20 | Blocked, optional, or low ROI work |

---

## Foundation Gaps Missing From The Original Plan

### F1: Fix Follow Mode Streaming
**Priority**: P0 | **Effort**: Medium | **Files**: `internal/ui/model.go`

**Problem**:
Follow mode likely stops after the first emitted line because `startFollow()` returns a single `logLineMsg` instead of continuously scheduling the next read.

**Why this matters**:
This is a core feature bug, not a polish issue.

**Scope**:
- Make follow mode continuously consume log lines until cancelled.
- Stop safely on service change, quit, and explicit toggle off.
- Preserve bottom-follow behavior while still allowing manual scrolling if desired.

**Acceptance Criteria**:
- Follow mode keeps updating for an active service without manual refresh.
- Toggling `f` off stops updates immediately.
- Changing selected service while following does not leak goroutines or stale updates.

---

### F2: Fix Initial Selection and Detail Pane Loading
**Priority**: P0 | **Effort**: Low | **Files**: `internal/ui/model.go`

**Problem**:
After the initial service load, the detail pane only fetches logs if a selected service is already known. Selection is not explicitly initialized from the new list state.

**Why this matters**:
The app should show useful detail content immediately after loading.

**Scope**:
- After setting list items, initialize selection if possible.
- Fetch logs/config for the selected service after initial load and after list rebuilds when selection changes.
- Keep behavior safe when the selected item is a group header.

**Acceptance Criteria**:
- On initial load, the first real service shows details automatically.
- Rebuilding the list after polling preserves a valid selection when possible.
- No detail fetch is attempted for group headers.

---

### F3: Correct `Enabled` State Derivation
**Priority**: P0 | **Effort**: Low | **Files**: `internal/service/client.go`

**Problem**:
`Enabled` is currently inferred from runtime state, which is not the same as unit-file enablement.

**Why this matters**:
This produces incorrect service metadata and will mislead future UI features.

**Scope**:
- Derive `Enabled` from `list-unit-files` state.
- Treat `enabled`, `enabled-runtime`, `static`, `indirect`, `disabled`, `masked`, and `generated` intentionally.
- Document the meaning of `Enabled` in the UI if needed.

**Acceptance Criteria**:
- `Enabled` reflects unit-file state, not active/running state.
- Tests cover enabled, disabled, static, and masked examples.

---

### F4: Refresh Stale Tests and Mock Interfaces
**Priority**: P0 | **Effort**: Medium | **Files**: `internal/service/client_test.go`, `internal/ui/model_test.go`

**Problem**:
Some service-layer mocks and tests no longer match the current `ServiceClient` interface shape.

**Why this matters**:
The current test suite gives less protection than it appears to.

**Scope**:
- Update mocks to use `context.Context` everywhere.
- Add focused tests for selection initialization, follow-mode lifecycle, and enabled-state mapping.
- Remove or rewrite stale interface-only tests that no longer verify real behavior.

**Acceptance Criteria**:
- Test doubles implement the current interface correctly.
- New foundation issues have regression tests.

---

## Revised Existing Issues

### Issue 12: Log Buffer Memory Limit
**Priority**: P0 | **Effort**: Low | **Files**: `internal/ui/model.go`, `internal/config/config.go`

**Revision Note**:
This is partially implemented already because follow-mode lines are capped to 1000 entries. The remaining work is to make the behavior robust and configurable.

**Scope**:
- Keep line cap.
- Add optional byte cap to avoid very large single-line growth.
- Show a trim notice once content has been truncated.
- Preserve viewport behavior when old lines are dropped.

**Acceptance Criteria**:
- Follow mode memory use stays bounded over long sessions.
- Trim behavior is visible and non-jarring.

---

### Issue 14: Golden File Tests For View Output
**Priority**: P1 | **Effort**: Medium | **Files**: `internal/ui/model_test.go`, `internal/ui/testdata/`

**Revision Note**:
Moved earlier. This should happen before or alongside UI refactors.

**Scope**:
- Add golden tests for stable `View()` output.
- Cover empty state, grouped list, filter active, confirmation mode, create modal, and detail pane states.
- Stabilize test width/height and theme so snapshots are deterministic.

**Acceptance Criteria**:
- UI refactors can be validated against snapshot output.
- Golden tests are easy to refresh intentionally.

---

### Issue 18: Extract Rendering And Commands From `model.go`
**Priority**: P1 | **Effort**: Medium | **Files**: new `internal/ui/view.go`, new `internal/ui/commands.go`, `internal/ui/model.go`

**Revision Note**:
Still valuable, but no longer first. It follows the foundation fixes and test safety net.

**Scope**:
- Keep `MainModel` state and `Update()` orchestration in `model.go`.
- Move rendering helpers to `view.go`.
- Move async command builders to `commands.go`.
- Do not change behavior during this refactor.

**Acceptance Criteria**:
- No functional change.
- `model.go` is materially smaller and easier to reason about.
- Golden tests stay green.

---

### Issue 1: Long Service Name Truncation
**Priority**: P1 | **Effort**: Medium | **Files**: `internal/ui/delegate.go`

**Revision Note**:
Use display width, not raw byte length.

**Scope**:
- Truncate service names based on rendered width.
- Preserve `.service` suffix when possible.
- Avoid breaking selected-row markers and source indicators.
- Keep full name visible in the detail pane.

**Acceptance Criteria**:
- Very long service names fit cleanly in narrow list panes.
- Truncation remains stable under resize.
- No malformed Unicode rendering.

---

### Issue 7: Consolidate Help System
**Priority**: P1 | **Effort**: Medium | **Files**: `internal/ui/model.go`, new `internal/ui/help.go`

**Revision Note**:
The current code already has a custom help bar and a partially unused Bubble help model. This work should consolidate them instead of layering another help system on top.

**Scope**:
- Decide on one help architecture:
  - context-aware inline help bar as primary, or
  - inline short help plus popup full help.
- Ensure `?` renders something real.
- Hide irrelevant keys in modal/confirm/filter contexts.

**Acceptance Criteria**:
- Help content matches the current interaction context.
- `?` reliably shows expanded help.
- No dead help state remains in the code.

---

### Issue 5+13: Theme-Aware Log Styling
**Priority**: P1 | **Effort**: Medium | **Files**: new `internal/ui/log_styling.go`, `internal/config/config.go`, `internal/ui/model.go`

**Revision Note**:
Keep this combined, but make it theme-driven.

**Scope**:
- Add theme tokens for log levels and optional line emphasis.
- Style obvious level keywords only.
- Strip or neutralize existing ANSI sequences before styling if needed.
- Keep performance acceptable for long logs.

**Acceptance Criteria**:
- Error/warn/info/debug lines are visually distinguishable.
- Styling works in dark, light, and high-contrast themes.
- Styling does not noticeably degrade viewport responsiveness.

---

### Issue 3: Header Service Count Badge
**Priority**: P2 | **Effort**: Low | **Files**: `internal/ui/model.go`

**Revision Note**:
Update the list title through the list component instead of treating it as a standalone panel string.

**Scope**:
- Show total count and filtered count in the list title.
- Use compact formatting for narrow widths.

**Acceptance Criteria**:
- Counts are accurate after filtering and polling.
- Title stays readable in narrow layouts.

---

### Issue 6: Status Bar Redesign
**Priority**: P2 | **Effort**: Low | **Files**: `internal/ui/model.go`

**Revision Note**:
Treat this as a small redesign, not just string concatenation.

**Scope**:
- Show service counts, filter mode, group mode, and follow state.
- Prioritize critical info from left to right.
- Degrade gracefully in narrow terminals.

**Acceptance Criteria**:
- Important status remains visible at 80x24.
- The bar is readable without wrapping awkwardly.

---

### Issue 8: Theme-Aware Colored Source Indicators
**Priority**: P2 | **Effort**: Low | **Files**: `internal/config/config.go`, `internal/ui/styles.go`, `internal/ui/delegate.go`

**Revision Note**:
Do not hardcode colors directly in the delegate. Add theme tokens first.

**Scope**:
- Add theme colors for source types.
- Keep shape-based indicators for accessibility.
- Render color only when it improves readability in the active theme.

**Acceptance Criteria**:
- User, system, transient, generated, static, and unknown are distinguishable.
- Light and high-contrast themes remain readable.

---

### Issue 9: Scroll Position Indicator
**Priority**: P2 | **Effort**: Low | **Files**: `internal/ui/model.go`

**Scope**:
- Show progress in the detail pane header.
- Hide it when the content fits entirely on one screen.

**Acceptance Criteria**:
- Users can tell where they are in long logs/status output.

---

### Issue 10: Resizable Split Panels
**Priority**: P2 | **Effort**: Medium | **Files**: `internal/ui/model.go`, `internal/config/config.go`

**Revision Note**:
Keep the feature, but use terminal-friendly keys and enforce minimum sizes.

**Scope**:
- Add configurable split ratios.
- Preserve ratio on resize.
- Use safer keybindings than raw `<` and `>` if terminal handling is unreliable.

**Acceptance Criteria**:
- Panels can be resized without breaking layout.
- Minimum usable widths are enforced.

---

### Issue 16: Status View Toggle
**Priority**: P2 | **Effort**: Medium | **Files**: `internal/client/types.go`, `internal/service/client.go`, `internal/ui/model.go`

**Revision Note**:
Do not overload the existing `GetStatus(ctx, name)` method. Add a new method for detailed textual status output.

**Proposed API**:
```go
GetStatusDetails(ctx context.Context, name string) (string, error)
```

**Scope**:
- Add a detail-pane toggle between logs and full unit status.
- Keep the existing lightweight `GetStatus` API for active-state lookups.
- Define clear refresh behavior in both modes.

**Acceptance Criteria**:
- Toggling views does not break polling or selection.
- The service client interface remains backwards-compatible.

---

### Issue 19: Structured Error Types
**Priority**: P2 | **Effort**: Medium | **Files**: new `internal/client/errors.go`, `internal/service/client.go`

**Revision Note**:
Useful, but only after command execution captures enough detail.

**Scope**:
- Capture stderr and exit code for service operations.
- Wrap errors with operation, unit, and parsed class when possible.
- Allow the UI to render more actionable messages.

**Acceptance Criteria**:
- Common systemctl failures are distinguishable in tests.
- UI can show better messages than generic `failed`.

---

## Deferred / Blocked Items

### Issue 4: Relative Timestamps In Logs
**Priority**: P3 | **Effort**: Medium | **Status**: Deferred

**Reason**:
This depends on using a stable, parseable log format. Parsing human-formatted journal text is brittle.

**Unblock Condition**:
- Switch log retrieval to a structured or machine-stable format first.

---

### Issue 11: Group Header Selection Stats
**Priority**: P3 | **Effort**: Medium | **Status**: Blocked

**Reason**:
The current group header item only stores a title, and service resource fields like `PID`, `Memory`, `CPU`, and `Since` are not populated yet.

**Unblock Condition**:
- Add real group metadata and populate service stats in the backend first.

---

### Issue 15: Clipboard Support
**Priority**: P3 | **Effort**: Low | **Status**: Optional

**Reason**:
Useful, but environment-sensitive and not core to service management.

**Scope**:
- Implement only with clear fallback/error handling for SSH, Wayland, X11, and headless sessions.

---

### Issue 17: Cache Service List Between Polls
**Priority**: P3 | **Effort**: Medium | **Status**: Measure First

**Reason**:
This is speculative optimization. Current service counts are small enough that correctness and UX matter more than in-place diffing.

**Unblock Condition**:
- Add only if profiling shows list rebuilds cause visible flicker or CPU churn.

---

### Issue 20: Refresh Interval Configuration Cleanup
**Priority**: P3 | **Effort**: Low | **Status**: Rewrite

**Revision Note**:
There is already config support via `general.refresh_interval`. Do not add a duplicate top-level `poll_interval` key.

**Scope**:
- Validate and document the existing `general.refresh_interval` setting.
- Optionally cap absurd values and surface warnings.

**Acceptance Criteria**:
- Existing config remains the single source of truth.
- Invalid values fall back safely.

---

## Recommended Implementation Order

### Phase 0: Correctness First
1. F1 - Fix follow mode streaming
2. F2 - Fix initial selection and detail loading
3. F3 - Correct enabled-state derivation
4. F4 - Refresh stale tests and mocks
5. Issue 12 - Finish bounded log buffer behavior

### Phase 1: Safety Net And Refactor
6. Issue 14 - Golden tests for `View()`
7. Issue 18 - Extract `view.go` and `commands.go`

### Phase 2: Highest-Value UX Wins
8. Issue 1 - Long name truncation
9. Issue 7 - Consolidated help system
10. Issue 5+13 - Theme-aware log styling

### Phase 3: Medium-Risk Product Improvements
11. Issue 3 - Header service count badge
12. Issue 6 - Status bar redesign
13. Issue 8 - Theme-aware source colors
14. Issue 9 - Scroll indicator
15. Issue 10 - Resizable panels
16. Issue 16 - Status view toggle
17. Issue 19 - Structured errors

### Phase 4: Deferred Work
18. Issue 20 - Refresh interval cleanup
19. Issue 15 - Clipboard support
20. Issue 4 - Relative timestamps
21. Issue 11 - Group header stats
22. Issue 17 - Poll caching only if profiling justifies it

---

## Cross-Cutting Acceptance Criteria

- `go test ./...` passes after every phase.
- UI remains usable at 80x24.
- Terminal resize does not break layout or corrupt state.
- No goroutine leaks from follow mode.
- Theme changes still work in dark, light, and high-contrast modes.
- Selection remains valid after refreshes, grouping changes, and filtering changes.

---

## Out Of Scope For This Plan

- New service-manager backends beyond systemd user services.
- Major redesign of the overall layout.
- Advanced metrics aggregation until service stats are actually populated.

---

## Summary Of What Changed

- Added four missing foundation items: follow streaming, initial detail loading, enabled-state correctness, and stale test cleanup.
- Kept Issue 12, but reclassified it as partially complete follow-up work.
- Moved golden tests earlier.
- Moved refactor work behind correctness and tests.
- Rewrote Issue 16 to add a new API instead of breaking `GetStatus`.
- Rewrote Issue 20 to use the existing config schema instead of adding a duplicate key.
- Deferred backend-blocked items instead of pretending they are implementation-ready.
