package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	operaton "github.com/kthoms/o8n/internal/operaton"
)

func TestFetchInstancesWithParams_Fallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/process-instance" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("extraParam") != "123" || q.Get("foo") != "bar" {
			t.Fatalf("unexpected query: %v", q)
		}
		resp := []map[string]interface{}{
			{"id": "1", "name": "one"},
			{"id": "2", "name": "two"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := operaton.NewConfiguration()
	cfg.Servers = operaton.ServerConfigurations{{URL: srv.URL}}
	c, err := New(cfg, false)
	if err != nil {
		t.Fatalf("New client: %v", err)
	}

	ctx := context.Background()
	params := map[string]string{"extraParam": "123", "foo": "bar"}
	out, err := c.FetchInstancesWithParams(ctx, params)
	if err != nil {
		t.Fatalf("FetchInstancesWithParams error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if out[0]["id"] != "1" {
		t.Fatalf("unexpected first id: %v", out[0]["id"])
	}
}

func TestFetchInstancesWithParams_GeneratedClient(t *testing.T) {
	// Server will validate that the generated-client Execute path is used
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/process-instance" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("processDefinitionKey") != "fromGenerated" {
			t.Fatalf("expected processDefinitionKey=fromGenerated, got %v", r.URL.Query())
		}
		resp := []map[string]interface{}{{"id": "g1"}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := operaton.NewConfiguration()
	cfg.Servers = operaton.ServerConfigurations{{URL: srv.URL}}
	// Ensure the generated client's HTTP client points to the test server
	cfg.HTTPClient = srv.Client()

	c, err := New(cfg, false)
	if err != nil {
		t.Fatalf("New client: %v", err)
	}

	ctx := context.Background()
	params := map[string]string{"processDefinitionKey": "fromGenerated"}
	out, err := c.FetchInstancesWithParams(ctx, params)
	if err != nil {
		t.Fatalf("FetchInstancesWithParams error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0]["id"] != "g1" {
		t.Fatalf("unexpected id: %v", out[0]["id"])
	}
}
