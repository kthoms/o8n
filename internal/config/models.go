// Package config provides configuration management for o8n
package config

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

// Variable represents a process instance variable
type Variable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
}
