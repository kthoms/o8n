package main

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kthoms/o8n/internal/client"
	"github.com/kthoms/o8n/internal/config"
)

func TestNewClientSetsTimeout(t *testing.T) {
	env := config.Environment{URL: "http://example.com"}
	c := client.NewClient(env)

	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestFetchProcessDefinitions(t *testing.T) {
	username := "user"
	password := "pass"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/process-definition" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("expected Accept header application/json, got %s", got)
		}
		if auth := r.Header.Get("Authorization"); auth != "Basic "+basic(username, password) {
			t.Fatalf("missing or invalid basic auth header: %s", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"1","key":"proc1","name":"Process One"}]`))
	}))
	defer server.Close()

	c := client.NewClient(config.Environment{
		URL:      server.URL,
		Username: username,
		Password: password,
	})

	defs, err := c.FetchProcessDefinitions()
	if err != nil {
		t.Fatalf("FetchProcessDefinitions returned error: %v", err)
	}
	if len(defs) != 1 || defs[0].ID != "1" || defs[0].Key != "proc1" {
		t.Fatalf("unexpected definitions: %+v", defs)
	}
}

func TestFetchInstances(t *testing.T) {
	username := "user"
	password := "pass"
	processKey := "testKey"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/process-instance" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("processDefinitionId"); got != processKey {
			t.Fatalf("expected processDefinitionId=%s, got %s", processKey, got)
		}
		if auth := r.Header.Get("Authorization"); auth != "Basic "+basic(username, password) {
			t.Fatalf("missing or invalid basic auth header: %s", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"inst1","definitionId":"def1"}]`))
	}))
	defer server.Close()

	c := client.NewClient(config.Environment{
		URL:      server.URL,
		Username: username,
		Password: password,
	})

	instances, err := c.FetchInstances(processKey)
	if err != nil {
		t.Fatalf("FetchInstances returned error: %v", err)
	}
	if len(instances) != 1 || instances[0].ID != "inst1" || instances[0].DefinitionID != "def1" {
		t.Fatalf("unexpected instances: %+v", instances)
	}
}

func TestTerminateInstance(t *testing.T) {
	username := "user"
	password := "pass"
	instanceID := "abc-123"

	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		receivedPath = r.URL.Path
		if !strings.Contains(receivedPath, instanceID) {
			t.Fatalf("expected path to contain instance id %s, got %s", instanceID, receivedPath)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("expected Accept header application/json, got %s", got)
		}
		if auth := r.Header.Get("Authorization"); auth != "Basic "+basic(username, password) {
			t.Fatalf("missing or invalid basic auth header: %s", auth)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := client.NewClient(config.Environment{
		URL:      server.URL,
		Username: username,
		Password: password,
	})

	if err := c.TerminateInstance(instanceID); err != nil {
		t.Fatalf("TerminateInstance returned error: %v", err)
	}
	if receivedPath != "/process-instance/"+url.PathEscape(instanceID) {
		t.Fatalf("unexpected path: %s", receivedPath)
	}
}

func TestTerminateInstance_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	c := client.NewClient(config.Environment{URL: server.URL})
	if err := c.TerminateInstance("bad"); err == nil {
		t.Fatalf("expected error for bad status code")
	}
}

func basic(user, pass string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
}
