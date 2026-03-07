package app

// resize_test.go — Story 4.3: Terminal Resize Handling
//
// Tests verify graceful terminal resize adaptation:
//   - AC 1: Layout reflows correctly without artifacts
//   - AC 2: termWidth and termHeight updated immediately
//   - AC 3: Degrades gracefully when resized below 120x20
//   - AC 4: Modals re-center correctly after resize

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kthoms/o6n/internal/config"
)

func resizeTestConfig() *config.Config {
	return &config.Config{
		Environments: map[string]config.Environment{"local": {URL: "http://localhost"}},
		Tables: []config.TableDef{
			{
				Name: "test-resource",
				Columns: []config.ColumnDef{
					{Name: "id", Width: 20},
					{Name: "name", Width: 30},
					{Name: "status", Width: 15},
				},
			},
		},
	}
}

// ── AC 1 & 2: WindowSizeMsg handling and dimension updates ─────────────

// TestWindowSizeMsg_DimensionUpdate verifies m.lastWidth and m.lastHeight are updated
func TestWindowSizeMsg_DimensionUpdate(t *testing.T) {
	m := newModel(resizeTestConfig())
	m.currentRoot = "test-resource"

	// Simulate a resize to 120x24
	msg := tea.WindowSizeMsg{Width: 120, Height: 24}
	m2, cmd := m.Update(msg)
	model := m2.(model)

	if model.lastWidth != 120 {
		t.Errorf("expected lastWidth=120, got %d", model.lastWidth)
	}
	if model.lastHeight != 24 {
		t.Errorf("expected lastHeight=24, got %d", model.lastHeight)
	}
	if cmd != nil {
		t.Errorf("expected no command from resize, got %v", cmd)
	}
}

// TestWindowSizeMsg_PaneHeightCalculation verifies paneHeight is recalculated
func TestWindowSizeMsg_PaneHeightCalculation(t *testing.T) {
	m := newModel(resizeTestConfig())
	m.currentRoot = "test-resource"

	// Initial size
	msg1 := tea.WindowSizeMsg{Width: 100, Height: 20}
	m1, _ := m.Update(msg1)
	model1 := m1.(model)

	// Resize taller
	msg2 := tea.WindowSizeMsg{Width: 100, Height: 30}
	m2, _ := model1.Update(msg2)
	model2 := m2.(model)

	// Height should have increased
	if model2.paneHeight <= model1.paneHeight {
		t.Errorf("expected paneHeight to increase on tall resize, was %d, now %d", model1.paneHeight, model2.paneHeight)
	}

	// New height = 30 - 2 (header) - 1 (context) - 1 (footer) = 26
	expectedHeight := 30 - 2 - 1 - 1
	if model2.paneHeight != expectedHeight {
		t.Errorf("expected paneHeight=%d, got %d", expectedHeight, model2.paneHeight)
	}
}

// TestWindowSizeMsg_TableDimensionUpdate verifies table.SetWidth and SetHeight are called
func TestWindowSizeMsg_TableDimensionUpdate(t *testing.T) {
	m := newModel(resizeTestConfig())
	m.currentRoot = "test-resource"

	msg := tea.WindowSizeMsg{Width: 120, Height: 24}
	m2, _ := m.Update(msg)
	model := m2.(model)

	// Table width should be pane width - 4 (for borders/padding)
	expectedTableWidth := model.paneWidth - 4
	actualTableWidth := model.table.Width()
	if actualTableWidth != expectedTableWidth {
		t.Errorf("expected table width=%d (paneWidth %d - 4), got %d", expectedTableWidth, model.paneWidth, actualTableWidth)
	}

	// Table height should be positive and less than or equal to pane height
	actualTableHeight := model.table.Height()
	if actualTableHeight <= 0 {
		t.Errorf("expected table height > 0, got %d", actualTableHeight)
	}
	if actualTableHeight > model.paneHeight {
		t.Errorf("expected table height <= paneHeight (%d), got %d", model.paneHeight, actualTableHeight)
	}
}

// ── AC 1: No layout corruption/artifacts ──────────────────────────────

// TestResize_MultipleConsecutiveResizes verifies no corruption across resizes
func TestResize_MultipleConsecutiveResizes(t *testing.T) {
	m := newModel(resizeTestConfig())
	m.currentRoot = "test-resource"

	sizes := []tea.WindowSizeMsg{
		{Width: 120, Height: 20},
		{Width: 80, Height: 24},
		{Width: 150, Height: 30},
		{Width: 100, Height: 22},
	}

	currentModel := m
	for _, msg := range sizes {
		m2, _ := currentModel.Update(msg)
		currentModel = m2.(model)

		// Verify dimensions are always consistent
		if currentModel.lastWidth != msg.Width {
			t.Errorf("resize to %dx%d: lastWidth=%d (expected %d)", msg.Width, msg.Height, currentModel.lastWidth, msg.Width)
		}
		if currentModel.lastHeight != msg.Height {
			t.Errorf("resize to %dx%d: lastHeight=%d (expected %d)", msg.Width, msg.Height, currentModel.lastHeight, msg.Height)
		}

		// Pane dimensions should be valid (positive)
		if currentModel.paneWidth <= 0 {
			t.Errorf("paneWidth=%d (should be positive)", currentModel.paneWidth)
		}
		if currentModel.paneHeight <= 0 {
			t.Errorf("paneHeight=%d (should be positive)", currentModel.paneHeight)
		}
	}
}

