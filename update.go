package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/dao"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case splashDoneMsg:
		m.splashActive = false

	case splashFrameMsg:
		// update frame and schedule next frame or end splash
		if msg.frame <= 0 {
			msg.frame = 1
		}
		if msg.frame >= totalSplashFrames {
			m.splashFrame = totalSplashFrames
			m.splashActive = false
		} else {
			m.splashFrame = msg.frame
			// schedule next frame
			next := msg.frame + 1
			return m, tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return splashFrameMsg{frame: next} })
		}

	case tea.KeyMsg:
		s := msg.String()

		// Handle search mode keys first (before any modal/popup checks)
		if m.searchMode {
			switch s {
			case "esc":
				m.searchMode = false
				m.searchInput.Blur()
				m.searchTerm = ""
				if m.originalRows != nil {
					m.table.SetRows(m.originalRows)
				}
				m.originalRows = nil
				m.filteredRows = nil
				if m.footerStatusKind == footerStatusInfo {
					m.footerError = ""
					m.footerStatusKind = footerStatusNone
				}
				m.paneHeight = m.computePaneHeight()
				m.table.SetHeight(m.paneHeight - 1)
				return m, nil
			case "enter":
				m.searchMode = false
				m.searchInput.Blur()
				m.searchTerm = m.searchInput.Value()
				// keep filtered rows as current table rows
				m.originalRows = nil
				m.filteredRows = nil
				m.paneHeight = m.computePaneHeight()
				m.table.SetHeight(m.paneHeight - 1)
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				// recompute filtered rows
				term := m.searchInput.Value()
				m.searchTerm = term
				if term == "" {
					m.table.SetRows(m.originalRows)
					m.filteredRows = nil
				} else {
					m.filteredRows = filterRows(m.originalRows, term)
					m.table.SetRows(m.filteredRows)
					m.table.SetCursor(0)
				}
				return m, cmd
			}
		}

		// Handle sort popup keys
		if m.activeModal == ModalSort {
			switch s {
			case "esc":
				m.activeModal = ModalNone
				return m, nil
			case "up", "k":
				if m.sortColumn >= 0 && m.sortPopupCursor == 0 {
					m.sortPopupCursor = -1 // move to clear item
				} else if m.sortPopupCursor > 0 {
					m.sortPopupCursor--
				}
				return m, nil
			case "down", "j":
				if m.sortPopupCursor == -1 {
					m.sortPopupCursor = 0
				} else {
					cols := m.table.Columns()
					if m.sortPopupCursor < len(cols)-1 {
						m.sortPopupCursor++
					}
				}
				return m, nil
			case "enter":
				if m.sortPopupCursor == -1 {
					// Clear sort: reset and re-fetch
					m.sortColumn = -1
					m.sortAscending = true
					m.applySortIndicatorToColumns()
					m.activeModal = ModalNone
					root := m.currentRoot
					if len(m.breadcrumb) > 0 {
						root = m.breadcrumb[len(m.breadcrumb)-1]
					}
					return m, tea.Batch(m.fetchForRoot(root), flashOnCmd())
				}
				cols := m.table.Columns()
				if m.sortPopupCursor >= 0 && m.sortPopupCursor < len(cols) {
					if m.sortColumn == m.sortPopupCursor {
						m.sortAscending = !m.sortAscending
					} else {
						m.sortColumn = m.sortPopupCursor
						m.sortAscending = true
					}
					rows := m.table.Rows()
					sorted := sortTableRows(rows, m.sortColumn, m.sortAscending)
					m.table.SetRows(sorted)
					m.applySortIndicatorToColumns()
				}
				m.activeModal = ModalNone
				return m, nil
			}
			return m, nil
		}

		// Handle actions menu keys
		if m.showActionsMenu {
			switch s {
			case "esc":
				m.showActionsMenu = false
				return m, nil
			case "up", "k":
				if m.actionsMenuCursor > 0 {
					m.actionsMenuCursor--
				}
				return m, nil
			case "down", "j":
				if m.actionsMenuCursor < len(m.actionsMenuItems)-1 {
					m.actionsMenuCursor++
				}
				return m, nil
			case "enter":
				if m.actionsMenuCursor >= 0 && m.actionsMenuCursor < len(m.actionsMenuItems) {
					item := m.actionsMenuItems[m.actionsMenuCursor]
					m.showActionsMenu = false
					if item.cmd != nil {
						return m, item.cmd(&m)
					}
				}
				m.showActionsMenu = false
				return m, nil
			default:
				// Check shortcut keys
				for _, item := range m.actionsMenuItems {
					if s == item.key {
						m.showActionsMenu = false
						if item.cmd != nil {
							return m, item.cmd(&m)
						}
						return m, nil
					}
				}
				return m, nil
			}
		}

		// Handle detail view keys
		if m.activeModal == ModalDetailView {
			detailLines := strings.Split(m.detailContent, "\n")
			detailViewH := m.lastHeight - 6
			if detailViewH < 3 {
				detailViewH = 3
			}
			maxDetailScroll := len(detailLines) - detailViewH
			if maxDetailScroll < 0 {
				maxDetailScroll = 0
			}
			switch s {
			case "esc", "q", "y":
				m.activeModal = ModalNone
				m.detailContent = ""
				return m, nil
			case "j", "down":
				if m.detailScroll < maxDetailScroll {
					m.detailScroll++
				}
				return m, nil
			case "k", "up":
				if m.detailScroll > 0 {
					m.detailScroll--
				}
				return m, nil
			case "ctrl+d":
				m.detailScroll += 10
				if m.detailScroll > maxDetailScroll {
					m.detailScroll = maxDetailScroll
				}
				return m, nil
			case "ctrl+u":
				m.detailScroll -= 10
				if m.detailScroll < 0 {
					m.detailScroll = 0
				}
				return m, nil
			case "G":
				m.detailScroll = maxDetailScroll
				return m, nil
			case "g":
				if m.pendingG {
					m.pendingG = false
					m.detailScroll = 0
					return m, nil
				}
				m.pendingG = true
				return m, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg { return clearPendingGMsg{} })
			}
			return m, nil
		}

		// Handle environment popup keys
		if m.activeModal == ModalEnvironment {
			switch s {
			case "esc":
				m.activeModal = ModalNone
				return m, nil
			case "up", "k":
				if m.envPopupCursor > 0 {
					m.envPopupCursor--
				}
				return m, nil
			case "down", "j":
				if m.envPopupCursor < len(m.envNames)-1 {
					m.envPopupCursor++
				}
				return m, nil
			case "enter":
				if m.envPopupCursor >= 0 && m.envPopupCursor < len(m.envNames) {
					targetEnv := m.envNames[m.envPopupCursor]
					m.activeModal = ModalNone
					if targetEnv != m.currentEnv {
						m.switchToEnvironment(targetEnv)
						m.resetViews()
						m.isLoading = true
						m.apiCallStarted = time.Now()
						return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), m.checkEnvironmentHealthCmd(targetEnv), spinnerTickCmd())
					}
				}
				return m, nil
			}
			return m, nil
		}

		// Handle modal-specific keys first
		if m.activeModal == ModalHelp {
			// compute help line count for scroll bounds
			helpLines := strings.Split(renderHelpContentForLineCount(m.viewMode, m.currentEnv), "\n")
			visibleH := m.lastHeight - 8
			if visibleH < 5 {
				visibleH = 5
			}
			maxScroll := len(helpLines) - visibleH
			if maxScroll < 0 {
				maxScroll = 0
			}
			switch s {
			case "j", "down", "ctrl+d":
				if m.helpScroll < maxScroll {
					m.helpScroll++
				}
				return m, nil
			case "k", "up", "ctrl+u":
				if m.helpScroll > 0 {
					m.helpScroll--
				}
				return m, nil
			default:
				m.activeModal = ModalNone
				m.helpScroll = 0
				return m, nil
			}
		}

		if m.activeModal == ModalEdit {
			row := m.currentEditRow()
			col := m.currentEditColumn()
			inputType, typeName := "text", "String"
			if row != nil && col != nil {
				inputType, typeName = m.resolveEditTypes(col.def, m.editTableKey, row)
			}
			switch s {
			case "esc":
				m.activeModal = ModalNone
				m.editError = ""
				m.editInput.Blur()
				return m, nil
			case "tab":
				switch m.editFocus {
				case editFocusInput:
					if len(m.editColumns) > 1 && m.editColumnPos < len(m.editColumns)-1 {
						m.setEditColumn(m.editColumnPos + 1) // still on input, next column
					} else {
						m.editFocus = editFocusSave
						m.editInput.Blur()
					}
				case editFocusSave:
					m.editFocus = editFocusCancel
				case editFocusCancel:
					m.editFocus = editFocusInput
					m.setEditColumn(0)
					m.editInput.Focus()
				}
				return m, nil
			case "shift+tab", "backtab":
				switch m.editFocus {
				case editFocusInput:
					if len(m.editColumns) > 1 && m.editColumnPos > 0 {
						m.setEditColumn(m.editColumnPos - 1) // still on input, prev column
					} else {
						m.editFocus = editFocusCancel
						m.editInput.Blur()
					}
				case editFocusSave:
					m.editFocus = editFocusInput
					m.setEditColumn(len(m.editColumns) - 1)
					m.editInput.Focus()
				case editFocusCancel:
					m.editFocus = editFocusSave
				}
				return m, nil
			case " ", "space":
				if inputType == "bool" && m.editFocus == editFocusInput {
					current := strings.TrimSpace(strings.ToLower(m.editInput.Value()))
					if current == "true" {
						m.editInput.SetValue("false")
					} else {
						m.editInput.SetValue("true")
					}
					m.editInput.CursorEnd()
					return m, nil
				}
			case "enter":
				if m.editFocus == editFocusCancel {
					m.activeModal = ModalNone
					m.editError = ""
					m.editInput.Blur()
					return m, nil
				}
				// editFocusInput or editFocusSave → save
				if row == nil || col == nil {
					m.editError = "No selection"
					return m, nil
				}
				if isVariableTable(m.editTableKey) {
					varName := m.variableNameForRow(m.editTableKey, row)
					if varName == "" {
						m.editError = "Variable name not found"
						return m, nil
					}
					parsedValue, err := parseInputValue(m.editInput.Value(), inputType)
					if err != nil {
						m.editError = err.Error()
						return m, nil
					}
					rowIndex := m.editRowIndex
					colIndex := col.index
					displayValue := m.editInput.Value()
					m.activeModal = ModalNone
					m.editError = ""
					m.editInput.Blur()
					return m, tea.Batch(m.setVariableCmd(m.selectedInstanceID, varName, parsedValue, typeName, rowIndex, colIndex, displayValue, varName), flashOnCmd())
				}
				m.editError = "Editing not supported for this table"
				return m, nil
			default:
				var cmd tea.Cmd
				m.editInput, cmd = m.editInput.Update(msg)
				return m, cmd
			}
		}

		if m.activeModal == ModalConfirmDelete {
			// Only <ctrl>+d confirms delete, any other key cancels
			if s == "ctrl+d" {
				// Confirm delete
				m.activeModal = ModalNone
				// Config-driven action confirmation
				if m.pendingAction != nil {
					act := *m.pendingAction
					resolvedPath := m.pendingActionPath
					m.pendingAction = nil
					m.pendingActionID = ""
					m.pendingActionPath = ""
					return m, tea.Batch(m.executeActionCmd(act, resolvedPath), flashOnCmd())
				}
				// Legacy: terminate process instance
				if m.pendingDeleteID != "" {
					return m, tea.Batch(m.terminateInstanceCmd(m.pendingDeleteID), flashOnCmd())
				}
			} else {
				// Cancel
				m.activeModal = ModalNone
				m.pendingDeleteID = ""
				m.pendingDeleteLabel = ""
				m.pendingAction = nil
				m.pendingActionID = ""
				m.pendingActionPath = ""
				m.footerError = "Cancelled"
				return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return clearErrorMsg{} })
			}
			return m, nil
		}

		// handle colon typed as a key string so it works across terminals
		if s == ":" {
			if !m.showRootPopup {
				m.showRootPopup = true
				m.rootInput = ""
				m.rootPopupCursor = -1
				m.footerError = ""
			} else {
				m.showRootPopup = false
				m.rootPopupCursor = -1
			}
			m.paneHeight = m.computePaneHeight()
			m.table.SetHeight(m.paneHeight - 1)
			return m, nil
		}

		switch s {
		case "q", "Q":
			// Ignore plain 'q'/'Q' to avoid accidental quit; only ctrl+c quits.
			return m, nil
		case "?":
			// Show help screen
			m.activeModal = ModalHelp
			m.helpScroll = 0
			return m, nil
		case "/":
			// Enter search/filter mode
			if !m.showRootPopup && m.activeModal == ModalNone {
				m.searchMode = true
				m.originalRows = append([]table.Row{}, m.table.Rows()...)
				m.searchInput.SetValue("")
				m.searchTerm = ""
				m.searchInput.Focus()
				m.paneHeight = m.computePaneHeight()
				m.table.SetHeight(m.paneHeight - 1)
				// Warn user if search is scoped to current page only
				currentRoot := m.currentRoot
				if len(m.breadcrumb) > 0 {
					currentRoot = m.breadcrumb[len(m.breadcrumb)-1]
				}
				if off, ok := m.pageOffsets[currentRoot]; ok && off > 0 {
					pageSize := m.getPageSize()
					pg := off/pageSize + 1
					m.footerError, m.footerStatusKind, _ = setFooterStatus(footerStatusInfo,
						fmt.Sprintf("Search limited to current page (pg %d) — PgUp to page 1 for full results", pg), 0)
				}
				return m, m.searchInput.Focus()
			}
			// fall through to root popup input if popup active
			if m.showRootPopup {
				m.rootInput += s
				return m, nil
			}
			return m, nil
		case "ctrl+c":
			// Quit via <ctrl>+c only; do not exit on plain 'q'
			if m.activeModal == ModalConfirmQuit {
				return m, tea.Quit
			}
			m.activeModal = ModalConfirmQuit
			return m, nil
		case "ctrl+e":
			// Open environment selection popup
			if m.activeModal == ModalNone && !m.showActionsMenu {
				m.activeModal = ModalEnvironment
				// Set cursor to current environment
				m.envPopupCursor = 0
				for i, name := range m.envNames {
					if name == m.currentEnv {
						m.envPopupCursor = i
						break
					}
				}
				return m, nil
			}
			return m, nil
		case "ctrl+r", "r":
			// Toggle auto-refresh
			m.autoRefresh = !m.autoRefresh
			if m.autoRefresh {
				m.isLoading = true
				m.apiCallStarted = time.Now()
				initialCmd := m.fetchForRoot(m.currentRoot)
				if initialCmd == nil {
					initialCmd = m.fetchDefinitionsCmd()
				}
				return m, tea.Batch(initialCmd, flashOnCmd(), tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} }), spinnerTickCmd())
			}
			return m, nil
		case "s":
			// Open sort popup
			if m.showRootPopup {
				m.rootInput += s
				return m, nil
			}
			if m.searchMode {
				msg2, kind, cmd := setFooterStatus(footerStatusInfo,
					"Clear search first (Esc) to sort all results", 3*time.Second)
				m.footerError = msg2
				m.footerStatusKind = kind
				return m, cmd
			}
			if m.activeModal == ModalNone && !m.showActionsMenu {
				m.activeModal = ModalSort
				m.sortPopupCursor = 0
				if m.sortColumn >= 0 {
					m.sortPopupCursor = m.sortColumn
				}
				return m, nil
			}
			return m, nil
		case " ":
			// Open context actions menu
			if m.showRootPopup || m.activeModal != ModalNone {
				return m, nil
			}
			row := m.table.SelectedRow()
			if len(row) == 0 {
				return m, nil
			}
			m.actionsMenuItems = m.buildActionsForRoot()
			if len(m.actionsMenuItems) > 0 {
				m.showActionsMenu = true
				m.actionsMenuCursor = 0
			}
			return m, nil
		case "y":
			// Open detail viewer
			if m.showRootPopup {
				m.rootInput += s
				return m, nil
			}
			if m.activeModal == ModalNone && !m.showActionsMenu {
				row := m.table.SelectedRow()
				if len(row) == 0 {
					return m, nil
				}
				m.detailContent = m.buildDetailContent(row)
				m.detailScroll = 0
				m.activeModal = ModalDetailView
				return m, nil
			}
			return m, nil
		case "e":
			if m.showRootPopup {
				m.rootInput += s
				return m, nil
			}
			tableKey := m.currentTableKey()
			if errMsg := m.startEdit(tableKey); errMsg != "" {
				msg2, kind, cmd := setFooterStatus(footerStatusError, errMsg, 5*time.Second)
				m.footerError = msg2
				m.footerStatusKind = kind
				return m, cmd
			}
			return m, nil
		case "esc":
			if m.activeModal == ModalConfirmQuit {
				m.activeModal = ModalNone
				return m, nil
			}
			if m.showRootPopup {
				m.showRootPopup = false
				m.rootInput = ""
				m.rootPopupCursor = -1
				return m, nil
			}
			// Pop from navigation stack and restore previous view state
			if len(m.navigationStack) > 0 {
				// pop last state
				prevState := m.navigationStack[len(m.navigationStack)-1]
				m.navigationStack = m.navigationStack[:len(m.navigationStack)-1]

				// restore complete state
				m.viewMode = prevState.viewMode
				m.breadcrumb = append([]string{}, prevState.breadcrumb...)
				m.contentHeader = prevState.contentHeader
				m.selectedDefinitionKey = prevState.selectedDefinitionKey
				m.selectedInstanceID = prevState.selectedInstanceID
				m.cachedDefinitions = prevState.cachedDefinitions
				m.genericParams = prevState.genericParams // restore drilldown filter params
				m.rowData = prevState.rowData             // restore raw row data for drilldown

				// restore table rows and cursor position
				// restore columns first to ensure rows align with expected column count
				if len(prevState.tableColumns) > 0 {
					m.table.SetColumns(prevState.tableColumns)
				}
				// normalize rows to the restored column count (defensive)
				cols := m.table.Columns()
				rows := prevState.tableRows
				if len(cols) > 0 {
					norm := normalizeRows(rows, len(cols))
					m.table.SetRows(norm)
				} else {
					m.table.SetRows(rows)
				}
				m.table.SetCursor(prevState.tableCursor)

				return m, flashOnCmd()
			}
			return m, nil
		case "enter":
			if m.showRootPopup {
				// If cursor selects from popup list, use that context
				matchingContexts := []string{}
				for _, rc := range m.rootContexts {
					if m.rootInput == "" || strings.HasPrefix(rc, m.rootInput) {
						matchingContexts = append(matchingContexts, rc)
					}
				}
				selectedContext := ""
				if m.rootPopupCursor >= 0 && m.rootPopupCursor < len(matchingContexts) {
					selectedContext = matchingContexts[m.rootPopupCursor]
				} else {
					// fall back to exact match on input
					for _, rc := range m.rootContexts {
						if rc == m.rootInput {
							selectedContext = rc
							break
						}
					}
				}
				if selectedContext != "" {
					rc := selectedContext
					m.currentRoot = rc
					m.showRootPopup = false
					m.rootInput = ""
					m.rootPopupCursor = -1
					// clear any footer error
					m.footerError = ""
					// clear drilldown filter params when switching root context
					m.genericParams = make(map[string]string)
					// reset navigationStack so Esc doesn't take us back to stale state
					m.navigationStack = nil
					// reset breadcrumb and header
					m.breadcrumb = []string{rc}
					m.contentHeader = rc
					// reset viewMode for new context so fallback drilldown doesn't misfire
					switch rc {
					case "process-definitions":
						m.viewMode = "definitions"
					case "process-instances":
						m.viewMode = "instances"
					default:
						m.viewMode = rc
					}
					// If we have a TableDef for this root, set columns and trigger the appropriate fetch
					if def := m.findTableDef(rc); def != nil {
						cols := m.buildColumnsFor(rc, m.paneWidth-4)
						if len(cols) > 0 {
							m.table.SetRows(normalizeRows(nil, len(cols)))
							m.table.SetColumns(cols)
						}
						m.isLoading = true
						m.apiCallStarted = time.Now()
						return m, tea.Batch(m.fetchForRoot(rc), flashOnCmd(), spinnerTickCmd())
					}
					// fallback to definitions fetch
					m.isLoading = true
					m.apiCallStarted = time.Now()
					return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), spinnerTickCmd())
				}
				// no match: ignore
				return m, nil
			}

			// identify the current table (use last breadcrumb entry when available)
			currentTableKey := m.currentRoot
			if len(m.breadcrumb) > 0 {
				currentTableKey = m.breadcrumb[len(m.breadcrumb)-1]
			}

			row := m.table.SelectedRow()
			if len(row) == 0 {
				return m, nil
			}

			// (deferred) save state only when we actually perform a drilldown

			// Config-driven drilldown: consult TableDef.Drilldown (if present)
			if def := m.findTableDef(currentTableKey); def != nil && len(def.Drilldown) > 0 {
				// choose the drill that best matches visible columns (fallback to first)
				var chosen *config.DrillDownDef
				for i := range def.Drilldown {
					d := &def.Drilldown[i]
					col := d.Column
					if col == "" {
						col = "id"
					}
					if idx := m.visibleColumnIndex(def, col); idx >= 0 && idx < len(row) {
						chosen = d
						break
					}
				}
				if chosen == nil {
					chosen = &def.Drilldown[0]
				}

				// resolve drilldown value: prefer rowData (includes hidden columns like id),
				// fall back to visible cell with focus-indicator prefix stripped
				colName := chosen.Column
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
					visIdx := m.visibleColumnIndex(def, colName)
					if visIdx >= 0 && visIdx < len(row) {
						val = stripFocusIndicatorPrefix(fmt.Sprintf("%v", row[visIdx]))
					} else {
						val = stripFocusIndicatorPrefix(fmt.Sprintf("%v", row[0]))
					}
					log.Printf("drilldown: column %q not in rowData at cursor %d, used visible cell: %q", colName, cursor, val)
				}

				// supported runtime targets -> dispatch
				switch chosen.Target {
				case "process-instance", "process-instances":
					// Save current state before performing the drilldown to instances
					cols2 := m.table.Columns()
					rowsCopy2 := append([]table.Row{}, m.table.Rows()...)
					if len(cols2) > 0 {
						norm2 := normalizeRows(rowsCopy2, len(cols2))
						rowsCopy2 = norm2
					}
					currentState2 := viewState{
						viewMode:              m.viewMode,
						breadcrumb:            append([]string{}, m.breadcrumb...),
						contentHeader:         m.contentHeader,
						selectedDefinitionKey: m.selectedDefinitionKey,
						selectedInstanceID:    m.selectedInstanceID,
						tableRows:             rowsCopy2,
						tableCursor:           m.table.Cursor(),
						cachedDefinitions:     m.cachedDefinitions,
						tableColumns:          append([]table.Column{}, cols2...),
						rowData:               append([]map[string]interface{}{}, m.rowData...),
					}
					m.navigationStack = append(m.navigationStack, currentState2)
					// definitions -> instances (expects a process key)
					m.selectedDefinitionKey = val
					m.viewMode = "instances"
					m.breadcrumb = []string{m.currentRoot, "process-instances"}
					m.contentHeader = fmt.Sprintf("%s(%s)", m.currentRoot, val)
					// reset cursor to row 0
					m.table.SetCursor(0)
					return m, tea.Batch(m.fetchInstancesCmd(chosen.Param, val), flashOnCmd())
				case "process-variables", "variables", "variable-instance", "variable-instances":
					// instances -> variables (expects an instance id)
					// Save current state before performing the drilldown to variables
					colsVar := m.table.Columns()
					rowsCopyVar := append([]table.Row{}, m.table.Rows()...)
					if len(colsVar) > 0 {
						normVar := normalizeRows(rowsCopyVar, len(colsVar))
						rowsCopyVar = normVar
					}
					currentStateVar := viewState{
						viewMode:              m.viewMode,
						breadcrumb:            append([]string{}, m.breadcrumb...),
						contentHeader:         m.contentHeader,
						selectedDefinitionKey: m.selectedDefinitionKey,
						selectedInstanceID:    m.selectedInstanceID,
						tableRows:             rowsCopyVar,
						tableCursor:           m.table.Cursor(),
						cachedDefinitions:     m.cachedDefinitions,
						tableColumns:          append([]table.Column{}, colsVar...),
						rowData:               append([]map[string]interface{}{}, m.rowData...),
					}
					m.navigationStack = append(m.navigationStack, currentStateVar)

					m.selectedInstanceID = val
					m.viewMode = "variables"
					m.breadcrumb = append(m.breadcrumb, "variables")
					m.contentHeader = fmt.Sprintf("process-instances(%s)", val)
					// reset cursor to row 0
					m.table.SetCursor(0)
					// immediately set variables columns and clear rows while loading to avoid showing previous rows
					colsVarView := m.buildColumnsFor(dao.ResourceProcessVariables, m.paneWidth-4)
					if len(colsVarView) > 0 {
						// set rows first to match the new column count, then set columns
						m.table.SetRows(normalizeRows(nil, len(colsVarView)))
						m.table.SetColumns(colsVarView)
					} else {
						m.table.SetRows([]table.Row{})
					}
					return m, tea.Batch(m.fetchVariablesCmd(val), flashOnCmd())

				default:
					// Generic drilldown for any configured target
					// Save current state before performing the drilldown
					colsGeneric := m.table.Columns()
					rowsCopyGeneric := append([]table.Row{}, m.table.Rows()...)
					if len(colsGeneric) > 0 {
						normGeneric := normalizeRows(rowsCopyGeneric, len(colsGeneric))
						rowsCopyGeneric = normGeneric
					}
					currentStateGeneric := viewState{
						viewMode:              m.viewMode,
						breadcrumb:            append([]string{}, m.breadcrumb...),
						contentHeader:         m.contentHeader,
						selectedDefinitionKey: m.selectedDefinitionKey,
						selectedInstanceID:    m.selectedInstanceID,
						tableRows:             rowsCopyGeneric,
						tableCursor:           m.table.Cursor(),
						cachedDefinitions:     m.cachedDefinitions,
						tableColumns:          append([]table.Column{}, colsGeneric...),
						genericParams:         m.genericParams, // save current filter params
						rowData:               append([]map[string]interface{}{}, m.rowData...),
					}
					m.navigationStack = append(m.navigationStack, currentStateGeneric)

					// Set up the drilldown: store the filter param and navigate to target table
					m.currentRoot = chosen.Target
					m.genericParams = map[string]string{chosen.Param: val}
					m.breadcrumb = append(m.breadcrumb, chosen.Target)
					m.contentHeader = fmt.Sprintf("%s(%s=%s)", chosen.Target, chosen.Param, val)
					// reset cursor to row 0
					m.table.SetCursor(0)
					// Build columns for target table and clear rows while loading
					colsTarget := m.buildColumnsFor(chosen.Target, m.paneWidth-4)
					if len(colsTarget) > 0 {
						m.table.SetRows(normalizeRows(nil, len(colsTarget)))
						m.table.SetColumns(colsTarget)
					} else {
						m.table.SetRows([]table.Row{})
					}
					return m, tea.Batch(m.fetchGenericCmd(chosen.Target), flashOnCmd())
				}
			}

			// fallback: preserve previous hard-coded drill behaviour
			if m.viewMode == "definitions" {
				// Save current state before fallback drilldown
				cols3 := m.table.Columns()
				rowsCopy3 := append([]table.Row{}, m.table.Rows()...)
				if len(cols3) > 0 {
					norm3 := normalizeRows(rowsCopy3, len(cols3))
					rowsCopy3 = norm3
				}
				currentState3 := viewState{
					viewMode:              m.viewMode,
					breadcrumb:            append([]string{}, m.breadcrumb...),
					contentHeader:         m.contentHeader,
					selectedDefinitionKey: m.selectedDefinitionKey,
					selectedInstanceID:    m.selectedInstanceID,
					tableRows:             rowsCopy3,
					tableCursor:           m.table.Cursor(),
					cachedDefinitions:     m.cachedDefinitions,
					tableColumns:          append([]table.Column{}, cols3...),
					rowData:               append([]map[string]interface{}{}, m.rowData...),
				}
				m.navigationStack = append(m.navigationStack, currentState3)

				key := stripFocusIndicatorPrefix(fmt.Sprintf("%v", row[0]))
				m.selectedDefinitionKey = key
				m.viewMode = "instances"
				m.breadcrumb = []string{m.currentRoot, "process-instances"}
				m.contentHeader = fmt.Sprintf("%s(%s)", m.currentRoot, key)
				// reset cursor to row 0
				m.table.SetCursor(0)
				return m, tea.Batch(m.fetchInstancesCmd("processDefinitionKey", key), flashOnCmd())
			} else if m.viewMode == "instances" {
				// Fallback instances -> variables: save state then drill
				cols4 := m.table.Columns()
				rowsCopy4 := append([]table.Row{}, m.table.Rows()...)
				if len(cols4) > 0 {
					norm4 := normalizeRows(rowsCopy4, len(cols4))
					rowsCopy4 = norm4
				}
				currentState4 := viewState{
					viewMode:              m.viewMode,
					breadcrumb:            append([]string{}, m.breadcrumb...),
					contentHeader:         m.contentHeader,
					selectedDefinitionKey: m.selectedDefinitionKey,
					selectedInstanceID:    m.selectedInstanceID,
					tableRows:             rowsCopy4,
					tableCursor:           m.table.Cursor(),
					cachedDefinitions:     m.cachedDefinitions,
					tableColumns:          append([]table.Column{}, cols4...),
					rowData:               append([]map[string]interface{}{}, m.rowData...),
				}
				m.navigationStack = append(m.navigationStack, currentState4)

				id := stripFocusIndicatorPrefix(fmt.Sprintf("%v", row[0]))
				m.selectedInstanceID = id
				m.viewMode = "variables"
				m.breadcrumb = append(m.breadcrumb, "variables")
				m.contentHeader = fmt.Sprintf("process-instances(%s)", id)
				// reset cursor to row 0
				m.table.SetCursor(0)
				// immediately set variables columns and clear rows while loading to avoid showing previous rows
				colsVarView2 := m.buildColumnsFor(dao.ResourceProcessVariables, m.paneWidth-4)
				if len(colsVarView2) > 0 {
					// set rows first to match the new column count, then set columns
					m.table.SetRows(normalizeRows(nil, len(colsVarView2)))
					m.table.SetColumns(colsVarView2)
				} else {
					m.table.SetRows([]table.Row{})
				}
				return m, tea.Batch(m.fetchVariablesCmd(id), flashOnCmd())
			}

			return m, nil
		case "tab":
			if m.showRootPopup {
				// compute matching contexts
				matchingContexts := []string{}
				for _, rc := range m.rootContexts {
					if m.rootInput == "" || strings.HasPrefix(rc, m.rootInput) {
						matchingContexts = append(matchingContexts, rc)
					}
				}
				if m.rootPopupCursor >= 0 && m.rootPopupCursor < len(matchingContexts) {
					m.rootInput = matchingContexts[m.rootPopupCursor]
				} else if len(m.rootInput) > 0 && len(matchingContexts) > 0 {
					m.rootInput = matchingContexts[0]
				}
			}
			return m, nil
		case "pgdown", "pagedown", "pgdn", "ctrl+f":
			// Page down: advance offset by visible rows and refetch
			root := m.currentRoot
			if len(m.breadcrumb) > 0 {
				root = m.breadcrumb[len(m.breadcrumb)-1]
			}
			pageSize := m.getPageSize()
			if pageSize <= 0 {
				pageSize = 10
			}
			curOff := 0
			if v, ok := m.pageOffsets[root]; ok {
				curOff = v
			}
			newOff := curOff + pageSize
			// if total known, clamp
			if t, ok := m.pageTotals[root]; ok && t > 0 {
				if newOff >= t {
					// move to last page where last row is last
					newOff = t - pageSize
					if newOff < 0 {
						newOff = 0
					}
				}
			}
			m.pageOffsets[root] = newOff
			if newOff == curOff {
				msg2, kind, cmd := setFooterStatus(footerStatusInfo, "Last page", 2*time.Second)
				m.footerError = msg2
				m.footerStatusKind = kind
				return m, cmd
			}
			// keep selection stable
			m.pendingCursorAfterPage = m.table.Cursor()
			return m, tea.Batch(m.fetchForRoot(root), flashOnCmd())
		case "pgup", "pageup", "ctrl+b":
			// Page up: decrease offset by visible rows and refetch
			root := m.currentRoot
			if len(m.breadcrumb) > 0 {
				root = m.breadcrumb[len(m.breadcrumb)-1]
			}
			pageSize := m.getPageSize()
			curOff := 0
			if v, ok := m.pageOffsets[root]; ok {
				curOff = v
			}
			if curOff == 0 {
				msg2, kind, cmd := setFooterStatus(footerStatusInfo, "First page", 2*time.Second)
				m.footerError = msg2
				m.footerStatusKind = kind
				return m, cmd
			}
			newOff := curOff - pageSize
			if newOff < 0 {
				newOff = 0
			}
			m.pageOffsets[root] = newOff
			m.pendingCursorAfterPage = m.table.Cursor()
			return m, tea.Batch(m.fetchForRoot(root), flashOnCmd())
		case "backspace":
			if m.showRootPopup {
				if len(m.rootInput) > 0 {
					m.rootInput = m.rootInput[:len(m.rootInput)-1]
					m.rootPopupCursor = -1
				}
				m.paneHeight = m.computePaneHeight()
				m.table.SetHeight(m.paneHeight - 1)
				return m, nil
			}
			return m, nil
		case "j":
			// Vim down — only when not in popup/modal/search
			if !m.showRootPopup && m.activeModal == ModalNone {
				m.table.MoveDown(1)
				return m, nil
			}
			if m.showRootPopup {
				matchingContexts := []string{}
				for _, rc := range m.rootContexts {
					if m.rootInput == "" || strings.HasPrefix(rc, m.rootInput) {
						matchingContexts = append(matchingContexts, rc)
					}
				}
				maxShow := 8
				if len(matchingContexts) > maxShow {
					matchingContexts = matchingContexts[:maxShow]
				}
				if len(matchingContexts) > 0 {
					if m.rootPopupCursor < 0 {
						m.rootPopupCursor = 1
					} else if m.rootPopupCursor < len(matchingContexts)-1 {
						m.rootPopupCursor++
					}
				}
				return m, nil
			}
			return m, nil
		case "k":
			// Vim up — only when not in popup/modal/search
			if !m.showRootPopup && m.activeModal == ModalNone {
				m.table.MoveUp(1)
				return m, nil
			}
			if m.showRootPopup {
				if m.rootPopupCursor > 0 {
					m.rootPopupCursor--
				} else if m.rootPopupCursor < 0 {
					m.rootPopupCursor = 0
				}
				return m, nil
			}
			return m, nil
		case "G":
			// Vim jump to bottom
			if !m.showRootPopup && m.activeModal == ModalNone {
				rows := m.table.Rows()
				if len(rows) > 0 {
					m.table.SetCursor(len(rows) - 1)
				}
				return m, nil
			}
			if m.showRootPopup {
				m.rootInput += s
			}
			return m, nil
		case "g":
			// Vim gg sequence: first g sets pendingG, second g jumps to top
			if m.showRootPopup {
				m.rootInput += s
				return m, nil
			}
			if m.activeModal == ModalNone {
				if m.pendingG {
					m.pendingG = false
					m.table.SetCursor(0)
					return m, nil
				}
				m.pendingG = true
				return m, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg { return clearPendingGMsg{} })
			}
			return m, nil
		case "ctrl+d":
			// Half-page down OR delete (existing behavior)
			if m.viewMode == "instances" {
				row := m.table.SelectedRow()
				if len(row) > 0 {
					m.pendingDeleteID = stripFocusIndicatorPrefix(fmt.Sprintf("%v", row[0]))
					if len(row) > 1 {
						m.pendingDeleteLabel = fmt.Sprintf("%v", row[1])
					} else {
						m.pendingDeleteLabel = ""
					}
					m.activeModal = ModalConfirmDelete
				}
				return m, nil
			}
			// Half-page down for non-instances views
			pageSize := m.getPageSize()
			if pageSize <= 0 {
				pageSize = 10
			}
			m.table.MoveDown(pageSize / 2)
			return m, nil
		case "up":
			if m.showRootPopup {
				if m.rootPopupCursor > 0 {
					m.rootPopupCursor--
				} else if m.rootPopupCursor < 0 {
					m.rootPopupCursor = 0
				}
				return m, nil
			}
			// fall through to table navigation via component update
		case "down":
			if m.showRootPopup {
				matchingContexts := []string{}
				for _, rc := range m.rootContexts {
					if m.rootInput == "" || strings.HasPrefix(rc, m.rootInput) {
						matchingContexts = append(matchingContexts, rc)
					}
				}
				maxShow := 8
				if len(matchingContexts) > maxShow {
					matchingContexts = matchingContexts[:maxShow]
				}
				if len(matchingContexts) > 0 {
					if m.rootPopupCursor < 0 {
						m.rootPopupCursor = 1
					} else if m.rootPopupCursor < len(matchingContexts)-1 {
						m.rootPopupCursor++
					}
				}
				return m, nil
			}
			// fall through to table navigation via component update
		case "ctrl+u":
			// Vim half-page up
			pageSize := m.getPageSize()
			if pageSize <= 0 {
				pageSize = 10
			}
			m.table.MoveUp(pageSize / 2)
			return m, nil
		case "1", "2", "3", "4":
			// numeric breadcrumb navigation when popup not active
			if !m.showRootPopup {
				n := int(s[0] - '1') // 0-based
				if n < len(m.breadcrumb) {
					return m, (&m).navigateToBreadcrumb(n)
				}
			}
			return m, nil
		default:
			// typing into the root input when popup active
			if m.showRootPopup {
				if len(s) == 1 {
					m.rootInput += s
					m.rootPopupCursor = -1
					m.paneHeight = m.computePaneHeight()
					m.table.SetHeight(m.paneHeight - 1)
				}
				return m, nil
			}
			// otherwise don't intercept the key: let the list/table components handle navigation
			// fall through to component updates below
		}

	case tea.WindowSizeMsg:
		// Resize UI to take full terminal size
		width := msg.Width
		height := msg.Height
		// store terminal size for View footer alignment
		m.lastWidth = width
		m.lastHeight = height
		// Reserve lines: compact header is 3 rows, context selection 1 line (when active), footer 1 line
		headerLines := 3 // compactHeader placed at 3 rows
		contextSelectionLines := 1
		footerLines := 1
		// content height = terminal height minus header, context box, and footer
		contentHeight := height - headerLines - contextSelectionLines - footerLines
		if contentHeight < 3 {
			contentHeight = 3
		}
		// compute left column width and content pane width
		leftW := width / 4
		if leftW < 12 {
			leftW = 12
		}
		m.paneWidth = width - leftW - 4
		if m.paneWidth < 10 {
			m.paneWidth = 10
		}
		m.paneHeight = contentHeight
		m.list.SetSize(leftW-2, contentHeight-1)
		// table inner width should be pane inner width minus border(2) and padding(2)
		tableInner := m.paneWidth - 4
		if tableInner < 10 {
			tableInner = 10
		}
		m.table.SetWidth(tableInner)
		m.table.SetHeight(contentHeight - 1)
		return m, nil
	case refreshMsg:
		if m.autoRefresh {
			cmd := m.fetchForRoot(m.currentRoot)
			if cmd == nil {
				cmd = m.fetchDefinitionsCmd()
			}
			return m, tea.Batch(cmd, flashOnCmd(), tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} }), spinnerTickCmd())
		}
	case healthTickMsg:
		return m, tea.Batch(
			m.checkEnvironmentHealthCmd(m.currentEnv),
			tea.Tick(60*time.Second, func(time.Time) tea.Msg { return healthTickMsg{} }),
		)
	case flashOnMsg:
		m.flashActive = true
		// schedule turning the flash off after 200ms
		return m, tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg { return flashOffMsg{} })
	case flashOffMsg:
		m.flashActive = false
	case spinnerTickMsg:
		if m.isLoading {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
			return m, spinnerTickCmd()
		}
	case definitionsLoadedMsg:
		m.pageTotals[dao.ResourceProcessDefinitions] = msg.count
		m.applyDefinitions(msg.definitions)
		if !m.apiCallStarted.IsZero() {
			m.lastAPILatency = time.Since(m.apiCallStarted)
			m.apiCallStarted = time.Time{}
		}
		m.isLoading = false
	case instancesLoadedMsg:
		m.applyInstances(msg.instances)
		if !m.apiCallStarted.IsZero() {
			m.lastAPILatency = time.Since(m.apiCallStarted)
			m.apiCallStarted = time.Time{}
		}
		m.isLoading = false
	case variablesLoadedMsg:
		// show variables table
		m.applyVariables(msg.variables)
		if !m.apiCallStarted.IsZero() {
			m.lastAPILatency = time.Since(m.apiCallStarted)
			m.apiCallStarted = time.Time{}
		}
		m.isLoading = false
	case instancesWithCountMsg:
		// set known total for instances root
		m.pageTotals[dao.ResourceProcessInstances] = msg.count
		m.applyInstances(msg.instances)
		if !m.apiCallStarted.IsZero() {
			m.lastAPILatency = time.Since(m.apiCallStarted)
			m.apiCallStarted = time.Time{}
		}
		m.isLoading = false
	case genericLoadedMsg:
		// Apply generic fetched collection into the table using the table definition if available

		// Strip _meta_count FIRST before inferring columns from data
		if len(msg.items) > 0 {
			if v, ok := msg.items[0]["_meta_count"]; ok {
				if n, ok2 := v.(float64); ok2 {
					m.pageTotals[msg.root] = int(n)
					msg.items = msg.items[1:]
				} else if n2, ok3 := v.(int); ok3 {
					m.pageTotals[msg.root] = n2
					msg.items = msg.items[1:]
				}
			}
		}

		def := m.findTableDef(msg.root)
		var cols []table.Column
		if def != nil {
			cols = m.buildColumnsFor(msg.root, m.paneWidth-4)
			// If buildColumnsFor returned only the "EMPTY" fallback (when all columns are invisible),
			// infer columns from the data instead
			if len(cols) == 1 && cols[0].Title == "EMPTY" {
				cols = nil
			}
		}
		if len(cols) == 0 {
			// infer columns from first item keys (after stripping _meta_count)
			if len(msg.items) > 0 {
				keys := make([]string, 0, len(msg.items[0]))
				for k := range msg.items[0] {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					cols = append(cols, table.Column{Title: strings.ToUpper(k), Width: 20})
				}
			}
		}
		// Build rows from items; also capture raw data for drilldown column lookup
		rows := make([]table.Row, 0, len(msg.items))
		rd := make([]map[string]interface{}, 0, len(msg.items))
		hasDrilldown := def != nil && len(def.Drilldown) > 0
		for _, it := range msg.items {
			rd = append(rd, it)
			if len(cols) == 0 {
				// fallback: add single column with JSON representation
				rows = append(rows, table.Row{fmt.Sprintf("%v", it)})
				continue
			}
			r := make(table.Row, len(cols))
			for i, col := range cols {
				// prefer original column name from TableDef when available
				key := strings.ToLower(col.Title)
				// If title contains spaces or was uppercased, try original key forms
				val := ""
				if v, ok := it[key]; ok {
					val = fmt.Sprintf("%v", v)
				} else {
					// try lowercase and camel-case variants
					if v, ok := it[strings.ToLower(col.Title)]; ok {
						val = fmt.Sprintf("%v", v)
					} else if v, ok := it[col.Title]; ok {
						val = fmt.Sprintf("%v", v)
					}
				}
				r[i] = val
			}
			if hasDrilldown && len(r) > 0 {
				r[0] = "▶ " + r[0]
			}
			rows = append(rows, r)
		}
		m.rowData = rd

		if len(cols) > 0 {
			m.table.SetColumns(cols)
		}
		normalized := normalizeRows(rows, len(cols))
		colorized := colorizeRows(msg.root, normalized, cols)
		m.setTableRowsSorted(colorized)
		if m.sortColumn >= 0 {
			m.applySortIndicatorToColumns()
		}
		// re-apply locked search filter if active
		if !m.searchMode && m.searchTerm != "" {
			filtered := filterRows(m.table.Rows(), m.searchTerm)
			m.table.SetRows(filtered)
		}
		// restore pending cursor after page operations for generic loads
		if m.pendingCursorAfterPage >= 0 {
			r := m.table.Rows()
			last := len(r) - 1
			pos := m.pendingCursorAfterPage
			if pos > last {
				pos = last
			}
			if pos < 0 {
				pos = 0
			}
			m.table.SetCursor(pos)
			m.pendingCursorAfterPage = -1
		}
		if !m.apiCallStarted.IsZero() {
			m.lastAPILatency = time.Since(m.apiCallStarted)
			m.apiCallStarted = time.Time{}
		}
		m.isLoading = false
	case editSavedMsg:
		rows := m.table.Rows()
		if msg.rowIndex >= 0 && msg.rowIndex < len(rows) {
			row := rows[msg.rowIndex]
			if msg.colIndex >= 0 && msg.colIndex < len(row) {
				row[msg.colIndex] = msg.value
				rows[msg.rowIndex] = row
				m.table.SetRows(rows)
			}
		}
		// keep rowData consistent with optimistic table update
		if msg.dataKey != "" && msg.rowIndex >= 0 && msg.rowIndex < len(m.rowData) {
			m.rowData[msg.rowIndex][msg.dataKey] = msg.value
		}
		msg2, kind, cmd := setFooterStatus(footerStatusSuccess, "✓ Saved", 2*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		return m, cmd
	case dataLoadedMsg:
		// keep backward compatibility: only apply definitions to avoid auto-drilldown
		m.applyDefinitions(msg.definitions)
	case terminatedMsg:
		// find index before removing (removeInstance shifts rows)
		rows := m.table.Rows()
		deleteIdx := -1
		for i, r := range rows {
			if rowInstanceID(r) == msg.id {
				deleteIdx = i
				break
			}
		}
		m.removeInstance(msg.id)
		// remove corresponding rowData entry
		if deleteIdx >= 0 && deleteIdx < len(m.rowData) {
			m.rowData = append(m.rowData[:deleteIdx], m.rowData[deleteIdx+1:]...)
		}
		// show success feedback (consistent with suspendedMsg/resumedMsg/retriedMsg)
		msg2, kind, cmd := setFooterStatus(footerStatusSuccess,
			fmt.Sprintf("✓ Terminated %s", msg.id), 3*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		return m, cmd
	case suspendedMsg:
		msg2, kind, statusCmd := setFooterStatus(footerStatusSuccess, fmt.Sprintf("✓ Suspended %s", msg.id), 3*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		m.isLoading = true
		m.apiCallStarted = time.Now()
		fetchCmd := m.fetchForRoot(m.currentRoot)
		if fetchCmd == nil {
			fetchCmd = m.fetchDefinitionsCmd()
		}
		return m, tea.Batch(statusCmd, fetchCmd, spinnerTickCmd())
	case resumedMsg:
		msg2, kind, statusCmd := setFooterStatus(footerStatusSuccess, fmt.Sprintf("✓ Resumed %s", msg.id), 3*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		m.isLoading = true
		m.apiCallStarted = time.Now()
		fetchCmd := m.fetchForRoot(m.currentRoot)
		if fetchCmd == nil {
			fetchCmd = m.fetchDefinitionsCmd()
		}
		return m, tea.Batch(statusCmd, fetchCmd, spinnerTickCmd())
	case retriedMsg:
		msg2, kind, statusCmd := setFooterStatus(footerStatusSuccess, fmt.Sprintf("✓ Retried %s", msg.id), 3*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		m.isLoading = true
		m.apiCallStarted = time.Now()
		fetchCmd := m.fetchForRoot(m.currentRoot)
		if fetchCmd == nil {
			fetchCmd = m.fetchDefinitionsCmd()
		}
		return m, tea.Batch(statusCmd, fetchCmd, spinnerTickCmd())
	case actionExecutedMsg:
		msg2, kind, statusCmd := setFooterStatus(footerStatusSuccess, fmt.Sprintf("✓ %s", msg.label), 3*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		m.isLoading = true
		m.apiCallStarted = time.Now()
		fetchCmd := m.fetchForRoot(m.currentRoot)
		if fetchCmd == nil {
			fetchCmd = m.fetchDefinitionsCmd()
		}
		return m, tea.Batch(statusCmd, fetchCmd, spinnerTickCmd())
	case envStatusMsg:
		// Update environment status
		m.envStatus[msg.env] = msg.status
	case errMsg:
		errText := friendlyError(m.currentEnv, msg.err) + " — Ctrl+r to retry"
		msg2, kind, cmd := setFooterStatus(footerStatusError, errText, 8*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		m.isLoading = false
		return m, cmd
	case clearErrorMsg:
		m.footerError = ""
		m.footerStatusKind = footerStatusNone
		m.isLoading = false
	case clearPendingGMsg:
		m.pendingG = false
	}

	var cmd tea.Cmd
	// update list/table internals as usual
	prevIndex := m.list.Index()
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	newIndex := m.list.Index()
	changed := prevIndex != newIndex || newIndex != m.lastListIndex
	_ = changed
	m.lastListIndex = newIndex

	// Defensive: ensure table rows match the number of columns before updating
	// This avoids panics in the underlying bubbles table when rows/cols get out of sync
	if cols := m.table.Columns(); len(cols) > 0 {
		rows := m.table.Rows()
		norm := normalizeRows(rows, len(cols))
		// Only reset rows if normalization changed anything
		changed := false
		if len(norm) != len(rows) {
			changed = true
		} else {
			for i := range norm {
				if len(norm[i]) != len(rows[i]) {
					changed = true
					break
				}
			}
		}
		if changed {
			log.Printf("normalized table rows: cols=%d rows_before=%d rows_after=%d", len(cols), len(rows), len(norm))
			m.table.SetRows(norm)
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	m.selectedInstanceID = rowInstanceID(m.table.SelectedRow())

	return m, tea.Batch(cmds...)
}
