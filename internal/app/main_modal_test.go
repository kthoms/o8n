package app

import (
	"strings"
	"testing"
)

// TestModalRegistry_AllTypesRegistered verifies that all expected ModalTypes
// have a config registered in the modalRegistry.
func TestModalRegistry_AllTypesRegistered(t *testing.T) {
	expected := []ModalType{
		ModalConfirmDelete,
		ModalConfirmQuit,
		ModalSort,
		ModalEnvironment,
		ModalEdit,
		ModalHelp,
		ModalDetailView,
		ModalTaskComplete,
	}
	for _, mt := range expected {
		if _, ok := modalRegistry[mt]; !ok {
			t.Errorf("ModalType %d not registered in modalRegistry", mt)
		}
	}
}

// TestModalRegistry_UnregisteredTypeReturnsEmpty verifies that renderModal
// returns an empty string for unregistered modal types without panicking.
func TestModalRegistry_UnregisteredTypeReturnsEmpty(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120
	m.lastHeight = 24

	// Use a high value that won't match any registered type
	var unknown ModalType = 999
	cfg, ok := modalRegistry[unknown]
	if ok {
		t.Fatalf("expected no config for unknown modal type 999, got: %+v", cfg)
	}
	// Verify View() with unregistered active modal doesn't panic
	m.activeModal = unknown
	m.splashActive = false
	output := m.View()
	if output == "" {
		t.Error("expected non-empty output from View() even with unknown modal type")
	}
}

// TestModalSizeHint_OverlayCenterReturnsStyledBox verifies that OverlayCenter
// modal configs produce styled output (with border characters).
func TestModalSizeHint_OverlayCenterReturnsStyledBox(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120
	m.lastHeight = 24
	m.pendingDeleteID = "proc-test-123"
	m.pendingDeleteLabel = "test-process"

	cfg, ok := modalRegistry[ModalConfirmDelete]
	if !ok {
		t.Fatal("ModalConfirmDelete not in registry")
	}
	if cfg.SizeHint != OverlayCenter {
		t.Errorf("expected OverlayCenter, got %d", cfg.SizeHint)
	}

	output := renderModal(m, cfg)
	if output == "" {
		t.Error("expected non-empty output for ModalConfirmDelete")
	}
	// Should contain the pending delete ID
	if !strings.Contains(output, "proc-test-123") {
		t.Error("expected pendingDeleteID in ConfirmDelete modal output")
	}
	// Should have Delete button label
	if !strings.Contains(output, "Delete") {
		t.Error("expected 'Delete' in ConfirmDelete modal output")
	}
}

// TestModalSizeHint_OverlayCenterQuit verifies ConfirmQuit config.
func TestModalSizeHint_OverlayCenterQuit(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120
	m.lastHeight = 24

	cfg, ok := modalRegistry[ModalConfirmQuit]
	if !ok {
		t.Fatal("ModalConfirmQuit not in registry")
	}
	if cfg.SizeHint != OverlayCenter {
		t.Errorf("expected OverlayCenter, got %d", cfg.SizeHint)
	}
	output := renderModal(m, cfg)
	if !strings.Contains(output, "Quit") {
		t.Error("expected 'Quit' in ConfirmQuit modal output")
	}
}

// TestModalSizeHint_OverlayLargeHelp verifies that ModalHelp uses OverlayLarge
// size hint and renders the HintLine.
func TestModalSizeHint_OverlayLargeHelp(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120
	m.lastHeight = 24

	cfg, ok := modalRegistry[ModalHelp]
	if !ok {
		t.Fatal("ModalHelp not in registry")
	}
	if cfg.SizeHint != OverlayLarge {
		t.Errorf("expected OverlayLarge for ModalHelp, got %d", cfg.SizeHint)
	}
	if len(cfg.HintLine) == 0 {
		t.Error("expected non-empty HintLine for ModalHelp (required for OverlayLarge)")
	}
	// Verify HintLine contains required fields
	hasEsc := false
	for _, h := range cfg.HintLine {
		if strings.Contains(h.Key, "Esc") || strings.Contains(h.Label, "close") {
			hasEsc = true
		}
	}
	if !hasEsc {
		t.Error("expected ModalHelp HintLine to include Esc/close hint")
	}

	output := renderModal(m, cfg)
	if output == "" {
		t.Error("expected non-empty output for ModalHelp")
	}
	// Should contain help content
	if !strings.Contains(output, "o8n Help") {
		t.Error("expected 'o8n Help' in ModalHelp output")
	}
	// OverlayLarge output should contain the hint line rendered from HintLine
	if !strings.Contains(output, "scroll") {
		t.Error("expected scroll hint in ModalHelp output")
	}
}

