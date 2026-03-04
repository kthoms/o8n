package app

import (
	"reflect"
	"testing"

	table "github.com/charmbracelet/bubbles/table"
)

// ── TransitionFull ──────────────────────────────────────────────────────────

func TestTransitionFull_ClearsAllFields(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalConfirmDelete
	m.footerError = "some error"
	m.searchTerm = "filter"
	m.searchMode = true
	m.sortColumn = 2
	m.sortAscending = false
	cols := []table.Column{{Title: "ID", Width: 10}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"a"}, {"b"}, {"c"}})
	m.table.SetCursor(2)
	m.navigationStack = []viewState{{viewMode: "process-definition"}}

	m.prepareStateTransition(TransitionFull)

	if m.activeModal != ModalNone {
		t.Error("activeModal not cleared")
	}
	if m.footerError != "" {
		t.Errorf("footerError not cleared, got %q", m.footerError)
	}
	if m.searchTerm != "" {
		t.Errorf("searchTerm not cleared, got %q", m.searchTerm)
	}
	if m.searchMode {
		t.Error("searchMode not cleared")
	}
	if m.sortColumn != -1 {
		t.Errorf("sortColumn not reset to -1, got %d", m.sortColumn)
	}
	if !m.sortAscending {
		t.Error("sortAscending not reset to true")
	}
	if m.table.Cursor() != 0 {
		t.Errorf("tableCursor not reset to 0, got %d", m.table.Cursor())
	}
	if len(m.navigationStack) != 0 {
		t.Errorf("navigationStack not cleared, got len=%d", len(m.navigationStack))
	}
}

func TestTransitionFull_ClearsIdentityParams(t *testing.T) {
	m := newTestModel(t)
	m.selectedDefinitionKey = "proc-def-key"
	m.selectedInstanceID = "inst-456"
	m.genericParams = map[string]string{"processInstanceId": "abc"}

	m.prepareStateTransition(TransitionFull)

	if m.selectedDefinitionKey != "" {
		t.Errorf("selectedDefinitionKey not cleared, got %q", m.selectedDefinitionKey)
	}
	if m.selectedInstanceID != "" {
		t.Errorf("selectedInstanceID not cleared, got %q", m.selectedInstanceID)
	}
	if len(m.genericParams) != 0 {
		t.Errorf("genericParams not cleared, got %v", m.genericParams)
	}
}

func TestTransitionFull_ClearsFilteredRows(t *testing.T) {
	m := newTestModel(t)
	m.originalRows = []table.Row{{"orig"}}
	m.filteredRows = []table.Row{{"filt"}}

	m.prepareStateTransition(TransitionFull)

	if m.originalRows != nil {
		t.Error("originalRows not cleared")
	}
	if m.filteredRows != nil {
		t.Error("filteredRows not cleared")
	}
}

// ── TransitionDrillDown ─────────────────────────────────────────────────────

func TestTransitionDrillDown_PushesStateBeforeClearing(t *testing.T) {
	m := newTestModel(t)
	m.viewMode = "process-definition"
	cols := []table.Column{{Title: "ID", Width: 10}}
	rows := []table.Row{{"a"}, {"b"}, {"c"}}
	m.table.SetColumns(cols)
	m.table.SetRows(rows)
	m.table.SetCursor(2)
	m.sortColumn = 1
	m.sortAscending = false

	m.prepareStateTransition(TransitionDrillDown)

	if len(m.navigationStack) != 1 {
		t.Fatalf("expected 1 entry pushed, got %d", len(m.navigationStack))
	}
	pushed := m.navigationStack[0]
	if pushed.viewMode != "process-definition" {
		t.Errorf("pushed wrong viewMode: got %q", pushed.viewMode)
	}
	if pushed.tableCursor != 2 {
		t.Errorf("pushed wrong tableCursor: got %d", pushed.tableCursor)
	}
}

func TestTransitionDrillDown_ClearsAfterPush(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalConfirmDelete
	m.searchTerm = "filter"
	m.searchMode = true
	m.sortColumn = 1
	m.sortAscending = false

	m.prepareStateTransition(TransitionDrillDown)

	if m.activeModal != ModalNone {
		t.Error("activeModal not cleared after DrillDown")
	}
	if m.searchTerm != "" {
		t.Errorf("searchTerm not cleared after DrillDown, got %q", m.searchTerm)
	}
	if m.searchMode {
		t.Error("searchMode not cleared after DrillDown")
	}
	if m.sortColumn != -1 {
		t.Errorf("sortColumn not reset after DrillDown, got %d", m.sortColumn)
	}
	if !m.sortAscending {
		t.Error("sortAscending not reset after DrillDown")
	}
}

func TestTransitionDrillDown_PushBeforeClear_CursorPreservedInSnapshot(t *testing.T) {
	// The cursor captured in the snapshot must be the PRE-clear cursor (3),
	// not the post-clear cursor (which would be 0 if clearing happened first).
	m := newTestModel(t)
	cols := []table.Column{{Title: "ID", Width: 10}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"a"}, {"b"}, {"c"}, {"d"}})
	m.table.SetCursor(3)

	m.prepareStateTransition(TransitionDrillDown)

	if len(m.navigationStack) == 0 {
		t.Fatal("stack empty after DrillDown")
	}
	if m.navigationStack[0].tableCursor != 3 {
		t.Errorf("push-before-clear violated: snapshot cursor=%d, want 3", m.navigationStack[0].tableCursor)
	}
}

