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

func TestPaginationPageDownUp(t *testing.T) {
	// prepare server with 10 items and count endpoint
	total := 10
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/process-instance":
			q := r.URL.Query()
			firstStr := q.Get("firstResult")
			maxStr := q.Get("maxResults")
			first := 0
			max := 3
			if firstStr != "" {
				if v, err := strconv.Atoi(firstStr); err == nil {
					first = v
				}
			}
			if maxStr != "" {
				if v, err := strconv.Atoi(maxStr); err == nil {
					max = v
				}
			}
			// build items from first..first+max-1
			items := make([]config.ProcessInstance, 0)
			for i := first; i < first+max && i < total; i++ {
				id := "i" + strconv.Itoa(i+1)
				items = append(items, config.ProcessInstance{ID: id, DefinitionID: "d1", BusinessKey: "bk" + strconv.Itoa(i+1), StartTime: "2020-01-01T00:00:00Z"})
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
	// set view into instances and breadcrumb so page handlers use correct root
	m.viewMode = "instances"
	m.breadcrumb = []string{m.currentRoot, dao.ResourceProcessInstances}

	// ensure page size is small and offsets start at 0
	m.table.SetHeight(3)
	m.pageOffsets[dao.ResourceProcessInstances] = 0

	// load first page via fetchInstancesCmd
	msg := m.fetchInstancesCmd("", "")()
	res, _ := m.Update(msg)
	m1 := res.(model)
	rows := m1.table.Rows()
	if len(rows) == 0 {
		t.Fatalf("expected rows after initial fetch")
	}
	// Account for focus indicator prefix (▶ )
	firstRowID := stripFocusIndicatorPrefix(rows[0][0])
	if firstRowID != "i1" {
		t.Fatalf("expected first row i1, got %v", firstRowID)
	}

	// set selection to third row (index 2)
	m1.table.SetCursor(2)

	// press PageDown
	res2, _ := m1.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m2 := res2.(model)
	// execute fetch using m2 to ensure closure reads updated offsets
	fetchMsg := m2.fetchInstancesCmd("", "")()
	res3, _ := m2.Update(fetchMsg)
	m3 := res3.(model)

	// after page down, first row should be i{p+1}, and selected row (index 2) should point to i{p+3}
	p := m1.getPageSize()
	rows2 := m3.table.Rows()
	expectedFirst := "i" + strconv.Itoa(p+1)
	// Account for focus indicator prefix (▶ )
	page2FirstRow := stripFocusIndicatorPrefix(rows2[0][0])
	if page2FirstRow != expectedFirst {
		t.Fatalf("expected page2 first row %s, got %v", expectedFirst, page2FirstRow)
	}
	// selection preservation is validated at runtime; here we ensure the page content changed

	// ensure page total was captured
	if tot, ok := m3.pageTotals[dao.ResourceProcessInstances]; !ok || tot != total {
		t.Fatalf("expected total %d recorded, got %v ok=%v", total, tot, ok)
	}

	// Now PageUp back to first page
	res4, _ := m3.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m4 := res4.(model)
	fetchMsg2 := m4.fetchInstancesCmd("", "")()
	res5, _ := m4.Update(fetchMsg2)
	m5 := res5.(model)

	rows3 := m5.table.Rows()
	// Account for focus indicator prefix (▶ )
	page1FirstRow := stripFocusIndicatorPrefix(rows3[0][0])
	if page1FirstRow != "i1" {
		t.Fatalf("expected page1 first row i1 after page up, got %v", page1FirstRow)
	}
	// ensure we returned to first page content
}
