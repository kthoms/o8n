package app

import (
	"testing"

	"github.com/kthoms/o6n/internal/config"
	"github.com/kthoms/o6n/internal/dao"
)

func applyEnvSwitchSequence(m *model, targetEnv string) {
	m.switchToEnvironment(targetEnv)
	m.prepareStateTransition(TransitionFull)
	m.currentRoot = dao.ResourceProcessDefinitions
	m.contentHeader = dao.ResourceProcessDefinitions
	m.breadcrumb = []string{m.currentRoot}
}

func applyContextSwitchSequence(m *model, root string) {
	m.activeModal = ModalNone
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
	// Start on a non-default root to exercise the stale-root fix.
	m.currentRoot = "task"
	m.breadcrumb = []string{"process-definition", "process-instance"}

	applyEnvSwitchSequence(&m, m.currentEnv)

	// After env switch, currentRoot and breadcrumb must be reset to the process-definitions home root.
	if m.currentRoot != dao.ResourceProcessDefinitions {
		t.Fatalf("expected currentRoot=%s after env switch, got %q", dao.ResourceProcessDefinitions, m.currentRoot)
	}
	if len(m.breadcrumb) != 1 || m.breadcrumb[0] != dao.ResourceProcessDefinitions {
		t.Fatalf("expected breadcrumb=[%s], got %v", dao.ResourceProcessDefinitions, m.breadcrumb)
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

func TestEnvSwitch_KeyHandler_UsesNewEnvironmentURLAndClearsState(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"dev":   {URL: "http://dev"},
			"local": {URL: "http://local"},
		},
		Active: "local",
	}
	m := newModel(cfg)
	m.splashActive = false
	m.activeModal = ModalEnvironment
	m.searchTerm = "invoice"
	m.searchMode = false
	m.navigationStack = []viewState{{viewMode: "process-instance"}}
	m.currentRoot = "task"

	// Select "dev" in popup and trigger real key-handler path.
	devIdx := -1
	for i, name := range m.envNames {
		if name == "dev" {
			devIdx = i
			break
		}
	}
	if devIdx < 0 {
		t.Fatalf("expected env list to contain dev, got %v", m.envNames)
	}
	m.envPopupCursor = devIdx
	m2, _ := sendKeyString(m, "enter")

	if m2.currentEnv != "dev" {
		t.Fatalf("expected currentEnv=dev, got %q", m2.currentEnv)
	}
	if env, ok := m2.config.Environments[m2.currentEnv]; !ok || env.URL != "http://dev" {
		t.Fatalf("expected active env URL http://dev, got %#v", env)
	}
	if m2.searchTerm != "" || m2.searchMode {
		t.Fatalf("expected cleared search state after env switch, got term=%q mode=%v", m2.searchTerm, m2.searchMode)
	}
	if len(m2.navigationStack) != 0 {
		t.Fatalf("expected cleared navigationStack after env switch, got %d", len(m2.navigationStack))
	}
}

func TestContextSwitch_KeyHandler_ClearsStateAndSetsRoot(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.rootContexts = append(m.rootContexts, "incidents")
	m.activeModal = ModalContextSwitcher
	m.popup.input = "incidents"
	m.popup.cursor = -1
	m.popup.offset = 3
	m.searchTerm = "foo"
	m.searchMode = false
	m.navigationStack = []viewState{{viewMode: "process-instance"}}

	m2, _ := sendKeyString(m, "enter")

	if m2.currentRoot != "incidents" || m2.viewMode != "incidents" || m2.contentHeader != "incidents" {
		t.Fatalf("expected incidents context, got root=%q view=%q header=%q", m2.currentRoot, m2.viewMode, m2.contentHeader)
	}
	if m2.activeModal != ModalNone {
		t.Fatalf("expected modal cleared, got %v", m2.activeModal)
	}
	if m2.searchTerm != "" || m2.searchMode {
		t.Fatalf("expected search cleared, got term=%q mode=%v", m2.searchTerm, m2.searchMode)
	}
	if len(m2.navigationStack) != 0 {
		t.Fatalf("expected navigationStack cleared, got %d", len(m2.navigationStack))
	}
	if m2.activeModal != ModalNone || m2.popup.offset != 0 {
		t.Fatalf("expected modal closed/reset, got modal=%v offset=%d", m2.activeModal, m2.popup.offset)
	}
}