// TestModalSizeHint_OverlayLargeDetailView verifies that ModalDetailView uses
// OverlayLarge size hint and has a populated HintLine.
func TestModalSizeHint_OverlayLargeDetailView(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120
	m.lastHeight = 24
	m.detailContent = `{"id": "test-123", "state": "ACTIVE"}`

	cfg, ok := modalRegistry[ModalDetailView]
	if !ok {
		t.Fatal("ModalDetailView not in registry")
	}
	if cfg.SizeHint != OverlayLarge {
		t.Errorf("expected OverlayLarge for ModalDetailView, got %d", cfg.SizeHint)
	}
	if len(cfg.HintLine) == 0 {
		t.Error("expected non-empty HintLine for ModalDetailView (required for OverlayLarge)")
	}

	output := renderModal(m, cfg)
	if output == "" {
		t.Error("expected non-empty output for ModalDetailView")
	}
	if !strings.Contains(output, "Detail View") {
		t.Error("expected 'Detail View' in ModalDetailView output")
	}
}

// TestModalSizeHint_FullScreenTaskComplete verifies that ModalTaskComplete uses
// FullScreen size hint and has a populated HintLine.
func TestModalSizeHint_FullScreenTaskComplete(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120
	m.lastHeight = 24
	m.taskCompleteTaskName = "Review Invoice"

	cfg, ok := modalRegistry[ModalTaskComplete]
	if !ok {
		t.Fatal("ModalTaskComplete not in registry")
	}
	if cfg.SizeHint != FullScreen {
		t.Errorf("expected FullScreen for ModalTaskComplete, got %d", cfg.SizeHint)
	}
	if len(cfg.HintLine) == 0 {
		t.Error("expected non-empty HintLine for ModalTaskComplete (required for FullScreen)")
	}

	output := renderModal(m, cfg)
	if output == "" {
		t.Error("expected non-empty output for ModalTaskComplete")
	}
}

// TestModalHintLine_Rendering verifies renderModalHintLine produces readable output.
func TestModalHintLine_Rendering(t *testing.T) {
	m := newTestModel(t)
	hints := []Hint{
		{Key: "↑↓", Label: "scroll", Priority: 1},
		{Key: "Esc", Label: "close", Priority: 1},
	}
	output := renderModalHintLine(m, hints)
	if output == "" {
		t.Error("expected non-empty hint line output")
	}
	if !strings.Contains(output, "↑↓") {
		t.Error("expected '↑↓' key in hint line output")
	}
	if !strings.Contains(output, "scroll") {
		t.Error("expected 'scroll' label in hint line output")
	}
	if !strings.Contains(output, "Esc") {
		t.Error("expected 'Esc' key in hint line output")
	}
	if !strings.Contains(output, "close") {
		t.Error("expected 'close' label in hint line output")
	}
}

// TestModalHintLine_EmptyReturnsEmpty verifies renderModalHintLine handles empty hints.
func TestModalHintLine_EmptyReturnsEmpty(t *testing.T) {
	m := newTestModel(t)
	output := renderModalHintLine(m, nil)
	if output != "" {
		t.Errorf("expected empty string for nil hints, got: %q", output)
	}
	output = renderModalHintLine(m, []Hint{})
	if output != "" {
		t.Errorf("expected empty string for empty hints, got: %q", output)
	}
}

// TestModalConfirmLabel_Populated verifies confirm/cancel labels are populated.
func TestModalConfirmLabel_Populated(t *testing.T) {
	cases := []struct {
		modal         ModalType
		expectConfirm string
		expectCancel  string
	}{
		{ModalConfirmDelete, "Delete", "Cancel"},
		{ModalConfirmQuit, "Quit", "Cancel"},
		{ModalEdit, "Save", "Cancel"},
	}
	for _, tc := range cases {
		cfg, ok := modalRegistry[tc.modal]
		if !ok {
			t.Errorf("ModalType %d not in registry", tc.modal)
			continue
		}
		if cfg.ConfirmLabel != tc.expectConfirm {
			t.Errorf("modal %d: expected ConfirmLabel %q, got %q", tc.modal, tc.expectConfirm, cfg.ConfirmLabel)
		}
		if cfg.CancelLabel != tc.expectCancel {
			t.Errorf("modal %d: expected CancelLabel %q, got %q", tc.modal, tc.expectCancel, cfg.CancelLabel)
		}
	}
}

