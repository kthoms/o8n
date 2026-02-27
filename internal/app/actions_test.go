package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kthoms/o8n/internal/config"
)

// testConfigWithActions returns a config with table definitions and actions
// for testing the config-driven action system.
func testConfigWithActions() *config.Config {
	return &config.Config{
		Environments: map[string]config.Environment{
			"local": {URL: "http://localhost:8080"},
		},
		Tables: []config.TableDef{
			{
				Name: "process-instance",
				Columns: []config.ColumnDef{
					{Name: "id", Width: 0},
					{Name: "definitionId", Width: 0},
				},
				Actions: []config.ActionDef{
					{Key: "s", Label: "Suspend Instance", Method: "PUT", Path: "/process-instance/{id}/suspended", Body: `{"suspended":true}`},
					{Key: "r", Label: "Resume Instance", Method: "PUT", Path: "/process-instance/{id}/suspended", Body: `{"suspended":false}`},
					{Key: "ctrl+d", Label: "Delete Instance", Method: "DELETE", Path: "/process-instance/{id}", Confirm: true},
				},
			},
			{
				Name: "job",
				Columns: []config.ColumnDef{
					{Name: "id", Width: 0},
				},
				Actions: []config.ActionDef{
					{Key: "r", Label: "Retry", Method: "PUT", Path: "/job/{id}/retries", Body: `{"retries":1}`},
					{Key: "x", Label: "Execute", Method: "POST", Path: "/job/{id}/execute"},
				},
			},
			{
				Name: "task",
				Columns: []config.ColumnDef{
					{Name: "id", Width: 0},
					{Name: "assignee", Width: 0},
				},
				Actions: []config.ActionDef{
					{Key: "c", Label: "Claim Task", Method: "POST", Path: "/task/{id}/claim", Body: `{"userId": "{currentUser}"}`},
					{Key: "u", Label: "Unclaim Task", Method: "POST", Path: "/task/{id}/unclaim"},
				},
			},
		},
	}
}

func TestActionsMenuOpensOnSpace(t *testing.T) {
	m := newTestModel(t)
	m.config = testConfigWithActions()
	m.splashActive = false
	m.currentRoot = "process-instance"
	cols := []table.Column{{Title: "id", Width: 20}}
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
	m.config = testConfigWithActions()
	m.splashActive = false
	m.currentRoot = "process-instance"
	m.breadcrumb = []string{"process-instance"}

	items := m.buildActionsForRoot()

	// Should have: Suspend, Resume, Delete + View as JSON (always appended)
	if len(items) < 4 {
		t.Errorf("expected at least 4 items for process-instance, got %d", len(items))
	}

	keys := make(map[string]bool)
	for _, item := range items {
		keys[item.key] = true
	}
	for _, expected := range []string{"s", "r", "ctrl+d", "y"} {
		if !keys[expected] {
			t.Errorf("expected key %q in actions menu", expected)
		}
	}
}

func TestActionsMenuItemsForJob(t *testing.T) {
	m := newTestModel(t)
	m.config = testConfigWithActions()
	m.splashActive = false
	m.currentRoot = "job"
	m.breadcrumb = []string{"job"}

	items := m.buildActionsForRoot()

	// Should have: Retry, Execute + View as JSON
	if len(items) < 3 {
		t.Errorf("expected at least 3 items for job, got %d", len(items))
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

func TestActionsMenuItemsForTask(t *testing.T) {
	m := newTestModel(t)
	m.config = testConfigWithActions()
	m.splashActive = false
	m.currentRoot = "task"
	m.breadcrumb = []string{"task"}

	items := m.buildActionsForRoot()

	// Should have: Claim, Unclaim + View as JSON (Complete is now a dialog-driven flow)
	if len(items) < 3 {
		t.Errorf("expected at least 3 items for task, got %d", len(items))
	}

	keys := make(map[string]bool)
	for _, item := range items {
		keys[item.key] = true
	}
	for _, expected := range []string{"c", "u", "y"} {
		if !keys[expected] {
			t.Errorf("expected key %q in task actions menu", expected)
		}
	}
	if keys["k"] {
		t.Error("key 'k' should no longer be in task actions (renamed to 'c' for Claim)")
	}
}

func TestActionsConfirmFlow(t *testing.T) {
	m := newTestModel(t)
	m.config = testConfigWithActions()
	m.splashActive = false
	m.currentRoot = "process-instance"
	m.breadcrumb = []string{"process-instance"}
	cols := []table.Column{{Title: "id", Width: 20}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"inst-42"}})
	m.table.SetCursor(0)

	// Build actions and find the confirm action (ctrl+d = Delete)
	items := m.buildActionsForRoot()
	t.Logf("Found %d actions for process-instance:", len(items))
	for _, it := range items {
		t.Logf("  key=%q label=%q", it.key, it.label)
	}
	var deleteItem *actionItem
	for i := range items {
		if items[i].key == "ctrl+d" {
			deleteItem = &items[i]
			break
		}
	}
	if deleteItem == nil {
		t.Fatal("expected ctrl+d action in process-instance actions")
	}

	// Execute the action - should set pending state and modal
	deleteItem.cmd(&m)

	if m.activeModal != ModalConfirmDelete {
		t.Error("expected ModalConfirmDelete after confirm action")
	}
	if m.pendingAction == nil {
		t.Error("expected pendingAction to be set")
	}
	if m.pendingActionPath != "/process-instance/inst-42" {
		t.Errorf("expected resolved path /process-instance/inst-42, got %s", m.pendingActionPath)
	}
}

func TestActionsDefaultViewAsJSON(t *testing.T) {
	m := newTestModel(t)
	// No config actions for this resource
	m.splashActive = false
	m.currentRoot = "some-unknown-resource"
	m.breadcrumb = []string{"some-unknown-resource"}

	items := m.buildActionsForRoot()

	// Should have at least View as JSON
	if len(items) < 1 {
		t.Error("expected at least 1 action (View as JSON)")
	}
	if items[len(items)-1].key != "y" {
		t.Errorf("expected last action key to be 'y', got %q", items[len(items)-1].key)
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
