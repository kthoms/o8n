# Story 3.2: Drill-Down Navigation & Breadcrumb

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want to drill down from a parent resource into related child resources and navigate back via Escape or breadcrumb,
so that I can traverse resource hierarchies (e.g., process definition → instances → variables) without losing my place.

## Acceptance Criteria

1. **Given** the current resource type has drilldown rules configured in `o6n-cfg.yaml`
   **When** the operator presses Enter on a row
   **Then** `prepareStateTransition(TransitionDrillDown)` is called, the current view is pushed onto the navigation stack, and the child resource type loads filtered to the selected parent.

2. **Given** the operator has drilled down one or more levels
   **When** the operator presses Escape
   **Then** `prepareStateTransition(TransitionPop)` is called and the previous view is fully restored (rows, cursor, columns, filters, breadcrumb).

3. **Given** the operator has drilled down multiple levels and the breadcrumb shows level labels
   **When** the operator selects a specific breadcrumb level (via mouse or future key binding)
   **Then** the application pops to that level and restores the view state captured at that level using `prepareStateTransition(TransitionPop)` or a series of pops.

## Tasks / Subtasks

- [x] Audit and Refine `executeDrilldown` in `internal/app/nav.go` (AC: 1)
  - [x] Verify `prepareStateTransition(TransitionDrillDown)` is called correctly.
  - [x] Ensure `m.navigationStack` correctly captures the `viewState` (rows, cursor, filters, etc.) before the transition.
  - [x] Verify `m.genericParams` is correctly populated with the drill-down filter.
  - [x] Ensure child columns are pre-set to avoid "column flash" from parent.
- [x] Audit and Refine `navigateToBreadcrumb` and Escape handling (AC: 2, 3)
  - [x] Verify `prepareStateTransition(TransitionPop)` is used when popping.
  - [x] Ensure `navigationStack` is correctly truncated and the restored state is applied to the model.
  - [x] Verify cursor position and search filters are correctly restored.
- [x] Add breadcrumb level selection logic (AC: 3)
  - [x] Ensure `navigateToBreadcrumb(idx)` correctly handles stack restoration for any previous level.
- [x] Tests and Validation (AC: 1, 2, 3)
  - [x] Create `internal/app/main_drilldown_nav_test.go`
  - [x] Test DrillDown: verify stack push and state clearing (except stack).
  - [x] Test Pop (Escape): verify stack pop and full state restoration.
  - [x] Test Jump: verify jumping to an early breadcrumb level restores the correct state.
- [x] Documentation (AC: 1, 2, 3)
  - [x] Ensure `specification.md` accurately describes the `TransitionDrillDown` and `TransitionPop` behavior.

## Dev Notes

- **State Transition Contract:** This story is the primary test for `TransitionDrillDown` and `TransitionPop`. Refer to `internal/app/nav.go` and `architecture.md` for the state clearing/restoration rules.
- **View State Persistence:** The `viewState` struct in `model.go` must be exhaustive enough to allow a perfect restoration of the parent view.
- **Navigation Stack:** The stack is the source of truth for the "Back" operation. Ensure it doesn't leak memory or grow boundlessly (though depth is usually shallow).
- **Wait for Data:** When drilling down, the UI should show a loading indicator while the child data is fetched.

### Project Structure Notes

- All logic resides in `internal/app/nav.go` (methods on `model`).
- Coordination with `update.go` for key handling (Enter and Escape).

### References

- Epic 3 Story 3.2: `_bmad/planning-artifacts/epics.md`
- State Transition Contract: `internal/app/nav.go`
- ViewState struct: `internal/app/model.go`

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.6 (implementation); Gemini 2.0 Flash (story creation)

### Debug Log References

N/A

### Completion Notes List

- **Root cause fixed: cursor clamped to -1 by SetRows([]Row{})**: Bubbles table v1.0.0 `SetRows(r)` clamps `cursor` to `len(r)-1`. When `r` is empty, cursor becomes -1. Subsequent `SetRows(rows)` only clamps downward, so cursor stays -1. Fixed by saving/restoring cursor around the clear-and-reload pattern in `applyDefinitions`, `applyInstances`, and `applyVariables`.
- **executeDrilldown: SetCursor(0) moved after final SetRows**: `SetCursor(0)` was called before `SetRows([]Row{})` in `executeDrilldown`, which would immediately be overridden to -1. Moved `SetCursor(0)` to after the final `SetRows` call.
- **Pre-existing test failures fixed**: 4 tests in `ui_test.go` were failing before this story (confirmed via investigation) — `TestConfigDrivenDrilldownFromDefinitionToInstances`, `TestConfigDrivenDrilldownFromInstancesToVariables`, `TestNavigationStackPreservesRowSelection`, `TestExtraEntersDontPushNavigationStack`. All now pass.
- **specification.md updated**: Added `genericParams` and `rowData` to the viewState field list at line 210.
- **navigateToBreadcrumb audited**: Logic is correct — idx=0 uses TransitionFull, idx>0 truncates stack and uses TransitionPop. No code changes needed.

### File List

- `internal/app/nav.go` — FIX: moved SetCursor(0) to after final SetRows in executeDrilldown
- `internal/app/table.go` — FIX: cursor save/restore in applyDefinitions, applyInstances, applyVariables
- `internal/app/main_drilldown_nav_test.go` — NEW: 12 navigation stack tests (AC 1, 2, 3)
- `specification.md` — DOC: added genericParams, rowData to viewState field list
