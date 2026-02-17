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