// TestModalFactory_ViewDispatch verifies that View() dispatches all registered
// modal types through the factory without panicking.
func TestModalFactory_ViewDispatch(t *testing.T) {
	registeredTypes := []ModalType{
		ModalConfirmDelete,
		ModalConfirmQuit,
		ModalSort,
		ModalEnvironment,
		ModalHelp,
		ModalDetailView,
	}

	for _, mt := range registeredTypes {
		m := newTestModel(t)
		m.splashActive = false
		m.lastWidth = 120
		m.lastHeight = 24
		m.activeModal = mt
		// Set state needed for modals that check preconditions
		if mt == ModalConfirmDelete {
			m.pendingDeleteID = "test-id"
		}
		if mt == ModalDetailView {
			m.detailContent = `{"key": "value"}`
		}

		output := m.View()
		if output == "" {
			t.Errorf("expected non-empty View() output for modal type %d", mt)
		}
	}
}

// TestOverlayFunctions_OverlayLargeDelegatesToCenter verifies overlayLarge
// produces centered output.
func TestOverlayFunctions_OverlayLargeDelegatesToCenter(t *testing.T) {
	bg := strings.Repeat("background content line here\n", 20)
	fg := "modal content"
	large := overlayLarge(bg, fg)
	center := overlayCenter(bg, fg)
	if large != center {
		t.Error("expected overlayLarge to produce same result as overlayCenter")
	}
}

// TestOverlayFunctions_OverlayFullscreenIgnoresBg verifies overlayFullscreen
// produces a result that does not include the background.
func TestOverlayFunctions_OverlayFullscreenIgnoresBg(t *testing.T) {
	bg := "background content that should not appear"
	fg := "fullscreen modal content"
	output := overlayFullscreen(bg, fg, 80, 24)
	if strings.Contains(output, "background content") {
		t.Error("expected overlayFullscreen to not include background content")
	}
	if !strings.Contains(output, "fullscreen modal content") {
		t.Error("expected overlayFullscreen output to contain fg content")
	}
}

// TestSizeHint_Constants verifies the SizeHint constants have the expected values.
func TestSizeHint_Constants(t *testing.T) {
	if OverlayCenter != 0 {
		t.Errorf("expected OverlayCenter = 0, got %d", OverlayCenter)
	}
	if OverlayLarge != 1 {
		t.Errorf("expected OverlayLarge = 1, got %d", OverlayLarge)
	}
	if FullScreen != 2 {
		t.Errorf("expected FullScreen = 2, got %d", FullScreen)
	}
}

// TestHintStruct_Fields verifies the Hint struct has the expected fields.
func TestHintStruct_Fields(t *testing.T) {
	h := Hint{
		Key:      "Esc",
		Label:    "close",
		MinWidth: 0,
		Priority: 1,
	}
	if h.Key != "Esc" {
		t.Errorf("expected Key 'Esc', got %q", h.Key)
	}
	if h.Label != "close" {
		t.Errorf("expected Label 'close', got %q", h.Label)
	}
	if h.MinWidth != 0 {
		t.Errorf("expected MinWidth 0, got %d", h.MinWidth)
	}
	if h.Priority != 1 {
		t.Errorf("expected Priority 1, got %d", h.Priority)
	}
}

// TestRegisterModal_Override verifies registerModal can override an existing registration.
func TestRegisterModal_Override(t *testing.T) {
	original, exists := modalRegistry[ModalSort]
	if !exists {
		t.Fatal("ModalSort should be registered")
	}

	// Register a temporary override
	registerModal(ModalSort, ModalConfig{
		SizeHint: OverlayLarge,
		BodyRenderer: func(m model) string {
			return "custom sort body"
		},
	})

	cfg := modalRegistry[ModalSort]
	if cfg.SizeHint != OverlayLarge {
		t.Error("expected override to take effect")
	}

	// Restore original
	registerModal(ModalSort, original)
	restored := modalRegistry[ModalSort]
	if restored.SizeHint != OverlayCenter {
		t.Error("expected original OverlayCenter restored")
	}
}

// ── Esc key handler tests (AC 1) ──────────────────────────────────────────────

// TestModalEsc_ConfirmDelete verifies Esc dismisses ModalConfirmDelete without executing the action.
func TestModalEsc_ConfirmDelete(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalConfirmDelete
	m.pendingDeleteID = "proc-123"

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Esc on ModalConfirmDelete, got %v", m2.activeModal)
	}
}

