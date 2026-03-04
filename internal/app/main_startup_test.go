package app

import (
	"testing"

	"github.com/kthoms/o8n/internal/config"
)

// testModelWithContexts creates a newTestModel and populates rootContexts.
func testModelWithContexts(t *testing.T) model {
	t.Helper()
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition", "process-instance", "process-variables"}
	return m
}

// ── AC 1: State Restoration ───────────────────────────────────────────────────

// TestStateRestoration_RootRestoredFromNavState verifies that restoreNavState
// correctly applies a saved NavState to the model (root, breadcrumb, drilldown params).
func TestStateRestoration_RootRestoredFromNavState(t *testing.T) {
	m := newTestModel(t)
	nav := config.NavState{
		Root:       "process-instance",
		Breadcrumb: []string{"process-definition", "process-instance"},
		GenericParams: map[string]string{
			"processDefinitionId": "def-123",
		},
	}

	m.restoreNavState(nav)

	if m.currentRoot != "process-instance" {
		t.Errorf("expected currentRoot 'process-instance', got %q", m.currentRoot)
	}
	if len(m.breadcrumb) != 2 {
		t.Fatalf("expected breadcrumb len 2, got %d", len(m.breadcrumb))
	}
	if m.breadcrumb[0] != "process-definition" || m.breadcrumb[1] != "process-instance" {
		t.Errorf("unexpected breadcrumb: %v", m.breadcrumb)
	}
	if m.genericParams["processDefinitionId"] != "def-123" {
		t.Errorf("expected genericParams[processDefinitionId]='def-123', got %q", m.genericParams["processDefinitionId"])
	}
}

// TestStateRestoration_EmptyRootSkipsRestore verifies that restoreNavState is
// a no-op on fresh start (nav.Root == ""), preserving the default state.
func TestStateRestoration_EmptyRootSkipsRestore(t *testing.T) {
	m := newTestModel(t)
	// Record default breadcrumb set by newModel()
	defaultBreadcrumb := append([]string{}, m.breadcrumb...)

	m.restoreNavState(config.NavState{}) // empty nav = first run

	if m.breadcrumb[0] != defaultBreadcrumb[0] {
		t.Errorf("expected breadcrumb unchanged after empty restoreNavState, got %v", m.breadcrumb)
	}
}

// TestStateRestoration_EnvRestoredFromAppState verifies that environment
// selection is correctly applied when the saved env exists in config.
func TestStateRestoration_EnvRestoredFromAppState(t *testing.T) {
	m := newTestModel(t)
	// Test config (from newTestModel) includes "local" environment
	savedEnv := "local"

	// Simulate what run.go:92-98 does after LoadAppState
	if _, ok := m.config.Environments[savedEnv]; ok {
		m.currentEnv = savedEnv
	}

	if m.currentEnv != savedEnv {
		t.Errorf("expected currentEnv %q after restoration, got %q", savedEnv, m.currentEnv)
	}
}

// TestStateRestoration_UnknownEnvNotRestored verifies that a saved env that no
// longer exists in config is silently ignored (safe fallback).
func TestStateRestoration_UnknownEnvNotRestored(t *testing.T) {
	m := newTestModel(t)
	original := m.currentEnv

	// Simulate what run.go:92-98 does — env not in config → no change
	savedEnv := "deleted-env"
	if _, ok := m.config.Environments[savedEnv]; ok {
		m.currentEnv = savedEnv
	}

	if m.currentEnv != original {
		t.Errorf("expected currentEnv unchanged for unknown env, got %q (was %q)", m.currentEnv, original)
	}
}

// TestStateRestoration_CurrentNavStateRoundTrip verifies currentNavState()
// captures the model's nav fields and restoreNavState() restores them exactly.
func TestStateRestoration_CurrentNavStateRoundTrip(t *testing.T) {
	m := newTestModel(t)
	m.currentRoot = "task"
	m.breadcrumb = []string{"task"}
	m.selectedDefinitionKey = "def-abc"
	m.selectedInstanceID = "inst-xyz"
	m.genericParams = map[string]string{"taskAssignee": "user1"}

	// Capture state
	nav := m.currentNavState()

	// Restore into a fresh model
	m2 := newTestModel(t)
	m2.restoreNavState(nav)

	if m2.currentRoot != "task" {
		t.Errorf("currentRoot: got %q, want 'task'", m2.currentRoot)
	}
	if len(m2.breadcrumb) != 1 || m2.breadcrumb[0] != "task" {
		t.Errorf("breadcrumb: got %v", m2.breadcrumb)
	}
	if m2.selectedDefinitionKey != "def-abc" {
		t.Errorf("selectedDefinitionKey: got %q", m2.selectedDefinitionKey)
	}
	if m2.selectedInstanceID != "inst-xyz" {
		t.Errorf("selectedInstanceID: got %q", m2.selectedInstanceID)
	}
	if m2.genericParams["taskAssignee"] != "user1" {
		t.Errorf("genericParams[taskAssignee]: got %q", m2.genericParams["taskAssignee"])
	}
}

// ── AC 2: FirstRunModal Opens on Fresh State ──────────────────────────────────