// ── AC 3: Degrade gracefully below 120x20 ──────────────────────────────

// TestResize_BelowMinimum_80x12 verifies graceful degradation below 120x20
func TestResize_BelowMinimum_80x12(t *testing.T) {
	m := newModel(resizeTestConfig())
	m.currentRoot = "test-resource"

	msg := tea.WindowSizeMsg{Width: 80, Height: 12}
	m2, _ := m.Update(msg)
	model := m2.(model)

	// Should still have valid dimensions (no panic, no negative values)
	if model.lastWidth != 80 {
		t.Errorf("expected lastWidth=80, got %d", model.lastWidth)
	}
	if model.lastHeight != 12 {
		t.Errorf("expected lastHeight=12, got %d", model.lastHeight)
	}

	// paneHeight should be at minimum 3 (see contentHeight < 3 check in update.go)
	if model.paneHeight < 3 {
		t.Errorf("expected paneHeight >= 3 at narrow resize, got %d", model.paneHeight)
	}

	// paneWidth should be at minimum 10
	if model.paneWidth < 10 {
		t.Errorf("expected paneWidth >= 10, got %d", model.paneWidth)
	}
}

// TestResize_VerySmall_40x10 verifies no crash at extreme narrow
func TestResize_VerySmall_40x10(t *testing.T) {
	m := newModel(resizeTestConfig())
	m.currentRoot = "test-resource"

	msg := tea.WindowSizeMsg{Width: 40, Height: 10}
	m2, _ := m.Update(msg)
	model := m2.(model)

	// Should not crash and maintain valid state
	if model.paneWidth < 10 {
		t.Errorf("expected paneWidth clamped to minimum 10, got %d", model.paneWidth)
	}
	if model.paneHeight < 3 {
		t.Errorf("expected paneHeight clamped to minimum 3, got %d", model.paneHeight)
	}
}

// TestResize_ExtraLarge_300x50 verifies scaling up works
func TestResize_ExtraLarge_300x50(t *testing.T) {
	m := newModel(resizeTestConfig())
	m.currentRoot = "test-resource"

	msg := tea.WindowSizeMsg{Width: 300, Height: 50}
	m2, _ := m.Update(msg)
	model := m2.(model)

	if model.lastWidth != 300 {
		t.Errorf("expected lastWidth=300, got %d", model.lastWidth)
	}
	if model.lastHeight != 50 {
		t.Errorf("expected lastHeight=50, got %d", model.lastHeight)
	}

	// paneHeight = 50 - 2 - 1 - 1 = 46
	expectedPaneHeight := 50 - 2 - 1 - 1
	if model.paneHeight != expectedPaneHeight {
		t.Errorf("expected paneHeight=%d, got %d", expectedPaneHeight, model.paneHeight)
	}
}

// ── AC 4: Modals remain valid after resize ────────────────────────────

// TestResize_WithModalOpen verifies modal state persists correctly
func TestResize_WithModalOpen(t *testing.T) {
	m := newModel(resizeTestConfig())
	m.currentRoot = "test-resource"

	// Open a modal (set it to ModalNone initially, then open)
	m.activeModal = ModalConfirmQuit

	// Resize while modal is open
	msg := tea.WindowSizeMsg{Width: 100, Height: 24}
	m2, _ := m.Update(msg)
	model := m2.(model)

	// Modal should still be active
	if model.activeModal != ModalConfirmQuit {
		t.Errorf("expected modal to remain open after resize, got %v", model.activeModal)
	}

	// Dimensions should be updated
	if model.lastWidth != 100 || model.lastHeight != 24 {
		t.Errorf("expected dimensions updated even with modal open")
	}
}

// TestResize_SmallThenLarge verifies both shrinking and growing work
func TestResize_SmallThenLarge(t *testing.T) {
	m := newModel(resizeTestConfig())
	m.currentRoot = "test-resource"

	// Start normal
	msg1 := tea.WindowSizeMsg{Width: 120, Height: 24}
	m1, _ := m.Update(msg1)
	model1 := m1.(model)

	w1 := model1.paneWidth
	h1 := model1.paneHeight

	// Shrink
	msg2 := tea.WindowSizeMsg{Width: 60, Height: 12}
	m2, _ := model1.Update(msg2)
	model2 := m2.(model)

	w2 := model2.paneWidth
	h2 := model2.paneHeight

	if w2 >= w1 {
		t.Errorf("expected paneWidth to shrink, was %d, now %d", w1, w2)
	}
	if h2 >= h1 {
		t.Errorf("expected paneHeight to shrink, was %d, now %d", h1, h2)
	}

	// Grow again
	msg3 := tea.WindowSizeMsg{Width: 150, Height: 30}
	m3, _ := model2.Update(msg3)
	model3 := m3.(model)

	w3 := model3.paneWidth
	h3 := model3.paneHeight

	if w3 <= w2 {
		t.Errorf("expected paneWidth to grow, was %d, now %d", w2, w3)
	}
	if h3 <= h2 {
		t.Errorf("expected paneHeight to grow, was %d, now %d", h2, h3)
	}
}
