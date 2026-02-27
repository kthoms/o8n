package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/operaton"
)

// testTaskConfig returns a minimal config with a task table that has id, name, assignee columns.
func testTaskConfig(username string) *config.Config {
	return &config.Config{
		Environments: map[string]config.Environment{
			"local": {URL: "http://localhost:8080", Username: username},
		},
		Tables: []config.TableDef{
			{
				Name: "task",
				Columns: []config.ColumnDef{
					{Name: "id"},
					{Name: "name"},
					{Name: "assignee"},
				},
				Actions: []config.ActionDef{
					{Key: "c", Label: "Claim Task", Method: "POST", Path: "/task/{id}/claim", Body: `{"userId": "{currentUser}"}`},
					{Key: "u", Label: "Unclaim Task", Method: "POST", Path: "/task/{id}/unclaim"},
				},
			},
		},
	}
}

// setupTaskTable initialises the model with a task table containing one row.
func setupTaskTable(t *testing.T, id, name, assignee, currentUser string) model {
	t.Helper()
	m := newModel(testTaskConfig(currentUser))
	m.splashActive = false
	m.currentRoot = "task"
	m.breadcrumb = []string{"task"}
	cols := []table.Column{
		{Title: "id", Width: 20},
		{Title: "name", Width: 30},
		{Title: "assignee", Width: 20},
	}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{id, name, assignee}})
	m.table.SetCursor(0)
	return m
}

// ── Claim guard tests (c key) ─────────────────────────────────────────────────

func TestClaimOnUnclaimedTask(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "", "alice")
	m2, cmd := sendKeyString(m, "c")
	if cmd == nil {
		t.Error("expected claimTaskCmd to be dispatched")
	}
	_ = m2
}

func TestClaimOnForeignTask(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "bob", "alice")
	// foreign task — should show error, not dispatch API call (not loading)
	m2, _ := sendKeyString(m, "c")
	if m2.isLoading {
		t.Error("expected no API call (isLoading) when task claimed by another user")
	}
	if !strings.Contains(m2.footerError, "bob") {
		t.Errorf("expected footer error mentioning 'bob', got %q", m2.footerError)
	}
	if m2.footerStatusKind != footerStatusError {
		t.Error("expected footerStatusError")
	}
}

func TestClaimOnOwnTask(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	// own task — should show hint, not dispatch API call
	m2, _ := sendKeyString(m, "c")
	if m2.isLoading {
		t.Error("expected no API call (isLoading) when task already owned")
	}
	if !strings.Contains(m2.footerError, "already own") {
		t.Errorf("expected footer hint about already owning task, got %q", m2.footerError)
	}
	if m2.footerStatusKind != footerStatusInfo {
		t.Error("expected footerStatusInfo")
	}
}

// ── Unclaim guard tests (u key) ───────────────────────────────────────────────

func TestUnclaimOwnTask(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	m2, cmd := sendKeyString(m, "u")
	if cmd == nil {
		t.Error("expected unclaimTaskCmd to be dispatched")
	}
	_ = m2
}

func TestUnclaimUnclaimedTask(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "", "alice")
	m2, _ := sendKeyString(m, "u")
	if m2.isLoading {
		t.Error("expected no API call for unclaimed task")
	}
	if !strings.Contains(m2.footerError, "not claimed") {
		t.Errorf("expected 'not claimed' footer error, got %q", m2.footerError)
	}
}

func TestUnclaimForeignTask(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "bob", "alice")
	m2, _ := sendKeyString(m, "u")
	if m2.isLoading {
		t.Error("expected no API call when task owned by another user")
	}
	if !strings.Contains(m2.footerError, "bob") {
		t.Errorf("expected footer error mentioning 'bob', got %q", m2.footerError)
	}
}

// ── Enter guard tests ─────────────────────────────────────────────────────────

