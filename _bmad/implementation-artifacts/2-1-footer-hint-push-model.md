# Story 2.1: Footer Hint Push Model

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want the primary available actions for the current view visible in the footer at all times,
so that I can discover what I can do without opening the `?` help screen.

## Acceptance Criteria

1. **Given** `Hint{Key, Label, MinWidth, Priority}` is defined in `internal/app/hints.go`
   **When** the footer renderer (`renderCompactHeader`) is called
   **Then** it receives a `[]Hint` slice from `currentViewHints(m)` — the footer renderer itself contains no hint declarations
   **And** hints with `MinWidth: 0` are always visible regardless of terminal width
   **And** hints with `MinWidth > 0` are hidden when the terminal is narrower than that threshold
   **And** when multiple hints must be dropped due to width, higher `Priority` integer hints are dropped first (lower int = higher priority)

2. **Given** `internal/app/main_hints_test.go` is created
   **When** `make cover` is run
   **Then** line coverage on `hints.go` is ≥ 80%

## Context & Background

### Pre-Work Completed

- **Story 1.1 (done):** Created `internal/app/hints.go` stub with the `Hint{Key, Label, MinWidth, Priority int}` struct. This is the correct new type.
- **Story 1.2 (done):** State transition contract enforced — unrelated to hints.

### Current State of Hint System (Code to Replace)

**Old type — `KeyHint` in `internal/app/model.go:133-138`:**
```go
// KeyHint represents a keyboard shortcut with priority
type KeyHint struct {
    Key         string
    Description string
    Priority    int // 1=always visible, 9=only on wide terminals
}
```
Note: `KeyHint.Description` corresponds to `Hint.Label` in the new type.

**Old function — `getKeyHints(width int) []KeyHint` in `internal/app/view.go:21-84`:**
- Method on `*model`; takes terminal `width` and uses inline `if width >= N` checks
- Mixes hint declarations with model-state queries and width-threshold logic
- Returns `[]KeyHint` (old type)

**Footer renderer — `renderCompactHeader` in `internal/app/view.go:129-154`:**
```go
hints := m.getKeyHints(width)                         // ← old call-site
sort.Slice(hints, func(i, j int) bool { ... })        // ← sorting happens here
for _, hint := range hints {
    part := fmt.Sprintf("%s %s", hint.Key, hint.Description) // ← uses .Description
    ...
}
```

**Tests using old API — `internal/app/quick_wins_test.go`:**
- `TestWin2_ContextAwareKeyHints` (line 67-113): calls `m.getKeyHints(100)`, checks `h.Description`
- `TestWin2_KeyHintsRespectTerminalWidth` (line 115-143): calls `m.getKeyHints(80/100)`, checks `h.Key`/`h.Description`
- Line 266: calls `m.getKeyHints(100)` in a third test

### Width Thresholds in Current `getKeyHints()` to Migrate to `Hint.MinWidth`

| Current `if width >= N` | New `Hint.MinWidth` |
|---|---|
| no threshold (always shown) | `0` |
| `width >= 88` | `88` |
| `width >= 90` | `90` |
| `width >= 100` | `100` |
| `width >= 110` | `110` |
| `width >= 112` | `112` |

Conditional hints based on **model state** (not width) remain `MinWidth: 0` and are included/excluded by `currentViewHints()` based on state checks (e.g., `len(m.navigationStack) > 0`).

### Architecture Decision 3 — Footer Hint Push Model

From `_bmad/planning-artifacts/architecture.md`:

> **Decision:** Each view handler returns a `[]Hint` slice declaring the hints it contributes. The footer renderer is stateless — it receives the hint slice and terminal width, filters by each hint's `minWidth` threshold, and renders the result.
>
> View handlers produce hints at render time — hints are not stored in model state. The footer renderer calls `currentViewHints(m)` which dispatches to the active view's hint function. This makes hints independently testable: `viewHints(m)` → `[]Hint`, no rendering required.

Function signatures required by architecture:
```go
// hints.go
func currentViewHints(m model) []Hint         // produces all hints for current view
func filterHints(hints []Hint, width int) []Hint  // sorts by Priority, applies MinWidth filter
```

The separation: `currentViewHints` is width-unaware (returns all candidates with MinWidth metadata); `filterHints` applies the width gate. Footer renderer chains both:
```go
hints := filterHints(currentViewHints(m), width)
```

## Tasks / Subtasks

- [x] Task 1: Implement `filterHints` and `currentViewHints` in `hints.go` (AC: 1)
  - [x] Add `filterHints(hints []Hint, width int) []Hint` to `hints.go`:
    - Sort the input slice by `Priority` ascending (lower int = higher priority = shown first)
    - Filter out any hint where `MinWidth > 0 && width < MinWidth`
    - Return the sorted, filtered slice
  - [x] Add `currentViewHints(m model) []Hint` to `hints.go`:
    - Migrate all hint declarations from `view.go:getKeyHints()` to this function
    - Each width-threshold inline `if width >= N` becomes `Hint.MinWidth = N` on the hint struct
    - Each conditional hint (based on model state: `navigationStack`, `breadcrumb`, `findTableDef`, etc.) remains conditional in `currentViewHints()` but with `MinWidth: 0`
    - `KeyHint.Description` maps to `Hint.Label` in the new struct
    - **Do NOT take `width` as a parameter** — `currentViewHints` is width-unaware; all width logic lives in `filterHints`

