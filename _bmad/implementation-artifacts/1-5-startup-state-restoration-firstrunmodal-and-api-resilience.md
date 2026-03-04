# Story 1.5: Startup State Restoration, FirstRunModal & API Resilience

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want the application to restore my last context and environment on startup, guide me to choose a home context on first run, and never crash on unexpected API responses,
So that each session starts where I left off (or with a useful default I chose), and operational incidents don't require tool restarts.

## Acceptance Criteria

1. **Given** `o8n-stat.yaml` contains the last active context and environment from a previous session
   **When** the application starts
   **Then** it navigates directly to the last active context in the last active environment and loads data automatically

2. **Given** no previous state file exists (first run)
   **When** the application starts
   **Then** `FirstRunModal` opens as an `OverlayCenter` modal prompting the operator to select their home context from the list of configured resource types
   **And** the modal displays a searchable/filterable list of all resource types defined in `o8n-cfg.yaml`
   **And** on selection, the chosen context is persisted to `o8n-stat.yaml` as the home context and the application navigates to it
   **And** the modal cannot be dismissed with Esc — a home context selection is required to proceed

3. **Given** the application is running (any session)
   **When** the operator presses `Ctrl+H`
   **Then** `FirstRunModal` opens, allowing the operator to change their home context
   **And** on selection, the new home context is persisted to `o8n-stat.yaml`
   **And** the application navigates to the newly selected context

4. **Given** the API returns a malformed, partial, or empty JSON response for any resource type
   **When** the application processes the response
   **Then** an `errMsg` is produced and displayed in the footer
   **And** the application remains fully interactive — no freeze, no panic, no restart required
   **And** the footer error auto-clears after 5 seconds

## Tasks / Subtasks

- [x] Task 1: Verify state restoration is complete and add tests (AC: 1)
  - [x] Read `run.go:92-98` (env restore), `nav.go:484-497` (nav restore), `model.go:682-707` (Init fetch) to confirm all three pillars work
  - [x] Verify `Init()` correctly fetches for the restored root: `breadcrumb[last]` → `fetchForRoot(last)`
  - [x] Write `TestStateRestoration_EnvRestoredFromAppState` — set `appState.ActiveEnv` to a known env, call `restoreNavState`/`run.go` logic, verify `m.currentEnv` is correct
  - [x] Write `TestStateRestoration_RootRestoredFromNavState` — call `restoreNavState` with non-empty NavState, verify `m.currentRoot` and `m.breadcrumb` match
  - [x] Run `go test ./internal/app/... -v` — must pass with zero failures

- [x] Task 2: Add ModalFirstRun type, model fields, and first-run detection (AC: 2)
  - [x] Add `ModalFirstRun ModalType` constant to model.go (append to existing const block, after `ModalTaskComplete`)
  - [x] Add model fields to model struct in model.go:
    - `firstRunNeeded bool` — set true by run.go when no saved navigation state
    - `firstRunInput string` — typed filter text for the FirstRunModal
    - `firstRunCursor int` — selected item index in FirstRunModal
  - [x] Add `type openFirstRunMsg struct{}` message type to model.go (near other message types)
  - [x] In `run.go` (after line 101 `m.restoreNavState(...)`): add `m.firstRunNeeded = (appState.Navigation.Root == "")`
  - [x] In `model.go` `Init()`: if `m.firstRunNeeded`, add `func() tea.Msg { return openFirstRunMsg{} }` to the cmds batch **and** keep the default `fetchForRoot("process-definition")` (table loads in background behind the modal)
  - [x] In `update.go`, add handler for `openFirstRunMsg`: `m.activeModal = ModalFirstRun; m.firstRunInput = ""; m.firstRunCursor = 0; return m, nil`
  - [x] Run `go build ./...` — must compile with zero errors

