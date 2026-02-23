package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/validation"
)

func (m *model) editableColumnsFor(tableKey string) []editableColumn {
	def := m.findTableDef(tableKey)
	if def == nil {
		return nil
	}
	cols := []editableColumn{}
	idx := 0
	for _, c := range def.Columns {
		if !c.IsVisible() {
			continue
		}
		if c.Editable {
			cols = append(cols, editableColumn{index: idx, def: c})
		}
		idx++
	}
	return cols
}

func (m *model) hasEditableColumns() bool {
	return len(m.editableColumnsFor(m.currentTableKey())) > 0
}

func inputTypeFromVariableType(variableType string) string {
	lower := strings.ToLower(variableType)
	if strings.Contains(lower, "bool") {
		return "bool"
	}
	if strings.Contains(lower, "int") || strings.Contains(lower, "long") {
		return "int"
	}
	if strings.Contains(lower, "double") || strings.Contains(lower, "float") || strings.Contains(lower, "number") {
		return "number"
	}
	return "text"
}

func typeNameForInputType(inputType string, variableType string) string {
	if variableType != "" && inputType == inputTypeFromVariableType(variableType) {
		return variableType
	}
	switch inputType {
	case "bool":
		return "Boolean"
	case "int":
		return "Integer"
	case "number":
		return "Double"
	default:
		return "String"
	}
}

func isVariableTable(tableKey string) bool {
	switch tableKey {
	case "process-variables", "variables", "variable-instance", "variable-instances":
		return true
	}
	return false
}

func (m *model) variableTypeForRow(tableKey string, row table.Row) string {
	if !isVariableTable(tableKey) {
		return ""
	}
	def := m.findTableDef(tableKey)
	if def == nil {
		return ""
	}
	nameIdx := m.visibleColumnIndex(def, "name")
	if nameIdx < 0 || nameIdx >= len(row) {
		return ""
	}
	name := fmt.Sprintf("%v", row[nameIdx])
	if v, ok := m.variablesByName[name]; ok {
		return v.Type
	}
	return ""
}

func (m *model) resolveEditTypes(col config.ColumnDef, tableKey string, row table.Row) (string, string) {
	inputType := strings.TrimSpace(strings.ToLower(col.InputType))
	variableType := m.variableTypeForRow(tableKey, row)
	if inputType == "" {
		inputType = "text"
	}
	if inputType == "auto" {
		inputType = inputTypeFromVariableType(variableType)
	}
	return inputType, typeNameForInputType(inputType, variableType)
}

func parseInputValue(input string, inputType string) (interface{}, error) {
	// delegate to validation package which centralizes parsing/validation rules
	return validation.ValidateAndParse(input, inputType)
}

func (m *model) currentEditRow() table.Row {
	rows := m.table.Rows()
	if m.editRowIndex < 0 || m.editRowIndex >= len(rows) {
		return nil
	}
	return rows[m.editRowIndex]
}

func (m *model) currentEditColumn() *editableColumn {
	if m.editColumnPos < 0 || m.editColumnPos >= len(m.editColumns) {
		return nil
	}
	return &m.editColumns[m.editColumnPos]
}

func (m *model) setEditColumn(pos int) {
	if len(m.editColumns) == 0 {
		return
	}
	if pos < 0 {
		pos = len(m.editColumns) - 1
	}
	if pos >= len(m.editColumns) {
		pos = 0
	}
	m.editColumnPos = pos
	m.editError = ""

	row := m.currentEditRow()
	col := m.currentEditColumn()
	value := ""
	if row != nil && col != nil && col.index < len(row) {
		value = fmt.Sprintf("%v", row[col.index])
	}
	m.editInput.SetValue(value)
	m.editInput.CursorEnd()
}

func (m *model) variableNameForRow(tableKey string, row table.Row) string {
	def := m.findTableDef(tableKey)
	if def == nil {
		return ""
	}
	nameIdx := m.visibleColumnIndex(def, "name")
	if nameIdx < 0 || nameIdx >= len(row) {
		return ""
	}
	return fmt.Sprintf("%v", row[nameIdx])
}

// startEdit opens the edit modal and returns an error if not possible.
// Returns empty string on success.
func (m *model) startEdit(tableKey string) string {
	cols := m.editableColumnsFor(tableKey)
	if len(cols) == 0 {
		return "No editable columns"
	}
	if len(m.table.Rows()) == 0 {
		return "No row selected"
	}
	m.editColumns = cols
	m.editTableKey = tableKey
	m.editRowIndex = m.table.Cursor()
	m.editColumnPos = 0
	m.editError = ""
	m.editFocus = editFocusInput
	m.editInput.Focus()
	m.setEditColumn(0)
	m.activeModal = ModalEdit
	return ""
}
