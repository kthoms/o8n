package app

import (
	"strings"
	"testing"
)

// baseViewMarker returns a string that appears in the base view footer/header
// but NOT in any modal content. Used to verify true overlay rendering.
func baseViewMarker(m model) string {
	// The footer always renders " | " as a separator column delimiter.
	// This is distinct from all modal content.
	return " | "
}

// TestT1_HelpModalIsOverlay verifies that the help modal renders as a true
// overlay — both the modal content AND base view content are present.
func TestT1_HelpModalIsOverlay(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 30
	m.activeModal = ModalHelp

	output := m.View()

	if !strings.Contains(output, "o8n Help") {
		t.Error("expected help modal content 'o8n Help' in output")
	}
	// Base view footer always contains " | " separators — proves bg is rendered
	if !strings.Contains(output, baseViewMarker(m)) {
		t.Error("expected base view content (footer separator ' | ') visible behind help modal overlay")
	}
}

// TestT1_SortModalIsOverlay verifies sort popup renders as overlay.
func TestT1_SortModalIsOverlay(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 30
	m.activeModal = ModalSort

	output := m.View()

	if !strings.Contains(output, "Sort by Column") {
		t.Error("expected sort popup content in output")
	}
	if !strings.Contains(output, baseViewMarker(m)) {
		t.Error("expected base view content visible behind sort popup overlay")
	}
}

// TestT1_DetailViewIsOverlay verifies detail view renders as overlay.
func TestT1_DetailViewIsOverlay(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 15 // small enough that modal (height=9) fits with footer visible
	m.activeModal = ModalDetailView
	m.detailContent = `{"id": "test-123"}`

	output := m.View()

	if !strings.Contains(output, "Detail View") {
		t.Error("expected detail view content in output")
	}
	if !strings.Contains(output, baseViewMarker(m)) {
		t.Error("expected base view content visible behind detail view overlay")
	}
}

// TestT1_EnvPopupIsOverlay verifies env popup renders as overlay.
func TestT1_EnvPopupIsOverlay(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 30
	m.activeModal = ModalEnvironment

	output := m.View()

	if !strings.Contains(output, "Select Environment") {
		t.Error("expected env popup content in output")
	}
	if !strings.Contains(output, baseViewMarker(m)) {
		t.Error("expected base view content visible behind env popup overlay")
	}
}
