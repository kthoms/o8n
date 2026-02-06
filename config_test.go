package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test config data
	configData := `environments:
  test:
    url: "http://localhost:8080/engine-rest"
    username: "testuser"
    password: "testpass"
    ui_color: "#00A8E1"
`
	if _, err := tmpFile.Write([]byte(configData)); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	tmpFile.Close()

	// Test loading the config
	config, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify the loaded config
	if config.Environments == nil {
		t.Fatal("Environments map is nil")
	}

	testEnv, ok := config.Environments["test"]
	if !ok {
		t.Fatal("Test environment not found")
	}

	if testEnv.URL != "http://localhost:8080/engine-rest" {
		t.Errorf("Expected URL 'http://localhost:8080/engine-rest', got '%s'", testEnv.URL)
	}

	if testEnv.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", testEnv.Username)
	}

	if testEnv.Password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", testEnv.Password)
	}

	if testEnv.UIColor != "#00A8E1" {
		t.Errorf("Expected UIColor '#00A8E1', got '%s'", testEnv.UIColor)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tmpFile, err := os.CreateTemp("", "invalid-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	invalidYAML := `environments:
  test:
    url: http://localhost:8080
    username: [invalid yaml structure
`
	if _, err := tmpFile.Write([]byte(invalidYAML)); err != nil {
		t.Fatalf("Failed to write invalid yaml: %v", err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}
