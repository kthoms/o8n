package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/dao"
)

// ── T1: Search scope warning ──────────────────────────────────────────────────

func TestSearchOnPage2SetsFooterWarning(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.currentRoot = dao.ResourceProcessDefinitions
	// Simulate being on page 2 (offset > 0)
	m.pageOffsets[dao.ResourceProcessDefinitions] = 50

	m2, _ := sendKeyString(m, "/")

	if m2.popup.mode != popupModeSearch {
		t.Fatal("expected popup.mode == popupModeSearch")
	}
	if m2.footerStatusKind != footerStatusInfo {
		t.Errorf("expected footerStatusInfo, got %v", m2.footerStatusKind)
	}
	if !strings.Contains(m2.footerError, "Search limited to current page") {
		t.Errorf("expected page warning in footer, got %q", m2.footerError)
	}
}

func TestSearchOnPage1NoFooterWarning(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.currentRoot = dao.ResourceProcessDefinitions
	// No offset → page 1

	m2, _ := sendKeyString(m, "/")

	if m2.popup.mode != popupModeSearch {
		t.Fatal("expected popup.mode == popupModeSearch")
	}
	if m2.footerStatusKind == footerStatusInfo {
		t.Errorf("expected no footerStatusInfo on page 1, got %v and %q", m2.footerStatusKind, m2.footerError)
	}
}

func TestSearchEscClearsFooterInfo(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.currentRoot = dao.ResourceProcessDefinitions
	m.pageOffsets[dao.ResourceProcessDefinitions] = 50

	// Enter search (sets footerStatusInfo warning)
	m2, _ := sendKeyString(m, "/")
	if m2.footerStatusKind != footerStatusInfo {
		t.Fatal("precondition: expected footerStatusInfo to be set")
	}

	// Esc should clear it
	m3, _ := sendKeyString(m2, "esc")
	if m3.footerStatusKind == footerStatusInfo {
		t.Errorf("expected footerStatusInfo cleared after Esc, still have %q", m3.footerError)
	}
}

// ── T2: Sort clear option ─────────────────────────────────────────────────────

func TestSortPopupCursorCanGoToNegativeOne(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	// Set up a sort that is active
	m.sortColumn = 0
	m.sortAscending = true
	m.activeModal = ModalSort
	m.sortPopupCursor = 0

	// Press Up — should move cursor to -1 (clear sort item)
	m2, _ := sendKeyString(m, "up")

	if m2.sortPopupCursor != -1 {
		t.Errorf("expected sortPopupCursor -1 after Up from 0 with active sort, got %d", m2.sortPopupCursor)
	}
}

func TestSortPopupEnterAtNegativeOneClearsSortColumn(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	rows := []table.Row{{"charlie"}, {"alpha"}, {"bravo"}}
	cols := []table.Column{{Title: "Name", Width: 10}}
	m.table.SetColumns(cols)
	m.table.SetRows(rows)
	m.sortColumn = 0
	m.sortAscending = true
	m.activeModal = ModalSort
	m.sortPopupCursor = -1

	// Press Enter at cursor -1 → should clear sort
	m2, _ := sendKeyString(m, "enter")

	if m2.sortColumn != -1 {
		t.Errorf("expected sortColumn -1 after Enter at clear option, got %d", m2.sortColumn)
	}
	if m2.activeModal == ModalSort {
		t.Error("expected sort popup to close after selecting clear")
	}
}

// ── T3: Sort persists via setTableRowsSorted ──────────────────────────────────

func TestSetTableRowsSortedAppliesSort(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	cols := []table.Column{{Title: "Name", Width: 10}}
	m.table.SetColumns(cols)

	m.sortColumn = 0
	m.sortAscending = true

	unsorted := []table.Row{{"charlie"}, {"alpha"}, {"bravo"}}
	m.setTableRowsSorted(unsorted)

	rows := m.table.Rows()
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	if rows[0][0] != "alpha" {
		t.Errorf("expected first row 'alpha', got %q", rows[0][0])
	}
	if rows[2][0] != "charlie" {
		t.Errorf("expected last row 'charlie', got %q", rows[2][0])
	}
}

func TestSetTableRowsSortedNoSortWhenColumnNegOne(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	cols := []table.Column{{Title: "Name", Width: 10}}
	m.table.SetColumns(cols)

	m.sortColumn = -1

	unsorted := []table.Row{{"charlie"}, {"alpha"}, {"bravo"}}
	m.setTableRowsSorted(unsorted)

	rows := m.table.Rows()
	// Order should be unchanged when no sort
	if rows[0][0] != "charlie" {
		t.Errorf("expected first row 'charlie' (no sort), got %q", rows[0][0])
	}
}

// ── T4: Edit modal label ──────────────────────────────────────────────────────

