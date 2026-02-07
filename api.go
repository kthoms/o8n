package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents an API client bound to a specific environment.
type Client struct {
	env        Environment
	httpClient *http.Client
}

// NewClient creates a Client with a default timeout.
func NewClient(env Environment) *Client {
	return &Client{
		env:        env,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) http() *http.Client {
	return c.httpClient
}

func (c *Client) buildURL(path string, query url.Values) string {
	base := strings.TrimRight(c.env.URL, "/")
	u := base + path
	if query != nil {
		return u + "?" + query.Encode()
	}
	return u
}

// FetchProcessDefinitions retrieves all process definitions using Basic Auth.
func (c *Client) FetchProcessDefinitions() ([]ProcessDefinition, error) {
	req, err := http.NewRequest(http.MethodGet, c.buildURL("/process-definition", nil), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.env.Username, c.env.Password)

	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch process definitions: status %d", resp.StatusCode)
	}

	var defs []ProcessDefinition
	if err := json.NewDecoder(resp.Body).Decode(&defs); err != nil {
		return nil, err
	}

	return defs, nil
}

// FetchInstances retrieves process instances filtered by process key using Basic Auth.
func (c *Client) FetchInstances(processKey string) ([]ProcessInstance, error) {
	query := url.Values{}
	query.Set("processDefinitionKey", processKey)

	req, err := http.NewRequest(http.MethodGet, c.buildURL("/process-instance", query), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.env.Username, c.env.Password)

	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch instances: status %d", resp.StatusCode)
	}

	var instances []ProcessInstance
	if err := json.NewDecoder(resp.Body).Decode(&instances); err != nil {
		return nil, err
	}

	return instances, nil
}

// FetchVariables retrieves variables for a process instance. The Operaton API may
// return either a JSON object map or an array; we handle both formats and return
// a slice of Variable.
func (c *Client) FetchVariables(instanceID string) ([]Variable, error) {
	req, err := http.NewRequest(http.MethodGet, c.buildURL("/process-instance/"+url.PathEscape(instanceID)+"/variables", nil), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.env.Username, c.env.Password)

	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch variables: status %d", resp.StatusCode)
	}

	// Try to decode into a map[string]{value,type}
	var asMap map[string]struct {
		Value interface{} `json:"value"`
		Type  string      `json:"type"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&asMap); err == nil {
		vars := make([]Variable, 0, len(asMap))
		for k, v := range asMap {
			vars = append(vars, Variable{Name: k, Value: fmt.Sprintf("%v", v.Value), Type: v.Type})
		}
		return vars, nil
	}

	// If not a map, try array form
	// Need to reset body - but cannot; instead, re-request
	resp.Body.Close()
	req2, err := http.NewRequest(http.MethodGet, c.buildURL("/process-instance/"+url.PathEscape(instanceID)+"/variables", nil), nil)
	if err != nil {
		return nil, err
	}
	req2.Header.Set("Accept", "application/json")
	req2.SetBasicAuth(c.env.Username, c.env.Password)

	resp2, err := c.http().Do(req2)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	var vars []Variable
	if err := json.NewDecoder(resp2.Body).Decode(&vars); err != nil {
		return nil, err
	}

	return vars, nil
}

// TerminateInstance terminates a process instance using a DELETE request with Basic Auth.
func (c *Client) TerminateInstance(instanceID string) error {
	req, err := http.NewRequest(http.MethodDelete, c.buildURL("/process-instance/"+url.PathEscape(instanceID), nil), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.env.Username, c.env.Password)

	resp, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to terminate instance %s: status %d", instanceID, resp.StatusCode)
	}

	return nil
}
