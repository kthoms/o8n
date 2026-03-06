package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// ── T1: applySortIndicatorToColumns ─────────────────────────────────────────

func TestApplySortIndicatorAddsArrow(t *testing.T) {
	m := newTestModel(t)
	cols := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "NAME", Width: 30},
	}
	m.table.SetColumns(cols)
	m.sortColumn = 1
	m.sortAscending = true

	m.applySortIndicatorToColumns()

	got := m.table.Columns()
	if !strings.HasSuffix(got[1].Title, " ▲") {
		t.Errorf("expected col[1] title to end with ' ▲', got %q", got[1].Title)
	}
	if strings.HasSuffix(got[0].Title, " ▲") || strings.HasSuffix(got[0].Title, " ▼") {
		t.Errorf("expected col[0] title to have no indicator, got %q", got[0].Title)
	}
}

func TestApplySortIndicatorDescending(t *testing.T) {
	m := newTestModel(t)
	cols := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "NAME", Width: 30},
	}
	m.table.SetColumns(cols)
	m.sortColumn = 0
	m.sortAscending = false

	m.applySortIndicatorToColumns()

	got := m.table.Columns()
	if !strings.HasSuffix(got[0].Title, " ▼") {
		t.Errorf("expected col[0] title to end with ' ▼', got %q", got[0].Title)
	}
}

func TestApplySortIndicatorClearsOnMinus1(t *testing.T) {
	m := newTestModel(t)
	cols := []table.Column{
		{Title: "ID ▲", Width: 20},
		{Title: "NAME ▼", Width: 30},
	}
	m.table.SetColumns(cols)
	m.sortColumn = -1

	m.applySortIndicatorToColumns()

	got := m.table.Columns()
	for _, c := range got {
		if strings.Contains(c.Title, " ▲") || strings.Contains(c.Title, " ▼") {
			t.Errorf("expected no sort indicator when sortColumn=-1, got %q", c.Title)
		}
	}
}

// ── T2: renderConfirmDeleteModal uses pendingDeleteID + pendingDeleteLabel ───

func TestConfirmDeleteModalUsesPendingID(t *testing.T) {
	m := newTestModel(t)
	m.pendingDeleteID = "proc-abc-123"
	m.pendingDeleteLabel = "My Process"

	// Place a different value in the table rows so we can verify the modal
	// reads from pendingDeleteID, not the live cursor row.
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 20}, {Title: "NAME", Width: 30}})
	m.table.SetRows([]table.Row{{"other-id", "Other Name"}})

	out := m.renderConfirmDeleteModal(80, 24)

	if !strings.Contains(out, "proc-abc-123") {
		t.Errorf("expected modal to contain pendingDeleteID 'proc-abc-123', got:\n%s", out)
	}
	if !strings.Contains(out, "My Process") {
		t.Errorf("expected modal to contain pendingDeleteLabel 'My Process', got:\n%s", out)
	}
	if strings.Contains(out, "other-id") {
		t.Errorf("expected modal NOT to use live cursor row 'other-id', got:\n%s", out)
	}
}

func TestConfirmDeleteModalEmptyWhenNoPendingID(t *testing.T) {
	m := newTestModel(t)
	m.pendingDeleteID = ""

	out := m.renderConfirmDeleteModal(80, 24)

	if out != "" {
		t.Errorf("expected empty string when pendingDeleteID is empty, got:\n%s", out)
	}
}

// ── T4: context switcher shows filtered list + hint line ─────────────────────

