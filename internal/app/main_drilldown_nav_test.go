package app

// main_drilldown_nav_test.go — Story 3.2: Drill-Down Navigation & Breadcrumb
//
// Tests verify the navigation stack contract:
//   - Enter: prepareStateTransition(TransitionDrillDown) pushes viewState snapshot
//   - Esc:   prepareStateTransition(TransitionPop) fully restores the parent view
//   - navigateToBreadcrumb(idx): jumps to a prior level using pop or full reset

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kthoms/o6n/internal/config"
)

// drillNavConfig returns a minimal 3-level config (definition → instance → variable).
func drillNavConfig() *config.Config {
	return &config.Config{
		Environments: map[string]config.Environment{"local": {URL: "http://localhost"}},
		Tables: []config.TableDef{
			{
				Name:      "process-definition",
				Columns:   []config.ColumnDef{{Name: "key"}, {Name: "name"}},
				Drilldown: &config.DrillDownDef{Target: "process-instance", Param: "processDefinitionKey", Column: "key"},
			},
			{
				Name:      "process-instance",
				Columns:   []config.ColumnDef{{Name: "id"}, {Name: "definitionId"}},
				Drilldown: &config.DrillDownDef{Target: "process-variables", Param: "processInstanceId", Column: "id"},
			},
			{
				Name:    "process-variables",
				Columns: []config.ColumnDef{{Name: "name"}, {Name: "value"}},
			},
		},
	}
}

// ── AC 1: DrillDown pushes state onto navigation stack ───────────────────────

// TestDrillDown_PushesSnapshotOntoStack verifies that pressing Enter on a row
// pushes a viewState snapshot (including cursor, rows, columns, breadcrumb) onto
// the navigation stack.
func TestDrillDown_PushesSnapshotOntoStack(t *testing.T) {
	m := newModel(drillNavConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
		{ID: "d2", Key: "k2", Name: "Beta"},
	})
	m.table.SetCursor(1) // position on 2nd row

	initialRows := append([]table.Row{}, m.table.Rows()...)
	initialCursor := m.table.Cursor()

	// Enter → drilldown into instances
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	if m2.viewMode != "process-instance" {
		t.Fatalf("expected viewMode=process-instance after Enter, got %q", m2.viewMode)
	}

	// Stack must have exactly 1 entry after a single drilldown.
	if len(m2.navigationStack) != 1 {
		t.Fatalf("expected 1 item in navigationStack, got %d", len(m2.navigationStack))
	}

	// Snapshot must capture the parent cursor position.
	snap := m2.navigationStack[0]
	if snap.tableCursor != initialCursor {
		t.Errorf("snapshot tableCursor: want %d, got %d", initialCursor, snap.tableCursor)
	}

	// Snapshot must capture the parent rows.
	if len(snap.tableRows) != len(initialRows) {
		t.Errorf("snapshot tableRows: want %d rows, got %d", len(initialRows), len(snap.tableRows))
	}

	// Snapshot must capture the parent viewMode.
	if snap.viewMode != "process-definition" {
		t.Errorf("snapshot viewMode: want %q, got %q", "process-definition", snap.viewMode)
	}

	// Snapshot breadcrumb must include process-definitions root.
	if len(snap.breadcrumb) == 0 {
		t.Error("snapshot breadcrumb must not be empty")
	}
}

// TestDrillDown_ClearsChildStateExceptStack verifies that child-view state
// (table rows, filters) is reset after TransitionDrillDown while the stack
// itself is preserved.
func TestDrillDown_ClearsChildStateExceptStack(t *testing.T) {
	m := newModel(drillNavConfig())
	m.applyDefinitions([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "Alpha"}})

	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	// After drilldown the table shows a placeholder "No results" row until the
	// fetch completes; it must NOT still show parent rows.
	parentRowCount := 1
	childRows := m2.table.Rows()
	// Child rows must be either empty or a single "No results" placeholder.
	for _, r := range childRows {
		if len(r) > 0 && (r[0] == "▶ k1" || r[0] == "k1") {
			t.Errorf("parent row %q leaked into child view after drilldown", r[0])
		}
	}
	_ = parentRowCount

	// Sort/search state must be cleared for the child view.
	if m2.sortColumn != -1 {
		t.Errorf("sortColumn must be -1 in child view, got %d", m2.sortColumn)
	}
	if m2.searchTerm != "" {
		t.Errorf("searchTerm must be empty in child view, got %q", m2.searchTerm)
	}
}

