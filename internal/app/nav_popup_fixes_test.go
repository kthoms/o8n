package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kthoms/o8n/internal/config"
)

// ── T1: Esc back restores currentRoot so title shows correct count ────────────

func TestBackNavRestoresCurrentRoot(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.currentRoot = "process-definition"
	m.viewMode = "process-definition"
	m.breadcrumb = []string{"process-definition"}
	m.pageTotals["process-definition"] = 5

	// Push a drilldown state
	prev := viewState{
		viewMode:   "process-definition",
		breadcrumb: []string{"process-definition"},
	}
	m.navigationStack = []viewState{prev}
	m.currentRoot = "process-instance"
	m.viewMode = "process-instance"
	m.breadcrumb = []string{"process-definition", "process-instance"}

	m2, _ := sendKeyString(m, "esc")

	if m2.currentRoot != "process-definition" {
		t.Errorf("expected currentRoot=process-definition after esc, got %q", m2.currentRoot)
	}
	if m2.viewMode != "process-definition" {
		t.Errorf("expected viewMode=process-definition after esc, got %q", m2.viewMode)
	}
}

// ── T2: resolvePathParams substitutes {placeholder} in URL path ───────────────

func TestResolvePathParams_DirectMatch(t *testing.T) {
	url := "http://host/process-instance/{processInstanceId}/variables"
	params := map[string]string{"processInstanceId": "abc-123"}
	resolved, remaining := resolvePathParams(url, params)
	want := "http://host/process-instance/abc-123/variables"
	if resolved != want {
		t.Errorf("expected %q, got %q", want, resolved)
	}
	if len(remaining) != 0 {
		t.Errorf("expected no remaining params, got %v", remaining)
	}
}

func TestResolvePathParams_ParentIdFallback(t *testing.T) {
	url := "http://host/process-instance/{parentId}/variables"
	params := map[string]string{"processInstanceId": "abc-123"}
	resolved, remaining := resolvePathParams(url, params)
	want := "http://host/process-instance/abc-123/variables"
	if resolved != want {
		t.Errorf("expected %q, got %q", want, resolved)
	}
	if len(remaining) != 0 {
		t.Errorf("expected no remaining params after parentId fallback, got %v", remaining)
	}
}

func TestResolvePathParams_NoMatch_AppendsAsQueryParam(t *testing.T) {
	url := "http://host/process-definition"
	params := map[string]string{"processDefinitionKey": "my-key"}
	resolved, remaining := resolvePathParams(url, params)
	if resolved != url {
		t.Errorf("url should not change when no placeholder, got %q", resolved)
	}
	if remaining["processDefinitionKey"] != "my-key" {
		t.Errorf("expected key in remaining params, got %v", remaining)
	}
}

// ── T3: Arrow-right triggers drilldown ───────────────────────────────────────

func TestArrowRightDrillsDown(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.config.Tables = []config.TableDef{
		{
			Name: "process-definition",
			Columns: []config.ColumnDef{
				{Name: "id"},
			},
			Drilldown: []config.DrillDownDef{
				{Target: "process-instance", Param: "processDefinitionId", Column: "id"},
			},
		},
		{
			Name: "process-instance",
			Columns: []config.ColumnDef{
				{Name: "id"},
			},
		},
	}
	m.currentRoot = "process-definition"
	m.viewMode = "process-definition"
	m.breadcrumb = []string{"process-definition"}
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 20}})
	m.table.SetRows([]table.Row{{"def-1"}})
	m.rowData = []map[string]interface{}{{"id": "def-1"}}

	m2, _ := sendKeyString(m, "right")

	if len(m2.navigationStack) == 0 {
		t.Error("expected drilldown push to navigationStack after right-arrow")
	}
}

func TestArrowRightIgnoredInPopup(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.popup.mode = popupModeContext

	m2, _ := sendKeyString(m, "right")

	// popup should remain open, no drilldown
	if m2.popup.mode == popupModeNone {
		t.Error("expected popup to remain open after right-arrow in popup mode")
	}
}

// ── T4: Popup scroll — no "…N more" truncation ───────────────────────────────

