package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestDetectRowStatusProcessInstanceSuspended(t *testing.T) {
	cols := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "SUSPENDED", Width: 10},
		{Title: "ENDED", Width: 10},
	}
	row := table.Row{"inst-1", "true", "false"}
	status := detectRowStatus("process-instance", row, cols)
	if status != "suspended" {
		t.Errorf("expected 'suspended', got %q", status)
	}
}

func TestDetectRowStatusProcessInstanceRunning(t *testing.T) {
	cols := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "SUSPENDED", Width: 10},
		{Title: "ENDED", Width: 10},
	}
	row := table.Row{"inst-1", "false", "false"}
	status := detectRowStatus("process-instance", row, cols)
	if status != "running" {
		t.Errorf("expected 'running', got %q", status)
	}
}

func TestDetectRowStatusProcessInstanceEnded(t *testing.T) {
	cols := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "SUSPENDED", Width: 10},
		{Title: "ENDED", Width: 10},
	}
	row := table.Row{"inst-1", "false", "true"}
	status := detectRowStatus("process-instance", row, cols)
	if status != "ended" {
		t.Errorf("expected 'ended', got %q", status)
	}
}

func TestDetectRowStatusJobFailed(t *testing.T) {
	cols := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "RETRIES", Width: 10},
	}
	row := table.Row{"job-1", "0"}
	status := detectRowStatus("job", row, cols)
	if status != "failed" {
		t.Errorf("expected 'failed', got %q", status)
	}
}

func TestDetectRowStatusJobNormal(t *testing.T) {
	cols := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "RETRIES", Width: 10},
	}
	row := table.Row{"job-1", "3"}
	status := detectRowStatus("job", row, cols)
	if status != "normal" {
		t.Errorf("expected 'normal', got %q", status)
	}
}

func TestDetectRowStatusIncident(t *testing.T) {
	cols := []table.Column{{Title: "ID", Width: 20}}
	row := table.Row{"inc-1"}
	status := detectRowStatus("incident", row, cols)
	if status != "failed" {
		t.Errorf("expected 'failed' for incident, got %q", status)
	}
}

func TestDetectRowStatusDefault(t *testing.T) {
	cols := []table.Column{{Title: "ID", Width: 20}}
	row := table.Row{"item-1"}
	status := detectRowStatus("some-unknown-resource", row, cols)
	if status != "normal" {
		t.Errorf("expected 'normal' for unknown resource, got %q", status)
	}
}

func TestColorizeRowsAppliesColor(t *testing.T) {
	cols := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "SUSPENDED", Width: 10},
		{Title: "ENDED", Width: 10},
	}
	rows := []table.Row{
		{"inst-1", "true", "false"},  // suspended → yellow
		{"inst-2", "false", "false"}, // running → green
	}

	colorized := colorizeRows("process-instance", rows, cols)

	if len(colorized) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(colorized))
	}

	// In test mode (no terminal), lipgloss may not apply ANSI escapes,
	// but the function should still process all rows and the text should be preserved.
	if !strings.Contains(colorized[0][0], "inst-1") {
		t.Error("expected first row first cell to still contain 'inst-1' text")
	}
	if !strings.Contains(colorized[1][0], "inst-2") {
		t.Error("expected second row first cell to still contain 'inst-2' text")
	}

	// Verify that the function processed the rows (not nil)
	if colorized[0] == nil || colorized[1] == nil {
		t.Error("expected non-nil rows after colorization")
	}
}

func TestColorizeRowsEmptyInput(t *testing.T) {
	cols := []table.Column{{Title: "ID", Width: 20}}
	var rows []table.Row
	result := colorizeRows("process-instance", rows, cols)
	if result != nil {
		t.Errorf("expected nil for empty rows, got %v", result)
	}
}

func TestColorizeRowsNormalStatus(t *testing.T) {
	cols := []table.Column{{Title: "ID", Width: 20}}
	rows := []table.Row{{"item-1"}}

	colorized := colorizeRows("some-unknown", rows, cols)

	// Normal status should not add styling
	if colorized[0][0] != "item-1" {
		t.Errorf("expected plain 'item-1' for normal status, got %q", colorized[0][0])
	}
}
