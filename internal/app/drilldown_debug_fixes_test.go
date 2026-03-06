package app

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kthoms/o6n/internal/config"
)

// ── AC-1..AC-4: drilldown target fields present in o6n-cfg.yaml ──────────────

// TestDrilldownTargets_Config verifies that the four drilldown blocks in
// o6n-cfg.yaml each carry a non-empty target field. Missing targets cause
// findTableDef("") to match the first config table (external-task) and show
// the wrong resource after Enter.
func TestDrilldownTargets_Config(t *testing.T) {
	appCfg, err := config.LoadAppConfig("o6n-cfg.yaml")
	if err != nil {
		t.Fatalf("load o6n-cfg.yaml: %v", err)
	}

	want := map[string]string{
		"process-definition": "process-instance",
		"process-instance":   "process-variables",
		"job-definition":     "job",
		"deployment":         "process-definition",
	}

	for _, tbl := range appCfg.Tables {
		if expected, ok := want[tbl.Name]; ok {
			if tbl.Drilldown == nil {
				t.Errorf("table %q: missing drilldown block", tbl.Name)
				continue
			}
			if tbl.Drilldown.Target != expected {
				t.Errorf("table %q drilldown.target: got %q, want %q",
					tbl.Name, tbl.Drilldown.Target, expected)
			}
		}
	}
}

// ── AC-5: Enter on process-definition navigates to process-instance ───────────

func TestEnterOnProcessDefinition_NavigatesToInstance(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: "http://localhost:8080"}},
		Tables: []config.TableDef{
			{
				Name:    "process-definition",
				Columns: []config.ColumnDef{{Name: "id"}, {Name: "key"}},
				Drilldown: &config.DrillDownDef{
					Target: "process-instance",
					Param:  "processDefinitionId",
					Column: "id",
				},
			},
			{
				Name:    "process-instance",
				Columns: []config.ColumnDef{{Name: "id"}},
			},
		},
	}
	m := newModel(cfg)
	m.splashActive = false
	m.currentRoot = "process-definition"
	m.viewMode = "process-definition"
	m.breadcrumb = []string{"process-definition"}
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 20}, {Title: "KEY", Width: 20}})
	m.table.SetRows([]table.Row{{"def-abc", "invoice"}})
	m.rowData = []map[string]interface{}{{"id": "def-abc", "key": "invoice"}}
	m.table.SetCursor(0)

	m2, _ := sendKeyString(m, "enter")

	if m2.currentRoot != "process-instance" {
		t.Errorf("currentRoot: got %q, want %q", m2.currentRoot, "process-instance")
	}
	if m2.genericParams["processDefinitionId"] == "" {
		t.Errorf("expected genericParams[processDefinitionId] to be set, got %v", m2.genericParams)
	}
}

// ── AC-6: Enter on process-instance navigates to process-variables ─────────────

func TestEnterOnProcessInstance_NavigatesToVariables(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: "http://localhost:8080"}},
		Tables: []config.TableDef{
			{
				Name:    "process-instance",
				Columns: []config.ColumnDef{{Name: "id"}, {Name: "businessKey"}},
				Drilldown: &config.DrillDownDef{
					Target: "process-variables",
					Param:  "processInstanceId",
					Column: "id",
				},
			},
			{
				Name:    "process-variables",
				Columns: []config.ColumnDef{{Name: "name"}, {Name: "value"}},
			},
		},
	}
	m := newModel(cfg)
	m.splashActive = false
	m.currentRoot = "process-instance"
	m.viewMode = "process-instance"
	m.breadcrumb = []string{"process-instance"}
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 20}, {Title: "BUSINESSKEY", Width: 20}})
	m.table.SetRows([]table.Row{{"inst-xyz", "order-1"}})
	m.rowData = []map[string]interface{}{{"id": "inst-xyz", "businessKey": "order-1"}}
	m.table.SetCursor(0)

	m2, _ := sendKeyString(m, "enter")

	if m2.currentRoot != "process-variables" {
		t.Errorf("currentRoot: got %q, want %q", m2.currentRoot, "process-variables")
	}
	if m2.genericParams["processInstanceId"] == "" {
		t.Errorf("expected genericParams[processInstanceId] to be set, got %v", m2.genericParams)
	}
}

// ── AC-13: Screen capture files are ANSI-free ────────────────────────────────

