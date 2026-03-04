# Story 1.3: Environment & Context Switching Correctness

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want switching between environments and resource contexts to always start fresh with no stale data,
So that I can trust the view I see reflects exactly the environment and context I selected.

## Acceptance Criteria

1. **Given** the operator is in any resource context with an active filter, modal, or cursor position
   **When** the operator switches to a different environment
   **Then** `prepareStateTransition(TransitionFull)` is called, clearing all prior view state
   **And** the new environment's accent color and API URL are active immediately
   **And** the table reloads from the new environment with no rows, filters, or modal from the previous environment visible

2. **Given** the operator is in any resource context with active state
   **When** the operator uses `:` to switch to a different resource type
   **Then** `prepareStateTransition(TransitionFull)` is called before the new context loads
   **And** the new context view has no residual search query, sort order, modal, or cursor from the previous context

## Tasks / Subtasks

- [x] Task 1: Audit callsites — confirm Story 1.2 migration is complete + check `nextEnvironment()` (AC: 1, 2)
  - [x] Verify `update.go:313` (env switch Enter handler) uses `TransitionFull` not `transitionEnvSwitch`
  - [x] Verify `update.go:1005` (context switch handler) uses `TransitionFull` not `transitionContextSwitch`
  - [x] If NOT migrated yet (Story 1.2 incomplete): change `transitionEnvSwitch` → `TransitionFull` and `transitionContextSwitch` → `TransitionFull` in those two callsites
  - [x] **Audit `nextEnvironment()` in `commands.go:23-39`**: `grep -rn "nextEnvironment" internal/app/` — if called from any key handler, this method currently omits `prepareStateTransition(TransitionFull)` and is a state leakage bug; fix by adding the transition call and breadcrumb reset before returning the `tea.Cmd`
  - [x] Run `go build ./...` — must compile with zero errors before writing any tests

- [x] Task 2: Update `state_transition_test.go` to use new `TransitionType` constants (AC: 1, 2)
  - [x] Replace `transitionEnvSwitch` with `TransitionFull` in all tests
  - [x] Replace `transitionContextSwitch` with `TransitionFull` in all tests
  - [x] Replace `transitionBack` with `TransitionPop` in all tests
  - [x] Replace `transitionDrilldown` with `TransitionDrillDown` in all tests
  - [x] Remove or rewrite `TestBreadcrumbNavToRootClearsNavStack` and `TestBreadcrumbNavToDepth1TruncatesNavStack` — the `transitionBreadcrumb` variant is gone; breadcrumb logic now lives in `navigateToBreadcrumb()` (caller truncates stack, then calls `TransitionPop`). Replace with `TestBreadcrumbNavigation*` tests that call `navigateToBreadcrumb()` directly.
  - [x] Update `TestAllTransitionsClearSortAndSearch` to iterate `[]TransitionType{TransitionFull, TransitionDrillDown}` (TransitionPop does NOT clear sort/search — it restores them from stack)
  - [x] Remove `TestPrepareStateTransitionExists` — the new `main_transition_test.go` (Story 1.2) covers this
  - [x] Run `go test ./internal/app/... -run TestEnv -v` — all env switch tests pass

- [x] Task 3: Write new behavioral tests in `main_env_switch_test.go` (AC: 1, 2)
  - [x] `TestEnvSwitch_ClearsActiveModal` — active modal set before switch → `ModalNone` after
  - [x] `TestEnvSwitch_ClearsSearchState` — `searchTerm` and `searchMode` active before switch → cleared after
  - [x] `TestEnvSwitch_SetsNewEnvironment` — call `switchToEnvironment("staging")` → `m.currentEnv == "staging"`
  - [x] `TestEnvSwitch_BreadcrumbReset` — verify breadcrumb is `[]string{m.currentRoot}` after env switch sequence
  - [x] `TestEnvSwitch_EscAfterSwitchIsNoop` — Esc after env switch with empty nav stack → no panic, no state restore
  - [x] `TestContextSwitch_ClearsModalAndSearch` — active modal + search before `:` switch → all cleared after
  - [x] `TestContextSwitch_SetsNewRoot` — after context switch to "incidents": `m.currentRoot == "incidents"`, `m.viewMode == "incidents"`, `m.contentHeader == "incidents"`
  - [x] `TestContextSwitch_BreadcrumbIsSingleElement` — breadcrumb has exactly 1 element after context switch
  - [x] `TestContextSwitch_ClearsNavigationStack` — nav stack with 2 entries → nil after context switch

