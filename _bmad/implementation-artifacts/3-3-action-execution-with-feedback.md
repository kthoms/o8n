# Story 3.3: Action Execution with Feedback

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want to execute any configured action on a selected row and receive clear success or error feedback,
so that I always know whether an action succeeded or failed without guessing.

## Acceptance Criteria

1. **Given** the operator selects a row and presses the configured action key
   **When** the action executes successfully
   **Then** a success message appears in the footer (e.g., `✓ Job retried`)
   **And** the table refreshes to reflect the updated state

2. **Given** a destructive action key is pressed (e.g., delete)
   **When** the operator presses the key
   **Then** a confirmation modal is shown before the action executes
   **And** confirming executes the action; Esc cancels without any side effect

3. **Given** an action's API call fails
   **When** the error is received
   **Then** an error message appears in the footer and auto-clears after 5 seconds
   **And** no silent failure occurs

## Tasks / Subtasks

- [x] Audit `executeActionCmd` in `internal/app/commands.go` (AC: 1, 3)
  - [x] Ensure it returns `actionExecutedMsg` with the correct label on success.
  - [x] Ensure it returns `errMsg` on failure.
- [x] Audit `actionExecutedMsg` handler in `internal/app/update.go` (AC: 1)
  - [x] Verify it uses `setFooterStatus` with `footerStatusSuccess`.
  - [x] Verify it triggers a data re-fetch for the current root.
- [x] Audit `errMsg` handler in `internal/app/update.go` (AC: 3)
  - [x] Verify it uses `setFooterStatus` with `footerStatusError` and 5s auto-clear.
  - [x] Verify it logs the error to `debug/o6n.log`.
- [x] Verify Destructive Action Flow (AC: 2)
  - [x] Check `buildActionsForRoot` in `internal/app/nav.go` correctly sets `m.activeModal = ModalConfirmDelete` for actions with `Confirm: true`.
  - [x] Verify the `ModalConfirmDelete` handler in `update.go` correctly calls `executeActionCmd` on confirmation.
- [x] Tests and Validation (AC: 1, 2, 3)
  - [x] Create `internal/app/main_action_feedback_test.go`.
  - [x] Test success path: verify footer message and re-fetch command.
  - [x] Test failure path: verify footer error message and auto-clear timer.
  - [x] Test confirmation path: verify modal appears and action only runs on confirm.

## Dev Notes

- **Footer Status Model:** Use `setFooterStatus(kind, msg, duration)` helper in `update.go` to ensure consistency.
- **Async Feedback:** Feedback must be non-blocking. The user should be able to continue navigating while the feedback message is displayed.
- **Refresh Strategy:** Data re-fetch after success ensures the UI reflects the server-side state change (e.g., instance removed, variable updated).

### Project Structure Notes

- Logic distributed between `commands.go` (cmd creation), `nav.go` (action routing), and `update.go` (message handling).
- Co-locate tests in `internal/app/`.

### References

- Epic 3 Story 3.3: `_bmad/planning-artifacts/epics.md`
- Footer Hint Push Model: `internal/app/hints.go`
- State Transition Contract: `internal/app/transition.go`

## Dev Agent Record

### Agent Model Used

Claude Haiku 4.5 (implementation)

### Debug Log References

N/A

### Completion Notes List

- **Action execution flow already fully implemented**: `executeActionCmd`, `actionExecutedMsg` handler, and `errMsg` handler all correctly implement the required functionality per AC 1–3.
- **Destructive action modal correctly configured**: `buildActionsForRoot` in nav.go (lines 146-152) correctly sets `m.activeModal = ModalConfirmDelete` when `act.Confirm: true`. `ModalConfirmDelete` handler in update.go (lines 670-717) correctly executes on confirmation via `executeActionCmd` and shows "Cancelled" message on Esc.
- **Comprehensive test suite created**: `internal/app/main_action_feedback_test.go` with 13 tests covering all three acceptance criteria:
  - AC 1 (success): `TestActionSuccess_FooterStatusSet`, `TestActionSuccess_TriggersDataRefetch`, `TestActionSuccess_FooterStatusClears`
  - AC 2 (destructive modal): `TestDestructiveAction_ModalAppears`, `TestDestructiveAction_ConfirmExecutes`, `TestDestructiveAction_EscCancels`, `TestDestructiveAction_TabTogglesFocus`, `TestDestructiveAction_CtrlDForceConfirms`
  - AC 3 (error): `TestActionError_FooterErrorSet`, `TestActionError_AutoClears`, `TestActionError_TableRowsCleared`, `TestActionError_TaskCompletePreservesRows`, `TestActionError_LogsToDebug`
- **All tests passing**: `make test` runs with 100% pass rate across all test suites.

### File List

- `internal/app/main_action_feedback_test.go` — NEW: 13 comprehensive action feedback tests (AC 1, 2, 3)
- `_bmad/implementation-artifacts/3-3-action-execution-with-feedback.md` — UPD: Marked all tasks complete, updated status to review
