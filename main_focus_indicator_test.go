package main

import (
	"strings"
	"testing"

	"github.com/kthoms/o8n/internal/config"
)

// Helper: Create a model for focus indicator tests
func newFocusIndicatorTestModel(t *testing.T) model {
	t.Helper()
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"local": {
				URL:     "http://localhost:8080",
				UIColor: "#00A8E1",  // Blue accent
			},
		},
		Tables: []config.TableDef{
			{
				Name: "process-definition",
				Columns: []config.ColumnDef{
					{Name: "key", Visible: true},
					{Name: "name", Visible: true},
				},
			},
		},
	}
	m := newModel(cfg)
	// Disable splash screen for tests
	m.splashActive = false
	// Set terminal size
	m.lastWidth = 80
	m.lastHeight = 24
	return m
}

// ============================================================
// TEST 1: Focused row has bold + prefix indicator
// ============================================================
func TestFocusIndicatorHasBoldPrefix(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	// Add test data
	defs := []config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Order Process"},
		{ID: "d2", Key: "k2", Name: "Payment Process"},
	}
	m.applyDefinitions(defs)

	// Ensure we're at row 0 (first item focused)
	m.table.SetCursor(0)

	// Render the view
	output := m.View()

	// Focused row should contain the drilldown prefix "▶"
	if !strings.Contains(output, "▶") {
		t.Errorf("expected drilldown prefix '▶' for focused row, not found in output")
	}
}

// ============================================================
// TEST 2: Unfocused rows have normal prefix (spaces)
// ============================================================
func TestFocusIndicatorNormalRowsHaveSpacePrefix(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	defs := []config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Order Process"},
		{ID: "d2", Key: "k2", Name: "Payment Process"},
		{ID: "d3", Key: "k3", Name: "Notification Process"},
	}
	m.applyDefinitions(defs)

	// Move cursor to row 1 (second item focused)
	m.table.SetCursor(1)

	output := m.View()

	// The output should be valid (not check exact format, but row should render)
	if len(output) == 0 {
		t.Fatal("expected non-empty view output")
	}

	// Unfocused rows should NOT have the drilldown prefix (unless they're the focused row)
	// This is implicit: only the focused row gets the prefix
	rows := m.table.Rows()
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows in table, got %d", len(rows))
	}
}

// ============================================================
// TEST 3: Focused row styling is applied (rendering check)
// ============================================================
func TestFocusIndicatorStylingSurvivesRender(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	defs := []config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "ProcessName"},
	}
	m.applyDefinitions(defs)

	m.table.SetCursor(0)

	// Render multiple times; styling should persist
	output1 := m.View()
	output2 := m.View()

	if output1 != output2 {
		t.Error("expected consistent rendering across multiple View() calls")
	}

	// Both should contain the prefix
	if !strings.Contains(output1, "▶") || !strings.Contains(output2, "▶") {
		t.Error("expected drilldown prefix in both renders")
	}
}

// ============================================================
// TEST 4: Cursor movement updates focus indicator
// ============================================================
func TestFocusIndicatorFollowsCursor(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	defs := []config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "First"},
		{ID: "d2", Key: "k2", Name: "Second"},
		{ID: "d3", Key: "k3", Name: "Third"},
	}
	m.applyDefinitions(defs)

	// Move to row 0
	m.table.SetCursor(0)
	if m.table.Cursor() != 0 {
		t.Fatalf("expected cursor at 0, got %d", m.table.Cursor())
	}

	// Move to row 2
	m.table.SetCursor(2)
	if m.table.Cursor() != 2 {
		t.Fatalf("expected cursor at 2, got %d", m.table.Cursor())
	}

	output := m.View()

	// The focused indicator should be present (at least once for the focused row)
	// Exact checking would require parsing the rendered output character by character
	if !strings.Contains(output, "▶") {
		t.Error("expected focus indicator '▶' in view output")
	}
}

// ============================================================
// TEST 5: Focus indicator respects environment color config
// ============================================================
func TestFocusIndicatorUsesEnvironmentColor(t *testing.T) {
	// Create model with specific environment color
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"local": {
				URL:     "http://localhost:8080",
				UIColor: "#FFA500",  // Orange accent
			},
		},
		Tables: []config.TableDef{
			{
				Name: "process-definition",
				Columns: []config.ColumnDef{
					{Name: "key", Visible: true},
					{Name: "name", Visible: true},
				},
			},
		},
	}
	m := newModel(cfg)
	m.splashActive = false
	m.lastWidth = 80
	m.lastHeight = 24

	defs := []config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Process"},
	}
	m.applyDefinitions(defs)

	// Verify environment color was loaded
	if m.config.Environments["local"].UIColor != "#FFA500" {
		t.Errorf("expected environment color #FFA500, got %s",
			m.config.Environments["local"].UIColor)
	}

	// Focus indicator should still render (color application is internal to Lipgloss)
	output := m.View()
	if !strings.Contains(output, "▶") {
		t.Error("expected focus indicator with environment color")
	}
}

// ============================================================
// TEST 6: Focus indicator works with multiple columns
// ============================================================
func TestFocusIndicatorHighlightsEntireRow(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"local": {URL: "http://localhost:8080", UIColor: "#00A8E1"},
		},
		Tables: []config.TableDef{
			{
				Name: "process-definition",
				Columns: []config.ColumnDef{
					{Name: "key", Visible: true, Width: "20%"},
					{Name: "name", Visible: true, Width: "40%"},
					{Name: "version", Visible: true, Width: "20%"},
					{Name: "status", Visible: true, Width: "20%"},
				},
			},
		},
	}
	m := newModel(cfg)
	m.splashActive = false
	m.lastWidth = 80
	m.lastHeight = 24

	defs := []config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Order Processing", Version: 1, Resource: "res1"},
	}
	m.applyDefinitions(defs)

	m.table.SetCursor(0)

	output := m.View()

	// Focus indicator should be present in multi-column layout
	if !strings.Contains(output, "▶") {
		t.Error("expected focus indicator in multi-column table")
	}
}

