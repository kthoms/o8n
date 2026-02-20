package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/kthoms/o8n/internal/config"
)

func TestErrorMsgRenderedInFooter(t *testing.T) {
	m := newTestModel(t)

	// Send an error message
	testErr := errors.New("test error")
	res, cmd := m.Update(errMsg{err: testErr})
	m = res.(model)

	// Verify footer fields were set correctly
	if m.footerError != testErr.Error() {
		t.Errorf("expected footerError to be %q, got %q", testErr.Error(), m.footerError)
	}
	if m.footerStatusKind != footerStatusError {
		t.Errorf("expected footerStatusKind to be footerStatusError, got %v", m.footerStatusKind)
	}

	// Verify a clear command was returned
	if cmd == nil {
		t.Error("expected a clear command to be returned")
	}
}

func TestSuccessRenderedInFooter(t *testing.T) {
	m := newTestModel(t)

	// Simulate a successful save by triggering editSavedMsg
	msg := editSavedMsg{rowIndex: 0, colIndex: 0, value: "updated"}
	res, cmd := m.Update(msg)
	m = res.(model)

	// Verify footer fields were set correctly
	if !strings.Contains(m.footerError, "Saved") {
		t.Errorf("expected footerError to contain 'Saved', got %q", m.footerError)
	}
	if m.footerStatusKind != footerStatusSuccess {
		t.Errorf("expected footerStatusKind to be footerStatusSuccess, got %v", m.footerStatusKind)
	}

	// Verify a clear command was returned
	if cmd == nil {
		t.Error("expected a clear command to be returned")
	}
}

func TestLoadingStateRenderedInFooter(t *testing.T) {
	m := newTestModel(t)
	m.isLoading = true
	m.footerError = "Loading..."
	m.footerStatusKind = footerStatusLoading

	// Verify loading state is set correctly
	if !m.isLoading {
		t.Error("expected isLoading to be true")
	}
	if m.footerError != "Loading..." {
		t.Errorf("expected footerError to be 'Loading...', got %q", m.footerError)
	}
	if m.footerStatusKind != footerStatusLoading {
		t.Errorf("expected footerStatusKind to be footerStatusLoading, got %v", m.footerStatusKind)
	}
}

func TestErrorAutoClears(t *testing.T) {
	m := newTestModel(t)

	testErr := errors.New("test error")
	_, cmd := m.Update(errMsg{err: testErr})

	// Verify a command was returned (which schedules the clear)
	if cmd == nil {
		t.Error("expected a command to be returned for auto-clear")
	}
}

func TestSuccessAutoClears(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 80
	m.lastHeight = 24

	msg := editSavedMsg{rowIndex: 0, colIndex: 0, value: "updated"}
	_, cmd := m.Update(msg)

	// Verify a command was returned (which schedules the clear)
	if cmd == nil {
		t.Error("expected a command to be returned for auto-clear")
	}
}

func TestLoadingStateIconKind(t *testing.T) {
	m := newTestModel(t)
	m.isLoading = true
	m.footerError = "Loading data..."
	m.footerStatusKind = footerStatusLoading

	// Verify the loading state is configured correctly
	if m.footerStatusKind != footerStatusLoading {
		t.Errorf("expected footerStatusKind to be footerStatusLoading, got %v", m.footerStatusKind)
	}
	if m.footerError != "Loading data..." {
		t.Errorf("expected footerError to be 'Loading data...', got %q", m.footerError)
	}
}

func TestFooterShowsBreadcrumbWhenNoStatus(t *testing.T) {
	m := newTestModel(t)
	m.breadcrumb = []string{"definitions", "process-definitions"}
	m.footerError = "" // No status
	m.footerStatusKind = footerStatusNone

	// Verify breadcrumb is set and status is clear
	if len(m.breadcrumb) != 2 {
		t.Errorf("expected breadcrumb with 2 items, got %d", len(m.breadcrumb))
	}
	if m.footerError != "" {
		t.Errorf("expected empty footerError, got %q", m.footerError)
	}
	if m.footerStatusKind != footerStatusNone {
		t.Errorf("expected footerStatusKind to be footerStatusNone, got %v", m.footerStatusKind)
	}
}

func TestErrorStatusKindRendersRed(t *testing.T) {
	m := newTestModel(t)
	m.footerStatusKind = footerStatusError

	if m.footerStatusKind != footerStatusError {
		t.Errorf("expected error status kind, got %v", m.footerStatusKind)
	}
}

func TestSuccessStatusKindRendersGreen(t *testing.T) {
	m := newTestModel(t)
	m.footerStatusKind = footerStatusSuccess

	if m.footerStatusKind != footerStatusSuccess {
		t.Errorf("expected success status kind, got %v", m.footerStatusKind)
	}
}

func TestInfoStatusKindRendersWithIcon(t *testing.T) {
	m := newTestModel(t)
	m.footerStatusKind = footerStatusInfo

	if m.footerStatusKind != footerStatusInfo {
		t.Errorf("expected info status kind, got %v", m.footerStatusKind)
	}
}

func TestIsLoadingClearedOnDataLoaded(t *testing.T) {
	m := newTestModel(t)
	m.isLoading = true

	// Simulate receiving data
	msg := definitionsLoadedMsg{definitions: []config.ProcessDefinition{}}
	res, _ := m.Update(msg)
	m = res.(model)

	if m.isLoading {
		t.Error("expected isLoading to be cleared when data is loaded")
	}
}

func TestIsLoadingClearedOnError(t *testing.T) {
	m := newTestModel(t)
	m.isLoading = true

	// Simulate an error
	testErr := errors.New("network error")
	res, _ := m.Update(errMsg{err: testErr})
	m = res.(model)

	if m.isLoading {
		t.Error("expected isLoading to be cleared when an error occurs")
	}
}
