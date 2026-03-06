package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/kthoms/o6n/internal/config"
	"github.com/kthoms/o6n/internal/dao"
)

// buildColumnsFor builds table.Column slice for a named table using the config definitions
// totalWidth is the available characters for the table content; if zero, returns reasonable defaults
func (m *model) buildColumnsFor(tableName string, totalWidth int) []table.Column {
	def := m.findTableDef(tableName)
	if def == nil {
		return []table.Column{{Title: "COL", Width: 20}, {Title: "COL2", Width: 20}}
	}

	// collect all columns that are not explicitly hidden, preserving config order
	type colEntry struct {
		def      config.ColumnDef
		title    string
		width    int // desired width
		minWidth int // minimum width (= header title length if not configured)
	}

	const drilldownPrefixWidth = 2 // "▶ " is prepended to first column when drilldown exists
	hasDrilldownPrefix := def.Drilldown != nil

	entries := make([]colEntry, 0, len(def.Columns))
	firstVisible := true
	for _, c := range def.Columns {
		if !c.IsVisible() {
			continue
		}
		title := strings.ToUpper(c.Name)
		if c.Editable {
			title = title + " ✎"
		}
		headerLen := lipgloss.Width(title)

		desired := c.Width
		if desired == 0 {
			desired = config.DefaultTypeWidth(c.Type)
		}
		// First column needs extra room for the "▶ " drilldown prefix
		if firstVisible && hasDrilldownPrefix {
			desired += drilldownPrefixWidth
		}

		minW := c.MinWidth
		if minW == 0 {
			minW = headerLen // implicit minimum = column header title width
		}
		firstVisible = false
		if desired < minW {
			desired = minW
		}

		entries = append(entries, colEntry{def: c, title: title, width: desired, minWidth: minW})
	}

	n := len(entries)
	if n == 0 {
		return []table.Column{{Title: "EMPTY", Width: 20}}
	}

	// extract ColumnDefs for hide sequence calculation
	defs := make([]config.ColumnDef, n)
	for i, e := range entries {
		defs[i] = e.def
	}

	// determine which columns to show given available totalWidth
	active := make([]bool, n)
	for i := range active {
		active[i] = true
	}

	if totalWidth > 0 {
		totalDesired := func() int {
			s := 0
			for i, e := range entries {
				if active[i] {
					s += e.width
				}
			}
			return s
		}

		hideSeq := config.HideSequence(defs)
		for _, hideIdx := range hideSeq {
			if totalDesired() <= totalWidth {
				break
			}
			active[hideIdx] = false
		}
	}

	// build final column list
	cols := make([]table.Column, 0, n)
	used := 0
	for i, e := range entries {
		if !active[i] {
			continue
		}
		used += e.width
		cols = append(cols, table.Column{Title: e.title, Width: e.width})
	}

	// last visible column stretches to fill any remaining space
	if len(cols) > 0 && totalWidth > 0 && used < totalWidth {
		cols[len(cols)-1].Width += totalWidth - used
	}
	return cols
}

// Update applyDefinitions and applyInstances to recover from panics and show footer error
func (m *model) applyDefinitions(defs []config.ProcessDefinition) {
	defer func() {
		if r := recover(); r != nil {
			m.footerError = fmt.Sprintf("Error rendering definitions: %v", r)
			m.footerStatusKind = footerStatusError
			log.Printf("applyDefinitions panic recovered: %v", r)
		}
	}()
	m.cachedDefinitions = defs
	items := make([]list.Item, 0, len(defs))
	rows := make([]table.Row, 0, len(defs))
	rd := make([]map[string]interface{}, 0, len(defs))
	for _, d := range defs {
		items = append(items, processDefinitionItem{definition: d})
		// Add drilldown prefix to first column for focus indicator
		rows = append(rows, table.Row{"▶ " + d.Key, d.Name, fmt.Sprintf("%d", d.Version), d.Resource})
		rd = append(rd, map[string]interface{}{
			"id": d.ID, "key": d.Key, "name": d.Name,
			"version": d.Version, "resource": d.Resource,
		})
	}
	m.rowData = rd
	m.list.SetItems(items)
	// determine available table width
	tableWidth := m.table.Width()
	if tableWidth <= 0 {
		tableWidth = m.paneWidth - 4
	}
	cols := m.buildColumnsFor(dao.ResourceProcessDefinitions, tableWidth)
	if len(cols) == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), tableWidth)
	}
	colsCount := len(cols)
	if colsCount == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), tableWidth)
		colsCount = len(cols)
	}
	m.table.SetRows([]table.Row{})
	m.table.SetColumns(cols)
	normRows := normalizeRows(rows, colsCount)
	m.setTableRowsSorted(normRows)
	if m.sortColumn >= 0 {
		m.applySortIndicatorToColumns()
	}
	// re-apply locked search filter if active
	if !m.searchMode && m.searchTerm != "" {
		filtered := filterRows(m.table.Rows(), m.searchTerm)
		m.table.SetRows(filtered)
	}
	m.viewMode = "process-definition"
}

