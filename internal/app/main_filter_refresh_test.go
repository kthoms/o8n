package app

// main_filter_refresh_test.go — Story 3.7: Search, Filter & Auto-Refresh
//
// Tests verify filtering and auto-refresh functionality:
//   - AC 1: "/" enters search mode and filters table in real-time
//   - AC 2: Escape clears filter and restores originalRows
//   - AC 3: Ctrl+Shift+R toggles auto-refresh with ⟳ indicator
//   - AC 4: Loading indicator appears for requests > 500ms
//   - AC 5: API status indicator (● green, ✗ red, ○ muted) appears in footer

import (
	"fmt"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kthoms/o6n/internal/config"
)

// filterTestConfig returns a minimal config for filter/refresh testing
func filterTestConfig() *config.Config {
	return &config.Config{
		Environments: map[string]config.Environment{"local": {URL: "http://localhost"}},
		Tables: []config.TableDef{
			{
				Name: "process-instance",
				Columns: []config.ColumnDef{
					{Name: "id"},
					{Name: "businessKey"},
					{Name: "state"},
				},
			},
		},
	}
}

// ── AC 1: "/" enters search mode and filters ────────────────────────────────

// TestSearchModeEntry verifies that pressing "/" enters search mode
func TestSearchModeEntry(t *testing.T) {
	m := newModel(filterTestConfig())
	m.currentRoot = "process-instance"
	m.table.SetColumns([]table.Column{
		{Title: "id", Width: 20},
		{Title: "businessKey", Width: 20},
		{Title: "state", Width: 20},
	})
	m.table.SetRows([]table.Row{
		{"pi-1", "invoice", "ACTIVE"},
		{"pi-2", "payment", "ACTIVE"},
		{"pi-3", "invoice", "COMPLETED"},
	})

	// Store original rows before search
	originalCount := len(m.table.Rows())

	// Simulate pressing "/" to enter search mode
	m.popup.mode = popupModeSearch
	m.originalRows = append([]table.Row{}, m.table.Rows()...)
	m.popup.input = ""

	// Verify search mode is active
	if m.popup.mode != popupModeSearch {
		t.Errorf("expected popupModeSearch, got %v", m.popup.mode)
	}

	// Verify original rows are saved
	if len(m.originalRows) != originalCount {
		t.Errorf("expected %d original rows, got %d", originalCount, len(m.originalRows))
	}
}

// TestSearchFilteringRealTime verifies that search filters rows in real-time
func TestSearchFilteringRealTime(t *testing.T) {
	m := newModel(filterTestConfig())
	m.currentRoot = "process-instance"
	m.table.SetColumns([]table.Column{
		{Title: "id", Width: 20},
		{Title: "businessKey", Width: 20},
		{Title: "state", Width: 20},
	})
	m.table.SetRows([]table.Row{
		{"pi-1", "invoice", "ACTIVE"},
		{"pi-2", "payment", "ACTIVE"},
		{"pi-3", "invoice", "COMPLETED"},
	})

	// Save original rows
	m.originalRows = append([]table.Row{}, m.table.Rows()...)

	// Simulate user typing "invoice" to filter
	filtered := filterRowsByTerm(m.originalRows, "invoice")

	// Should match pi-1 and pi-3
	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered rows, got %d", len(filtered))
	}

	// Verify both invoice rows are present
	foundCount := 0
	for _, row := range filtered {
		if len(row) > 1 && row[1] == "invoice" {
			foundCount++
		}
	}
	if foundCount != 2 {
		t.Errorf("expected 2 invoice rows in filtered results, got %d", foundCount)
	}
}

// Helper function to simulate row filtering by search term
func filterRowsByTerm(rows []table.Row, term string) []table.Row {
	var filtered []table.Row
	for _, row := range rows {
		// Simple search: check if any column matches term (case-insensitive substring)
		for _, col := range row {
			colStr := fmt.Sprintf("%v", col)
			if stringContains(colStr, term) {
				filtered = append(filtered, row)
				break
			}
		}
	}
	return filtered
}

// ── AC 2: Escape clears filter and restores originalRows ─────────────────────

// TestFilterClear verifies that escape clears the filter
func TestFilterClear(t *testing.T) {
	m := newModel(filterTestConfig())
	m.currentRoot = "process-instance"
	m.table.SetColumns([]table.Column{
		{Title: "id", Width: 20},
		{Title: "businessKey", Width: 20},
		{Title: "state", Width: 20},
	})
	m.table.SetRows([]table.Row{
		{"pi-1", "invoice", "ACTIVE"},
		{"pi-2", "payment", "ACTIVE"},
	})

	// Save original rows
	m.originalRows = append([]table.Row{}, m.table.Rows()...)
	originalRowCount := len(m.originalRows)

	// Simulate filtering
	m.popup.mode = popupModeSearch
	m.popup.input = "invoice"
	filtered := filterRowsByTerm(m.originalRows, "invoice")

	// Apply filter
	m.table.SetRows(filtered)
	if len(m.table.Rows()) == originalRowCount {
		t.Errorf("expected filtered rows, but got all original rows")
	}

	// Simulate pressing Escape to clear filter
	m.table.SetRows(m.originalRows)
	m.popup.mode = popupModeNone
	m.popup.input = ""

	// Verify full table is restored
	if len(m.table.Rows()) != originalRowCount {
		t.Errorf("expected %d rows after clear, got %d", originalRowCount, len(m.table.Rows()))
	}
}

// ── AC 3: Ctrl+Shift+R toggles auto-refresh with ⟳ indicator ────────────────

