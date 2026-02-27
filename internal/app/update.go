package app

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/dao"
	"github.com/kthoms/o8n/internal/operaton"
	"github.com/kthoms/o8n/internal/validation"
)

func (m model) Update(msg tea.Msg) (retModel tea.Model, retCmd tea.Cmd) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			id := fmt.Sprintf("%d", time.Now().UnixNano())
			_ = os.MkdirAll("./debug", 0755)
			screenFile := fmt.Sprintf("./debug/screen-%s.txt", id)
			_ = os.WriteFile(screenFile, []byte(lastRenderedView), 0644)
			log.Printf("[panic] %v | screen: debug/screen-%s.txt\n%s", r, id, stack)
			// Mutate return values to surface error in UI without crashing
			recovered := m
			recovered.footerError = fmt.Sprintf("internal error: %v — see debug/o8n.log", r)
			recovered.footerStatusKind = footerStatusError
			recovered.isLoading = false
			retModel = recovered
			retCmd = nil
		}
	}()

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
			case "up":
				if m.sortColumn >= 0 && m.sortPopupCursor == 0 {
					m.sortPopupCursor = -1 // move to clear item
				} else if m.sortPopupCursor > 0 {
					m.sortPopupCursor--
				}
				return m, nil
			case "down":
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
			case "up":
				if m.actionsMenuCursor > 0 {
					m.actionsMenuCursor--
				}
				return m, nil
			case "down":
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
			case "down":
				if m.detailScroll < maxDetailScroll {
					m.detailScroll++
				}
				return m, nil
			case "up":
				if m.detailScroll > 0 {
					m.detailScroll--
				}
				return m, nil
			case "ctrl+d":
				// Vim half-page scroll in detail view (vim mode only)
				if m.vimMode {
					m.detailScroll += 10
					if m.detailScroll > maxDetailScroll {
						m.detailScroll = maxDetailScroll
					}
				}
				return m, nil
			case "ctrl+u":
				// Vim half-page scroll in detail view (vim mode only)
				if m.vimMode {
					m.detailScroll -= 10
					if m.detailScroll < 0 {
						m.detailScroll = 0
					}
				}
				return m, nil
			case "G":
				// Vim jump to bottom in detail view (vim mode only)
				if m.vimMode {
					m.detailScroll = maxDetailScroll
				}
				return m, nil
			case "g":
				// Vim gg in detail view (vim mode only)
				if m.vimMode {
					if m.pendingG {
						m.pendingG = false
						m.detailScroll = 0
						return m, nil
					}
					m.pendingG = true
					return m, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg { return clearPendingGMsg{} })
				}
				return m, nil
			case "j":
				// Vim scroll down in detail view (vim mode only)
				if m.vimMode {
					m.detailScroll++
					if m.detailScroll > maxDetailScroll {
						m.detailScroll = maxDetailScroll
					}
				}
				return m, nil
			case "k":
				// Vim scroll up in detail view (vim mode only)
				if m.vimMode {
					m.detailScroll--
					if m.detailScroll < 0 {
						m.detailScroll = 0
					}
				}
				return m, nil
			}
			return m, nil
		}

		// Handle environment popup keys
		if m.activeModal == ModalEnvironment {
			switch s {
			case "esc":
				m.activeModal = ModalNone
				return m, nil
			case "up":
				if m.envPopupCursor > 0 {
					m.envPopupCursor--
				}
				return m, nil
			case "down":
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
						m.prepareStateTransition(transitionEnvSwitch)
						m.breadcrumb = []string{m.currentRoot}
						m.isLoading = true
						m.apiCallStarted = time.Now()
						return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), m.checkEnvironmentHealthCmd(targetEnv), spinnerTickCmd(), m.saveStateCmd())
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
			case "down", "ctrl+d":
				if m.helpScroll < maxScroll {
					m.helpScroll++
				}
				return m, nil
			case "up", "ctrl+u":
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
				// Generic save via EditAction config (preferred)
				if def := m.findTableDef(m.editTableKey); def != nil && def.EditAction != nil {
					act := *def.EditAction
					varName := m.variableNameForRow(m.editTableKey, row)
					parsedValue, err := parseInputValue(m.editInput.Value(), inputType)
					if err != nil {
						m.editError = err.Error()
						return m, nil
					}
					valueStr := fmt.Sprintf("%v", parsedValue)
					rowIndex := m.editRowIndex
					colIndex := col.index
					displayValue := m.editInput.Value()
					idCol := act.IDColumn
					if idCol == "" {
						idCol = "id"
					}
					nameCol := act.NameColumn
					if nameCol == "" {
						nameCol = "name"
					}
					id := m.resolveRowValue(row, idCol)
					name := m.resolveRowValue(row, nameCol)
					m.activeModal = ModalNone
					m.editError = ""
					m.editInput.Blur()
					return m, tea.Batch(m.executeEditActionCmd(act, id, name, m.selectedInstanceID, valueStr, typeName, rowIndex, colIndex, displayValue, varName), flashOnCmd())
				}
				// Fallback: legacy variable table save
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
			confirmAction := func() (tea.Model, tea.Cmd) {
				m.activeModal = ModalNone
				m.confirmFocusedBtn = 1 // reset to cancel for next time
				if m.pendingAction != nil {
					act := *m.pendingAction
					resolvedPath := m.pendingActionPath
					m.pendingAction = nil
					m.pendingActionID = ""
					m.pendingActionPath = ""
					return m, tea.Batch(m.executeActionCmd(act, resolvedPath), flashOnCmd())
				}
				if m.pendingDeleteID != "" {
					return m, tea.Batch(m.terminateInstanceCmd(m.pendingDeleteID), flashOnCmd())
				}
				return m, nil
			}
			cancelAction := func() (tea.Model, tea.Cmd) {
				m.activeModal = ModalNone
				m.confirmFocusedBtn = 1
				m.pendingDeleteID = ""
				m.pendingDeleteLabel = ""
				m.pendingAction = nil
				m.pendingActionID = ""
				m.pendingActionPath = ""
				m.footerError = "Cancelled"
				return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return clearErrorMsg{} })
			}

			switch {
			case s == "ctrl+d": // confirm key always confirms
				return confirmAction()
			case s == "tab":
				if m.confirmFocusedBtn == 0 {
					m.confirmFocusedBtn = 1
				} else {
					m.confirmFocusedBtn = 0
				}
			case s == "enter":
				if m.confirmFocusedBtn == 0 {
					return confirmAction()
				}
				return cancelAction()
			case s == "esc":
				return cancelAction()
			}
			return m, nil
		}

		if m.activeModal == ModalConfirmQuit {
			switch {
			case s == "ctrl+c": // confirm key always quits
				m.quitting = true
				return m, tea.Quit
			case s == "tab":
				if m.confirmFocusedBtn == 0 {
					m.confirmFocusedBtn = 1
				} else {
					m.confirmFocusedBtn = 0
				}
			case s == "enter":
				if m.confirmFocusedBtn == 0 {
					m.quitting = true
					return m, tea.Quit
				}
				m.activeModal = ModalNone
				m.confirmFocusedBtn = 1
			case s == "esc":
				m.activeModal = ModalNone
				m.confirmFocusedBtn = 1
			}
			return m, nil
		}

		if m.activeModal == ModalTaskComplete {
			switch s {
			case "esc":
				m.closeTaskCompleteDialog()
				return m, nil
			case "tab":
				m.taskCompleteTabForward()
				return m, nil
			case "shift+tab", "backtab":
				m.taskCompleteTabBackward()
				return m, nil
			case " ", "space":
				if m.taskCompleteFocus == focusTaskField && len(m.taskCompleteFields) > 0 {
					f := &m.taskCompleteFields[m.taskCompletePos]
					if f.varType == "bool" {
						current := strings.TrimSpace(strings.ToLower(f.input.Value()))
						if current == "true" {
							f.input.SetValue("false")
						} else {
							f.input.SetValue("true")
						}
						f.input.CursorEnd()
						f.error = ""
					}
				}
				return m, nil
			case "enter":
				switch m.taskCompleteFocus {
				case focusTaskField:
					// Advance focus to [Complete] button
					m.blurCurrentTaskField()
					m.taskCompleteFocus = focusTaskComplete
				case focusTaskComplete:
					if cmd := m.submitTaskComplete(); cmd != nil {
						return m, cmd
					}
				case focusTaskBack:
					m.closeTaskCompleteDialog()
				}
				return m, nil
			default:
				// Feed printable keystrokes to the focused field
				if m.taskCompleteFocus == focusTaskField && len(m.taskCompleteFields) > 0 {
					var cmd tea.Cmd
					f := &m.taskCompleteFields[m.taskCompletePos]
					f.input, cmd = f.input.Update(msg)
					f.error = m.validateTaskFieldValue(*f)
					return m, cmd
				}
			}
			return m, nil
		}

		// handle colon typed as a key string so it works across terminals
		if s == ":" {
			if m.popup.mode == popupModeNone {
				m.popup.mode = popupModeContext
				m.popup.input = ""
				m.popup.cursor = -1
				m.footerError = ""
			} else {
				m.popup.mode = popupModeNone
				m.popup.cursor = -1
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
			// Open popup in search mode (replaces inline search bar)
			if m.popup.mode == popupModeNone && m.activeModal == ModalNone && !m.searchMode {
				// close existing inline search if active
				m.searchMode = false
				m.searchInput.Blur()
				// save original rows and open popup in search mode
				m.originalRows = append([]table.Row{}, m.table.Rows()...)
				m.searchTerm = ""
				m.popup.mode = popupModeSearch
				m.popup.input = ""
				m.popup.cursor = -1
				m.popup.offset = 0
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
				return m, nil
			}
			// fall through to root popup input if popup active
			if m.popup.mode != popupModeNone {
				m.popup.input += s
				if m.popup.mode == popupModeSearch {
					m.applySearchFromPopup()
				}
				return m, nil
			}
			return m, nil
		case "ctrl+c":
			// Quit via <ctrl>+c only; do not exit on plain 'q'
			m.activeModal = ModalConfirmQuit
			m.confirmFocusedBtn = 1 // default Cancel focused (safer)
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
		case "ctrl+t":
			// Open skin/theme picker
			if m.activeModal == ModalNone && m.popup.mode == popupModeNone {
				m.popup.mode = popupModeSkin
				m.popup.input = ""
				m.popup.cursor = -1
				m.popup.offset = 0
				m.popup.title = "theme"
				m.popup.hint = "↑↓:preview  Enter:apply  Esc:revert"
				m.popup.previewSkin = m.activeSkin
				m.popup.items = m.availableSkins
				m.paneHeight = m.computePaneHeight()
				m.table.SetHeight(m.paneHeight - 1)
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
		case "L":
			// Toggle latency display in footer
			m.showLatency = !m.showLatency
			return m, m.saveStateCmd()
		case "s":
			// Open sort popup
			if m.popup.mode != popupModeNone {
				m.popup.input += s
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
			if m.popup.mode != popupModeNone || m.activeModal != ModalNone {
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
			if m.popup.mode != popupModeNone {
				m.popup.input += s
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
			if m.popup.mode != popupModeNone {
				m.popup.input += s
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
		case "c":
			if m.popup.mode != popupModeNone {
				m.popup.input += s
				return m, nil
			}
			// Task claim: only active on task table with no modal/menu open
			if def := m.findTableDef(m.currentTableKey()); def != nil && def.Name == "task" && m.activeModal == ModalNone && !m.showActionsMenu {
				row := m.table.SelectedRow()
				if len(row) == 0 {
					return m, nil
				}
				assignee := m.resolveRowValue(row, "assignee")
				taskID := m.resolveRowValue(row, "id")
				taskName := m.resolveRowValue(row, "name")
				currentUser := m.currentUsername()
				if assignee != "" && assignee != currentUser {
					msg2, kind, cmd := setFooterStatus(footerStatusError, fmt.Sprintf("Already claimed by %s", assignee), 5*time.Second)
					m.footerError = msg2
					m.footerStatusKind = kind
					return m, cmd
				}
				if assignee == currentUser && currentUser != "" {
					msg2, kind, cmd := setFooterStatus(footerStatusInfo, "You already own this task — press Enter to complete", 5*time.Second)
					m.footerError = msg2
					m.footerStatusKind = kind
					return m, cmd
				}
				// assignee is empty — claim it
				m.isLoading = true
				m.apiCallStarted = time.Now()
				return m, tea.Batch(m.claimTaskCmd(taskID, currentUser, taskName), spinnerTickCmd())
			}
			return m, nil
		case "u":
			if m.popup.mode != popupModeNone {
				m.popup.input += s
				return m, nil
			}
			// Task unclaim: only active on task table with no modal/menu open
			if def := m.findTableDef(m.currentTableKey()); def != nil && def.Name == "task" && m.activeModal == ModalNone && !m.showActionsMenu {
				row := m.table.SelectedRow()
				if len(row) == 0 {
					return m, nil
				}
				assignee := m.resolveRowValue(row, "assignee")
				taskID := m.resolveRowValue(row, "id")
				taskName := m.resolveRowValue(row, "name")
				currentUser := m.currentUsername()
				if assignee == "" {
					msg2, kind, cmd := setFooterStatus(footerStatusError, "Task is not claimed", 5*time.Second)
					m.footerError = msg2
					m.footerStatusKind = kind
					return m, cmd
				}
				if assignee != currentUser {
					msg2, kind, cmd := setFooterStatus(footerStatusError, fmt.Sprintf("Task is assigned to %s, not you", assignee), 5*time.Second)
					m.footerError = msg2
					m.footerStatusKind = kind
					return m, cmd
				}
				// assignee == currentUser — unclaim
				m.isLoading = true
				m.apiCallStarted = time.Now()
				return m, tea.Batch(m.unclaimTaskCmd(taskID, taskName), spinnerTickCmd())
			}
			return m, nil
		case "esc":
			if m.popup.mode == popupModeSkin {
				// Revert skin to what it was before preview
				if m.popup.previewSkin != "" {
					m.activeSkin = m.popup.previewSkin
					if skin, err := loadSkin(m.popup.previewSkin + ".yaml"); err == nil {
						m.skin = skin
						m.applyStyle()
					}
				}
				m.popup.mode = popupModeNone
				m.popup.input = ""
				m.popup.cursor = -1
				return m, nil
			}
			if m.popup.mode == popupModeSearch {
				// Cancel search: restore original rows
				m.popup.mode = popupModeNone
				m.popup.input = ""
				m.popup.cursor = -1
				m.popup.offset = 0
				m.searchTerm = ""
				m.table.SetRows(m.originalRows)
				m.footerError = ""
				m.footerStatusKind = footerStatusNone
				m.paneHeight = m.computePaneHeight()
				m.table.SetHeight(m.paneHeight - 1)
				return m, nil
			}
			if m.popup.mode != popupModeNone {
				m.popup.mode = popupModeNone
				m.popup.input = ""
				m.popup.cursor = -1
				m.popup.offset = 0
				return m, nil
			}
			// Pop from navigation stack and restore previous view state
			if len(m.navigationStack) > 0 {
				// clear sort/search state before restoring parent state
				m.prepareStateTransition(transitionBack)
				// pop last state
				prevState := m.navigationStack[len(m.navigationStack)-1]
				m.navigationStack = m.navigationStack[:len(m.navigationStack)-1]

				// restore complete state
				m.viewMode = prevState.viewMode
				m.currentRoot = prevState.viewMode // fix: restore currentRoot so title shows correct count
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

				// re-fetch to refresh data and count; preserve cursor after load
				m.pendingCursorAfterPage = prevState.tableCursor
				m.isLoading = true
				m.apiCallStarted = time.Now()
				return m, tea.Batch(m.fetchForRoot(m.currentRoot), flashOnCmd(), spinnerTickCmd(), m.saveStateCmd())
			}
			return m, nil
		case "enter", "right":
			// right arrow only handles drilldown (not popup selection or other modals)
			if s == "right" && (m.popup.mode != popupModeNone || m.activeModal != ModalNone || m.searchMode) {
				return m, nil
			}
			if m.popup.mode == popupModeSkin {
				// Commit current skin — already applied via live preview
				m.popup.mode = popupModeNone
				m.popup.input = ""
				m.popup.cursor = -1
				return m, m.saveStateCmd()
			}
			if m.popup.mode == popupModeSearch {
				// Lock the search filter and close popup
				m.popup.mode = popupModeNone
				m.popup.input = ""
				m.popup.cursor = -1
				m.popup.offset = 0
				// searchTerm is already set; filtered rows are in the table
				m.paneHeight = m.computePaneHeight()
				m.table.SetHeight(m.paneHeight - 1)
				return m, nil
			}
			if m.popup.mode != popupModeNone {
				// If cursor selects from popup list, use that context
				matchingContexts := m.popupItems()
				selectedContext := ""
				if m.popup.cursor >= 0 && m.popup.cursor < len(matchingContexts) {
					selectedContext = matchingContexts[m.popup.cursor]
				} else {
					// fall back to exact match on input
					for _, rc := range m.rootContexts {
						if rc == m.popup.input {
							selectedContext = rc
							break
						}
					}
				}
				if selectedContext != "" {
					rc := selectedContext
					m.currentRoot = rc
					m.popup.mode = popupModeNone
					m.popup.input = ""
					m.popup.cursor = -1
					// clear any footer error
					m.footerError = ""
					// centralized state cleanup for context switch
					m.prepareStateTransition(transitionContextSwitch)
					// reset breadcrumb and header
					m.breadcrumb = []string{rc}
					m.contentHeader = rc
					m.viewMode = rc
					// If we have a TableDef for this root, set columns and trigger the appropriate fetch
					if def := m.findTableDef(rc); def != nil {
						cols := m.buildColumnsFor(rc, m.paneWidth-4)
						if len(cols) > 0 {
							m.table.SetRows(normalizeRows(nil, len(cols)))
							m.table.SetColumns(cols)
						}
						m.isLoading = true
						m.apiCallStarted = time.Now()
						return m, tea.Batch(m.fetchForRoot(rc), flashOnCmd(), spinnerTickCmd(), m.saveStateCmd())
					}
					// no table def: still attempt fetch using /{rc} path
					m.isLoading = true
					m.apiCallStarted = time.Now()
					return m, tea.Batch(m.fetchForRoot(rc), flashOnCmd(), spinnerTickCmd(), m.saveStateCmd())
				}
				// no match: ignore
				return m, nil
			}

			// Task table: intercept Enter to open completion dialog (instead of drilldown)
			if s == "enter" && m.popup.mode == popupModeNone {
				// Only intercept Enter for the explicit "task" TableDef; otherwise fall through to generic drilldown
				if def := m.findTableDef(m.currentTableKey()); def != nil && def.Name == "task" {
					row := m.table.SelectedRow()
					if len(row) == 0 {
						return m, nil
					}
					assignee := m.resolveRowValue(row, "assignee")
					taskID := m.resolveRowValue(row, "id")
					taskName := m.resolveRowValue(row, "name")
					currentUser := m.currentUsername()
					if assignee == "" {
						msg2, kind, cmd := setFooterStatus(footerStatusError, "Claim this task first (c)", 5*time.Second)
						m.footerError = msg2
						m.footerStatusKind = kind
						return m, cmd
					}
					if assignee != currentUser {
						msg2, kind, cmd := setFooterStatus(footerStatusError, fmt.Sprintf("Task is assigned to %s", assignee), 5*time.Second)
						m.footerError = msg2
						m.footerStatusKind = kind
						return m, cmd
					}
					// Own task — fetch variables and open dialog
					m.isLoading = true
					m.apiCallStarted = time.Now()
					m.footerError, m.footerStatusKind, _ = setFooterStatus(footerStatusLoading, "Loading task variables…", 0)
					return m, tea.Batch(m.fetchTaskVariablesCmd(taskID, taskName), spinnerTickCmd())
				}
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

				// supported runtime targets -> generic drilldown for all configured targets
				{
					// Save current state before drilldown, then clear sort/search for new view
					m.prepareStateTransition(transitionDrilldown)
					colsDrill := m.table.Columns()
					rowsCopyDrill := append([]table.Row{}, m.table.Rows()...)
					if len(colsDrill) > 0 {
						rowsCopyDrill = normalizeRows(rowsCopyDrill, len(colsDrill))
					}
					currentStateDrill := viewState{
						viewMode:              m.viewMode,
						breadcrumb:            append([]string{}, m.breadcrumb...),
						contentHeader:         m.contentHeader,
						selectedDefinitionKey: m.selectedDefinitionKey,
						selectedInstanceID:    m.selectedInstanceID,
						tableRows:             rowsCopyDrill,
						tableCursor:           m.table.Cursor(),
						cachedDefinitions:     m.cachedDefinitions,
						tableColumns:          append([]table.Column{}, colsDrill...),
						genericParams:         m.genericParams,
						rowData:               append([]map[string]interface{}{}, m.rowData...),
					}
					m.navigationStack = append(m.navigationStack, currentStateDrill)

					// Persist values used by edit/save (variable editing needs selectedInstanceID)
					switch chosen.Target {
					case "process-instance":
						m.selectedDefinitionKey = val
					case "process-variables":
						m.selectedInstanceID = val
					}

					// breadcrumb label: use configured label or target name
					label := chosen.Label
					if label == "" {
						label = chosen.Target
					}

					m.currentRoot = chosen.Target
					m.viewMode = chosen.Target
					m.genericParams = map[string]string{chosen.Param: val}
					m.breadcrumb = append(m.breadcrumb, label)

					// Build content header: use title_attribute if configured, else fallback to param value
					titleVal := val
					if chosen.TitleAttribute != "" && cursor >= 0 && cursor < len(m.rowData) {
						if tv, ok := m.rowData[cursor][chosen.TitleAttribute]; ok && tv != nil && fmt.Sprintf("%v", tv) != "" {
							titleVal = fmt.Sprintf("%v", tv)
						}
					}
					m.contentHeader = fmt.Sprintf("%s — %s", chosen.Target, titleVal)
					m.table.SetCursor(0)

					// Pre-set columns for target table to avoid stale columns during load
					colsTarget := m.buildColumnsFor(chosen.Target, m.paneWidth-4)
					if len(colsTarget) > 0 {
						m.table.SetRows(normalizeRows(nil, len(colsTarget)))
						m.table.SetColumns(colsTarget)
					} else {
						m.table.SetRows([]table.Row{})
					}
					return m, tea.Batch(m.fetchGenericCmd(chosen.Target), flashOnCmd(), m.saveStateCmd())
				}
			}

		case "tab":
			if m.popup.mode != popupModeNone && m.popup.mode != popupModeSearch {
				// compute matching contexts
				matchingContexts := m.popupItems()
				// Only complete if user has typed something or explicitly moved cursor
				if m.popup.cursor >= 0 && m.popup.cursor < len(matchingContexts) && len(m.popup.input) > 0 {
					m.popup.input = matchingContexts[m.popup.cursor]
				} else if len(m.popup.input) > 0 && len(matchingContexts) > 0 {
					m.popup.input = matchingContexts[0]
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
			if m.popup.mode != popupModeNone {
				if len(m.popup.input) > 0 {
					m.popup.input = m.popup.input[:len(m.popup.input)-1]
					m.popup.cursor = -1
				}
				if m.popup.mode == popupModeSearch {
					m.applySearchFromPopup()
				}
				m.paneHeight = m.computePaneHeight()
				m.table.SetHeight(m.paneHeight - 1)
				return m, nil
			}
			return m, nil
		case "G":
			// Vim jump to bottom (only in vim mode)
			if m.vimMode && m.popup.mode == popupModeNone && m.activeModal == ModalNone {
				rows := m.table.Rows()
				if len(rows) > 0 {
					m.table.SetCursor(len(rows) - 1)
				}
				return m, nil
			}
			if m.popup.mode != popupModeNone {
				m.popup.input += s
			}
			return m, nil
		case "g":
			// Vim gg sequence: only active in vim mode
			if m.popup.mode != popupModeNone {
				m.popup.input += s
				return m, nil
			}
			if m.vimMode && m.activeModal == ModalNone {
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
			// Half-page down (vim mode only) OR delete (always available for instance views)
			if m.viewMode == "process-instance" {
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
			// Half-page down only in vim mode
			if m.vimMode {
				pageSize := m.getPageSize()
				if pageSize <= 0 {
					pageSize = 10
				}
				m.table.MoveDown(pageSize / 2)
			}
			return m, nil
		case "up":
			if m.popup.mode == popupModeSkin {
				items := m.skinPopupItems()
				if m.popup.cursor > 0 {
					m.popup.cursor--
				} else if m.popup.cursor < 0 {
					m.popup.cursor = 0
				}
				// scroll up if cursor moved above visible window
				if m.popup.cursor < m.popup.offset {
					m.popup.offset = m.popup.cursor
				}
				if m.popup.cursor >= 0 && m.popup.cursor < len(items) {
					m.previewSkinByName(items[m.popup.cursor])
				}
				return m, nil
			}
			if m.popup.mode != popupModeNone {
				if m.popup.cursor > 0 {
					m.popup.cursor--
				} else if m.popup.cursor < 0 {
					m.popup.cursor = 0
				}
				// scroll up if cursor moved above visible window
				if m.popup.cursor < m.popup.offset {
					m.popup.offset = m.popup.cursor
				}
				if m.popup.mode == popupModeSearch {
					m.applySearchFromPopup()
				}
				return m, nil
			}
			// fall through to table navigation via component update
		case "down":
			if m.popup.mode == popupModeSkin {
				items := m.skinPopupItems()
				if len(items) > 0 {
					if m.popup.cursor < 0 {
						m.popup.cursor = 0
					} else if m.popup.cursor < len(items)-1 {
						m.popup.cursor++
					}
				}
				// scroll down if cursor moved past visible window
				const maxShow = 8
				if m.popup.cursor >= m.popup.offset+maxShow {
					m.popup.offset = m.popup.cursor - maxShow + 1
				}
				if m.popup.cursor >= 0 && m.popup.cursor < len(items) {
					m.previewSkinByName(items[m.popup.cursor])
				}
				return m, nil
			}
			if m.popup.mode != popupModeNone && m.popup.mode != popupModeSkin {
				matchingContexts := m.popupItems()
				if len(matchingContexts) > 0 {
					if m.popup.cursor < 0 {
						m.popup.cursor = 0
					} else if m.popup.cursor < len(matchingContexts)-1 {
						m.popup.cursor++
					}
				}
				// scroll down if cursor moved past visible window
				const maxShow = 8
				if m.popup.cursor >= m.popup.offset+maxShow {
					m.popup.offset = m.popup.cursor - maxShow + 1
				}
				if m.popup.mode == popupModeSearch {
					m.applySearchFromPopup()
				}
				return m, nil
			}
			// fall through to table navigation via component update
		case "ctrl+u":
			// Vim half-page up (vim mode only)
			if m.vimMode {
				pageSize := m.getPageSize()
				if pageSize <= 0 {
					pageSize = 10
				}
				m.table.MoveUp(pageSize / 2)
			}
			return m, nil
		case "ctrl+a":
			// Server-side search all pages (only active when search term is set)
			if m.searchTerm != "" {
				def := m.findTableDef(m.currentRoot)
				if def != nil && def.SearchParam != "" {
					// Fetch with server-side search param
					m.isLoading = true
					m.apiCallStarted = time.Now()
					return m, tea.Batch(
						m.fetchGenericWithParamCmd(m.currentRoot, def.SearchParam, m.searchTerm),
						spinnerTickCmd(),
					)
				}
				// No search_param configured — show feedback
				msg2, kind, cmd := setFooterStatus(footerStatusInfo, "Server-side search not available for this resource", 4*time.Second)
				m.footerError = msg2
				m.footerStatusKind = kind
				return m, cmd
			}
			return m, nil
		case "1", "2", "3", "4":
			// numeric breadcrumb navigation when popup not active
			if m.popup.mode == popupModeNone {
				n := int(s[0] - '1') // 0-based
				if n < len(m.breadcrumb) {
					return m, (&m).navigateToBreadcrumb(n)
				}
			}
			return m, nil
		case "j":
			// Vim down: navigate table rows (vim mode only, not in popup/modal/search)
			if m.vimMode && m.popup.mode == popupModeNone && m.activeModal == ModalNone && !m.searchMode {
				m.table.MoveDown(1)
				return m, nil
			}
			// In search mode: type into search input (handled by default case)
			if m.searchMode || m.popup.mode != popupModeNone {
				m.popup.input += s
				if m.popup.mode == popupModeSearch {
					m.applySearchFromPopup()
				}
				return m, nil
			}
			return m, nil
		case "k":
			// Vim up: navigate table rows (vim mode only, not in popup/modal/search)
			if m.vimMode && m.popup.mode == popupModeNone && m.activeModal == ModalNone && !m.searchMode {
				m.table.MoveUp(1)
				return m, nil
			}
			if m.searchMode || m.popup.mode != popupModeNone {
				m.popup.input += s
				if m.popup.mode == popupModeSearch {
					m.applySearchFromPopup()
				}
				return m, nil
			}
			return m, nil
		case "home":
			// Jump to first row (always available, not in popup/modal)
			if m.popup.mode == popupModeNone && m.activeModal == ModalNone {
				m.table.SetCursor(0)
				return m, nil
			}
			return m, nil
		case "end":
			// Jump to last row (always available, not in popup/modal)
			if m.popup.mode == popupModeNone && m.activeModal == ModalNone {
				rows := m.table.Rows()
				if len(rows) > 0 {
					m.table.SetCursor(len(rows) - 1)
				}
				return m, nil
			}
			return m, nil
		default:
			// typing into the root input when popup active
			if m.popup.mode != popupModeNone {
				if len(s) == 1 {
					m.popup.input += s
					m.popup.cursor = -1
					if m.popup.mode == popupModeSearch {
						m.applySearchFromPopup()
					}
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
		if m.debugEnabled {
			log.Printf("[resize] %dx%d", width, height)
		}
		// store terminal size for View footer alignment
		m.lastWidth = width
		m.lastHeight = height
		// Reserve lines: compact header is 2 rows, context selection 1 line (when active), footer 1 line
		headerLines := 2 // compactHeader placed at 2 rows
		contextSelectionLines := 1
		footerLines := 1
		// content height = terminal height minus header, context box, and footer
		contentHeight := height - headerLines - contextSelectionLines - footerLines
		if contentHeight < 3 {
			contentHeight = 3
		}
		// full terminal width — no ghost left pane
		m.paneWidth = width
		if m.paneWidth < 10 {
			m.paneWidth = 10
		}
		m.paneHeight = contentHeight
		// table inner width should be pane inner width minus border(2) and padding(2)
		tableInner := m.paneWidth - 4
		if tableInner < 10 {
			tableInner = 10
		}
		m.table.SetWidth(tableInner)
		m.table.SetHeight(contentHeight - 1)
		m.applyStyle()
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
	case taskVariablesLoadedMsg:
		m.isLoading = false
		if !m.apiCallStarted.IsZero() {
			m.lastAPILatency = time.Since(m.apiCallStarted)
			m.apiCallStarted = time.Time{}
		}
		m.taskCompleteTaskID = msg.taskID
		m.taskCompleteTaskName = msg.taskName
		m.taskInputVars = msg.inputVars
		m.taskCompleteFields = m.buildTaskCompleteFields(msg.formVars, msg.inputVars)
		m.taskCompletePos = 0
		m.taskCompleteFocus = focusTaskField
		if len(m.taskCompleteFields) > 0 {
			m.taskCompleteFields[0].input.Focus()
		}
		m.footerError = ""
		m.footerStatusKind = footerStatusNone
		m.activeModal = ModalTaskComplete
		return m, nil
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
				val := ""
				var v interface{}
				var found bool
				if vv, ok := it[key]; ok {
					v = vv
					found = true
				} else if vv, ok := it[strings.ToLower(col.Title)]; ok {
					v = vv
					found = true
				} else if vv, ok := it[col.Title]; ok {
					v = vv
					found = true
				}
				if found {
					if v == nil {
						val = ""
					} else if s, ok := v.(string); ok {
						val = s
					} else {
						val = fmt.Sprintf("%v", v)
					}
				} else {
					val = ""
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
		colorized := colorizeRows(msg.root, normalized, cols, RowStyles{
			Running:   m.styles.RowRunning,
			Suspended: m.styles.RowSuspended,
			Failed:    m.styles.RowFailed,
			Ended:     m.styles.RowEnded,
		})
		m.setTableRowsSorted(colorized)
		if m.sortColumn >= 0 {
			m.applySortIndicatorToColumns()
		}
		// Track which resource is currently displayed.
		m.viewMode = msg.root
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
		m.clampCursorAfterRowRemoval()
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
		if msg.closeTaskDialog {
			m.closeTaskCompleteDialog()
		}
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
	case skinsLoadedMsg:
		m.availableSkins = msg.names
		if m.popup.mode == popupModeSkin {
			m.popup.items = m.availableSkins
		}
	case errMsg:
		// Clear stale table rows so old data doesn't persist alongside the error.
		// Exception: preserve rows when the task complete dialog is open (keeps context visible).
		if m.activeModal != ModalTaskComplete {
			m.table.SetRows([]table.Row{})
		}
		m.isLoading = false
		// Log error always (not just in debug mode) — stack trace only in debug mode
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		_ = os.MkdirAll("./debug", 0755)
		screenFile := fmt.Sprintf("./debug/screen-%s.txt", id)
		_ = os.WriteFile(screenFile, []byte(lastRenderedView), 0644)
		if m.debugEnabled {
			log.Printf("[error] %v | screen: debug/screen-%s.txt\n%s", msg.err, id, debug.Stack())
		} else {
			log.Printf("[error] %v | screen: debug/screen-%s.txt", msg.err, id)
		}
		errText := friendlyError(m.currentEnv, msg.err) + " — Ctrl+r to retry"
		msg2, kind, cmd := setFooterStatus(footerStatusError, errText, 8*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
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

// currentUsername returns the username for the active environment.
func (m *model) currentUsername() string {
	if m.config == nil {
		return ""
	}
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return ""
	}
	return env.Username
}

// closeTaskCompleteDialog resets all task completion state and closes the modal.
func (m *model) closeTaskCompleteDialog() {
	m.activeModal = ModalNone
	m.taskCompleteTaskID = ""
	m.taskCompleteTaskName = ""
	m.taskInputVars = nil
	m.taskCompleteFields = nil
	m.taskCompletePos = 0
	m.taskCompleteFocus = focusTaskField
}

// blurCurrentTaskField removes focus from the currently focused text input.
func (m *model) blurCurrentTaskField() {
	if len(m.taskCompleteFields) > 0 && m.taskCompletePos < len(m.taskCompleteFields) {
		m.taskCompleteFields[m.taskCompletePos].input.Blur()
	}
}

// focusCurrentTaskField sets focus on the current task field.
func (m *model) focusCurrentTaskField() {
	if len(m.taskCompleteFields) > 0 && m.taskCompletePos < len(m.taskCompleteFields) {
		m.taskCompleteFields[m.taskCompletePos].input.Focus()
	}
}

// taskCompleteTabForward advances focus: field → ... → Complete → Back → field.
func (m *model) taskCompleteTabForward() {
	switch m.taskCompleteFocus {
	case focusTaskField:
		if m.taskCompletePos < len(m.taskCompleteFields)-1 {
			m.blurCurrentTaskField()
			m.taskCompletePos++
			m.focusCurrentTaskField()
		} else {
			m.blurCurrentTaskField()
			m.taskCompleteFocus = focusTaskComplete
		}
	case focusTaskComplete:
		m.taskCompleteFocus = focusTaskBack
	case focusTaskBack:
		m.taskCompleteFocus = focusTaskField
		m.taskCompletePos = 0
		m.focusCurrentTaskField()
	}
}

// taskCompleteTabBackward reverses the Tab cycle.
func (m *model) taskCompleteTabBackward() {
	switch m.taskCompleteFocus {
	case focusTaskField:
		if m.taskCompletePos > 0 {
			m.blurCurrentTaskField()
			m.taskCompletePos--
			m.focusCurrentTaskField()
		} else {
			m.blurCurrentTaskField()
			m.taskCompleteFocus = focusTaskBack
		}
	case focusTaskComplete:
		m.taskCompleteFocus = focusTaskField
		if len(m.taskCompleteFields) > 0 {
			m.taskCompletePos = len(m.taskCompleteFields) - 1
		}
		m.focusCurrentTaskField()
	case focusTaskBack:
		m.taskCompleteFocus = focusTaskComplete
	}
}

// validateTaskFieldValue validates a task field's current input value.
// Returns empty string when valid, or an error message.
func (m *model) validateTaskFieldValue(f taskCompleteField) string {
	val := f.input.Value()
	if val == "" {
		return "" // empty is allowed (will submit as empty string)
	}
	_, err := validation.ValidateAndParse(val, f.varType)
	if err != nil {
		return err.Error()
	}
	return ""
}

// taskCompleteHasErrors returns true if any field has a validation error.
func (m *model) taskCompleteHasErrors() bool {
	for _, f := range m.taskCompleteFields {
		if f.error != "" {
			return true
		}
		// live-validate each field
		if err := m.validateTaskFieldValue(f); err != "" {
			return true
		}
	}
	return false
}

// submitTaskComplete builds the variable map and dispatches completeTaskCmd.
// Returns nil if validation fails (Complete button should be disabled).
func (m *model) submitTaskComplete() tea.Cmd {
	if m.taskCompleteHasErrors() {
		return nil
	}
	vars := make(map[string]operaton.VariableValueDto, len(m.taskCompleteFields))
	for _, f := range m.taskCompleteFields {
		parsedVal, _ := validation.ValidateAndParse(f.input.Value(), f.varType)
		v := operaton.VariableValueDto{}
		v.SetValue(parsedVal)
		v.SetType(f.origType)
		vars[f.name] = v
	}
	m.isLoading = true
	m.apiCallStarted = time.Now()
	return tea.Batch(m.completeTaskCmd(m.taskCompleteTaskID, m.taskCompleteTaskName, vars), spinnerTickCmd())
}

// mapAPITypeToVarType maps an Operaton API type name to the internal validation type string.
func mapAPITypeToVarType(apiType string) string {
	switch strings.ToLower(apiType) {
	case "boolean":
		return "bool"
	case "integer", "long", "short":
		return "int"
	case "double", "float":
		return "float"
	case "json", "object":
		return "json"
	default:
		return "text"
	}
}

// buildTaskCompleteFields constructs the slice of editable form fields from the
// form variable definitions, pre-filling values from matching input variables.
func (m *model) buildTaskCompleteFields(formVars, inputVars map[string]variableValue) []taskCompleteField {
	names := make([]string, 0, len(formVars))
	for name := range formVars {
		names = append(names, name)
	}
	sort.Strings(names)

	fields := make([]taskCompleteField, 0, len(names))
	for _, name := range names {
		fv := formVars[name]
		origType := fv.TypeName
		varType := mapAPITypeToVarType(origType)

		ti := textinput.New()
		// Pre-fill: input variable with same name takes priority, then form var default value
		prefill := ""
		if iv, ok := inputVars[name]; ok && iv.Value != nil {
			prefill = fmt.Sprintf("%v", iv.Value)
		} else if fv.Value != nil {
			prefill = fmt.Sprintf("%v", fv.Value)
		}
		ti.SetValue(prefill)

		fields = append(fields, taskCompleteField{
			name:     name,
			varType:  varType,
			origType: origType,
			input:    ti,
		})
	}
	return fields
}
