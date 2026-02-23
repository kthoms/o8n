package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kthoms/o8n/internal/client"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/dao"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) nextEnvironment() tea.Cmd {
	if len(m.envNames) == 0 {
		return nil
	}
	idx := 0
	for i, name := range m.envNames {
		if name == m.currentEnv {
			idx = i
			break
		}
	}
	idx = (idx + 1) % len(m.envNames)
	m.currentEnv = m.envNames[idx]
	m.applyStyle()
	// Check health of the newly selected environment
	return m.checkEnvironmentHealthCmd(m.currentEnv)
}

// executeActionCmd creates a command that performs a config-driven REST action.
func (m model) executeActionCmd(action config.ActionDef, resolvedPath string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env, m.debugEnabled)
	label := action.Label
	return func() tea.Msg {
		if err := c.ExecuteAction(action.Method, resolvedPath, action.Body); err != nil {
			return errMsg{err}
		}
		return actionExecutedMsg{label: label}
	}
}

// buildActionsForRoot returns context-specific action items for the current root resource.
// Actions are loaded from the config file's table definitions. A "View as JSON" action
// is always appended as the last item.

// suspendInstanceCmd creates a command to suspend/resume a process instance.
func (m model) suspendInstanceCmd(id string, suspend bool) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env, m.debugEnabled)
	return func() tea.Msg {
		if err := c.SuspendProcessInstance(id, suspend); err != nil {
			return errMsg{err}
		}
		if suspend {
			return suspendedMsg{id: id}
		}
		return resumedMsg{id: id}
	}
}

// setJobRetriesCmd creates a command to set retries on a job.
func (m model) setJobRetriesCmd(id string, retries int) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env, m.debugEnabled)
	return func() tea.Msg {
		if err := c.SetJobRetries(id, retries); err != nil {
			return errMsg{err}
		}
		return retriedMsg{id: id}
	}
}

func (m model) fetchDefinitionsCmd() tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env, m.debugEnabled)
	return func() tea.Msg {
		defs, err := c.FetchProcessDefinitions()
		if err != nil {
			return errMsg{err}
		}
		// Fetch count separately (non-fatal if it fails)
		count := 0
		if countVal, err := c.FetchProcessDefinitionsCount(); err == nil {
			count = countVal
		}
		return definitionsLoadedMsg{definitions: defs, count: count}
	}
}

func (m model) fetchInstancesCmd(paramName, paramValue string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	// Use paged HTTP fetch similar to generic fetch so we can pass firstResult/maxResults
	return func() tea.Msg {
		// If caller asked for a definition id but provided a key, resolve it from cache
		value := paramValue
		if paramName == "processDefinitionId" {
			// try to map key -> id when possible
			for _, d := range m.cachedDefinitions {
				if d.Key == paramValue {
					value = d.ID
					break
				}
			}
		}

		base := strings.TrimRight(env.URL, "/")
		// API uses singular path for instances endpoint
		urlStr := base + "/process-instance"
		// add filter param if provided
		q := ""
		if paramName != "" && value != "" {
			q = fmt.Sprintf("%s=%s", paramName, value)
		}
		offset := 0
		if v, ok := m.pageOffsets[dao.ResourceProcessInstances]; ok {
			offset = v
		}
		limit := m.getPageSize()
		if q != "" {
			urlStr = urlStr + "?" + q + fmt.Sprintf("&firstResult=%d&maxResults=%d", offset, limit)
		} else {
			urlStr = urlStr + fmt.Sprintf("?firstResult=%d&maxResults=%d", offset, limit)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			return errMsg{err}
		}
		req.Header.Set("Accept", "application/json")
		if env.Username != "" {
			req.SetBasicAuth(env.Username, env.Password)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			data, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("failed to fetch instances: %s", string(data))}
		}
		var items []map[string]interface{}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&items); err != nil {
			return errMsg{err}
		}

		// Try to load count using the process-instance table def's count path.
		count := -1
		instanceCountPath := "/process-instance/count"
		if def := m.findTableDef("process-instance"); def != nil && def.CountPath != "" {
			instanceCountPath = def.CountPath
		}
		countURL := base + "/" + strings.TrimLeft(instanceCountPath, "/")
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		req2, err2 := http.NewRequestWithContext(ctx2, http.MethodGet, countURL, nil)
		if err2 == nil {
			req2.Header.Set("Accept", "application/json")
			if env.Username != "" {
				req2.SetBasicAuth(env.Username, env.Password)
			}
			if resp2, err2 := http.DefaultClient.Do(req2); err2 == nil {
				defer resp2.Body.Close()
				if resp2.StatusCode < 400 {
					var cntBody map[string]interface{}
					dec2 := json.NewDecoder(resp2.Body)
					if err3 := dec2.Decode(&cntBody); err3 == nil {
						if v, ok := cntBody["count"]; ok {
							if n, ok := v.(float64); ok {
								count = int(n)
							}
						}
					}
				}
			}
		}

		// Convert items to typed instances
		instances := make([]config.ProcessInstance, 0, len(items))
		for _, it := range items {
			pi := config.ProcessInstance{}
			if v, ok := it["id"]; ok {
				pi.ID = fmt.Sprintf("%v", v)
			}
			if v, ok := it["processDefinitionId"]; ok {
				pi.DefinitionID = fmt.Sprintf("%v", v)
			}
			if v, ok := it["businessKey"]; ok {
				pi.BusinessKey = fmt.Sprintf("%v", v)
			}
			if v, ok := it["startTime"]; ok {
				pi.StartTime = fmt.Sprintf("%v", v)
			}
			instances = append(instances, pi)
		}

		// return typed instances message (tests expect this path)
		// Note: count handling for instances is not propagated in this message.
		return instancesWithCountMsg{instances: instances, count: count}
	}
}