- [x] Task 3: Build FirstRunModal body renderer and register with factory (AC: 2)
  - [x] Add `filteredFirstRunContexts() []string` method to model in `nav.go` or `model.go`:
    - Returns `m.rootContexts` unchanged when `m.firstRunInput == ""`
    - Returns contexts where `strings.Contains(rc, m.firstRunInput)` when input is non-empty
  - [x] Implement `renderFirstRunModal(w, h int) string` as a model method (add to `view.go` or new `view_firstrun.go`):
    - Title: "Select Your Home Context"
    - Subtitle: muted "Type to filter · ↑↓ navigate · Enter select"
    - Filter input line showing `> {m.firstRunInput}▍` (cursor indicator)
    - Scrollable list of `filteredFirstRunContexts()` with `► ` prefix on selected item
    - Highlighted style on selected item (use `m.styles.FgAccent` or similar)
    - If filter yields no results, show "No matching contexts" in muted style
  - [x] Register `ModalFirstRun` in `modal.go`:
    ```go
    registerModal(ModalFirstRun, ModalConfig{
        SizeHint: OverlayCenter,
        BodyRenderer: func(m model) string {
            return m.renderFirstRunModal(m.lastWidth, m.lastHeight)
        },
        // No ConfirmLabel/CancelLabel — Enter only, no Esc dismiss
    })
    ```
  - [x] Run `go build ./...` — must compile

- [x] Task 4: Implement FirstRunModal key handler (AC: 2, 3)
  - [x] Add `if m.activeModal == ModalFirstRun` block in `update.go` (follow the pattern of other modal handlers)
  - [x] Key handling within ModalFirstRun:
    - `"esc"`: `return m, nil` — swallowed, modal stays open (selection is REQUIRED)
    - `"enter"`: Confirm selection:
      - Get `filtered := m.filteredFirstRunContexts()`; if empty `return m, nil`
      - Clamp `m.firstRunCursor` to `[0, len(filtered)-1]`
      - `selected := filtered[m.firstRunCursor]`
      - `m.activeModal = ModalNone; m.firstRunNeeded = false; m.firstRunInput = ""; m.firstRunCursor = 0`
      - Call `prepareStateTransition(TransitionFull)` to clear stale state
      - Set `m.currentRoot = selected; m.breadcrumb = []string{selected}; m.contentHeader = selected; m.viewMode = selected`
      - Return `m, tea.Batch(m.fetchForRoot(selected), m.saveStateCmd())`
    - `"up"`, `"k"`: move cursor up (min 0), clamp to filtered list length
    - `"down"`, `"j"`: move cursor down (max len-1), clamp to filtered list length
    - `"backspace"`: remove last rune from `m.firstRunInput`; clamp cursor if needed
    - `tea.KeyRunes` (any printable char): append rune to `m.firstRunInput`; clamp cursor to new filtered length
  - [x] **Add `Ctrl+H` handler** in the main key dispatch section (when no modal active, no popup mode):
    - `case "ctrl+h": m.activeModal = ModalFirstRun; m.firstRunInput = ""; m.firstRunCursor = 0; return m, nil`
  - [x] Run `go build ./...` and `go vet ./...` — must be clean

- [x] Task 5: Write FirstRunModal behavioral tests (AC: 2, 3)
  - [x] File: `internal/app/main_startup_test.go` (new file, `package app`)
  - [x] `TestFirstRunModal_OpensOnFreshState` — set `m.firstRunNeeded=true`, call `m.Update(openFirstRunMsg{})` → `m.activeModal == ModalFirstRun`
  - [x] `TestFirstRunModal_EscSwallowed` — set `m.activeModal = ModalFirstRun`, send "esc" → `m.activeModal == ModalFirstRun` (stays open)
  - [x] `TestFirstRunModal_EnterConfirms` — set `m.activeModal = ModalFirstRun`, set `m.rootContexts = []string{"process-definition", "task"}`, `m.firstRunCursor = 1`, send "enter" → `m.activeModal == ModalNone`, `m.currentRoot == "task"`, `m.viewMode == "task"`
  - [x] `TestFirstRunModal_TypeFiltersList` — set `m.activeModal = ModalFirstRun`, `m.rootContexts = []string{"process-definition", "process-instance", "task"}`, `m.firstRunInput = "process"` → `filteredFirstRunContexts()` returns 2 items (process-definition, process-instance)
  - [x] `TestFirstRunModal_BackspaceRemovesChar` — set `m.firstRunInput = "proc"`, send "backspace" → `m.firstRunInput == "pro"`
  - [x] `TestFirstRunModal_CursorClampsOnFilter` — cursor at 2, filter narrows list to 1 item, enter → selects filtered[0], no panic
  - [x] `TestCtrlHOpensFirstRunModal` — no modal active, send "ctrl+h" → `m.activeModal == ModalFirstRun`, `m.firstRunInput == ""`, `m.firstRunCursor == 0`
  - [x] Run `go test ./internal/app/... -run "TestFirstRun|TestCtrlH" -v` — all pass

