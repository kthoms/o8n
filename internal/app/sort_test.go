package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestSortPopupOpens(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	m2, _ := sendKeyString(m, "s")

	if m2.activeModal != ModalSort {
		t.Errorf("expected ModalSort after 's', got %v", m2.activeModal)
	}
}

func TestSortPopupEscCloses(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalSort

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Errorf("expected ModalNone after Esc, got %v", m2.activeModal)
	}
}

func TestSortByColumn(t *testing.T) {
	rows := []table.Row{
		{"charlie", "3"},
		{"alpha", "1"},
		{"bravo", "2"},
	}

	sorted := sortTableRows(rows, 0, true)
	if sorted[0][0] != "alpha" {
		t.Errorf("expected first row 'alpha' after asc sort, got %q", sorted[0][0])
	}
	if sorted[2][0] != "charlie" {
		t.Errorf("expected last row 'charlie' after asc sort, got %q", sorted[2][0])
	}
}

func TestSortToggleDirection(t *testing.T) {
	rows := []table.Row{
		{"alpha", "1"},
		{"bravo", "2"},
		{"charlie", "3"},
	}

	sorted := sortTableRows(rows, 0, false)
	if sorted[0][0] != "charlie" {
		t.Errorf("expected first row 'charlie' after desc sort, got %q", sorted[0][0])
	}
	if sorted[2][0] != "alpha" {
		t.Errorf("expected last row 'alpha' after desc sort, got %q", sorted[2][0])
	}
}

func TestSortNumeric(t *testing.T) {
	rows := []table.Row{
		{"b", "10"},
		{"a", "2"},
		{"c", "100"},
	}

	sorted := sortTableRows(rows, 1, true)
	if sorted[0][1] != "2" {
		t.Errorf("expected first row value '2' after numeric asc sort, got %q", sorted[0][1])
	}
	if sorted[1][1] != "10" {
		t.Errorf("expected second row value '10' after numeric asc sort, got %q", sorted[1][1])
	}
	if sorted[2][1] != "100" {
		t.Errorf("expected third row value '100' after numeric asc sort, got %q", sorted[2][1])
	}
}

func TestSortDate(t *testing.T) {
	rows := []table.Row{
		{"b", "2024-03-15"},
		{"a", "2024-01-01"},
		{"c", "2024-12-31"},
	}

	sorted := sortTableRows(rows, 1, true)
	if sorted[0][1] != "2024-01-01" {
		t.Errorf("expected first row '2024-01-01' after date asc sort, got %q", sorted[0][1])
	}
	if sorted[2][1] != "2024-12-31" {
		t.Errorf("expected last row '2024-12-31' after date asc sort, got %q", sorted[2][1])
	}
}

func TestSortInvalidColumn(t *testing.T) {
	rows := []table.Row{{"a"}, {"b"}}
	sorted := sortTableRows(rows, -1, true)
	if len(sorted) != 2 {
		t.Errorf("expected 2 rows for invalid column, got %d", len(sorted))
	}
}

func TestSortPopupNavigation(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalSort
	m.sortPopupCursor = 0
	cols := []table.Column{{Title: "A", Width: 10}, {Title: "B", Width: 10}, {Title: "C", Width: 10}}
	m.table.SetColumns(cols)

	// Move down
	m2, _ := sendKeyString(m, "down")
	if m2.sortPopupCursor != 1 {
		t.Errorf("expected sortPopupCursor 1 after down, got %d", m2.sortPopupCursor)
	}

	// Move up
	m3, _ := sendKeyString(m2, "up")
	if m3.sortPopupCursor != 0 {
		t.Errorf("expected sortPopupCursor 0 after up, got %d", m3.sortPopupCursor)
	}
}