// New: fetch variables for a given instance
func (m model) fetchVariablesCmd(instanceID string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env, m.debugEnabled)
	return func() tea.Msg {
		vars, err := c.FetchVariables(instanceID)
		if err != nil {
			return errMsg{err}
		}
		return variablesLoadedMsg{variables: vars}
	}
}

func (m model) fetchDataCmd() tea.Cmd {
	// keep original behaviour for tests: fetch defs and instances for selected key
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}

	var selectedKey string
	items := m.list.Items()
	if len(items) > 0 {
		index := m.list.Index()
		if index >= 0 && index < len(items) {
			if item, ok := items[index].(processDefinitionItem); ok {
				selectedKey = item.definition.Key
			}
		}
	}

	c := client.NewClient(env, m.debugEnabled)

	return func() tea.Msg {
		defs, err := c.FetchProcessDefinitions()
		if err != nil {
			return errMsg{err}
		}

		var instances []config.ProcessInstance
		if selectedKey != "" {
			instances, err = c.FetchInstances("processDefinitionKey", selectedKey)
			if err != nil {
				return errMsg{err}
			}
		}

		return dataLoadedMsg{definitions: defs, instances: instances}
	}
}

// fetchForRoot returns a command that fetches data for the given root resource.
// All resources are fetched via fetchGenericCmd which reads api_path and count_path from config.
func (m model) fetchForRoot(root string) tea.Cmd {
	if def := m.findTableDef(root); def != nil {
		return m.fetchGenericCmd(root)
	}
	// root has no table def — fall back to process-definition list
	return m.fetchGenericCmd("process-definition")
}

// fetchGenericCmd performs a GET to the environment server for the provided
// collection resource (root) and returns a genericLoadedMsg with the parsed
// JSON array of objects.
func (m model) fetchGenericCmd(root string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	// ensure paging defaults
	if m.pageOffsets == nil {
		m.pageOffsets = make(map[string]int)
	}
	if m.pageTotals == nil {
		m.pageTotals = make(map[string]int)
	}

	// Resolve API path: prefer TableDef.ApiPath, then fall back to /{name}.
	apiPath := "/" + strings.TrimLeft(root, "/")
	countPath := ""
	if def := m.findTableDef(root); def != nil {
		if def.ApiPath != "" {
			apiPath = def.ApiPath
		} else {
			apiPath = "/" + strings.TrimLeft(def.Name, "/")
		}
		if def.CountPath != "" {
			countPath = def.CountPath
		}
	}
	if countPath == "" {
		countPath = strings.TrimRight(apiPath, "/") + "/count"
	}

	// Copy active filter params for thread-safe use inside the goroutine.
	paramsCopy := make(map[string]string, len(m.genericParams))
	for k, v := range m.genericParams {
		paramsCopy[k] = v
	}

	return func() tea.Msg {
		base := strings.TrimRight(env.URL, "/")
		offset := 0
		if v, ok := m.pageOffsets[root]; ok {
			offset = v
		}
		limit := m.getPageSize()
		urlStr := base + "/" + strings.TrimLeft(apiPath, "/")
		// append drilldown filter params, then paging params
		for k, v := range paramsCopy {
			if strings.Contains(urlStr, "?") {
				urlStr = fmt.Sprintf("%s&%s=%s", urlStr, k, v)
			} else {
				urlStr = fmt.Sprintf("%s?%s=%s", urlStr, k, v)
			}
		}
		if limit > 0 {
			if strings.Contains(urlStr, "?") {
				urlStr = fmt.Sprintf("%s&firstResult=%d&maxResults=%d", urlStr, offset, limit)
			} else {
				urlStr = urlStr + fmt.Sprintf("?firstResult=%d&maxResults=%d", offset, limit)
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			return errMsg{err}
		}
		req.Header.Set("Accept", "application/json")
		if env.Username != "" {
			req.SetBasicAuth(env.Username, env.Password)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			data, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("failed to fetch %s: %s", root, string(data))}
		}
		var items []map[string]interface{}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&items); err != nil {
			return errMsg{err}
		}

		// Try to load count using the correct count endpoint for this table.
		count := -1
		countURL := base + "/" + strings.TrimLeft(countPath, "/")
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		req2, err2 := http.NewRequestWithContext(ctx2, http.MethodGet, countURL, nil)
		if err2 == nil {
			req2.Header.Set("Accept", "application/json")
			if env.Username != "" {
				req2.SetBasicAuth(env.Username, env.Password)
			}
			if resp2, err2 := http.DefaultClient.Do(req2); err2 == nil {
				defer resp2.Body.Close()
				if resp2.StatusCode < 400 {
					var cntBody map[string]interface{}
					dec2 := json.NewDecoder(resp2.Body)
					if err3 := dec2.Decode(&cntBody); err3 == nil {
						if v, ok := cntBody["count"]; ok {
							if n, ok := v.(float64); ok {
								count = int(n)
							}
						}
					}
				}
			}
		}

		msg := genericLoadedMsg{root: root, items: items}
		if count >= 0 {
			if msg.items == nil {
				msg.items = []map[string]interface{}{}
			}
			meta := map[string]interface{}{"_meta_count": count}
			msg.items = append([]map[string]interface{}{meta}, msg.items...)
		}
		return msg
	}
}