func TestRenderEditModalSingleColumnShowsHeader(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	cols := []table.Column{{Title: "Value", Width: 20}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"hello"}})

	m.editColumns = []editableColumn{
		{index: 0, def: config.ColumnDef{Name: "myField", InputType: "text"}},
	}
	m.editColumnPos = 0
	m.editRowIndex = 0
	m.editTableKey = "test-table"
	m.editInput.SetValue("hello")
	m.activeModal = ModalEdit

	out := m.renderEditModal(80, 24)

	if !strings.Contains(out, "Editing: myField") {
		t.Errorf("expected 'Editing: myField' in single-column edit modal, got:\n%s", out)
	}
	if !strings.Contains(out, "type: text") {
		t.Errorf("expected 'type: text' in single-column edit modal, got:\n%s", out)
	}
}

func TestRenderEditModalMultiColumnNoHeader(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	cols := []table.Column{{Title: "A", Width: 10}, {Title: "B", Width: 10}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"val1", "val2"}})

	m.editColumns = []editableColumn{
		{index: 0, def: config.ColumnDef{Name: "fieldA", InputType: "text"}},
		{index: 1, def: config.ColumnDef{Name: "fieldB", InputType: "text"}},
	}
	m.editColumnPos = 0
	m.editRowIndex = 0
	m.editTableKey = "test-table"
	m.editInput.SetValue("val1")
	m.activeModal = ModalEdit

	out := m.renderEditModal(80, 24)

	if strings.Contains(out, "Editing:") {
		t.Errorf("expected no 'Editing:' header for multi-column modal, got:\n%s", out)
	}
}

// ── T6: Detail scroll without detailMaxScroll ─────────────────────────────────

func TestDetailScrollDownKey(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	// Build content with many lines so scrolling is possible
	lines := make([]string, 30)
	for i := range lines {
		lines[i] = "line"
	}
	m.detailContent = strings.Join(lines, "\n")
	m.detailScroll = 0
	m.lastHeight = 24
	m.activeModal = ModalDetailView

	m2, _ := sendKeyString(m, "down")

	if m2.detailScroll != 1 {
		t.Errorf("expected detailScroll 1 after down, got %d", m2.detailScroll)
	}
}

func TestDetailScrollUpKey(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	lines := make([]string, 30)
	for i := range lines {
		lines[i] = "line"
	}
	m.detailContent = strings.Join(lines, "\n")
	m.detailScroll = 5
	m.lastHeight = 24
	m.activeModal = ModalDetailView

	m2, _ := sendKeyString(m, "up")

	if m2.detailScroll != 4 {
		t.Errorf("expected detailScroll 4 after up, got %d", m2.detailScroll)
	}
}

func TestDetailScrollGKey(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	lines := make([]string, 30)
	for i := range lines {
		lines[i] = "line"
	}
	m.detailContent = strings.Join(lines, "\n")
	m.detailScroll = 0
	m.lastHeight = 24
	m.activeModal = ModalDetailView

	m2, _ := sendKeyString(m, "G")

	// detailViewH = lastHeight - 6 = 18; maxScroll = 30-18 = 12
	expectedMax := 30 - (24 - 6)
	if m2.detailScroll != expectedMax {
		t.Errorf("expected detailScroll %d after G, got %d", expectedMax, m2.detailScroll)
	}
}

// ── T7: Raw API data in detail view ──────────────────────────────────────────

func TestBuildDetailContentUsesRowData(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	cols := []table.Column{{Title: "id", Width: 10}, {Title: "name", Width: 20}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"proc-1", "MyProcess"}})

	// Provide raw API data with extra hidden fields
	m.rowData = []map[string]interface{}{
		{"id": "proc-1", "name": "MyProcess", "hiddenField": "secretValue"},
	}

	row := m.table.Rows()[0]
	out := m.buildDetailContent(row)

	if !strings.Contains(out, "hiddenField") {
		t.Errorf("expected hiddenField in detail content from rowData, got:\n%s", out)
	}
	if !strings.Contains(out, "secretValue") {
		t.Errorf("expected secretValue in detail content from rowData, got:\n%s", out)
	}
}

func TestBuildDetailContentFallsBackToColumns(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	cols := []table.Column{{Title: "id", Width: 10}, {Title: "name", Width: 20}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"proc-1", "MyProcess"}})

	// No rowData — should fall back to display columns
	m.rowData = nil

	row := m.table.Rows()[0]
	out := m.buildDetailContent(row)

	if !strings.Contains(out, "id") {
		t.Errorf("expected 'id' column in fallback detail content, got:\n%s", out)
	}
	if !strings.Contains(out, "proc-1") {
		t.Errorf("expected 'proc-1' value in fallback detail content, got:\n%s", out)
	}
}