// TestModalEsc_ConfirmQuit verifies Esc dismisses ModalConfirmQuit without quitting.
func TestModalEsc_ConfirmQuit(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalConfirmQuit

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Esc on ModalConfirmQuit, got %v", m2.activeModal)
	}
	if m2.quitting {
		t.Fatal("expected quitting=false after Esc on ModalConfirmQuit")
	}
}

// TestModalEsc_Sort verifies Esc dismisses ModalSort.
func TestModalEsc_Sort(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalSort

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Esc on ModalSort, got %v", m2.activeModal)
	}
}

// TestModalEsc_Help verifies Esc dismisses ModalHelp and resets scroll position.
func TestModalEsc_Help(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalHelp
	m.helpScroll = 5

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Esc on ModalHelp, got %v", m2.activeModal)
	}
	if m2.helpScroll != 0 {
		t.Fatalf("expected helpScroll=0 after Esc on ModalHelp, got %d", m2.helpScroll)
	}
}

// TestModalEsc_Edit verifies Esc dismisses ModalEdit and clears editError.
func TestModalEsc_Edit(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalEdit
	m.editError = "enter an integer"

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Esc on ModalEdit, got %v", m2.activeModal)
	}
	if m2.editError != "" {
		t.Fatalf("expected editError cleared after Esc on ModalEdit, got %q", m2.editError)
	}
}

// TestModalEsc_Environment verifies Esc dismisses ModalEnvironment.
func TestModalEsc_Environment(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalEnvironment

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Esc on ModalEnvironment, got %v", m2.activeModal)
	}
}

// TestModalEsc_DetailView verifies Esc dismisses ModalDetailView.
func TestModalEsc_DetailView(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalDetailView
	m.detailContent = `{"id": "test-123"}`

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Esc on ModalDetailView, got %v", m2.activeModal)
	}
}

// TestModalEsc_TaskComplete verifies Esc dismisses ModalTaskComplete via closeTaskCompleteDialog.
func TestModalEsc_TaskComplete(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalTaskComplete
	m.taskCompleteTaskName = "Review Invoice"

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Esc on ModalTaskComplete, got %v", m2.activeModal)
	}
}

// ── ModalHelp key behavior tests ──────────────────────────────────────────────

// TestModalHelp_QCloses verifies 'q' closes ModalHelp (consistent with ModalDetailView convention).
func TestModalHelp_QCloses(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalHelp
	m.helpScroll = 3

	m2, _ := sendKeyString(m, "q")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after 'q' on ModalHelp, got %v", m2.activeModal)
	}
	if m2.helpScroll != 0 {
		t.Fatalf("expected helpScroll=0 after 'q' on ModalHelp, got %d", m2.helpScroll)
	}
}

// TestModalHelp_EnterSwallowed verifies Enter does NOT close ModalHelp —
// only Esc and 'q' are explicit close keys; other keys are silently swallowed.
func TestModalHelp_EnterSwallowed(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalHelp

	m2, _ := sendKeyString(m, "enter")

	if m2.activeModal != ModalHelp {
		t.Fatalf("expected ModalHelp to stay open after Enter (only Esc/q close), got %v", m2.activeModal)
	}
}

// TestModalHelp_HintShowsQEsc verifies the ModalHelp HintLine accurately reflects that q/Esc closes it.
func TestModalHelp_HintShowsQEsc(t *testing.T) {
	cfg, ok := modalRegistry[ModalHelp]
	if !ok {
		t.Fatal("ModalHelp not in registry")
	}
	hasClose := false
	for _, h := range cfg.HintLine {
		if strings.Contains(h.Key, "q") && strings.Contains(h.Key, "Esc") {
			hasClose = true
		}
	}
	if !hasClose {
		t.Error("expected ModalHelp HintLine to contain 'q/Esc' close key (hint matches actual behavior)")
	}
}

// ── Enter / confirm handler tests (AC 2) ─────────────────────────────────────

// TestModalEnter_ConfirmQuit_QuitButton verifies Enter on the Quit button triggers quit.
func TestModalEnter_ConfirmQuit_QuitButton(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalConfirmQuit
	m.confirmFocusedBtn = 0 // Quit button

	m2, _ := sendKeyString(m, "enter")

	if !m2.quitting {
		t.Fatal("expected quitting=true after Enter on Quit button of ModalConfirmQuit")
	}
}

