package app

import (
	"testing"
)

func TestSkinColorLookup(t *testing.T) {
	s := &Skin{}
	s.Colors.Accent = "#2F81F7"
	s.Colors.Danger = "#F85149"
	s.Colors.Fg = "#E6EDF3"

	if got := s.Color("accent"); got != "#2F81F7" {
		t.Errorf("accent: want #2F81F7, got %q", got)
	}
	if got := s.Color("danger"); got != "#F85149" {
		t.Errorf("danger: want #F85149, got %q", got)
	}
	if got := s.Color("fg"); got != "#E6EDF3" {
		t.Errorf("fg: want #E6EDF3, got %q", got)
	}
}

func TestSkinMissingRoleReturnsEmpty(t *testing.T) {
	s := &Skin{}
	if got := s.Color("accent"); got != "" {
		t.Errorf("unset role should return empty string, got %q", got)
	}
	if got := s.Color("nonexistent"); got != "" {
		t.Errorf("unknown role should return empty string, got %q", got)
	}
}

func TestSkinNilColorReturnsEmpty(t *testing.T) {
	var s *Skin
	if got := s.Color("accent"); got != "" {
		t.Errorf("nil skin should return empty string, got %q", got)
	}
}

func TestSkinMigrateFromLegacy(t *testing.T) {
	s := &Skin{}
	s.O8n.Body.FgColor = "#FFFFFF"
	s.O8n.Body.BgColor = "#000000"
	s.O8n.Body.LogoColor = "#FF6600"
	s.O8n.Frame.Border.FocusColor = "#00AAFF"
	s.O8n.Frame.Status.ErrorColor = "#FF0000"
	s.O8n.Frame.Status.AddColor = "#00FF00"
	s.O8n.Views.YAML.KeyColor = "#AABBCC"

	s.migrateFromLegacy()

	if s.Colors.Fg != "#FFFFFF" {
		t.Errorf("fg: want #FFFFFF, got %q", s.Colors.Fg)
	}
	if s.Colors.Bg != "#000000" {
		t.Errorf("bg: want #000000, got %q", s.Colors.Bg)
	}
	if s.Colors.AccentAlt != "#FF6600" {
		t.Errorf("accentAlt: want #FF6600, got %q", s.Colors.AccentAlt)
	}
	if s.Colors.BorderFocus != "#00AAFF" {
		t.Errorf("borderFocus: want #00AAFF, got %q", s.Colors.BorderFocus)
	}
	if s.Colors.Danger != "#FF0000" {
		t.Errorf("danger: want #FF0000, got %q", s.Colors.Danger)
	}
	if s.Colors.Success != "#00FF00" {
		t.Errorf("success: want #00FF00, got %q", s.Colors.Success)
	}
	if s.Colors.JSONKey != "#AABBCC" {
		t.Errorf("jsonKey: want #AABBCC, got %q", s.Colors.JSONKey)
	}
}

func TestSkinMigrateDoesNotOverwriteExisting(t *testing.T) {
	s := &Skin{}
	s.Colors.Fg = "#EXPLICIT"
	s.O8n.Body.FgColor = "#LEGACY"
	s.migrateFromLegacy()
	if s.Colors.Fg != "#EXPLICIT" {
		t.Errorf("migrate should not overwrite existing value, got %q", s.Colors.Fg)
	}
}

func TestLoadSkinFile(t *testing.T) {
	// stock.yaml must load without error
	skin, err := loadSkin("stock.yaml")
	if err != nil {
		t.Fatalf("loadSkin(stock.yaml): %v", err)
	}
	if skin == nil {
		t.Fatal("expected non-nil skin")
	}
}

func TestLoadSkinFileNotFound(t *testing.T) {
	_, err := loadSkin("nonexistent-skin-xyz.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent skin file")
	}
}

func TestAllRolesAccessible(t *testing.T) {
	roles := []string{
		"bg", "fg", "fgMuted", "accent", "accentAlt", "surface", "surfaceAlt",
		"success", "warning", "danger", "info", "borderFg", "borderFocus",
		"crumbBg", "crumbFg", "crumbActiveBg", "crumbActiveFg",
		"jsonKey", "jsonValue", "jsonNumber", "jsonBool",
		"btnPrimaryBg", "btnPrimaryFg", "btnSecondaryBg", "btnSecondaryFg",
	}
	s := &Skin{}
	// Just ensure Color() doesn't panic for any known role
	for _, role := range roles {
		_ = s.Color(role)
	}
}