func TestContextBoxShowsMatchingContexts(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.rootContexts = []string{"process-definition", "process-instance", "task", "user-task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "task"
	m.lastWidth = 120
	m.lastHeight = 30

	out := m.View()

	if !strings.Contains(out, "task") {
		t.Errorf("expected View() output to contain 'task' when popup.input='task', got:\n%s", out)
	}
}

func TestContextBoxShowsHintLine(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.rootContexts = []string{"process-definition", "task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = ""
	m.lastWidth = 120
	m.lastHeight = 30

	out := m.View()

	if !strings.Contains(out, "↑↓ nav") { // Modal factory uses its own HintLine
		t.Errorf("expected '↑↓ nav' hint in context switcher modal, got:\n%s", out)
	}
}

// ── T5: computePaneHeight ────────────────────────────────────────────────────

func TestComputePaneHeightBase(t *testing.T) {
	m := newTestModel(t)
	m.lastHeight = 30
	m.popup.mode = popupModeNone
	m.searchMode = false

	h := m.computePaneHeight()
	// 30 - 2 (header) - 2 (footer) - 1 (safe) = 25
	want := 25
	if h != want {
		t.Errorf("expected pane height %d, got %d", want, h)
	}
}

func TestComputePaneHeightWithSkinPopup(t *testing.T) {
	m := newTestModel(t)
	m.lastHeight = 30
	m.popup.mode = popupModeSkin
	m.rootContexts = nil // no items → popup height = 4 (2 borders + input + hint)
	m.searchMode = false

	h := m.computePaneHeight()
	want := 21 // 30 - 2(header) - 2(footer) - 1(safe) - 4(popup)
	if h != want {
		t.Errorf("expected pane height %d with popup, got %d", want, h)
	}
}

func TestComputePaneHeightWithSearch(t *testing.T) {
	m := newTestModel(t)
	m.lastHeight = 30
	m.popup.mode = popupModeNone
	m.searchMode = true

	h := m.computePaneHeight()
	want := 24
	if h != want {
		t.Errorf("expected pane height %d with search, got %d", want, h)
	}
}

func TestComputePaneHeightBothActive(t *testing.T) {
	m := newTestModel(t)
	m.lastHeight = 30
	m.popup.mode = popupModeSkin
	m.rootContexts = nil // no items → popup height = 4
	m.searchMode = true

	h := m.computePaneHeight()
	want := 20 // 30 - 2 - 2 - 1 - 4(popup) - 1(search)
	if h != want {
		t.Errorf("expected pane height %d with both active, got %d", want, h)
	}
}

func TestComputePaneHeightModalContextSwitcherDoesNotAffectHeight(t *testing.T) {
	m := newTestModel(t)
	m.lastHeight = 30
	m.activeModal = ModalContextSwitcher
	m.searchMode = false

	h := m.computePaneHeight()
	want := 25 // Should be same as Base because it's an overlay
	if h != want {
		t.Errorf("expected pane height %d when ModalContextSwitcher active, got %d", want, h)
	}
}

// ── T6: ctrl+c → ModalConfirmQuit; second ctrl+c quits; Esc cancels ─────────

func TestCtrlCFirstPressShowsQuitModal(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalNone

	m2, _ := sendKeyString(m, "ctrl+c")

	if m2.activeModal != ModalConfirmQuit {
		t.Errorf("expected ModalConfirmQuit after first ctrl+c, got modal=%v", m2.activeModal)
	}
}

func TestCtrlCSecondPressQuits(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalConfirmQuit

	_, cmd := sendKeyString(m, "ctrl+c")

	if cmd == nil {
		t.Error("expected non-nil quit command when ctrl+c pressed with ModalConfirmQuit active")
	}
	// Execute the command to verify it produces a quit message
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg from quit command, got %T", msg)
	}
}

func TestEscClosesQuitModal(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalConfirmQuit

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Errorf("expected ModalNone after Esc, got modal=%v", m2.activeModal)
	}
}

// ── T8: env popup marks current active env with ✓ ───────────────────────────

func TestEnvPopupActiveMarker(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	// currentEnv is set by newModel; newTestModel sets "local"
	m.currentEnv = "local"

	out := m.renderEnvPopup(80, 24)

	if !strings.Contains(out, "✓") {
		t.Errorf("expected '✓' marker in env popup, got:\n%s", out)
	}

	// Verify the marker appears on the local line
	lines := strings.Split(out, "\n")
	foundMarkerOnLocal := false
	for _, line := range lines {
		if strings.Contains(line, "local") && strings.Contains(line, "✓") {
			foundMarkerOnLocal = true
			break
		}
	}
	if !foundMarkerOnLocal {
		t.Errorf("expected '✓' to appear on the 'local' environment line, got:\n%s", out)
	}
}
