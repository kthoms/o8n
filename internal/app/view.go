package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/kthoms/o8n/internal/contentassist"
	"github.com/kthoms/o8n/internal/validation"
)

// lastRenderedView holds the most recently rendered frame, used for screen dumps on errors.
// Package-level because View() has a value receiver and cannot mutate model fields.
var lastRenderedView string

// getKeyHints returns context-aware keyboard hints based on current view and terminal width
func (m *model) getKeyHints(width int) []KeyHint {
	hints := []KeyHint{}

	// Global hints (always relevant)
	hints = append(hints,
		KeyHint{"?", "help", 1},
		KeyHint{":", "switch", 2},
	)

	// Context-specific hints based on viewMode
	if m.viewMode == "process-definition" {
		hints = append(hints,
			KeyHint{"↑↓", "nav", 3},
			KeyHint{"Enter", "drill", 4},
		)
		// Drill-down drilldown hint
		if width >= 85 {
			hints = append(hints, KeyHint{"e", "Edit def", 4})
		}
	} else if m.viewMode == "process-instance" {
		hints = append(hints,
			KeyHint{"Esc", "back", 5},
			KeyHint{"↑↓", "nav", 3},
			KeyHint{"Enter", "vars", 4},
		)
		// Terminate instance hint in instances view
		if width >= 100 {
			hints = append(hints, KeyHint{"Ctrl+d", "terminate", 7})
		}
	} else if m.viewMode == "process-variables" {
		hints = append(hints,
			KeyHint{"Esc", "back", 5},
			KeyHint{"↑↓", "nav", 3},
		)
		// Show edit hint for variables when columns are editable
		if m.hasEditableColumns() {
			hints = append(hints, KeyHint{"e", "edit var", 4})
		}
	}

	// Add other hints based on width thresholds
	if width >= 88 && m.activeModal == ModalNone {
		hints = append(hints, KeyHint{"s", "sort", 5})
	}
	if width >= 100 && m.activeModal == ModalNone {
		hints = append(hints, KeyHint{"Space", "actions", 6})
	}
	if width >= 112 && m.activeModal == ModalNone {
		hints = append(hints, KeyHint{"y", "detail", 6})
	}
	if width >= 90 {
		hints = append(hints, KeyHint{"Ctrl+r", "refresh", 6})
	}
	// Latency hint intentionally removed from header hints — keep toggle available via help/README
	hints = append(hints, KeyHint{"PgDn/PgUp", "page", 3})
	if width >= 110 {
		hints = append(hints, KeyHint{"Ctrl+c", "quit", 8})
	}
	if width >= 105 {
		hints = append(hints, KeyHint{"Ctrl+e", "env", 9})
	}

	return hints
}