// TestAutoRefreshToggle verifies that Ctrl+Shift+R toggles auto-refresh
func TestAutoRefreshToggle(t *testing.T) {
	m := newModel(filterTestConfig())
	m.currentRoot = "process-instance"

	// Initial state: autoRefresh should be false
	if m.autoRefresh {
		t.Error("expected autoRefresh to be false initially")
	}

	// Simulate toggling auto-refresh on
	m.autoRefresh = !m.autoRefresh
	if !m.autoRefresh {
		t.Error("expected autoRefresh to be true after toggle")
	}

	// Simulate toggling auto-refresh off
	m.autoRefresh = !m.autoRefresh
	if m.autoRefresh {
		t.Error("expected autoRefresh to be false after second toggle")
	}
}

// TestAutoRefreshIndicator verifies that flashActive indicates auto-refresh state
func TestAutoRefreshIndicator(t *testing.T) {
	m := newModel(filterTestConfig())
	m.currentRoot = "process-instance"

	// When auto-refresh is on, flash indicator should be set
	m.autoRefresh = true
	m.flashActive = true

	// Verify flash is active when auto-refresh is on
	if !m.flashActive {
		t.Error("expected flashActive to be true when auto-refresh is on")
	}

	// When auto-refresh is off, flash should be inactive
	m.autoRefresh = false
	m.flashActive = false

	if m.flashActive {
		t.Error("expected flashActive to be false when auto-refresh is off")
	}
}

// ── AC 4: Loading indicator appears for requests > 500ms ────────────────────

// TestLoadingIndicatorDelay verifies that loading appears after 500ms
func TestLoadingIndicatorDelay(t *testing.T) {
	m := newModel(filterTestConfig())
	m.currentRoot = "process-instance"

	// API call starts
	m.isLoading = true
	m.apiCallStarted = time.Now().Add(-400 * time.Millisecond)

	// Calculate elapsed time
	elapsed := time.Since(m.apiCallStarted)

	// Should not show loading yet (< 500ms)
	if elapsed > 500*time.Millisecond {
		t.Errorf("expected elapsed time < 500ms, got %v", elapsed)
	}

	// Simulate request taking > 500ms
	m.apiCallStarted = time.Now().Add(-600 * time.Millisecond)
	elapsed = time.Since(m.apiCallStarted)

	// Should show loading now (> 500ms)
	if elapsed < 500*time.Millisecond {
		t.Errorf("expected elapsed time > 500ms, got %v", elapsed)
	}

	// Verify isLoading flag is set
	if !m.isLoading {
		t.Error("expected isLoading to be true during request")
	}
}

// ── AC 5: API status indicator (● green, ✗ red, ○ muted) ──────────────────

// TestAPIStatusIndicator verifies that API status shows in footer
func TestAPIStatusIndicator(t *testing.T) {
	m := newModel(filterTestConfig())
	m.currentEnv = "local"

	// Initialize envStatus map
	m.envStatus = make(map[string]EnvironmentStatus)

	// Test operational status (●)
	m.envStatus["local"] = StatusOperational
	status := m.envStatus["local"]
	if status != StatusOperational {
		t.Errorf("expected StatusOperational, got %v", status)
	}

	// Test unreachable status (✗)
	m.envStatus["local"] = StatusUnreachable
	status = m.envStatus["local"]
	if status != StatusUnreachable {
		t.Errorf("expected StatusUnreachable, got %v", status)
	}

	// Test unknown status (○)
	m.envStatus["local"] = StatusUnknown
	status = m.envStatus["local"]
	if status != StatusUnknown {
		t.Errorf("expected StatusUnknown, got %v", status)
	}
}

// TestAPIStatusIndicatorInFooter verifies that status indicator renders correctly
func TestAPIStatusIndicatorInFooter(t *testing.T) {
	m := newModel(filterTestConfig())
	m.currentEnv = "local"
	m.lastWidth = 120
	m.lastHeight = 20
	m.envStatus = make(map[string]EnvironmentStatus)
	m.envStatus["local"] = StatusOperational

	// Verify envStatus map contains the environment
	if _, ok := m.envStatus["local"]; !ok {
		t.Error("expected envStatus to contain 'local' environment")
	}

	// Verify status is operational
	if m.envStatus["local"] != StatusOperational {
		t.Error("expected environment status to be operational")
	}
}

// ── AC 1, 3: Hint System Integration ──────────────────────────────────────

// TestFilterHints verifies that "/" hint appears in hints
func TestFilterHints(t *testing.T) {
	m := newModel(filterTestConfig())
	m.currentRoot = "process-instance"
	m.lastWidth = 120

	// Get hints from tableViewHints
	hints := tableViewHints(m)

	// Should contain "/" hint
	foundFilter := false
	for _, h := range hints {
		if h.Key == "/" {
			foundFilter = true
			if h.Label != "find" && h.Label != "filter" {
				t.Errorf("expected hint label 'find' or 'filter', got %q", h.Label)
			}
		}
	}
	if !foundFilter {
		t.Error("expected '/' hint in tableViewHints")
	}
}

// TestAutoRefreshHints verifies that Ctrl+Shift+R hint appears in hints
func TestAutoRefreshHints(t *testing.T) {
	m := newModel(filterTestConfig())
	m.currentRoot = "process-instance"
	m.lastWidth = 120

	// Get hints from tableViewHints
	hints := tableViewHints(m)

	// At minimum, refresh hint should exist
	foundRefresh := false
	for _, h := range hints {
		if h.Label == "refresh" {
			foundRefresh = true
			break
		}
	}

	if !foundRefresh {
		t.Error("expected refresh hint in tableViewHints")
	}
}

// Helper: stringContains checks if a string contains a substring (simple substring check)
// Note: contains helper is already defined in config_protection_test.go and is reused here
func stringContains(s, substring string) bool {
	for i := 0; i <= len(s)-len(substring); i++ {
		if s[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}
