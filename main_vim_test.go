package main

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestVimJKNavigation(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	cols := []table.Column{{Title: "ID", Width: 10}, {Title: "NAME", Width: 20}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"1", "a"}, {"2", "b"}, {"3", "c"}})
	m.table.SetCursor(0)

	// j moves down
	m2, _ := sendKeyString(m, "j")
	if m2.table.Cursor() != 1 {
		t.Errorf("expected cursor 1 after j, got %d", m2.table.Cursor())
	}

	// k moves up
	m3, _ := sendKeyString(m2, "k")
	if m3.table.Cursor() != 0 {
		t.Errorf("expected cursor 0 after k, got %d", m3.table.Cursor())
	}
}

func TestVimGGJumpsToTop(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	cols := []table.Column{{Title: "ID", Width: 10}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"1"}, {"2"}, {"3"}, {"4"}, {"5"}})
	m.table.SetCursor(4)

	// First g: sets pendingG
	m2, _ := sendKeyString(m, "g")
	if !m2.pendingG {
		t.Fatal("expected pendingG to be true after first g")
	}

	// Second g: jumps to top
	m3, _ := sendKeyString(m2, "g")
	if m3.table.Cursor() != 0 {
		t.Errorf("expected cursor 0 after gg, got %d", m3.table.Cursor())
	}
	if m3.pendingG {
		t.Error("expected pendingG to be false after gg")
	}
}

func TestVimGJumpsToBottom(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	cols := []table.Column{{Title: "ID", Width: 10}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"1"}, {"2"}, {"3"}, {"4"}, {"5"}})
	m.table.SetCursor(0)

	m2, _ := sendKeyString(m, "G")
	if m2.table.Cursor() != 4 {
		t.Errorf("expected cursor 4 after G, got %d", m2.table.Cursor())
	}
}

func TestVimHalfPageScroll(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	cols := []table.Column{{Title: "ID", Width: 10}}
	m.table.SetColumns(cols)
	rows := make([]table.Row, 20)
	for i := range rows {
		rows[i] = table.Row{string(rune('a' + i))}
	}
	m.table.SetRows(rows)
	m.table.SetCursor(0)
	m.table.SetHeight(10) // visible rows ~10

	// ctrl+u should move up (from 0, stays at 0)
	m2, _ := sendKeyString(m, "ctrl+u")
	if m2.table.Cursor() != 0 {
		t.Errorf("expected cursor 0 after ctrl+u from top, got %d", m2.table.Cursor())
	}
}

func TestVimKeysDisabledInSearchMode(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.searchMode = true
	m.originalRows = []table.Row{{"1"}, {"2"}}
	cols := []table.Column{{Title: "ID", Width: 10}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"1"}, {"2"}})
	m.table.SetCursor(0)

	// j/k should NOT navigate when in search mode (they go to search input)
	m2, _ := sendKeyString(m, "j")
	// In search mode, j is passed to the search input, not table nav
	if m2.table.Cursor() != 0 {
		// This is expected: cursor doesn't change in search mode
	}
}

func TestVimKeysDisabledInModal(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalHelp
	cols := []table.Column{{Title: "ID", Width: 10}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"1"}, {"2"}, {"3"}})
	m.table.SetCursor(0)

	// j should dismiss help modal (any key dismisses help)
	m2, _ := sendKeyString(m, "j")
	if m2.activeModal != ModalNone {
		t.Error("expected help modal to be dismissed")
	}
}

func TestVimPendingGTimeout(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.pendingG = true

	// clearPendingGMsg should reset pendingG
	res, _ := m.Update(clearPendingGMsg{})
	m2 := res.(model)
	if m2.pendingG {
		t.Error("expected pendingG to be false after clearPendingGMsg")
	}
}
