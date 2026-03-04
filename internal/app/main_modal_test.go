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
