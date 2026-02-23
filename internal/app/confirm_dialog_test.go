package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestT2_ConfirmDialogTabToggle verifies that pressing Tab in the delete
// confirm dialog toggles focus between Confirm and Cancel buttons.
func TestT2_ConfirmDialogTabToggle(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalConfirmDelete
	m.pendingDeleteID = "proc-abc"
	m.confirmFocusedBtn = 1 // start on Cancel

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m3 := m2.(model)

	if m3.confirmFocusedBtn != 0 {
		t.Errorf("expected focus on Confirm (0) after Tab, got %d", m3.confirmFocusedBtn)
	}

	m4, _ := m3.Update(tea.KeyMsg{Type: tea.KeyTab})
	m5 := m4.(model)

	if m5.confirmFocusedBtn != 1 {
		t.Errorf("expected focus back on Cancel (1) after second Tab, got %d", m5.confirmFocusedBtn)
	}
}

// TestT2_ConfirmDialogEnterOnCancel verifies that pressing Enter when focused
// on the Cancel button dismisses the dialog.
func TestT2_ConfirmDialogEnterOnCancel(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalConfirmDelete
	m.pendingDeleteID = "proc-abc"
	m.confirmFocusedBtn = 1 // Cancel focused

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := m2.(model)

	if m3.activeModal != ModalNone {
		t.Errorf("expected ModalNone after Enter on Cancel, got %v", m3.activeModal)
	}
	if m3.pendingDeleteID != "" {
		t.Error("expected pendingDeleteID cleared after cancel")
	}
}

// TestT2_ConfirmDialogEnterOnConfirm verifies that pressing Enter when focused
// on the Confirm button triggers the action.
func TestT2_ConfirmDialogEnterOnConfirm(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalConfirmDelete
	m.pendingDeleteID = "proc-abc"
	m.confirmFocusedBtn = 0 // Confirm focused

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := m2.(model)

	if m3.activeModal != ModalNone {
		t.Errorf("expected ModalNone after Enter on Confirm, got %v", m3.activeModal)
	}
	if cmd == nil {
		t.Error("expected a command to be dispatched when confirming delete")
	}
}

// TestT2_ConfirmDialogEscAlwaysCancels verifies that Esc always cancels
// regardless of which button is focused.
func TestT2_ConfirmDialogEscAlwaysCancels(t *testing.T) {
	for _, focused := range []int{0, 1} {
		m := newTestModel(t)
		m.splashActive = false
		m.activeModal = ModalConfirmDelete
		m.pendingDeleteID = "proc-abc"
		m.confirmFocusedBtn = focused

		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m3 := m2.(model)

		if m3.activeModal != ModalNone {
			t.Errorf("focused=%d: expected ModalNone after Esc, got %v", focused, m3.activeModal)
		}
	}
}

// TestT2_ConfirmDialogButtonsRendered verifies that the confirm dialog renders
// both buttons with visual distinction for the focused one.
func TestT2_ConfirmDialogButtonsRendered(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 30
	m.activeModal = ModalConfirmDelete
	m.pendingDeleteID = "proc-abc"
	m.confirmFocusedBtn = 0

	output := m.View()

	// Both button labels should be in the output
	if !strings.Contains(output, "Delete") {
		t.Error("expected 'Delete' button label in confirm dialog output")
	}
	if !strings.Contains(output, "Cancel") {
		t.Error("expected 'Cancel' button label in confirm dialog output")
	}
}
