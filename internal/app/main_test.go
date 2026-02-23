package app

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain wraps the entire test suite to protect sensitive files from being
// overwritten by tests that trigger UI actions (e.g. environment switching,
// which calls SaveConfig → writes o8n-env.yaml).
func TestMain(m *testing.M) {
	// Tests run with CWD = internal/app; walk up to project root so
	// relative paths like o8n-env.yaml and o8n-cfg.yaml resolve correctly.
	if root := findProjectRoot(); root != "" {
		_ = os.Chdir(root)
	}

	const envFile = "o8n-env.yaml"

	// Back up the current env file before any test runs.
	original, readErr := os.ReadFile(envFile)

	code := m.Run()

	// Restore original content regardless of test outcome.
	if readErr == nil {
		_ = os.WriteFile(envFile, original, 0600)
	}

	os.Exit(code)
}

// findProjectRoot walks up from the current directory until it finds go.mod.
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
