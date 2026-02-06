package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateCyclesEnvironmentClearsViews(t *testing.T) {
	cfg := &Config{
		Environments: map[string]Environment{
			"dev":  {UIColor: "#ff0000"},
			"prod": {UIColor: "#00ff00"},
		},
	}

	m := newModel(cfg)
	m.list.SetItems([]list.Item{processDefinitionItem{name: "one"}})
	m.table.SetRows([]table.Row{{"i1", "d1"}})
	m.selectedInstanceID = "i1"
	m.showKillModal = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	got := updated.(model)

	if got.currentEnv != "prod" {
		t.Fatalf("expected currentEnv to cycle to prod, got %s", got.currentEnv)
	}
	if len(got.list.Items()) != 0 {
		t.Fatalf("expected list to be cleared")
	}
	if len(got.table.Rows()) != 0 {
		t.Fatalf("expected table to be cleared")
	}
	if got.selectedInstanceID != "" {
		t.Fatalf("expected selectedInstanceID to reset")
	}
	if got.showKillModal {
		t.Fatalf("expected kill modal to close when switching env")
	}
}

func TestTerminateInstanceWhenModalConfirmed(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	cfg := &Config{
		Environments: map[string]Environment{
			"test": {URL: server.URL},
		},
	}

	m := newModel(cfg)
	m.showKillModal = true
	m.selectedInstanceID = "abc"

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if updated.(model).showKillModal {
		t.Fatalf("expected modal to close after confirmation")
	}
	if cmd == nil {
		t.Fatalf("expected terminate command to be returned")
	}
	msg := cmd()
	if _, ok := msg.(terminateResultMsg); !ok {
		t.Fatalf("expected terminateResultMsg, got %T", msg)
	}
	if !called {
		t.Fatalf("expected terminate endpoint to be called")
	}
}

func TestAutoRefreshTick(t *testing.T) {
	cfg := &Config{Environments: map[string]Environment{}}
	m := newModel(cfg)

	var tickDuration time.Duration
	m.tick = func(d time.Duration) tea.Cmd {
		tickDuration = d
		return func() tea.Msg { return refreshMsg{} }
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m2 := updated.(model)

	if !m2.autoRefresh {
		t.Fatalf("expected autoRefresh to be enabled")
	}
	if tickDuration != 5*time.Second {
		t.Fatalf("expected tick duration 5s, got %v", tickDuration)
	}
	if cmd == nil {
		t.Fatalf("expected tick command")
	}
	if _, ok := cmd().(refreshMsg); !ok {
		t.Fatalf("expected refreshMsg from tick command")
	}

	_, cmd2 := m2.Update(refreshMsg{})
	if cmd2 == nil {
		t.Fatalf("expected command batch from refresh handling")
	}
}
