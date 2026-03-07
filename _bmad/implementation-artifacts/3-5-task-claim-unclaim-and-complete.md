# Story 3.5: Task Claim, Unclaim & Complete

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator** (Priya persona),
I want to **claim, unclaim, and complete user tasks** via the TUI including a form variable dialog,
so that I can process my task queue efficiently without switching to Operaton Cockpit.

## Acceptance Criteria

1. **Given** the operator is viewing the task resource type and selects an unassigned task
   **When** the operator executes the **Claim** action
   **Then** the task is claimed via API (`POST /task/{id}/claim`), the assignee column updates immediately, and the footer confirms.

2. **Given** the operator has a claimed task selected
   **When** the operator executes the **Unclaim** action
   **Then** the task is unclaimed via API (`POST /task/{id}/unclaim`) and the assignee column clears.

3. **Given** the operator has a claimed task selected and executes the **Complete** action
   **When** the **task completion dialog** opens
   **Then** input variables with no corresponding output variable are displayed read-only
   **And** output (form) variables are displayed as editable fields (`textinput.Model`) with type validation (string, integer, boolean)
   **And** output variables whose name matches an input variable are pre-populated with that input variable's value (converted to string) — the input variable is not shown separately as a duplicate read-only row
   **And** output variables with no matching input variable are presented as empty editable fields
   **And** pressing Enter on `[Complete]` submits the completion (`POST /task/{id}/complete`) and the task disappears from the table
   **And** the footer confirms: `✓ Completed: [task name]`.

## Tasks / Subtasks

- [x] **Implement Claim/Unclaim logic (AC: 1, 2)**
  - [x] Add `claimTaskCmd` and `unclaimTaskCmd` to `internal/app/commands.go` using established `client.NewClient` pattern.
  - [x] Map actions in `o6n-cfg.yaml` for `task` resource to these commands.
  - [x] Dispatch `actionExecutedMsg` on success to trigger footer feedback and automated refresh.
- [x] **Initialize Task Completion State (AC: 3, 4)**
  - [x] Add `taskCompleteFields []taskCompleteField` and `taskInputVars map[string]variableValue` to `model.go`.
  - [x] Implement `fetchTaskVariablesCmd` to fetch `GET /task/{id}/variables` and `GET /task/{id}/form-variables` in parallel.
- [x] **Register Modal Factory Dialog (AC: 4)**
  - [x] Register `ModalTaskComplete` in `internal/app/modal.go` using `FullScreen` size hint.
  - [x] Implement `renderTaskCompleteModal` in `view.go` to render read-only context variables and editable form fields.
  - [x] Use `github.com/charmbracelet/bubbles/textinput` for all editable fields.
- [x] **Implement Submission & Validation (AC: 5)**
  - [x] Handle keyboard focus navigation (Tab/Shift+Tab) between text inputs and Submit/Cancel buttons.
  - [x] Add field-level validation for integer and boolean types before allowing submission.
  - [x] Implement `completeTaskCmd` to submit the form data.
  - [x] On success, close modal and dispatch `actionExecutedMsg` with `label` and `closeTaskDialog: true`.
- [x] **Verify & Test (AC: all)**
  - [x] Create `internal/app/main_task_ops_test.go` verifying the variable merging and command emission logic.
  - [x] Run `make test` to ensure zero regressions in navigation or existing modal rendering.

## Dev Notes

### Architecture Compliance
- **Modal Factory:** The Task Completion dialog MUST be rendered via the factory. `FullScreen` modals replace the main table view entirely during the flow.
- **State Transition Contract:** Use `prepareStateTransition(TransitionFull)` if navigation context changes, though Claim/Unclaim should primarily use the `actionExecutedMsg` refresh pattern.
- **Async Pattern:** All network operations must return `tea.Cmd`. Never write to the model from a goroutine.

### UI/UX & Component Usage
- **Bubbles:** Use `textinput.Model` for form fields. Use `table.Model` for the underlying task list.
- **Merged View:** If an output variable name exists in the input variables set, pre-fill the input value into the form field and hide the read-only input entry.
- **Feedback:** Use `setFooterStatus` via the `Update()` loop for all success/error messaging.

### Project Structure Notes
- **Implementation:** `internal/app/commands.go`, `internal/app/view.go`, `internal/app/modal.go`.
- **Testing:** `internal/app/main_task_ops_test.go`.

### References
- [Source: `_bmad/planning-artifacts/epics.md#Story 3.5`]
- [Source: `_bmad/planning-artifacts/architecture.md#Decision 1: Modal Factory Pattern`]
- [Source: `_bmad/implementation-artifacts/3-3-action-execution-with-feedback.md`] (Feedback patterns)

## Dev Agent Record

### Agent Model Used

Claude Haiku 4.5 (implementation)

### Debug Log References

N/A

### Completion Notes List

- **Infrastructure already in place**: Most of the task operations infrastructure was already implemented (claimTaskCmd, unclaimTaskCmd, completeTaskCmd, renderTaskCompleteModal, ModalTaskComplete types). Story focused on integrating missing pieces.
- **Complete action added**: Added key "o" → "Complete" action to task table config in o6n-cfg.yaml. This triggers the task completion modal workflow.
- **Complete action handler added**: Added "o" case in update.go key handler to open ModalTaskComplete and fetch task variables (input + form variables in parallel).
- **Variable merging implementation verified**: Modal correctly merges input variables (read-only context) with form variables (editable fields). Output variables matching input var names are pre-filled; input-only vars shown read-only; output-only vars shown as empty editable fields.
- **Type validation in place**: taskCompleteField tracks varType (lowercased: "string", "boolean", "integer", "double", "json") and origType for API submission.
- **Comprehensive test suite created**: `internal/app/main_task_ops_test.go` with 8 tests covering all acceptance criteria:
  - AC 1 (Claim): User resolution and success feedback
  - AC 2 (Unclaim): Validation for current user ownership
  - AC 3 (Complete): Modal opening and variable fetch
  - AC 4 (Modal): Variable merging, read-only input vars, editable output fields
- **All tests passing**: `make test` runs with 100% pass rate; all existing tests continue to pass.

### File List

- `o6n-cfg.yaml` — UPD: Added "o" (Complete) action to task table
- `internal/app/update.go` — UPD: Added "o" key handler to open task completion modal and fetch variables
- `internal/app/main_task_ops_test.go` — NEW: 8 comprehensive task operations tests (AC 1-4)
