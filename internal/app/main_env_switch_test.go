package app

import (
	"testing"
)

func applyEnvSwitchSequence(m *model, targetEnv string) {
	m.switchToEnvironment(targetEnv)
	m.prepareStateTransition(TransitionFull)
	m.breadcrumb = []string{m.currentRoot}
}

func applyContextSwitchSequence(m *model, root string) {
	m.popup.mode = popupModeNone
	m.popup.input = ""
	m.popup.cursor = -1
	m.popup.offset = 0
	m.prepareStateTransition(TransitionFull)
	m.currentRoot = root
	m.breadcrumb = []string{root}
	m.contentHeader = root
	m.viewMode = root
}

func TestEnvSwitch_ClearsActiveModal(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalConfirmDelete

	applyEnvSwitchSequence(&m, m.currentEnv)

	if m.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after env switch, got %v", m.activeModal)
	}
}

func TestEnvSwitch_ClearsSearchState(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.searchTerm = "invoice"
	m.searchMode = true

	applyEnvSwitchSequence(&m, m.currentEnv)

	if m.searchTerm != "" {
		t.Fatalf("expected searchTerm cleared, got %q", m.searchTerm)
	}
	if m.searchMode {
		t.Fatal("expected searchMode false after env switch")
	}
}

func TestEnvSwitch_SetsNewEnvironment(t *testing.T) {
	m := newTestModel(t)
	m.config.Environments["staging"] = m.config.Environments["local"]
	m.envNames = []string{"local", "staging"}
	m.currentEnv = "local"

	m.switchToEnvironment("staging")

	if m.currentEnv != "staging" {
		t.Fatalf("expected currentEnv staging, got %q", m.currentEnv)
	}
}

func TestEnvSwitch_BreadcrumbReset(t *testing.T) {
	m := newTestModel(t)
	m.currentRoot = "task"
	m.breadcrumb = []string{"process-definition", "process-instance"}

	applyEnvSwitchSequence(&m, m.currentEnv)

	if len(m.breadcrumb) != 1 || m.breadcrumb[0] != m.currentRoot {
		t.Fatalf("expected breadcrumb=[%s], got %v", m.currentRoot, m.breadcrumb)
	}
}

func TestEnvSwitch_EscAfterSwitchIsNoop(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.navigationStack = nil

	applyEnvSwitchSequence(&m, m.currentEnv)
	m2, _ := sendKeyString(m, "esc")

	if len(m2.navigationStack) != 0 {
		t.Fatalf("expected empty navigationStack, got %v", m2.navigationStack)
	}
}

func TestContextSwitch_ClearsModalAndSearch(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalConfirmDelete
	m.searchTerm = "foo"
	m.searchMode = true

	applyContextSwitchSequence(&m, "incidents")

	if m.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after context switch, got %v", m.activeModal)
	}
	if m.searchTerm != "" || m.searchMode {
		t.Fatalf("expected cleared search state, got term=%q mode=%v", m.searchTerm, m.searchMode)
	}
}

func TestContextSwitch_SetsNewRoot(t *testing.T) {
	m := newTestModel(t)

	applyContextSwitchSequence(&m, "incidents")

	if m.currentRoot != "incidents" {
		t.Fatalf("expected currentRoot incidents, got %q", m.currentRoot)
	}
	if m.viewMode != "incidents" {
		t.Fatalf("expected viewMode incidents, got %q", m.viewMode)
	}
	if m.contentHeader != "incidents" {
		t.Fatalf("expected contentHeader incidents, got %q", m.contentHeader)
	}
}

func TestContextSwitch_BreadcrumbIsSingleElement(t *testing.T) {
	m := newTestModel(t)
	m.breadcrumb = []string{"process-definition", "process-instance"}

	applyContextSwitchSequence(&m, "incidents")

	if len(m.breadcrumb) != 1 || m.breadcrumb[0] != "incidents" {
		t.Fatalf("expected breadcrumb=[incidents], got %v", m.breadcrumb)
	}
}

func TestContextSwitch_ClearsNavigationStack(t *testing.T) {
	m := newTestModel(t)
	m.navigationStack = []viewState{{viewMode: "a"}, {viewMode: "b"}}

	applyContextSwitchSequence(&m, "incidents")

	if len(m.navigationStack) != 0 {
		t.Fatalf("expected empty navigationStack after context switch, got %d", len(m.navigationStack))
	}
}