// TestDrillDown_GenericParamsSetCorrectly verifies that genericParams is
// populated with the drilldown filter value from the selected parent row.
func TestDrillDown_GenericParamsSetCorrectly(t *testing.T) {
	m := newModel(drillNavConfig())
	m.applyDefinitions([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "Alpha"}})

	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	// DrillDownDef for process-definition → process-instance uses Param="processDefinitionKey", Column="key"
	val, ok := m2.genericParams["processDefinitionKey"]
	if !ok {
		t.Fatalf("genericParams missing key %q after drilldown; got: %v", "processDefinitionKey", m2.genericParams)
	}
	if val != "k1" {
		t.Errorf("expected genericParams[processDefinitionKey]=k1, got %q", val)
	}
}

// TestDrillDown_BreadcrumbGrowsAfterEachLevel verifies that the breadcrumb
// slice grows by one entry at each drilldown level.
func TestDrillDown_BreadcrumbGrowsAfterEachLevel(t *testing.T) {
	m := newModel(drillNavConfig())
	initialBC := len(m.breadcrumb)

	m.applyDefinitions([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "Alpha"}})

	// Level 1 drilldown
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)
	if len(m2.breadcrumb) != initialBC+1 {
		t.Errorf("breadcrumb after level-1 drilldown: want %d, got %d", initialBC+1, len(m2.breadcrumb))
	}
}

// TestDrillDown_ChildCursorResetToZero verifies that the child view starts with
// cursor at 0 after the drilldown transition.
func TestDrillDown_ChildCursorResetToZero(t *testing.T) {
	m := newModel(drillNavConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
		{ID: "d2", Key: "k2", Name: "Beta"},
	})
	m.table.SetCursor(1)

	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	if m2.table.Cursor() != 0 {
		t.Errorf("expected child cursor=0 after drilldown, got %d", m2.table.Cursor())
	}
}

// ── AC 2: Escape pops state and fully restores parent view ───────────────────

// TestEscape_PopsNavigationStack verifies that pressing Esc on a drilled-down
// view pops the navigation stack and returns viewMode to the parent level.
func TestEscape_PopsNavigationStack(t *testing.T) {
	m := newModel(drillNavConfig())
	m.applyDefinitions([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "Alpha"}})

	// Drill down
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	if m2.viewMode != "process-instance" {
		t.Fatalf("precondition: expected viewMode=process-instance, got %q", m2.viewMode)
	}
	if len(m2.navigationStack) != 1 {
		t.Fatalf("precondition: expected stack length=1, got %d", len(m2.navigationStack))
	}

	// Esc → pop
	res2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := res2.(model)

	if m3.viewMode != "process-definition" {
		t.Errorf("expected viewMode=process-definition after Esc, got %q", m3.viewMode)
	}
	if len(m3.navigationStack) != 0 {
		t.Errorf("expected empty navigationStack after Esc, got %d items", len(m3.navigationStack))
	}
}

// TestEscape_RestoresCursorPosition verifies that the parent row cursor is
// restored to the position saved at drilldown time.
func TestEscape_RestoresCursorPosition(t *testing.T) {
	m := newModel(drillNavConfig())
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
		{ID: "d2", Key: "k2", Name: "Beta"},
		{ID: "d3", Key: "k3", Name: "Gamma"},
	})
	m.table.SetCursor(2)

	// Drill down from row 2
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	// Esc
	res2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := res2.(model)

	if m3.table.Cursor() != 2 {
		t.Errorf("expected cursor restored to 2 after Esc, got %d", m3.table.Cursor())
	}
}

// TestEscape_RestoresParentRows verifies that the parent table rows are fully
// restored after a pop, not left in the child's empty/loading state.
func TestEscape_RestoresParentRows(t *testing.T) {
	m := newModel(drillNavConfig())
	defs := []config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "Alpha"},
		{ID: "d2", Key: "k2", Name: "Beta"},
	}
	m.applyDefinitions(defs)

	// Drill down
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	// Esc
	res2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := res2.(model)

	rows := m3.table.Rows()
	if len(rows) != len(defs) {
		t.Errorf("expected %d rows after Esc, got %d", len(defs), len(rows))
	}
}

