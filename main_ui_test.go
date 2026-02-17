package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kthoms/o8n/internal/config"
)

func TestApplyDataPopulatesTable(t *testing.T) {
	cfg := &config.Config{Environments: map[string]config.Environment{"local": {URL: "http://example"}}}
	m := newModel(cfg)
	defs := []config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "One"}}
	insts := []config.ProcessInstance{{ID: "i1", DefinitionID: "d1", BusinessKey: "bk1", StartTime: "2020-01-01T00:00:00Z"}}
	m.applyData(defs, insts)

	rows := m.table.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][0] != "i1" {
		t.Fatalf("expected instance id i1, got %s", rows[0][0])
	}
}

func TestSelectionChangeTriggersManualRefreshFlag(t *testing.T) {
	// Create a dummy server for the client to call (won't be reached in this test)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	cfg := &config.Config{Environments: map[string]config.Environment{"local": {URL: server.URL}}}
	m := newModel(cfg)
	// populate list with two definitions so moving down changes selection
	defs := []config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "One"}, {ID: "d2", Key: "k2", Name: "Two"}}
	m.applyData(defs, nil)

	// ensure autoRefresh is off
	m.autoRefresh = false

	// Simulate that the lastListIndex is outdated so Update detects a change
	m.lastListIndex = 999

	// call Update with an unknown message type so Update goes through the default path
	res, _ := m.Update(struct{}{})
	m2, ok := res.(model)
	if !ok {
		t.Fatalf("expected model type from Update result")
	}

	if !m2.manualRefreshTriggered {
		t.Fatalf("expected manualRefreshTriggered to be true after selection change")
	}
}

func TestFetchCmdExecutesAndLoadsData(t *testing.T) {
	// Prepare server responses: definitions and instances
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/process-definition" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "One"}})
			return
		}
		if r.URL.Path == "/process-instance" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]config.ProcessInstance{{ID: "i1", DefinitionID: "d1", BusinessKey: "bk1", StartTime: "2020-01-01T00:00:00Z"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{Environments: map[string]config.Environment{"local": {URL: server.URL}}}
	m := newModel(cfg)
	// pre-populate list so selectedKey will be k1 after selection
	m.applyData([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "One"}}, nil)
	m.autoRefresh = false

	// Simulate selection change (no-op if only one item, so force index change by setting index then sending a no-op)
	// Ensure index is 0, then call fetch cmd directly.
	cmd := m.fetchDataCmd()
	if cmd == nil {
		t.Fatalf("expected non-nil fetch cmd")
	}

	msg := cmd()
	switch mm := msg.(type) {
	case dataLoadedMsg:
		// apply and ensure table updated
		m.applyData(mm.definitions, mm.instances)
		rows := m.table.Rows()
		if len(rows) != 1 || rows[0][0] != "i1" {
			t.Fatalf("expected table to contain instance i1, got %+v", rows)
		}
	default:
		t.Fatalf("expected dataLoadedMsg, got %T", mm)
	}
}

func TestFlashOnOff(t *testing.T) {
	cfg := &config.Config{Environments: map[string]config.Environment{"local": {URL: "http://example"}}}
	m := newModel(cfg)

	// Send flashOnMsg as if a remote signal was issued
	res, cmd := m.Update(flashOnMsg{})
	m2, ok := res.(model)
	if !ok {
		t.Fatalf("expected model from Update flashOnMsg")
	}
	if !m2.flashActive {
		t.Fatalf("expected flashActive true after flashOnMsg")
	}

	// The cmd should schedule the flashOffMsg; execute it to get the message
	if cmd == nil {
		t.Fatalf("expected non-nil cmd returned when flashOnMsg handled")
	}
	msg := cmd()
	switch msg.(type) {
	case flashOffMsg:
		// now pass the flashOffMsg into Update to clear the flash
		res2, _ := m2.Update(msg)
		m3, ok := res2.(model)
		if !ok {
			t.Fatalf("expected model from Update flashOffMsg")
		}
		if m3.flashActive {
			t.Fatalf("expected flashActive false after flashOffMsg")
		}
	default:
		t.Fatalf("expected flashOffMsg from cmd, got %T", msg)
	}
}

