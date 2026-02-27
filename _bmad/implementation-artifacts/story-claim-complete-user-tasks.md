# Story: Claim and Complete User Tasks

## Summary

Enable users to claim user tasks, fill in output variables via a structured completion dialog, and unclaim tasks they no longer want to work on. Pressing `c` claims an unassigned task; `Enter` on an owned task opens a full-screen completion dialog showing read-only input variables and editable output (form) variable fields; `u` releases a claimed task back to the pool.

## Motivation

The task table currently supports completing tasks with a single keypress that sends an empty body — no variables, no context. Real Operaton user tasks have form variables (output) and process variables (input context). Without a proper completion dialog, users must use the REST API or Operaton Cockpit to actually fill in task data. This feature closes that gap and makes o8n a first-class tool for human task work.

## Acceptance Criteria

### Claim (`c`)

- [x] **AC-1:** Pressing `c` on a task with an empty `assignee` column calls `POST /task/{id}/claim` with body `{"userId": "<env.username>"}`
- [x] **AC-2:** On success (HTTP 204) the task row refreshes and `assignee` shows `env.username`; footer shows `✓ Claimed: <task name>`
- [x] **AC-3:** Pressing `c` when `assignee` is a different user shows footer error `Already claimed by <assignee>` — no API call
- [x] **AC-4:** Pressing `c` when `assignee` is already the current user shows footer hint `You already own this task — press Enter to complete` — no API call
- [x] **AC-5:** The claim body uses `{currentUser}` as a new body template placeholder resolved from `m.config.Environments[m.currentEnv].Username`

### Open Completion Dialog (`Enter` on task table)

- [x] **AC-6:** Pressing `Enter` on a task where `assignee == env.Username` fetches `GET /task/{id}/variables` and `GET /task/{id}/form-variables` in parallel and opens `ModalTaskComplete`
- [x] **AC-7:** Footer shows `Loading task variables…` (with spinner) while fetches are in-flight
- [x] **AC-8:** If either fetch fails, the dialog does NOT open and the error is shown in the footer
- [x] **AC-9:** Pressing `Enter` when `assignee` is empty shows footer error `Claim this task first (c)` — no fetch
- [x] **AC-10:** Pressing `Enter` when `assignee` is a different user shows footer error `Task is assigned to <assignee>` — no fetch
- [x] **AC-11:** Pressing `Enter` on a non-task table retains existing drilldown behaviour unchanged

### Completion Dialog Layout

- [x] **AC-12:** The dialog is a full-screen overlay with a rounded border and title `Complete Task`
- [x] **AC-13:** A task name subtitle appears directly below the title
- [x] **AC-14:** The `INPUT VARIABLES` section lists all variables from `GET /task/{id}/variables` as `name : value` rows — read-only, muted style, sorted alphabetically
- [x] **AC-15:** If no input variables exist, the section shows `No context variables`
- [x] **AC-16:** The `OUTPUT VARIABLES` section lists all fields from `GET /task/{id}/form-variables` as editable input rows — `name │ <input>` — sorted alphabetically, no type label displayed
- [x] **AC-17:** If no form variables exist, the section shows `No output variables defined for this task`
- [x] **AC-18:** Output fields are pre-populated: if `formVarName` matches a key in `taskInputVars`, the input is pre-filled with that variable's value
- [x] **AC-19:** Boolean fields toggle between `true`/`false` on `Space`; no other type-specific UI treatment is visible
- [x] **AC-20:** Invalid values show an inline `⚠ <message>` below the affected field; the `[Complete]` button is rendered disabled (`BtnDisabled` style) until all errors are resolved
- [x] **AC-21:** The hint line at the bottom of the dialog reads: `Tab: next  Space: toggle bool  Esc: back`

### Completion Dialog Navigation

- [x] **AC-22:** `Tab` advances focus: field 0 → … → field N-1 → `[Complete]` → `[Back]` → field 0
- [x] **AC-23:** `Shift+Tab` reverses the cycle
- [x] **AC-24:** `Enter` on an output field advances focus to `[Complete]` without submitting
- [x] **AC-25:** `Enter` on `[Complete]` submits (when enabled); `Enter` on `[Back]` closes without submitting

### Submit / Complete

