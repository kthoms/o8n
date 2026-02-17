package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	cfgpkg "github.com/kthoms/o8n/internal/config"

	operaton "github.com/kthoms/o8n/internal/operaton"
)

var (
	dbgOnce   sync.Once
	dbgWriter io.Writer
	dbgErr    error
)

func getDebugWriter() io.Writer {
	dbgOnce.Do(func() {
		if os.Getenv("O8N_DEBUG") == "1" {
			_ = os.MkdirAll("./debug", 0o755)
			fpath := filepath.Join(".", "debug", "access.log")
			f, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				dbgErr = err
				dbgWriter = nil
				return
			}
			dbgWriter = f
		}
	})
	return dbgWriter
}

type Client struct {
	api    *operaton.APIClient
	base   string
	debug  bool
	mu     sync.Mutex
	writer io.Writer
	cfg    *operaton.Configuration
}

// New creates a new Client. If debug is true, logs are written to ./debug/access.log.
func New(cfg *operaton.Configuration, debug bool) (*Client, error) {
	api := operaton.NewAPIClient(cfg)
	c := &Client{api: api, debug: debug}
	c.cfg = cfg
	if debug {
		if err := os.MkdirAll("./debug", 0o755); err != nil {
			return nil, err
		}
		fpath := filepath.Join(".", "debug", "access.log")
		f, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
		c.writer = f
	}
	return c, nil
}

func (c *Client) logf(format string, args ...interface{}) {
	if !c.debug || c.writer == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Fprintf(c.writer, format+"\n", args...)
}

// FetchProcessDefinitions fetches process definitions.
func (c *Client) FetchProcessDefinitions(ctx context.Context) ([]operaton.ProcessDefinitionDto, error) {
	req := c.api.ProcessDefinitionAPI.GetProcessDefinitions(ctx)
	c.logf("API: FetchProcessDefinitions()")
	// Concrete HTTP log
	c.logf("API: GET /process-definition")
	res, r, err := req.Execute()
	_ = r
	if err != nil {
		return nil, err
	}
	return res, nil
}

// FetchInstances fetches process instances using a single param name/value.
func (c *Client) FetchInstances(ctx context.Context, paramName, paramValue string) ([]map[string]interface{}, error) {
	m := map[string]string{paramName: paramValue}
	return c.FetchInstancesWithParams(ctx, m)
}