- [x] Task 2: Update `view.go` footer renderer to use new hint API (AC: 1)
  - [x] In `renderCompactHeader`, replace `m.getKeyHints(width)` + inline sort with:
    ```go
    hints := filterHints(currentViewHints(m), width)
    ```
  - [x] Update hint field reference: `hint.Description` → `hint.Label`
  - [x] Remove the `sort.Slice` call from `renderCompactHeader` (sorting now in `filterHints`)
  - [x] Delete `getKeyHints()` function from `view.go` entirely

- [x] Task 3: Remove `KeyHint` struct from `model.go` (AC: 1)
  - [x] Verify `KeyHint` is no longer referenced anywhere (after updating quick_wins_test.go below)
  - [x] Delete the `KeyHint` struct definition from `model.go:133-138`

- [x] Task 4: Update existing hint tests in `quick_wins_test.go` (prerequisite for passing tests)
  - [x] `TestWin2_ContextAwareKeyHints`: change `m.getKeyHints(100)` → `currentViewHints(m)`, change all `h.Description` → `h.Label`
  - [x] `TestWin2_KeyHintsRespectTerminalWidth`: change `m.getKeyHints(80)` → `filterHints(currentViewHints(m), 80)` and similarly for width=100, change `h.Description` → `h.Label`
  - [x] Line ~266 (third caller): update similarly

- [x] Task 5: Create `internal/app/main_hints_test.go` with ≥80% coverage on `hints.go` (AC: 2)
  - [x] Test `filterHints` — MinWidth = 0 always passes through
  - [x] Test `filterHints` — hint with MinWidth > width is excluded
  - [x] Test `filterHints` — hint with MinWidth ≤ width is included
  - [x] Test `filterHints` — output is sorted by Priority ascending
  - [x] Test `filterHints` — empty input returns empty output
  - [x] Test `currentViewHints` — always includes `?`/`help` (Priority 1, MinWidth 0)
  - [x] Test `currentViewHints` — includes `Esc`/`back` only when `navigationStack` is non-empty
  - [x] Test `currentViewHints` — includes `Enter`/`drill` only when current table def has a drilldown
  - [x] Test `currentViewHints` — `Ctrl+r`/`refresh` hint has MinWidth: 90

- [x] Task 6: Verify and run tests (AC: 2)
  - [x] `make test` passes with no failures
  - [x] `make cover` shows ≥ 80% line coverage on `hints.go`
  - [x] `go vet ./...` reports no issues
  - [x] `gofmt -w .` leaves no changes

## Dev Notes

### Implementation Order

1. Implement `filterHints` + `currentViewHints` in `hints.go` (Task 1)
2. Update `view.go` footer renderer (Task 2) — at this point the app compiles but uses new types
3. Delete `getKeyHints()` from `view.go`
4. Update `quick_wins_test.go` so existing tests pass (Task 4)
5. Remove `KeyHint` from `model.go` (Task 3) — only safe after step 4
6. Create `main_hints_test.go` (Task 5)
7. `make test` + `make cover` (Task 6)

### `currentViewHints` Hint Declarations

Migrate the following from `getKeyHints()` to `currentViewHints()`. Use `Hint.MinWidth` for width thresholds:

```go
// Always shown (MinWidth: 0)
{Key: "?",         Label: "help",    MinWidth: 0,   Priority: 1},
{Key: ":",         Label: "switch",  MinWidth: 0,   Priority: 2},
{Key: "↑↓",        Label: "nav",     MinWidth: 0,   Priority: 3},
{Key: "/",         Label: "find",    MinWidth: 0,   Priority: 3},
{Key: "PgDn/PgUp", Label: "page",    MinWidth: 0,   Priority: 3},

// Conditional on model state (MinWidth: 0 but only appended when condition holds)
// if def != nil && def.Drilldown != nil:
{Key: "Enter",     Label: "drill",   MinWidth: 0,   Priority: 4},
// if m.hasEditableColumns():
{Key: "e",         Label: "edit",    MinWidth: 0,   Priority: 4},
// if len(m.navigationStack) > 0:
{Key: "Esc",       Label: "back",    MinWidth: 0,   Priority: 5},
// if len(m.breadcrumb) > 1:
{Key: "1–N",       Label: "back",    MinWidth: 0,   Priority: 5},  // N = len(breadcrumb)-1

// Width-gated (MinWidth matches current threshold)
{Key: "s",         Label: "sort",    MinWidth: 88,  Priority: 5},   // was: width >= 88
{Key: "Ctrl+r",    Label: "refresh", MinWidth: 90,  Priority: 6},   // was: width >= 90
{Key: "Ctrl+T",    Label: "skin",    MinWidth: 90,  Priority: 6},
{Key: "Ctrl+e",    Label: "env",     MinWidth: 90,  Priority: 6},
{Key: "Ctrl+Space",Label: "actions", MinWidth: 100, Priority: 6},   // was: Space (key updated to Ctrl+Space per FR16)
{Key: "J",         Label: "json",    MinWidth: 112, Priority: 6},   // was: y/detail (key updated per FR31)
{Key: "Ctrl+c",    Label: "quit",    MinWidth: 110, Priority: 8},   // was: width >= 110
```