func (m *model) applyInstances(instances []config.ProcessInstance) {
	defer func() {
		if r := recover(); r != nil {
			m.footerError = fmt.Sprintf("Error rendering instances: %v", r)
			m.footerStatusKind = footerStatusError
			log.Printf("applyInstances panic recovered: %v", r)
		}
	}()
	rows := make([]table.Row, 0, len(instances))
	rd := make([]map[string]interface{}, 0, len(instances))
	for _, inst := range instances {
		// Add drilldown prefix to first column for focus indicator
		rows = append(rows, table.Row{"▶ " + inst.ID, inst.DefinitionID, inst.BusinessKey, inst.StartTime})
		rd = append(rd, map[string]interface{}{
			"id": inst.ID, "definitionId": inst.DefinitionID,
			"businessKey": inst.BusinessKey, "startTime": inst.StartTime,
		})
	}
	m.rowData = rd
	tableWidth := m.table.Width()
	if tableWidth <= 0 {
		tableWidth = m.paneWidth - 4
	}
	cols := m.buildColumnsFor(dao.ResourceProcessInstances, tableWidth)
	if len(cols) == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), tableWidth)
	}
	colsCount := len(cols)
	if colsCount == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), tableWidth)
		colsCount = len(cols)
	}
	m.table.SetRows([]table.Row{})
	m.table.SetColumns(cols)
	normRows := normalizeRows(rows, colsCount)
	m.setTableRowsSorted(normRows)
	m.viewMode = "process-instance"
	if m.sortColumn >= 0 {
		m.applySortIndicatorToColumns()
	}
	// re-apply locked search filter if active
	if !m.searchMode && m.searchTerm != "" {
		filtered := filterRows(m.table.Rows(), m.searchTerm)
		m.table.SetRows(filtered)
	}
	// restore cursor position requested for paging operations
	if m.pendingCursorAfterPage >= 0 {
		last := len(normRows) - 1
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
}

// New: variables table
func (m *model) applyVariables(vars []config.Variable) {
	defer func() {
		if r := recover(); r != nil {
			m.footerError = fmt.Sprintf("Error loading variables: %v", r)
			m.footerStatusKind = footerStatusError
			log.Printf("variables panic recovered: %v", r)
		}
	}()

	rows := make([]table.Row, 0, len(vars))
	m.variablesByName = make(map[string]config.Variable, len(vars))
	for _, v := range vars {
		// Add drilldown prefix to first column for focus indicator
		rows = append(rows, table.Row{"▶ " + v.Name, v.Value})
		m.variablesByName[v.Name] = v
	}
	tableWidth := m.table.Width()
	if tableWidth <= 0 {
		tableWidth = m.paneWidth - 4
	}
	cols := m.buildColumnsFor(dao.ResourceProcessVariables, tableWidth)
	if len(cols) == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), tableWidth)
	}
	colsCount := len(cols)
	if colsCount == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), tableWidth)
		colsCount = len(cols)
	}
	m.table.SetRows([]table.Row{})
	m.table.SetColumns(cols)
	normRows := normalizeRows(rows, colsCount)
	m.setTableRowsSorted(normRows)
	if m.sortColumn >= 0 {
		m.applySortIndicatorToColumns()
	}
	// re-apply locked search filter if active
	if !m.searchMode && m.searchTerm != "" {
		filtered := filterRows(m.table.Rows(), m.searchTerm)
		m.table.SetRows(filtered)
	}
	m.viewMode = "process-variables"
}

// Backwards-compatible wrapper used by tests: applies both definitions and instances
func (m *model) applyData(defs []config.ProcessDefinition, instances []config.ProcessInstance) {
	m.applyDefinitions(defs)
	m.applyInstances(instances)
}

