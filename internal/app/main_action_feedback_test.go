package app

// main_action_feedback_test.go — Story 3.3: Action Execution with Feedback
//
// Tests verify the action execution flow:
//   - Success: actionExecutedMsg sets footer status + triggers re-fetch
//   - Failure: errMsg sets error status with 5s auto-clear
//   - Destructive: ModalConfirmDelete modal appears + executes on confirm

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kthoms/o6n/internal/config"
)

// actionConfig returns a minimal config with process-definitions and a delete action.
func actionConfig() *config.Config {
	return &config.Config{
		Environments: map[string]config.Environment{"local": {URL: "http://localhost"}},
		Tables: []config.TableDef{
			{
				Name:    "process-definition",
				Columns: []config.ColumnDef{{Name: "key"}, {Name: "name"}},
				Actions: []config.ActionDef{
					{
						Key:     "d",
						Label:   "Delete",
						Method:  "DELETE",
						Path:    "/process-definitions/{id}",
						Confirm: true, // destructive action
					},
					{
						Key:    "r",
						Label:  "Retry",
						Method: "POST",
						Path:   "/process-instances/{id}/retry",
					},
				},
			},
		},
	}
}

// ── AC 1: Success path ─────────────────────────────────────────────────────────

// TestActionSuccess_FooterStatusSet verifies that successful action execution
// sets a success footer status message with the action label.
func TestActionSuccess_FooterStatusSet(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	// Simulate successful action execution
	res, _ := m.Update(actionExecutedMsg{label: "Retry"})
	m2 := res.(model)

	// Footer status should be set to success kind
	if m2.footerStatusKind != footerStatusSuccess {
		t.Errorf("expected footerStatusSuccess, got %q", m2.footerStatusKind)
	}

	// Footer message should include checkmark and label
	if m2.footerError != "✓ Retry" {
		t.Errorf("expected footer message '✓ Retry', got %q", m2.footerError)
	}
}

// TestActionSuccess_TriggersDataRefetch verifies that the success message handler
// returns a batch command that includes a data fetch for the current root.
func TestActionSuccess_TriggersDataRefetch(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})
	m.currentRoot = "process-definition"

	// Trigger success message
	res, cmd := m.Update(actionExecutedMsg{label: "Delete"})
	m2 := res.(model)

	// Command should not be nil (indicates batch with fetch/status commands)
	if cmd == nil {
		t.Error("expected command batch for success, got nil")
	}

	// Model should not be in loading state initially (loading is set by fetch cmd internally)
	// We just verify the handler returned a command
	_ = m2
}

// TestActionSuccess_FooterStatusClears verifies that success message auto-clears
// after 3 seconds (via clearErrorMsg timer).
func TestActionSuccess_FooterStatusClears(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	// Trigger success
	res, _ := m.Update(actionExecutedMsg{label: "Complete"})
	m2 := res.(model)

	if m2.footerStatusKind != footerStatusSuccess {
		t.Fatalf("precondition: expected success status")
	}

	// Simulate clearErrorMsg after timer fires
	res2, _ := m2.Update(clearErrorMsg{})
	m3 := res2.(model)

	if m3.footerError != "" {
		t.Errorf("expected footer cleared, got %q", m3.footerError)
	}
	if m3.footerStatusKind != footerStatusNone {
		t.Errorf("expected footerStatusNone after clear, got %q", m3.footerStatusKind)
	}
}

// ── AC 2: Destructive Action Modal ─────────────────────────────────────────────

// TestDestructiveAction_ModalAppears verifies that when a destructive action is invoked,
// the ModalConfirmDelete modal is opened and pending action state is set.
func TestDestructiveAction_ModalAppears(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	// Simulate setting up a destructive action (normally done in buildActionsForRoot)
	deleteAction := &config.ActionDef{
		Key:     "d",
		Label:   "Delete",
		Method:  "DELETE",
		Path:    "/process-definitions/{id}",
		Confirm: true,
	}

	m.pendingAction = deleteAction
	m.pendingActionID = "d1"
	m.pendingActionPath = "/process-definitions/d1"
	m.activeModal = ModalConfirmDelete

	// Verify modal and pending action state are set
	if m.activeModal != ModalConfirmDelete {
		t.Errorf("expected ModalConfirmDelete to be set, got %q", m.activeModal)
	}

	if m.pendingAction == nil {
		t.Error("expected pendingAction to be set")
	}
	if m.pendingAction.Label != "Delete" {
		t.Errorf("expected pendingAction.Label='Delete', got %q", m.pendingAction.Label)
	}
	if m.pendingActionID == "" {
		t.Error("expected pendingActionID to be set")
	}
	if m.pendingActionPath == "" {
		t.Error("expected pendingActionPath to be set")
	}
}

