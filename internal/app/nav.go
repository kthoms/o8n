package app

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/kthoms/o6n/internal/config"
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
		// normalize title (remove editable marker)
		title := col.Title
		if strings.HasSuffix(title, " ✎") {
			title = strings.TrimSuffix(title, " ✎")
		}
		if strings.EqualFold(title, colName) && i < len(row) {
			return stripFocusIndicatorPrefix(row[i])
		}
	}
	// check rowData for hidden columns
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.rowData) {
		if v, ok := m.rowData[cursor][colName]; ok {
			if v == nil {
				return ""
			}
			if s, ok := v.(string); ok {
				return s
			}
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
					if act.Type == "navigate" {
						// Build navigate action: resolves ID from rowData/visible cell and triggers drilldown
						colName := act.Column
						if colName == "" {
							colName = "id"
						}
						items = append(items, actionItem{
							key:        act.Key,
							label:      act.Label + " →",
							isNavigate: true,
							cmd: func(m *model) tea.Cmd {
								cursor := m.table.Cursor()
								val := ""
								if cursor >= 0 && cursor < len(m.rowData) {
									if v, ok := m.rowData[cursor][colName]; ok && v != nil {
										val = fmt.Sprintf("%v", v)
									}
								}
								if val == "" {
									// fallback to visible cell
									def := m.findTableDef(m.currentTableKey())
									visIdx := -1
									if def != nil {
										visIdx = m.visibleColumnIndex(def, colName)
									}
									row := m.table.SelectedRow()
									if visIdx >= 0 && visIdx < len(row) {
										val = stripFocusIndicatorPrefix(fmt.Sprintf("%v", row[visIdx]))
									} else if len(row) > 0 {
										val = stripFocusIndicatorPrefix(fmt.Sprintf("%v", row[0]))
									}
								}
								if val == "" {
									return nil
								}
								d := &config.DrillDownDef{
									Target: act.Target,
									Param:  act.Param,
									Column: colName,
									Label:  act.Label,
								}
								newM, cmd := m.executeDrilldown(d)
								*m = newM
								return cmd
							},
						})
					} else {
						// HTTP mutation action (existing logic)
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
									m.confirmFocusedBtn = 1 // default to Cancel (safe)
									return nil
								}
								return tea.Batch(m.executeActionCmd(act, resolvedPath), flashOnCmd())
							},
						})
					}
				}
				break
			}
		}
	}

	// Always add "View as JSON" and "Copy as JSON" as the last two actions
	items = append(items, actionItem{key: "J", label: "View as JSON", cmd: func(m *model) tea.Cmd {
		row := m.table.SelectedRow()
		if len(row) == 0 {
			return nil
		}
		m.detailContent = m.buildDetailContent(row)
		m.detailScroll = 0
		m.activeModal = ModalDetailView
		return nil
	}})
	items = append(items, actionItem{key: "ctrl+j", label: "Copy as JSON", cmd: func(m *model) tea.Cmd {
		row := m.table.SelectedRow()
		if len(row) == 0 {
			return nil
		}
		content := m.buildDetailContent(row)
		_ = clipboard.WriteAll(content)
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
	var matchCount int
	if m.popup.mode == popupModeSearch {
		matchCount = len(m.table.Rows())
	} else {
		for _, rc := range m.rootContexts {
			if m.popup.input == "" || strings.HasPrefix(rc, m.popup.input) {
				matchCount++
			}
		}
	}
	const maxShow = 8
	shown := min(matchCount, maxShow)
	// 2 border lines + 1 input line + 1 hint line + shown match lines
	return 2 + 2 + shown
}

// computePaneHeight recalculates the table pane height based on current overlay state.
func (m *model) computePaneHeight() int {
	h := m.lastHeight - 2 - 2 - 1 // header(2) - footer(2) - safe(1)
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
	if idx == 0 || idx >= len(m.navigationStack) {
		// Full reset for root navigation or when stack doesn't have the target entry.
		m.prepareStateTransition(TransitionFull)
	} else {
		// Truncate stack so the target state sits at the top, then pop and restore it.
		// stack[idx] holds the state captured when drilling into breadcrumb level idx+1.
		m.navigationStack = m.navigationStack[:idx+1]
		m.prepareStateTransition(TransitionPop)
	}
	// Set active navigation context from the breadcrumb target.
	m.breadcrumb = append([]string{}, m.breadcrumb[:idx+1]...)
	last := m.breadcrumb[len(m.breadcrumb)-1]
	m.currentRoot = last
	m.viewMode = last
	m.contentHeader = last
	return tea.Batch(m.fetchForRoot(last), flashOnCmd())
}

// executeDrilldown performs the full navigation-stack push and resource fetch
// for the given drilldown definition, using the current table cursor as context.
func (m model) executeDrilldown(d *config.DrillDownDef) (model, tea.Cmd) {
	// Push current viewState snapshot and clear non-stack state for the child view.
	m.prepareStateTransition(TransitionDrillDown)

	// Resolve drilldown value: prefer rowData (includes hidden columns like id),
	// fall back to visible cell with focus-indicator prefix stripped
	colName := d.Column
	if colName == "" {
		colName = "id"
	}
	val := ""
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.rowData) {
		if v, ok := m.rowData[cursor][colName]; ok && v != nil {
			val = fmt.Sprintf("%v", v)
		}
	}
	if val == "" {
		visIdx := m.visibleColumnIndex(m.findTableDef(m.currentTableKey()), colName)
		row := m.table.SelectedRow()
		if visIdx >= 0 && visIdx < len(row) {
			val = stripFocusIndicatorPrefix(fmt.Sprintf("%v", row[visIdx]))
		} else if len(row) > 0 {
			val = stripFocusIndicatorPrefix(fmt.Sprintf("%v", row[0]))
		}
		log.Printf("drilldown: column %q not in rowData at cursor %d, used visible cell: %q", colName, cursor, val)
	}

	// Preserve context state needed by edit/save flows
	switch d.Target {
	case "process-instance":
		m.selectedDefinitionKey = val
	case "process-variables":
		m.selectedInstanceID = val
	}

	// breadcrumb label: use configured label or target name
	label := d.Label
	if label == "" {
		label = d.Target
	}

	m.currentRoot = d.Target
	m.viewMode = d.Target
	m.genericParams = map[string]string{d.Param: val}
	m.breadcrumb = append(m.breadcrumb, label)

	// Build content header: use title_attribute if configured, else fallback to param value
	titleVal := val
	if d.TitleAttribute != "" && cursor >= 0 && cursor < len(m.rowData) {
		if tv, ok := m.rowData[cursor][d.TitleAttribute]; ok && tv != nil && fmt.Sprintf("%v", tv) != "" {
			titleVal = fmt.Sprintf("%v", tv)
		}
	}
	m.contentHeader = fmt.Sprintf("%s — %s", d.Target, titleVal)
	m.table.SetCursor(0)

	// Pre-set columns for target table to avoid stale columns during load
	colsTarget := m.buildColumnsFor(d.Target, m.paneWidth-4)
	m.table.SetRows([]table.Row{})
	if len(colsTarget) > 0 {
		m.table.SetColumns(colsTarget)
		m.table.SetRows(normalizeRows(nil, len(colsTarget)))
	}

	return m, tea.Batch(m.fetchGenericCmd(d.Target), flashOnCmd(), m.saveStateCmd())
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

// filteredFirstRunContexts returns m.rootContexts filtered by m.firstRunInput.
// Returns all contexts when input is empty; filters by substring match otherwise.
func (m *model) filteredFirstRunContexts() []string {
	if m.firstRunInput == "" {
		return m.rootContexts
	}
	var out []string
	for _, rc := range m.rootContexts {
		if strings.Contains(rc, m.firstRunInput) {
			out = append(out, rc)
		}
	}
	return out
}

// popupItems returns the list of items for the current popup mode.
// For context mode: rootContexts filtered by input prefix.
// For skin mode: all skin names.
// For search mode: first-column values of current table rows.
func (m *model) popupItems() []string {
	if m.activeModal == ModalContextSwitcher {
		var out []string
		for _, rc := range m.rootContexts {
			if m.popup.input == "" || strings.HasPrefix(rc, m.popup.input) {
				out = append(out, rc)
			}
		}
		return out
	}

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
	default:
		return nil
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
