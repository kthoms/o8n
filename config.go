package main

import (
	"fmt"
	"os"

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

// ColumnDef defines a table column in the UI config
type ColumnDef struct {
	Name      string `yaml:"name"`
	Visible   bool   `yaml:"visible"`
	Width     string `yaml:"width"` // percentage like "25%" or empty for auto
	Align     string `yaml:"align"` // left/right/center (currently informational)
	Editable  bool   `yaml:"editable,omitempty"`
	InputType string `yaml:"input_type,omitempty"` // text/number/bool/auto (optional)
}

// DrillDownDef describes a drill-down target for a table (target collection and query parameter)
type DrillDownDef struct {
	Target string `yaml:"target"`           // target table name (e.g. process-instance)
	Param  string `yaml:"param"`            // query parameter to set on target (e.g. processInstanceId)
	Column string `yaml:"column,omitempty"` // source column to read the value from (defaults to id)
}

// TableDef defines a named table and its columns
type TableDef struct {
	Name      string         `yaml:"name"`
	Columns   []ColumnDef    `yaml:"columns"`
	Drilldown []DrillDownDef `yaml:"drilldown,omitempty"`
}

// EnvConfig holds environment-specific configuration (moved to o8n-env.yaml)
type EnvConfig struct {
	Environments map[string]Environment `yaml:"environments"`
	Active       string                 `yaml:"active,omitempty"`
}

// AppConfig holds application-level configuration (moved to o8n-cfg.yaml)
type AppConfig struct {
	Tables []TableDef `yaml:"tables,omitempty"`
}

// LoadEnvConfig loads the environment YAML file (o8n-env.yaml)
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

// Compatibility Config used by the rest of the app
type Config struct {
	Environments map[string]Environment `yaml:"environments"`
	Active       string                 `yaml:"active,omitempty"`
	Tables       []TableDef             `yaml:"tables,omitempty"`
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
	cfg.Tables = appCfg.Tables
	return cfg, nil
}

// LoadConfig attempts to load legacy config file at _path; it returns an error if the file is missing or invalid.
func LoadConfig(_path string) (*Config, error) {
	data, err := os.ReadFile(_path)
	if err != nil {
		return nil, fmt.Errorf("failed to read legacy config %s: %w", _path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse legacy config %s: %w", _path, err)
	}
	return &cfg, nil
}

// SaveConfig persists the env.Active field back to o8n-env.yaml (best-effort)
func SaveConfig(_path string, cfg *Config) error {
	// Write out env file if possible
	envCfg := &EnvConfig{Environments: cfg.Environments, Active: cfg.Active}
	if err := SaveEnvConfig("o8n-env.yaml", envCfg); err != nil {
		return err
	}
	// Also persist app config tables if provided
	appCfg := &AppConfig{Tables: cfg.Tables}
	if err := SaveAppConfig("o8n-cfg.yaml", appCfg); err != nil {
		// non-fatal
		return err
	}
	return nil
}
