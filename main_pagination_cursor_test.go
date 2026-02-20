package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/dao"
)

// Test that the cursor (selected row index) is preserved across page down/up
// and that the selected row maps to the expected item given the page offset.
func TestCursorPreservedAcrossPages(t *testing.T) {
	total := 10
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/process-instance":
			q := r.URL.Query()
			first := 0
			max := 3
			if v := q.Get("firstResult"); v != "" {
				if n, err := strconv.Atoi(v); err == nil {
					first = n
				}
			}
			if v := q.Get("maxResults"); v != "" {
				if n, err := strconv.Atoi(v); err == nil {
					max = n
				}
			}
			items := make([]config.ProcessInstance, 0)
			for i := first; i < first+max && i < total; i++ {
				id := "i" + strconv.Itoa(i+1)
				items = append(items, config.ProcessInstance{ID: id, DefinitionID: "d1"})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(items)
			return
		case "/process-instance/count":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"count": total})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	cfg := &config.Config{Environments: map[string]config.Environment{"local": {URL: server.URL}}}
	m := newModel(cfg)
	m.viewMode = "instances"
	m.breadcrumb = []string{m.currentRoot, dao.ResourceProcessInstances}

	// deterministic page size
	m.table.SetHeight(3)
	m.pageOffsets[dao.ResourceProcessInstances] = 0

	// initial fetch
	msg := m.fetchInstancesCmd("", "")()
	res, _ := m.Update(msg)
	m1 := res.(model)
	// set selection to index 2
	m1.table.SetCursor(2)
	preCursor := m1.table.Cursor()

	// page down
	res2, _ := m1.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m2 := res2.(model)
	// trigger fetch using the updated model so it reads new offsets
	fetch := m2.fetchInstancesCmd("", "")()
	res3, _ := m2.Update(fetch)
	m3 := res3.(model)

	// cursor index should be preserved
	postCursor := m3.table.Cursor()
	if preCursor != postCursor {
		t.Fatalf("expected cursor index preserved (%d==%d)", preCursor, postCursor)
	}

	// selected row should correspond to id = offset + index + 1
	offset := m2.pageOffsets[dao.ResourceProcessInstances]
	sel := m3.table.SelectedRow()
	if len(sel) == 0 {
		t.Fatalf("no selected row after page down")
	}
	expectedID := "i" + strconv.Itoa(offset+postCursor+1)
	// Account for focus indicator prefix (▶ )
	selectedID := stripFocusIndicatorPrefix(sel[0])
	if selectedID != expectedID {
		t.Fatalf("expected selected id %s, got %v", expectedID, selectedID)
	}
}

// Test that when paging to the final page with fewer rows than the current cursor
// the cursor is clamped to the last available row.
func TestCursorClampedAtEndPage(t *testing.T) {
	total := 7
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/process-instance":
			q := r.URL.Query()
			first := 0
			max := 4
			if v := q.Get("firstResult"); v != "" {
				if n, err := strconv.Atoi(v); err == nil {
					first = n
				}
			}
			if v := q.Get("maxResults"); v != "" {
				if n, err := strconv.Atoi(v); err == nil {
					max = n
				}
			}
			items := make([]config.ProcessInstance, 0)
			for i := first; i < first+max && i < total; i++ {
				id := "i" + strconv.Itoa(i+1)
				items = append(items, config.ProcessInstance{ID: id, DefinitionID: "d1"})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(items)
			return
		case "/process-instance/count":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"count": total})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	cfg := &config.Config{Environments: map[string]config.Environment{"local": {URL: server.URL}}}
	m := newModel(cfg)
	m.viewMode = "instances"
	m.breadcrumb = []string{m.currentRoot, dao.ResourceProcessInstances}

	// page size 4, total 7 -> pages: [1-4], [5-7]
	m.table.SetHeight(4)
	m.pageOffsets[dao.ResourceProcessInstances] = 0

	// initial fetch and set cursor to 3 (last index)
	msg := m.fetchInstancesCmd("", "")()
	res, _ := m.Update(msg)
	m1 := res.(model)
	m1.table.SetCursor(3)
	// page down to second page (offset 4)
	res2, _ := m1.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m2 := res2.(model)
	fetch := m2.fetchInstancesCmd("", "")()
	res3, _ := m2.Update(fetch)
	m3 := res3.(model)

	// now the last page has only 3 rows; cursor should be clamped to last index (2)
	if m3.table.Cursor() != 2 {
		t.Fatalf("expected cursor clamped to 2, got %d", m3.table.Cursor())
	}
}
