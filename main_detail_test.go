package main

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestDetailViewerOpensOnY(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	cols := []table.Column{{Title: "ID", Width: 10}, {Title: "NAME", Width: 20}}
	m.table.SetColumns(cols)
	m.table.SetRows([]table.Row{{"inst-1", "MyProcess"}})
	m.table.SetCursor(0)

	m2, _ := sendKeyString(m, "y")

	if m2.activeModal != ModalDetailView {
		t.Errorf("expected ModalDetailView after 'y', got %v", m2.activeModal)
	}
	if m2.detailContent == "" {
		t.Error("expected non-empty detailContent")
	}
}

func TestDetailViewerScrolling(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalDetailView
	// Create content with many lines
	content := ""
	for i := 0; i < 50; i++ {
		content += "line content\n"
	}
	m.detailContent = content
	m.detailScroll = 0
	m.detailMaxScroll = 40

	// j scrolls down
	m2, _ := sendKeyString(m, "j")
	if m2.detailScroll != 1 {
		t.Errorf("expected detailScroll 1 after j, got %d", m2.detailScroll)
	}

	// k scrolls up
	m3, _ := sendKeyString(m2, "k")
	if m3.detailScroll != 0 {
		t.Errorf("expected detailScroll 0 after k, got %d", m3.detailScroll)
	}
}

func TestDetailViewerEscCloses(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalDetailView
	m.detailContent = "some content"

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Errorf("expected ModalNone after Esc, got %v", m2.activeModal)
	}
	if m2.detailContent != "" {
		t.Errorf("expected empty detailContent after close, got %q", m2.detailContent)
	}
}

func TestDetailViewerQCloses(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalDetailView
	m.detailContent = "some content"

	m2, _ := sendKeyString(m, "q")

	if m2.activeModal != ModalNone {
		t.Errorf("expected ModalNone after q, got %v", m2.activeModal)
	}
}

func TestSyntaxHighlightJSON(t *testing.T) {
	// Basic key-value pair
	result := syntaxHighlightJSON(`  "name": "value"`)
	if result == "" {
		t.Error("expected non-empty result from syntaxHighlightJSON")
	}

	// Number value
	result2 := syntaxHighlightJSON(`  "count": 42`)
	if result2 == "" {
		t.Error("expected non-empty result for number value")
	}

	// Boolean value
	result3 := syntaxHighlightJSON(`  "active": true`)
	if result3 == "" {
		t.Error("expected non-empty result for boolean value")
	}

	// Non-JSON line should pass through
	result4 := syntaxHighlightJSON("{")
	if result4 != "{" {
		t.Errorf("expected '{' unchanged, got %q", result4)
	}
}

func TestDetailViewerYKeyCloses(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalDetailView
	m.detailContent = "content"

	m2, _ := sendKeyString(m, "y")
	if m2.activeModal != ModalNone {
		t.Errorf("expected ModalNone after y in detail view, got %v", m2.activeModal)
	}
}

func TestDetailViewerNoRowNoOpen(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.table.SetRows([]table.Row{})

	m2, _ := sendKeyString(m, "y")

	if m2.activeModal == ModalDetailView {
		t.Error("expected ModalDetailView NOT to open with no rows")
	}
}