**Note on key corrections:** The current `getKeyHints()` still uses the old keys `"Space"` for actions and `"y"` for detail. These must be updated to `"Ctrl+Space"` (FR16) and `"J"` (FR31) in `currentViewHints()`. The label for `J` should be `"json"` not `"detail"`.

The `m.activeModal == ModalNone` guard currently on `s`/`Space`/`y` hints should be **dropped** — the footer always shows the primary actions regardless of modal state (the modal itself renders on top and provides its own hint line per Story 1.1).

### `filterHints` Implementation Notes

```go
func filterHints(hints []Hint, width int) []Hint {
    // Sort: lower Priority int = higher importance = shown first
    sort.Slice(hints, func(i, j int) bool {
        return hints[i].Priority < hints[j].Priority
    })
    // Filter: remove hints that require more width than available
    result := hints[:0]   // reuse slice, avoids allocation for typical case
    for _, h := range hints {
        if h.MinWidth == 0 || width >= h.MinWidth {
            result = append(result, h)
        }
    }
    return result
}
```

Note: `hints[:0]` reuse only works correctly if the caller doesn't need to retain the original slice after calling `filterHints`. Since we always call it at render time and discard the result, this is safe.

### Footer Renderer After Migration

In `renderCompactHeader`, the hints section changes from:
```go
hints := m.getKeyHints(width)
sort.Slice(hints, func(i, j int) bool { return hints[i].Priority < hints[j].Priority })
...
for _, hint := range hints {
    part := fmt.Sprintf("%s %s", hint.Key, hint.Description)
```
to:
```go
hints := filterHints(currentViewHints(m), width)
...
for _, hint := range hints {
    part := fmt.Sprintf("%s %s", hint.Key, hint.Label)
```

### Test Helpers

`newTestModel(t)` is defined in `internal/app/main_test.go` and creates a model with default test state. Use it for hint tests. The model needs a skin loaded to avoid nil pointer on `m.skin`. Check `newTestModel` implementation before writing tests.

## Dev Agent Record

_To be filled by the implementing agent._

- **Agent Model Used:** GPT-5.3-Codex (gpt-5.3-codex)
- **Completion Notes:**
  - Implemented `filterHints` and `currentViewHints` in `internal/app/hints.go` and moved all footer hint declarations out of `view.go`.
  - Updated `renderCompactHeader` to use `filterHints(currentViewHints(*m), width)` and switched footer hint rendering to `hint.Label`.
  - Removed deprecated `KeyHint` model type and migrated existing quick-win tests to the new hint API.
  - Added `internal/app/main_hints_test.go` with focused tests for filtering, sorting, conditional hints, and refresh width constraints.
  - Validation run: `go test ./internal/app`, `make test`, `make cover`, `go vet ./...`, `gofmt -w .`; `hints.go` coverage is `filterHints=100%`, `currentViewHints=90%`.
- **File List:**
  - `internal/app/hints.go` — modified (add `filterHints`, `currentViewHints`)
  - `internal/app/view.go` — modified (replace `getKeyHints` call-site, update `hint.Description` → `hint.Label`, remove `sort.Slice`, delete `getKeyHints` function; update help screen key text)
  - `internal/app/model.go` — modified (remove `KeyHint` struct)
  - `internal/app/quick_wins_test.go` — modified (update hint test callers)
  - `internal/app/main_hints_test.go` — created
  - `internal/app/update.go` — modified (Space→`ctrl+space` for actions, `y`→`J` for detail view, FR16/FR31)
  - `internal/app/nav.go` — modified (`y`→`J` key on View as JSON action item, FR31)
  - `internal/app/actions_test.go` — modified (update key assertions to `J`/`ctrl+space`)
  - `internal/app/detail_test.go` — modified (rename tests and update key strings to `J`)

### Change Log

| Date | Change | Reason |
|---|---|---|
| 2026-03-04 | Story file created | First backlog story in Epic 2 |
| 2026-03-04 | Implemented footer hint push model and tests | Complete Story 2.1 acceptance criteria |
| 2026-03-04 | Migrated keybindings Space→`Ctrl+Space` (FR16), `y`→`J` (FR31); updated help screen text; updated test files | CR finding: hint labels must match actual keybindings |
| 2026-03-04 | Fixed `Ctrl+T` hint key casing; added `hasEditableColumns` test coverage | CR findings L1/L2 |