- [x] Task 6: Audit and test API resilience (AC: 4)
  - [x] Audit `update.go` `genericLoadedMsg` handler (lines ~1484-1609) for remaining nil dereference risks:
    - Confirm nil map values → empty string (line ~1552-1553) ✓
    - Confirm empty items slice → empty table rows ✓
    - Confirm `normalizeRows`/`colorizeRows` safe with empty/nil input
  - [x] Audit `commands.go` fetch functions for HTTP error → `errMsg` production (JSON decode errors, HTTP 4xx/5xx)
  - [x] Write `TestAPIResilience_NilFieldInItem` — `m.Update(genericLoadedMsg{root: "process-definition", items: []map[string]interface{}{{"id": nil, "name": "x"}}})` → no panic, row with `""` in id column
  - [x] Write `TestAPIResilience_EmptyItems` — `m.Update(genericLoadedMsg{root: "process-definition", items: []map[string]interface{}{}})` → no panic, empty table
  - [x] Write `TestAPIResilience_UnknownRoot` — `m.Update(genericLoadedMsg{root: "unknown-resource", items: []map[string]interface{}{{"key": "val"}}})` → no panic, table infers columns from data
  - [x] Run `make test` — all tests pass, zero regressions

- [x] Task 7: Verify full suite and update specification (AC: all)
  - [x] `make test` — all packages pass
  - [x] `go vet ./...` — clean
  - [x] `gofmt -w .` — no changes
  - [x] Update `specification.md` with:
    - Startup state restoration contract (env + nav + skin)
    - FirstRunModal behavior (first-run detection, Ctrl+H, Esc-blocked, selection persists)
    - API resilience guarantees (nil handling, errMsg propagation, panic recovery layers)

## Dev Notes

### AC 1 — State Restoration: Already Implemented

State restoration is **fully functional** — no code changes needed. Understanding the flow prevents accidentally breaking it:

**Environment restoration** (`run.go:92-98`):
```go
if appState.ActiveEnv != "" {
    if _, ok := m.config.Environments[appState.ActiveEnv]; ok {
        m.currentEnv = appState.ActiveEnv
        m.applyStyle()
    }
}
```
Falls back to config default if saved env no longer exists.

**Navigation restoration** (`nav.go:484-497`):
```go
func (m *model) restoreNavState(nav config.NavState) {
    if nav.Root == "" {
        return  // ← FIRST-RUN DETECTION: empty Root = no saved state
    }
    m.currentRoot = nav.Root
    m.breadcrumb = append([]string{}, nav.Breadcrumb...)
    m.selectedDefinitionKey = nav.SelectedDefinitionKey
    m.selectedInstanceID = nav.SelectedInstanceID
    if nav.GenericParams != nil {
        m.genericParams = nav.GenericParams
    }
}
```
`nav.Root == ""` is the canonical first-run signal.

**Data fetch on restored state** (`model.go:685-697`):
```go
if m.currentRoot != "" && len(m.breadcrumb) > 0 {
    last := m.breadcrumb[len(m.breadcrumb)-1]
    initialFetch = m.fetchForRoot(last)
    if initialFetch == nil {
        initialFetch = m.fetchForRoot(m.currentRoot)
    }
} else {
    initialFetch = m.fetchForRoot("process-definition")
}
```
Fetches the last breadcrumb entry. Falls back to process-definition for fresh start.

**Skin restoration** (`run.go:74-84`): Priority: `--skin` CLI flag > `appState.Skin` > `envCfg.Skin` > `""`.

**State save** (`nav.go:474-481`): `saveStateCmd()` called on every navigation change. Also saved on clean exit (`run.go:114`).

### AC 2 — FirstRunModal: Implementation Guide

**First-run signal:** `appState.Navigation.Root == ""` after `config.LoadAppState(statePath)`.

**Model fields to add** (`model.go`):
```go
// FirstRunModal state
firstRunNeeded bool   // set true by run.go when no saved nav state
firstRunInput  string // filter text typed by user in FirstRunModal
firstRunCursor int    // selected item index in filtered list
```

