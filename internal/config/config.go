// Package config provides configuration management for o8n
package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Environment represents a connection environment for Operaton
// Note: For production use, consider storing sensitive credentials like passwords
// in environment variables or a secure secrets manager rather than in config files
type Environment struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"` // Consider using environment variables for sensitive data
	UIColor  string `yaml:"ui_color"` // hex string like "#FF5733"
}

// ColumnDef defines a table column in the UI config.
//
// Column visibility defaults to true (visible). Set Visible to an explicit false pointer to permanently hide.
// Width is in characters; 0 means derive from Type. MinWidth defaults to the header title length.
// HideOrder controls when a column is hidden when space is tight:
//   - 0 (default/unset): eligible to be hidden first; among these, rightmost column hides first
//   - positive N: hidden after all 0-order columns; lower N hidden before higher N
type ColumnDef struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type,omitempty"`       // string/bool/int/float/datetime/id; drives implicit width
	Width     int    `yaml:"width,omitempty"`       // initial width in chars (0 = derive from Type)
	MinWidth  int    `yaml:"min_width,omitempty"`   // minimum width in chars (0 = use header title length)
	HideOrder int    `yaml:"hide_order,omitempty"`  // 0 = hidden first (rightmost first); positive = hidden later
	Align     string `yaml:"align,omitempty"`       // left/right/center
	Editable  bool   `yaml:"editable,omitempty"`   // whether this column can be edited
	InputType string `yaml:"input_type,omitempty"` // text/number/bool/auto (optional)
	Visible   *bool  `yaml:"visible,omitempty"`    // nil = visible by default; explicit false = always hidden
}

// UnmarshalYAML implements custom YAML parsing for ColumnDef.
// It handles the legacy width format ("25%", "") by ignoring percent/empty values,
// accepting plain integer widths in the new format.
func (c *ColumnDef) UnmarshalYAML(value *yaml.Node) error {
	// Use a raw struct to capture width as a string node for special handling
	type rawCol struct {
		Name      string    `yaml:"name"`
		Type      string    `yaml:"type"`
		Width     yaml.Node `yaml:"width"`
		MinWidth  int       `yaml:"min_width"`
		HideOrder int       `yaml:"hide_order"`
		Align     string    `yaml:"align"`
		Editable  bool      `yaml:"editable"`
		InputType string    `yaml:"input_type"`
		Visible   *bool     `yaml:"visible"`
	}
	var raw rawCol
	if err := value.Decode(&raw); err != nil {
		return err
	}
	c.Name = raw.Name
	c.Type = raw.Type
	c.MinWidth = raw.MinWidth
	c.HideOrder = raw.HideOrder
	c.Align = raw.Align
	c.Editable = raw.Editable
	c.InputType = raw.InputType
	c.Visible = raw.Visible
	// parse width: accept integer values; ignore percent strings ("25%") and empty
	if raw.Width.Kind == yaml.ScalarNode {
		v := strings.TrimSpace(raw.Width.Value)
		if v != "" && !strings.HasSuffix(v, "%") {
			fmt.Sscan(v, &c.Width)
		}
	}
	return nil
}

// IsVisible reports whether the column should be displayed.
// Columns with no Visible field set (nil) are visible by default.
func (c ColumnDef) IsVisible() bool {
	return c.Visible == nil || *c.Visible
}

// DefaultTypeWidth returns the default initial column width in characters for a given type.
func DefaultTypeWidth(colType string) int {
	switch strings.ToLower(colType) {
	case "bool":
		return 6
	case "int":
		return 8
	case "float":
		return 10
	case "datetime":
		return 20
	case "id":
		return 36
	default: // "string" or anything else
		return 20
	}
}

// HideSequence returns the column indices from cols in the order they should be hidden
// when available width is insufficient. Columns with HideOrder == 0 come first (rightmost
// position first), followed by configured columns in ascending HideOrder (rightmost first
// within each group).
func HideSequence(cols []ColumnDef) []int {
	seq := make([]int, 0, len(cols))
	// group 0: unspecified hide_order — rightmost (highest index) first
	for i := len(cols) - 1; i >= 0; i-- {
		if cols[i].HideOrder == 0 {
			seq = append(seq, i)
		}
	}
	// groups > 0: ascending HideOrder, rightmost first within each group
	groups := map[int][]int{}
	for i, col := range cols {
		if col.HideOrder > 0 {
			groups[col.HideOrder] = append(groups[col.HideOrder], i)
		}
	}
	keys := make([]int, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		idxs := groups[k]
		sort.Sort(sort.Reverse(sort.IntSlice(idxs)))
		seq = append(seq, idxs...)
	}
	return seq
}

