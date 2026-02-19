package main

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kthoms/o8n/internal/config"
)

// newTestModel builds a minimal model for keybinding tests.
func newTestModel(t *testing.T) model {
	t.Helper()
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"local": {URL: "http://localhost:8080", UIColor: "#00A8E1"},
		},
	}
	return newModel(cfg)
}

// sendKeyString feeds a raw key string through the switch in Update.
// It wraps the string in a tea.KeyMsg by setting the internal field via
// casting, which isn't possible — instead we use the public API approach:
// send an actual key message with matching Type for ctrl keys.
func sendKeyString(m model, keyStr string) (model, tea.Cmd) {
	var msg tea.Msg
	switch keyStr {
	case "ctrl+c":
		msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+d":
		msg = tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+e":
		msg = tea.KeyMsg{Type: tea.KeyCtrlE}
	case "ctrl+r":
		msg = tea.KeyMsg{Type: tea.KeyCtrlR}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "backspace":
		msg = tea.KeyMsg{Type: tea.KeyBackspace}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyStr)}
	}
	res, cmd := m.Update(msg)
	return res.(model), cmd
}

// TestTabCompletionInContextPopup verifies the bug fix: Tab in context popup
// completes rootInput to the first matching root context.
func TestTabCompletionInContextPopup(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition", "process-instance", "task"}
	m.showRootPopup = true
	m.rootInput = "proc"

	m2, _ := sendKeyString(m, "tab")

	if !m2.showRootPopup {
		t.Error("expected popup to remain open after Tab")
	}
	if m2.rootInput != "process-definition" {
		t.Errorf("expected rootInput completed to 'process-definition', got %q", m2.rootInput)
	}
}

// TestTabCompletionRequiresInput verifies Tab does nothing when rootInput is empty.
func TestTabCompletionRequiresInput(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition"}
	m.showRootPopup = true
	m.rootInput = ""

	m2, _ := sendKeyString(m, "tab")

	if m2.rootInput != "" {
		t.Errorf("expected rootInput to stay empty, got %q", m2.rootInput)
	}
}

// TestTabNoPopupIsNoOp verifies Tab when popup is closed does nothing (no crash).
func TestTabNoPopupIsNoOp(t *testing.T) {
	m := newTestModel(t)
	m.showRootPopup = false
	m.rootInput = ""

	m2, _ := sendKeyString(m, "tab")
	_ = m2 // must not panic
}

// TestHelpModalOpenAndDismiss verifies ? opens ModalHelp and any key closes it.
func TestHelpModalOpenAndDismiss(t *testing.T) {
	m := newTestModel(t)

	m2, _ := sendKeyString(m, "?")
	if m2.activeModal != ModalHelp {
		t.Fatalf("expected ModalHelp after '?', got %v", m2.activeModal)
	}

	// Any key should dismiss the help modal
	m3, _ := sendKeyString(m2, "x")
	if m3.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after key in help modal, got %v", m3.activeModal)
	}
}

// TestCtrlDOpensConfirmModal verifies first ctrl+d opens the confirmation modal.
func TestCtrlDOpensConfirmModal(t *testing.T) {
	m := newTestModel(t)
	m.viewMode = "instances"
	m.selectedInstanceID = "i1"
	m.table.SetRows([]table.Row{{"i1", "k1", "bk1", "2020-01-01"}})

	m2, _ := sendKeyString(m, "ctrl+d")

	if m2.activeModal != ModalConfirmDelete {
		t.Fatalf("expected ModalConfirmDelete after first ctrl+d, got %v", m2.activeModal)
	}
	if m2.pendingDeleteID != "i1" {
		t.Errorf("expected pendingDeleteID 'i1', got %q", m2.pendingDeleteID)
	}
}

// TestCtrlDCancelClearsModal verifies a non-confirm key cancels the delete modal.
func TestCtrlDCancelClearsModal(t *testing.T) {
	m := newTestModel(t)
	m.viewMode = "instances"
	m.selectedInstanceID = "i1"
	m.table.SetRows([]table.Row{{"i1", "k1", "bk1", "2020-01-01"}})

	// open modal
	m2, _ := sendKeyString(m, "ctrl+d")
	if m2.activeModal != ModalConfirmDelete {
		t.Fatalf("modal not opened")
	}

	// cancel with a different key
	m3, _ := sendKeyString(m2, "esc")
	if m3.activeModal != ModalNone {
		t.Errorf("expected ModalNone after cancel, got %v", m3.activeModal)
	}
	if m3.footerError != "Cancelled" {
		t.Errorf("expected footerError 'Cancelled', got %q", m3.footerError)
	}
}