- [x] **AC-26:** `[Complete]` calls `POST /task/{id}/complete` with `CompleteTaskDto{ Variables: map[name]→VariableValueDto{Value: parsedValue, Type: originalType} }`
- [x] **AC-27:** `Type` in the submitted payload uses the original casing from the API response (e.g. `"String"`, `"Boolean"`, `"Integer"`)
- [x] **AC-28:** On HTTP 204: dialog closes, task list refreshes (row disappears or updates), footer shows `✓ Completed: <task name>`
- [x] **AC-29:** On API error: dialog stays open, footer shows error, field values are preserved
- [x] **AC-30:** When form variables are empty, `[Complete]` sends `{"variables": {}}` and behaves identically on success

### Back / Escape

- [x] **AC-31:** `Esc` from any focus area closes the dialog without any API call
- [x] **AC-32:** `[Back]` + `Enter` is identical to `Esc`
- [x] **AC-33:** Re-opening the dialog after back/escape re-fetches fresh data (no state carried over)
- [x] **AC-34:** All `taskComplete*` model state fields are cleared on dialog close

### Unclaim (`u`)

- [x] **AC-35:** Pressing `u` when `assignee == env.Username` calls `POST /task/{id}/unclaim` (no body)
- [x] **AC-36:** On success (HTTP 204) the `assignee` column clears; footer shows `✓ Unclaimed: <task name>`
- [x] **AC-37:** Pressing `u` when `assignee` is empty shows footer error `Task is not claimed` — no API call
- [x] **AC-38:** Pressing `u` when `assignee` is a different user shows footer error `Task is assigned to <assignee>, not you` — no API call

## Tasks

### Task 1: Config and body template changes

**Files:** `o8n-cfg.yaml`, `internal/app/commands.go` (or wherever body template resolution lives)

1. In `o8n-cfg.yaml` task table: rename claim action key from `k` → `c`; update body to `{"userId": "{currentUser}"}`
2. Remove the `c` (Complete Task, `POST /task/{id}/complete body:'{}'`) action — it is replaced by the dialog flow
3. Add `{currentUser}` as a supported placeholder in body template resolution, resolved from `m.config.Environments[m.currentEnv].Username`

### Task 2: Model state extensions

**Files:** `internal/app/model.go`

1. Add `ModalTaskComplete` to the modal type enum
2. Add `taskCompleteField` struct:
   ```go
   type taskCompleteField struct {
       name    string
       varType string          // lowercased: "string", "boolean", "integer", "double", "json"
       origType string         // original casing from API for submission
       input   textinput.Model
       error   string
   }
   ```
3. Add `taskCompleteFocusArea` enum: `focusTaskField`, `focusTaskComplete`, `focusTaskBack`
4. Add fields to `model` struct:
   - `taskCompleteTaskID   string`
   - `taskCompleteTaskName string`
   - `taskInputVars        map[string]operaton.VariableValueDto`
   - `taskCompleteFields   []taskCompleteField`
   - `taskCompletePos      int`
   - `taskCompleteFocus    taskCompleteFocusArea`

### Task 3: Command implementations

**Files:** `internal/app/commands.go`

1. `fetchTaskVariablesCmd(taskID string) tea.Cmd` — calls `GET /task/{id}/variables` and `GET /task/{id}/form-variables` in parallel; returns a new `taskVariablesLoadedMsg{inputVars, formVars}` message
2. `claimTaskCmd(taskID, userID string) tea.Cmd` — calls `POST /task/{id}/claim` with `UserIdDto{UserId: &userID}`; returns `actionExecutedMsg` or `errMsg`
3. `unclaimTaskCmd(taskID string) tea.Cmd` — calls `POST /task/{id}/unclaim`; returns `actionExecutedMsg` or `errMsg`
4. `completeTaskCmd(taskID string, vars map[string]operaton.VariableValueDto) tea.Cmd` — calls `POST /task/{id}/complete` with `CompleteTaskDto{Variables: vars}`; returns `actionExecutedMsg` or `errMsg`

### Task 4: Key handling in update.go

**Files:** `internal/app/update.go`

1. `c` key handler (task table context):
   - Read `assignee` from selected row
   - Guards: empty → footer error; foreign → footer error; own → footer hint; else → `claimTaskCmd`
2. `Enter` key handler — intercept when `m.currentRoot == "task"` before drilldown logic:
   - Guards: empty assignee → footer error; foreign → footer error; own → `fetchTaskVariablesCmd` + spinner
