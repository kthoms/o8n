# Story 2.2: Per-View Hint Functions

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want each view to declare its own available actions as hints in the footer,
So that the footer always reflects the actions relevant to my current context.

## Acceptance Criteria

1. **Given** the operator is in the main table view (no modal, no popup, no search mode)
   **When** the footer renders
   **Then** primary actions for that view (e.g., `Enter` to drill down, `Ctrl+Space` for actions, `/` to filter, `Ctrl+r` for refresh, `?` for help) appear as key hints in the footer

2. **Given** the operator is in the context switcher popup (`popupModeContext` — `:` key)
   **When** the footer renders
   **Then** the hints reflect only context switcher actions: `↑↓ select`, `Tab/Enter switch`, `Esc cancel`
   **And** the table's primary hints (Enter/drill, Ctrl+Space/actions, etc.) are NOT shown

3. **Given** the operator is in the search popup (`popupModeSearch` — `/` key)
   **When** the footer renders
   **Then** the hints reflect only search actions: `↑↓ select`, `Enter jump`, `Esc cancel`

4. **Given** the operator has the skin picker open (`popupModeSkin` — `Ctrl+T` key)
   **When** the footer renders
   **Then** the hints reflect only skin picker actions: `↑↓ preview`, `Enter apply`, `Esc revert`

5. **Given** any modal is active (ModalHelp, ModalEdit, ModalConfirmDelete, ModalSort, ModalEnvironment, etc.)
   **When** the footer renders
   **Then** the hints reflect only that modal's available actions (from the modal's HintLine configuration)
   **And** the table's primary hints are NOT shown

6. **Given** the terminal width is reduced below a hint's `MinWidth` threshold
   **When** the footer renders in any sub-view
   **Then** that hint is omitted without breaking the layout of remaining hints (existing behaviour preserved)

## Tasks / Subtasks

- [x] Task 1: Add HintLine to all OverlayCenter modals lacking one (AC: 5)
  - [x] In `modal.go`, add `HintLine` to `ModalConfirmDelete`:
    - `{Key: "Enter", Label: "confirm", Priority: 1}`, `{Key: "Esc", Label: "cancel", Priority: 2}`
  - [x] In `modal.go`, add `HintLine` to `ModalConfirmQuit`:
    - `{Key: "Enter", Label: "quit", Priority: 1}`, `{Key: "Esc", Label: "cancel", Priority: 2}`
  - [x] In `modal.go`, add `HintLine` to `ModalSort`:
    - `{Key: "↑↓", Label: "select", Priority: 1}`, `{Key: "Enter", Label: "apply", Priority: 1}`, `{Key: "Esc", Label: "cancel", Priority: 2}`
  - [x] In `modal.go`, add `HintLine` to `ModalEnvironment`:
    - `{Key: "↑↓", Label: "select", Priority: 1}`, `{Key: "Enter", Label: "switch", Priority: 1}`, `{Key: "Esc", Label: "cancel", Priority: 2}`
  - [x] In `modal.go`, add `HintLine` to `ModalEdit`:
    - `{Key: "Tab", Label: "switch", Priority: 1}`, `{Key: "Enter", Label: "save", Priority: 1}`, `{Key: "Esc", Label: "cancel", Priority: 2}`
  - [x] In `modal.go`, add `HintLine` to `ModalFirstRun` (intentionally NO Esc, AC2 of Story 1.5):
    - `{Key: "↑↓", Label: "nav", Priority: 1}`, `{Key: "Enter", Label: "select", Priority: 1}`
  - [x] Run `go build ./...` — must compile

- [x] Task 2: Refactor `currentViewHints` to dispatch by active state (AC: 1–5)
  - [x] In `internal/app/hints.go`, rename the body of current `currentViewHints` to a new function `tableViewHints(m model) []Hint` (no signature change — all the existing hint declarations move there)
  - [x] Replace `currentViewHints(m model) []Hint` body with a dispatch function:
    ```go
    func currentViewHints(m model) []Hint {
        // Active modal → use the modal's HintLine (empty if not configured)
        if m.activeModal != ModalNone {
            if cfg, ok := modalRegistry[m.activeModal]; ok && len(cfg.HintLine) > 0 {
                return cfg.HintLine
            }
            return nil
        }
        // Active popup mode → per-popup hints
        switch m.popup.mode {
        case popupModeContext:
            return contextSwitcherHints()
        case popupModeSearch:
            return searchPopupHints()
        case popupModeSkin:
            return skinPickerHints()
        }
        // Default: main table view hints
        return tableViewHints(m)
    }
    ```
  - [x] Add `contextSwitcherHints() []Hint` to `hints.go`
  - [x] Add `searchPopupHints() []Hint` to `hints.go`
  - [x] Add `skinPickerHints() []Hint` to `hints.go`
  - [x] Run `go build ./...` and `go vet ./...` — must be clean

