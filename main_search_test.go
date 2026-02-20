package main

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestSearchModeActivation(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetRows([]table.Row{{"a", "b"}, {"c", "d"}})

	m2, _ := sendKeyString(m, "/")

	if !m2.searchMode {
		t.Error("expected searchMode to be true after pressing /")
	}
	if m2.originalRows == nil {
		t.Error("expected originalRows to be saved")
	}
	if len(m2.originalRows) != 2 {
		t.Errorf("expected 2 original rows, got %d", len(m2.originalRows))
	}
}

func TestSearchFiltersRows(t *testing.T) {
	rows := []table.Row{
		{"proc-1", "OrderProcess", "active"},
		{"proc-2", "PaymentFlow", "suspended"},
		{"proc-3", "OrderCancel", "active"},
	}
	filtered := filterRows(rows, "order")

	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered rows for 'order', got %d", len(filtered))
	}
}

func TestSearchEscRestoresRows(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	original := []table.Row{{"a", "b"}, {"c", "d"}, {"e", "f"}}
	m.table.SetRows(original)

	// Enter search mode
	m2, _ := sendKeyString(m, "/")
	if !m2.searchMode {
		t.Fatal("expected search mode")
	}

	// Esc should restore original rows
	m3, _ := sendKeyString(m2, "esc")

	if m3.searchMode {
		t.Error("expected searchMode to be false after Esc")
	}
	if len(m3.table.Rows()) != 3 {
		t.Errorf("expected 3 rows after Esc, got %d", len(m3.table.Rows()))
	}
	if m3.searchTerm != "" {
		t.Errorf("expected empty searchTerm after Esc, got %q", m3.searchTerm)
	}
}

func TestSearchEnterLocksFilter(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.searchMode = true
	m.originalRows = []table.Row{{"a", "b"}, {"c", "d"}}
	m.searchInput.SetValue("test")
	m.searchTerm = "test"

	m2, _ := sendKeyString(m, "enter")

	if m2.searchMode {
		t.Error("expected searchMode to be false after Enter")
	}
	if m2.searchTerm != "test" {
		t.Errorf("expected searchTerm 'test', got %q", m2.searchTerm)
	}
	if m2.originalRows != nil {
		t.Error("expected originalRows to be nil after Enter (filter locked)")
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	rows := []table.Row{
		{"proc-1", "OrderProcess", "active"},
		{"proc-2", "PaymentFlow", "suspended"},
	}

	filtered := filterRows(rows, "ORDER")
	if len(filtered) != 1 {
		t.Errorf("expected 1 row for 'ORDER' (case insensitive), got %d", len(filtered))
	}

	filtered2 := filterRows(rows, "payment")
	if len(filtered2) != 1 {
		t.Errorf("expected 1 row for 'payment' (case insensitive), got %d", len(filtered2))
	}
}

func TestSearchEmptyResult(t *testing.T) {
	rows := []table.Row{
		{"proc-1", "OrderProcess", "active"},
	}
	filtered := filterRows(rows, "nonexistent")
	if len(filtered) != 0 {
		t.Errorf("expected 0 rows for 'nonexistent', got %d", len(filtered))
	}
}

func TestSearchAcrossAllColumns(t *testing.T) {
	rows := []table.Row{
		{"id-1", "name1", "status-running"},
		{"id-2", "name2", "status-stopped"},
		{"id-3", "match-name", "status-running"},
	}

	// Match in column 0
	filtered := filterRows(rows, "id-1")
	if len(filtered) != 1 {
		t.Errorf("expected 1 match for 'id-1' in column 0, got %d", len(filtered))
	}

	// Match in column 2
	filtered2 := filterRows(rows, "stopped")
	if len(filtered2) != 1 {
		t.Errorf("expected 1 match for 'stopped' in column 2, got %d", len(filtered2))
	}

	// Match in column 1
	filtered3 := filterRows(rows, "match-name")
	if len(filtered3) != 1 {
		t.Errorf("expected 1 match for 'match-name' in column 1, got %d", len(filtered3))
	}
}

func TestSearchEmptyTerm(t *testing.T) {
	rows := []table.Row{
		{"a", "b"},
		{"c", "d"},
	}
	filtered := filterRows(rows, "")
	if len(filtered) != 2 {
		t.Errorf("expected all rows returned for empty term, got %d", len(filtered))
	}
}

func TestSearchNotActiveInRootPopup(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.showRootPopup = true
	m.rootInput = ""

	m2, _ := sendKeyString(m, "/")

	if m2.searchMode {
		t.Error("expected searchMode to remain false when root popup is active")
	}
	if m2.rootInput != "/" {
		t.Errorf("expected rootInput '/', got %q", m2.rootInput)
	}
}