3. `taskVariablesLoadedMsg` handler: build `taskCompleteFields` (pre-fill + init textinputs), set `m.activeModal = ModalTaskComplete`
4. `u` key handler (task table context):
   - Guards: empty → footer error; foreign → footer error; own → `unclaimTaskCmd`
5. `ModalTaskComplete` key handlers:
   - `Tab` / `Shift+Tab`: advance/reverse focus through fields → Complete → Back
   - `Space`: toggle boolean field value when focused on a field
   - `Enter` on field: advance to `[Complete]`; on Complete (if valid): `completeTaskCmd`; on Back: close
   - `Esc`: close dialog, clear state
   - Any printable character when on a field: feed to `textinput.Model`
6. `actionExecutedMsg` handler for claim/unclaim/complete: refresh task list row

### Task 5: Dialog rendering in view.go

**Files:** `internal/app/view.go`

1. Add `renderTaskCompleteModal(width, height int) string` function:
   - Title bar: `Complete Task`
   - Subtitle: `m.taskCompleteTaskName`
   - Section separator (36 dashes)
   - `INPUT VARIABLES` header + rows `"  name : value"` in `FgMuted` style (or `No context variables`)
   - Section separator
   - `OUTPUT VARIABLES` header + rows `"  name │ <input.View()>"` with inline `⚠ error` lines (or `No output variables defined for this task`)
   - Section separator
   - Button row: `[ Complete ]  [ Back ]` with focus styles matching edit modal (`BtnSaveFocused`, `BtnDisabled`, `BtnSave`, `BtnCancel`, `BtnCancelFocused`)
   - Hint line: `Tab: next  Space: toggle bool  Esc: back` in `FgMuted`
   - Full-screen overlay via `overlayCenter()`
2. Add `ModalTaskComplete` branch to the modal overlay dispatch in `View()`

### Task 6: Tests

**Files:** `internal/app/claim_complete_task_test.go`

1. Claim guard: `c` on unclaimed → `claimTaskCmd` dispatched; on foreign → footer error; on own → footer hint
2. Unclaim guard: `u` on own → `unclaimTaskCmd`; on empty → footer error; on foreign → footer error
3. Enter guard: on task table with own task → `fetchTaskVariablesCmd`; on non-task table → drilldown unchanged
4. Pre-fill: output field matching input variable name gets input variable value
5. Tab cycle: field[0] → field[1] → Complete → Back → field[0]
6. Boolean toggle: Space on boolean field cycles true/false
7. Submit: `completeTaskCmd` called with correctly assembled `Variables` map
8. Close: `Esc` clears all `taskComplete*` state, `m.activeModal == ModalNone`
9. Validation gate: `[Complete]` disabled when any field has non-empty error

## Dev Notes

### Architecture

- **New modal:** `ModalTaskComplete` joins the existing `ModalNone / ModalConfirmDelete / ModalConfirmQuit / ModalHelp / ModalEdit / ModalSort / ModalDetailView` enum. Overlay rendering follows the same `overlayCenter(baseView, overlay)` pattern used by `renderEditModal`.
- **`{currentUser}` placeholder:** Add to body template resolution in the same location as `{id}`, `{name}`, `{value}`, `{type}` substitutions (`commands.go:executeEditActionCmd` or wherever `ActionDef.Body` is resolved). Value: `m.config.Environments[m.currentEnv].Username`.
- **Parallel fetch:** `fetchTaskVariablesCmd` must issue both API calls and return a single message when both complete. Use a goroutine pair with a channel or two sequential calls in the same goroutine (simpler, acceptable latency for task variables).
- **Variable type mapping** (API → validation system):
  - `"String"` → `"text"`
  - `"Boolean"` → `"bool"`
  - `"Integer"` → `"int"`
  - `"Double"` → `"float"`
  - `"Json"` → `"json"`
  - unknown → `"text"`
- **Ordering:** `map[string]VariableValueDto` is unordered; sort both input variables and output fields alphabetically by name for deterministic display.
- **Multiple `textinput.Model` instances:** Unlike the edit modal (one `editInput`), the completion dialog needs one `textinput.Model` per form variable field. Initialize all at dialog open; focus the first one. Only the currently focused field receives key events.
- **Enter on task table:** Intercept before the drilldown logic in `update.go`. Condition: `m.currentRoot == "task" || m.currentTableKey() == "task"`. The task table's two drilldowns (`variable-instance`, `history-detail`) remain accessible via `→` / right arrow.
- **`o8n-cfg.yaml` task action cleanup:** The old bare `c: POST /task/{id}/complete body:'{}'` must be removed. Leaving it would conflict with the new claim binding and bypass the dialog.

