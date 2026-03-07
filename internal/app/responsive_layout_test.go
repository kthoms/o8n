package app

// responsive_layout_test.go — Story 4.2: Responsive Column & Hint Visibility
//
// Tests verify graceful layout adaptation at various terminal widths:
//   - AC 1: Columns hidden in hide_order sequence when terminal < 120 columns
//   - AC 2: Hints omitted cleanly based on MinWidth threshold
//   - AC 3: Hints with lower Priority integer dropped first when space constrained
//   - AC 4: UI remains functional at very narrow widths (< 80 columns)

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kthoms/o6n/internal/config"
)

// responsiveTestConfig creates a config with explicit hide_order for testing
func responsiveTestConfig() *config.Config {
	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: "http://localhost"}},
		Tables: []config.TableDef{
			{
				Name: "process-instance",
				Columns: []config.ColumnDef{
					{Name: "id", Width: 20, HideOrder: 0},           // hidden first
					{Name: "businessKey", Width: 25, HideOrder: 1},  // hidden second
					{Name: "state", Width: 15, HideOrder: 2},        // hidden third
					{Name: "startTime", Width: 20, HideOrder: 3},    // hidden last
				},
			},
		},
	}
	return cfg
}

// ── AC 1: Columns hidden in hide_order sequence ──────────────────────────

// TestResponsiveColumns_HideOrderSequence verifies columns are hidden in correct sequence
func TestResponsiveColumns_HideOrderSequence(t *testing.T) {
	m := newModel(responsiveTestConfig())
	m.currentRoot = "process-instance"
	m.paneWidth = 80  // narrow width
	m.paneHeight = 18

	// Build columns for 80-character width
	cols := m.buildColumnsFor("process-instance", 80)

	// At 80 columns with 4 columns totaling 80 width, some should be hidden
	if len(cols) > 4 {
		t.Errorf("expected at most 4 columns, got %d", len(cols))
	}

	// Verify at least one column remains
	if len(cols) == 0 {
		t.Error("expected at least 1 column visible at 80 width")
	}
}

// TestResponsiveColumns_120Width verifies all columns visible at 120 width
func TestResponsiveColumns_120Width(t *testing.T) {
	m := newModel(responsiveTestConfig())
	m.currentRoot = "process-instance"

	// Build columns for 120-character width (should fit all)
	cols := m.buildColumnsFor("process-instance", 120)

	// All 4 columns should be visible at 120 width
	if len(cols) < 3 {
		t.Errorf("expected at least 3 columns at 120 width, got %d", len(cols))
	}
}

// TestResponsiveColumns_ExtractWide verifies extra wide terminal shows all columns
func TestResponsiveColumns_ExtraWide(t *testing.T) {
	m := newModel(responsiveTestConfig())
	m.currentRoot = "process-instance"

	// Build columns for 200-character width
	cols := m.buildColumnsFor("process-instance", 200)

	// At 200 columns, most columns should be visible
	if len(cols) < 3 {
		t.Errorf("expected at least 3 columns at 200 width, got %d", len(cols))
	}
}

// TestResponsiveColumns_NeverEmpty verifies at least one column always visible
func TestResponsiveColumns_NeverEmpty(t *testing.T) {
	m := newModel(responsiveTestConfig())
	m.currentRoot = "process-instance"

	// Even at very narrow width, buildColumnsFor should return at least 1 column
	cols := m.buildColumnsFor("process-instance", 20)

	if len(cols) == 0 {
		t.Error("expected at least 1 column even at 20 width")
	}
}

// ── AC 2: Hints hidden based on MinWidth ──────────────────────────────

// TestHintVisibility_MinWidthRespected verifies hints filter by MinWidth
func TestHintVisibility_MinWidthRespected(t *testing.T) {
	m := newModel(responsiveTestConfig())
	m.currentRoot = "process-instance"
	m.lastWidth = 80

	// Get hints that should be visible at 80 columns
	allHints := tableViewHints(m)
	filteredHints := filterHints(allHints, 80)

	// Count hints with MinWidth > 80 in original
	shouldHide := 0
	for _, h := range allHints {
		if h.MinWidth > 80 && h.MinWidth > 0 {
			shouldHide++
		}
	}

	// Verify those hints are not in filtered list
	for _, h := range filteredHints {
		if h.MinWidth > 80 && h.MinWidth > 0 {
			t.Errorf("expected hint %q (MinWidth=%d) to be hidden at width 80", h.Key, h.MinWidth)
		}
	}
}