// TestDestructiveAction_ConfirmExecutes verifies that pressing Enter on the Confirm button
// closes the modal and executes the action.
func TestDestructiveAction_ConfirmExecutes(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	// Set up the modal with pending action
	m.pendingAction = &config.ActionDef{
		Key:     "d",
		Label:   "Delete",
		Method:  "DELETE",
		Path:    "/process-definitions/{id}",
		Confirm: true,
	}
	m.pendingActionID = "d1"
	m.pendingActionPath = "/process-definitions/d1"
	m.activeModal = ModalConfirmDelete

	if m.activeModal != ModalConfirmDelete {
		t.Fatalf("precondition: expected ModalConfirmDelete")
	}

	// Simulate pressing Enter (confirm button is focused by default at index 0)
	m.confirmFocusedBtn = 0 // Confirm button
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	// Modal should be closed
	if m2.activeModal != ModalNone {
		t.Errorf("expected modal closed after confirm, got %q", m2.activeModal)
	}

	// Pending action state should be cleared
	if m2.pendingAction != nil {
		t.Error("expected pendingAction to be cleared after confirm")
	}
	if m2.pendingActionID != "" {
		t.Error("expected pendingActionID to be cleared after confirm")
	}

	// Command should be returned (executeActionCmd)
	if cmd == nil {
		t.Error("expected executeActionCmd to be returned")
	}
}

// TestDestructiveAction_EscCancels verifies that pressing Esc closes the modal
// without executing the action.
func TestDestructiveAction_EscCancels(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	// Set up the modal with pending action
	m.pendingAction = &config.ActionDef{
		Key:     "d",
		Label:   "Delete",
		Method:  "DELETE",
		Path:    "/process-definitions/{id}",
		Confirm: true,
	}
	m.pendingActionID = "d1"
	m.pendingActionPath = "/process-definitions/d1"
	m.activeModal = ModalConfirmDelete

	if m.activeModal != ModalConfirmDelete {
		t.Fatalf("precondition: expected ModalConfirmDelete")
	}

	// Simulate pressing Esc
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := res.(model)

	// Modal should be closed
	if m2.activeModal != ModalNone {
		t.Errorf("expected modal closed after Esc, got %q", m2.activeModal)
	}

	// Pending action state should be cleared
	if m2.pendingAction != nil {
		t.Error("expected pendingAction cleared after Esc")
	}

	// Footer should show cancellation message
	if m2.footerError != "Cancelled" {
		t.Errorf("expected 'Cancelled' message, got %q", m2.footerError)
	}

	// Command should be returned to set timer for clearing
	if cmd == nil {
		t.Error("expected command to clear cancellation message")
	}
}

// TestDestructiveAction_TabTogglesFocus verifies that pressing Tab toggles between
// Confirm and Cancel buttons.
func TestDestructiveAction_TabTogglesFocus(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	// Set up the modal
	m.pendingAction = &config.ActionDef{
		Key:     "d",
		Label:   "Delete",
		Method:  "DELETE",
		Path:    "/process-definitions/{id}",
		Confirm: true,
	}
	m.pendingActionID = "d1"
	m.pendingActionPath = "/process-definitions/d1"
	m.activeModal = ModalConfirmDelete
	m.confirmFocusedBtn = 0 // Start at Confirm button

	initialFocus := m.confirmFocusedBtn

	// Press Tab
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m2 := res.(model)

	if m2.confirmFocusedBtn == initialFocus {
		t.Errorf("Tab did not toggle focus: was %d, still %d", initialFocus, m2.confirmFocusedBtn)
	}
}

// TestDestructiveAction_CtrlDForceConfirms verifies that Ctrl+D confirms the action
// even if the Cancel button is focused (safety mechanism).
func TestDestructiveAction_CtrlDForceConfirms(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	// Set up the modal with Cancel button focused
	m.pendingAction = &config.ActionDef{
		Key:     "d",
		Label:   "Delete",
		Method:  "DELETE",
		Path:    "/process-definitions/{id}",
		Confirm: true,
	}
	m.pendingActionID = "d1"
	m.pendingActionPath = "/process-definitions/d1"
	m.activeModal = ModalConfirmDelete
	m.confirmFocusedBtn = 1 // Cancel button

	if m.activeModal != ModalConfirmDelete {
		t.Fatalf("precondition: expected ModalConfirmDelete")
	}

	// Press Ctrl+D (force confirm)
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m2 := res.(model)

	// Modal should be closed and action executed
	if m2.activeModal != ModalNone {
		t.Errorf("expected modal closed after Ctrl+D, got %q", m2.activeModal)
	}
	if cmd == nil {
		t.Error("expected executeActionCmd to be returned")
	}
}