// DrillDownDef describes a drill-down target for a table (target collection and query parameter)
type DrillDownDef struct {
	Target string `yaml:"target"`           // target table name (e.g. process-instance)
	Param  string `yaml:"param"`            // query parameter to set on target (e.g. processInstanceId)
	Column string `yaml:"column,omitempty"` // source column to read the value from (defaults to id)
	Label  string `yaml:"label,omitempty"`  // breadcrumb label for the drilldown (defaults to target name)
}

// EditActionDef describes how to save an edited cell value via a REST call.
// Path and BodyTemplate support placeholders: {id}, {name}, {parentId}, {value}, {type}.
type EditActionDef struct {
	Method       string `yaml:"method"`                  // HTTP method: PUT, POST, PATCH
	Path         string `yaml:"path"`                    // URL path template (e.g. /process-instance/{parentId}/variables/{name})
	BodyTemplate string `yaml:"body_template"`           // JSON body template (e.g. {"value": {value}, "type": "{type}"})
	IDColumn     string `yaml:"id_column,omitempty"`     // row column used as {id} (defaults to "id")
	NameColumn   string `yaml:"name_column,omitempty"`   // row column used as {name} (defaults to "name")
}

// ActionDef defines an action that can be performed on a selected item
type ActionDef struct {
	Key      string `yaml:"key"`                 // shortcut key (e.g. "s", "r", "ctrl+d")
	Label    string `yaml:"label"`               // display label (e.g. "Suspend Instance")
	Method   string `yaml:"method"`              // HTTP method: GET, PUT, POST, DELETE
	Path     string `yaml:"path"`                // URL path with {id} placeholder (e.g. /process-instance/{id}/suspended)
	Body     string `yaml:"body,omitempty"`       // optional JSON body to send
	Confirm  bool   `yaml:"confirm,omitempty"`    // require double-press confirmation
	IDColumn string `yaml:"id_column,omitempty"`  // column to read ID from (defaults to "id")
}

// TableDef defines a named table and its columns
type TableDef struct {
	Name       string         `yaml:"name"`
	ApiPath    string         `yaml:"api_path,omitempty"`    // REST collection path (defaults to /{name})
	CountPath  string         `yaml:"count_path,omitempty"`  // count endpoint (defaults to {api_path}/count)
	Columns    []ColumnDef    `yaml:"columns"`
	Drilldown  []DrillDownDef `yaml:"drilldown,omitempty"`
	Actions    []ActionDef    `yaml:"actions,omitempty"`
	EditAction *EditActionDef `yaml:"edit_action,omitempty"` // generic save config for editable columns
}

// EnvConfig holds environment-specific configuration (moved to o8n-env.yaml)
type EnvConfig struct {
	Environments map[string]Environment `yaml:"environments"`
	Active       string                 `yaml:"active,omitempty"`
	Skin         string                 `yaml:"skin,omitempty"`
}

// AppConfig holds application-level configuration (moved to o8n-cfg.yaml)
type AppConfig struct {
	Tables []TableDef `yaml:"tables,omitempty"`
	UI     *UIConfig  `yaml:"ui,omitempty"`
}

// Config is a compatibility type combining environment and app config
type Config struct {
	Environments map[string]Environment `yaml:"environments"`
	Active       string                 `yaml:"active,omitempty"`
	Skin         string                 `yaml:"skin,omitempty"`
	Tables       []TableDef             `yaml:"tables,omitempty"`
	UI           *UIConfig              `yaml:"ui,omitempty"`
}

// UIConfig holds UI-related configuration, e.g., edit modal styling
type UIConfig struct {
	EditModal *EditModalConfig `yaml:"edit_modal,omitempty"`
}

type EditModalConfig struct {
	Width          int               `yaml:"width,omitempty"`
	Border         string            `yaml:"border,omitempty"`
	BorderColor    string            `yaml:"border_color,omitempty"`
	OverlayOpacity float64           `yaml:"overlay_opacity,omitempty"`
	Buttons        *EditModalButtons `yaml:"buttons,omitempty"`
}