// TestCtrlDConfirmExecutesTerminate verifies second ctrl+d confirms and issues a command.
func TestCtrlDConfirmExecutesTerminate(t *testing.T) {
	m := newTestModel(t)
	m.viewMode = "instances"
	m.selectedInstanceID = "i1"
	m.table.SetRows([]table.Row{{"i1", "k1", "bk1", "2020-01-01"}})

	// first ctrl+d: open modal
	m2, _ := sendKeyString(m, "ctrl+d")

	// second ctrl+d: confirm
	m3, cmd := sendKeyString(m2, "ctrl+d")
	if m3.activeModal != ModalNone {
		t.Errorf("expected modal closed after confirm, got %v", m3.activeModal)
	}
	if cmd == nil {
		t.Error("expected a non-nil command (terminate) after confirming delete")
	}
}

// TestCtrlEEnvironmentSwitch verifies ctrl+e cycles to the next environment.
func TestCtrlEEnvironmentSwitch(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"dev":   {URL: "http://dev", UIColor: "#FFA500"},
			"local": {URL: "http://local", UIColor: "#00A8E1"},
		},
		Active: "local",
	}
	m := newModel(cfg)
	initial := m.currentEnv

	m2, _ := sendKeyString(m, "ctrl+e")

	if m2.currentEnv == initial {
		t.Errorf("expected environment to change from %q, but it did not", initial)
	}
}

// TestErrMsgDisplaysInFooter verifies errMsg is stored in footerError.
func TestErrMsgDisplaysInFooter(t *testing.T) {
	m := newTestModel(t)

	res, cmd := m.Update(errMsg{err: fmt.Errorf("connection refused")})
	m2 := res.(model)

	if m2.footerError != "connection refused" {
		t.Errorf("expected footerError 'connection refused', got %q", m2.footerError)
	}
	if cmd == nil {
		t.Error("expected a clearError cmd to be scheduled")
	}
}

// TestClearErrorMsgClearsFooter verifies clearErrorMsg resets footerError.
func TestClearErrorMsgClearsFooter(t *testing.T) {
	m := newTestModel(t)
	m.footerError = "some error"

	res, _ := m.Update(clearErrorMsg{})
	m2 := res.(model)

	if m2.footerError != "" {
		t.Errorf("expected empty footerError after clearErrorMsg, got %q", m2.footerError)
	}
}

// TestEditModalTabCyclesColumns verifies Tab cycles to the next editable column.
func TestEditModalTabCyclesColumns(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {UIColor: "#00FF00"}},
		Tables: []config.TableDef{
			{
				Name: "process-variables",
				Columns: []config.ColumnDef{
					{Name: "name", Visible: true, Editable: false},
					{Name: "value", Visible: true, Editable: true, InputType: "text"},
					{Name: "type", Visible: true, Editable: true, InputType: "text"},
				},
			},
		},
	}
	m := newModel(cfg)
	m.editColumns = []editableColumn{
		{index: 1, def: config.ColumnDef{Name: "value", Editable: true, InputType: "text"}},
		{index: 2, def: config.ColumnDef{Name: "type", Editable: true, InputType: "text"}},
	}
	m.editColumnPos = 0
	m.editTableKey = "process-variables"
	m.activeModal = ModalEdit
	m.table.SetRows([]table.Row{{"myVar", "hello", "String"}})
	m.editRowIndex = 0
	m.editInput.SetValue("hello")

	m2, _ := sendKeyString(m, "tab")

	if m2.editColumnPos != 1 {
		t.Errorf("expected editColumnPos 1 after Tab, got %d", m2.editColumnPos)
	}
}

// TestEditModalEscCancels verifies Esc closes the edit modal without saving.
func TestEditModalEscCancels(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalEdit
	m.editInput.SetValue("unsaved")

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Errorf("expected ModalNone after Esc in edit modal, got %v", m2.activeModal)
	}
}

// TestBreadcrumbNumericNavigation verifies pressing "1" navigates to breadcrumb level 0.
func TestBreadcrumbNumericNavigation(t *testing.T) {
	m := newTestModel(t)
	// simulate having drilled down: breadcrumb has 2 entries
	m.breadcrumb = []string{"process-definitions", "process-instances"}
	m.viewMode = "instances"
	m.showRootPopup = false

	// pressing "1" should navigate to breadcrumb index 0 (process-definitions)
	res, _ := sendKeyString(m, "1")
	m2 := res

	// After navigating to level 0, viewMode should return to definitions
	if m2.viewMode != "definitions" {
		t.Errorf("expected definitions view after pressing '1', got %q", m2.viewMode)
	}
}
