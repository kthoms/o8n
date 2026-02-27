# Story: State Transition Contract and Bug Fixes

## Summary

Introduce a centralized state transition contract that eliminates state leakage across environment switches, context switches, and breadcrumb navigation. Fix 4 confirmed interaction bugs and 1 design gap found during the state machine audit.

## Motivation

The state machine has 10 interactive states and 60+ key bindings with priority-based routing. The routing itself is sound — no stuck states, all modals have exit paths. But state *cleanup* between transitions is inconsistent:

1. Environment switch doesn't clear `navigationStack` — user sees data from the wrong environment on Esc
2. Environment switch doesn't clear `genericParams` — drilldown filters leak across environments
3. Context switch doesn't clear `sortColumn` — column index carries over to unrelated table
4. Breadcrumb navigation bypasses the navigation stack — Esc behavior becomes unpredictable
5. Row deletion doesn't adjust cursor bounds — potential out-of-bounds after deleting last row

A centralized `prepareStateTransition(scope)` function solves all of these with a single abstraction.

## Acceptance Criteria

### Bug Fix: Environment Switch State Leakage

- [x] **AC-1:** `navigationStack` is cleared (`nil`) when switching environments via `Ctrl+E`
- [x] **AC-2:** `genericParams` is cleared (empty map) when switching environments
- [x] **AC-3:** `breadcrumb` is reset to `[currentRoot]` when switching environments
- [x] **AC-4:** `selectedDefinitionKey` and `selectedInstanceID` are cleared on environment switch
- [x] **AC-5:** After switching environments, `Esc` does not restore state from the previous environment

### Bug Fix: Sort State Leakage

- [x] **AC-6:** `sortColumn` is reset to `-1` when switching context via `:`
- [x] **AC-7:** `sortAscending` is reset to `true` when switching context
- [x] **AC-8:** Sort indicator is not shown in the new context's table header after switch

### Bug Fix: Breadcrumb Navigation Stack Mismatch

- [x] **AC-9:** When pressing `1`-`4` for breadcrumb navigation, the navigation stack is truncated to the target depth (not left intact)
- [x] **AC-10:** Pressing `1` (root) clears the navigation stack entirely
- [x] **AC-11:** Pressing `2` with a 3-level breadcrumb pops the stack once (removes the deepest level)
- [x] **AC-12:** After breadcrumb navigation, pressing `Esc` pops to the correct parent level (not a stale deeper level)

### Bug Fix: Cursor Bounds After Row Deletion

- [x] **AC-13:** After deleting a row (terminatedMsg, actionExecutedMsg), if the cursor exceeds the remaining row count, it is adjusted to the last row
- [x] **AC-14:** After deleting the only remaining row, the cursor is set to 0 (empty state)

### State Transition Contract

- [x] **AC-15:** A `prepareStateTransition(scope)` function (or equivalent clear method) exists that encapsulates state cleanup for each transition type
- [x] **AC-16:** All transition call sites use this function instead of ad-hoc state clearing
- [x] **AC-17:** The transition scopes are:

| Scope | Clears |
|---|---|
| `transitionEnvSwitch` | navStack, genericParams, breadcrumb, selectedDefinitionKey, selectedInstanceID, sort, search, searchTerm |
| `transitionContextSwitch` | navStack, genericParams, sort, search, searchTerm |
| `transitionDrilldown` | sort, search (navStack push happens separately) |
| `transitionBack` | sort, search (navStack pop restores previous state) |
| `transitionBreadcrumb(depth)` | navStack truncated to depth, sort, search |

## Tasks

### Task 1: Implement `prepareStateTransition`

**File:** `internal/app/nav.go` (or new file `internal/app/transition.go`)

```go
type transitionScope int

const (
    transitionEnvSwitch transitionScope = iota
    transitionContextSwitch
    transitionDrilldown
    transitionBack
    transitionBreadcrumb
)

func (m *model) prepareStateTransition(scope transitionScope, depth ...int) {
    // Clear sort state for all transitions
    m.sortColumn = -1
    m.sortAscending = true

    // Clear search state for all transitions
    m.searchTerm = ""
    m.originalRows = nil
    m.filteredRows = nil

    switch scope {
    case transitionEnvSwitch:
        m.navigationStack = nil
        m.genericParams = make(map[string]string)
        m.selectedDefinitionKey = ""
        m.selectedInstanceID = ""
        // breadcrumb reset handled by caller (depends on new root)

    case transitionContextSwitch:
        m.navigationStack = nil
        m.genericParams = make(map[string]string)

    case transitionDrilldown:
        // navStack push handled by caller
        // sort/search already cleared above

    case transitionBack:
        // navStack pop handled by caller
        // sort/search already cleared above

    case transitionBreadcrumb:
        if len(depth) > 0 && depth[0] >= 0 {
            if depth[0] == 0 {
                m.navigationStack = nil
            } else if depth[0] < len(m.navigationStack) {
                m.navigationStack = m.navigationStack[:depth[0]]
            }
        }
    }
}
```

