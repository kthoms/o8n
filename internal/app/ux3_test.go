package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kthoms/o8n/internal/config"
)

// ── T2: genericLoadedMsg adds ▶ prefix for drilldown tables ──────────────────

func TestGenericLoadedMsgAddsDrilldownPrefix(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	// Inject a table def with Drilldown configured
	m.config.Tables = []config.TableDef{
		{
			Name: "widgets",
			Columns: []config.ColumnDef{
				{Name: "id", Type: "text"},
				{Name: "name", Type: "text"},
			},
			Drilldown: []config.DrillDownDef{
				{Target: "widget-details", Param: "widgetId"},
			},
		},
	}
	m.paneWidth = 120

	msg := genericLoadedMsg{
		root: "widgets",
		items: []map[string]interface{}{
			{"id": "w1", "name": "Widget One"},
			{"id": "w2", "name": "Widget Two"},
		},
	}

	res, _ := m.Update(msg)
	m2 := res.(model)

	rows := m2.table.Rows()
	if len(rows) == 0 {
		t.Fatal("expected rows after genericLoadedMsg")
	}
	if !strings.HasPrefix(rows[0][0], "▶ ") {
		t.Errorf("expected first cell to start with '▶ ', got %q", rows[0][0])
	}
}

func TestGenericLoadedMsgNoDrilldownPrefixWhenNoDrilldown(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	// Table def without Drilldown
	m.config.Tables = []config.TableDef{
		{
			Name: "tags",
			Columns: []config.ColumnDef{
				{Name: "id", Type: "text"},
				{Name: "label", Type: "text"},
			},
		},
	}
	m.paneWidth = 120

	msg := genericLoadedMsg{
		root: "tags",
		items: []map[string]interface{}{
			{"id": "t1", "label": "Alpha"},
		},
	}

	res, _ := m.Update(msg)
	m2 := res.(model)

	rows := m2.table.Rows()
	if len(rows) == 0 {
		t.Fatal("expected rows after genericLoadedMsg")
	}
	if strings.HasPrefix(rows[0][0], "▶ ") {
		t.Errorf("expected no '▶ ' prefix for table without drilldown, got %q", rows[0][0])
	}
}

// ── T3: renderCompactHeader shows ↺ badge when autoRefresh is true ───────────

func TestAutoRefreshHeaderIndicatorPresent(t *testing.T) {
	m := newTestModel(t)
	m.autoRefresh = true
	m.lastWidth = 120

	out := m.renderCompactHeader(120)

	if !strings.Contains(out, "↺") {
		t.Errorf("expected '↺' in header when autoRefresh=true, got:\n%s", out)
	}
}

func TestAutoRefreshHeaderIndicatorAbsent(t *testing.T) {
	m := newTestModel(t)
	m.autoRefresh = false
	m.lastWidth = 120

	out := m.renderCompactHeader(120)

	if strings.Contains(out, "↺") {
		t.Errorf("expected no '↺' in header when autoRefresh=false, got:\n%s", out)
	}
}

// ── T5: buildColumnsFor appends ✎ to editable column titles ─────────────────

func TestBuildColumnsForEditableMarker(t *testing.T) {
	m := newTestModel(t)

	m.config.Tables = []config.TableDef{
		{
			Name: "forms",
			Columns: []config.ColumnDef{
				{Name: "id", Type: "text"},
				{Name: "value", Type: "text", Editable: true},
			},
		},
	}

	cols := m.buildColumnsFor("forms", 200)

	found := false
	for _, c := range cols {
		if strings.HasSuffix(c.Title, " ✎") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected at least one column title ending with ' ✎', got: %v", cols)
	}
}

func TestBuildColumnsForNonEditableNoMarker(t *testing.T) {
	m := newTestModel(t)

	m.config.Tables = []config.TableDef{
		{
			Name: "forms",
			Columns: []config.ColumnDef{
				{Name: "id", Type: "text"},
				{Name: "readonly", Type: "text", Editable: false},
			},
		},
	}

	cols := m.buildColumnsFor("forms", 200)

	for _, c := range cols {
		if strings.HasSuffix(c.Title, " ✎") {
			t.Errorf("expected no ' ✎' marker on non-editable column, got title: %q", c.Title)
		}
	}
}