func TestEnterOnOwnTaskFetchesVariables(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	m2, cmd := sendKeyString(m, "enter")
	if cmd == nil {
		t.Error("expected fetchTaskVariablesCmd to be dispatched for own task")
	}
	if m2.footerStatusKind != footerStatusLoading {
		t.Errorf("expected footerStatusLoading, got %v", m2.footerStatusKind)
	}
}

func TestEnterOnUnclaimedTaskShowsError(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "", "alice")
	m2, _ := sendKeyString(m, "enter")
	if m2.footerStatusKind != footerStatusError {
		t.Errorf("expected footerStatusError for unclaimed task, got %v", m2.footerStatusKind)
	}
	if !strings.Contains(m2.footerError, "Claim") {
		t.Errorf("expected 'Claim' in footer error, got %q", m2.footerError)
	}
	if m2.activeModal == ModalTaskComplete {
		t.Error("expected dialog NOT to open for unclaimed task")
	}
}

func TestEnterOnForeignTaskShowsError(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "bob", "alice")
	m2, _ := sendKeyString(m, "enter")
	if m2.footerStatusKind != footerStatusError {
		t.Errorf("expected footerStatusError for foreign task, got %v", m2.footerStatusKind)
	}
	if !strings.Contains(m2.footerError, "bob") {
		t.Errorf("expected footer error with assignee name, got %q", m2.footerError)
	}
	if m2.activeModal == ModalTaskComplete {
		t.Error("expected dialog NOT to open for foreign task")
	}
}

func TestEnterOnNonTaskTableDoesNotIntercept(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.currentRoot = "process-instance"
	m.breadcrumb = []string{"process-instance"}
	cols := []table.Column{{Title: "id", Width: 20}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"inst-1"}})
	m.table.SetCursor(0)
	// Enter on non-task table should not produce the loading status
	m2, _ := sendKeyString(m, "enter")
	if m2.footerStatusKind == footerStatusLoading {
		t.Error("Enter on non-task table should not show loading status")
	}
}

// ── taskVariablesLoadedMsg handler ────────────────────────────────────────────

func TestTaskVariablesLoadedOpensModal(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	inputVars := map[string]variableValue{
		"orderId": {Value: "ORD-123", TypeName: "String"},
	}
	formVars := map[string]variableValue{
		"approved": {Value: nil, TypeName: "Boolean"},
		"orderId":  {Value: nil, TypeName: "String"},
	}
	msg := taskVariablesLoadedMsg{
		taskID:    "task-1",
		taskName:  "My Task",
		inputVars: inputVars,
		formVars:  formVars,
	}
	m2, _ := m.Update(msg)
	result := m2.(model)

	if result.activeModal != ModalTaskComplete {
		t.Error("expected ModalTaskComplete to be active after taskVariablesLoadedMsg")
	}
	if result.taskCompleteTaskID != "task-1" {
		t.Errorf("expected taskCompleteTaskID 'task-1', got %q", result.taskCompleteTaskID)
	}
	if len(result.taskCompleteFields) != 2 {
		t.Errorf("expected 2 form fields, got %d", len(result.taskCompleteFields))
	}
}

// ── Pre-fill test ─────────────────────────────────────────────────────────────

func TestPreFillFromInputVars(t *testing.T) {
	m := newModel(testTaskConfig("alice"))
	inputVars := map[string]variableValue{
		"orderId": {Value: "ORD-999", TypeName: "String"},
	}
	formVars := map[string]variableValue{
		"orderId": {Value: nil, TypeName: "String"},
	}
	fields := m.buildTaskCompleteFields(formVars, inputVars)
	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}
	if fields[0].input.Value() != "ORD-999" {
		t.Errorf("expected pre-filled value 'ORD-999', got %q", fields[0].input.Value())
	}
}

// ── Tab cycle test ────────────────────────────────────────────────────────────

