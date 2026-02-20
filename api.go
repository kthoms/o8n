package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kthoms/o8n/internal/client"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/operaton"
)

// Client represents an API client bound to a specific environment.
type Client struct {
	env         Environment
	httpClient  *http.Client
	operatonAPI *operaton.APIClient
	authContext context.Context
}

// NewClient creates a Client with a default timeout.
func NewClient(env Environment) *Client {
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

	result := make([]config.ProcessDefinition, len(defs))
	for i, def := range defs {
		result[i] = config.ProcessDefinition{
			ID:           client.GetStringValue(def.Id),
			Key:          client.GetStringValue(def.Key),
			Category:     client.GetStringValue(def.Category),
			Description:  client.GetStringValue(def.Description),
			Name:         client.GetStringValue(def.Name),
			Version:      int(client.GetInt32Value(def.Version)),
			Resource:     client.GetStringValue(def.Resource),
			DeploymentID: client.GetStringValue(def.DeploymentId),
			Diagram:      client.GetStringValue(def.Diagram),
			Suspended:    client.GetBoolValue(def.Suspended),
			TenantID:     client.GetStringValue(def.TenantId),
		}
	}

	return result, nil
}

// FetchProcessDefinitionsCount returns the total count of process definitions.
func (c *Client) FetchProcessDefinitionsCount() (int, error) {
	countResp, _, err := c.operatonAPI.ProcessDefinitionAPI.GetProcessDefinitionsCount(c.authContext).Execute()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch process definitions count: %w", err)
	}
	if countResp == nil || countResp.Count == nil {
		return 0, nil
	}
	return int(*countResp.Count), nil
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

	result := make([]config.ProcessInstance, len(instances))
	for i, inst := range instances {
		result[i] = config.ProcessInstance{
			ID:             client.GetStringValue(inst.Id),
			DefinitionID:   client.GetStringValue(inst.DefinitionId),
			BusinessKey:    client.GetStringValue(inst.BusinessKey),
			CaseInstanceID: client.GetStringValue(inst.CaseInstanceId),
			Ended:          client.GetBoolValue(inst.Ended),
			Suspended:      client.GetBoolValue(inst.Suspended),
			TenantID:       client.GetStringValue(inst.TenantId),
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

	vars := make([]config.Variable, 0, len(*varsMap))
	for name, varValue := range *varsMap {
		value := ""
		if varValue.Value != nil {
			value = fmt.Sprintf("%v", varValue.Value)
		}
		vars = append(vars, config.Variable{
			Name:  name,
			Value: value,
			Type:  client.GetStringValue(varValue.Type),
		})
	}

	return vars, nil
}

// TerminateInstance terminates a process instance using the generated client.
func (c *Client) TerminateInstance(instanceID string) error {
	_, err := c.operatonAPI.ProcessInstanceAPI.DeleteProcessInstance(c.authContext, instanceID).Execute()
	if err != nil {
		return fmt.Errorf("failed to terminate instance %s: %w", instanceID, err)
	}

	return nil
}