// FetchInstancesWithParams tries to use the generated client's request setters for params.
// If any param cannot be applied via setters, it falls back to a manual HTTP GET with the full query string.
func (c *Client) FetchInstancesWithParams(ctx context.Context, params map[string]string) ([]map[string]interface{}, error) {
	// First try to use the generated client's request builder setters via reflection.
	// This is a best-effort, safe attempt: if any step fails or not all params can
	// be applied, fall back to a manual HTTP GET.
	if c.api != nil && c.api.ProcessInstanceAPI != nil {
		// Use reflection to call GetProcessInstances(ctx) on the service
		// and then call setter methods for each param when available.
		defer func() {
			_ = recover()
		}()
		svcVal := reflect.ValueOf(c.api.ProcessInstanceAPI)
		getMethod := svcVal.MethodByName("GetProcessInstances")
		if getMethod.IsValid() {
			res := getMethod.Call([]reflect.Value{reflect.ValueOf(ctx)})
			if len(res) >= 1 {
				reqBuilder := res[0]
				applied := 0
				for k, v := range params {
					setterName := toPascal(k)
					meth := reqBuilder.MethodByName(setterName)
					if !meth.IsValid() {
						continue
					}
					// ensure the method accepts exactly one argument
					if meth.Type().NumIn() != 1 {
						continue
					}
					inType := meth.Type().In(0)
					var arg reflect.Value
					switch inType.Kind() {
					case reflect.String:
						arg = reflect.ValueOf(v)
					case reflect.Bool:
						bv, err := strconv.ParseBool(v)
						if err != nil {
							continue
						}
						arg = reflect.ValueOf(bv)
					case reflect.Int32:
						iv, err := strconv.ParseInt(v, 10, 32)
						if err != nil {
							continue
						}
						arg = reflect.ValueOf(int32(iv))
					case reflect.Int, reflect.Int64:
						iv, err := strconv.ParseInt(v, 10, 64)
						if err != nil {
							continue
						}
						// use int64 value
						arg = reflect.ValueOf(iv)
					default:
						// try to pass as string if assignable
						if inType.AssignableTo(reflect.TypeOf("")) {
							arg = reflect.ValueOf(v)
						} else {
							continue
						}
					}
					out := meth.Call([]reflect.Value{arg})
					if len(out) >= 1 {
						reqBuilder = out[0]
					}
					applied++
				}
				if applied == len(params) {
					// All params applied — execute the generated request
					exec := reqBuilder.MethodByName("Execute")
					if exec.IsValid() {
						out := exec.Call(nil)
						if len(out) >= 3 {
							// (result, *http.Response, error)
							last := out[len(out)-1]
							if !last.IsNil() {
								return nil, last.Interface().(error)
							}
							first := out[0]
							b, err := json.Marshal(first.Interface())
							if err != nil {
								return nil, err
							}
							var outv []map[string]interface{}
							if err := json.Unmarshal(b, &outv); err != nil {
								var single map[string]interface{}
								if err2 := json.Unmarshal(b, &single); err2 == nil {
									return []map[string]interface{}{single}, nil
								}
								return nil, err
							}
							// Log the concrete HTTP path we expect (query string)
							q := url.Values{}
							for k, v := range params {
								q.Set(k, v)
							}
							c.logf("API: FetchInstancesWithParams()")
							c.logf("API: GET /process-instance?%s", q.Encode())
							return outv, nil
						}
					}
				}
			}
		}
	}

	// Manual HTTP GET to /process-instance with query string (fallback-only implementation)
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	raw := "/process-instance"
	if enc := q.Encode(); enc != "" {
		raw = raw + "?" + enc
	}
	c.logf("API: FetchInstancesWithParams()")
	c.logf("API: GET %s", raw)

	// Build manual request using API client's configuration to find base path
	urlStr := raw
	if c.cfg != nil {
		if base, err := c.cfg.ServerURLWithContext(ctx, ""); err == nil {
			urlStr = strings.TrimRight(base, "/") + raw
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	// Use default client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out []map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// toPascal converts camelCase or lowerCamel to PascalCase suitable for method names.
func toPascal(s string) string {
	// split on non-letter/digit boundaries and uppercase first letter
	var parts []string
	cur := ""
	for i, r := range s {
		if r == '_' || r == '-' {
			if cur != "" {
				parts = append(parts, cur)
				cur = ""
			}
			continue
		}
		if i == 0 {
			cur += strings.ToUpper(string(r))
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	return strings.Join(parts, "")
}

// SetProcessVariable sets a variable on a process instance.
func (c *Client) SetProcessVariable(ctx context.Context, instanceId, name string, value interface{}) error {
	c.logf("API: SetProcessVariable(%s, %s)", instanceId, name)
	body := map[string]interface{}{"value": value}
	b, _ := json.Marshal(body)
	// We'll perform a manual PUT to set the variable
	var basePath string
	if c.cfg != nil {
		if base, err := c.cfg.ServerURLWithContext(ctx, ""); err == nil {
			basePath = strings.TrimRight(base, "/")
		}
	}
	urlStr := fmt.Sprintf("%s/process-instance/%s/variables/%s", basePath, instanceId, url.PathEscape(name))
	r, err := http.NewRequestWithContext(ctx, http.MethodPut, urlStr, bytes.NewReader(b))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error setting variable: %s", string(data))
	}
	return nil
}

// --- Compatibility wrapper to match the previous package API used by main and tests ---

// Use config package types for compatibility

// CompatClient provides the historic API surface used by the rest of the app/tests.
type CompatClient struct {
	env         cfgpkg.Environment
	httpClient  *http.Client
	operatonAPI *operaton.APIClient
	authContext context.Context
}

func (c *CompatClient) logf(format string, args ...interface{}) {
	w := getDebugWriter()
	if w == nil {
		return
	}
	fmt.Fprintf(w, format+"\n", args...)
}

// NewClient creates a CompatClient from the environment config (keeps previous API).
func NewClient(env cfgpkg.Environment) *CompatClient {
	httpClient := &http.Client{Timeout: 10 * time.Second}

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

	apiClient := operaton.NewAPIClient(cfg)

	auth := operaton.BasicAuth{
		UserName: env.Username,
		Password: env.Password,
	}
	authContext := context.WithValue(context.Background(), operaton.ContextBasicAuth, auth)

	return &CompatClient{
		env:         env,
		httpClient:  httpClient,
		operatonAPI: apiClient,
		authContext: authContext,
	}
}

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

// FetchProcessDefinitions retrieves process definitions.
func (c *CompatClient) FetchProcessDefinitions() ([]cfgpkg.ProcessDefinition, error) {
	c.logf("API: FetchProcessDefinitions()")
	c.logf("API: GET /process-definition")
	defs, _, err := c.operatonAPI.ProcessDefinitionAPI.GetProcessDefinitions(c.authContext).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch process definitions: %w", err)
	}
	out := make([]cfgpkg.ProcessDefinition, len(defs))
	for i, d := range defs {
		out[i] = cfgpkg.ProcessDefinition{
			ID:           getStringValue(d.Id),
			Key:          getStringValue(d.Key),
			Category:     getStringValue(d.Category),
			Description:  getStringValue(d.Description),
			Name:         getStringValue(d.Name),
			Version:      int(getInt32Value(d.Version)),
			Resource:     getStringValue(d.Resource),
			DeploymentID: getStringValue(d.DeploymentId),
			Diagram:      getStringValue(d.Diagram),
			Suspended:    getBoolValue(d.Suspended),
			TenantID:     getStringValue(d.TenantId),
		}
	}
	return out, nil
}

// FetchInstances retrieves process instances; if processKey is non-empty it uses that as processDefinitionKey.
func (c *CompatClient) FetchInstances(paramName, paramValue string) ([]cfgpkg.ProcessInstance, error) {
	c.logf("API: FetchInstances(%s=%s)", paramName, paramValue)
	// Log concrete GET path the compat wrapper will use
	q := url.Values{}
	if paramName != "" && paramValue != "" {
		q.Set(paramName, paramValue)
	}
	raw := "/process-instance"
	if enc := q.Encode(); enc != "" {
		raw = raw + "?" + enc
	}
	c.logf("API: GET %s", raw)

	req := c.operatonAPI.ProcessInstanceAPI.GetProcessInstances(c.authContext)
	if paramName != "" && paramValue != "" {
		// Support common drilldown params: processDefinitionKey and processDefinitionId
		if paramName == "processDefinitionKey" {
			req = req.ProcessDefinitionKey(paramValue)
		} else if paramName == "processDefinitionId" {
			// generated client may expose ProcessDefinitionId setter
			// use it when available
			req = req.ProcessDefinitionId(paramValue)
		}
	}
	instances, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch instances: %w", err)
	}
	out := make([]cfgpkg.ProcessInstance, len(instances))
	for i, inst := range instances {
		out[i] = cfgpkg.ProcessInstance{
			ID:             getStringValue(inst.Id),
			DefinitionID:   getStringValue(inst.DefinitionId),
			BusinessKey:    getStringValue(inst.BusinessKey),
			CaseInstanceID: getStringValue(inst.CaseInstanceId),
			Ended:          getBoolValue(inst.Ended),
			Suspended:      getBoolValue(inst.Suspended),
			TenantID:       getStringValue(inst.TenantId),
		}
	}
	return out, nil
}

// FetchVariables retrieves variables (legacy wrapper) for a process instance.
func (c *CompatClient) FetchVariables(instanceID string) ([]cfgpkg.Variable, error) {
	c.logf("API: FetchVariables(instanceId=%s)", instanceID)
	c.logf("API: GET /process-instance/%s/variables", instanceID)
	varsMap, _, err := c.operatonAPI.ProcessInstanceAPI.GetProcessInstanceVariables(c.authContext, instanceID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch variables: %w", err)
	}
	vars := make([]cfgpkg.Variable, 0, len(*varsMap))
	for name, varValue := range *varsMap {
		value := ""
		if varValue.Value != nil {
			value = fmt.Sprintf("%v", varValue.Value)
		}
		vars = append(vars, cfgpkg.Variable{
			Name:  name,
			Value: value,
			Type:  getStringValue(varValue.Type),
		})
	}
	return vars, nil
}

// TerminateInstance terminates a process instance.
func (c *CompatClient) TerminateInstance(instanceID string) error {
	c.logf("API: TerminateInstance(%s)", instanceID)
	c.logf("API: DELETE /process-instance/%s", instanceID)
	_, err := c.operatonAPI.ProcessInstanceAPI.DeleteProcessInstance(c.authContext, instanceID).Execute()
	if err != nil {
		return fmt.Errorf("failed to terminate instance %s: %w", instanceID, err)
	}
	return nil
}

// SetProcessInstanceVariable sets a process variable on an instance (legacy API used by UI).
func (c *CompatClient) SetProcessInstanceVariable(instanceID, varName string, value interface{}, valueType string) error {
	v := operaton.NewVariableValueDto()
	v.SetValue(value)
	if valueType != "" {
		v.SetType(valueType)
	}
	c.logf("API: SetProcessInstanceVariable(%s, %s)", instanceID, varName)
	c.logf("API: PUT /process-instance/%s/variables/%s", instanceID, url.PathEscape(varName))
	_, err := c.operatonAPI.ProcessInstanceAPI.SetProcessInstanceVariable(c.authContext, instanceID, varName).VariableValueDto(*v).Execute()
	if err != nil {
		return fmt.Errorf("failed to set variable: %w", err)
	}
	return nil
}
