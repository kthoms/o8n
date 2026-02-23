package app

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/kthoms/o8n/internal/config"
)

// resolveActionID extracts the ID value from the selected row for a given action.
// It uses the action's IDColumn (defaults to "id") and finds the matching table column.
func (m *model) resolveActionID(action config.ActionDef) string {
	row := m.table.SelectedRow()
	if len(row) == 0 {
		return ""
	}
	idCol := action.IDColumn
	if idCol == "" {
		idCol = "id"
	}
	cols := m.table.Columns()
	for i, col := range cols {
		if col.Title == idCol && i < len(row) {
			return stripFocusIndicatorPrefix(row[i])
		}
	}
	// Fallback: use first column
	if len(row) > 0 {
		return stripFocusIndicatorPrefix(row[0])
	}
	return ""
}

// resolveRowValue returns the value in a row for the given column name.
// Falls back to the first column value if the column is not found.
func (m *model) resolveRowValue(row []string, colName string) string {
	if len(row) == 0 {
		return ""
	}
	cols := m.table.Columns()
	for i, col := range cols {
		if col.Title == colName && i < len(row) {
			return stripFocusIndicatorPrefix(row[i])
		}
	}
	// check rowData for hidden columns
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.rowData) {
		if v, ok := m.rowData[cursor][colName]; ok {
			return fmt.Sprintf("%v", v)
		}
	}
	return stripFocusIndicatorPrefix(row[0])
}

func (m *model) buildActionsForRoot() []actionItem {
	root := m.currentRoot
	if len(m.breadcrumb) > 0 {
		root = m.breadcrumb[len(m.breadcrumb)-1]
	}

	var items []actionItem

	// Find table definition for current root and load config-driven actions
	if m.config != nil {
		tableKey := m.currentTableKey()
		for _, td := range m.config.Tables {
			if td.Name == tableKey || td.Name == root {
				for _, action := range td.Actions {
					act := action // capture loop variable
					items = append(items, actionItem{
						key:   act.Key,
						label: act.Label,
						cmd: func(m *model) tea.Cmd {
							id := m.resolveActionID(act)
							if id == "" {
								return nil
							}
							resolvedPath := strings.Replace(act.Path, "{id}", id, 1)
							if act.Confirm {
								m.pendingAction = &act
								m.pendingActionID = id
								m.pendingActionPath = resolvedPath
								m.activeModal = ModalConfirmDelete
								return nil
							}
							return tea.Batch(m.executeActionCmd(act, resolvedPath), flashOnCmd())
						},
					})
				}
				break
			}
		}
	}

	// Always add "View as JSON" as the last action
	items = append(items, actionItem{key: "y", label: "View as JSON", cmd: func(m *model) tea.Cmd {
		row := m.table.SelectedRow()
		if len(row) == 0 {
			return nil
		}
		m.detailContent = m.buildDetailContent(row)
		m.detailScroll = 0
		m.activeModal = ModalDetailView
		return nil
	}})

	return items
}

// buildDetailContent builds a JSON representation of the selected row.
func (m *model) buildDetailContent(row table.Row) string {
	// Prefer full raw API object when available
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.rowData) {
		b, err := json.MarshalIndent(m.rowData[cursor], "", "  ")
		if err == nil {
			return string(b)
		}
	}
	// Fallback: build from visible display columns
	cols := m.table.Columns()
	data := make(map[string]string)
	for i, col := range cols {
		val := ""
		if i < len(row) {
			val = ansi.Strip(row[i])
		}
		data[col.Title] = val
	}

	// Pretty-print as JSON
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return string(b)
}

// switchToEnvironment switches to the named environment (extracted from cycling logic).
func (m *model) switchToEnvironment(name string) {
	m.currentEnv = name
	m.applyStyle()
}

func (m *model) resetViews() {
	m.list.SetItems([]list.Item{})
	m.table.SetRows([]table.Row{})
	m.selectedInstanceID = ""
}

func (m *model) findTableDef(name string) *config.TableDef {
	// Flexible lookup: accept exact names and common singular/plural variants so
	// config table names can be either `process-definition` or `process-definitions`.
	if m.config == nil {
		return nil
	}
	// exact match first
	for i := range m.config.Tables {
		if m.config.Tables[i].Name == name {
			return &m.config.Tables[i]
		}
	}

	// try simple singular/plural variants and common suffix swaps
	variants := []string{}
	if strings.HasSuffix(name, "s") {
		variants = append(variants, strings.TrimSuffix(name, "s"))
	} else {
		variants = append(variants, name+"s")
	}
	if strings.HasSuffix(name, "-definitions") {
		variants = append(variants, strings.TrimSuffix(name, "s"))
	} else if strings.HasSuffix(name, "-definition") {
		variants = append(variants, name+"s")
	}
	if strings.HasSuffix(name, "-instances") {
		variants = append(variants, strings.TrimSuffix(name, "s"))
	} else if strings.HasSuffix(name, "-instance") {
		variants = append(variants, name+"s")
	}

	for _, v := range variants {
		for i := range m.config.Tables {
			if m.config.Tables[i].Name == v {
				return &m.config.Tables[i]
			}
		}
	}

	// last resort: match by prefix (useful for small naming differences)
	base := strings.TrimSuffix(name, "s")
	for i := range m.config.Tables {
		if strings.HasPrefix(m.config.Tables[i].Name, base) {
			return &m.config.Tables[i]
		}
	}

	return nil
}