func TestPopupScrollOffset_DownMovesOffset(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	// Create enough contexts to trigger scrolling (more than 8)
	for i := 0; i < 12; i++ {
		m.rootContexts = append(m.rootContexts, strings.Repeat("ctx-", 1)+string(rune('a'+i)))
	}
	m.popup.mode = popupModeContext
	m.popup.cursor = 7 // at maxShow boundary

	m2, _ := sendKeyString(m, "down")

	// cursor should advance
	if m2.popup.cursor != 8 {
		t.Errorf("expected cursor=8, got %d", m2.popup.cursor)
	}
	// offset should have scrolled to keep cursor visible
	if m2.popup.offset == 0 {
		t.Error("expected popup.offset > 0 when cursor moved past maxShow")
	}
}

func TestPopupViewNoTruncation(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 40
	for i := 0; i < 15; i++ {
		m.rootContexts = append(m.rootContexts, strings.Repeat("ctx-", 1)+string(rune('a'+i)))
	}
	m.popup.mode = popupModeContext
	m.popup.cursor = -1

	out := m.View()

	if strings.Contains(out, "more") {
		t.Error("expected no '...N more' truncation text in popup view")
	}
}

// ── T5: Context switch no longer falls back to process-definition ─────────────

func TestFetchForRoot_NoFallback(t *testing.T) {
	m := newTestModel(t)
	// "engines" is not in config.Tables — fetchForRoot should still return a non-nil cmd
	cmd := m.fetchForRoot("engines")
	if cmd == nil {
		t.Error("expected fetchForRoot to return non-nil cmd for unknown root")
	}
}

// ── T6: Search via popup ────────────────────────────────────────────────────

func TestSearchPopupOpensOnSlash(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetRows([]table.Row{{"row-a"}, {"row-b"}})

	m2, _ := sendKeyString(m, "/")

	if m2.popup.mode != popupModeSearch {
		t.Errorf("expected popup.mode=popupModeSearch, got %v", m2.popup.mode)
	}
	if m2.searchMode {
		t.Error("expected legacy searchMode to be false in new popup search mode")
	}
}

func TestSearchPopupFiltersRowsLive(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetRows([]table.Row{{"invoice-1"}, {"payment-2"}, {"invoice-3"}})

	m2, _ := sendKeyString(m, "/")
	// type "inv"
	m3, _ := sendKeyString(m2, "i")
	m4, _ := sendKeyString(m3, "n")
	m5, _ := sendKeyString(m4, "v")

	if len(m5.table.Rows()) != 2 {
		t.Errorf("expected 2 rows after typing 'inv', got %d", len(m5.table.Rows()))
	}
}

func TestSearchPopupEscRestoresRows(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetRows([]table.Row{{"a"}, {"b"}, {"c"}})

	m2, _ := sendKeyString(m, "/")
	m3, _ := sendKeyString(m2, "a")
	m4, _ := sendKeyString(m3, "esc")

	if m4.popup.mode != popupModeNone {
		t.Error("expected popup closed after Esc")
	}
	if len(m4.table.Rows()) != 3 {
		t.Errorf("expected 3 original rows after Esc, got %d", len(m4.table.Rows()))
	}
}

func TestSearchPopupEnterLocksFilter(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetRows([]table.Row{{"alpha"}, {"nothing"}, {"nowhere"}})

	m2, _ := sendKeyString(m, "/")
	// type "alpha" to match only the first row
	for _, ch := range "alpha" {
		m2, _ = sendKeyString(m2, string(ch))
	}
	m3, _ := sendKeyString(m2, "enter")

	if m3.popup.mode != popupModeNone {
		t.Error("expected popup closed after Enter")
	}
	// Filter is locked: rows remain filtered
	if len(m3.table.Rows()) != 1 {
		t.Errorf("expected 1 filtered row locked after Enter, got %d", len(m3.table.Rows()))
	}
}

func TestSearchPopupNotOpenedWhenPopupActive(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.popup.mode = popupModeContext

	m2, _ := sendKeyString(m, "/")

	if m2.popup.mode != popupModeContext {
		t.Errorf("expected popup mode unchanged (context), got %v", m2.popup.mode)
	}
	// "/" typed into context input
	if m2.popup.input != "/" {
		t.Errorf("expected '/' appended to popup input, got %q", m2.popup.input)
	}
}
