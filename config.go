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

// Config represents the application configuration
type Config struct {
	Environments map[string]Environment `yaml:"environments"`
}

// ProcessDefinition represents a BPMN process definition from the Operaton REST API
type ProcessDefinition struct {
	ID           string `json:"id"`
	Key          string `json:"key"`
	Category     string `json:"category"`
	Description  string `json:"description"`
	Name         string `json:"name"`
	Version      int    `json:"version"`
	Resource     string `json:"resource"`
	DeploymentID string `json:"deploymentId"`
	Diagram      string `json:"diagram"`
	Suspended    bool   `json:"suspended"`
	TenantID     string `json:"tenantId"`
}

// ProcessInstance represents a BPMN process instance from the Operaton REST API
type ProcessInstance struct {
	ID             string `json:"id"`
	DefinitionID   string `json:"definitionId"`
	BusinessKey    string `json:"businessKey"`
	CaseInstanceID string `json:"caseInstanceId"`
	Ended          bool   `json:"ended"`
	Suspended      bool   `json:"suspended"`
	TenantID       string `json:"tenantId"`
	StartTime      string `json:"startTime"`
	EndTime        string `json:"endTime"`
}

// LoadConfig loads and parses a YAML configuration file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}