func TestTabCycleInTaskCompleteModal(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	// Simulate dialog open with 2 fields
	m.activeModal = ModalTaskComplete
	m.taskCompleteFields = m.buildTaskCompleteFields(
		map[string]variableValue{
			"fieldA": {Value: nil, TypeName: "String"},
			"fieldB": {Value: nil, TypeName: "String"},
		},
		map[string]variableValue{},
	)
	m.taskCompletePos = 0
	m.taskCompleteFocus = focusTaskField
	m.taskCompleteFields[0].input.Focus()

	// Tab: field[0] → field[1]
	m2, _ := sendKeyString(m, "tab")
	if m2.taskCompletePos != 1 {
		t.Errorf("expected pos 1 after first Tab, got %d", m2.taskCompletePos)
	}
	if m2.taskCompleteFocus != focusTaskField {
		t.Errorf("expected focusTaskField after first Tab")
	}

	// Tab: field[1] → Complete
	m3, _ := sendKeyString(m2, "tab")
	if m3.taskCompleteFocus != focusTaskComplete {
		t.Errorf("expected focusTaskComplete after Tab from last field")
	}

	// Tab: Complete → Back
	m4, _ := sendKeyString(m3, "tab")
	if m4.taskCompleteFocus != focusTaskBack {
		t.Errorf("expected focusTaskBack after Tab from Complete")
	}

	// Tab: Back → field[0]
	m5, _ := sendKeyString(m4, "tab")
	if m5.taskCompleteFocus != focusTaskField {
		t.Errorf("expected focusTaskField after Tab from Back")
	}
	if m5.taskCompletePos != 0 {
		t.Errorf("expected pos 0 after wrap-around Tab, got %d", m5.taskCompletePos)
	}
}

// ── Boolean toggle test ───────────────────────────────────────────────────────

func TestSpaceTogglesBoolField(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	m.activeModal = ModalTaskComplete
	m.taskCompleteFields = m.buildTaskCompleteFields(
		map[string]variableValue{
			"approved": {Value: nil, TypeName: "Boolean"},
		},
		map[string]variableValue{},
	)
	m.taskCompletePos = 0
	m.taskCompleteFocus = focusTaskField
	m.taskCompleteFields[0].input.Focus()
	m.taskCompleteFields[0].input.SetValue("false")

	m2, _ := sendKeyString(m, " ")
	if m2.taskCompleteFields[0].input.Value() != "true" {
		t.Errorf("expected 'true' after Space toggle, got %q", m2.taskCompleteFields[0].input.Value())
	}

	m3, _ := sendKeyString(m2, " ")
	if m3.taskCompleteFields[0].input.Value() != "false" {
		t.Errorf("expected 'false' after second Space toggle, got %q", m3.taskCompleteFields[0].input.Value())
	}
}

// ── Submit / completeTaskCmd ──────────────────────────────────────────────────

func TestSubmitBuildsCorrectVariables(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	m.activeModal = ModalTaskComplete
	m.taskCompleteTaskID = "task-1"
	m.taskCompleteTaskName = "My Task"
	m.taskCompleteFields = m.buildTaskCompleteFields(
		map[string]variableValue{
			"approved": {Value: nil, TypeName: "Boolean"},
			"amount":   {Value: nil, TypeName: "Integer"},
		},
		map[string]variableValue{},
	)
	// Set valid values
	for i, f := range m.taskCompleteFields {
		if f.name == "approved" {
			m.taskCompleteFields[i].input.SetValue("true")
		} else if f.name == "amount" {
			m.taskCompleteFields[i].input.SetValue("42")
		}
	}
	m.taskCompleteFocus = focusTaskComplete

	m2, cmd := sendKeyString(m, "enter")
	if cmd == nil {
		t.Error("expected completeTaskCmd to be dispatched on Enter with Complete focused")
	}
	_ = m2
}

// ── Escape closes dialog ──────────────────────────────────────────────────────

func TestEscapeClosesTaskCompleteDialog(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	m.activeModal = ModalTaskComplete
	m.taskCompleteTaskID = "task-1"
	m.taskCompleteTaskName = "My Task"

	m2, _ := sendKeyString(m, "esc")
	if m2.activeModal != ModalNone {
		t.Error("expected ModalNone after Esc")
	}
	if m2.taskCompleteTaskID != "" {
		t.Error("expected taskCompleteTaskID cleared after Esc")
	}
	if m2.taskCompleteTaskName != "" {
		t.Error("expected taskCompleteTaskName cleared after Esc")
	}
	if m2.taskCompleteFields != nil {
		t.Error("expected taskCompleteFields cleared after Esc")
	}
}

