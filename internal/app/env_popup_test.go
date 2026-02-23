package app

import (
	"testing"

	"github.com/kthoms/o8n/internal/config"
)

func TestEnvPopupOpens(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	m2, _ := sendKeyString(m, "ctrl+e")

	if m2.activeModal != ModalEnvironment {
		t.Errorf("expected ModalEnvironment, got %v", m2.activeModal)
	}
}

func TestEnvPopupShowsAllEnvs(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"dev":     {URL: "http://dev", UIColor: "#FFA500"},
			"local":   {URL: "http://local", UIColor: "#00A8E1"},
			"staging": {URL: "http://staging", UIColor: "#50C878"},
		},
		Active: "local",
	}
	m := newModel(cfg)

	if len(m.envNames) != 3 {
		t.Errorf("expected 3 environments, got %d", len(m.envNames))
	}
}

func TestEnvPopupEscCloses(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalEnvironment

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Errorf("expected ModalNone after Esc, got %v", m2.activeModal)
	}
}

func TestEnvPopupEnterSwitches(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"dev":   {URL: "http://dev", UIColor: "#FFA500"},
			"local": {URL: "http://local", UIColor: "#00A8E1"},
		},
		Active: "local",
	}
	m := newModel(cfg)
	m.splashActive = false
	m.activeModal = ModalEnvironment

	// Find dev index
	devIdx := -1
	for i, name := range m.envNames {
		if name == "dev" {
			devIdx = i
			break
		}
	}
	if devIdx < 0 {
		t.Fatal("could not find 'dev' in envNames")
	}
	m.envPopupCursor = devIdx

	m2, cmd := sendKeyString(m, "enter")

	if m2.activeModal != ModalNone {
		t.Errorf("expected ModalNone after Enter, got %v", m2.activeModal)
	}
	if m2.currentEnv != "dev" {
		t.Errorf("expected currentEnv 'dev', got %q", m2.currentEnv)
	}
	if cmd == nil {
		t.Error("expected a non-nil command after switching environment")
	}
}

func TestEnvPopupNavigation(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"dev":   {URL: "http://dev", UIColor: "#FFA500"},
			"local": {URL: "http://local", UIColor: "#00A8E1"},
		},
		Active: "local",
	}
	m := newModel(cfg)
	m.splashActive = false
	m.activeModal = ModalEnvironment
	m.envPopupCursor = 0

	// Move down
	m2, _ := sendKeyString(m, "j")
	if m2.envPopupCursor != 1 {
		t.Errorf("expected envPopupCursor 1 after j, got %d", m2.envPopupCursor)
	}

	// Move up
	m3, _ := sendKeyString(m2, "k")
	if m3.envPopupCursor != 0 {
		t.Errorf("expected envPopupCursor 0 after k, got %d", m3.envPopupCursor)
	}
}

func TestEnvPopupEnterSameEnvNoSwitch(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"local": {URL: "http://local", UIColor: "#00A8E1"},
		},
		Active: "local",
	}
	m := newModel(cfg)
	m.splashActive = false
	m.activeModal = ModalEnvironment
	m.envPopupCursor = 0

	m2, cmd := sendKeyString(m, "enter")

	if m2.activeModal != ModalNone {
		t.Errorf("expected ModalNone, got %v", m2.activeModal)
	}
	// Same env — no fetch command should be issued
	if cmd != nil {
		t.Error("expected nil command when selecting same environment")
	}
}
