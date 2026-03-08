package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/kthoms/o6n/internal/config"
	"gopkg.in/yaml.v3"
)

// ── Story 5.2: Credential Security — Permission Check ────────────────────────

func TestLoadEnvConfig_WarnOnTooOpenPermissions(t *testing.T) {
	if os.Getenv("CI") != "" {
		// Some CI environments run as root; skip permission test there.
		t.Skip("skipping permission check in CI root environment")
	}
	f, err := os.CreateTemp("", "o6n-env-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString("environments:\n  local:\n    url: http://localhost\n")
	f.Close()

	// Set overly permissive mode (world-readable) — should trigger warning.
	if err := os.Chmod(f.Name(), 0o644); err != nil {
		t.Skipf("chmod not supported: %v", err)
	}

	_, err = config.LoadEnvConfig(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.PermissionWarning == "" {
		t.Error("expected a non-empty PermissionWarning for 0644 permissions")
	}
}

func TestLoadEnvConfig_NoWarnOnRestrictedPermissions(t *testing.T) {
	f, err := os.CreateTemp("", "o6n-env-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString("environments:\n  local:\n    url: http://localhost\n")
	f.Close()

	// Set strictly correct mode — should NOT trigger warning.
	if err := os.Chmod(f.Name(), 0o600); err != nil {
		t.Skipf("chmod not supported: %v", err)
	}

	_, err = config.LoadEnvConfig(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.PermissionWarning != "" {
		t.Errorf("expected empty PermissionWarning for 0600 permissions, got: %s", config.PermissionWarning)
	}
}

func TestTableDefNewFields_RoundTrip(t *testing.T) {
	raw := `
tables:
  - name: history-process-instance
    api_path: /history/process-instance
    count_path: /history/process-instance/count
    columns:
      - name: id
    drilldown:
      target: history-activity-instance
      param: processInstanceId
      column: id
      label: Activity Instances
    edit_action:
      method: PUT
      path: /process-instance/{parentId}/variables/{name}
      body_template: '{"value": {value}, "type": "{type}"}'
      id_column: id
      name_column: name
`
	var cfg config.AppConfig
	if err := yaml.NewDecoder(strings.NewReader(raw)).Decode(&cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(cfg.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(cfg.Tables))
	}
	tbl := cfg.Tables[0]

	if tbl.ApiPath != "/history/process-instance" {
		t.Errorf("ApiPath: got %q, want %q", tbl.ApiPath, "/history/process-instance")
	}
	if tbl.CountPath != "/history/process-instance/count" {
		t.Errorf("CountPath: got %q, want %q", tbl.CountPath, "/history/process-instance/count")
	}
	if tbl.Drilldown == nil || tbl.Drilldown.Label != "Activity Instances" {
		t.Errorf("DrillDownDef.Label: got %q, want %q", tbl.Drilldown.Label, "Activity Instances")
	}
	if tbl.EditAction == nil {
		t.Fatal("EditAction is nil")
	}
	if tbl.EditAction.Method != "PUT" {
		t.Errorf("EditAction.Method: got %q, want PUT", tbl.EditAction.Method)
	}
	if tbl.EditAction.Path != "/process-instance/{parentId}/variables/{name}" {
		t.Errorf("EditAction.Path: got %q", tbl.EditAction.Path)
	}
	if tbl.EditAction.IDColumn != "id" {
		t.Errorf("EditAction.IDColumn: got %q, want id", tbl.EditAction.IDColumn)
	}
	if tbl.EditAction.NameColumn != "name" {
		t.Errorf("EditAction.NameColumn: got %q, want name", tbl.EditAction.NameColumn)
	}
}

func TestTableDef_ApiPathDefaults(t *testing.T) {
	raw := `
tables:
  - name: process-instance
    columns:
      - name: id
`
	var cfg config.AppConfig
	if err := yaml.NewDecoder(strings.NewReader(raw)).Decode(&cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	tbl := cfg.Tables[0]
	// When not set, ApiPath and CountPath should be empty strings (caller derives defaults)
	if tbl.ApiPath != "" {
		t.Errorf("expected empty ApiPath when not set, got %q", tbl.ApiPath)
	}
	if tbl.CountPath != "" {
		t.Errorf("expected empty CountPath when not set, got %q", tbl.CountPath)
	}
	if tbl.EditAction != nil {
		t.Errorf("expected nil EditAction when not set")
	}
}

func TestDrillDownDef_LabelOptional(t *testing.T) {
	raw := `
tables:
  - name: process-definition
    columns:
      - name: id
    drilldown:
      target: process-instance
      param: processDefinitionId
      column: id
`
	var cfg config.AppConfig
	if err := yaml.NewDecoder(strings.NewReader(raw)).Decode(&cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	dd := cfg.Tables[0].Drilldown
	if dd == nil {
		t.Fatal("expected drilldown present")
	}
	if dd.Label != "" {
		t.Errorf("expected empty Label when not set, got %q", dd.Label)
	}
	if dd.Target != "process-instance" {
		t.Errorf("Target: got %q", dd.Target)
	}
}
