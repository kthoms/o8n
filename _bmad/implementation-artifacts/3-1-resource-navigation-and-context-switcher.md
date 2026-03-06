# Story 3.1: Resource Navigation & Context Switcher

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want to navigate to any of the 35 configured resource types using the context switcher (key `:`),
so that I can reach any operational view in seconds from anywhere in the application.

## Acceptance Criteria

1. **Given** 35 resource types are defined in `o6n-cfg.yaml`
   **When** the operator presses `:` to open the context switcher
   **Then** a searchable list of all configured resource types is displayed in an `OverlayCenter` modal.

2. **Given** the context switcher is open
   **When** the operator types a partial name and selects a match
   **Then** `prepareStateTransition(TransitionFull)` is called, clearing all prior view state
   **And** the selected resource type loads with a fresh, paginated table
   **And** the breadcrumb shows the new context name.

3. **Given** the selected resource type returns data from the API
   **When** the table loads
   **Then** the correct columns for that resource type (as defined in `o6n-cfg.yaml`) are displayed and the first row is selected.

## Tasks / Subtasks

- [x] Migrate `popupModeContext` legacy logic to `ModalContextSwitcher` factory modal (AC: 1, 2)
  - [x] Add `ModalContextSwitcher` to `ModalType` in `internal/app/model.go`
  - [x] Update `update.go` to set `m.activeModal = ModalContextSwitcher` on `:` instead of setting `popupModeContext`
  - [x] Remove `popupModeContext` from `popupMode` enum and clean up associated legacy logic in `model.go`, `update.go`, `view.go`, `nav.go`
- [x] Implement `ModalContextSwitcher` registration in `modal.go` (AC: 1)
  - [x] Register with `SizeHint: OverlayCenter`
  - [x] Register body renderer `renderContextSwitcherBody`
  - [x] Define `HintLine` showing `↑↓ Nav`, `Enter Select`, `Esc Close`
- [x] Create `renderContextSwitcherBody` in `internal/app/view.go` (AC: 1)
  - [x] Reuse `m.rootContexts` (filtered by `m.popup.input`) as items
  - [x] Use `lipgloss.Place` to render a centered list with current skin styles
- [x] Handle selection and navigation in `internal/app/update.go` (AC: 2, 3)
  - [x] In `ModalContextSwitcher` key handler, call `prepareStateTransition(TransitionFull)` on `Enter`
  - [x] Update `m.currentRoot`, `m.breadcrumb`, and `m.viewMode` to the selected resource
  - [x] Dispatch `m.fetchForRoot(selected)` cmd
  - [x] Ensure `m.table.SetCursor(0)` is called for the new context
- [x] Tests and Validation (AC: 1, 2, 3)
  - [x] Create `internal/app/main_context_switcher_test.go`
  - [x] Test `:` trigger, partial name search, selection, and transition state clearing
- [x] Documentation (AC: 3)
  - [x] Update `specification.md` modal keyboard behavior table with `ModalContextSwitcher`

## Dev Notes

- **Architecture Compliance:** Must use the `ModalConfig` and `renderModal()` factory established in Story 1.1. Do NOT add a `case ModalContextSwitcher` to the main `View` render switch.
- **State Transition Contract:** Mandatory use of `prepareStateTransition(TransitionFull)` from Story 1.2 to ensure zero state leakage between resource contexts.
- **Config-Driven:** Resource list is already filtered to only those with `TableDef` in `o6n-cfg.yaml` within `newModel`.
- **Legacy Cleanup:** The current `:` implementation in `update.go` (line ~708) and `view.go` (line ~630) is a legacy "popup" system. Migrating it to the modal factory is a key architectural goal of this story.

### Project Structure Notes

- All changes in `internal/app/`.
- No changes to `internal/operaton/` or `internal/client/`.
- Test file `main_context_switcher_test.go` co-located with implementation.

### References

- Epic 3 section: `_bmad/planning-artifacts/epics.md`
- Modal Factory: `internal/app/modal.go`
- State Transition Contract: `internal/app/nav.go` (check `prepareStateTransition`)
- Resource definitions: `o6n-cfg.yaml`

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6 (Claude Code)

### Debug Log References

N/A — audited committed implementation (commit 48f6040) and completed missing deliverables.

### Completion Notes List

- Core implementation was committed in 48f6040 by Gemini 2.0 Flash: ModalContextSwitcher registered in modal.go, renderContextSwitcherBody in view.go, key handling in update.go, popupModeContext removed, popupItems() updated for ModalContextSwitcher
- Added missing deliverables in this session:
  - Created `internal/app/main_context_switcher_test.go` with 18 tests covering `:` trigger, partial name search/filtering, arrow key navigation, selection via cursor, transition state clearing, and renderContextSwitcherBody rendering
  - Updated `specification.md`: added ModalContextSwitcher to modal types table and modal keyboard contract table
- Note: 3 pre-existing test failures in ui_test.go are not related to this story
- Note: `:` key when ModalContextSwitcher is active appends to filter input (not toggle-close); the toggle code in update.go line 829 is unreachable dead code

### File List

- `internal/app/model.go` — ModalContextSwitcher added to ModalType enum
- `internal/app/modal.go` — ModalContextSwitcher registered with OverlayCenter, renderContextSwitcherBody, HintLine
- `internal/app/view.go` — renderContextSwitcherBody() implemented
- `internal/app/update.go` — `:` key handler sets ModalContextSwitcher; Enter handler calls prepareStateTransition(TransitionFull) + fetch + SetCursor(0)
- `internal/app/nav.go` — popupItems() updated to handle ModalContextSwitcher; popupModeContext removed
- `internal/app/main_context_switcher_test.go` — NEW: 18 tests for ModalContextSwitcher
- `specification.md` — ModalContextSwitcher added to modal types table and keyboard contract table

## Change Log

- 2026-03-06: Story completed — tests and documentation added for implementation committed in 48f6040