### API Endpoints Used

| Purpose | Method | Path | Body / Response |
|---|---|---|---|
| Claim | POST | `/task/{id}/claim` | `UserIdDto{userId}` → 204 |
| Unclaim | POST | `/task/{id}/unclaim` | — → 204 |
| Input variables | GET | `/task/{id}/variables` | → `map[string]VariableValueDto` |
| Form (output) variables | GET | `/task/{id}/form-variables` | → `map[string]VariableValueDto` |
| Complete | POST | `/task/{id}/complete` | `CompleteTaskDto{variables}` → 204 |

### Key Bindings (task table)

| Key | Condition | Action |
|---|---|---|
| `c` | assignee empty | Claim → `POST /task/{id}/claim` |
| `c` | assignee = me | Footer hint only |
| `c` | assignee = other | Footer error only |
| `Enter` | assignee = me | Fetch variables → open dialog |
| `Enter` | assignee ≠ me | Footer error only |
| `Enter` | non-task table | Existing drilldown |
| `u` | assignee = me | Unclaim → `POST /task/{id}/unclaim` |
| `u` | assignee empty/other | Footer error only |

## Dev Agent Record

### Implementation Plan

Implemented all 6 tasks in order:
1. Config changes + `{currentUser}` placeholder
2. Model state extensions (`ModalTaskComplete`, `taskCompleteField`, `taskCompleteFocusArea`, model fields)
3. New commands (`fetchTaskVariablesCmd`, `claimTaskCmd`, `unclaimTaskCmd`, `completeTaskCmd`)
4. Key handling (`c`, `u`, `Enter` guards; `ModalTaskComplete` handlers; helper methods)
5. Dialog rendering (`renderTaskCompleteModal`, overlay dispatch)
6. Tests (`claim_complete_task_test.go` — 26 test functions covering all ACs)

Key design decisions:
- `variableValue` struct in model.go avoids importing `operaton` in model (clean separation)
- `OperatonAPI()` and `AuthContext()` accessor methods added to `CompatClient` to expose SDK
- `closeTaskDialog bool` field added to `actionExecutedMsg` to signal dialog close on complete
- `errMsg` handler preserves table rows when `ModalTaskComplete` is active (AC-29)
- `sortStrings()` bubble sort helper in view.go avoids importing `sort` in view package
- Tab cycle: field[0]→…→field[N-1]→Complete→Back→field[0]
- `testConfigWithActions()` in actions_test.go updated to match new task table config

### Debug Log
_No issues encountered_

### Completion Notes
All 38 ACs implemented and verified. All existing tests pass. 26 new tests cover claim/unclaim/enter guards, dialog lifecycle, tab navigation, boolean toggle, pre-fill, submit, and validation.

## File List

- `o8n-cfg.yaml` — removed `c: Complete Task`, renamed `k: Claim Task` → `c: Claim Task` with `{currentUser}` body
- `internal/app/model.go` — added `ModalTaskComplete`, `taskCompleteFocusArea`, `taskCompleteField`, `variableValue`, `taskVariablesLoadedMsg`, `closeTaskDialog` field on `actionExecutedMsg`, task complete model fields
- `internal/app/commands.go` — added `operaton` import, `{currentUser}` resolution in `executeActionCmd`, `fetchTaskVariablesCmd`, `claimTaskCmd`, `unclaimTaskCmd`, `completeTaskCmd`, `getVarTypeName`
- `internal/client/client.go` — added `OperatonAPI()` and `AuthContext()` accessor methods
- `internal/app/update.go` — added `textinput`, `operaton`, `validation` imports; `ModalTaskComplete` key handler; `c`/`u` key handlers; `Enter` intercept for task table; `taskVariablesLoadedMsg` handler; modified `actionExecutedMsg` and `errMsg` handlers; added helper methods
- `internal/app/view.go` — added `renderTaskCompleteModal`, `sortStrings`, `ModalTaskComplete` branch in View()
- `internal/app/actions_test.go` — updated `testConfigWithActions()` to reflect new task table config
- `internal/app/claim_complete_task_test.go` — new test file (26 test functions)

## Change Log

- 2026-02-27: Initial implementation — all 6 tasks complete, all 38 ACs covered

## Status

done
