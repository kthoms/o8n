package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func TestActionsMenuOpensOnSpace(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.currentRoot = "process-instance"
	cols := []table.Column{{Title: "ID", Width: 20}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"inst-1"}})
	m.table.SetCursor(0)

	m2, _ := sendKeyString(m, " ")

	if !m2.showActionsMenu {
		t.Error("expected showActionsMenu to be true after Space")
	}
	if len(m2.actionsMenuItems) == 0 {
		t.Error("expected actions menu items to be populated")
	}
}

func TestActionsMenuItemsForProcessInstance(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.currentRoot = "process-instance"
	m.breadcrumb = []string{"process-instance"}

	items := m.buildActionsForRoot()

	// Should have: View Variables, Suspend, Resume, View as JSON
	if len(items) < 4 {
		t.Errorf("expected at least 4 items for process-instance, got %d", len(items))
	}

	// Check keys
	keys := make(map[string]bool)
	for _, item := range items {
		keys[item.key] = true
	}
	for _, expected := range []string{"v", "s", "r", "y"} {
		if !keys[expected] {
			t.Errorf("expected key %q in actions menu", expected)
		}
	}
}

func TestActionsMenuItemsForJob(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.currentRoot = "job"
	m.breadcrumb = []string{"job"}

	items := m.buildActionsForRoot()

	if len(items) < 2 {
		t.Errorf("expected at least 2 items for job, got %d", len(items))
	}

	keys := make(map[string]bool)
	for _, item := range items {
		keys[item.key] = true
	}
	if !keys["r"] {
		t.Error("expected 'r' (retry) in job actions")
	}
	if !keys["y"] {
		t.Error("expected 'y' (view) in job actions")
	}
}

func TestActionsMenuEscCloses(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.showActionsMenu = true
	m.actionsMenuItems = []actionItem{{key: "y", label: "Test"}}

	m2, _ := sendKeyString(m, "esc")

	if m2.showActionsMenu {
		t.Error("expected showActionsMenu to be false after Esc")
	}
}

func TestActionsMenuKeyShortcut(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.showActionsMenu = true
	executed := false
	m.actionsMenuItems = []actionItem{
		{key: "t", label: "Test Action", cmd: func(m *model) tea.Cmd {
			executed = true
			return nil
		}},
	}

	m2, _ := sendKeyString(m, "t")

	if m2.showActionsMenu {
		t.Error("expected showActionsMenu to be false after shortcut key")
	}
	if !executed {
		t.Error("expected action command to be executed")
	}
}

func TestActionsMenuNoRowNoOpen(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	// No rows in table
	m.table.SetRows([]table.Row{})

	m2, _ := sendKeyString(m, " ")

	if m2.showActionsMenu {
		t.Error("expected showActionsMenu to remain false with no rows")
	}
}

func TestBuildDetailContent(t *testing.T) {
	m := newTestModel(t)
	cols := []table.Column{{Title: "ID", Width: 10}, {Title: "NAME", Width: 20}}
	m.table.SetColumns(cols)

	row := table.Row{"inst-1", "MyProcess"}
	content := m.buildDetailContent(row)

	if content == "" {
		t.Error("expected non-empty detail content")
	}
	if !strings.Contains(content, "inst-1") {
		t.Error("expected content to contain 'inst-1'")
	}
	if !strings.Contains(content, "MyProcess") {
		t.Error("expected content to contain 'MyProcess'")
	}
}
