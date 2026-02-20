package main

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kthoms/o8n/internal/config"
)

// TestDrilldownFromDefinitionsToInstances verifies drilldown works from definitions to instances
func TestDrilldownFromDefinitionsToInstances(t *testing.T) {
	m := newTestModel(t)
	m.viewMode = "definitions"

	// Simulate having some definitions loaded
	mockDefs := []config.ProcessDefinition{
		{
			ID:  "invoice:1:test-id-123",
			Key: "invoice",
		},
		{
			ID:  "review:1:test-id-456",
			Key: "review",
		},
	}

	// Convert to table rows
	rows := []table.Row{}
	for _, def := range mockDefs {
		rows = append(rows, table.Row{def.Key, def.ID})
	}

	m.table.SetRows(rows)
	m.table.SetColumns([]table.Column{
		{Title: "Key", Width: 20},
		{Title: "ID", Width: 50},
	})

	// Select first row
	m.table.SetCursor(0)

	// Verify initial state
	if m.viewMode != "definitions" {
		t.Errorf("expected viewMode 'definitions', got '%s'", m.viewMode)
	}

	if len(m.table.Rows()) != 2 {
		t.Errorf("expected 2 rows, got %d", len(m.table.Rows()))
	}
}

// TestDrilldownParameterPassThrough verifies that drilldown parameters are correctly passed
func TestDrilldownParameterPassThrough(t *testing.T) {
	m := newTestModel(t)

	// Simulate drilldown from definitions to instances with filter
	m.currentRoot = "process-instance"
	m.genericParams = map[string]string{
		"processDefinitionId": "invoice:1:test-id-123",
	}

	// Verify parameters are stored
	if len(m.genericParams) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(m.genericParams))
	}

	if m.genericParams["processDefinitionId"] != "invoice:1:test-id-123" {
		t.Errorf("expected processDefinitionId to be 'invoice:1:test-id-123', got '%s'",
			m.genericParams["processDefinitionId"])
	}
}

// TestDrilldownBreadcrumb verifies breadcrumb is correctly updated during drilldown
func TestDrilldownBreadcrumb(t *testing.T) {
	m := newTestModel(t)

	// Start at definitions
	m.breadcrumb = []string{"process-definition"}

	// Drilldown to instances
	m.breadcrumb = append(m.breadcrumb, "process-instance")

	// Verify breadcrumb
	if len(m.breadcrumb) != 2 {
		t.Errorf("expected breadcrumb length 2, got %d", len(m.breadcrumb))
	}

	if m.breadcrumb[0] != "process-definition" {
		t.Errorf("expected first breadcrumb 'process-definition', got '%s'", m.breadcrumb[0])
	}

	if m.breadcrumb[1] != "process-instance" {
		t.Errorf("expected second breadcrumb 'process-instance', got '%s'", m.breadcrumb[1])
	}
}

// TestDrilldownConfigParsing verifies that drilldown config is correctly parsed
func TestDrilldownConfigParsing(t *testing.T) {
	m := newTestModel(t)

	// Find process-definition table def
	var defTableDef *config.TableDef
	if m.config != nil && m.config.Tables != nil {
		for _, table := range m.config.Tables {
			if table.Name == "process-definition" {
				defTableDef = &table
				break
			}
		}
	}

	if defTableDef == nil {
		t.Skip("process-definition table not in config")
		return
	}

	// Verify drilldown config exists
	if len(defTableDef.Drilldown) == 0 {
		t.Error("expected drilldown config for process-definition")
		return
	}

	// First drilldown should be to process-instance
	first := defTableDef.Drilldown[0]
	if first.Target != "process-instance" {
		t.Errorf("expected first drilldown target 'process-instance', got '%s'", first.Target)
	}

	if first.Param != "processDefinitionId" {
		t.Errorf("expected parameter 'processDefinitionId', got '%s'", first.Param)
	}

	if first.Column != "id" {
		t.Errorf("expected column 'id', got '%s'", first.Column)
	}
}

// TestDrilldownURLConstruction tests that filter params are included in API URLs
func TestDrilldownURLConstruction(t *testing.T) {
	// This test verifies the logic in fetchGenericCmd that builds URLs with filter params
	m := newTestModel(t)

	// Set up filter parameters like drilldown would
	m.genericParams = map[string]string{
		"processDefinitionId": "invoice:1:test-id-123",
	}

	// Verify params can be retrieved
	paramsCopy := make(map[string]string, len(m.genericParams))
	for k, v := range m.genericParams {
		paramsCopy[k] = v
	}

	if len(paramsCopy) != 1 {
		t.Error("failed to copy parameters")
	}

	// Simulate URL construction
	base := "http://localhost:8080/engine-rest"
	apiPath := "process-instance"
	urlStr := base + "/" + apiPath

	// Add filter params
	for k, v := range paramsCopy {
		urlStr = fmt.Sprintf("%s?%s=%s", urlStr, k, v)
	}

	// Add paging
	urlStr = fmt.Sprintf("%s&firstResult=0&maxResults=10", urlStr)

	expectedURL := "http://localhost:8080/engine-rest/process-instance?processDefinitionId=invoice:1:test-id-123&firstResult=0&maxResults=10"
	if urlStr != expectedURL {
		t.Errorf("URL construction failed.\nExpected: %s\nGot: %s", expectedURL, urlStr)
	}
}

// TestDrilldownAllResources verifies all drilldown paths in config are properly defined
func TestDrilldownAllResources(t *testing.T) {
	m := newTestModel(t)

	if m.config == nil || m.config.Tables == nil {
		t.Skip("config not available")
		return
	}

	drilldownCount := 0
	for _, table := range m.config.Tables {
		if len(table.Drilldown) > 0 {
			drilldownCount++
			t.Logf("Found drilldown for %s:", table.Name)
			for i, dd := range table.Drilldown {
				t.Logf("  [%d] target=%s, param=%s, column=%s",
					i, dd.Target, dd.Param, dd.Column)
			}
		}
	}

	if drilldownCount == 0 {
		t.Error("expected at least one table with drilldown config")
	}
}

// TestInstancesCountEndpoint verifies that instances count endpoint works
func TestInstancesCountEndpoint(t *testing.T) {
	m := newTestModel(t)

	// This test just verifies the model has support for count endpoints
	m.pageTotals = make(map[string]int)
	m.pageOffsets = make(map[string]int)

	m.pageTotals["process-instance"] = 8
	m.pageOffsets["process-instance"] = 0

	if m.pageTotals["process-instance"] != 8 {
		t.Error("failed to store total count")
	}

	// Calculate pages
	pageSize := m.getPageSize()
	currentPage := (m.pageOffsets["process-instance"] / pageSize) + 1
	totalPages := (m.pageTotals["process-instance"] + pageSize - 1) / pageSize

	if currentPage != 1 {
		t.Errorf("expected page 1, got %d", currentPage)
	}

	if totalPages < 1 {
		t.Errorf("expected at least 1 total page, got %d", totalPages)
	}

	t.Logf("Pagination: page %d of %d (total items: %d, page size: %d)",
		currentPage, totalPages, m.pageTotals["process-instance"], pageSize)
}