// renderCompactHeader renders a 2-row header
func (m *model) renderCompactHeader(width int) string {
	if width <= 0 {
		width = 80
	}

	// Row 1: Status line (environment with status indicator instead of URL/user)
	status, ok := m.envStatus[m.currentEnv]
	if !ok {
		status = StatusUnknown
	}

	// Determine status symbol and color
	statusSymbol := "●"
	statusColor := col(m.skin, "warning")
	switch status {
	case StatusOperational:
		statusSymbol = "●"
		statusColor = col(m.skin, "success")
	case StatusUnreachable:
		statusSymbol = "✗"
		statusColor = col(m.skin, "danger")
	case StatusUnknown:
		statusSymbol = "○"
		statusColor = col(m.skin, "warning")
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	envInfo := fmt.Sprintf("%s %s", m.currentEnv, statusStyle.Render(statusSymbol))

	row1 := fmt.Sprintf("o8n %s │ %s", m.version, envInfo)
	if m.autoRefresh {
		badge := m.styles.Accent.Render("↺")
		row1 = row1 + " " + badge
	}
	if lipgloss.Width(row1) > width-4 {
		row1 = truncateString(row1, width-7) + "..."
	}

	// Row 2: Key hints (priority-based)
	hints := m.getKeyHints(width)
	row2Parts := []string{}
	for _, hint := range hints {
		part := fmt.Sprintf("%s %s", hint.Key, hint.Description)
		row2Parts = append(row2Parts, part)
	}
	row2 := strings.Join(row2Parts, "  ")
	if lipgloss.Width(row2) > width-4 {
		row2 = truncateString(row2, width-7) + "..."
	}

	// Join rows
	header := fmt.Sprintf("%s\n%s", row1, row2)

	// Render header using default terminal colors (no forced background/foreground).
	headerStyle := lipgloss.NewStyle().Width(width).Padding(0, 1).Bold(true)
	return headerStyle.Render(header)
}

// renderConfirmDeleteModal renders a modal for confirming delete action
func (m *model) renderConfirmDeleteModal(width, height int) string {
	if m.pendingDeleteID == "" && m.pendingAction == nil {
		return ""
	}
	resourceLabel := strings.ToUpper(m.currentRoot)

	nameDetail := ""
	if m.pendingDeleteLabel != "" {
		nameDetail = fmt.Sprintf("Name:          %s\n", m.pendingDeleteLabel)
	}

	message := fmt.Sprintf(
		"⚠️  DELETE %s\n\n"+
			"You are about to DELETE this item:\n\n"+
			"ID:            %s\n"+
			"%s\n"+
			"⚠️  WARNING: This action CANNOT be undone!",
		resourceLabel, m.pendingDeleteID, nameDetail)

	confirmBtn := m.styles.BtnSave.Render(" Delete ")
	cancelBtn := m.styles.BtnCancelFocused.Render(" Cancel ")
	if m.confirmFocusedBtn == 0 {
		confirmBtn = m.styles.BtnSaveFocused.Render(" Delete ")
		cancelBtn = m.styles.BtnCancel.Render(" Cancel ")
	}
	buttons := confirmBtn + "  " + cancelBtn
	hint := m.styles.FgMuted.Render("Tab: switch  Enter: activate  Ctrl+d: delete  Esc: cancel")

	modalContent := message + "\n\n" + buttons + "\n" + hint

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(col(m.skin, "borderFocus")).
		Padding(1, 2).
		Width(54)

	// Return just the styled box — overlayCenter handles centering
	return modalStyle.Render(modalContent)
}

// renderConfirmQuitModal renders a modal asking the user to confirm quitting.
// Returns just the styled box; View() wraps it with overlayCenter.
func (m *model) renderConfirmQuitModal(_, _ int) string {
	confirmBtn := m.styles.BtnSave.Render(" Quit ")
	cancelBtn := m.styles.BtnCancelFocused.Render(" Cancel ")
	if m.confirmFocusedBtn == 0 {
		confirmBtn = m.styles.BtnSaveFocused.Render(" Quit ")
		cancelBtn = m.styles.BtnCancel.Render(" Cancel ")
	}
	buttons := confirmBtn + "  " + cancelBtn
	hint := m.styles.FgMuted.Render("Tab: switch  Enter: activate  Ctrl+c: quit  Esc: cancel")

	modalContent := "Quit o8n?\n\n" + buttons + "\n" + hint
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(col(m.skin, "borderFocus")).
		Padding(1, 2).
		Width(44)
	return modalStyle.Render(modalContent)
}

// renderHelpContentForLineCount returns the static help text for line-count purposes (scroll bound computation).
func renderHelpContentForLineCount(viewMode, currentEnv string) string {
	return `o8n Help

NAVIGATION               │  ACTIONS                │  GLOBAL
──────────────────────   │  ────────────────────   │  ──────────────────
↑/↓      Navigate list   │  Ctrl+e  Switch env     │  ?     This help
PgUp/Dn  Page up/down    │  Ctrl+r  Auto-refresh   │  :     Switch view
Home/End First/last row  │  Space   Actions menu   │  Ctrl+c Quit
Enter    Drill down      │
→        Drill down      │
Esc      Go back         │

SEARCH                   │  CONTEXT
──────────────────────   │  ──────────────────────
/        Search/filter   │  Tab    Complete
Esc      Clear filter    │  Enter  Confirm
Enter    Lock filter     │  Esc    Cancel
Ctrl+a   Search all pgs  │  s      Sort
                         │  y      Detail view

STATUS INDICATORS
────────────────────────────────────────────
● Running    ● Suspended    ✗ Failed/Incident    ○ Ended

Current View: ` + viewMode + `
Environment: ` + currentEnv + `

↑/↓: scroll  Any other key: close`
}

// renderHelpScreen renders the help screen modal
func (m model) renderHelpScreen(width, height int) string {
	// Build dynamic help content with resource-specific actions
	vimSection := ""
	if m.vimMode {
		vimSection = `
VIM NAVIGATION (vim mode active)
──────────────────────────────────
j/k      Navigate rows
gg/G     Top/bottom
Ctrl+u   Half-page up
Ctrl+d   Half-page dn`
	}

	// Resource-specific actions section
	resourceActionsSection := ""
	if def := m.findTableDef(m.currentRoot); def != nil && len(def.Actions) > 0 {
		var actionLines []string
		for _, act := range def.Actions {
			key := act.Key
			label := act.Label
			actionLines = append(actionLines, fmt.Sprintf("%-10s %s", key, label))
		}
		resourceActionsSection = "\nRESOURCE ACTIONS (" + m.currentRoot + ")\n" +
			"──────────────────────────────────\n" +
			strings.Join(actionLines, "\n")
	}

	helpContent := `o8n Help

NAVIGATION               │  ACTIONS                │  GLOBAL
──────────────────────   │  ────────────────────   │  ──────────────────
↑/↓      Navigate list   │  Ctrl+e  Switch env     │  ?     This help
PgUp/Dn  Page up/down    │  Ctrl+r  Auto-refresh   │  :     Switch view
Home/End First/last row  │  Space   Actions menu   │  Ctrl+c Quit
Enter    Drill down      │  Ctrl+a  Search all pg  │
→        Drill down      │                         │
Esc      Go back         │  SEARCH                 │  CONTEXT
                         │  ────────────────────   │  ──────────────────
                         │  /       Search/filter  │  Tab    Complete
                         │  Esc     Clear filter   │  Enter  Confirm
                         │  Enter   Lock filter    │  Esc    Cancel
                         │                         │  s      Sort
                         │                         │  y      Detail view` +
		vimSection + resourceActionsSection + `

STATUS INDICATORS
────────────────────────────────────────────
● Running    ● Suspended    ✗ Failed/Incident    ○ Ended
(symbol + position differentiate, not color alone)

Current View: ` + m.currentRoot + `
Environment: ` + m.currentEnv + `

↑/↓: scroll  Any other key: close`

	// Apply scroll window
	lines := strings.Split(helpContent, "\n")
	visibleH := height - 8
	if visibleH < 5 {
		visibleH = 5
	}
	maxScroll := len(lines) - visibleH
	if maxScroll < 0 {
		maxScroll = 0
	}
	scroll := m.helpScroll
	if scroll > maxScroll {
		scroll = maxScroll
	}
	start := scroll
	end := start + visibleH
	if end > len(lines) {
		end = len(lines)
	}
	visibleLines := make([]string, end-start)
	copy(visibleLines, lines[start:end])
	if scroll > 0 {
		visibleLines[0] = "  ↑ more above"
	}
	if end < len(lines) {
		visibleLines = append(visibleLines, "  ↓ more below")
	}
	helpContent = strings.Join(visibleLines, "\n")

	// Get color for styling

	helpWidth := width - 6
	if helpWidth > 76 {
		helpWidth = 76
	}
	if helpWidth < 40 {
		helpWidth = 40
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(col(m.skin, "borderFocus")).
		Padding(1, 2).
		Width(helpWidth)

	modal := modalStyle.Render(helpContent)

	return modal
}

func (m *model) renderEditModal(width, height int) string {
	row := m.currentEditRow()
	editCol := m.currentEditColumn()
	if row == nil || editCol == nil {
		return ""
	}
	inputType, _ := m.resolveEditTypes(editCol.def, m.editTableKey, row)

	singleColumnHeader := ""
	if len(m.editColumns) == 1 {
		typeLabel := inputType
		if typeLabel == "" {
			typeLabel = "text"
		}
		singleColumnHeader = fmt.Sprintf("Editing: %s  (type: %s)\n%s\n\n",
			m.editColumns[0].def.Name,
			typeLabel,
			strings.Repeat("─", 36))
	}

	columnsLine := ""
	if len(m.editColumns) > 1 {
		parts := make([]string, 0, len(m.editColumns))
		for i, c := range m.editColumns {
			label := fmt.Sprintf("%d:%s", i+1, c.def.Name)
			if i == m.editColumnPos {
				label = "[" + label + "]"
			}
			parts = append(parts, label)
		}
		columnsLine = strings.Join(parts, " ") + "\n\n"
	}

	errorLine := ""
	if m.editError != "" {
		errorLine = m.styles.ValidationError.Render("⚠ "+m.editError) + "\n\n"
	}

	// Build body (without header)
	body := columnsLine + m.editInput.View() + "\n\n" + errorLine

	// Determine button styles from config or defaults
	saveStyle := m.styles.BtnSave
	cancelStyle := m.styles.BtnCancel
	disabledSaveStyle := m.styles.BtnDisabled

	saveLabel := " Save "
	cancelLabel := " Cancel "
	if m.config != nil && m.config.UI != nil && m.config.UI.EditModal != nil && m.config.UI.EditModal.Buttons != nil {
		if b := m.config.UI.EditModal.Buttons.Save; b != nil && b.Label != "" {
			saveLabel = " " + b.Label + " "
		}
		if b := m.config.UI.EditModal.Buttons.Cancel; b != nil && b.Label != "" {
			cancelLabel = " " + b.Label + " "
		}
	}

	// Proactive validation: validate current input to determine if Save should be enabled
	saveDisabled := false
	if inputType != "" {
		if _, err := validation.ValidateAndParse(m.editInput.Value(), inputType); err != nil {
			saveDisabled = true
			// if no explicit editError set, show the validation error inline
			if m.editError == "" {
				errorLine = m.styles.ValidationError.Render("⚠ "+err.Error()) + "\n\n"
			}
		}
	}

	// Rebuild body to include any validation errorLine that may have been set above
	// If input type is `user` include content-assist suggestions from the static cache.
	suggestionLine := ""
	if inputType == "user" {
		sugg := contentassist.SuggestUsers(m.editInput.Value())
		if len(sugg) > 0 {
			suggestionLine = "Suggestions: " + strings.Join(sugg, ", ") + "\n\n"
		}
	}
	body = singleColumnHeader + columnsLine + m.editInput.View() + "\n\n" + errorLine + suggestionLine

	// Render buttons with focus styles
	savedFocusedStyle := m.styles.BtnSaveFocused
	cancelFocusedStyle := m.styles.BtnCancelFocused

	var saveBtn, cancelBtn string
	switch m.editFocus {
	case editFocusSave:
		if saveDisabled {
			saveBtn = disabledSaveStyle.Copy().
				Border(lipgloss.NormalBorder()).
				BorderForeground(col(m.skin, "danger")).
				Render(saveLabel)
		} else {
			saveBtn = savedFocusedStyle.Render(saveLabel)
		}
		cancelBtn = cancelStyle.Render(cancelLabel)
	case editFocusCancel:
		if saveDisabled {
			saveBtn = disabledSaveStyle.Render(saveLabel)
		} else {
			saveBtn = saveStyle.Render(saveLabel)
		}
		cancelBtn = cancelFocusedStyle.Render(cancelLabel)
	default: // editFocusInput
		if saveDisabled {
			saveBtn = disabledSaveStyle.Render(saveLabel)
		} else {
			saveBtn = saveStyle.Render(saveLabel)
		}
		cancelBtn = cancelStyle.Render(cancelLabel)
	}
	buttons := saveBtn + "  " + cancelBtn

	modalBody := body + "\n" + buttons

	modalBorderColor := col(m.skin, "borderFocus")
	if m.config != nil && m.config.UI != nil && m.config.UI.EditModal != nil && m.config.UI.EditModal.BorderColor != "" {
		modalBorderColor = lipgloss.Color(m.config.UI.EditModal.BorderColor)
	}

	modalStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(modalBorderColor).Padding(1, 2)
	if m.config != nil && m.config.UI != nil && m.config.UI.EditModal != nil && m.config.UI.EditModal.Width > 0 {
		modalStyle = modalStyle.Width(m.config.UI.EditModal.Width)
	} else {
		modalStyle = modalStyle.Width(60)
	}

	modal := modalStyle.Render(modalBody)
	// Return just the styled box — overlayCenter handles centering
	return modal
}

func (m model) asciiArt() string {
	return "   ____\n____  ( __ )\n / __ \\/ __  / __ \\\n/ /_/ / /_/ / / / /\n\\____/\\____/_/ /_/\n" + "o8n"
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	// If splash active, render animated splash centered
	if m.splashActive {
		logo := m.asciiArt()
		lines := strings.Split(logo, "\n")
		totalLines := len(lines)
		if totalLines == 0 {
			totalLines = 1
		}
		// determine how many lines to reveal based on current frame
		show := (m.splashFrame * totalLines) / totalSplashFrames
		if show < 1 {
			show = 1
		}
		if show > totalLines {
			show = totalLines
		}
		displayed := strings.Join(lines[:show], "\n")
		logoRendered := m.splashLogoStyle.Render(displayed)

		// animate info: fade in during last half of frames by showing once frame > half
		info := "v" + m.version
		infoRendered := ""
		if m.splashFrame >= totalSplashFrames/2 {
			infoRendered = m.splashInfoStyle.Render(info)
		} else {
			// slight spacer to keep vertical centering stable
			infoRendered = ""
		}

		content := lipgloss.JoinVertical(lipgloss.Center, logoRendered, infoRendered)
		// ensure we use full terminal size
		w := m.lastWidth
		h := m.lastHeight
		if w <= 0 {
			w = 80
		}
		if h <= 0 {
			h = 24
		}
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
	}

	// Main UI - use compact 3-row header
	compactHeader := m.renderCompactHeader(m.lastWidth)
	// Ensure compact header occupies exactly 3 rows
	compactHeader = lipgloss.Place(m.lastWidth, 2, lipgloss.Left, lipgloss.Center, compactHeader)

	// get border color

	// render context selection box (generic popup palette)
	var contextSelectionBox string
	if m.popup.mode != popupModeNone {
		var items []string
		var hint string
		if m.popup.mode == popupModeSkin {
			items = m.skinPopupItems()
			hint = m.popup.hint
		} else if m.popup.mode == popupModeSearch {
			// Search mode: list shows matching rows (first-column values)
			for _, row := range m.table.Rows() {
				if len(row) > 0 {
					items = append(items, fmt.Sprintf("%v", row[0]))
				}
			}
			hint = "↑↓:select  Enter:jump  Esc:cancel"
			// Show Ctrl+A hint when there are more results beyond the current page
			if total, ok := m.pageTotals[m.currentRoot]; ok {
				pageSize := m.getPageSize()
				if pageSize <= 0 {
					pageSize = 1
				}
				if total > pageSize {
					hint = hint + "  Ctrl+A:search all pages"
				}
			}
		} else {
			// context mode
			for _, rc := range m.rootContexts {
				if m.popup.input == "" || strings.HasPrefix(rc, m.popup.input) {
					items = append(items, rc)
				}
			}
			hint = "↑↓:select  Tab/Enter:switch  Esc:cancel"
		}

		completion := ""
		if m.popup.mode != popupModeSearch && m.popup.input != "" && len(items) > 0 && items[0] != m.popup.input {
			completion = items[0][len(m.popup.input):]
		}

		var inputLabel string
		if m.popup.mode == popupModeSearch {
			inputLabel = "/"
		}
		displayText := m.styles.PopupInput.Render(inputLabel+m.popup.input) + m.styles.PopupCompletion.Render(completion)
		hintLine := m.styles.PopupHint.Render(hint)

		// Scrollable list: show a window of maxShow items around the cursor
		const maxShow = 8
		offset := m.popup.offset
		if offset < 0 {
			offset = 0
		}
		if len(items) > 0 && offset > len(items)-1 {
			offset = len(items) - 1
		}
		end := offset + maxShow
		if end > len(items) {
			end = len(items)
		}
		shown := items[offset:end]
		selectedStyle := lipgloss.NewStyle().Foreground(col(m.skin, "borderFocus")).Bold(true)
		var listLines []string
		for i, rc := range shown {
			globalIdx := offset + i
			cursor := "  "
			if globalIdx == m.popup.cursor {
				cursor = "\u25b8 "
				listLines = append(listLines, selectedStyle.Render(fmt.Sprintf("%s%s", cursor, rc)))
			} else {
				listLines = append(listLines, fmt.Sprintf("%s%s", cursor, rc))
			}
		}
		listText := strings.Join(listLines, "\n")

		fullContent := displayText + "\n" + hintLine
		if len(listLines) > 0 {
			fullContent = fullContent + "\n" + listText
		}

		boxStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(col(m.skin, "borderFocus")).
			Width(m.lastWidth-4).
			Padding(0, 1)
		contextSelectionBox = boxStyle.Render(fullContent)
	} else {
		contextSelectionBox = ""
	}

	// combine compact header only; the content header will be rendered
	// embedded into the content border by the custom box renderer below.
	headerStack := compactHeader

	// Content box should use full terminal width
	pw := m.lastWidth
	if pw < 10 {
		pw = 10
	}

	// render the main content box with title embedded into top border
	// Append total count for current root if known
	baseTitle := m.contentHeader
	if m.searchTerm != "" {
		var matchCount int
		if m.filteredRows != nil {
			matchCount = len(m.filteredRows)
		} else {
			matchCount = len(m.table.Rows())
		}
		if total, ok := m.pageTotals[m.currentRoot]; ok && total >= 0 {
			baseTitle = fmt.Sprintf("%s [/%s/ — %d of %d]", m.contentHeader, m.searchTerm, matchCount, total)
		} else {
			baseTitle = fmt.Sprintf("%s [/%s/ — %d matches]", m.contentHeader, m.searchTerm, matchCount)
		}
	} else if total, ok := m.pageTotals[m.currentRoot]; ok && total >= 0 {
		baseTitle = fmt.Sprintf("%s — %d items", m.contentHeader, total)
	}
	title := baseTitle

	// Render search bar when in search mode
	searchBar := ""
	if m.searchMode {
		searchStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(col(m.skin, "borderFocus")).
			Width(pw-4).
			Padding(0, 1)
		searchBar = searchStyle.Render(m.searchInput.View())
	}

	// Build table content — show empty state message when no rows
	tableContent := m.table.View()
	if len(m.table.Rows()) == 0 && !m.isLoading {
		displayName := strings.ReplaceAll(m.currentRoot, "-", " ")
		emptyMsg := "No " + displayName + " found"
		if m.footerStatusKind == footerStatusError {
			emptyMsg = "Error loading data — press r to retry"
		} else if len(m.navigationStack) > 0 && m.selectedDefinitionKey != "" {
			emptyMsg = "No " + displayName + " for " + m.selectedDefinitionKey
		}
		emptyStyle := lipgloss.NewStyle().
			Foreground(col(m.skin, "fgMuted")).
			Width(m.paneWidth - 6).
			Align(lipgloss.Center)
		tableContent = emptyStyle.Render(emptyMsg)
	}
	mainBox := renderBoxWithTitle(tableContent, pw, m.paneHeight, title, m.styles.BorderFocus)

	// Footer (1 row with 3 columns per specification):
	// Column 1: Context tag with breadcrumb navigation hints
	// Column 2: Status message (error/success/loading/info)
	// Column 3: Remote activity indicator (⚡)
	// Columns separated by " | "

	crumbs := make([]string, 0, len(m.breadcrumb))
	lastIdx := len(m.breadcrumb) - 1
	for i, c := range m.breadcrumb {
		style := lipgloss.NewStyle()
		if i < len(m.breadcrumbStyles) {
			style = m.breadcrumbStyles[i]
		}
		if i == lastIdx {
			// Last crumb = current location; no hotkey needed
			crumbs = append(crumbs, style.Render(fmt.Sprintf("<%s>", c)))
		} else {
			// Navigable ancestors get a [n] hotkey
			hint := fmt.Sprintf("%d", i+1)
			crumbs = append(crumbs, style.Render(fmt.Sprintf("[%s] <%s>", hint, c)))
		}
	}
	breadcrumbRendered := strings.Join(crumbs, " ")
	// Cap breadcrumb at 50% terminal width using ANSI-aware clipping
	maxBreadcrumbW := m.lastWidth / 2
	if lipgloss.Width(breadcrumbRendered) > maxBreadcrumbW && maxBreadcrumbW > 0 {
		breadcrumbRendered = lipgloss.NewStyle().MaxWidth(maxBreadcrumbW).Inline(true).Render(breadcrumbRendered)
	}

	// Render remote flash as a fixed-width symbol on the right, plus latency and pagination if available
	remoteSymbol := " "
	rpStyle := m.styles.FlashBase
	if m.flashActive {
		remoteSymbol = "⚡"
		rpStyle = rpStyle.Foreground(col(m.skin, "success")).Bold(true)
	} else {
		remoteSymbol = " "
		rpStyle = rpStyle.Foreground(col(m.skin, "fgMuted"))
	}
	latencyStr := ""
	if m.showLatency && m.lastAPILatency > 0 {
		latencyStr = fmt.Sprintf(" %dms", m.lastAPILatency.Milliseconds())
	}
	// Add pagination status if available
	paginationStr := ""
	if m.currentRoot != "" {
		if total, ok := m.pageTotals[m.currentRoot]; ok && total > 0 {
			pageSize := m.getPageSize()
			currentPage := (m.pageOffsets[m.currentRoot] / pageSize) + 1
			totalPages := (total + pageSize - 1) / pageSize
			paginationStr = fmt.Sprintf(" [%d/%d]", currentPage, totalPages)
		}
	}
	pageIndicator := ""
	if paginationStr != "" {
		pageIndicator = m.styles.PageCounter.Render(paginationStr) + " "
	}
	rightPart := pageIndicator + rpStyle.Render(remoteSymbol+latencyStr)

	// Layout footer: [breadcrumb] | [status] | [remote]
	// Format: leftPart | middlePart | rightPart (all separated by " | ")
	totalW := m.lastWidth
	const sepWidth = 3 // width of " | "

	leftPart := breadcrumbRendered
	leftPartW := lipgloss.Width(leftPart)

	// Remote indicator always at the right (1 char + " | " before it = 4 total)
	remotePart := " | " + rightPart
	remotePartW := lipgloss.Width(remotePart)

	// Available width for middle column (status message)
	middleW := totalW - leftPartW - remotePartW - sepWidth
	if middleW < 0 {
		middleW = 0
	}

	// Build status message: truncate plain text first, then style
	statusMessage := ""
	if m.footerError != "" {
		var plainIcon string
		var style lipgloss.Style
		switch m.footerStatusKind {
		case footerStatusError:
			plainIcon = "✗ "
			style = m.styles.ErrorFooter
		case footerStatusSuccess:
			plainIcon = "✓ "
			style = m.styles.SuccessFooter
		case footerStatusLoading:
			plainIcon = spinnerFrames[m.spinnerFrame%len(spinnerFrames)] + " "
			style = m.styles.LoadingFooter
		default: // footerStatusInfo
			plainIcon = "ℹ "
			style = m.styles.InfoFooter
		}
		plainText := m.footerError
		iconWidth := lipgloss.Width(plainIcon)
		maxText := middleW - iconWidth
		if maxText < 0 {
			maxText = 0
		}
		if lipgloss.Width(plainText) > maxText {
			plainText = truncateString(plainText, maxText)
		}
		statusMessage = style.Render(plainIcon + plainText)
	}

	// Pad to fill the column
	padW := middleW - lipgloss.Width(statusMessage)
	if padW > 0 {
		statusMessage = statusMessage + strings.Repeat(" ", padW)
	}

	footerLine := leftPart + " | " + statusMessage + remotePart

	// Compose final vertical layout: header, context box, search bar, main content, footer (1 row)
	baseView := lipgloss.JoinVertical(lipgloss.Left, headerStack, contextSelectionBox, searchBar, mainBox, footerLine)

	// If modal is active, overlay it
	if m.activeModal == ModalConfirmDelete {
		overlay := m.renderConfirmDeleteModal(m.lastWidth, m.lastHeight)
		return overlayCenter(baseView, overlay)
	} else if m.activeModal == ModalEdit {
		overlay := m.renderEditModal(m.lastWidth, m.lastHeight)
		return overlayCenter(baseView, overlay)
	} else if m.activeModal == ModalHelp {
		overlay := m.renderHelpScreen(m.lastWidth, m.lastHeight)
		return overlayCenter(baseView, overlay)
	} else if m.activeModal == ModalSort {
		overlay := m.renderSortPopup(m.lastWidth, m.lastHeight)
		return overlayCenter(baseView, overlay)
	} else if m.activeModal == ModalDetailView {
		overlay := m.renderDetailView(m.lastWidth, m.lastHeight)
		return overlayCenter(baseView, overlay)
	} else if m.activeModal == ModalEnvironment {
		overlay := m.renderEnvPopup(m.lastWidth, m.lastHeight)
		return overlayCenter(baseView, overlay)
	} else if m.activeModal == ModalConfirmQuit {
		overlay := m.renderConfirmQuitModal(m.lastWidth, m.lastHeight)
		return overlayCenter(baseView, overlay)
	} else if m.activeModal == ModalTaskComplete {
		overlay := m.renderTaskCompleteModal(m.lastWidth, m.lastHeight)
		return overlayCenter(baseView, overlay)
	}

	// If actions menu is active, overlay it
	if m.showActionsMenu {
		overlay := m.renderActionsMenu(m.lastWidth, m.lastHeight)
		return overlayCenter(baseView, overlay)
	}

	// Ensure the main UI uses the full terminal area to avoid trailing space artifacts
	w := m.lastWidth
	h := m.lastHeight
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	// Capture last rendered frame for screen dumps (used by error/panic logging)
	lastRenderedView = baseView
	if m.debugEnabled {
		_ = os.MkdirAll("./debug", 0755)
		_ = os.WriteFile("./debug/last-screen.txt", []byte(baseView), 0644)
	}
	return lipgloss.Place(w, h, lipgloss.Left, lipgloss.Top, baseView)
}

// renderBoxWithTitle draws a simple single-line-border box of the given
// totalWidth/totalHeight and embeds title centered into the top border.
// The content is clipped/padded to fit the inner area. If color is non-empty
// the entire box text is colorized using that foreground color.
func renderBoxWithTitle(content string, totalWidth, totalHeight int, title string, borderStyle lipgloss.Style) string {
	innerWidth := totalWidth - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Determine available content height (exclude top/bottom border)
	contentHeight := totalHeight - 2
	if contentHeight < 0 {
		contentHeight = 0
	}

	// Split content into lines and prepare a fixed-height slice
	contentLines := strings.Split(content, "\n")
	lines := make([]string, contentHeight)
	for i := 0; i < contentHeight; i++ {
		var l string
		if i < len(contentLines) {
			l = contentLines[i]
		} else {
			l = ""
		}
		// Truncate or pad to innerWidth using rune-aware slicing
		if lipgloss.Width(l) > innerWidth {
			runes := []rune(l)
			if len(runes) > innerWidth {
				l = string(runes[:innerWidth])
			}
		}
		pad := innerWidth - lipgloss.Width(l)
		if pad < 0 {
			pad = 0
		}
		lines[i] = l + strings.Repeat(" ", pad)
	}

	// Prepare title, trimming if necessary
	t := title
	if lipgloss.Width(t) > innerWidth {
		runes := []rune(t)
		if len(runes) > innerWidth {
			t = string(runes[:innerWidth])
		}
	}
	left := (innerWidth - lipgloss.Width(t)) / 2
	if left < 0 {
		left = 0
	}
	right := innerWidth - left - lipgloss.Width(t)

	top := "┌" + strings.Repeat("─", left) + t + strings.Repeat("─", right) + "┐"
	bottom := "└" + strings.Repeat("─", innerWidth) + "┘"

	var b strings.Builder
	// Always apply the border style (callers pass empty style for no coloring)
	useBorderColor := true
	if useBorderColor {
		b.WriteString(borderStyle.Render(top))
	} else {
		b.WriteString(top)
	}
	b.WriteString("\n")
	for _, l := range lines {
		if useBorderColor {
			b.WriteString(borderStyle.Render("│"))
			b.WriteString(l)
			b.WriteString(borderStyle.Render("│"))
		} else {
			b.WriteString("│")
			b.WriteString(l)
			b.WriteString("│")
		}
		b.WriteString("\n")
	}
	if useBorderColor {
		b.WriteString(borderStyle.Render(bottom))
	} else {
		b.WriteString(bottom)
	}

	return b.String()
}

// renderSortPopup renders the column sort selection popup.
func (m *model) renderSortPopup(width, height int) string {
	cols := m.table.Columns()

	var b strings.Builder
	showClear := m.sortColumn >= 0
	if showClear {
		cursor := "  "
		if m.sortPopupCursor == -1 {
			cursor = "▸ "
		}
		b.WriteString(fmt.Sprintf("%s— clear sort —\n", cursor))
	}
	for i, tableCol := range cols {
		cursor := "  "
		if i == m.sortPopupCursor {
			cursor = "▸ "
		}
		indicator := "  "
		if i == m.sortColumn {
			if m.sortAscending {
				indicator = " ▲"
			} else {
				indicator = " ▼"
			}
		}
		b.WriteString(fmt.Sprintf("%s%s%s\n", cursor, tableCol.Title, indicator))
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(col(m.skin, "borderFocus")).
		Padding(0, 1)

	// Dynamic width: max(30, longestColumnName+8), capped at terminalWidth-10
	longestName := 14 // len("Sort by Column") as minimum content
	for _, tableCol := range cols {
		if len(tableCol.Title) > longestName {
			longestName = len(tableCol.Title)
		}
	}
	sortWidth := longestName + 8
	if sortWidth < 30 {
		sortWidth = 30
	}
	if m.lastWidth > 10 && sortWidth > m.lastWidth-10 {
		sortWidth = m.lastWidth - 10
	}
	modalStyle = modalStyle.Width(sortWidth)

	title := "Sort by Column"
	content := b.String()
	modal := modalStyle.Render(title + "\n" + content + "\nEnter: Select  Esc: Close")

	return modal
}

// renderDetailView renders the YAML/JSON detail viewer overlay.
func (m *model) renderDetailView(width, height int) string {

	// Calculate visible area
	viewHeight := height - 6
	if viewHeight < 3 {
		viewHeight = 3
	}

	lines := strings.Split(m.detailContent, "\n")
	maxScroll := len(lines) - viewHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.detailScroll > maxScroll {
		m.detailScroll = maxScroll
	}
	if m.detailScroll < 0 {
		m.detailScroll = 0
	}

	// Extract visible lines
	end := m.detailScroll + viewHeight
	if end > len(lines) {
		end = len(lines)
	}
	visibleLines := lines[m.detailScroll:end]

	// Add line numbers and syntax highlighting
	var b strings.Builder
	for i, line := range visibleLines {
		lineNum := m.detailScroll + i + 1
		b.WriteString(fmt.Sprintf("%4d │ %s\n", lineNum, m.syntaxHighlightJSON(line)))
	}

	scrollInfo := fmt.Sprintf("[%d/%d]", m.detailScroll+1, len(lines))

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(col(m.skin, "borderFocus")).
		Padding(1, 2).
		Width(width - 4).
		Height(viewHeight + 2)

	content := b.String()
	title := "Detail View  " + scrollInfo + "  (scroll ↑↓, q/Esc close)"
	modal := modalStyle.Render(title + "\n" + content)

	return modal
}

// syntaxHighlightJSON applies basic syntax highlighting to a JSON line.
func (m *model) syntaxHighlightJSON(line string) string {
	keyStyle := m.styles.JSONKey
	stringStyle := m.styles.JSONValue
	numberStyle := m.styles.JSONNumber
	boolStyle := m.styles.JSONBool

	trimmed := strings.TrimSpace(line)

	// Key-value detection: "key": value
	if idx := strings.Index(trimmed, ":"); idx > 0 {
		key := strings.TrimSpace(trimmed[:idx])
		if strings.HasPrefix(key, "\"") && strings.HasSuffix(key, "\"") {
			val := strings.TrimSpace(trimmed[idx+1:])
			val = strings.TrimSuffix(val, ",")
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]

			styledKey := keyStyle.Render(key)
			styledVal := val
			if strings.HasPrefix(val, "\"") {
				styledVal = stringStyle.Render(val)
			} else if val == "true" || val == "false" || val == "null" {
				styledVal = boolStyle.Render(val)
			} else if _, err := strconv.ParseFloat(val, 64); err == nil {
				styledVal = numberStyle.Render(val)
			}

			trailing := ""
			if strings.HasSuffix(strings.TrimSpace(trimmed[idx+1:]), ",") {
				trailing = ","
			}
			return indent + styledKey + ": " + styledVal + trailing
		}
	}
	return line
}

// renderEnvPopup renders the environment selection popup.
func (m *model) renderEnvPopup(width, height int) string {

	var b strings.Builder
	for i, name := range m.envNames {
		cursor := "  "
		if i == m.envPopupCursor {
			cursor = "▸ "
		}

		// Status indicator
		status, ok := m.envStatus[name]
		if !ok {
			status = StatusUnknown
		}
		statusIcon := "○"
		switch status {
		case StatusOperational:
			statusIcon = "●"
		case StatusUnreachable:
			statusIcon = "✗"
		}

		envStyle := m.styles.Accent

		url := ""
		if env, ok := m.config.Environments[name]; ok {
			url = env.URL
		}

		activeMarker := ""
		if name == m.currentEnv {
			activeMarker = m.styles.FgMuted.Render(" ✓")
		}
		line := fmt.Sprintf("%s%s %s%s  %s", cursor, statusIcon, envStyle.Render(name), activeMarker, url)
		b.WriteString(line + "\n")
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(col(m.skin, "borderFocus")).
		Padding(0, 1).
		Width(50)

	content := "Select Environment\n\n" + b.String() + "\nEnter: Select  Esc: Close"
	modal := modalStyle.Render(content)

	return modal
}

// renderActionsMenu renders the context actions menu popup.
func (m *model) renderActionsMenu(width, height int) string {

	var b strings.Builder
	root := m.currentRoot
	row := m.table.SelectedRow()
	rowLabel := ""
	if len(row) > 0 {
		rowLabel = "\n" + ansi.Strip(stripFocusIndicatorPrefix(row[0]))
	}
	b.WriteString(fmt.Sprintf("Actions: %s%s\n\n", root, rowLabel))
	for i, item := range m.actionsMenuItems {
		cursor := "  "
		if i == m.actionsMenuCursor {
			cursor = "▸ "
		}
		b.WriteString(fmt.Sprintf("%s[%s] %s\n", cursor, item.key, item.label))
	}
	b.WriteString("\nEsc: Close")

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(col(m.skin, "borderFocus")).
		Padding(0, 1).
		Width(35)

	modal := modalStyle.Render(b.String())
	return overlayCenter(lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, ""), modal)
}

// renderTaskCompleteModal renders the full-screen task completion dialog.
func (m *model) renderTaskCompleteModal(width, height int) string {
	const sep = "────────────────────────────────────"

	var sb strings.Builder

	// Title + subtitle
	sb.WriteString("Complete Task\n")
	sb.WriteString(m.styles.FgMuted.Render(m.taskCompleteTaskName) + "\n")
	sb.WriteString(sep + "\n\n")

	// INPUT VARIABLES section
	sb.WriteString("INPUT VARIABLES\n")
	if len(m.taskInputVars) == 0 {
		sb.WriteString(m.styles.FgMuted.Render("  No context variables") + "\n")
	} else {
		// Sort alphabetically
		names := make([]string, 0, len(m.taskInputVars))
		for n := range m.taskInputVars {
			names = append(names, n)
		}
		sortStrings(names)
		for _, n := range names {
			v := m.taskInputVars[n]
			val := ""
			if v.Value != nil {
				val = fmt.Sprintf("%v", v.Value)
			}
			sb.WriteString(m.styles.FgMuted.Render(fmt.Sprintf("  %s : %s", n, val)) + "\n")
		}
	}
	sb.WriteString("\n" + sep + "\n\n")

	// OUTPUT VARIABLES section
	sb.WriteString("OUTPUT VARIABLES\n")
	if len(m.taskCompleteFields) == 0 {
		sb.WriteString(m.styles.FgMuted.Render("  No output variables defined for this task") + "\n")
	} else {
		for i, f := range m.taskCompleteFields {
			focused := m.taskCompleteFocus == focusTaskField && m.taskCompletePos == i
			_ = focused
			inputView := f.input.View()
			sb.WriteString(fmt.Sprintf("  %s │ %s\n", f.name, inputView))
			if f.error != "" {
				sb.WriteString(m.styles.ValidationError.Render("  ⚠ "+f.error) + "\n")
			}
		}
	}
	sb.WriteString("\n" + sep + "\n\n")

	// Buttons
	completeDisabled := m.taskCompleteHasErrors()
	var completeBtn, backBtn string
	switch m.taskCompleteFocus {
	case focusTaskComplete:
		if completeDisabled {
			completeBtn = m.styles.BtnDisabled.Copy().
				Border(lipgloss.NormalBorder()).
				BorderForeground(col(m.skin, "danger")).
				Render("  Complete  ")
		} else {
			completeBtn = m.styles.BtnSaveFocused.Render("  Complete  ")
		}
		backBtn = m.styles.BtnCancel.Render("  Back  ")
	case focusTaskBack:
		if completeDisabled {
			completeBtn = m.styles.BtnDisabled.Render("  Complete  ")
		} else {
			completeBtn = m.styles.BtnSave.Render("  Complete  ")
		}
		backBtn = m.styles.BtnCancelFocused.Render("  Back  ")
	default:
		if completeDisabled {
			completeBtn = m.styles.BtnDisabled.Render("  Complete  ")
		} else {
			completeBtn = m.styles.BtnSave.Render("  Complete  ")
		}
		backBtn = m.styles.BtnCancel.Render("  Back  ")
	}
	sb.WriteString(completeBtn + "  " + backBtn + "\n\n")

	// Hint line
	sb.WriteString(m.styles.FgMuted.Render("Tab: next  Space: toggle bool  Esc: back"))

	// Wrap in modal style
	modalW := width - 8
	if modalW < 60 {
		modalW = 60
	}
	if modalW > 100 {
		modalW = 100
	}
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(col(m.skin, "borderFocus")).
		Padding(1, 2).
		Width(modalW)

	return modalStyle.Render(sb.String())
}

// sortStrings sorts a string slice in place (helper to avoid importing sort in view.go).
func sortStrings(ss []string) {
	for i := 1; i < len(ss); i++ {
		for j := i; j > 0 && ss[j] < ss[j-1]; j-- {
			ss[j], ss[j-1] = ss[j-1], ss[j]
		}
	}
}

// overlayCenter places the fg string centered over the bg string.
// The left side of each overlaid line preserves ANSI styling from bg.
// The right side uses plain text (stripped ANSI) — acceptable since
// the modal box covers the visually important center region.
func overlayCenter(bg, fg string) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")
	bgH, fgH := len(bgLines), len(fgLines)
	bgW := 0
	for _, l := range bgLines {
		if w := lipgloss.Width(l); w > bgW {
			bgW = w
		}
	}
	fgW := 0
	for _, l := range fgLines {
		if w := lipgloss.Width(l); w > fgW {
			fgW = w
		}
	}
	startRow := (bgH - fgH) / 2
	startCol := (bgW - fgW) / 2
	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}
	result := make([]string, bgH)
	copy(result, bgLines)
	for i, fgLine := range fgLines {
		row := startRow + i
		if row < 0 || row >= bgH {
			continue
		}
		bgLine := bgLines[row]
		left := ansi.Truncate(bgLine, startCol, "")
		leftW := lipgloss.Width(left)
		if leftW < startCol {
			left += strings.Repeat(" ", startCol-leftW)
		}
		// Right portion: use stripped bg text beyond the overlay area
		plain := ansi.Strip(bgLine)
		runes := []rune(plain)
		rightStart := startCol + lipgloss.Width(fgLine)
		right := ""
		if rightStart < len(runes) {
			right = string(runes[rightStart:])
		}
		result[row] = left + fgLine + right
	}
	return strings.Join(result, "\n")
}
