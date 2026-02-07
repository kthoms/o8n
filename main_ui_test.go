package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplyDataPopulatesTable(t *testing.T) {
	cfg := &Config{Environments: map[string]Environment{"local": {URL: "http://example"}}}
	m := newModel(cfg)
	defs := []ProcessDefinition{{ID: "d1", Key: "k1", Name: "One"}}
	insts := []ProcessInstance{{ID: "i1", DefinitionID: "d1", BusinessKey: "bk1", StartTime: "2020-01-01T00:00:00Z"}}
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

	cfg := &Config{Environments: map[string]Environment{"local": {URL: server.URL}}}
	m := newModel(cfg)
	// populate list with two definitions so moving down changes selection
	defs := []ProcessDefinition{{ID: "d1", Key: "k1", Name: "One"}, {ID: "d2", Key: "k2", Name: "Two"}}
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
			json.NewEncoder(w).Encode([]ProcessDefinition{{ID: "d1", Key: "k1", Name: "One"}})
			return
		}
		if r.URL.Path == "/process-instance" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]ProcessInstance{{ID: "i1", DefinitionID: "d1", BusinessKey: "bk1", StartTime: "2020-01-01T00:00:00Z"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &Config{Environments: map[string]Environment{"local": {URL: server.URL}}}
	m := newModel(cfg)
	// pre-populate list so selectedKey will be k1 after selection
	m.applyData([]ProcessDefinition{{ID: "d1", Key: "k1", Name: "One"}}, nil)
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
	cfg := &Config{Environments: map[string]Environment{"local": {URL: "http://example"}}}
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
