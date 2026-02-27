package app

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Colors holds the 25 semantic color roles that drive all UI styling.
// Empty string means "terminal default" for that role.
type Colors struct {
	Bg             string `yaml:"bg"`
	Fg             string `yaml:"fg"`
	FgMuted        string `yaml:"fgMuted"`
	Accent         string `yaml:"accent"`
	AccentAlt      string `yaml:"accentAlt"`
	Surface        string `yaml:"surface"`
	SurfaceAlt     string `yaml:"surfaceAlt"`
	Success        string `yaml:"success"`
	Warning        string `yaml:"warning"`
	Danger         string `yaml:"danger"`
	Info           string `yaml:"info"`
	BorderFg       string `yaml:"borderFg"`
	BorderFocus    string `yaml:"borderFocus"`
	CrumbBg        string `yaml:"crumbBg"`
	CrumbFg        string `yaml:"crumbFg"`
	CrumbActiveBg  string `yaml:"crumbActiveBg"`
	CrumbActiveFg  string `yaml:"crumbActiveFg"`
	JSONKey        string `yaml:"jsonKey"`
	JSONValue      string `yaml:"jsonValue"`
	JSONNumber     string `yaml:"jsonNumber"`
	JSONBool       string `yaml:"jsonBool"`
	BtnPrimaryBg   string `yaml:"btnPrimaryBg"`
	BtnPrimaryFg   string `yaml:"btnPrimaryFg"`
	BtnSecondaryBg string `yaml:"btnSecondaryBg"`
	BtnSecondaryFg string `yaml:"btnSecondaryFg"`
}

// Skin is the top-level skin definition loaded from a YAML file.
// The new schema uses a flat colors: section. For backward compatibility,
// the old nested o8n: structure is also read and used to fill missing roles.
type Skin struct {
	Colors Colors `yaml:"colors"`
	// Legacy nested structure for backward-compat reading
	O8n struct {
		Body struct {
			FgColor   string `yaml:"fgColor"`
			BgColor   string `yaml:"bgColor"`
			LogoColor string `yaml:"logoColor"`
		} `yaml:"body"`
		Frame struct {
			Border struct {
				FgColor    string `yaml:"fgColor"`
				FocusColor string `yaml:"focusColor"`
			} `yaml:"border"`
			Crumbs struct {
				FgColor     string `yaml:"fgColor"`
				BgColor     string `yaml:"bgColor"`
				ActiveColor string `yaml:"activeColor"`
			} `yaml:"crumbs"`
			Status struct {
				ErrorColor     string `yaml:"errorColor"`
				AddColor       string `yaml:"addColor"`
				ModifyColor    string `yaml:"modifyColor"`
				HighlightColor string `yaml:"highlightColor"`
				KillColor      string `yaml:"killColor"`
				CompletedColor string `yaml:"completedColor"`
			} `yaml:"status"`
		} `yaml:"frame"`
		Dialog struct {
			FgColor            string `yaml:"fgColor"`
			BgColor            string `yaml:"bgColor"`
			ButtonFgColor      string `yaml:"buttonFgColor"`
			ButtonBgColor      string `yaml:"buttonBgColor"`
			ButtonFocusFgColor string `yaml:"buttonFocusFgColor"`
			ButtonFocusBgColor string `yaml:"buttonFocusBgColor"`
			LabelFgColor       string `yaml:"labelFgColor"`
			FieldFgColor       string `yaml:"fieldFgColor"`
		} `yaml:"dialog"`
		Views struct {
			Table struct {
				FgColor       string `yaml:"fgColor"`
				BgColor       string `yaml:"bgColor"`
				CursorFgColor string `yaml:"cursorFgColor"`
				CursorBgColor string `yaml:"cursorBgColor"`
				Header        struct {
					FgColor string `yaml:"fgColor"`
					BgColor string `yaml:"bgColor"`
				} `yaml:"header"`
			} `yaml:"table"`
			YAML struct {
				KeyColor   string `yaml:"keyColor"`
				ValueColor string `yaml:"valueColor"`
			} `yaml:"yaml"`
		} `yaml:"views"`
	} `yaml:"o8n"`
}

