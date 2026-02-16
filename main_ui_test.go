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
	fetchMsg := m2.fetchInstancesCmd("k1")()
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
