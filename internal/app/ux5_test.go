package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kthoms/o8n/internal/config"
)

// ── T1: terminatedMsg success feedback + rowData cleanup ─────────────────────

func TestTerminatedMsgShowsSuccessFeedback(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 20}, {Title: "NAME", Width: 30}})
	m.table.SetRows([]table.Row{{"▶ abc", "Process A"}, {"▶ xyz", "Process B"}})
	m.rowData = []map[string]interface{}{{"id": "abc"}, {"id": "xyz"}}

	res, _ := m.Update(terminatedMsg{id: "abc"})
	got := res.(model)

	if !strings.Contains(got.footerError, "Terminated") {
		t.Errorf("expected footerError to contain 'Terminated', got %q", got.footerError)
	}
	if !strings.Contains(got.footerError, "abc") {
		t.Errorf("expected footerError to contain 'abc', got %q", got.footerError)
	}
}

func TestTerminatedMsgCleansRowData(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 20}, {Title: "NAME", Width: 30}})
	m.table.SetRows([]table.Row{{"▶ abc", "Process A"}, {"▶ xyz", "Process B"}})
	m.rowData = []map[string]interface{}{{"id": "abc"}, {"id": "xyz"}}

	res, _ := m.Update(terminatedMsg{id: "abc"})
	got := res.(model)

	if len(got.rowData) != 1 {
		t.Errorf("expected rowData length 1 after termination, got %d", len(got.rowData))
	}
}

// ── T2: s key blocked during searchMode ──────────────────────────────────────

func TestSortKeyDuringSearchShowsHint(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.searchMode = true

	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	got := res.(model)

	// Sort modal must not open while search is active.
	if got.activeModal == ModalSort {
		t.Error("expected ModalSort NOT to open when searchMode is true")
	}
}

// ── T3: filterRows strips ANSI codes before matching ─────────────────────────

func TestFilterRowsStripsANSI(t *testing.T) {
	rows := []table.Row{
		{"\x1b[32mACTIVE\x1b[0m", "instance-1"},
		{"\x1b[31mFAILED\x1b[0m", "instance-2"},
	}

	matched := filterRows(rows, "active")
	if len(matched) != 1 {
		t.Errorf("expected 1 match for 'active' after ANSI strip, got %d", len(matched))
	}
}

func TestFilterRowsDoesNotMatchANSICodes(t *testing.T) {
	rows := []table.Row{
		{"\x1b[32mACTIVE\x1b[0m", "instance-1"},
	}

	matched := filterRows(rows, "[32m")
	if len(matched) != 0 {
		t.Errorf("expected 0 matches for raw ANSI escape '[32m', got %d", len(matched))
	}
}

// ── T4: suspendedMsg/resumedMsg/actionExecutedMsg trigger reload ──────────────

func TestSuspendedMsgTriggersReload(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	_, cmd := m.Update(suspendedMsg{id: "x"})
	if cmd == nil {
		t.Error("expected non-nil command (reload) after suspendedMsg")
	}
}

func TestResumedMsgTriggersReload(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	_, cmd := m.Update(resumedMsg{id: "x"})
	if cmd == nil {
		t.Error("expected non-nil command (reload) after resumedMsg")
	}
}

func TestActionExecutedMsgTriggersReload(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	_, cmd := m.Update(actionExecutedMsg{label: "doSomething"})
	if cmd == nil {
		t.Error("expected non-nil command (reload) after actionExecutedMsg")
	}
}

// ── T5: editSavedMsg with dataKey syncs rowData ───────────────────────────────

func TestEditSavedMsgUpdatesRowData(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetColumns([]table.Column{{Title: "VALUE", Width: 20}})
	m.table.SetRows([]table.Row{{"old"}})
	m.rowData = []map[string]interface{}{{"value": "old"}}

	res, _ := m.Update(editSavedMsg{rowIndex: 0, colIndex: 0, value: "new", dataKey: "value"})
	got := res.(model)

	if got.rowData[0]["value"] != "new" {
		t.Errorf("expected rowData[0]['value'] == 'new', got %v", got.rowData[0]["value"])
	}
}

// ── T7: locked search re-applied after data load ──────────────────────────────

func TestLockedSearchReappliedAfterLoad(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.searchTerm = "invoice"
	m.searchMode = false // locked filter, not actively typing

	instances := []config.ProcessInstance{
		{ID: "inst-1", BusinessKey: "invoice-2024"},
		{ID: "inst-2", BusinessKey: "order-999"},
		{ID: "inst-3", BusinessKey: "invoice-2025"},
	}
	m.applyInstances(instances)

	rows := m.table.Rows()
	if len(rows) >= 3 {
		t.Errorf("expected filter to reduce rows below 3, got %d rows", len(rows))
	}
	for _, r := range rows {
		found := false
		for _, cell := range r {
			if strings.Contains(strings.ToLower(cell), "invoice") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected all rows to match 'invoice', got row: %v", r)
		}
	}
}

// ── T8: renderHelpScreen responsive width ────────────────────────────────────

func TestHelpScreenResponsiveWidth(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	// Should not panic
	out := m.renderHelpScreen(50, 30)
	if out == "" {
		t.Error("expected non-empty output from renderHelpScreen(50, 30)")
	}

	// Max line width (trimmed) should be <= width
	for _, line := range strings.Split(out, "\n") {
		stripped := stripANSIForTest(strings.TrimSpace(line))
		if len([]rune(stripped)) > 50 {
			t.Errorf("modal line exceeds terminal width of 50: %q (len=%d)", stripped, len([]rune(stripped)))
		}
	}
}

func TestHelpScreenMaxWidth(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	out := m.renderHelpScreen(200, 50)
	if out == "" {
		t.Error("expected non-empty output from renderHelpScreen(200, 50)")
	}

	// helpWidth capped at 76, plus borders(2) + padding(4) = 82.
	// Trim surrounding centering whitespace before measuring.
	for _, line := range strings.Split(out, "\n") {
		stripped := stripANSIForTest(strings.TrimSpace(line))
		if len([]rune(stripped)) > 82 {
			t.Errorf("modal line exceeds max width of 82: %q (len=%d)", stripped, len([]rune(stripped)))
		}
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

// stripANSIForTest removes ANSI escape codes for width measurement.
func stripANSIForTest(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