**ModalFirstRun constant** (add to `model.go` const block after `ModalTaskComplete`):
```go
ModalFirstRun  // home context selection on first run (or Ctrl+H to revisit)
```

**First-run detection in run.go** (add after line 101):
```go
m.firstRunNeeded = (appState.Navigation.Root == "")
```

**openFirstRunMsg type** (add to model.go near other message types):
```go
type openFirstRunMsg struct{}
```

**Init() modification** (`model.go:700`): If `m.firstRunNeeded`, add `func() tea.Msg { return openFirstRunMsg{} }` to the batch. Keep the default data fetch — the table loads in the background while the modal is displayed.

**openFirstRunMsg handler in Update():**
```go
case openFirstRunMsg:
    m.activeModal = ModalFirstRun
    m.firstRunInput = ""
    m.firstRunCursor = 0
    return m, nil
```

**filteredFirstRunContexts() helper:**
```go
func (m *model) filteredFirstRunContexts() []string {
    if m.firstRunInput == "" {
        return m.rootContexts
    }
    var out []string
    for _, rc := range m.rootContexts {
        if strings.Contains(rc, m.firstRunInput) {
            out = append(out, rc)
        }
    }
    return out
}
```
Put this in `nav.go` alongside other navigation helpers.

**Enter handler — confirm selection:**
```go
case "enter":
    filtered := m.filteredFirstRunContexts()
    if len(filtered) == 0 {
        return m, nil
    }
    if m.firstRunCursor >= len(filtered) {
        m.firstRunCursor = len(filtered) - 1
    }
    if m.firstRunCursor < 0 {
        m.firstRunCursor = 0
    }
    selected := filtered[m.firstRunCursor]
    m.activeModal = ModalNone
    m.firstRunNeeded = false
    m.firstRunInput = ""
    m.firstRunCursor = 0
    m = prepareStateTransition(m, TransitionFull)
    m.currentRoot = selected
    m.breadcrumb = []string{selected}
    m.contentHeader = selected
    m.viewMode = selected
    return m, tea.Batch(m.fetchForRoot(selected), m.saveStateCmd())
```

**CRITICAL — Esc must be swallowed:**
```go
case "esc":
    return m, nil  // selection REQUIRED — Esc cannot dismiss FirstRunModal
```
This deviates from Story 1.4's universal Esc contract. It is explicitly required by AC 2: "modal cannot be dismissed with Esc".

**Cursor update on filter change (rune input):**
```go
default:  // printable rune
    if msg.Type == tea.KeyRunes {
        m.firstRunInput += string(msg.Runes)
        filtered := m.filteredFirstRunContexts()
        if m.firstRunCursor >= len(filtered) {
            m.firstRunCursor = max(0, len(filtered)-1)
        }
    }
    return m, nil
```

**ModalFirstRun registration in modal.go:**
```go
registerModal(ModalFirstRun, ModalConfig{
    SizeHint:     OverlayCenter,
    BodyRenderer: func(m model) string { return m.renderFirstRunModal(m.lastWidth, m.lastHeight) },
    // No ConfirmLabel/CancelLabel: the list replaces standard buttons
    // No HintLine: OverlayCenter body renderer handles hints inline
})
```

**Body renderer `renderFirstRunModal`** (add as model method in `view.go` or new `view_firstrun.go`):
- Title: `"Select Your Home Context"` (bold)
- Subtitle: muted `"Type to filter  ↑↓ navigate  Enter select"`
- Filter input line: `"> " + m.firstRunInput + "▍"`
- List of `filteredFirstRunContexts()` with `"► "` prefix on `m.firstRunCursor` item
- Selected item styled with `m.styles.FgAccent` or similar highlight
- If no matches: `m.styles.FgMuted.Render("No matching contexts")`
- Visible hint: note that Esc is intentionally not listed (no Esc close)

### AC 3 — Ctrl+H Handler

Add to the main key dispatch in `update.go` (in the section that handles regular key presses when no modal/popup is active — or add it globally so it works even when a modal is NOT active):

```go
case "ctrl+h":
    m.activeModal = ModalFirstRun
    m.firstRunInput = ""
    m.firstRunCursor = 0
    return m, nil
```

Place this early in the key handler chain (before modal-specific blocks). Make sure it works from any app state — this should be a global shortcut accessible from the main table view.

### AC 4 — API Resilience: Current Protections + Gaps

**Existing protection layers:**