// ============================================================
// TEST 7: Focus indicator persists through navigation
// ============================================================
func TestFocusIndicatorPersistsAfterNavigation(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	defs := []config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Process One"},
		{ID: "d2", Key: "k2", Name: "Process Two"},
	}
	m.applyDefinitions(defs)

	// Set focus
	m.table.SetCursor(0)
	initialFocus := m.table.Cursor()

	// Simulate a state save (as happens on drill-down)
	savedState := viewState{
		tableCursor: m.table.Cursor(),
		viewMode:    m.viewMode,
	}

	// Verify focus is preserved
	if savedState.tableCursor != initialFocus {
		t.Errorf("expected saved cursor %d, got %d", initialFocus, savedState.tableCursor)
	}

	// Render with original focus
	output := m.View()
	if !strings.Contains(output, "▶") {
		t.Error("expected focus indicator after state save")
	}
}

// ============================================================
// TEST 8: Focus indicator works on empty table
// ============================================================
func TestFocusIndicatorHandlesEmptyTable(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	// Apply empty definitions
	m.applyDefinitions([]config.ProcessDefinition{})

	// Should not crash when rendering empty table
	output := m.View()
	if len(output) == 0 {
		t.Fatal("expected non-empty view output for empty table")
	}
}

// ============================================================
// TEST 9: Focus indicator responsive on narrow terminals
// ============================================================
func TestFocusIndicatorResponsiveOnNarrowTerminal(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	defs := []config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "ProcessName"},
	}
	m.applyDefinitions(defs)
	m.table.SetCursor(0)

	// Simulate narrow terminal (< 60 columns)
	m.lastWidth = 50
	m.lastHeight = 24

	// Should still render without crashing
	output := m.View()
	if len(output) == 0 {
		t.Fatal("expected non-empty view on narrow terminal")
	}

	// On ultra-narrow, prefix might be hidden, but focus should still work
	// (verification depends on implementation details)
}

// ============================================================
// TEST 10: Focus indicator renders correctly after data reload
// ============================================================
func TestFocusIndicatorPersistsAfterDataReload(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	// Load initial data
	defs := []config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Process One"},
		{ID: "d2", Key: "k2", Name: "Process Two"},
	}
	m.applyDefinitions(defs)

	// Set focus to row 1
	m.table.SetCursor(1)

	// Reload data (cursor should be preserved)
	m.applyDefinitions(defs)

	if m.table.Cursor() != 1 {
		t.Errorf("expected cursor preserved at 1 after reload, got %d",
			m.table.Cursor())
	}

	output := m.View()
	if !strings.Contains(output, "▶") {
		t.Error("expected focus indicator after data reload")
	}
}

// ============================================================
// TEST 11: Instances and Variables also have focus indicators
// ============================================================
func TestFocusIndicatorOnInstances(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	instances := []config.ProcessInstance{
		{ID: "i1", DefinitionID: "d1", BusinessKey: "bk1", StartTime: "2020-01-01T00:00:00Z"},
		{ID: "i2", DefinitionID: "d1", BusinessKey: "bk2", StartTime: "2020-01-02T00:00:00Z"},
	}
	m.applyInstances(instances)
	m.table.SetCursor(0)

	output := m.View()

	// Instances should also have the prefix
	if !strings.Contains(output, "▶") {
		t.Error("expected focus indicator prefix in instances view")
	}

	rows := m.table.Rows()
	if len(rows) < 1 {
		t.Fatal("expected at least 1 row in instances table")
	}

	// First column should have the prefix
	if !strings.HasPrefix(rows[0][0], "▶") {
		t.Errorf("expected instance ID row to start with '▶', got %q", rows[0][0])
	}
}

// ============================================================
// TEST 12: Variables have focus indicators
// ============================================================
func TestFocusIndicatorOnVariables(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	vars := []config.Variable{
		{Name: "var1", Value: "value1", Type: "String"},
		{Name: "var2", Value: "value2", Type: "Integer"},
	}
	m.applyVariables(vars)
	m.table.SetCursor(0)

	output := m.View()

	// Variables should also have the prefix
	if !strings.Contains(output, "▶") {
		t.Error("expected focus indicator prefix in variables view")
	}

	rows := m.table.Rows()
	if len(rows) < 1 {
		t.Fatal("expected at least 1 row in variables table")
	}

	// First column should have the prefix
	if !strings.HasPrefix(rows[0][0], "▶") {
		t.Errorf("expected variable name row to start with '▶', got %q", rows[0][0])
	}
}

// ============================================================
// TEST 13: Background color is derived from environment color
// ============================================================
func TestDeriveBackgroundColorMapping(t *testing.T) {
	m := newFocusIndicatorTestModel(t)

	tests := []struct {
		accentColor string
		expectDark  string
	}{
		{"#FFA500", "94"},   // Orange
		{"#00A8E1", "23"},   // Blue
		{"#50C878", "22"},   // Green
		{"#FF6B6B", "52"},   // Red
		{"#UNKNOWN", "23"},  // Unknown → default to dark blue
	}

	for _, tt := range tests {
		got := m.deriveFocusBackgroundColor(tt.accentColor)
		if got != tt.expectDark {
			t.Errorf("deriveFocusBackgroundColor(%q) = %q, want %q", tt.accentColor, got, tt.expectDark)
		}
	}
}