### Task 2: Wire environment switch through contract

**File:** `internal/app/update.go` (environment switch handler, ~line 273-288)

1. Replace ad-hoc `resetViews()` with `m.prepareStateTransition(transitionEnvSwitch)`
2. Keep the env-specific logic (switching client, triggering fetch, health check)
3. Verify that `resetViews()` is only called from the env switch path — if so, remove it

### Task 3: Wire context switch through contract

**File:** `internal/app/update.go` (context switch handler, ~line 814-845)

1. Add `m.prepareStateTransition(transitionContextSwitch)` before setting `m.currentRoot`
2. Remove the existing ad-hoc `m.navigationStack = nil` and `m.genericParams = make(...)` lines — the contract handles it

### Task 4: Wire breadcrumb navigation through contract

**File:** `internal/app/update.go` (breadcrumb handler, ~line 296-315)

1. Replace the current direct state manipulation with:
   ```go
   m.prepareStateTransition(transitionBreadcrumb, idx)
   ```
2. Keep the breadcrumb-specific logic (setting `m.currentRoot`, `m.breadcrumb`, triggering fetch)

### Task 5: Wire drilldown and back through contract

**File:** `internal/app/update.go` (Enter handler, Esc handler)

1. Before pushing navStack on Enter: `m.prepareStateTransition(transitionDrilldown)`
2. Before popping navStack on Esc: `m.prepareStateTransition(transitionBack)`
3. The navStack push/pop remains in the caller — the contract only handles cleanup

### Task 6: Fix cursor bounds after row deletion

**File:** `internal/app/update.go` (terminatedMsg handler, actionExecutedMsg handler)

After removing a row from the table:
```go
if m.table.Cursor() >= len(m.table.Rows()) {
    newCursor := len(m.table.Rows()) - 1
    if newCursor < 0 {
        newCursor = 0
    }
    m.table.SetCursor(newCursor)
}
```

Apply in:
- `terminatedMsg` handler (~line 1444-1464)
- `actionExecutedMsg` handler (if it removes rows)
- Any other handler that calls `removeInstance()` or equivalent

### Task 7: Update tests

1. **Env switch test:** Drill into instances in env A, switch to env B, press Esc — verify navStack is empty, not restoring env A state
2. **Env switch params test:** Set genericParams in env A, switch to env B — verify genericParams is empty
3. **Context switch sort test:** Sort by column 3, switch context — verify sortColumn == -1
4. **Breadcrumb nav test:** Drill 3 levels deep, press `1` — verify navStack is nil
5. **Breadcrumb depth test:** Drill 3 levels deep, press `2` — verify navStack has 1 entry (not 2)
6. **Breadcrumb Esc test:** Drill 3 levels, press `2`, then press `Esc` — verify pop to root (not to level 3)
7. **Cursor bounds test:** Delete last row in table — verify cursor adjusted to new last row
8. **Cursor bounds empty test:** Delete only row — verify cursor is 0
9. **Contract coverage test:** Verify `prepareStateTransition` is called in all 5 transition paths

## Technical Notes

- `resetViews()` in `model.go` currently only clears table data — it does NOT clear navStack, genericParams, sort, or search. This is why the bugs exist. The new contract replaces it.
- The `prepareStateTransition` function mutates the model directly (pointer receiver). All callers in `Update()` already work with `m` by reference.
- Sort state (`sortColumn`, `sortAscending`) is purely client-side. Clearing it on transition is always safe — no API interaction needed.
- `searchTerm` persists after search lock (Enter). Clearing it on transition means the user loses their search filter when navigating — this is correct behavior (same as k9s).
- The breadcrumb jump with stack truncation must handle the edge case where `depth[0] >= len(navigationStack)` — in that case, no truncation needed (already at or past target depth).