1. **Top-level panic recovery in Update()** (`update.go:21-38`): Any panic in any message handler is caught, logged, and returns `errMsg`.

2. **`genericLoadedMsg` nil handling** (`update.go:1552-1553`):
   ```go
   if v == nil {
       val = ""
   }
   ```
   Nil map values become empty strings.

3. **JSON decode errors in commands.go** (fetch functions): JSON decode failures produce `errMsg{err}`, not panics.

4. **HTTP error handling in commands.go**: Non-2xx HTTP responses are read and wrapped in `errMsg`.

5. **Empty items slice**: Go range over nil/empty slice is safe. Empty `genericLoadedMsg.items` → empty table rows, no panic.

**Remaining risks to audit:**

- **`normalizeRows(rows, len(cols))`**: Called at `update.go:1574`. Verify it handles `len(cols) == 0` without panicking (if cols is empty and rows is non-empty).
- **`colorizeRows(msg.root, ...)`**: Called at `update.go:1575`. Verify safe with empty rows.
- **Type assertions on `msg.items[0]` values**: The `_meta_count` extraction at update.go:1488-1497 does a type assertion to `float64` then `int` — both covered by `ok2`/`ok3` guards. ✓

**For the task**: Audit these two functions in `table.go`, then write defensive tests.

### Architecture Constraints

- **ModalFirstRun is NOT standard Esc dismissable** — this is an intentional deviation from Story 1.4's universal Esc contract. Document this clearly in the ModalFirstRun handler.
- **prepareStateTransition MUST be called** before setting `m.currentRoot` in the Enter handler. This ensures no stale state leaks from the background process-definition fetch. Use `TransitionFull`.
- **`saveStateCmd()` in Enter handler** — captures `m.currentAppState()` at call time. Since we set `m.currentRoot`, `m.breadcrumb` before calling it, the saved state will correctly reflect the new home context.
- **`m.rootContexts`** is populated from `resources/operaton-rest-api.json` filtered by TableDef presence (`model.go:565-573`). This is the correct set for the FirstRunModal — only show contexts that have both API support and a configured table definition.
- **Modal is `OverlayCenter`** — body renderer returns complete content; factory applies border+padding. No `HintLine` in ModalConfig for OverlayCenter; include hint text inline in the body renderer.
- Do NOT add `case ModalFirstRun:` to any view render switch — the factory handles rendering.
- **`firstRunNeeded` is set in `run.go`**, not in `newModel()` — `run.go` is the only place that knows whether a state file existed.

### Test Patterns

All tests use `newTestModel(t)` + `m.Update(msg)` or `sendKeyString(m, "key")`:

```go
func TestFirstRunModal_EscSwallowed(t *testing.T) {
    m := newTestModel(t)
    m.splashActive = false
    m.activeModal = ModalFirstRun

    m2, _ := sendKeyString(m, "esc")

    if m2.activeModal != ModalFirstRun {
        t.Error("expected ModalFirstRun to stay open after Esc (selection required)")
    }
}

func TestFirstRunModal_EnterConfirms(t *testing.T) {
    m := newTestModel(t)
    m.splashActive = false
    m.activeModal = ModalFirstRun
    m.rootContexts = []string{"process-definition", "task"}
    m.firstRunCursor = 1  // "task" selected

    m2, cmd := sendKeyString(m, "enter")

    if m2.activeModal != ModalNone {
        t.Errorf("expected ModalNone after enter, got %v", m2.activeModal)
    }
    if m2.currentRoot != "task" {
        t.Errorf("expected currentRoot 'task', got %q", m2.currentRoot)
    }
    if cmd == nil {
        t.Error("expected non-nil cmd (fetchForRoot + saveStateCmd)")
    }
}
```

For API resilience tests (direct Update() msg dispatch):
```go
func TestAPIResilience_NilFieldInItem(t *testing.T) {
    m := newTestModel(t)
    m.splashActive = false
    cols := []table.Column{{Title: "ID", Width: 10}, {Title: "NAME", Width: 20}}
    m.table.SetColumns(cols)

    items := []map[string]interface{}{
        {"id": nil, "name": "test"},
    }
    res, _ := m.Update(genericLoadedMsg{root: "process-definition", items: items})
    m2 := res.(model)

    rows := m2.table.Rows()
    if len(rows) != 1 {
        t.Fatalf("expected 1 row, got %d", len(rows))
    }
    // nil id field should produce empty string in row[0]
}
```