func (m model) terminateInstanceCmd(id string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env, m.debugEnabled)

	return func() tea.Msg {
		if err := c.TerminateInstance(id); err != nil {
			return errMsg{err}
		}
		return terminatedMsg{id: id}
	}
}

func (m model) setVariableCmd(instanceID, varName string, value interface{}, valueType string, rowIndex, colIndex int, displayValue string, dataKey string) tea.Cmd {
	if instanceID == "" || varName == "" {
		return func() tea.Msg { return errMsg{fmt.Errorf("missing instance or variable name")} }
	}
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env, m.debugEnabled)
	return func() tea.Msg {
		if err := c.SetProcessInstanceVariable(instanceID, varName, value, valueType); err != nil {
			return errMsg{err}
		}
		return editSavedMsg{rowIndex: rowIndex, colIndex: colIndex, value: displayValue, dataKey: dataKey}
	}
}

// executeEditActionCmd saves an edited value using a TableDef.EditAction template.
// Placeholders in Path and BodyTemplate: {id}, {name}, {parentId}, {value}, {type}.
func (m model) executeEditActionCmd(act config.EditActionDef, id, name, parentID, value, typeName string, rowIndex, colIndex int, displayValue, dataKey string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return func() tea.Msg { return errMsg{fmt.Errorf("unknown environment %q", m.currentEnv)} }
	}
	replace := func(s string) string {
		s = strings.ReplaceAll(s, "{id}", id)
		s = strings.ReplaceAll(s, "{name}", name)
		s = strings.ReplaceAll(s, "{parentId}", parentID)
		s = strings.ReplaceAll(s, "{value}", value)
		s = strings.ReplaceAll(s, "{type}", typeName)
		return s
	}
	url := strings.TrimRight(env.URL, "/") + replace(act.Path)
	body := replace(act.BodyTemplate)
	method := act.Method
	if method == "" {
		method = "PUT"
	}
	return func() tea.Msg {
		req, err := http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			return errMsg{err}
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(env.Username, env.Password)
		resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			b, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("edit action %s %s: HTTP %d: %s", method, url, resp.StatusCode, string(b))}
		}
		return editSavedMsg{rowIndex: rowIndex, colIndex: colIndex, value: displayValue, dataKey: dataKey}
	}
}


// helper to create a command that triggers the flash indicator
func flashOnCmd() tea.Cmd {
	return func() tea.Msg { return flashOnMsg{} }
}

// checkEnvironmentHealthCmd performs a simple health check on the given environment
func (m *model) checkEnvironmentHealthCmd(envName string) tea.Cmd {
	return func() tea.Msg {
		if m.config == nil {
			return envStatusMsg{env: envName, status: StatusUnknown}
		}

		env, ok := m.config.Environments[envName]
		if !ok {
			return envStatusMsg{env: envName, status: StatusUnknown}
		}

		// Make a simple HTTP GET request to check connectivity
		httpClient := &http.Client{
			Timeout: 2 * time.Second,
		}

		// Try a simple identity endpoint to check health
		req, err := http.NewRequest("GET", env.URL+"/identity/current", nil)
		if err != nil {
			return envStatusMsg{env: envName, status: StatusUnreachable, err: err}
		}

		// Add basic auth
		req.SetBasicAuth(env.Username, env.Password)

		resp, err := httpClient.Do(req)
		if err != nil {
			return envStatusMsg{env: envName, status: StatusUnreachable, err: err}
		}
		defer resp.Body.Close()

		// Any 2xx or 3xx response = operational (ignore auth errors for now)
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			return envStatusMsg{env: envName, status: StatusOperational}
		}

		// 4xx/5xx = unreachable/unhealthy
		return envStatusMsg{env: envName, status: StatusUnreachable}
	}
}