- [x] Task 4: Verify all tests pass (AC: all)
  - [x] `make test` passes with zero regressions
  - [x] `go vet ./...` passes
  - [x] `gofmt -w .` produces no changes

## Dev Notes

### Dependency on Story 1.2

**Story 1.2 MUST be complete before implementing Story 1.3.**

Story 1.2 defines `TransitionType` + `prepareStateTransition(t TransitionType)` in `transition.go` and migrates all nav callsites. As of Story 1.3 start, `transition.go` is already updated, but `update.go` still has old constants that cause a compile error:

```
internal/app/update.go:919:30: undefined: transitionBack
```

**Current failing callsites in `update.go` (from `go build`):**
- `update.go:313`: `m.prepareStateTransition(transitionEnvSwitch)` → must become `m.prepareStateTransition(TransitionFull)`
- `update.go:919`: `m.prepareStateTransition(transitionBack)` → must become `m.prepareStateTransition(TransitionPop)` (Story 1.2 scope)
- `update.go:1005`: `m.prepareStateTransition(transitionContextSwitch)` → must become `m.prepareStateTransition(TransitionFull)`

**Current failing callsites in `nav.go`:**
- `nav.go:365`: `m.prepareStateTransition(transitionBreadcrumb, idx)` → Story 1.2 scope
- `nav.go:385`: `m.prepareStateTransition(transitionDrilldown)` → Story 1.2 scope

**If Story 1.2 is NOT complete, Task 1 in this story handles the env/context-switch callsites (`update.go:313` and `update.go:1005`). The Esc/back and nav.go callsites MUST be fixed by Story 1.2 first.**

### Current Code State (`transition.go`)

`transition.go` already has the NEW API fully implemented:

```go
type TransitionType int
const (
    TransitionFull      TransitionType = iota
    TransitionDrillDown
    TransitionPop
)

func (m *model) prepareStateTransition(t TransitionType) {
    switch t {
    case TransitionFull:
        m.activeModal = ModalNone
        m.footerError = ""
        m.footerStatusKind = footerStatusNone
        m.sortColumn = -1
        m.sortAscending = true
        m.searchTerm = ""
        m.searchMode = false
        m.searchInput.Blur()
        m.originalRows = nil
        m.filteredRows = nil
        m.navigationStack = nil
        m.genericParams = make(map[string]string)
        m.selectedDefinitionKey = ""
        m.selectedInstanceID = ""
        m.table.SetCursor(0)
        if m.popup.mode != popupModeNone {
            m.popup.mode = popupModeNone
            ...
        }
    case TransitionDrillDown:
        // push-before-clear
        ...
    case TransitionPop:
        // restore from top of stack
        ...
    }
}
```

### Current Env Switch Sequence in `update.go:308-322`

```go
// inside ModalEnvironment key handler (update.go):
case "enter":
    if m.envPopupCursor >= 0 && m.envPopupCursor < len(m.envNames) {
        targetEnv := m.envNames[m.envPopupCursor]
        m.activeModal = ModalNone
        if targetEnv != m.currentEnv {
            m.switchToEnvironment(targetEnv)    // sets currentEnv + applyStyle()
            m.prepareStateTransition(transitionEnvSwitch)  // ← CHANGE TO TransitionFull
            m.breadcrumb = []string{m.currentRoot}
            m.isLoading = true
            m.apiCallStarted = time.Now()
            return m, tea.Batch(m.fetchDefinitionsCmd(), ...)
        }
    }
```