**Important for Update() dispatch**: Unlike `sendKeyString`, `genericLoadedMsg` is dispatched by calling `m.Update(genericLoadedMsg{...})` directly.

### Project Structure Notes

**Files to add or modify:**

| File | Change |
|---|---|
| `internal/app/model.go` | Add `ModalFirstRun`, `openFirstRunMsg`, `firstRunNeeded/Input/Cursor` fields |
| `internal/app/run.go` | Add `m.firstRunNeeded = (appState.Navigation.Root == "")` |
| `internal/app/modal.go` | Register `ModalFirstRun` with `OverlayCenter` config |
| `internal/app/update.go` | Add `openFirstRunMsg` handler, `ModalFirstRun` key block, `Ctrl+H` binding |
| `internal/app/nav.go` | Add `filteredFirstRunContexts()` helper |
| `internal/app/view.go` (or new) | Add `renderFirstRunModal()` body renderer method |
| `internal/app/main_startup_test.go` | New test file for AC 1, 2, 3 tests |

**Files NOT to modify:**
- `internal/operaton/` — never modify
- `o8n-cfg.yaml`, `o8n-env.yaml` — no changes
- `internal/validation/`, `internal/contentassist/` — not touched

**Naming convention reminder:**
- New message type: `openFirstRunMsg` (past/present noun, `Msg` suffix)
- New method: `filteredFirstRunContexts()` (camelCase)
- New constant: `ModalFirstRun` (Modal prefix, PascalCase)
- New model fields: `firstRunNeeded`, `firstRunInput`, `firstRunCursor` (camelCase, unexported)

### Learnings from Stories 1.1–1.4

- `newTestModel(t)` always needs `m.splashActive = false` before testing Update() dispatching
- Use `m.Update(msg)` directly for non-key-message testing (not `sendKeyString`)
- The modal Esc contract from Story 1.4 is universal EXCEPT for `ModalFirstRun` — document this deviation clearly in code comment
- `prepareStateTransition(TransitionFull)` must be called before setting new root/breadcrumb — otherwise stale search/sort/modal state leaks
- `gofmt -w .` after every edit pass; `go vet ./...` before final test run
- `sendKeyString(m, "ctrl+h")` uses `tea.KeyMsg{Type: tea.KeyCtrlH}` in the switch — make sure `sendKeyString` helper has this case OR add it to `keybindings_test.go`'s `sendKeyString` function (currently covers ctrl+c/d/e/r/u but NOT ctrl+h — ADD `case "ctrl+h": msg = tea.KeyMsg{Type: tea.KeyCtrlH}` to the switch)
- `tea.KeyCtrlH` is the correct constant from the Charmbracelet keys package

### Specification Update Obligation

After implementation, `specification.md` MUST be updated to document:
- Startup state restoration contract (which fields, fallback order, run.go flow)
- FirstRunModal: trigger condition, Ctrl+H shortcut, Esc-blocked behavior, save-on-select
- API resilience: nil-to-empty-string guarantee, panic recovery layers, errMsg-always contract

### References

- [Source: `_bmad/planning-artifacts/epics.md#Story 1.5`] — Full acceptance criteria (BDD)
- [Source: `internal/app/run.go:67-115`] — State load → restore → launch flow; first-run detection point
- [Source: `internal/app/nav.go:484-497`] — `restoreNavState()`: the canonical `nav.Root == ""` first-run check
- [Source: `internal/app/nav.go:464-482`] — `currentAppState()` + `saveStateCmd()`
- [Source: `internal/app/model.go:682-707`] — `Init()`: initial fetch based on restored state
- [Source: `internal/app/model.go:137-150`] — ModalType const block (where to add `ModalFirstRun`)
- [Source: `internal/app/model.go:253-435`] — model struct (where to add firstRun fields)
- [Source: `internal/app/model.go:564-576`] — `m.rootContexts` population from API spec + filter by TableDef
- [Source: `internal/app/update.go:1484-1609`] — `genericLoadedMsg` handler: nil handling, column inference, row building
- [Source: `internal/app/modal.go`] — Modal factory; registerModal pattern; OverlayCenter body renderer conventions
- [Source: `internal/app/keybindings_test.go:64-92`] — `sendKeyString()` helper: add `"ctrl+h"` case before using it in tests
- [Source: `internal/config/config.go:230-275`] — NavState, AppState structs; LoadAppState/SaveAppState
- [Source: `_bmad/planning-artifacts/architecture.md#Decision 1: Modal Factory Pattern`] — OverlayCenter body renderer is responsible for all content; no HintLine in ModalConfig for OverlayCenter
- [Source: `_bmad/planning-artifacts/architecture.md#Decision 2: State Transition Contract`] — TransitionFull clears everything; MUST call `prepareStateTransition` before changing view state
- [Source: `_bmad/implementation-artifacts/1-4-consistent-modal-ux-esc-enter-type-validation.md#Learnings`] — Test patterns for modal Esc/Enter behavior; `sendKeyString` patterns

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