func TestConfigDrivenDrilldownFromDefinitionToInstances(t *testing.T) {
	// Server: respond to /process-instance queries
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/process-instance" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]config.ProcessInstance{{ID: "i1", DefinitionID: "d1", BusinessKey: "bk1", StartTime: "2020-01-01T00:00:00Z"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: server.URL}},
		Tables: []config.TableDef{
			{
				Name:      "process-definition",
				Columns:   []config.ColumnDef{{Name: "key", Visible: true}, {Name: "name", Visible: true}, {Name: "version", Visible: true}, {Name: "resource", Visible: true}},
				Drilldown: []config.DrillDownDef{{Target: "process-instance", Param: "processDefinitionKey", Column: "key"}},
			},
			{
				Name:    "process-instance",
				Columns: []config.ColumnDef{{Name: "id", Visible: true}, {Name: "definitionId", Visible: true}, {Name: "businessKey", Visible: true}, {Name: "startTime", Visible: true}},
			},
		},
	}

	m := newModel(cfg)
	// populate definitions table (selectable row)
	m.applyDefinitions([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "One", Version: 1, Resource: "res"}})

	// ensure we are in definitions view
	if m.viewMode != "definitions" {
		t.Fatalf("expected viewMode definitions, got %s", m.viewMode)
	}

	// press Enter to drill down (should use config-driven drill)
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2, ok := res.(model)
	if !ok {
		t.Fatalf("expected model from Update")
	}
	if m2.viewMode != "instances" {
		t.Fatalf("expected viewMode instances after drill, got %s", m2.viewMode)
	}
	// simulate the fetch that would be scheduled by the program (fetchInstancesCmd)
	fetchMsg := m2.fetchInstancesCmd("processDefinitionKey", "k1")()
	res2, _ := m2.Update(fetchMsg)
	m3 := res2.(model)
	rows := m3.table.Rows()
	if len(rows) != 1 || rows[0][0] != "i1" {
		t.Fatalf("expected instance i1 after drill, got %+v", rows)
	}
}

func TestConfigDrivenDrilldownFromInstancesToVariables(t *testing.T) {
	// Server: respond to /process-instance/{id}/variables
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/process-instance/") && strings.HasSuffix(r.URL.Path, "/variables") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]map[string]interface{}{"var1": {"value": "v1", "type": "String"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: server.URL}},
		Tables: []config.TableDef{
			{
				Name:      "process-instance",
				Columns:   []config.ColumnDef{{Name: "id", Visible: true}, {Name: "definitionId", Visible: true}},
				Drilldown: []config.DrillDownDef{{Target: "process-variables", Param: "processInstanceId", Column: "id"}},
			},
			{
				Name:    "process-variables",
				Columns: []config.ColumnDef{{Name: "name", Visible: true}, {Name: "value", Visible: true}},
			},
		},
	}

	m := newModel(cfg)
	// populate instances table and set view to instances
	m.applyInstances([]config.ProcessInstance{{ID: "i1", DefinitionID: "d1", BusinessKey: "bk1", StartTime: "2020-01-01T00:00:00Z"}})
	m.viewMode = "instances"
	m.breadcrumb = []string{m.currentRoot, "process-instances"}

	// press Enter to drill into variables (config-driven)
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2, ok := res.(model)
	if !ok {
		t.Fatalf("expected model from Update")
	}
	if m2.viewMode != "variables" {
		t.Fatalf("expected viewMode variables after drill, got %s", m2.viewMode)
	}
	// simulate the fetch that would be scheduled by the program (fetchVariablesCmd)
	fetchMsg := m2.fetchVariablesCmd("i1")()
	res2, _ := m2.Update(fetchMsg)
	m3 := res2.(model)
	rows := m3.table.Rows()
	if len(rows) != 1 || rows[0][0] != "var1" {
		t.Fatalf("expected variable 'var1' after drill, got %+v", rows)
	}
}