// Backwards-compatible removeInstance (filters table rows by instance id)
func (m *model) removeInstance(id string) {
	rows := m.table.Rows()
	filtered := make([]table.Row, 0, len(rows))
	for _, r := range rows {
		if rowInstanceID(r) == id {
			continue
		}
		filtered = append(filtered, r)
	}
	m.table.SetRows(filtered)
	m.selectedInstanceID = rowInstanceID(m.table.SelectedRow())
}

// helper to build N equal-width columns when no config is available
func defaultColumns(n int, totalWidth int) []table.Column {
	if n <= 0 {
		n = 1
	}
	if totalWidth <= 0 {
		// fallback width per column
		cols := make([]table.Column, 0, n)
		for i := 0; i < n; i++ {
			cols = append(cols, table.Column{Title: fmt.Sprintf("COL%d", i+1), Width: 20})
		}
		return cols
	}
	if totalWidth < n*3 {
		totalWidth = n * 3
	}
	cols := make([]table.Column, 0, n)
	per := totalWidth / n
	if per < 3 {
		per = 3
	}
	used := 0
	for i := 0; i < n; i++ {
		w := per
		used += w
		cols = append(cols, table.Column{Title: fmt.Sprintf("COL%d", i+1), Width: w})
	}
	if used < totalWidth {
		cols[len(cols)-1].Width += totalWidth - used
	}
	return cols
}

// filterRows returns rows where any cell contains the search term (case-insensitive).
func filterRows(rows []table.Row, term string) []table.Row {
	if term == "" {
		return rows
	}
	lower := strings.ToLower(term)
	var result []table.Row
	for _, row := range rows {
		for _, cell := range row {
			if strings.Contains(strings.ToLower(ansi.Strip(cell)), lower) {
				result = append(result, row)
				break
			}
		}
	}
	return result
}

// RowStyles holds skin-aware styles for row status colorization.
type RowStyles struct {
	Running   lipgloss.Style
	Suspended lipgloss.Style
	Failed    lipgloss.Style
	Ended     lipgloss.Style
}

// detectRowStatus determines the status of a row based on its resource type and column values.
// Returns one of: "running", "suspended", "failed", "ended", "normal"
func detectRowStatus(root string, row table.Row, columns []table.Column) string {
	// Build a column name -> index map
	colIdx := make(map[string]int)
	for i, col := range columns {
		colIdx[strings.ToLower(col.Title)] = i
	}

	getVal := func(name string) string {
		if idx, ok := colIdx[name]; ok && idx < len(row) {
			return strings.ToLower(strings.TrimSpace(row[idx]))
		}
		return ""
	}

	switch root {
	case "process-instance", "process-instances":
		if getVal("suspended") == "true" {
			return "suspended"
		}
		if getVal("ended") == "true" {
			return "ended"
		}
		return "running"
	case "job", "jobs":
		retries := getVal("retries")
		if retries == "0" {
			return "failed"
		}
		return "normal"
	case "incident", "incidents":
		return "failed"
	case "external-task", "external-tasks":
		if getVal("locked") == "true" || getVal("workerid") != "" {
			return "running"
		}
		return "normal"
	default:
		return "normal"
	}
}

// colorizeRows applies status-based color to each row using skin-aware styles.
func colorizeRows(root string, rows []table.Row, columns []table.Column, rs RowStyles) []table.Row {
	if len(rows) == 0 || len(columns) == 0 {
		return rows
	}
	result := make([]table.Row, len(rows))
	for i, row := range rows {
		status := detectRowStatus(root, row, columns)
		newRow := make(table.Row, len(row))
		copy(newRow, row)

		var style lipgloss.Style
		switch status {
		case "running":
			style = rs.Running
		case "suspended":
			style = rs.Suspended
		case "failed":
			style = rs.Failed
		case "ended":
			style = rs.Ended
		default:
			result[i] = newRow
			continue
		}

		for j, cell := range newRow {
			newRow[j] = style.Render(cell)
		}
		result[i] = newRow
	}
	return result
}