After fix:
```go
m.switchToEnvironment(targetEnv)
m.prepareStateTransition(TransitionFull)  // clears modal, search, sort, cursor, navStack, etc.
m.breadcrumb = []string{m.currentRoot}
```

Note: `m.activeModal = ModalNone` before the call is redundant since `TransitionFull` also clears it, but it's harmless — leave it or remove it (either is fine).

### Current Context Switch Sequence in `update.go:990-1025`

```go
// inside popup popupModeSearch Enter handler (update.go):
m.currentRoot = rc
m.popup.mode = popupModeNone
m.popup.input = ""
m.popup.cursor = -1
m.popup.offset = 0
m.footerError = ""               // redundant after TransitionFull; harmless to leave
m.prepareStateTransition(transitionContextSwitch)  // ← CHANGE TO TransitionFull
m.breadcrumb = []string{rc}
m.contentHeader = rc
m.viewMode = rc
// fetch logic...
```

After fix:
```go
m.currentRoot = rc
m.popup.mode = popupModeNone
m.popup.input = ""
m.popup.cursor = -1
m.popup.offset = 0
m.prepareStateTransition(TransitionFull)  // clears modal, search, sort, cursor, navStack etc.
m.breadcrumb = []string{rc}
m.contentHeader = rc
m.viewMode = rc
// fetch logic...
```