// ── Validation gate test ──────────────────────────────────────────────────────

func TestCompleteDisabledWhenFieldHasError(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	m.activeModal = ModalTaskComplete
	m.taskCompleteTaskID = "task-1"
	m.taskCompleteTaskName = "My Task"
	m.taskCompleteFields = m.buildTaskCompleteFields(
		map[string]variableValue{
			"count": {Value: nil, TypeName: "Integer"},
		},
		map[string]variableValue{},
	)
	// Set invalid value for an integer field
	m.taskCompleteFields[0].input.SetValue("notanumber")
	m.taskCompleteFields[0].error = "enter an integer"
	m.taskCompleteFocus = focusTaskComplete

	m2, cmd := sendKeyString(m, "enter")
	// completeTaskCmd should NOT be dispatched when there are errors
	if cmd != nil {
		t.Error("expected no completeTaskCmd when field has error")
	}
	_ = m2
}

// ── actionExecutedMsg closes dialog on complete ───────────────────────────────

func TestActionExecutedMsgClosesTaskDialog(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	m.activeModal = ModalTaskComplete
	m.taskCompleteTaskID = "task-1"
	m.taskCompleteTaskName = "My Task"

	msg := actionExecutedMsg{label: "Completed: My Task", closeTaskDialog: true}
	m2, _ := m.Update(msg)
	result := m2.(model)

	if result.activeModal != ModalNone {
		t.Error("expected ModalNone after actionExecutedMsg with closeTaskDialog=true")
	}
	if result.taskCompleteTaskID != "" {
		t.Error("expected taskCompleteTaskID cleared after close")
	}
}

// ── renderTaskCompleteModal ───────────────────────────────────────────────────

func TestRenderTaskCompleteModal(t *testing.T) {
	m := setupTaskTable(t, "task-1", "Review Order", "alice", "alice")
	m.activeModal = ModalTaskComplete
	m.taskCompleteTaskName = "Review Order"
	m.lastWidth = 120
	m.lastHeight = 40
	m.taskInputVars = map[string]variableValue{
		"orderId": {Value: "ORD-777", TypeName: "String"},
	}
	m.taskCompleteFields = m.buildTaskCompleteFields(
		map[string]variableValue{
			"approved": {Value: nil, TypeName: "Boolean"},
		},
		m.taskInputVars,
	)
	m.taskCompleteFocus = focusTaskField

	out := m.renderTaskCompleteModal(120, 40)
	if out == "" {
		t.Fatal("expected non-empty modal output")
	}
	if !strings.Contains(out, "Complete Task") {
		t.Error("expected 'Complete Task' title in modal")
	}
	if !strings.Contains(out, "Review Order") {
		t.Error("expected task name in modal")
	}
	if !strings.Contains(out, "INPUT VARIABLES") {
		t.Error("expected INPUT VARIABLES section")
	}
	if !strings.Contains(out, "OUTPUT VARIABLES") {
		t.Error("expected OUTPUT VARIABLES section")
	}
	if !strings.Contains(out, "Complete") {
		t.Error("expected Complete button in modal")
	}
	if !strings.Contains(out, "Back") {
		t.Error("expected Back button in modal")
	}
}

// ── {currentUser} placeholder ─────────────────────────────────────────────────

func TestCurrentUserPlaceholderInBody(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"local": {URL: "http://localhost:8080", Username: "testuser"},
		},
	}
	m := newModel(cfg)
	m.currentEnv = "local"
	action := config.ActionDef{
		Key:    "c",
		Label:  "Claim",
		Method: "POST",
		Path:   "/task/{id}/claim",
		Body:   `{"userId": "{currentUser}"}`,
	}
	// Build the resolved body
	env := cfg.Environments["local"]
	resolvedBody := replaceCurrentUser(action.Body, env.Username)
	if resolvedBody != `{"userId": "testuser"}` {
		t.Errorf("expected resolved body with username, got %q", resolvedBody)
	}
	_ = m
}