// sortTableRows sorts table rows by a given column index.
// It detects numeric and date values for proper sorting.
func sortTableRows(rows []table.Row, colIndex int, ascending bool) []table.Row {
	if colIndex < 0 || len(rows) == 0 {
		return rows
	}
	sorted := make([]table.Row, len(rows))
	copy(sorted, rows)

	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := "", ""
		if colIndex < len(sorted[i]) {
			a = sorted[i][colIndex]
		}
		if colIndex < len(sorted[j]) {
			b = sorted[j][colIndex]
		}

		// Strip ANSI escape sequences for comparison
		a = ansi.Strip(a)
		b = ansi.Strip(b)

		// Try numeric comparison
		af, aErr := strconv.ParseFloat(a, 64)
		bf, bErr := strconv.ParseFloat(b, 64)
		if aErr == nil && bErr == nil {
			if ascending {
				return af < bf
			}
			return af > bf
		}

		// Try date comparison (common ISO format)
		for _, layout := range []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		} {
			at, aErr := time.Parse(layout, a)
			bt, bErr := time.Parse(layout, b)
			if aErr == nil && bErr == nil {
				if ascending {
					return at.Before(bt)
				}
				return at.After(bt)
			}
		}

		// Lexicographic comparison
		if ascending {
			return strings.ToLower(a) < strings.ToLower(b)
		}
		return strings.ToLower(a) > strings.ToLower(b)
	})

	return sorted
}

// setTableRowsSorted sets the table rows, re-applying active sort if any.
func (m *model) setTableRowsSorted(rows []table.Row) {
	if m.sortColumn >= 0 {
		rows = sortTableRows(rows, m.sortColumn, m.sortAscending)
	}
	m.table.SetRows(rows)
}

// applySortIndicatorToColumns updates column titles to reflect the active sort column and direction.
// Must be called from Update() whenever m.sortColumn or m.sortAscending changes.
func (m *model) applySortIndicatorToColumns() {
	cols := m.table.Columns()
	if len(cols) == 0 {
		return
	}
	newCols := make([]table.Column, len(cols))
	copy(newCols, cols)
	// strip any existing sort indicator from all columns
	for i := range newCols {
		newCols[i].Title = strings.TrimSuffix(strings.TrimSuffix(newCols[i].Title, " ▲"), " ▼")
	}
	if m.sortColumn >= 0 && m.sortColumn < len(newCols) {
		indicator := " ▲"
		if !m.sortAscending {
			indicator = " ▼"
		}
		newCols[m.sortColumn].Title = newCols[m.sortColumn].Title + indicator
	}
	m.table.SetColumns(newCols)
}

// normalizeRows adjusts each row to have exactly colsCount columns (pad with empty strings or truncate).
func normalizeRows(rows []table.Row, colsCount int) []table.Row {
	if colsCount <= 0 {
		return rows
	}
	if len(rows) == 0 {
		// return single empty row to avoid table rendering issues
		empty := make(table.Row, colsCount)
		for i := range empty {
			empty[i] = ""
		}
		if colsCount > 0 {
			empty[0] = "No results"
		}
		return []table.Row{empty}
	}
	out := make([]table.Row, 0, len(rows))
	for _, r := range rows {
		nr := make(table.Row, colsCount)
		for i := 0; i < colsCount; i++ {
			if i < len(r) {
				nr[i] = r[i]
			} else {
				nr[i] = ""
			}
		}
		out = append(out, nr)
	}
	return out
}

// loadRootContexts extracts top-level paths from the OpenAPI spec file and returns pluralized names
func loadRootContexts(specPath string) []string {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return []string{dao.ResourceProcessDefinitions, dao.ResourceProcessInstances, dao.ResourceProcessVariables, dao.ResourceTasks, dao.ResourceJobs, dao.ResourceExternalTasks}
	}
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return []string{dao.ResourceProcessDefinitions, dao.ResourceProcessInstances, dao.ResourceProcessVariables, dao.ResourceTasks, dao.ResourceJobs, dao.ResourceExternalTasks}
	}
	pathsI, ok := doc["paths"]
	if !ok {
		return []string{dao.ResourceProcessDefinitions, dao.ResourceProcessInstances, dao.ResourceProcessVariables, dao.ResourceTasks, dao.ResourceJobs, dao.ResourceExternalTasks}
	}
	pathsMap, ok := pathsI.(map[string]interface{})
	if !ok {
		return []string{dao.ResourceProcessDefinitions, dao.ResourceProcessInstances, dao.ResourceProcessVariables, dao.ResourceTasks, dao.ResourceJobs, dao.ResourceExternalTasks}
	}
	set := map[string]struct{}{}
	for p := range pathsMap {
		seg := strings.TrimPrefix(p, "/")
		if seg == "" {
			continue
		}
		// take first segment before '/'
		parts := strings.Split(seg, "/")
		root := parts[0]
		// pluralize simply by appending 's' when not already ending with s
		if !strings.HasSuffix(root, "s") {
			root = root + "s"
		}
		set[root] = struct{}{}
	}
	roots := make([]string, 0, len(set))
	for r := range set {
		roots = append(roots, r)
	}
	sort.Strings(roots)
	return roots
}