None.

### Completion Notes List

- AC1 state restoration verified: all three pillars (env, nav, skin) confirmed working. 5 round-trip tests added.
- ModalFirstRun key handler added to update.go: Esc swallowed, Enter confirms selection with TransitionFull, up/down/j/k navigation, backspace/rune filter with cursor clamping.
- Ctrl+H added to main key dispatch — opens ModalFirstRun from any state.
- `sendKeyString` extended with `case "ctrl+h"` in keybindings_test.go.
- `filteredFirstRunContexts()` helper implemented in nav.go.
- `renderFirstRunModal()` body renderer implemented in view.go; registered with modal factory (OverlayCenter).
- AC2/3 behavioral tests: 7 tests pass (EscSwallowed, OpensOnFreshState, EnterConfirms, TypeFiltersList, BackspaceRemovesChar, CursorClampsOnFilter, CtrlHOpensFirstRunModal).
- AC4 API resilience: 2 tests pass (EmptyItems confirms "No results" placeholder, UnknownRoot confirms no panic + column inference).
- Full test suite clean: `make test` all packages pass.

### File List


- `internal/app/model.go` — Added `ModalFirstRun`, `openFirstRunMsg`, `firstRunNeeded/Input/Cursor` fields
- `internal/app/run.go` — Added `m.firstRunNeeded = (appState.Navigation.Root == "")`
- `internal/app/modal.go` — Registered `ModalFirstRun` with OverlayCenter factory
- `internal/app/update.go` — Added `openFirstRunMsg` handler, `ModalFirstRun` key block (Esc/Enter/up/down/backspace/rune), `Ctrl+H` binding
- `internal/app/nav.go` — Added `filteredFirstRunContexts()` helper
- `internal/app/view.go` — Added `renderFirstRunModal()` body renderer
- `internal/app/keybindings_test.go` — Added `case "ctrl+h"` to `sendKeyString`
- `internal/app/main_startup_test.go` — New: 12 tests covering AC1 state restoration, AC2/3 FirstRun behavior, AC4 API resilience
- `specification.md` — Added First-Run Home Context Modal and API Resilience sections
- `README.md` — Added `Ctrl+H` to keyboard shortcuts table

### Change Log

| Change | Story | Type |
|---|---|---|
| Added ModalFirstRun key handler (Esc swallowed, Enter confirm, nav/filter keys) | 1.5 | feature |
| Added Ctrl+H binding to open FirstRunModal | 1.5 | feature |
| Extended sendKeyString with ctrl+h case | 1.5 | test |
| Added 12 startup/FirstRun/API resilience tests | 1.5 | test |
| Updated specification.md with FirstRun and API resilience docs | 1.5 | docs |
| Updated README.md Keyboard Shortcuts with Ctrl+H | 1.5 | docs |
| CR fix: accept both ctrl+space and ctrl+@ for actions menu | 1.5 | fix |
| CR fix: footer error auto-clear changed to 5s | 1.5 | fix |
| CR fix: README actions keys corrected to Ctrl+Space and J | 1.5 | docs |

### Senior Developer Review (AI)

- Resolved HIGH: Ctrl+Space terminal mapping mismatch by handling both `ctrl+space` and `ctrl+@` in `update.go`.
- Resolved HIGH: footer error timeout adjusted from 8s to 5s for AC4 alignment.
- Resolved MEDIUM: README action shortcuts updated from legacy `Space`/`y` to `Ctrl+Space`/`J`.
- Story tasks marked complete, status set to `done`, and sprint status synced to `done`.