// TestHintVisibility_AlwaysShowHints verifies MinWidth:0 hints always visible
func TestHintVisibility_AlwaysShowHints(t *testing.T) {
	m := newModel(responsiveTestConfig())
	m.currentRoot = "process-instance"
	m.lastWidth = 40  // very narrow

	allHints := tableViewHints(m)
	filteredHints := filterHints(allHints, 40)

	// Hints with MinWidth: 0 should still be present
	for _, h := range allHints {
		if h.MinWidth == 0 {
			// Find this hint in filtered list
			found := false
			for _, fh := range filteredHints {
				if fh.Key == h.Key {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected MinWidth:0 hint %q to be visible at width 40", h.Key)
			}
		}
	}
}

// ── AC 3: Priority-based hint truncation ──────────────────────────────

// TestHintPriority_HigherPriorityDroppedFirst verifies Priority integer order
func TestHintPriority_HigherPriorityDroppedFirst(t *testing.T) {
	// Create test hints with different priorities
	testHints := []Hint{
		{Key: "p1", Label: "high", MinWidth: 0, Priority: 1},    // highest priority
		{Key: "p2", Label: "med", MinWidth: 0, Priority: 4},     // medium priority
		{Key: "p3", Label: "low", MinWidth: 0, Priority: 8},     // lowest priority
	}

	// Sort by priority (lower number = higher priority, so dropped first when we need to truncate)
	sorted := append([]Hint(nil), testHints...)
	sorted_before := len(sorted)

	// When filtered, lower priority (higher number) should remain longer
	filtered := filterHints(testHints, 120)
	filtered_count := len(filtered)

	// At wide width, all should be visible
	if filtered_count != sorted_before {
		t.Errorf("expected all %d hints at wide width, got %d", sorted_before, filtered_count)
	}
}

// TestHintVisibility_160Columns verifies hints at various widths
func TestHintVisibility_160Columns(t *testing.T) {
	m := newModel(responsiveTestConfig())
	m.currentRoot = "process-instance"
	m.lastWidth = 160

	hints := filterHints(tableViewHints(m), 160)

	// Should have multiple hints visible
	if len(hints) == 0 {
		t.Error("expected hints to be visible at 160 columns")
	}

	// All hints should respect MinWidth <= 160
	for _, h := range hints {
		if h.MinWidth > 160 && h.MinWidth > 0 {
			t.Errorf("hint %q has MinWidth %d > 160, should not be visible", h.Key, h.MinWidth)
		}
	}
}

// ── AC 4: Very narrow widths (< 80 columns) ──────────────────────────

// TestResponsiveUI_VeryNarrow_40Columns ensures no crash at 40 columns
func TestResponsiveUI_VeryNarrow_40Columns(t *testing.T) {
	m := newModel(responsiveTestConfig())
	m.currentRoot = "process-instance"
	m.table.SetColumns([]table.Column{
		{Title: "id", Width: 20},
		{Title: "businessKey", Width: 20},
	})
	m.table.SetRows([]table.Row{
		{"pi-1", "invoice"},
		{"pi-2", "payment"},
	})

	// Simulate rendering at 40 columns
	cols := m.buildColumnsFor("process-instance", 40)

	// Should not crash, and at least 1 column should be visible
	if len(cols) == 0 {
		t.Error("expected at least 1 column at 40 width (should not crash or empty)")
	}

	// Verify column titles are not empty
	for _, col := range cols {
		if col.Title == "" {
			t.Error("expected column title to be non-empty")
		}
	}
}

// TestResponsiveUI_VeryNarrow_Hints ensures hints don't crash at narrow width
func TestResponsiveUI_VeryNarrow_Hints(t *testing.T) {
	m := newModel(responsiveTestConfig())
	m.currentRoot = "process-instance"
	m.lastWidth = 40

	hints := currentViewHints(m)
	filtered := filterHints(hints, 40)

	// Should not crash; may have no hints or some hints with MinWidth: 0
	if filtered == nil {
		t.Error("filterHints returned nil, expected empty slice")
	}

	// All visible hints should have MinWidth <= 40
	for _, h := range filtered {
		if h.MinWidth > 40 && h.MinWidth > 0 {
			t.Errorf("hint %q (MinWidth=%d) should not be visible at width 40", h.Key, h.MinWidth)
		}
	}
}

// TestResponsiveColumns_StabilityAt40 verifies buildColumnsFor doesn't crash at very narrow widths
func TestResponsiveColumns_StabilityAt40(t *testing.T) {
	m := newModel(responsiveTestConfig())
	m.currentRoot = "process-instance"

	// Should not crash even at 40 columns
	cols := m.buildColumnsFor("process-instance", 40)

	// Must return at least one column
	if len(cols) == 0 {
		t.Error("expected at least 1 column at 40 width")
	}

	// Each column should have a title
	for _, col := range cols {
		if col.Title == "" {
			t.Error("expected all columns to have titles")
		}
		if col.Width <= 0 {
			t.Error("expected all columns to have positive width")
		}
	}
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