// TestFirstRunModal_OpensOnFreshState verifies that when firstRunNeeded is true
// and the app dispatches openFirstRunMsg, ModalFirstRun becomes the active modal.
func TestFirstRunModal_OpensOnFreshState(t *testing.T) {
	m := testModelWithContexts(t)
	m.firstRunNeeded = true
	m.activeModal = ModalNone

	m2, _ := m.Update(openFirstRunMsg{})
	res := m2.(model)

	if res.activeModal != ModalFirstRun {
		t.Errorf("expected ModalFirstRun active, got %v", res.activeModal)
	}
}

// TestFirstRunModal_EscSwallowed verifies that Esc is ignored when ModalFirstRun
// is active — selection is mandatory (Story 1.5 AC2, documented exception to Story 1.4).
func TestFirstRunModal_EscSwallowed(t *testing.T) {
	m := testModelWithContexts(t)
	m.activeModal = ModalFirstRun

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalFirstRun {
		t.Errorf("expected ModalFirstRun still active after Esc, got %v", m2.activeModal)
	}
}

// TestFirstRunModal_EnterConfirms verifies that Enter on a valid cursor selection
// sets currentRoot and dismisses ModalFirstRun.
func TestFirstRunModal_EnterConfirms(t *testing.T) {
	m := testModelWithContexts(t)
	m.activeModal = ModalFirstRun
	m.firstRunCursor = 0 // select first context

	m2, _ := sendKeyString(m, "enter")

	if m2.activeModal != ModalNone {
		t.Errorf("expected modal dismissed after Enter, got %v", m2.activeModal)
	}
	if m2.currentRoot != "process-definition" {
		t.Errorf("expected currentRoot 'process-definition', got %q", m2.currentRoot)
	}
	if m2.firstRunNeeded {
		t.Error("expected firstRunNeeded=false after selection")
	}
}

// TestFirstRunModal_TypeFiltersList verifies that typing updates firstRunInput
// and cursor clamps when filtered list shrinks.
func TestFirstRunModal_TypeFiltersList(t *testing.T) {
	m := testModelWithContexts(t)
	m.activeModal = ModalFirstRun
	m.firstRunCursor = 2 // last item

	m2, _ := sendKeyString(m, "z") // 'z' matches nothing

	if m2.firstRunInput != "z" {
		t.Errorf("expected firstRunInput 'z', got %q", m2.firstRunInput)
	}
	// No matches → cursor clamped to 0
	if m2.firstRunCursor != 0 {
		t.Errorf("expected cursor clamped to 0 on empty filter, got %d", m2.firstRunCursor)
	}
}

// TestFirstRunModal_BackspaceRemovesChar verifies that backspace removes the
// last character from firstRunInput.
func TestFirstRunModal_BackspaceRemovesChar(t *testing.T) {
	m := testModelWithContexts(t)
	m.activeModal = ModalFirstRun
	m.firstRunInput = "proce"

	m2, _ := sendKeyString(m, "backspace")

	if m2.firstRunInput != "proc" {
		t.Errorf("expected firstRunInput 'proc' after backspace, got %q", m2.firstRunInput)
	}
}

// TestFirstRunModal_CursorClampsOnFilter verifies that typing a filter that
// reduces the list clamps firstRunCursor to len(filtered)-1.
func TestFirstRunModal_CursorClampsOnFilter(t *testing.T) {
	m := testModelWithContexts(t)
	m.activeModal = ModalFirstRun
	m.firstRunCursor = 2

	// Typing "process-d" should match only "process-definition" (1 result)
	m2, _ := sendKeyString(m, "process-d")

	if m2.firstRunCursor != 0 {
		t.Errorf("expected cursor clamped to 0 for single-result filter, got %d", m2.firstRunCursor)
	}
}

// TestCtrlHOpensFirstRunModal verifies that Ctrl+H opens ModalFirstRun from normal state.
func TestCtrlHOpensFirstRunModal(t *testing.T) {
	m := testModelWithContexts(t)
	m.activeModal = ModalNone

	m2, _ := sendKeyString(m, "ctrl+h")

	if m2.activeModal != ModalFirstRun {
		t.Errorf("expected ModalFirstRun after Ctrl+H, got %v", m2.activeModal)
	}
	if m2.firstRunInput != "" {
		t.Errorf("expected cleared firstRunInput, got %q", m2.firstRunInput)
	}
}

// ── AC 4: API Resilience ──────────────────────────────────────────────────────

// TestAPIResilience_EmptyItems verifies that an empty items slice produces a
// "No results" placeholder row rather than panicking (via genericLoadedMsg).
func TestAPIResilience_EmptyItems(t *testing.T) {
	m := newTestModel(t)
	m.currentRoot = "process-definition"

	msg := genericLoadedMsg{root: "process-definition", items: nil}
	m2, _ := m.Update(msg)
	res := m2.(model)

	rows := res.table.Rows()
	if len(rows) != 1 {
		t.Errorf("expected 1 placeholder row for nil items, got %d", len(rows))
	} else if len(rows[0]) == 0 || rows[0][0] != "No results" {
		t.Errorf("expected 'No results' placeholder, got %q", rows[0])
	}
}

// TestAPIResilience_UnknownRoot verifies that a genericLoadedMsg for an unknown
// root key does not crash the model.
func TestAPIResilience_UnknownRoot(t *testing.T) {
	m := newTestModel(t)
	// unknown-root has no TableDef; the model must ignore or handle gracefully
	msg := genericLoadedMsg{root: "unknown-root", items: []map[string]interface{}{{"key": "val"}}}
	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic for unknown root: %v", r)
		}
	}()
	m.Update(msg)
}