func TestNavigationStackPreservesRowSelection(t *testing.T) {
	// Server: respond to both definitions and instances
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/process-definition" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]config.ProcessDefinition{
				{ID: "d1", Key: "k1", Name: "One", Version: 1},
				{ID: "d2", Key: "k2", Name: "Two", Version: 1},
				{ID: "d3", Key: "k3", Name: "Three", Version: 1},
			})
			return
		}
		if r.URL.Path == "/process-instance" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]config.ProcessInstance{
				{ID: "i1", DefinitionID: "d2", BusinessKey: "bk1", StartTime: "2020-01-01T00:00:00Z"},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: server.URL}},
		Tables: []config.TableDef{
			{
				Name:      "process-definition",
				Columns:   []config.ColumnDef{{Name: "key", Visible: true}, {Name: "name", Visible: true}},
				Drilldown: []config.DrillDownDef{{Target: "process-instance", Param: "processDefinitionKey", Column: "key"}},
			},
		},
	}

	m := newModel(cfg)
	// populate definitions table
	m.applyDefinitions([]config.ProcessDefinition{
		{ID: "d1", Key: "k1", Name: "One", Version: 1},
		{ID: "d2", Key: "k2", Name: "Two", Version: 1},
		{ID: "d3", Key: "k3", Name: "Three", Version: 1},
	})

	// move cursor to row 2 (3rd definition)
	m.table.SetCursor(2)
	initialCursor := m.table.Cursor()
	if initialCursor != 2 {
		t.Fatalf("expected cursor at row 2, got %d", initialCursor)
	}

	// press Enter to drill down (should save state and reset cursor to 0)
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)

	// verify we drilled down
	if m2.viewMode != "instances" {
		t.Fatalf("expected instances view, got %s", m2.viewMode)
	}

	// simulate the fetch
	fetchMsg := m2.fetchInstancesCmd("processDefinitionKey", "k3")()
	res2, _ := m2.Update(fetchMsg)
	m3 := res2.(model)

	// verify cursor was reset to 0
	if m3.table.Cursor() != 0 {
		t.Fatalf("expected cursor reset to 0 after drill, got %d", m3.table.Cursor())
	}

	// verify navigation stack has saved state
	if len(m3.navigationStack) != 1 {
		t.Fatalf("expected 1 item in navigation stack, got %d", len(m3.navigationStack))
	}

	// press Esc to go back
	res3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m4 := res3.(model)

	// verify we're back to definitions view
	if m4.viewMode != "definitions" {
		t.Fatalf("expected definitions view after Esc, got %s", m4.viewMode)
	}

	// verify cursor was restored to row 2
	if m4.table.Cursor() != 2 {
		t.Fatalf("expected cursor restored to row 2 after Esc, got %d", m4.table.Cursor())
	}

	// verify navigation stack is now empty
	if len(m4.navigationStack) != 0 {
		t.Fatalf("expected empty navigation stack after Esc, got %d items", len(m4.navigationStack))
	}

	// verify table rows were restored
	rows := m4.table.Rows()
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows restored after Esc, got %d", len(rows))
	}
}

func TestEditableColumnsMarkedWithIndicator(t *testing.T) {
	// Setup config with editable column
	appCfg := &config.AppConfig{
		Tables: []config.TableDef{
			{
				Name: "process-variables",
				Columns: []config.ColumnDef{
					{Name: "name", Visible: true, Width: "30%", Editable: false},
					{Name: "value", Visible: true, Width: "70%", Editable: true, InputType: "text"},
				},
			},
		},
	}
	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: "http://example"}},
		Tables:       appCfg.Tables,
	}
	m := newModel(cfg)

	// Apply variables
	vars := []config.Variable{
		{Name: "myVar", Value: "test123", Type: "String"},
		{Name: "count", Value: "42", Type: "Integer"},
	}
	m.applyVariables(vars)

	rows := m.table.Rows()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// First column (name) should NOT have [E] marker in cell values
	if strings.Contains(rows[0][0], "[E]") {
		t.Errorf("name column should not be marked in cell values, got: %s", rows[0][0])
	}

	// Editable columns are now indicated in the header with a write emoji.
	cols := m.table.Columns()
	if len(cols) < 2 {
		t.Fatalf("expected at least 2 columns, got %d", len(cols))
	}
	// First column header should NOT contain the write emoji
	if strings.Contains(cols[0].Title, "🖍️") {
		t.Errorf("name column header should not be marked editable, got: %s", cols[0].Title)
	}
	// Second column header SHOULD contain the write emoji
	if !strings.Contains(cols[1].Title, "🖍️") {
		t.Errorf("value column header should be marked editable with 🖍️, got: %s", cols[1].Title)
	}
}

