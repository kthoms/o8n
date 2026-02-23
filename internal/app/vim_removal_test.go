package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// TestT3_JKeyNoLongerNavigates verifies that pressing 'j' does NOT move the
// table cursor when not in popup mode.
func TestT3_JKeyNoLongerNavigates(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalNone

	// Initial cursor position
	initial := m.table.Cursor()

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m3 := m2.(model)

	// Cursor should not have moved via j (j is now just text input if popup is open,
	// or ignored when no popup is active — no more vim navigation)
	if m3.table.Cursor() != initial {
		t.Errorf("expected table cursor unchanged after 'j' key (removed vim nav), got cursor=%d", m3.table.Cursor())
	}
}

// TestT3_KKeyNoLongerNavigates verifies that pressing 'k' does NOT move the
// table cursor when not in popup mode.
func TestT3_KKeyNoLongerNavigates(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalNone

	initial := m.table.Cursor()

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m3 := m2.(model)

	if m3.table.Cursor() != initial {
		t.Errorf("expected table cursor unchanged after 'k' key (removed vim nav), got cursor=%d", m3.table.Cursor())
	}
}

// TestT3_ArrowDownStillNavigates verifies the ↓ arrow key still moves the cursor.
func TestT3_ArrowDownStillNavigates(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	// Load some rows so there's something to navigate
	rows := []table.Row{{"row1"}, {"row2"}, {"row3"}}
	m.table.SetRows(rows)
	m.table.SetCursor(0)

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m3 := m2.(model)

	if m3.table.Cursor() <= 0 {
		t.Errorf("expected table cursor to move down with arrow key, got cursor=%d", m3.table.Cursor())
	}
}
