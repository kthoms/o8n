package app

import (
	"testing"

	"github.com/kthoms/o8n/internal/client"
	"github.com/kthoms/o8n/internal/config"
)

// TestLiveAPIIntegration tests against running Operaton instance at http://localhost:8080/engine-rest
func TestLiveAPIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	env := config.Environment{
		URL:      "http://localhost:8080/engine-rest",
		Username: "",
		Password: "",
	}

	c := client.NewClient(env, false)

	// Test 1: Fetch definitions
	defs, err := c.FetchProcessDefinitions()
	if err != nil {
		t.Fatalf("Failed to fetch definitions: %v", err)
	}
	if len(defs) == 0 {
		t.Fatal("Expected at least one definition, got none")
	}
	t.Logf("✓ Fetched %d definitions", len(defs))

	// Test 1b: Fetch definitions count
	count, err := c.FetchProcessDefinitionsCount()
	if err != nil {
		t.Logf("Warning: Failed to fetch definitions count: %v", err)
	} else {
		t.Logf("✓ Total process definitions count: %d", count)
	}

	// Test 2: Fetch instances
	instances, err := c.FetchInstances("", "")
	if err != nil {
		t.Fatalf("Failed to fetch instances: %v", err)
	}
	t.Logf("✓ Fetched %d instances", len(instances))

	// Test 3: Check for data consistency
	for _, def := range defs {
		if def.Key == "" {
			t.Error("❌ Definition key is empty")
		}
		if def.Name == "" {
			t.Errorf("❌ Definition %s has empty name", def.ID)
		}
		if def.Version <= 0 {
			t.Logf("Warning: Definition %s has invalid version: %d", def.Key, def.Version)
		}
	}

	// Test 4: Fetch variables for first instance if available
	if len(instances) > 0 {
		vars, err := c.FetchVariables(instances[0].ID)
		if err != nil {
			t.Logf("Note: Could not fetch variables for instance %s: %v", instances[0].ID, err)
		} else {
			t.Logf("✓ Fetched %d variables for instance %s", len(vars), instances[0].ID)
		}
	}

	t.Log("✓ All integration tests passed")
}
