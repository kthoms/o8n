# Story 4.3: Terminal Resize Handling

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want **the application to adapt cleanly when I resize my terminal window**,
so that **layout corruption doesn't interrupt my workflow**.

## Acceptance Criteria

1. **Given** the application is running and displaying any view
   **When** the operator resizes the terminal window (larger or smaller)
   **Then** Bubble Tea's `tea.WindowSizeMsg` is handled and the layout reflows correctly within one render cycle.
   **And** no text overflows, no borders break, and no content from a previous size persists as artifacts.

2. **Given** a terminal resize event
   **When** the new dimensions are calculated
   **Then** the table, header, and footer proportions are recalculated based on the new terminal dimensions.
   **And** `m.termWidth` and `m.termHeight` are updated in the model.

3. **Given** the terminal is resized to below the 120×20 minimum
   **When** the application renders
   **Then** it degrades gracefully per the responsive rules in Story 4.2 (column and hint visibility) — no panic, no corruption.

4. **Given** a resize event
   **When** a modal is currently open
   **Then** the modal is recalculated and re-centered based on the new dimensions without closing or losing state.

## Tasks / Subtasks

- [ ] **Implement WindowSizeMsg Handling (AC: 1, 2)**
  - [ ] Update `Update()` in `internal/app/update.go` to handle `tea.WindowSizeMsg`.
  - [ ] Ensure `m.termWidth` and `m.termHeight` are updated immediately.
  - [ ] Trigger a re-calculation of table dimensions (including `maxRows`) upon resize.
- [ ] **Modal Resize Resilience (AC: 4)**
  - [ ] Verify `renderModal()` in `internal/app/modal.go` uses dynamic dimensions from the model instead of hardcoded or stale values.
  - [ ] Ensure `OverlayCenter` and `OverlayLarge` modals re-center correctly after a resize.
- [ ] **Artifact Prevention (AC: 1)**
  - [ ] Audit `View()` in `internal/app/view.go` to ensure no stale strings or hardcoded line breaks persist between renders.
  - [ ] Use `lipgloss.Place` or similar tools to ensure clean clearing of the terminal buffer if necessary (usually handled by Bubble Tea).
- [ ] **Testing & Validation (AC: all)**
  - [ ] Create `internal/app/resize_test.go` to simulate `tea.WindowSizeMsg` and assert model dimension updates.
  - [ ] Manually verify resize behavior in VSCode and iTerm2.

## Dev Notes

### Previous Story Intelligence
- **Story 4.1 & 4.2 Learnings:** Layout decisions MUST use `m.termWidth` and `m.termHeight`. Hardcoded values are anti-patterns.
- **Chrome Allocation:** The 1-row header and 1-row footer are fixed; the table area (including FilterBar) must expand/contract to fill the remaining height.

### Architecture Compliance
- **WindowSizeMsg:** This is the standard Bubble Tea way to handle resizes. It must be processed in the main `Update` loop.
- **Dynamic Layout:** The `maxRows` calculation for the table renderer MUST be updated whenever `m.termHeight` changes.

### Anti-Patterns to Avoid (LLM Guardrails)
- **❌ DO NOT** ignore the `tea.WindowSizeMsg`.
- **❌ DO NOT** wait for the next data load to reflow the UI; reflow must happen immediately on resize.
- **❌ DO NOT** use `os.Stdout` to get terminal size; use the message provided by Bubble Tea.

### UI/UX Standards
- **Smooth Reflow:** The transition should be near-instant and visual artifacts (ghost text) must be avoided.

### Project Structure Notes
- **Update Logic:** `internal/app/update.go` (message handling).
- **View Logic:** `internal/app/view.go` (layout assembly).
- **Table Logic:** `internal/app/table.go` (row calculation).

### References
- [Source: `_bmad/planning-artifacts/epics.md#Story 4.3`]
- [Source: `_bmad/planning-artifacts/prd.md#FR37`]
- [Source: `_bmad/planning-artifacts/architecture.md#TUI Architecture`]
- [Source: `_bmad/planning-artifacts/ux-design-specification.md#Handling Terminal Resizes`]

## Dev Agent Record

### Agent Model Used

Gemini 2.0 Flash

### Debug Log References

### Completion Notes List

### File List