- [x] Task 3: Update existing tests for renamed function (AC: 1)
  - [x] In `internal/app/main_hints_test.go`: all existing `TestCurrentViewHints_*` tests call `currentViewHints(m)` on a model with no active modal/popup — verify they still pass without modification (they should, since `currentViewHints` dispatches to `tableViewHints` in the default case and the test model has `m.activeModal == ModalNone` and `m.popup.mode == popupModeNone`)
  - [x] Run `go test ./internal/app/... -run "TestCurrentViewHints|TestFilterHints" -v` — all pass

- [x] Task 4: Add per-view hint dispatch tests (AC: 2–5)
  - [x] File: `internal/app/main_hints_test.go` (extend existing file)
  - [x] `TestCurrentViewHints_ContextSwitcherPopup` — set `m.popup.mode = popupModeContext` → includes `Tab/Enter switch`, does NOT include `?/help`
  - [x] `TestCurrentViewHints_SearchPopup` — set `m.popup.mode = popupModeSearch` → includes `Enter jump`, does NOT include `?/help`
  - [x] `TestCurrentViewHints_SkinPickerPopup` — set `m.popup.mode = popupModeSkin` → includes `Enter apply`, does NOT include `?/help`
  - [x] `TestCurrentViewHints_ModalHelpActive` — set `m.activeModal = ModalHelp` → includes `↑↓/scroll` and `q/Esc/close`, does NOT include `?/help`
  - [x] `TestCurrentViewHints_ModalEditActive` — set `m.activeModal = ModalEdit` → includes `Tab/switch` and `Enter/save`, does NOT include `?/help`
  - [x] `TestCurrentViewHints_ModalFirstRunNoEsc` — set `m.activeModal = ModalFirstRun` → includes `↑↓/nav` and `Enter/select`, does NOT include `Esc/cancel`
  - [x] `TestCurrentViewHints_ModalConfirmDeleteActive` — set `m.activeModal = ModalConfirmDelete` → includes `Enter/confirm` and `Esc/cancel`
  - [x] `TestCurrentViewHints_MainTableUnchanged` — default model → same hints as Story 2.1 (help/switch/nav still present)
  - [x] Run `go test ./internal/app/... -run "TestCurrentViewHints" -v` — all pass

- [x] Task 5: Verify full suite and coverage (AC: all)
  - [x] `make test` — all packages pass, zero regressions
  - [x] `go vet ./...` — clean
  - [x] `gofmt -w .` — no diff
  - [x] `make cover` — `hints.go` coverage 100% (all functions, all branches)
  - [x] Update `specification.md` to note per-view hint dispatch and that footer always reflects active context

## Dev Notes

### Current State After Story 2.1

Story 2.1 created the `Hint` struct, `filterHints`, and `currentViewHints` in `hints.go`. The footer calls:
```go
hints := filterHints(currentViewHints(*m), width)  // view.go:63
```

`currentViewHints` currently returns the same table-view hints regardless of what popup or modal is active:
```go
func currentViewHints(m model) []Hint {
    hints := []Hint{
        {Key: "?", Label: "help", ...},    // always returned even when ModalHelp is open!
        {Key: ":", Label: "switch", ...},  // always returned even when context switcher IS open!
        ...
    }
    // conditional model-state checks...
    return hints
}
```

Story 2.2 completes the architecture intent: `currentViewHints` dispatches to per-view hint functions.

### Dispatch Priority (MUST be this order)

```
1. Active modal (ModalFirstRun, ModalHelp, ModalEdit, etc.) → modal's HintLine
2. Active popup mode (popupModeContext, popupModeSearch, popupModeSkin) → per-popup hints
3. Default → tableViewHints(m) [renamed from current currentViewHints body]
```

**Why modal check first:** A popup could theoretically be open while a modal opens on top of it. The modal takes precedence for the footer.

### Modal HintLine Additions (Task 1)