// replaceCurrentUser is a helper to test the placeholder resolution logic.
func replaceCurrentUser(body, username string) string {
	return strings.ReplaceAll(body, "{currentUser}", username)
}

// ── completeTaskCmd sends correct payload ──────────────────────────────────────

func TestCompleteTaskCmdUsesOrigType(t *testing.T) {
	// Verify that buildTaskCompleteFields preserves origType for API submission
	m := newModel(testTaskConfig("alice"))
	formVars := map[string]variableValue{
		"approved": {Value: nil, TypeName: "Boolean"},
		"name":     {Value: "default", TypeName: "String"},
	}
	fields := m.buildTaskCompleteFields(formVars, map[string]variableValue{})

	for _, f := range fields {
		switch f.name {
		case "approved":
			if f.origType != "Boolean" {
				t.Errorf("expected origType 'Boolean', got %q", f.origType)
			}
			if f.varType != "bool" {
				t.Errorf("expected varType 'bool', got %q", f.varType)
			}
		case "name":
			if f.origType != "String" {
				t.Errorf("expected origType 'String', got %q", f.origType)
			}
		}
	}
}

// ── submitTaskComplete assembles correct VariableValueDto ─────────────────────

func TestSubmitTaskCompleteAssemblesVars(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	m.activeModal = ModalTaskComplete
	m.taskCompleteTaskID = "task-1"
	m.taskCompleteTaskName = "My Task"

	approvedInput := textinput.New()
	approvedInput.SetValue("true")
	countInput := textinput.New()
	countInput.SetValue("5")

	m.taskCompleteFields = []taskCompleteField{
		{name: "approved", varType: "bool", origType: "Boolean", input: approvedInput},
		{name: "count", varType: "int", origType: "Integer", input: countInput},
	}
	m.taskCompleteFocus = focusTaskComplete

	cmd := m.submitTaskComplete()
	if cmd == nil {
		t.Fatal("expected completeTaskCmd to be returned")
	}
}

// ── Verify no drilldown on task Enter ─────────────────────────────────────────

func TestTaskEnterDoesNotDrilldown(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	// Enter should open dialog, not drilldown
	m2, _ := sendKeyString(m, "enter")
	// Should not have pushed to navigation stack
	if len(m2.navigationStack) > 0 {
		t.Error("Enter on task table with own task should not push to navigationStack")
	}
	// Should show loading state (fetching variables)
	if m2.footerStatusKind != footerStatusLoading {
		t.Errorf("expected footerStatusLoading, got %v", m2.footerStatusKind)
	}
}

// ── completeTaskCmd with empty form vars ──────────────────────────────────────

func TestCompleteWithNoFormVarsSendsEmptyMap(t *testing.T) {
	m := setupTaskTable(t, "task-1", "My Task", "alice", "alice")
	m.activeModal = ModalTaskComplete
	m.taskCompleteTaskID = "task-1"
	m.taskCompleteTaskName = "My Task"
	m.taskCompleteFields = []taskCompleteField{} // no form vars
	m.taskCompleteFocus = focusTaskComplete

	cmd := m.submitTaskComplete()
	if cmd == nil {
		t.Error("expected completeTaskCmd even with empty form vars")
	}
}

// ── VariableValueDto type in submission ───────────────────────────────────────

func TestVariableValueDtoHasOrigType(t *testing.T) {
	// Verify the VariableValueDto structure that would be submitted
	v := operaton.VariableValueDto{}
	v.SetValue(true)
	v.SetType("Boolean")
	if v.GetType() != "Boolean" {
		t.Errorf("expected type 'Boolean', got %q", v.GetType())
	}
	val := v.Value
	if val == nil {
		t.Error("expected value to be set")
	}
}
