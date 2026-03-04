package app

import (
	"testing"

	table "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// ── TransitionFull: environment switch clears all leaking state ────────────

func TestEnvSwitchClearsNavigationStack(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.navigationStack = []viewState{{viewMode: "process-definition"}}

	m.prepareStateTransition(TransitionFull)

	if m.navigationStack != nil {
		t.Errorf("expected navigationStack nil after TransitionFull, got %v", m.navigationStack)
	}
}

func TestEnvSwitchClearsGenericParams(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.genericParams = map[string]string{"processInstanceId": "abc"}

	m.prepareStateTransition(TransitionFull)

	if len(m.genericParams) != 0 {
		t.Errorf("expected empty genericParams after TransitionFull, got %v", m.genericParams)
	}
}

func TestEnvSwitchClearsSelectedKeys(t *testing.T) {
	m := newTestModel(t)
	m.selectedDefinitionKey = "my-process"
	m.selectedInstanceID = "inst-123"

	m.prepareStateTransition(TransitionFull)

	if m.selectedDefinitionKey != "" {
		t.Errorf("expected empty selectedDefinitionKey, got %q", m.selectedDefinitionKey)
	}
	if m.selectedInstanceID != "" {
		t.Errorf("expected empty selectedInstanceID, got %q", m.selectedInstanceID)
	}
}

// ── TransitionFull: context switch clears sort and nav state ──────────────

func TestContextSwitchClearsSortState(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.sortColumn = 3
	m.sortAscending = false

	m.prepareStateTransition(TransitionFull)

	if m.sortColumn != -1 {
		t.Errorf("expected sortColumn -1 after TransitionFull, got %d", m.sortColumn)
	}
	if !m.sortAscending {
		t.Errorf("expected sortAscending true after TransitionFull")
	}
}

func TestContextSwitchClearsNavStack(t *testing.T) {
	m := newTestModel(t)
	m.navigationStack = []viewState{{viewMode: "process-definition"}, {viewMode: "process-instance"}}

	m.prepareStateTransition(TransitionFull)

	if m.navigationStack != nil {
		t.Errorf("expected nil navigationStack after TransitionFull, got %v", m.navigationStack)
	}
}

// ── Breadcrumb navigation: TransitionFull (root) and TransitionPop (depth) ─

func TestBreadcrumbNavToRootClearsNavStack(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.navigationStack = []viewState{
		{viewMode: "a"},
		{viewMode: "b"},
		{viewMode: "c"},
	}

	// Root breadcrumb navigation uses TransitionFull (full reset)
	m.prepareStateTransition(TransitionFull)

	if m.navigationStack != nil {
		t.Errorf("expected nil navStack after TransitionFull (breadcrumb root), got %v", m.navigationStack)
	}
}

func TestBreadcrumbNavToDepthRestoresState(t *testing.T) {
	m := newTestModel(t)
	m.navigationStack = []viewState{
		{viewMode: "a", tableCursor: 3},
		{viewMode: "b", tableCursor: 7},
		{viewMode: "c", tableCursor: 1},
	}

	// Breadcrumb jump to depth 1: caller truncates stack so target is at top, then calls TransitionPop
	m.navigationStack = m.navigationStack[:2] // keep entries 0 and 1 (idx+1=2)
	m.prepareStateTransition(TransitionPop)

	// TransitionPop pops the top (entry 1: viewMode="b", tableCursor=7)
	if len(m.navigationStack) != 1 {
		t.Errorf("expected navStack len 1 after truncate+pop, got %d", len(m.navigationStack))
	}
	if m.viewMode != "b" {
		t.Errorf("expected viewMode=b, got %q", m.viewMode)
	}
}

// ── TransitionFull and TransitionDrillDown clear sort and search ──────────

func TestTransitionFullClearsSortAndSearch(t *testing.T) {
	m := newTestModel(t)
	m.sortColumn = 2
	m.sortAscending = false
	m.searchTerm = "myterm"
	m.originalRows = []table.Row{{"id1"}}

	m.prepareStateTransition(TransitionFull)

	if m.sortColumn != -1 {
		t.Errorf("TransitionFull: expected sortColumn -1, got %d", m.sortColumn)
	}
	if !m.sortAscending {
		t.Errorf("TransitionFull: expected sortAscending true")
	}
	if m.searchTerm != "" {
		t.Errorf("TransitionFull: expected empty searchTerm, got %q", m.searchTerm)
	}
	if m.originalRows != nil {
		t.Errorf("TransitionFull: expected nil originalRows")
	}
}

func TestTransitionDrillDownClearsSortAndSearch(t *testing.T) {
	m := newTestModel(t)
	m.sortColumn = 2
	m.sortAscending = false
	m.searchTerm = "myterm"
	m.originalRows = []table.Row{{"id1"}}

	m.prepareStateTransition(TransitionDrillDown)

	if m.sortColumn != -1 {
		t.Errorf("TransitionDrillDown: expected sortColumn -1, got %d", m.sortColumn)
	}
	if !m.sortAscending {
		t.Errorf("TransitionDrillDown: expected sortAscending true")
	}
	if m.searchTerm != "" {
		t.Errorf("TransitionDrillDown: expected empty searchTerm, got %q", m.searchTerm)
	}
	if m.originalRows != nil {
		t.Errorf("TransitionDrillDown: expected nil originalRows")
	}
}

// ── Cursor bounds after row deletion ─────────────────────────────────────

func TestCursorBoundsAfterTerminate(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	cols := []table.Column{{Title: "ID", Width: 10}}
	rows := []table.Row{{"row1"}, {"row2"}, {"row3"}}
	m.table.SetColumns(cols)
	m.table.SetRows(rows)
	m.table.SetCursor(2)

	m2raw, _ := m.Update(terminatedMsg{id: "row3"})
	m2 := m2raw.(model)

	remaining := m2.table.Rows()
	cursor := m2.table.Cursor()
	if cursor >= len(remaining) && len(remaining) > 0 {
		t.Errorf("cursor %d out of bounds for %d rows after terminate", cursor, len(remaining))
	}
}

func TestCursorBoundsAfterDeleteOnlyRow(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	cols := []table.Column{{Title: "ID", Width: 10}}
	rows := []table.Row{{"only-row"}}
	m.table.SetColumns(cols)
	m.table.SetRows(rows)
	m.table.SetCursor(0)

	m2raw, _ := m.Update(terminatedMsg{id: "only-row"})
	m2 := m2raw.(model)

	cursor := m2.table.Cursor()
	if cursor > 0 {
		t.Errorf("expected cursor <= 0 after deleting only row, got %d", cursor)
	}
}

// ── Integration: Esc after env switch doesn't restore stale state ─────────

func TestEscAfterEnvSwitchDoesNotRestoreOldStack(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.navigationStack = []viewState{{viewMode: "process-definition"}}
	m.genericParams = map[string]string{"processInstanceId": "abc"}

	// Simulate env switch (TransitionFull clears navStack)
	m.prepareStateTransition(TransitionFull)

	// Now press Esc — should NOT pop (stack is nil after TransitionFull)
	m2raw, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := m2raw.(model)

	if len(m2.navigationStack) > 0 {
		t.Errorf("expected empty navStack after Esc post-env-switch, got %v", m2.navigationStack)
	}
}