// Color returns the color string for the given semantic role.
// Returns "" (terminal default) if the role is not set.
func (s *Skin) Color(role string) string {
	if s == nil {
		return ""
	}
	c := s.Colors
	switch role {
	case "bg":
		return c.Bg
	case "fg":
		return c.Fg
	case "fgMuted":
		return c.FgMuted
	case "accent":
		return c.Accent
	case "accentAlt":
		return c.AccentAlt
	case "surface":
		return c.Surface
	case "surfaceAlt":
		return c.SurfaceAlt
	case "success":
		return c.Success
	case "warning":
		return c.Warning
	case "danger":
		return c.Danger
	case "info":
		return c.Info
	case "borderFg":
		return c.BorderFg
	case "borderFocus":
		return c.BorderFocus
	case "crumbBg":
		return c.CrumbBg
	case "crumbFg":
		return c.CrumbFg
	case "crumbActiveBg":
		return c.CrumbActiveBg
	case "crumbActiveFg":
		return c.CrumbActiveFg
	case "jsonKey":
		return c.JSONKey
	case "jsonValue":
		return c.JSONValue
	case "jsonNumber":
		return c.JSONNumber
	case "jsonBool":
		return c.JSONBool
	case "btnPrimaryBg":
		return c.BtnPrimaryBg
	case "btnPrimaryFg":
		return c.BtnPrimaryFg
	case "btnSecondaryBg":
		return c.BtnSecondaryBg
	case "btnSecondaryFg":
		return c.BtnSecondaryFg
	}
	return ""
}

// migrateFromLegacy fills missing Colors roles from the old nested o8n.* structure.
func (s *Skin) migrateFromLegacy() {
	c := &s.Colors
	o := s.O8n
	fill := func(dest *string, src string) {
		if *dest == "" && src != "" {
			*dest = src
		}
	}
	fill(&c.Bg, o.Body.BgColor)
	fill(&c.Fg, o.Body.FgColor)
	fill(&c.AccentAlt, o.Body.LogoColor)
	fill(&c.BorderFg, o.Frame.Border.FgColor)
	fill(&c.BorderFocus, o.Frame.Border.FocusColor)
	fill(&c.Accent, o.Frame.Border.FocusColor)
	fill(&c.CrumbFg, o.Frame.Crumbs.FgColor)
	fill(&c.CrumbBg, o.Frame.Crumbs.BgColor)
	fill(&c.CrumbActiveBg, o.Frame.Crumbs.ActiveColor)
	fill(&c.CrumbActiveFg, o.Body.FgColor)
	fill(&c.Danger, o.Frame.Status.ErrorColor)
	fill(&c.Success, o.Frame.Status.AddColor)
	fill(&c.Info, o.Frame.Status.ModifyColor)
	fill(&c.Warning, o.Frame.Status.HighlightColor)
	fill(&c.Surface, o.Views.Table.Header.BgColor)
	fill(&c.SurfaceAlt, o.Views.Table.CursorBgColor)
	fill(&c.JSONKey, o.Views.YAML.KeyColor)
	fill(&c.JSONValue, o.Views.YAML.ValueColor)
	fill(&c.BtnPrimaryBg, o.Dialog.ButtonBgColor)
	fill(&c.BtnPrimaryFg, o.Dialog.ButtonFgColor)
	fill(&c.BtnSecondaryBg, o.Body.BgColor)
	fill(&c.BtnSecondaryFg, o.Frame.Status.CompletedColor)
	fill(&c.FgMuted, o.Frame.Status.CompletedColor)
}

func loadSkin(skinName string) (*Skin, error) {
	if skinName == "" {
		skinName = "stock.yaml"
	}
	path := filepath.Join("skins", skinName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read skin file %s: %w", path, err)
	}

	var skin Skin
	if err := yaml.Unmarshal(data, &skin); err != nil {
		return nil, fmt.Errorf("failed to parse skin file %s: %w", path, err)
	}

	// If new colors section is empty, migrate from legacy nested structure
	if skin.Colors.Fg == "" && skin.Colors.Bg == "" {
		skin.migrateFromLegacy()
	}

	return &skin, nil
}
