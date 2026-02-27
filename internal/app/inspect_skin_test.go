package app

import (
	"github.com/kthoms/o8n/internal/config"
	"testing"
)

func TestInspectSkin(t *testing.T) {
	envCfg, err := config.LoadEnvConfig("../..//o8n-env.yaml")
	if err != nil {
		// Try root path
		envCfg, err = config.LoadEnvConfig("o8n-env.yaml")
		if err != nil {
			t.Fatalf("failed to load env config: %v", err)
		}
	}
	appCfg, err := config.LoadAppConfig("o8n-cfg.yaml")
	if err != nil {
		t.Fatalf("failed to load app config: %v", err)
	}
	state, err := config.LoadAppState("o8n-stat.yml")
	if err != nil {
		t.Fatalf("failed to load app state: %v", err)
	}
	t.Logf("state.Skin=%q", state.Skin)
	m := newModelEnvApp(envCfg, appCfg, state.Skin)
	if m.skin == nil {
		t.Fatalf("m.skin == nil (activeSkin=%q)", m.activeSkin)
	}
	t.Logf("m.activeSkin=%q", m.activeSkin)
	t.Logf("m.skin.Color('borderFocus')=%q", m.skin.Color("borderFocus"))
	t.Logf("m.styles.BorderFocus.Foreground=%q", m.styles.BorderFocus.GetForeground())
}

func TestInspectNarsingh(t *testing.T) {
	// Ensure narsingh skin file loads and produces colors
	envCfg, err := config.LoadEnvConfig("o8n-env.yaml")
	if err != nil {
		t.Fatalf("failed to load env config: %v", err)
	}
	appCfg, err := config.LoadAppConfig("o8n-cfg.yaml")
	if err != nil {
		t.Fatalf("failed to load app config: %v", err)
	}
	m := newModelEnvApp(envCfg, appCfg, "narsingh")
	if m.skin == nil {
		t.Fatalf("narsingh skin not loaded; m.skin == nil (activeSkin=%q)", m.activeSkin)
	}
	t.Logf("narsingh m.activeSkin=%q", m.activeSkin)
	t.Logf("narsingh borderFocus=%q", m.skin.Color("borderFocus"))
	t.Logf("narsingh fg=%q", m.skin.Color("fg"))
}