// ── T6: friendlyError translates network errors ───────────────────────────────

func TestFriendlyErrorConnectionRefused(t *testing.T) {
	err := errors.New("dial tcp: connection refused")
	out := friendlyError("local", err)
	if !strings.Contains(out, "Cannot connect") {
		t.Errorf("expected 'Cannot connect' for connection refused, got %q", out)
	}
}

func TestFriendlyErrorTimeout(t *testing.T) {
	err := errors.New("request timeout: context deadline exceeded")
	out := friendlyError("local", err)
	if !strings.Contains(out, "timed out") {
		t.Errorf("expected 'timed out' for timeout error, got %q", out)
	}
}

func TestFriendlyErrorDeadlineExceeded(t *testing.T) {
	err := errors.New("deadline exceeded waiting for response")
	out := friendlyError("local", err)
	if !strings.Contains(out, "timed out") {
		t.Errorf("expected 'timed out' for deadline exceeded, got %q", out)
	}
}

func TestFriendlyErrorTLS(t *testing.T) {
	err := errors.New("x509: certificate signed by unknown authority")
	out := friendlyError("local", err)
	if !strings.Contains(out, "TLS") {
		t.Errorf("expected 'TLS' for x509 error, got %q", out)
	}
}

func TestFriendlyErrorNoSuchHost(t *testing.T) {
	err := errors.New("dial tcp: no such host")
	out := friendlyError("staging", err)
	if !strings.Contains(out, "Unknown host") {
		t.Errorf("expected 'Unknown host' for no such host, got %q", out)
	}
}

func TestFriendlyErrorUnknownPassthrough(t *testing.T) {
	msg := "some completely unknown error"
	err := errors.New(msg)
	out := friendlyError("local", err)
	if out != msg {
		t.Errorf("expected original message for unknown error, got %q", out)
	}
}

// ── T8: healthTickMsg dispatches health check command ────────────────────────

func TestHealthTickMsgReturnsNonNilCmd(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	_, cmd := m.Update(healthTickMsg{})

	if cmd == nil {
		t.Error("expected non-nil command from healthTickMsg handler")
	}
}

// ── T9: normalizeRows returns "No results" for empty input ───────────────────

func TestNormalizeRowsNilInput(t *testing.T) {
	result := normalizeRows(nil, 3)
	if len(result) != 1 {
		t.Fatalf("expected 1 row for nil input, got %d", len(result))
	}
	if result[0][0] != "No results" {
		t.Errorf("expected first cell 'No results', got %q", result[0][0])
	}
	if len(result[0]) != 3 {
		t.Errorf("expected row length 3, got %d", len(result[0]))
	}
}

func TestNormalizeRowsEmptySlice(t *testing.T) {
	result := normalizeRows([]table.Row{}, 2)
	if len(result) != 1 {
		t.Fatalf("expected 1 row for empty slice, got %d", len(result))
	}
	if result[0][0] != "No results" {
		t.Errorf("expected first cell 'No results', got %q", result[0][0])
	}
}

func TestNormalizeRowsNonEmptyUnchanged(t *testing.T) {
	rows := []table.Row{{"hello", "world"}}
	result := normalizeRows(rows, 2)
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	if result[0][0] != "hello" {
		t.Errorf("expected 'hello', got %q", result[0][0])
	}
}

// ── T1: refreshMsg uses fetchForRoot(currentRoot) ────────────────────────────

func TestRefreshMsgWithAutoRefreshReturnsCmd(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.autoRefresh = true
	m.currentRoot = "process-definition"

	_, cmd := m.Update(refreshMsg{})

	if cmd == nil {
		t.Error("expected non-nil command from refreshMsg when autoRefresh=true")
	}
}

func TestRefreshMsgWithoutAutoRefreshReturnsNilCmd(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.autoRefresh = false

	_, cmd := m.Update(refreshMsg{})

	if cmd != nil {
		t.Error("expected nil command from refreshMsg when autoRefresh=false")
	}
}
