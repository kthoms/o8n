package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kthoms/o8n/internal/config"
)

func TestDelegateActionBodyContainsCurrentUser(t *testing.T) {
	envCfg, err := config.LoadEnvConfig("o8n-env.yaml")
	if err != nil {
		t.Fatalf("failed to load env config: %v", err)
	}
	appCfg, err := config.LoadAppConfig("o8n-cfg.yaml")
	if err != nil {
		t.Fatalf("failed to load app config: %v", err)
	}
	m := newModelEnvApp(envCfg, appCfg, "")
	a := config.ActionDef{Body: `{"userId":"{currentUser}"}`}
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		t.Fatalf("current env %q not found in model config", m.currentEnv)
	}
	b := strings.ReplaceAll(a.Body, "{currentUser}", env.Username)
	exp := `{"userId":"` + env.Username + `"}`
	if b != exp {
		t.Fatalf("expected body %q, got %q", exp, b)
	}
}

func TestVariableInstanceEditOpensModal(t *testing.T) {
	envCfg, err := config.LoadEnvConfig("o8n-env.yaml")
	if err != nil {
		t.Fatalf("load env config: %v", err)
	}
	appCfg, err := config.LoadAppConfig("o8n-cfg.yaml")
	if err != nil {
		t.Fatalf("load app config: %v", err)
	}
	m := newModelEnvApp(envCfg, appCfg, "")
	m.breadcrumb = []string{"variable-instance"}
	m.table.SetRows([]table.Row{{"name1", "val1", "proc1"}})
	m.table.SetCursor(0)
	if errMsg := m.startEdit(m.currentTableKey()); errMsg != "" {
		t.Fatalf("startEdit failed: %s", errMsg)
	}
	if m.activeModal != ModalEdit {
		t.Fatalf("expected activeModal ModalEdit, got %v", m.activeModal)
	}
}

func TestSearchModeClearedOnDrilldown(t *testing.T) {
	envCfg, err := config.LoadEnvConfig("o8n-env.yaml")
	if err != nil {
		t.Fatalf("load env config: %v", err)
	}
	appCfg, err := config.LoadAppConfig("o8n-cfg.yaml")
	if err != nil {
		t.Fatalf("load app config: %v", err)
	}
	m := newModelEnvApp(envCfg, appCfg, "")
	m.searchMode = true
	m.prepareStateTransition(TransitionDrillDown)
	if m.searchMode {
		t.Fatalf("expected searchMode false after drilldown, got true")
	}
}