**Note:** `popup.mode/input/cursor/offset` reset is NOT part of `TransitionFull` (it's specific to popup close). Keep those 4 lines — they clean up the context switcher UI state.

### `nextEnvironment()` — commands.go:23-39 — Potential State Leakage Bug

`commands.go` contains a `nextEnvironment()` method (original env cycling implementation):

```go
func (m *model) nextEnvironment() tea.Cmd {
    // cycles m.currentEnv to the next in m.envNames
    idx = (idx + 1) % len(m.envNames)
    m.currentEnv = m.envNames[idx]
    m.applyStyle()
    return m.checkEnvironmentHealthCmd(m.currentEnv)
}
```

This does NOT call `prepareStateTransition(TransitionFull)` — it only cycles `m.currentEnv`, applies style, and checks health. If this function is still called in any key handler path, it leaves all prior view state (search, sort, modal, navStack) intact — a state leakage bug violating AC 1.

**Task 1 action:** `grep -rn "nextEnvironment" internal/app/`
- If zero results in key handlers: dead code or test-only; safe to leave
- If called in update.go: replace call with the correct sequence (see env switch pattern in Task 2 audit)

If a fix is needed, pattern:
```go
// In the key handler, before returning the cmd:
m.switchToEnvironment(m.envNames[newIdx])
m.prepareStateTransition(TransitionFull)
m.breadcrumb = []string{m.currentRoot}
m.isLoading = true
m.apiCallStarted = time.Now()
return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), spinnerTickCmd(), m.saveStateCmd())
```

### `switchToEnvironment()` — nav.go:207-210

```go
func (m *model) switchToEnvironment(name string) {
    m.currentEnv = name
    m.applyStyle()    // applies the new env's skin (accent color, env_name role, etc.)
}
```

This sets `m.currentEnv` and triggers `m.applyStyle()` which rebuilds all Lipgloss styles using the new environment's `ui_color` and skin. This is how the accent color changes immediately on env switch.

### `state_transition_test.go` — Existing Tests to Update

File: `internal/app/state_transition_test.go` — currently uses OLD `transitionScope` constants. After Story 1.2 removes those constants, this file fails to compile.

**Tests to UPDATE (keep but update constants):**
- `TestEnvSwitchClearsNavigationStack`: `transitionEnvSwitch` → `TransitionFull`
- `TestEnvSwitchClearsGenericParams`: `transitionEnvSwitch` → `TransitionFull`
- `TestEnvSwitchClearsSelectedKeys`: `transitionEnvSwitch` → `TransitionFull`
- `TestContextSwitchClearsSortState`: `transitionContextSwitch` → `TransitionFull`
- `TestContextSwitchClearsNavStack`: `transitionContextSwitch` → `TransitionFull`
- `TestEscAfterEnvSwitchDoesNotRestoreOldStack`: `transitionEnvSwitch` → `TransitionFull`
- `TestCursorBoundsAfterTerminate` — no changes needed (uses Update() with terminatedMsg, no transitionScope)
- `TestCursorBoundsAfterDeleteOnlyRow` — no changes needed

**Tests to REWRITE:**
- `TestAllTransitionsClearSortAndSearch`: iterate over `[]TransitionType{TransitionFull, TransitionDrillDown}` (TransitionPop restores state, doesn't clear sort/search — remove it from this test)
- `TestBreadcrumbNavToRootClearsNavStack` + `TestBreadcrumbNavToDepth1TruncatesNavStack`: Old `transitionBreadcrumb` is gone. Replace with tests that call `m.navigateToBreadcrumb(idx)` directly — but this requires a running environment/fetch cmd, so use a simpler unit test: truncate stack manually then call `prepareStateTransition(TransitionPop)` and verify restore.

**Tests to REMOVE:**
- `TestPrepareStateTransitionExists`: redundant, Story 1.2's `main_transition_test.go` covers this

### Test Patterns — Use `newTestModel(t)`

All tests use `newTestModel(t)` from `internal/app/main_test.go`. Pattern from existing tests:

```go
func TestEnvSwitch_ClearsActiveModal(t *testing.T) {
    m := newTestModel(t)
    m.splashActive = false

    // Pre-condition: active modal
    m.activeModal = ModalConfirmDelete

    // Execute env switch sequence
    m.prepareStateTransition(TransitionFull)  // Story 1.2 already provides this

    if m.activeModal != ModalNone {
        t.Errorf("expected ModalNone after TransitionFull, got %v", m.activeModal)
    }
}

func TestEnvSwitch_SetsNewEnvironment(t *testing.T) {
    m := newTestModel(t)
    m.splashActive = false

    // switchToEnvironment is the callable method from nav.go
    m.switchToEnvironment("staging")

    if m.currentEnv != "staging" {
        t.Errorf("expected currentEnv=staging, got %q", m.currentEnv)
    }
}

func TestContextSwitch_SetsNewRoot(t *testing.T) {
    m := newTestModel(t)
    m.splashActive = false

    rc := "incidents"
    m.currentRoot = rc
    m.prepareStateTransition(TransitionFull)
    m.breadcrumb = []string{rc}
    m.contentHeader = rc
    m.viewMode = rc

    if m.viewMode != "incidents" {
        t.Errorf("expected viewMode=incidents, got %q", m.viewMode)
    }
    if len(m.breadcrumb) != 1 || m.breadcrumb[0] != "incidents" {
        t.Errorf("expected breadcrumb=[incidents], got %v", m.breadcrumb)
    }
}
```

### Critical: `config_quality_test.go` Uses `transitionDrilldown`

File: `internal/app/config_quality_test.go:65`
```go
m.prepareStateTransition(transitionDrilldown)
```

This ALSO fails to compile after Story 1.2 removes `transitionScope`. Check if Story 1.2 fixed this. If not, update this line to `m.prepareStateTransition(TransitionDrillDown)` as part of this story's Task 2.

### Esc Handler — Also Uses Old Constant

`update.go:919`: `m.prepareStateTransition(transitionBack)` → causes compile error. This is Story 1.2's scope (TransitionPop for back navigation). **Do NOT modify this in Story 1.3** — if it still uses `transitionBack`, that means Story 1.2 is incomplete. Fix the env/context switch callsites (Task 1) and note that Story 1.2 still needs to fix `transitionBack`.

### Architecture Constraints

- `prepareStateTransition` is a pure synchronous model method — no `tea.Cmd`, no goroutines
- Called inside `Update()` message handlers before view state changes
- `TransitionFull` guarantees: `activeModal=ModalNone`, `footerError=""`, `searchTerm=""`, `searchMode=false`, `sortColumn=-1`, `sortAscending=true`, `navigationStack=nil`, `genericParams=make(...)`, `selectedDefinitionKey=""`, `selectedInstanceID=""`, `table.SetCursor(0)`, popup cleared
- `switchToEnvironment()` MUST be called BEFORE `TransitionFull` in the env switch sequence so `applyStyle()` fires with the new env (current code ordering is correct)

### Test File Location

- **Update:** `internal/app/state_transition_test.go` (existing file, update in-place)
- **New:** `internal/app/main_env_switch_test.go` (new file, user-facing behavioral tests)

Both files are co-located with the production files they test (in `internal/app/`). Do NOT create a `tests/` subdirectory.

### Project Structure Notes

**Files to modify:**
- `internal/app/update.go` — only if Story 1.2 hasn't already migrated `transitionEnvSwitch`/`transitionContextSwitch` (lines 313, 1005)
- `internal/app/state_transition_test.go` — update old `transitionScope` constants to new `TransitionType`
- `internal/app/config_quality_test.go` — update `transitionDrilldown` → `TransitionDrillDown` (line 65) if Story 1.2 didn't do this

**New files:**
- `internal/app/main_env_switch_test.go` — new behavioral test file

**No changes needed to:**
- `internal/app/transition.go` — already correct
- `internal/app/nav.go` — Story 1.2 handles nav.go callsites
- `internal/operaton/` — never modify
- `o8n-cfg.yaml` / `o8n-env.yaml` — no changes

### Specification Update Obligation

Per architecture definition of done: after implementation, `specification.md` MUST be updated to document:
- The confirmed env switch sequence: `switchToEnvironment()` → `prepareStateTransition(TransitionFull)` → breadcrumb reset → fetch
- The confirmed context switch sequence: popup close → `prepareStateTransition(TransitionFull)` → root/view/breadcrumb update → fetch
- Any edge cases discovered during testing

### Learnings from Story 1.1

- Test file naming: `main_<feature>_test.go` co-located with production file (NOT in `tests/` subdirectory)
- Coverage target: ≥80% on primary production file modified (`transition.go` is covered by Story 1.2's `main_transition_test.go`)
- `gofmt -w .` after every edit pass
- `go vet ./...` before and after each task
- Use `newTestModel(t)` from `main_test.go` as the model factory
- All tasks complete = `make test` passes with zero regressions

### References

- [Source: `_bmad/planning-artifacts/architecture.md#Decision 2: State Transition Contract`] — TransitionType enum, field-clearing specs, mandatory gate requirement
- [Source: `_bmad/planning-artifacts/epics.md#Story 1.3`] — Acceptance criteria (BDD format)
- [Source: `internal/app/transition.go`] — `TransitionType` enum + `prepareStateTransition(t TransitionType)` (already correct)
- [Source: `internal/app/update.go:308-322`] — Env switch Enter handler (uses `transitionEnvSwitch` → change to `TransitionFull`)
- [Source: `internal/app/update.go:990-1025`] — Context switch handler (uses `transitionContextSwitch` → change to `TransitionFull`)
- [Source: `internal/app/nav.go:207-210`] — `switchToEnvironment()` sets `m.currentEnv` + `m.applyStyle()`
- [Source: `internal/app/state_transition_test.go`] — Existing tests using old `transitionScope` constants (must be updated)
- [Source: `internal/app/config_quality_test.go:65`] — Uses `transitionDrilldown` (must be updated)
- [Source: `_bmad/implementation-artifacts/1-1-modal-factory-foundation.md`] — Test structure patterns, `newTestModel(t)` usage
- [Source: `_bmad/implementation-artifacts/1-2-state-transition-contract.md`] — Task 5 (env/context-switch callsite migration), Task 7 (new main_transition_test.go)

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

None.
### Completion Notes List

- Verified env/context switch callsites now use `prepareStateTransition(TransitionFull)` and fixed ordering in context switch path so transition happens before root mutation.
- Audited `nextEnvironment()` usage; function exists but has no active key-handler callsites, so no additional runtime leak path is present.
- Added `internal/app/main_env_switch_test.go` with 9 behavioral tests covering env switch and context switch reset expectations.
- Confirmed transition tests already migrated in `state_transition_test.go` and validated with `go test ./internal/app/... -run TestEnv -v`.
- Updated specification with explicit env/context transition sequences and the Esc-after-switch edge-case note.
- Validation executed: `go test ./internal/app/... -run 'TestEnvSwitch_|TestContextSwitch_' -v`, `make test`, `go vet ./...`, `go build ./...`.

### File List

- `internal/app/update.go` (MODIFIED) — Env switch: reset `currentRoot`/`contentHeader` to process-definitions; context switch: add `popup.offset = 0`
- `internal/app/main_env_switch_test.go` (NEW, then MODIFIED) — Env/context switch behavioral coverage; updated helper and `TestEnvSwitch_BreadcrumbReset` to assert correct post-switch root
- `internal/app/state_transition_test.go` (MODIFIED) — Updated from old `transitionScope` constants to `TransitionType`; rewrote breadcrumb tests
- `internal/app/main_transition_test.go` (NEW) — Full coverage of `TransitionFull`, `TransitionDrillDown`, `TransitionPop` (Story 1.2 test work committed here)
- `internal/app/config_quality_test.go` (MODIFIED) — `transitionDrilldown` → `TransitionDrillDown`
- `specification.md` (MODIFIED) — Documented env/context switch transition sequences and Esc no-op edge case
- `_bmad/implementation-artifacts/1-3-environment-and-context-switching-correctness.md` (MODIFIED) — Status, tasks, Dev Agent Record, File List, Change Log
- `_bmad/implementation-artifacts/sprint-status.yaml` (MODIFIED) — Story status sync

### Senior Developer Review (AI)

**Reviewer:** claude-sonnet-4-6 | **Date:** 2026-03-04

**Findings Fixed (3):**
- **H1 FIXED** — Env switch now resets `m.currentRoot` and `m.contentHeader` to `dao.ResourceProcessDefinitions` before building breadcrumb. `TestEnvSwitch_BreadcrumbReset` updated to assert correct post-switch root. [`update.go:314-315`]
- **M1 FIXED** — Added `m.popup.offset = 0` to context switch popup-close path, consistent with all 5 other popup-close sites. [`update.go:976`]
- **M2+L1 FIXED** — File List completed (3 missing files added); Agent Model Used corrected from fictitious "GPT-5.3-Codex" to `claude-sonnet-4-6`.

**Remaining LOW items (accepted, not blocking):**
- L2: Tests use helper functions rather than exercising the real `Update()` handler — acceptable for unit-level coverage; integration coverage is added for the H1 fix
- L3: `nextEnvironment()` dead code retained — not a runtime issue; separate cleanup story if desired

**Follow-up fixes applied (CR round):**
- Added key-handler integration coverage in `internal/app/main_env_switch_test.go`:
  - `TestEnvSwitch_KeyHandler_UsesNewEnvironmentURLAndClearsState`
  - `TestContextSwitch_KeyHandler_ClearsStateAndSetsRoot`
- This closes the remaining medium concerns around URL activation assertion and handler-path validation.

### Change Log

- 2026-03-04: Implemented Story 1.3 env/context switch correctness, added behavioral tests, validated full suite, and updated spec.
- 2026-03-04: Code review (claude-sonnet-4-6) — fixed H1 (stale currentRoot on env switch), M1 (missing popup.offset reset), M2+L1 (incomplete file list, wrong model).
- 2026-03-04: CR follow-up — added env/context key-handler integration tests and set story back to review for final verification.