Currently, these OverlayCenter modals have NO HintLine in their `ModalConfig`:
- `ModalConfirmDelete` — needs Enter/Esc
- `ModalConfirmQuit` — needs Enter(quit)/Esc
- `ModalSort` — needs ↑↓/Enter/Esc
- `ModalEnvironment` — needs ↑↓/Enter/Esc
- `ModalEdit` — needs Tab/Enter/Esc
- `ModalFirstRun` — needs ↑↓/Enter (NO Esc — documented exception from Story 1.5)

These already have `OverlayLarge` HintLine and need no changes:
- `ModalHelp`: `↑↓ scroll, q/Esc close` ✓
- `ModalDetailView`: `↑↓ scroll, q/Esc close` ✓
- `ModalTaskComplete`: `Tab switch, Enter confirm, Esc cancel` ✓

### `tableViewHints(m model) []Hint` — Renamed, Not Changed

Move the **entire body** of current `currentViewHints` to `tableViewHints`. The function signature changes from `currentViewHints(m model) []Hint` to `tableViewHints(m model) []Hint`. Nothing else changes in the implementation — all existing hint declarations, MinWidth values, and conditional logic remain identical.

```go
func tableViewHints(m model) []Hint {
    // (entire current body of currentViewHints goes here, unchanged)
    hints := []Hint{
        {Key: "?", Label: "help", MinWidth: 0, Priority: 1},
        ...
    }
    // conditional hints...
    return hints
}
```

### Per-Popup Hint Functions

These are **width-unaware** (MinWidth: 0 for all, since popup hints are always short and fit in any reasonable terminal). `filterHints` is still called on the result by the footer renderer, so MinWidth gating is still respected if needed in future.

The hints mirror the existing inline hint strings already used inside the popup box in `view.go`:
- `popupModeContext` inline: `"↑↓:select  Tab/Enter:switch  Esc:cancel"` → `contextSwitcherHints()`
- `popupModeSearch` inline: `"↑↓:select  Enter:jump  Esc:cancel"` → `searchPopupHints()`
- `popupModeSkin` via `m.popup.hint`: `"↑↓:preview  Enter:apply  Esc:revert"` → `skinPickerHints()`

After this story, the footer will be consistent with the inline popup hints.

### Test Patterns

All tests extend `main_hints_test.go`:
```go
func TestCurrentViewHints_ContextSwitcherPopup(t *testing.T) {
    m := newTestModel(t)
    m.popup.mode = popupModeContext

    hints := currentViewHints(m)

    if _, ok := findHint(hints, "Tab/Enter", "switch"); !ok {
        t.Fatal("expected Tab/Enter switch hint for context switcher")
    }
    // Table hints must NOT appear when context switcher is open
    if _, ok := findHint(hints, "?", "help"); ok {
        t.Fatal("did not expect table help hint while context switcher is active")
    }
}

func TestCurrentViewHints_ModalHelpActive(t *testing.T) {
    m := newTestModel(t)
    m.activeModal = ModalHelp

    hints := currentViewHints(m)

    if _, ok := findHint(hints, "↑↓", "scroll"); !ok {
        t.Fatal("expected scroll hint for ModalHelp")
    }
    if _, ok := findHint(hints, "?", "help"); ok {
        t.Fatal("did not expect table help hint while ModalHelp is active")
    }
}

func TestCurrentViewHints_ModalFirstRunNoEsc(t *testing.T) {
    m := newTestModel(t)
    m.activeModal = ModalFirstRun

    hints := currentViewHints(m)

    if _, ok := findHint(hints, "Enter", "select"); !ok {
        t.Fatal("expected Enter select hint for ModalFirstRun")
    }
    // Esc MUST NOT appear — documented exception: selection is required (Story 1.5 AC2)
    for _, h := range hints {
        if h.Key == "Esc" {
            t.Fatalf("Esc hint must not appear in ModalFirstRun (selection required): got %+v", h)
        }
    }
}
```

**Important:** The `findHint` helper already exists in `main_hints_test.go`. The tests use `m.popup.mode` and `m.activeModal` directly — both are exported-accessible within the same `app` package.

### Files to Modify

| File | Change |
|---|---|
| `internal/app/hints.go` | Rename `currentViewHints` body → `tableViewHints`; add dispatch in `currentViewHints`; add `contextSwitcherHints`, `searchPopupHints`, `skinPickerHints` |
| `internal/app/modal.go` | Add `HintLine` to 6 OverlayCenter modal registrations |
| `internal/app/main_hints_test.go` | Add 8 new per-view dispatch tests |