func TestEditableColumnsFor(t *testing.T) {
	appCfg := &config.AppConfig{
		Tables: []config.TableDef{
			{
				Name: "process-variables",
				Columns: []config.ColumnDef{
					{Name: "name", Visible: true, Editable: false},
					{Name: "value", Visible: true, Editable: true},
					{Name: "type", Visible: true, Editable: false},
				},
			},
		},
	}
	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: "http://example"}},
		Tables:       appCfg.Tables,
	}
	m := newModel(cfg)

	editableCols := m.editableColumnsFor("process-variables")
	if len(editableCols) != 1 {
		t.Fatalf("expected 1 editable column, got %d", len(editableCols))
	}

	if editableCols[0].index != 1 {
		t.Errorf("expected editable column at index 1, got %d", editableCols[0].index)
	}

	if editableCols[0].def.Name != "value" {
		t.Errorf("expected editable column name 'value', got '%s'", editableCols[0].def.Name)
	}
}

func TestExtraEntersDontPushNavigationStack(t *testing.T) {
	// Server: respond to definitions, instances and variables
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/process-definition" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "One"}})
			return
		}
		if r.URL.Path == "/process-instance" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]config.ProcessInstance{{ID: "i1", DefinitionID: "d1", BusinessKey: "bk1", StartTime: "2020-01-01T00:00:00Z"}})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/process-instance/") && strings.HasSuffix(r.URL.Path, "/variables") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]map[string]interface{}{"var1": {"value": "v1", "type": "String"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: server.URL}},
		Tables: []config.TableDef{
			{
				Name:      "process-definition",
				Columns:   []config.ColumnDef{{Name: "key", Visible: true}, {Name: "name", Visible: true}},
				Drilldown: []config.DrillDownDef{{Target: "process-instance", Param: "processDefinitionKey", Column: "key"}},
			},
			{
				Name:      "process-instance",
				Columns:   []config.ColumnDef{{Name: "id", Visible: true}, {Name: "definitionId", Visible: true}},
				Drilldown: []config.DrillDownDef{{Target: "process-variables", Param: "processInstanceId", Column: "id"}},
			},
			{
				Name:    "process-variables",
				Columns: []config.ColumnDef{{Name: "name", Visible: true}, {Name: "value", Visible: true}},
			},
		},
	}

	m := newModel(cfg)
	// populate definitions
	m.applyDefinitions([]config.ProcessDefinition{{ID: "d1", Key: "k1", Name: "One"}})

	// Enter -> instances
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := res.(model)
	if m2.viewMode != "instances" {
		t.Fatalf("expected instances, got %s", m2.viewMode)
	}
	// simulate fetchInstances
	fetchMsg := m2.fetchInstancesCmd("processDefinitionKey", "k1")()
	res2, _ := m2.Update(fetchMsg)
	m3 := res2.(model)

	// Enter -> variables
	res3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m4 := res3.(model)
	if m4.viewMode != "variables" {
		t.Fatalf("expected variables, got %s", m4.viewMode)
	}
	// simulate fetchVariables
	fetchMsg2 := m4.fetchVariablesCmd("i1")()
	res4, _ := m4.Update(fetchMsg2)
	m5 := res4.(model)

	// Press Enter two extra times in variables view (should NOT push stack)
	_, _ = m5.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, _ = m5.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// navigation stack should contain two saved states: definitions->instances and instances->variables
	if len(m5.navigationStack) != 2 {
		t.Fatalf("expected navigationStack size 2 after drilldowns, got %d", len(m5.navigationStack))
	}

	// Press Esc once: should go back to instances and leave one saved state (definitions->instances)
	res5, _ := m5.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m6 := res5.(model)
	if m6.viewMode != "instances" {
		t.Fatalf("expected instances after Esc, got %s", m6.viewMode)
	}
	if len(m6.navigationStack) != 1 {
		t.Fatalf("expected navigationStack size 1 after first Esc, got %d", len(m6.navigationStack))
	}

	// Press Esc second time: should return to definitions and have empty stack
	res6, _ := m6.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m7 := res6.(model)
	if m7.viewMode != "definitions" {
		t.Fatalf("expected definitions after second Esc, got %s", m7.viewMode)
	}
	if len(m7.navigationStack) != 0 {
		t.Fatalf("expected navigationStack empty after second Esc, got %d", len(m7.navigationStack))
	}
}