// TestModalEnter_ConfirmQuit_CancelButton verifies Enter on the Cancel button dismisses without quitting.
func TestModalEnter_ConfirmQuit_CancelButton(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalConfirmQuit
	m.confirmFocusedBtn = 1 // Cancel button

	m2, _ := sendKeyString(m, "enter")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Enter on Cancel button of ModalConfirmQuit, got %v", m2.activeModal)
	}
	if m2.quitting {
		t.Fatal("expected quitting=false after Enter on Cancel button of ModalConfirmQuit")
	}
}

// TestModalEnter_Sort_ClosesModal verifies Enter applies sort and closes ModalSort.
func TestModalEnter_Sort_ClosesModal(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalSort
	m.sortPopupCursor = -1 // clear-sort option

	m2, _ := sendKeyString(m, "enter")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Enter on ModalSort (clear sort), got %v", m2.activeModal)
	}
	if m2.sortColumn != -1 {
		t.Fatalf("expected sortColumn=-1 after clearing sort, got %d", m2.sortColumn)
	}
	if !m2.sortAscending {
		t.Fatal("expected sortAscending=true after clearing sort")
	}
}

// ── Type validation tests (AC 3) ─────────────────────────────────────────────

// TestParseInputValue_IntegerRejectsText verifies parseInputValue returns an error for non-integer text.
func TestParseInputValue_IntegerRejectsText(t *testing.T) {
	_, err := parseInputValue("abc", "integer")
	if err == nil {
		t.Fatal("expected error when parsing 'abc' as integer, got nil")
	}
}

// TestParseInputValue_IntegerAcceptsNumber verifies parseInputValue accepts a valid integer string.
func TestParseInputValue_IntegerAcceptsNumber(t *testing.T) {
	_, err := parseInputValue("42", "integer")
	if err != nil {
		t.Fatalf("expected no error when parsing '42' as integer, got: %v", err)
	}
}

// TestParseInputValue_BoolRejectsInvalid verifies parseInputValue rejects non-boolean strings.
func TestParseInputValue_BoolRejectsInvalid(t *testing.T) {
	_, err := parseInputValue("yes", "bool")
	if err == nil {
		t.Fatal("expected error when parsing 'yes' as bool, got nil")
	}
}

// TestParseInputValue_BoolAcceptsTrue verifies parseInputValue accepts "true".
func TestParseInputValue_BoolAcceptsTrue(t *testing.T) {
	_, err := parseInputValue("true", "bool")
	if err != nil {
		t.Fatalf("expected no error when parsing 'true' as bool, got: %v", err)
	}
}

// TestParseInputValue_BoolAcceptsFalse verifies parseInputValue accepts "false".
func TestParseInputValue_BoolAcceptsFalse(t *testing.T) {
	_, err := parseInputValue("false", "bool")
	if err != nil {
		t.Fatalf("expected no error when parsing 'false' as bool, got: %v", err)
	}
}

// TestParseInputValue_JsonRejectsInvalid verifies parseInputValue rejects malformed JSON.
func TestParseInputValue_JsonRejectsInvalid(t *testing.T) {
	_, err := parseInputValue("{not valid json}", "json")
	if err == nil {
		t.Fatal("expected error when parsing invalid JSON, got nil")
	}
}

// TestParseInputValue_JsonAcceptsValid verifies parseInputValue accepts well-formed JSON.
func TestParseInputValue_JsonAcceptsValid(t *testing.T) {
	_, err := parseInputValue(`{"key": "value"}`, "json")
	if err != nil {
		t.Fatalf("expected no error when parsing valid JSON, got: %v", err)
	}
}

// TestParseInputValue_TextAlwaysPasses verifies parseInputValue accepts any string as text type.
func TestParseInputValue_TextAlwaysPasses(t *testing.T) {
	cases := []string{"anything", "123", "true", "{broken", ""}
	for _, input := range cases {
		_, err := parseInputValue(input, "text")
		if err != nil {
			t.Errorf("expected no error for text type with input %q, got: %v", input, err)
		}
	}
}

// TestModalEdit_EscClearsErrorAndDismisses verifies the ModalEdit Esc handler
// clears any pending editError before dismissing (prevents stale error on next open).
func TestModalEdit_EscClearsErrorAndDismisses(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalEdit
	m.editError = "enter an integer"

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Esc, got %v", m2.activeModal)
	}
	if m2.editError != "" {
		t.Fatalf("expected editError cleared after Esc, got %q", m2.editError)
	}
}