func TestScreenDumpIsAnsiFree(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	// Run in a temp dir to avoid polluting the real debug/ directory.
	origDir, _ := os.Getwd()
	tmp := t.TempDir()
	_ = os.Chdir(tmp)
	defer os.Chdir(origDir)

	// Restore lastRenderedView after test to avoid contaminating other tests.
	origView := lastRenderedView
	defer func() { lastRenderedView = origView }()

	// Inject ANSI content into lastRenderedView to simulate a real rendered frame.
	lastRenderedView = "\x1b[1mo6n 0.1.0\x1b[0m │ \x1b[38;2;195;64;67m✗\x1b[0m"

	m.Update(errMsg{err: fmt.Errorf("GET http://x/y: simulated error")})

	// Locate the written screen-*.txt file.
	entries, err := os.ReadDir("debug")
	if err != nil {
		t.Fatalf("debug dir not created: %v", err)
	}
	var screenFile string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "screen-") && strings.HasSuffix(e.Name(), ".txt") {
			screenFile = "debug/" + e.Name()
			break
		}
	}
	if screenFile == "" {
		t.Fatal("expected debug/screen-*.txt to be created by errMsg handler")
	}

	data, err := os.ReadFile(screenFile)
	if err != nil {
		t.Fatalf("read screen file: %v", err)
	}
	if strings.Contains(string(data), "\x1b") {
		t.Errorf("screen dump contains ANSI escape sequences (\\x1b); should be stripped before writing")
	}
}

// ── AC-15: Context switch with mismatched column count does not panic ─────────

func TestContextSwitch_DifferentColumnCount_NoPanic(t *testing.T) {
	// Table "alpha" has 3 columns; table "beta" has 4 columns.
	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: "http://localhost:8080"}},
		Tables: []config.TableDef{
			{
				Name: "alpha",
				Columns: []config.ColumnDef{
					{Name: "id"}, {Name: "name"}, {Name: "status"},
				},
			},
			{
				Name: "beta",
				Columns: []config.ColumnDef{
					{Name: "id"}, {Name: "type"}, {Name: "total"}, {Name: "tenant"},
				},
			},
		},
	}
	m := newModel(cfg)
	m.splashActive = false
	m.rootContexts = []string{"alpha", "beta"}
	// Wide terminal so buildColumnsFor keeps all columns rather than hiding them.
	m.lastWidth = 120
	m.paneWidth = 120

	// Load alpha with 3-cell rows.
	m.table.SetColumns([]table.Column{
		{Title: "ID", Width: 10},
		{Title: "NAME", Width: 10},
		{Title: "STATUS", Width: 10},
	})
	m.table.SetRows([]table.Row{
		{"a1", "foo", "active"},
		{"a2", "bar", "inactive"},
	})
	m.currentRoot = "alpha"
	m.viewMode = "alpha"

	// Open context modal and switch to "beta" (which has 4 columns).
	m.activeModal = ModalContextSwitcher
	m.popup.input = "beta"
	m.popup.cursor = -1

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("context switch panicked: %v", r)
		}
	}()

	m2, _ := sendKeyString(m, "enter")

	if m2.currentRoot != "beta" {
		t.Errorf("currentRoot: got %q, want %q", m2.currentRoot, "beta")
	}
	if len(m2.table.Columns()) != 4 {
		t.Errorf("expected 4 columns after switch to beta, got %d", len(m2.table.Columns()))
	}
}

// ── AC-7/AC-8: HTTP requests logged when debugEnabled=true ───────────────────

func TestFetchGenericCmd_DebugLogsHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/count") {
			w.Write([]byte(`{"count":0}`))
		} else {
			w.Write([]byte(`[]`))
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: server.URL}},
		Tables:       []config.TableDef{{Name: "widget", Columns: []config.ColumnDef{{Name: "id"}}}},
	}
	m := newModel(cfg)
	m.debugEnabled = true

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	m.fetchGenericCmd("widget")()

	logged := buf.String()
	// AC-7: main request logged
	if !strings.Contains(logged, "[http] GET") || !strings.Contains(logged, server.URL+"/widget") {
		t.Errorf("AC-7: expected [http] GET %s/widget in log, got: %s", server.URL, logged)
	}
	// AC-8: count request logged
	if !strings.Contains(logged, "(count)") {
		t.Errorf("AC-8: expected (count) in log, got: %s", logged)
	}
}

// ── AC-9: No HTTP log when debugEnabled=false ─────────────────────────────────

func TestFetchGenericCmd_NoDebugNoLog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/count") {
			w.Write([]byte(`{"count":0}`))
		} else {
			w.Write([]byte(`[]`))
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: server.URL}},
		Tables:       []config.TableDef{{Name: "widget", Columns: []config.ColumnDef{{Name: "id"}}}},
	}
	m := newModel(cfg)
	m.debugEnabled = false

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	m.fetchGenericCmd("widget")()

	if strings.Contains(buf.String(), "[http]") {
		t.Errorf("AC-9: expected no [http] log when debugEnabled=false, got: %s", buf.String())
	}
}
