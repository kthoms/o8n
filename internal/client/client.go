// Package client provides API client functionality for Operaton
package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/operaton"
)

// Client represents an API client bound to a specific environment.
type Client struct {
	env         config.Environment
	httpClient  *http.Client
	operatonAPI *operaton.APIClient
	authContext context.Context
}

// NewClient creates a Client with a default timeout.
func NewClient(env config.Environment) *Client {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Create OpenAPI configuration
	cfg := operaton.NewConfiguration()
	cfg.HTTPClient = httpClient
	cfg.Servers = operaton.ServerConfigurations{
		{
			URL:         "{url}",
			Description: "Custom Operaton server",
			Variables: map[string]operaton.ServerVariable{
				"url": {
					Description:  "Server URL",
					DefaultValue: env.URL,
				},
			},
		},
	}

	// Create the generated API client
	apiClient := operaton.NewAPIClient(cfg)

	// Create context with basic auth
	auth := operaton.BasicAuth{
		UserName: env.Username,
		Password: env.Password,
	}
	authContext := context.WithValue(context.Background(), operaton.ContextBasicAuth, auth)

	return &Client{
		env:         env,
		httpClient:  httpClient,
		operatonAPI: apiClient,
		authContext: authContext,
	}
}

// FetchProcessDefinitions retrieves all process definitions using the generated client.
func (c *Client) FetchProcessDefinitions() ([]config.ProcessDefinition, error) {
	defs, _, err := c.operatonAPI.ProcessDefinitionAPI.GetProcessDefinitions(c.authContext).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch process definitions: %w", err)
	}

	// Convert from generated DTOs to our model
	result := make([]config.ProcessDefinition, len(defs))
	for i, def := range defs {
		result[i] = config.ProcessDefinition{
			ID:           getStringValue(def.Id),
			Key:          getStringValue(def.Key),
			Category:     getStringValue(def.Category),
			Description:  getStringValue(def.Description),
			Name:         getStringValue(def.Name),
			Version:      int(getInt32Value(def.Version)),
			Resource:     getStringValue(def.Resource),
			DeploymentID: getStringValue(def.DeploymentId),
			Diagram:      getStringValue(def.Diagram),
			Suspended:    getBoolValue(def.Suspended),
			TenantID:     getStringValue(def.TenantId),
		}
	}

	return result, nil
}

// FetchInstances retrieves process instances filtered by process key using the generated client.
func (c *Client) FetchInstances(processKey string) ([]config.ProcessInstance, error) {
	req := c.operatonAPI.ProcessInstanceAPI.GetProcessInstances(c.authContext)
	if processKey != "" {
		req = req.ProcessDefinitionKey(processKey)
	}

	instances, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch instances: %w", err)
	}

	// Convert from generated DTOs to our model
	result := make([]config.ProcessInstance, len(instances))
	for i, inst := range instances {
		result[i] = config.ProcessInstance{
			ID:             getStringValue(inst.Id),
			DefinitionID:   getStringValue(inst.DefinitionId),
			BusinessKey:    getStringValue(inst.BusinessKey),
			CaseInstanceID: getStringValue(inst.CaseInstanceId),
			Ended:          getBoolValue(inst.Ended),
			Suspended:      getBoolValue(inst.Suspended),
			TenantID:       getStringValue(inst.TenantId),
		}
	}

	return result, nil
}

// FetchVariables retrieves variables for a process instance using the generated client.
// The Operaton API returns a map of variable names to variable values.
func (c *Client) FetchVariables(instanceID string) ([]config.Variable, error) {
	varsMap, _, err := c.operatonAPI.ProcessInstanceAPI.GetProcessInstanceVariables(c.authContext, instanceID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch variables: %w", err)
	}

	// Convert from map to slice
	vars := make([]config.Variable, 0, len(*varsMap))
	for name, varValue := range *varsMap {
		value := ""
		if varValue.Value != nil {
			value = fmt.Sprintf("%v", varValue.Value)
		}
		vars = append(vars, config.Variable{
			Name:  name,
			Value: value,
			Type:  getStringValue(varValue.Type),
		})
	}

	return vars, nil
}

// SetProcessInstanceVariable updates a single variable on a process instance.
func (c *Client) SetProcessInstanceVariable(instanceID, varName string, value interface{}, valueType string) error {
	dto := operaton.VariableValueDto{Value: value}
	if valueType != "" {
		dto.SetType(valueType)
	}
	_, err := c.operatonAPI.ProcessInstanceAPI.SetProcessInstanceVariable(c.authContext, instanceID, varName).
		VariableValueDto(dto).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to set variable %s on instance %s: %w", varName, instanceID, err)
	}
	return nil
}

// TerminateInstance terminates a process instance using the generated client.
func (c *Client) TerminateInstance(instanceID string) error {
	_, err := c.operatonAPI.ProcessInstanceAPI.DeleteProcessInstance(c.authContext, instanceID).Execute()
	if err != nil {
		return fmt.Errorf("failed to terminate instance %s: %w", instanceID, err)
	}

	return nil
}

// Helper functions to safely handle Nullable types from the generated client
func getStringValue(nullable operaton.NullableString) string {
	ptr := nullable.Get()
	if ptr == nil {
		return ""
	}
	return *ptr
}

func getInt32Value(nullable operaton.NullableInt32) int32 {
	ptr := nullable.Get()
	if ptr == nil {
		return 0
	}
	return *ptr
}

func getBoolValue(nullable operaton.NullableBool) bool {
	ptr := nullable.Get()
	if ptr == nil {
		return false
	}
	return *ptr
}