### Files NOT to Modify

- `internal/app/view.go` — footer caller `filterHints(currentViewHints(*m), width)` stays unchanged ✓
- `internal/app/quick_wins_test.go` — tests don't call `currentViewHints` directly with popup/modal state ✓
- Any file in `internal/operaton/` — never modify

### Architecture Compliance

From `_bmad/planning-artifacts/architecture.md` Decision 3:
> "The footer renderer calls `currentViewHints(m)` which dispatches to the active view's hint function."

This story completes that intent. After Story 2.1, `currentViewHints` was a single monolithic function; this story introduces the dispatch layer so it truly "dispatches to the active view's hint function."

### Learnings from Stories 2.1, 1.4, 1.5

- `newTestModel(t)` initializes with `m.activeModal = ModalNone` and `m.popup.mode = popupModeNone` → all existing `TestCurrentViewHints_*` tests will pass unchanged (they test the default dispatch path to `tableViewHints`)
- The `modalRegistry` is populated by `init()` in `modal.go` — it's accessible in test code since tests are in `package app`
- Keep HintLine hints short (Key ≤ 10 chars) — they need to render in the footer which can be narrow
- Story 1.5 established ModalFirstRun's Esc exception: the `HintLine` for `ModalFirstRun` must NOT include any `Esc` hint
- `gofmt -w .` after every edit pass; `go vet ./...` before final test run

### References

- [Source: `internal/app/hints.go`] — `currentViewHints`, `filterHints`, `Hint` struct (Story 2.1 implementation)
- [Source: `internal/app/modal.go`] — `ModalConfig.HintLine`, `modalRegistry`, all `registerModal` calls
- [Source: `internal/app/view.go:63`] — `filterHints(currentViewHints(*m), width)` — footer call site (unchanged by this story)
- [Source: `internal/app/model.go:181-188`] — `popupMode` enum (`popupModeNone`, `popupModeContext`, `popupModeSkin`, `popupModeSearch`)
- [Source: `internal/app/model.go:192-199`] — `popup` struct with `mode`, `hint`, `input`, `cursor` fields
- [Source: `internal/app/main_hints_test.go`] — existing hint tests; `findHint` helper; test model patterns
- [Source: `_bmad/planning-artifacts/architecture.md#Decision 3: Footer Hint Rendering — Push Model`] — dispatch intent
- [Source: `_bmad/planning-artifacts/epics.md#Story 2.2`] — full AC definition
- [Source: `_bmad/implementation-artifacts/2-1-footer-hint-push-model.md`] — Story 2.1 learnings and hint declarations

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

None.

### Completion Notes List

- Task 1: Added `HintLine` to 6 OverlayCenter modals in `modal.go` — ModalConfirmDelete, ModalConfirmQuit, ModalSort, ModalEnvironment, ModalEdit, ModalFirstRun. ModalFirstRun intentionally omits Esc per Story 1.5 AC2.
- Task 2: Refactored `currentViewHints` to dispatch by priority: active modal → popup mode → table view. Extracted `tableViewHints`, added `contextSwitcherHints`, `searchPopupHints`, `skinPickerHints`. All 5 functions hit 100% coverage.
- Task 3: All 10 existing `TestCurrentViewHints_*` and `TestFilterHints_*` tests pass unchanged — dispatch to `tableViewHints` in default case is transparent to existing tests.
- Task 4: Added 8 new per-view dispatch tests covering all popup modes and representative modals including the ModalFirstRun Esc-exception.
- Task 5: Full suite passes (0 regressions across all packages), `go vet` clean, `gofmt` no diff, `hints.go` coverage 100%. `specification.md` updated with Per-View Hint Dispatch section.

### File List

- `internal/app/hints.go` — refactored `currentViewHints` dispatch; added `tableViewHints`, `contextSwitcherHints`, `searchPopupHints`, `skinPickerHints`
- `internal/app/modal.go` — added `HintLine` to 6 OverlayCenter modal registrations
- `internal/app/main_hints_test.go` — added 8 new per-view dispatch tests
- `specification.md` — added Per-View Hint Dispatch section under Key Hint Priority System
- `_bmad/implementation-artifacts/2-2-per-view-hint-functions.md` — story file (this file)
- `_bmad/implementation-artifacts/sprint-status.yaml` — status updated to `review`