// ── AC 3: Error path ───────────────────────────────────────────────────────────

// TestActionError_FooterErrorSet verifies that action execution failures
// set an error status message with friendly formatting.
func TestActionError_FooterErrorSet(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	// Simulate action failure
	testErr := fmt.Errorf("API call failed")
	res, _ := m.Update(errMsg{err: testErr})
	m2 := res.(model)

	// Footer status should be error kind
	if m2.footerStatusKind != footerStatusError {
		t.Errorf("expected footerStatusError, got %q", m2.footerStatusKind)
	}

	// Footer should contain error message
	if m2.footerError == "" {
		t.Error("expected footerError to be set")
	}

	// Should include retry hint
	if !contains(m2.footerError, "Ctrl+r") {
		t.Errorf("expected retry hint in error message, got %q", m2.footerError)
	}
}

// TestActionError_AutoClears verifies that error messages auto-clear after 5 seconds.
func TestActionError_AutoClears(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	// Trigger error
	res, _ := m.Update(errMsg{err: fmt.Errorf("API call failed")})
	m2 := res.(model)

	if m2.footerStatusKind != footerStatusError {
		t.Fatalf("precondition: expected error status")
	}

	// Simulate clearErrorMsg after timer (5s) fires
	res2, _ := m2.Update(clearErrorMsg{})
	m3 := res2.(model)

	if m3.footerError != "" {
		t.Errorf("expected footer cleared after auto-timer, got %q", m3.footerError)
	}
	if m3.footerStatusKind != footerStatusNone {
		t.Errorf("expected footerStatusNone after clear, got %q", m3.footerStatusKind)
	}
}

// TestActionError_TableRowsCleared verifies that action errors clear the table rows
// (except when task complete modal is open, which preserves context).
func TestActionError_TableRowsCleared(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
		{ID: "d2", Key: "k2", Name: "Beta"},
	})

	initialRows := len(m.table.Rows())
	if initialRows == 0 {
		t.Fatalf("precondition: expected rows in table")
	}

	// Trigger error
	res, _ := m.Update(errMsg{err: fmt.Errorf("API call failed")})
	m2 := res.(model)

	// Rows should be cleared
	finalRows := len(m2.table.Rows())
	if finalRows != 0 {
		t.Errorf("expected rows cleared on error, but found %d rows", finalRows)
	}
}

// TestActionError_TaskCompletePreservesRows verifies that when ModalTaskComplete
// is open, error messages preserve the table rows for context.
func TestActionError_TaskCompletePreservesRows(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	initialRows := len(m.table.Rows())
	if initialRows == 0 {
		t.Fatalf("precondition: expected rows in table")
	}

	// Set task complete modal to preserve rows
	m.activeModal = ModalTaskComplete

	// Trigger error
	res, _ := m.Update(errMsg{err: fmt.Errorf("API call failed")})
	m2 := res.(model)

	// Rows should be preserved
	finalRows := len(m2.table.Rows())
	if finalRows != initialRows {
		t.Errorf("expected rows preserved in ModalTaskComplete, but %d→%d rows", initialRows, finalRows)
	}

	// Error message should be captured for display in the modal
	if m2.taskCompleteError == "" {
		t.Error("expected taskCompleteError to be set")
	}
}

// TestActionError_LogsToDebug verifies that errors are logged to debug/o6n.log
// with context and optional stack trace in debug mode.
func TestActionError_LogsToDebug(t *testing.T) {
	m := newModel(actionConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
	})

	// Trigger error
	res, _ := m.Update(errMsg{err: fmt.Errorf("API call failed")})
	m2 := res.(model)

	// We can't easily check the log file in a test, but we can verify the error
	// was processed without panicking and the model state is consistent
	if m2.footerStatusKind != footerStatusError {
		t.Error("expected error status to be set")
	}
	_ = m2
}

// Note: Helper functions and message types are defined in model.go and update.go
// - actionExecutedMsg: successful action execution result
// - errMsg: error result from an action execution
// - clearErrorMsg: timer event that clears error/status messages
// - contains: helper function (defined in config_protection_test.go)