// visibleColumnIndex returns the 0-based index of a visible column (by name) in the
// TableDef. Returns -1 if not found.
func (m *model) visibleColumnIndex(def *config.TableDef, column string) int {
	if def == nil || column == "" {
		return -1
	}
	idx := 0
	for _, c := range def.Columns {
		if !c.IsVisible() {
			continue
		}
		if c.Name == column || strings.EqualFold(c.Name, column) {
			return idx
		}
		idx++
	}
	// fallback: if looking for `id`, try to find any column named id
	if strings.EqualFold(column, "id") {
		idx = 0
		for _, c := range def.Columns {
			if !c.IsVisible() {
				continue
			}
			if strings.EqualFold(c.Name, "id") {
				return idx
			}
		}
	}
	return -1
}

func (m *model) currentTableKey() string {
	if len(m.breadcrumb) > 0 {
		last := m.breadcrumb[len(m.breadcrumb)-1]
		return last
	}
	return m.currentRoot
}

// contextPopupHeight returns the rendered line-height of the context switcher popup,
// or 0 when the popup is hidden. Used to shrink the content pane accordingly.
func (m *model) contextPopupHeight() int {
	if m.popup.mode == popupModeNone {
		return 0
	}
	matchCount := 0
	for _, rc := range m.rootContexts {
		if m.popup.input == "" || strings.HasPrefix(rc, m.popup.input) {
			matchCount++
		}
	}
	const maxShow = 8
	shown := matchCount
	if shown > maxShow {
		shown = maxShow + 1 // one extra for "… N more" line
	}
	// 2 border lines + 1 input line + 1 hint line + shown match lines
	return 2 + 2 + shown
}

// computePaneHeight recalculates the table pane height based on current overlay state.
func (m *model) computePaneHeight() int {
	h := m.lastHeight - 3 - 2 - 1 // header(3) - footer(2) - safe(1)
	h -= m.contextPopupHeight()
	if m.searchMode {
		h -= 1
	}
	if h < 3 {
		h = 3
	}
	return h
}

// getPageSize returns the preferred number of rows to load per page
func (m *model) getPageSize() int {
	if m.table.Height() > 0 {
		return m.table.Height()
	}
	if m.paneHeight > 0 {
		n := m.paneHeight - 2
		if n <= 0 {
			return 1
		}
		return n
	}
	return 10
}

// navigateToBreadcrumb moves the UI state to the breadcrumb level at idx (0-based).
func (m *model) navigateToBreadcrumb(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.breadcrumb) {
		m.footerError = "Invalid breadcrumb index"
		return nil
	}
	// truncate breadcrumb
	m.breadcrumb = append([]string{}, m.breadcrumb[:idx+1]...)
	last := m.breadcrumb[len(m.breadcrumb)-1]
	m.currentRoot = last
	m.viewMode = last
	m.contentHeader = last
	// Clear drilldown-specific state when navigating up
	if idx == 0 {
		m.selectedDefinitionKey = ""
		m.selectedInstanceID = ""
		m.genericParams = nil
	}
	return tea.Batch(m.fetchForRoot(last), flashOnCmd())
}

// currentNavState returns the current navigation position as a serialisable NavState.
func (m *model) currentNavState() config.NavState {
	return config.NavState{
		Root:                  m.currentRoot,
		Breadcrumb:            append([]string{}, m.breadcrumb...),
		SelectedDefinitionKey: m.selectedDefinitionKey,
		SelectedInstanceID:    m.selectedInstanceID,
		GenericParams:         m.genericParams,
	}
}

// currentAppState builds an AppState snapshot from the running model.
func (m *model) currentAppState() *config.AppState {
	return &config.AppState{
		ActiveEnv:   m.currentEnv,
		Skin:        m.activeSkin,
		ShowLatency: m.showLatency,
		Navigation:  m.currentNavState(),
	}
}

// saveStateCmd returns a tea.Cmd that persists the current AppState to disk (best-effort).
func (m *model) saveStateCmd() tea.Cmd {
	state := m.currentAppState()
	path := m.statePath
	return func() tea.Msg {
		_ = config.SaveAppState(path, state)
		return nil
	}
}

// restoreNavState applies a persisted NavState so the app opens at the last view.
func (m *model) restoreNavState(nav config.NavState) {
	if nav.Root == "" {
		return
	}
	m.currentRoot = nav.Root
	m.breadcrumb = append([]string{}, nav.Breadcrumb...)
	m.selectedDefinitionKey = nav.SelectedDefinitionKey
	m.selectedInstanceID = nav.SelectedInstanceID
	if nav.GenericParams != nil {
		m.genericParams = nav.GenericParams
	}
	// viewMode is set by the first genericLoadedMsg received after Init.
}

// popupItems returns the list of items for the current popup mode.
// For context mode: rootContexts filtered by input prefix.
// For skin mode: all skin names.
// For search mode: first-column values of current table rows.
func (m *model) popupItems() []string {
switch m.popup.mode {
case popupModeSkin:
return m.skinPopupItems()
case popupModeSearch:
var out []string
for _, row := range m.table.Rows() {
if len(row) > 0 {
out = append(out, fmt.Sprintf("%v", row[0]))
}
}
return out
default: // popupModeContext
var out []string
for _, rc := range m.rootContexts {
if m.popup.input == "" || strings.HasPrefix(rc, m.popup.input) {
out = append(out, rc)
}
}
return out
}
}

// applySearchFromPopup filters the table rows using popup.input as the search term.
func (m *model) applySearchFromPopup() {
m.searchTerm = m.popup.input
if m.popup.input == "" {
m.table.SetRows(m.originalRows)
} else {
filtered := filterRows(m.originalRows, m.popup.input)
m.table.SetRows(filtered)
}
}