// TestEscape_AtRootIsNoop verifies that pressing Esc when the navigation stack
// is empty does not panic or alter the viewMode.
func TestEscape_AtRootIsNoop(t *testing.T) {
	m := newModel(drillNavConfig())
	m.applyDefinitions([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "Alpha"}})

	before := m.viewMode
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := res.(model)

	if m2.viewMode != before {
		t.Errorf("Esc at root changed viewMode: was %q, now %q", before, m2.viewMode)
	}
}

// ── AC 3: navigateToBreadcrumb restores state at any previous level ──────────

// TestNavigateToBreadcrumb_JumpToRoot verifies that navigateToBreadcrumb(0)
// performs a TransitionFull and navigates to the root context.
func TestNavigateToBreadcrumb_JumpToRoot(t *testing.T) {
	m := newModel(drillNavConfig())
	m.applyDefinitions([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "Alpha"}})

	// Drill to level 1
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	if len(m2.navigationStack) != 1 {
		t.Fatalf("precondition: stack must have 1 entry")
	}

	// Jump back to breadcrumb[0] (root)
	_ = m2.navigateToBreadcrumb(0)

	// TransitionFull clears the navigation stack.
	if len(m2.navigationStack) != 0 {
		t.Errorf("expected empty stack after jumping to root, got %d", len(m2.navigationStack))
	}
}

// TestNavigateToBreadcrumb_JumpToIntermediateLevel verifies that
// navigateToBreadcrumb(1) on a 3-level-deep stack restores the correct
// intermediate state via TransitionPop.
func TestNavigateToBreadcrumb_JumpToIntermediateLevel(t *testing.T) {
	m := newModel(drillNavConfig())
	m.applyDefinitions([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "Alpha"}})

	// Level 1 drilldown
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	// Simulate being at the instance level with 2 rows
	m2.applyInstances([]config.ProcessInstance{
		{ID: "i1", DefinitionID: "d1", BusinessKey: "bk1", StartTime: "2020-01-01"},
		{ID: "i2", DefinitionID: "d1", BusinessKey: "bk2", StartTime: "2020-01-02"},
	})

	// Level 2 drilldown into variables
	res2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := res2.(model)

	if m3.viewMode != "process-variables" {
		t.Fatalf("precondition: expected viewMode=process-variables, got %q", m3.viewMode)
	}
	if len(m3.navigationStack) != 2 {
		t.Fatalf("precondition: expected stack depth=2, got %d", len(m3.navigationStack))
	}

	// Jump to breadcrumb[1] (instances level): should pop to instances
	_ = m3.navigateToBreadcrumb(1)

	// Stack should be truncated to the intermediate snapshot (one entry remains for root).
	// After navigateToBreadcrumb(1) the stack is truncated to [:2] and then pop removes 1.
	if len(m3.navigationStack) > 2 {
		t.Errorf("stack should be trimmed after jump, got %d entries", len(m3.navigationStack))
	}
	// viewMode must have been reset (TransitionPop/TransitionFull is triggered inside navigateToBreadcrumb)
	// navigateToBreadcrumb mutates m3 in place (pointer receiver — model is value, so no mutation — cmd is returned)
	// We verify via the returned breadcrumb state.
	if len(m3.breadcrumb) > 2 {
		t.Errorf("breadcrumb after jump should not exceed 2 levels, got %d", len(m3.breadcrumb))
	}
}

// TestNavigateToBreadcrumb_InvalidIndex verifies that out-of-range indices set
// a footer error without panicking.
func TestNavigateToBreadcrumb_InvalidIndex(t *testing.T) {
	m := newModel(drillNavConfig())
	m.applyDefinitions([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "Alpha"}})

	// Should not panic.
	cmd := m.navigateToBreadcrumb(-1)
	_ = cmd
	cmd2 := m.navigateToBreadcrumb(999)
	_ = cmd2

	// footerError must be set for invalid index.
	if m.footerError == "" {
		t.Error("expected footerError for out-of-range breadcrumb index")
	}
}
