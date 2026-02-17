package main

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
		Tables:       []config.TableDef{{Name: "widgets", Columns: []config.ColumnDef{{Name: "id", Visible: true}, {Name: "name", Visible: true}}}},
	}

	m := newModel(cfg)
	// fetch via generic fetch path for known table
	cmd := m.fetchForRoot("widgets")
	if cmd == nil {
		t.Fatalf("expected fetchForRoot to return a cmd for widgets")
	}
	msg := cmd()
	// deliver message to Update and inspect model
	res, _ := m.Update(msg)
	m2 := res.(model)

	// page total should be recorded
	if tot, ok := m2.pageTotals["widgets"]; !ok || tot != total {
		t.Fatalf("expected page total %d for widgets, got %v ok=%v", total, tot, ok)
	}

	// rows should contain expected values (check that any cell contains the id or name)
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
