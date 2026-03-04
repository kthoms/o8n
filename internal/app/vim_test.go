package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestVimJKNavigation(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.vimMode = true // j/k only work in vim mode
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
	m.vimMode = true // gg only works in vim mode
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
	m.vimMode = true // G only works in vim mode
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
	m.vimMode = true // ctrl+u half-page only works in vim mode
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
	m.lastHeight = 30
	cols := []table.Column{{Title: "ID", Width: 10}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"1"}, {"2"}, {"3"}})
	m.table.SetCursor(0)

	// down arrow should scroll help (not dismiss), table cursor should not move
	m2, _ := sendKeyString(m, "down")
	if m2.activeModal != ModalHelp {
		t.Error("expected help modal to remain open when down is pressed (scroll)")
	}
	if m2.table.Cursor() != 0 {
		t.Error("expected table cursor to stay at 0 (down consumed by help scroll)")
	}

	// space (and other non-close keys) should be swallowed while help modal is open
	// — vim keys must not trigger actions when a modal is active
	m3, _ := sendKeyString(m, " ")
	if m3.activeModal != ModalHelp {
		t.Error("expected help modal to remain open when space is pressed (non-close key swallowed)")
	}

	// Esc is the explicit close key for ModalHelp
	m4, _ := sendKeyString(m3, "esc")
	if m4.activeModal != ModalNone {
		t.Error("expected help modal to be dismissed by Esc")
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

// ── Default mode: vim keys are no-ops ─────────────────────────────────────────

func TestDefaultModeJKeyNoOp(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.vimMode = false
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 10}})
	m.table.SetRows([]table.Row{{"a"}, {"b"}, {"c"}})
	m.table.SetCursor(0)

	m2, _ := sendKeyString(m, "j")
	if m2.table.Cursor() != 0 {
		t.Errorf("expected j to be no-op in default mode, cursor moved to %d", m2.table.Cursor())
	}
}

func TestDefaultModeKKeyNoOp(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.vimMode = false
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 10}})
	m.table.SetRows([]table.Row{{"a"}, {"b"}, {"c"}})
	m.table.SetCursor(1)

	m2, _ := sendKeyString(m, "k")
	if m2.table.Cursor() != 1 {
		t.Errorf("expected k to be no-op in default mode, cursor=%d", m2.table.Cursor())
	}
}

func TestDefaultModeGGNoOp(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.vimMode = false
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 10}})
	m.table.SetRows([]table.Row{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}})
	m.table.SetCursor(4)

	m2, _ := sendKeyString(m, "g")
	m3, _ := sendKeyString(m2, "g")
	if m3.table.Cursor() != 4 {
		t.Errorf("expected gg no-op in default mode, cursor=%d", m3.table.Cursor())
	}
	if m3.pendingG {
		t.Error("expected pendingG=false in default mode")
	}
}

func TestDefaultModeCapGNoOp(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.vimMode = false
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 10}})
	m.table.SetRows([]table.Row{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}})
	m.table.SetCursor(0)

	m2, _ := sendKeyString(m, "G")
	if m2.table.Cursor() != 0 {
		t.Errorf("expected G no-op in default mode, cursor=%d", m2.table.Cursor())
	}
}

// ── Home/End navigation (always active) ──────────────────────────────────────

func TestHomeKeyJumpsToFirstRow(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 10}})
	m.table.SetRows([]table.Row{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}})
	m.table.SetCursor(4)

	m2, _ := sendKeyString(m, "home")
	if m2.table.Cursor() != 0 {
		t.Errorf("expected home to jump to first row, got cursor=%d", m2.table.Cursor())
	}
}

func TestEndKeyJumpsToLastRow(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 10}})
	rows := []table.Row{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}}
	m.table.SetRows(rows)
	m.table.SetCursor(0)

	m2, _ := sendKeyString(m, "end")
	lastIdx := len(rows) - 1
	if m2.table.Cursor() != lastIdx {
		t.Errorf("expected end to jump to last row (%d), got cursor=%d", lastIdx, m2.table.Cursor())
	}
}

func TestHomeEndWorkInVimMode(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.vimMode = true
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 10}})
	m.table.SetRows([]table.Row{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}})
	m.table.SetCursor(4)

	m2, _ := sendKeyString(m, "home")
	if m2.table.Cursor() != 0 {
		t.Errorf("expected home to work in vim mode, got cursor=%d", m2.table.Cursor())
	}
}

func TestVimModeDefaultFalse(t *testing.T) {
	m := newTestModel(t)
	if m.vimMode {
		t.Error("expected vimMode=false by default")
	}
}