// ── TransitionPop ───────────────────────────────────────────────────────────

func TestTransitionPop_RestoresAllFields(t *testing.T) {
	m := newTestModel(t)
	savedBreadcrumb := []string{"root", "child"}
	savedParams := map[string]string{"processInstanceId": "inst-99"}
	savedRowData := []map[string]interface{}{
		{"id": "a", "name": "first"},
	}
	m.navigationStack = []viewState{
		{
			viewMode:              "process-definition",
			breadcrumb:            savedBreadcrumb,
			contentHeader:         "Process Definitions",
			selectedDefinitionKey: "my-def",
			selectedInstanceID:    "inst-99",
			genericParams:         savedParams,
			rowData:               savedRowData,
			tableCursor:           5,
		},
	}

	m.prepareStateTransition(TransitionPop)

	if len(m.navigationStack) != 0 {
		t.Errorf("stack not emptied, got len=%d", len(m.navigationStack))
	}
	if m.viewMode != "process-definition" {
		t.Errorf("viewMode not restored, got %q", m.viewMode)
	}
	if m.contentHeader != "Process Definitions" {
		t.Errorf("contentHeader not restored, got %q", m.contentHeader)
	}
	if m.selectedDefinitionKey != "my-def" {
		t.Errorf("selectedDefinitionKey not restored, got %q", m.selectedDefinitionKey)
	}
	if m.selectedInstanceID != "inst-99" {
		t.Errorf("selectedInstanceID not restored, got %q", m.selectedInstanceID)
	}
	if !reflect.DeepEqual(m.breadcrumb, savedBreadcrumb) {
		t.Errorf("breadcrumb not restored, got %v", m.breadcrumb)
	}
	if !reflect.DeepEqual(m.genericParams, savedParams) {
		t.Errorf("genericParams not restored, got %v", m.genericParams)
	}
	if !reflect.DeepEqual(m.rowData, savedRowData) {
		t.Errorf("rowData not restored, got %v", m.rowData)
	}
}

func TestTransitionPop_RestoresTableCursor(t *testing.T) {
	m := newTestModel(t)
	cols := []table.Column{{Title: "ID", Width: 10}}
	rows := []table.Row{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}, {"f"}, {"g"}, {"h"}}
	m.navigationStack = []viewState{
		{
			viewMode:     "parent",
			tableCursor:  6,
			tableColumns: cols,
			tableRows:    rows,
		},
	}

	m.prepareStateTransition(TransitionPop)

	if m.table.Cursor() != 6 {
		t.Errorf("tableCursor not restored, got %d", m.table.Cursor())
	}
}

func TestTransitionPop_EmptyStack_IsNoop(t *testing.T) {
	m := newTestModel(t)
	m.navigationStack = nil
	m.viewMode = "current"

	// Must not panic
	m.prepareStateTransition(TransitionPop)

	// State unchanged
	if m.viewMode != "current" {
		t.Errorf("viewMode changed on empty-stack pop, got %q", m.viewMode)
	}
}

func TestTransitionPop_MultipleEntries_PopsTopOnly(t *testing.T) {
	m := newTestModel(t)
	m.navigationStack = []viewState{
		{viewMode: "level-0"},
		{viewMode: "level-1"},
		{viewMode: "level-2"},
	}

	m.prepareStateTransition(TransitionPop)

	if len(m.navigationStack) != 2 {
		t.Errorf("expected 2 entries remaining, got %d", len(m.navigationStack))
	}
	if m.viewMode != "level-2" {
		t.Errorf("expected top entry (level-2) restored, got %q", m.viewMode)
	}
}

// ── Round-trip: DrillDown then Pop restores original state ─────────────────

func TestDrillDownThenPop_RestoresOriginalState(t *testing.T) {
	m := newTestModel(t)
	m.viewMode = "process-definition"
	m.contentHeader = "Process Definitions"
	cols := []table.Column{{Title: "Key", Width: 20}}
	rows := []table.Row{{"proc-a"}, {"proc-b"}}
	m.table.SetColumns(cols)
	m.table.SetRows(rows)
	m.table.SetCursor(1)

	// Drill down into child view
	m.prepareStateTransition(TransitionDrillDown)
	m.viewMode = "process-instance"
	m.contentHeader = "Process Instances"

	// Pop back
	m.prepareStateTransition(TransitionPop)

	if m.viewMode != "process-definition" {
		t.Errorf("viewMode not restored, got %q", m.viewMode)
	}
	if m.contentHeader != "Process Definitions" {
		t.Errorf("contentHeader not restored, got %q", m.contentHeader)
	}
	if m.table.Cursor() != 1 {
		t.Errorf("tableCursor not restored, got %d", m.table.Cursor())
	}
	if len(m.navigationStack) != 0 {
		t.Errorf("stack should be empty after pop, got len=%d", len(m.navigationStack))
	}
}
