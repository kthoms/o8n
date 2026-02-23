package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kthoms/o8n/internal/config"
)

func TestGenericFetchUsesTableDefAndCount(t *testing.T) {
	// Server returns items and count for /widgets
	total := 5
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/widgets":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]interface{}{{"id": "w1", "name": "one"}, {"id": "w2", "name": "two"}})
			return
		case "/widgets/count":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"count": total})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: server.URL}},
		Tables:       []config.TableDef{{Name: "widgets", Columns: []config.ColumnDef{{Name: "id"}, {Name: "name"}}}},
	}

	m := newModel(cfg)
	cmd := m.fetchForRoot("widgets")
	if cmd == nil {
		t.Fatalf("expected fetchForRoot to return a cmd for widgets")
	}
	msg := cmd()
	res, _ := m.Update(msg)
	m2 := res.(model)

	if tot, ok := m2.pageTotals["widgets"]; !ok || tot != total {
		t.Fatalf("expected page total %d for widgets, got %v ok=%v", total, tot, ok)
	}

	rows := m2.table.Rows()
	if len(rows) == 0 {
		t.Fatalf("expected rows for widgets after fetch")
	}
	found := false
	for _, r := range rows {
		for _, c := range r {
			if c == "w1" || c == "one" {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatalf("expected widget id or name present in rows, got %v", rows)
	}
}

// TestGenericFetchUsesApiPath verifies that TableDef.ApiPath is used for the fetch URL.
func TestGenericFetchUsesApiPath(t *testing.T) {
	total := 3
	hitList := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitList[r.URL.Path]++
		switch r.URL.Path {
		case "/history/process-instance":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]interface{}{{"id": "h1"}})
			return
		case "/history/process-instance/count":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"count": total})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: server.URL}},
		Tables: []config.TableDef{{
			Name:      "history-process-instance",
			ApiPath:   "/history/process-instance",
			CountPath: "/history/process-instance/count",
			Columns:   []config.ColumnDef{{Name: "id"}},
		}},
	}

	m := newModel(cfg)
	cmd := m.fetchGenericCmd("history-process-instance")
	if cmd == nil {
		t.Fatalf("expected fetchGenericCmd to return a cmd")
	}
	msg := cmd()
	res, _ := m.Update(msg)
	m2 := res.(model)

	if hitList["/history/process-instance"] == 0 {
		t.Errorf("expected fetch to hit /history/process-instance, got hits: %v", hitList)
	}
	if hitList["/history/process-instance/count"] == 0 {
		t.Errorf("expected count fetch to hit /history/process-instance/count, got hits: %v", hitList)
	}
	if tot, ok := m2.pageTotals["history-process-instance"]; !ok || tot != total {
		t.Errorf("expected page total %d, got %v ok=%v", total, tot, ok)
	}
}

// TestGenericFetchCountPathFallback verifies count URL falls back to {api_path}/count when CountPath is not set.
func TestGenericFetchCountPathFallback(t *testing.T) {
	hitList := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitList[r.URL.Path]++
		switch r.URL.Path {
		case "/custom-path":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]interface{}{{"id": "x1"}})
			return
		case "/custom-path/count":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"count": 7})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Environments: map[string]config.Environment{"local": {URL: server.URL}},
		Tables: []config.TableDef{{
			Name:    "my-resource",
			ApiPath: "/custom-path",
			// CountPath intentionally not set — should derive /custom-path/count
			Columns: []config.ColumnDef{{Name: "id"}},
		}},
	}

	m := newModel(cfg)
	cmd := m.fetchGenericCmd("my-resource")
	msg := cmd()
	res, _ := m.Update(msg)
	m2 := res.(model)

	if hitList["/custom-path/count"] == 0 {
		t.Errorf("expected count fetch to hit /custom-path/count, got hits: %v", hitList)
	}
	if tot := m2.pageTotals["my-resource"]; tot != 7 {
		t.Errorf("expected page total 7, got %d", tot)
	}
}

func TestGenericFetchInfersColumnsWhenNoTableDef(t *testing.T) {
	// Server returns objects with keys a/b for /misc
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/misc":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]interface{}{{"a": 1, "b": "x"}})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	cfg := &config.Config{Environments: map[string]config.Environment{"local": {URL: server.URL}}}
	m := newModel(cfg)

	// call fetchGenericCmd directly for a root without a TableDef
	cmd := m.fetchGenericCmd("misc")
	if cmd == nil {
		t.Fatalf("expected fetchGenericCmd to return a cmd")
	}
	msg := cmd()
	res, _ := m.Update(msg)
	m2 := res.(model)

	cols := m2.table.Columns()
	if len(cols) < 2 {
		t.Fatalf("expected inferred columns, got %v", cols)
	}
	// titles are uppercased keys
	if cols[0].Title != "A" && cols[1].Title != "B" {
		// at least ensure keys present in titles in some order
		t.Fatalf("unexpected inferred column titles: %v", cols)
	}
}
