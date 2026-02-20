package main

import (
	"os"
	"testing"
)

// TestValidateConfigFiles verifies that critical config files are protected from deletion/corruption
func TestValidateConfigFiles(t *testing.T) {
	// This test ensures o8n-cfg.yaml and o8n-env.yaml are present and not corrupted
	err := validateConfigFiles()
	if err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	// Verify files have expected content
	cfgData, _ := os.ReadFile("o8n-cfg.yaml")
	if len(cfgData) < 100 {
		t.Fatalf("o8n-cfg.yaml is suspiciously small (%d bytes). This may indicate deletion/corruption.", len(cfgData))
	}

	envData, _ := os.ReadFile("o8n-env.yaml")
	if len(envData) < 50 {
		t.Fatalf("o8n-env.yaml is suspiciously small (%d bytes). This may indicate deletion/corruption.", len(envData))
	}

	t.Logf("✓ o8n-cfg.yaml: %d bytes (%d lines)", len(cfgData), len(string(cfgData))-len(string(cfgData)[:1]))
	t.Logf("✓ o8n-env.yaml: %d bytes", len(envData))
}

// TestConfigNotEmptied ensures o8n-cfg.yaml was not reduced to just "{}"
func TestConfigNotEmptied(t *testing.T) {
	data, err := os.ReadFile("o8n-cfg.yaml")
	if err != nil {
		t.Skip("o8n-cfg.yaml not found")
		return
	}

	// Check for the specific corruption case: file emptied to "{}"
	if string(data) == "{}" || string(data) == "{}\n" {
		t.Fatal("CRITICAL: o8n-cfg.yaml was emptied to '{}'. This indicates corruption from JSON parsing.")
	}

	// Should have ~760 lines of YAML
	lineCount := 0
	for _, c := range data {
		if c == '\n' {
			lineCount++
		}
	}

	if lineCount < 700 {
		t.Fatalf("o8n-cfg.yaml has only %d lines (expected ~760). File may be corrupted.", lineCount)
	}

	t.Logf("✓ o8n-cfg.yaml integrity verified: %d lines", lineCount)
}

// TestEnvFileNotEmptied ensures o8n-env.yaml was not corrupted
func TestEnvFileNotEmptied(t *testing.T) {
	data, err := os.ReadFile("o8n-env.yaml")
	if err != nil {
		t.Skip("o8n-env.yaml not found")
		return
	}

	if len(data) < 50 {
		t.Fatalf("o8n-env.yaml is too small (%d bytes). File may be corrupted.", len(data))
	}

	// Should contain 'environments:' key
	if !contains(string(data), "environments:") {
		t.Fatal("o8n-env.yaml is corrupted: missing 'environments:' key")
	}

	t.Logf("✓ o8n-env.yaml integrity verified: %d bytes", len(data))
}

// contains is a simple substring check
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
