package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kthoms/o8n/internal/config"
)

func TestRenderEditModal(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.Environment{
			"local": {UIColor: "#00FF00"},
		},
	}
	m := newModel(cfg)

	// Prepare editable column and a row
	col := config.ColumnDef{Name: "age", InputType: "int", Editable: true, Visible: true}
	m.editColumns = []editableColumn{{index: 1, def: col}}
	m.editColumnPos = 0
	m.editRowIndex = 0
	m.table.SetRows([]table.Row{{"1", "100"}})
	m.editInput.SetValue("abc") // invalid int value should trigger validation
	m.editTableKey = "instances"

	out := m.renderEditModal(80, 24)
	if out == "" {
		t.Fatal("expected non-empty modal output")
	}
	if !strings.Contains(out, "Save") || !strings.Contains(out, "Cancel") {
		t.Fatalf("modal output missing buttons: %q", out)
	}

	// valid input should still render modal
	m.editInput.SetValue("42")
	out2 := m.renderEditModal(80, 24)
	if out2 == "" {
		t.Fatal("expected non-empty modal output for valid input")
	}
}