type EditModalButtons struct {
	Save   *ButtonStyle `yaml:"save,omitempty"`
	Cancel *ButtonStyle `yaml:"cancel,omitempty"`
}

type ButtonStyle struct {
	Label      string `yaml:"label,omitempty"`
	Key        string `yaml:"key,omitempty"`
	Background string `yaml:"background,omitempty"`
	Foreground string `yaml:"foreground,omitempty"`
	Style      string `yaml:"style,omitempty"`
}

// NavState captures the user's last navigation position for state restore.
type NavState struct {
	Root                 string            `yaml:"root,omitempty"`
	Breadcrumb           []string          `yaml:"breadcrumb,omitempty"`
	SelectedDefinitionKey string           `yaml:"selected_definition_key,omitempty"`
	SelectedInstanceID   string            `yaml:"selected_instance_id,omitempty"`
	GenericParams        map[string]string `yaml:"generic_params,omitempty"`
}

// AppState holds mutable runtime state persisted to o8n-stat.yml.
// This keeps o8n-env.yaml stable (credentials only) and separates user preferences.
type AppState struct {
	ActiveEnv   string   `yaml:"active_env,omitempty"`
	Skin        string   `yaml:"skin,omitempty"`
	ShowLatency bool     `yaml:"show_latency,omitempty"`
	Navigation  NavState `yaml:"navigation,omitempty"`
}

// LoadAppState loads runtime state from the given path (o8n-stat.yml).
// Returns an empty state (not an error) if the file does not exist yet.
func LoadAppState(path string) (*AppState, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &AppState{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read state file %s: %w", path, err)
	}
	var s AppState
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse state file %s: %w", path, err)
	}
	return &s, nil
}

// SaveAppState writes runtime state to the given path with 0600 permissions.
func SaveAppState(path string, s *AppState) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal app state: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file %s: %w", path, err)
	}
	return nil
}


func LoadEnvConfig(path string) (*EnvConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read env config file %s: %w", path, err)
	}
	var cfg EnvConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env config file %s: %w", path, err)
	}
	return &cfg, nil
}

// SaveEnvConfig writes env config to the given path with 0600 permissions
func SaveEnvConfig(path string, cfg *EnvConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal env config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write env config file %s: %w", path, err)
	}
	return nil
}

// LoadAppConfig loads application config (o8n-cfg.yaml)
func LoadAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read app config file %s: %w", path, err)
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse app config file %s: %w", path, err)
	}
	return &cfg, nil
}

// SaveAppConfig writes the app config to the given path with 0600 permissions
func SaveAppConfig(path string, cfg *AppConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal app config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write app config file %s: %w", path, err)
	}
	return nil
}

// LoadSplitConfig loads configuration from the split files o8n-env.yaml and o8n-cfg.yaml
func LoadSplitConfig() (*Config, error) {
	envCfg, envErr := LoadEnvConfig("o8n-env.yaml")
	appCfg, appErr := LoadAppConfig("o8n-cfg.yaml")
	if envErr != nil || appErr != nil {
		return nil, fmt.Errorf("failed to load split configs (env: %v, app: %v)", envErr, appErr)
	}
	cfg := &Config{}
	cfg.Environments = envCfg.Environments
	cfg.Active = envCfg.Active
	cfg.Skin = envCfg.Skin
	cfg.Tables = appCfg.Tables
	return cfg, nil
}

// LoadConfig attempts to load legacy config file at path; it returns an error if the file is missing or invalid.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read legacy config %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse legacy config %s: %w", path, err)
	}
	return &cfg, nil
}

// SaveConfig persists the env.Active and env.Skin fields back to o8n-env.yaml (best-effort).
// WARNING: This only saves environment configuration, NOT table definitions.
// Table definitions in o8n-cfg.yaml are static and should never be programmatically overwritten.
func SaveConfig(path string, cfg *Config) error {
	// Only persist environment settings (which environment is active)
	// NEVER touch o8n-cfg.yaml (table definitions) - it is a static configuration file
	envCfg := &EnvConfig{Environments: cfg.Environments, Active: cfg.Active, Skin: cfg.Skin}
	if err := SaveEnvConfig("o8n-env.yaml", envCfg); err != nil {
		return err
	}
	return nil
}
