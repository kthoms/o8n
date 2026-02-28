package app

import "github.com/charmbracelet/bubbles/table"

// transitionScope identifies the category of state transition so
// prepareStateTransition knows what to clean up.
type transitionScope int

const (
	transitionEnvSwitch     transitionScope = iota // switching environments (Ctrl+E)
	transitionContextSwitch                        // switching root context (:)
	transitionDrilldown                            // drilling into a child resource (Enter/→)
	transitionBack                                 // navigating back (Esc)
	transitionBreadcrumb                           // jumping to a breadcrumb level (1-4)
)

// prepareStateTransition cleans up model state appropriate to the given transition scope.
// It always clears sort and search state. Additional cleanup depends on scope.
// For transitionBreadcrumb, pass the target depth (0-based index) as the optional depth arg.
func (m *model) prepareStateTransition(scope transitionScope, depth ...int) {
	// Clear sort state for all transitions
	m.sortColumn = -1
	m.sortAscending = true

	// Clear search state for all transitions
	m.searchTerm = ""
	m.searchMode = false
	m.searchInput.Blur()
	m.originalRows = nil
	m.filteredRows = nil

	// Reset popup if it was search mode
	if m.popup.mode == popupModeSearch {
		m.popup.mode = popupModeNone
		m.popup.input = ""
		m.popup.cursor = -1
		m.popup.offset = 0
		// restore original rows so table isn't stuck on filtered subset
		if m.originalRows != nil {
			m.table.SetRows(m.originalRows)
		}
	}

	switch scope {
	case transitionEnvSwitch:
		m.navigationStack = nil
		m.genericParams = make(map[string]string)
		m.selectedDefinitionKey = ""
		m.selectedInstanceID = ""
		// breadcrumb reset is handled by the caller (depends on new root)

	case transitionContextSwitch:
		m.navigationStack = nil
		m.genericParams = make(map[string]string)

	case transitionDrilldown:
		// navStack push handled by caller
		// sort/search already cleared above

	case transitionBack:
		// navStack pop handled by caller
		// sort/search already cleared above

	case transitionBreadcrumb:
		if len(depth) > 0 {
			d := depth[0]
			if d <= 0 {
				m.navigationStack = nil
			} else if d < len(m.navigationStack) {
				m.navigationStack = m.navigationStack[:d]
			}
			// if d >= len(navigationStack), no truncation needed
		}
	}
}

// clampCursorAfterRowRemoval ensures the table cursor stays within bounds after
// rows are removed (e.g. after terminate, delete). Safe to call always.
func (m *model) clampCursorAfterRowRemoval() {
	rows := m.table.Rows()
	cursor := m.table.Cursor()
	if len(rows) == 0 {
		m.table.SetCursor(0)
		return
	}
	if cursor >= len(rows) {
		m.table.SetCursor(len(rows) - 1)
	}
}

// clearSearch resets client-side search state without touching sort or nav.
// Used when search is explicitly cancelled (Esc in search popup).
func (m *model) clearSearch() {
	m.searchTerm = ""
	if m.originalRows != nil {
		m.table.SetRows(m.originalRows)
	}
	m.originalRows = nil
	m.filteredRows = nil
}

// clearSort resets sort state and removes sort indicators from column headers.
func (m *model) clearSort() {
	m.sortColumn = -1
	m.sortAscending = true
	// rebuild columns without sort indicators
	cols := m.table.Columns()
	cleaned := make([]table.Column, len(cols))
	for i, c := range cols {
		title := c.Title
		if len(title) > 2 && (title[len(title)-2:] == " ^" || title[len(title)-2:] == " v") {
			title = title[:len(title)-2]
		}
		cleaned[i] = table.Column{Title: title, Width: c.Width}
	}
	m.table.SetColumns(cleaned)
}
