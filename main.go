package main

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
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kthoms/o8n/internal/client"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/dao"
	"golang.org/x/term"
)

const refreshInterval = 5 * time.Second

type refreshMsg struct{}

type dataLoadedMsg struct {
	definitions []config.ProcessDefinition
	instances   []config.ProcessInstance
}

type definitionsLoadedMsg struct{ definitions []config.ProcessDefinition }
type instancesLoadedMsg struct{ instances []config.ProcessInstance }

// New: fetch variables for a given instance
type variablesLoadedMsg struct{ variables []config.Variable }

type terminatedMsg struct{ id string }

type errMsg struct{ err error }

// Messages for the flash indicator
type flashOnMsg struct{}
type flashOffMsg struct{}

// message used to clear footer errors
type clearErrorMsg struct{}

// splash done message
type splashDoneMsg struct{}

// splash frame message for animation
type splashFrameMsg struct{ frame int }

const totalSplashFrames = 15

type processDefinitionItem struct {
	definition config.ProcessDefinition
}

func (i processDefinitionItem) Title() string {
	if i.definition.Name != "" {
		return i.definition.Name
	}
	return i.definition.Key
}

func (i processDefinitionItem) Description() string {
	return fmt.Sprintf("Key: %s v%d", i.definition.Key, i.definition.Version)
}

func (i processDefinitionItem) FilterValue() string {
	if i.definition.Name != "" {
		return i.definition.Name
	}
	return i.definition.Key
}

// KeyHint represents a keyboard shortcut with priority
type KeyHint struct {
	Key         string
	Description string
	Priority    int // 1=always visible, 9=only on wide terminals
}

// Modal types
type ModalType int

const (
	ModalNone ModalType = iota
	ModalConfirmDelete
	ModalConfirmQuit
	ModalHelp
	ModalEdit
)

type editSavedMsg struct {
	rowIndex int
	colIndex int
	value    string
}

type editableColumn struct {
	index int
	def   config.ColumnDef
}

// viewState captures the complete state of a view for navigation history
type viewState struct {
	viewMode              string
	breadcrumb            []string
	contentHeader         string
	selectedDefinitionKey string
	selectedInstanceID    string
	tableRows             []table.Row
	tableCursor           int
	cachedDefinitions     []config.ProcessDefinition
	tableColumns          []table.Column
}

type model struct {
	config *config.Config

	envNames []string

	currentEnv         string
	autoRefresh        bool
	showKillModal      bool
	selectedInstanceID string
	// manualRefreshTriggered is set to true shortly after a manual refresh
	// (selection-change when auto-refresh is off). It's used to render a
	// small visual hint in the header.
	manualRefreshTriggered bool

	list  list.Model
	table table.Model

	// cached definitions for drilldown/back
	cachedDefinitions []config.ProcessDefinition

	// navigation history stack for back/forward
	navigationStack []viewState

	// view mode: "definitions" or "instances"
	viewMode string

	// Modal state
	activeModal     ModalType
	modalConfirmKey string // The key to press to confirm (e.g., "ctrl+d")
	pendingDeleteID string // ID pending deletion confirmation

	// Edit modal state
	editInput     textinput.Model
	editColumns   []editableColumn
	editColumnPos int
	editRowIndex  int
	editTableKey  string
	editError     string

	// lastListIndex stores the last-known list index so we can detect
	// selection changes even when list.Update doesn't change the index in tests.
	lastListIndex int

	// flashActive indicates the footer flash should be shown
	flashActive bool

	// last known terminal size (set on WindowSizeMsg)
	lastWidth  int
	lastHeight int

	// pane sizes (set on WindowSizeMsg)
	paneWidth  int
	paneHeight int

	style lipgloss.Style

	// EnvConfig and AppConfig for dynamic loading
	envConfig *config.EnvConfig
	appConfig *config.AppConfig

	// colon popup
	showRootPopup bool
	// available root contexts (computed from API spec)
	rootContexts []string
	rootSelected int

	// current root context
	currentRoot string

	// root input for context selection
	rootInput string

	// footer error message
	footerError string

	// splash screen active
	splashActive bool

	// current splash frame
	splashFrame int

	// splash styles
	splashLogoStyle lipgloss.Style
	splashInfoStyle lipgloss.Style

	// breadcrumb navigation
	breadcrumb    []string
	contentHeader string

	// selected definition key when drilled into instances
	selectedDefinitionKey string

	// breadcrumb styles per level
	breadcrumbStyles []lipgloss.Style

	// variables cache for type-aware editing
	variablesByName map[string]config.Variable

	// Version number
	version      string
	debugEnabled bool
}

// getKeyHints returns keyboard hints based on current view and terminal width
func (m *model) getKeyHints(width int) []KeyHint {
	hints := []KeyHint{}

	// Global hints (always relevant)
	hints = append(hints,
		KeyHint{"?", "help", 1},
		KeyHint{":", "switch", 2},
	)

	if m.viewMode == "definitions" {
		hints = append(hints,
			KeyHint{"↑↓", "nav", 3},
			KeyHint{"Enter", "drill", 4},
		)
	} else if m.viewMode == "instances" {
		hints = append(hints,
			KeyHint{"Esc", "back", 5},
			KeyHint{"↑↓", "nav", 3},
			KeyHint{"Enter", "vars", 4},
		)
	} else if m.viewMode == "variables" {
		hints = append(hints,
			KeyHint{"Esc", "back", 5},
			KeyHint{"↑↓", "nav", 3},
		)
		if m.hasEditableColumns() {
			hints = append(hints, KeyHint{"e", "edit", 4})
		}
	}

	// Add other hints based on width thresholds
	if width >= 90 {
		hints = append(hints, KeyHint{"<ctrl>-r", "refresh", 6})
	}
	if width >= 100 && m.viewMode == "instances" {
		hints = append(hints, KeyHint{"<ctrl>+d", "delete", 7})
	}
	if width >= 110 {
		hints = append(hints, KeyHint{"<ctrl>+c", "quit", 8})
	}
	if width >= 130 {
		hints = append(hints, KeyHint{"<ctrl>+e", "env", 9})
	}

	return hints
}

// renderCompactHeader renders a 3-row header
func (m *model) renderCompactHeader(width int) string {
	if width <= 0 {
		width = 80
	}

	// Row 1: Status line (environment, URL, user, connection status)
	envInfo := fmt.Sprintf("%s @ %s", m.currentEnv, "localhost:8080")
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			envInfo = fmt.Sprintf("%s @ %s │ %s", m.currentEnv, env.URL, env.Username)
		}
	}

	row1 := fmt.Sprintf("[H] o8n %s │ %s", m.version, envInfo)
	if len(row1) > width-4 {
		row1 = row1[:width-7] + "..."
	}

	// Row 2: Key hints (priority-based)
	hints := m.getKeyHints(width)
	row2Parts := []string{}
	for _, hint := range hints {
		part := fmt.Sprintf("%s %s", hint.Key, hint.Description)
		row2Parts = append(row2Parts, part)
	}
	row2 := strings.Join(row2Parts, "  ")
	if len(row2) > width-4 {
		row2 = row2[:width-7] + "..."
	}

	// Row 3: Empty spacer
	row3 := ""

	// Join rows
	header := fmt.Sprintf("%s\n%s\n%s", row1, row2, row3)

	// Always render header with a visible background so it's readable across terminals
	// Use environment UI color as accent if provided
	fg := "white"
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			if env.UIColor != "" {
				fg = env.UIColor
			}
		}
	}
	headerStyle := lipgloss.NewStyle().Width(width).Padding(0, 1).Background(lipgloss.Color("#2b2b2b")).Foreground(lipgloss.Color(fg)).Bold(true)
	return headerStyle.Render(header)
}

func newModel(cfg *config.Config) model {
	envNames := make([]string, 0, len(cfg.Environments))
	for name := range cfg.Environments {
		envNames = append(envNames, name)
	}
	sort.Strings(envNames)

	delegate := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, delegate, 30, 20)
	l.Title = "Process Definitions"

	// table will be configured dynamically based on view
	cols := []table.Column{
		{Title: "KEY", Width: 20},
		{Title: "NAME", Width: 28},
		{Title: "VERSION", Width: 8},
		{Title: "RESOURCE", Width: 30},
	}
	t := table.New(table.WithColumns(cols), table.WithFocused(true))

	// Determine starting environment: use cfg.Active, else 'local' if present, else first
	current := ""
	if cfg.Active != "" {
		if _, ok := cfg.Environments[cfg.Active]; ok {
			current = cfg.Active
		}
	}
	if current == "" {
		if _, ok := cfg.Environments["local"]; ok {
			current = "local"
		}
	}
	if current == "" && len(envNames) > 0 {
		current = envNames[0]
	}

	m := model{
		config:          cfg,
		envNames:        envNames,
		currentEnv:      current,
		list:            l,
		table:           t,
		viewMode:        "definitions",
		splashActive:    true,
		splashFrame:     1,
		activeModal:     ModalNone,
		version:         "0.1.0",
		variablesByName: map[string]config.Variable{},
		debugEnabled:    false,
	}

	// edit input defaults
	editInput := textinput.New()
	editInput.Placeholder = "value"
	editInput.Prompt = "> "
	editInput.CharLimit = 0
	editInput.Width = 40
	m.editInput = editInput
	m.applyStyle()
	// initialize lastListIndex
	m.lastListIndex = m.list.Index()

	// initialize breadcrumb: start with current root
	m.breadcrumb = []string{dao.ResourceProcessDefinitions}
	m.contentHeader = dao.ResourceProcessDefinitions

	// sensible defaults so the header is visible immediately
	m.lastWidth = 80
	m.lastHeight = 24
	// try to detect actual terminal size right away so we use full width
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		if w > 0 {
			m.lastWidth = w
		}
		if h > 0 {
			m.lastHeight = h
		}
	}
	// compute default paneWidth as remaining width after left column + margins
	leftW := m.lastWidth / 4
	if leftW < 12 {
		leftW = 12
	}
	m.paneWidth = m.lastWidth - leftW - 4
	if m.paneWidth < 20 {
		m.paneWidth = m.lastWidth - 4
	}
	// compute content height reserving header/context/footer lines so header is visible
	// compactHeader (3 lines) + content header (1 line) = 4 header lines total
	headerLines := 4
	contextSelectionLines := 1
	footerLines := 1
	// reserve an extra safe line to avoid off-by-one overflow
	contentHeight := m.lastHeight - headerLines - contextSelectionLines - footerLines - 1
	if contentHeight < 3 {
		contentHeight = 3
	}
	m.paneHeight = contentHeight

	// initialize list/table sizes to match detected terminal size
	m.list.SetSize(leftW-2, contentHeight-1)
	tableInner := m.paneWidth - 4
	if tableInner < 10 {
		tableInner = 10
	}
	m.table.SetWidth(tableInner)
	m.table.SetHeight(contentHeight - 1)

	// set root contexts and currentRoot
	m.rootContexts = loadRootContexts("resources/operaton-rest-api.json")
	if len(m.rootContexts) > 0 {
		m.currentRoot = dao.ResourceProcessDefinitions
	} else {
		m.currentRoot = dao.ResourceProcessDefinitions
	}

	return m
}

func newModelEnvApp(envCfg *config.EnvConfig, appCfg *config.AppConfig) model {
	// Build compatibility Config from split configs
	cfg := &config.Config{
		Environments: make(map[string]config.Environment),
		Active:       "",
		Tables:       appCfg.Tables,
	}
	for k, v := range envCfg.Environments {
		cfg.Environments[k] = config.Environment{
			URL:      v.URL,
			Username: v.Username,
			Password: v.Password,
			UIColor:  v.UIColor,
		}
	}
	cfg.Active = envCfg.Active

	m := newModel(cfg)
	m.envConfig = envCfg
	m.appConfig = appCfg
	// copy tables into m.config for backward compatibility
	m.config = cfg
	return m
}

// renderConfirmDeleteModal renders a modal for confirming delete action
func (m *model) renderConfirmDeleteModal(width, height int) string {
	selected := m.table.SelectedRow()
	if len(selected) == 0 {
		return ""
	}

	instanceID := fmt.Sprintf("%v", selected[0])
	defName := "Unknown"
	if len(selected) > 1 {
		defName = fmt.Sprintf("%v", selected[1])
	}

	modalContent := fmt.Sprintf(
		"⚠️  DELETE PROCESS INSTANCE\n\n"+
			"You are about to DELETE this instance:\n\n"+
			"Instance ID:   %s\n"+
			"Definition:    %s\n\n"+
			"⚠️  WARNING: This action CANNOT be undone!\n"+
			"The instance will be terminated.\n\n"+
			"<ctrl>+d  Confirm Delete    Esc  Cancel",
		instanceID, defName)

	// Get color for styling
	color := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			color = env.UIColor
		}
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(color)).
		Padding(1, 2).
		Width(54)

	modal := modalStyle.Render(modalContent)

	// Center the modal
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}

// renderHelpScreen renders the help screen modal
func (m *model) renderHelpScreen(width, height int) string {
	helpContent := `o8n Help

NAVIGATION              │  ACTIONS                │  GLOBAL
─────────────────────   │  ────────────────────   │  ──────────────────
↑/↓      Navigate list  │  <ctrl>+e  Switch env   │  ?     This help
PgUp/Dn  Page up/down   │  <ctrl>-r  Auto-refresh │  :     Switch view
Home     First item     │  <ctrl>+d  Delete item  │  <ctrl>+c Quit
End      Last item      │  <ctrl>-R  Force refresh│
Enter    Drill down     │                         │  CONTEXT
Esc      Go back        │  VIEW SPECIFIC          │  ───────────────────
                        │  (varies by view)       │  Tab      Complete
SEARCH (Coming Soon)    │                         │  Enter    Confirm
─────────────────────   │  In Process Instances:  │  Esc      Cancel
/        Search/filter  │  v  View variables      │
n        Next match     │  <ctrl>+d Kill instance │
N        Prev match     │  s  Suspend instance    │
                        │  r  Resume instance     │
					   │  In Variables:          │
					   │  e  Edit value          │

Current View: ` + m.viewMode + `
Environment: ` + m.currentEnv + `

Press any key to close`

	// Get color for styling
	color := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			color = env.UIColor
		}
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Padding(1, 2).
		Width(76)

	modal := modalStyle.Render(helpContent)

	// Center the modal
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}

func (m *model) renderEditModal(width, height int) string {
	row := m.currentEditRow()
	col := m.currentEditColumn()
	if row == nil || col == nil {
		return ""
	}
	inputType, _ := m.resolveEditTypes(col.def, m.editTableKey, row)

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
		columnsLine = "Columns: " + strings.Join(parts, " ") + "\n\n"
	}

	errorLine := ""
	if m.editError != "" {
		errorLine = "Error: " + m.editError + "\n\n"
	}

	modalContent := fmt.Sprintf(
		"EDIT VALUE\n\n"+
			"Table: %s\n"+
			"Column: %s\n"+
			"Type: %s\n\n"+
			"%s"+
			"%s"+
			"%s\n\n"+
			"Enter Save    Esc Cancel    Tab Next Column",
		m.editTableKey,
		col.def.Name,
		inputType,
		columnsLine,
		m.editInput.View(),
		errorLine,
	)

	color := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			color = env.UIColor
		}
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Padding(1, 2).
		Width(60)

	modal := modalStyle.Render(modalContent)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}

func (m *model) applyStyle() {
	color := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			color = env.UIColor
		}
	}
	m.style = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(color)).Bold(true)

	listStyles := list.DefaultStyles()
	listStyles.Title = listStyles.Title.BorderForeground(lipgloss.Color(color))
	m.list.Styles = listStyles

	// Table header color is white as requested. Remove header bottom border to hide separator line.
	tStyles := table.DefaultStyles()
	tStyles.Header = tStyles.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(color)).BorderBottom(false).Foreground(lipgloss.Color("white")).Bold(true)
	tStyles.Selected = tStyles.Selected.Foreground(lipgloss.Color(color)).Bold(true)
	m.table.SetStyles(tStyles)

	// splash styles
	m.splashLogoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Align(lipgloss.Center)
	m.splashInfoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Align(lipgloss.Center)

	// breadcrumb styles (up to 4 levels)
	m.breadcrumbStyles = []lipgloss.Style{
		lipgloss.NewStyle().Background(lipgloss.Color(color)).Foreground(lipgloss.Color("black")).Padding(0, 1),
		lipgloss.NewStyle().Background(lipgloss.Color("#e6e6fa")).Foreground(lipgloss.Color("black")).Padding(0, 1),
		lipgloss.NewStyle().Background(lipgloss.Color("#f0fff0")).Foreground(lipgloss.Color("black")).Padding(0, 1),
		lipgloss.NewStyle().Background(lipgloss.Color("#fffaf0")).Foreground(lipgloss.Color("black")).Padding(0, 1),
	}
}

func (m model) Init() tea.Cmd {
	// fetch definitions at start and flash, and start splash animation (150ms per frame, total 1.5s)
	// we start with frame 1 already set; schedule frame 2 after 150ms
	firstTick := tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return splashFrameMsg{frame: 2} })
	return tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), firstTick)
}

func (m *model) nextEnvironment() {
	if len(m.envNames) == 0 {
		return
	}
	idx := 0
	for i, name := range m.envNames {
		if name == m.currentEnv {
			idx = i
			break
		}
	}
	idx = (idx + 1) % len(m.envNames)
	m.currentEnv = m.envNames[idx]
	m.applyStyle()
	// persist active environment in the config file; best-effort
	if m.config != nil {
		m.config.Active = m.currentEnv
		if err := config.SaveConfig("config.yaml", m.config); err != nil {
			log.Printf("warning: failed to save config active environment: %v", err)
		}
	}
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
		if !c.Visible {
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
			if !c.Visible {
				continue
			}
			if strings.EqualFold(c.Name, "id") {
				return idx
			}
			idx++
		}
	}
	return -1
}

func (m *model) currentTableKey() string {
	if len(m.breadcrumb) > 0 {
		last := m.breadcrumb[len(m.breadcrumb)-1]
		if last == "variables" {
			return dao.ResourceProcessVariables
		}
		return last
	}
	return m.currentRoot
}

func (m *model) editableColumnsFor(tableKey string) []editableColumn {
	def := m.findTableDef(tableKey)
	if def == nil {
		return nil
	}
	cols := []editableColumn{}
	idx := 0
	for _, c := range def.Columns {
		if !c.Visible {
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

func (m *model) variableTypeForRow(tableKey string, row table.Row) string {
	if tableKey != "process-variables" && tableKey != "variables" && tableKey != "variable-instance" && tableKey != "variable-instances" {
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
	trimmed := strings.TrimSpace(input)
	switch inputType {
	case "bool":
		if strings.EqualFold(trimmed, "true") || trimmed == "1" {
			return true, nil
		}
		if strings.EqualFold(trimmed, "false") || trimmed == "0" {
			return false, nil
		}
		return nil, fmt.Errorf("enter true or false")
	case "int":
		v, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("enter an integer")
		}
		return v, nil
	case "number":
		v, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("enter a number")
		}
		return v, nil
	default:
		return input, nil
	}
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

func (m *model) startEdit(tableKey string) {
	cols := m.editableColumnsFor(tableKey)
	if len(cols) == 0 {
		m.footerError = "No editable columns"
		return
	}
	if len(m.table.Rows()) == 0 {
		m.footerError = "No row selected"
		return
	}
	m.editColumns = cols
	m.editTableKey = tableKey
	m.editRowIndex = m.table.Cursor()
	m.editColumnPos = 0
	m.editError = ""
	m.editInput.Focus()
	m.setEditColumn(0)
	m.activeModal = ModalEdit
}

// buildColumnsFor builds table.Column slice for a named table using the config definitions
// totalWidth is the available characters for the table content; if zero, returns reasonable defaults
// addEditableMarkers adds [E] suffix to editable columns in table rows
func (m *model) addEditableMarkers(rows []table.Row, tableName string) []table.Row {
	def := m.findTableDef(tableName)
	if def == nil {
		return rows
	}
	// find which visible column indices are editable
	editableIndices := make(map[int]bool)
	visIdx := 0
	for _, c := range def.Columns {
		if !c.Visible {
			continue
		}
		if c.Editable {
			editableIndices[visIdx] = true
		}
		visIdx++
	}
	if len(editableIndices) == 0 {
		return rows
	}
	// clone and mark editable cells
	marked := make([]table.Row, len(rows))
	for i, row := range rows {
		marked[i] = make(table.Row, len(row))
		for j, cell := range row {
			if editableIndices[j] {
				marked[i][j] = fmt.Sprintf("%v [E]", cell)
			} else {
				marked[i][j] = cell
			}
		}
	}
	return marked
}

func (m *model) buildColumnsFor(tableName string, totalWidth int) []table.Column {
	def := m.findTableDef(tableName)
	if def == nil {
		// fallback: reasonable default column
		return []table.Column{{Title: "COL", Width: 20}, {Title: "COL2", Width: 20}}
	}

	// collect visible columns
	visible := make([]config.ColumnDef, 0, len(def.Columns))
	for _, c := range def.Columns {
		if c.Visible {
			visible = append(visible, c)
		}
	}
	n := len(visible)
	if n == 0 {
		return []table.Column{{Title: "EMPTY", Width: 20}}
	}

	contentWidth := totalWidth
	if contentWidth < n*3 {
		contentWidth = n * 3
	}

	// parse percentages
	percentTotal := 0
	percentCols := make([]int, n)
	unspecified := 0
	for i, c := range visible {
		if len(c.Width) > 0 && c.Width[len(c.Width)-1] == '%' {
			var p int
			fmt.Sscanf(c.Width, "%d%%", &p)
			percentTotal += p
			percentCols[i] = p
		} else {
			unspecified++
			percentCols[i] = 0
		}
	}

	remaining := 100 - percentTotal
	if remaining < 0 {
		// normalize: scale down specified percentages proportionally
		for i := range percentCols {
			if percentCols[i] > 0 {
				percentCols[i] = percentCols[i] * 100 / percentTotal
			}
		}
		remaining = 100 - func() int {
			s := 0
			for _, v := range percentCols {
				s += v
			}
			return s
		}()
	}
	// distribute remaining among unspecified equally
	if unspecified > 0 {
		per := remaining / unspecified
		for i := range percentCols {
			if percentCols[i] == 0 {
				percentCols[i] = per
			}
		}
	}

	// if totalWidth known, compute column widths using contentWidth
	cols := make([]table.Column, 0, n)
	used := 0
	for i := range visible {
		w := contentWidth * percentCols[i] / 100
		if w < 3 {
			w = 3
		}
		used += w
		title := strings.ToUpper(visible[i].Name)
		if visible[i].Editable {
			title = title + "*"
		}
		cols = append(cols, table.Column{Title: title, Width: w})
	}
	if used < contentWidth {
		cols[len(cols)-1].Width += contentWidth - used
	}
	return cols
}

// Update applyDefinitions and applyInstances to recover from panics and show footer error
func (m *model) applyDefinitions(defs []config.ProcessDefinition) {
	defer func() {
		if r := recover(); r != nil {
			m.footerError = fmt.Sprintf("Error rendering definitions: %v", r)
			log.Printf("applyDefinitions panic recovered: %v", r)
		}
	}()
	m.cachedDefinitions = defs
	items := make([]list.Item, 0, len(defs))
	rows := make([]table.Row, 0, len(defs))
	for _, d := range defs {
		items = append(items, processDefinitionItem{definition: d})
		rows = append(rows, table.Row{d.Key, d.Name, fmt.Sprintf("%d", d.Version), d.Resource})
	}
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
	normRows := normalizeRows(rows, colsCount)
	normRows = m.addEditableMarkers(normRows, dao.ResourceProcessDefinitions)
	m.table.SetColumns(cols)
	m.table.SetRows(normRows)
	m.viewMode = "definitions"
}

func (m *model) applyInstances(instances []config.ProcessInstance) {
	defer func() {
		if r := recover(); r != nil {
			m.footerError = fmt.Sprintf("Error rendering instances: %v", r)
			log.Printf("applyInstances panic recovered: %v", r)
		}
	}()
	rows := make([]table.Row, 0, len(instances))
	for _, inst := range instances {
		rows = append(rows, table.Row{inst.ID, inst.DefinitionID, inst.BusinessKey, inst.StartTime})
	}
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
	normRows := normalizeRows(rows, colsCount)
	normRows = m.addEditableMarkers(normRows, dao.ResourceProcessInstances)
	m.table.SetColumns(cols)
	m.table.SetRows(normRows)
	m.viewMode = "instances"
}

// New: variables table
func (m *model) applyVariables(vars []config.Variable) {
	defer func() {
		if r := recover(); r != nil {
			m.footerError = fmt.Sprintf("Error loading variables: %v", r)
			log.Printf("variables panic recovered: %v", r)
		}
	}()

	rows := make([]table.Row, 0, len(vars))
	m.variablesByName = make(map[string]config.Variable, len(vars))
	for _, v := range vars {
		rows = append(rows, table.Row{v.Name, v.Value})
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
	normRows := normalizeRows(rows, colsCount)
	normRows = m.addEditableMarkers(normRows, dao.ResourceProcessVariables)
	m.table.SetColumns(cols)
	m.table.SetRows(normRows)
	m.viewMode = "variables"
}

func (m model) fetchDefinitionsCmd() tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env)
	return func() tea.Msg {
		defs, err := c.FetchProcessDefinitions()
		if err != nil {
			return errMsg{err}
		}
		return definitionsLoadedMsg{definitions: defs}
	}
}

func (m model) fetchInstancesCmd(processKey string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env)
	return func() tea.Msg {
		instances, err := c.FetchInstances(processKey)
		if err != nil {
			return errMsg{err}
		}
		return instancesLoadedMsg{instances: instances}
	}
}

// New: fetch variables for a given instance
func (m model) fetchVariablesCmd(instanceID string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env)
	return func() tea.Msg {
		vars, err := c.FetchVariables(instanceID)
		if err != nil {
			return errMsg{err}
		}
		return variablesLoadedMsg{variables: vars}
	}
}

func (m model) fetchDataCmd() tea.Cmd {
	// keep original behaviour for tests: fetch defs and instances for selected key
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}

	var selectedKey string
	items := m.list.Items()
	if len(items) > 0 {
		index := m.list.Index()
		if index >= 0 && index < len(items) {
			if item, ok := items[index].(processDefinitionItem); ok {
				selectedKey = item.definition.Key
			}
		}
	}

	c := client.NewClient(env)

	return func() tea.Msg {
		defs, err := c.FetchProcessDefinitions()
		if err != nil {
			return errMsg{err}
		}

		var instances []config.ProcessInstance
		if selectedKey != "" {
			instances, err = c.FetchInstances(selectedKey)
			if err != nil {
				return errMsg{err}
			}
		}

		return dataLoadedMsg{definitions: defs, instances: instances}
	}
}

func (m model) terminateInstanceCmd(id string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env)

	return func() tea.Msg {
		if err := c.TerminateInstance(id); err != nil {
			return errMsg{err}
		}
		return terminatedMsg{id: id}
	}
}

func (m model) setVariableCmd(instanceID, varName string, value interface{}, valueType string, rowIndex, colIndex int, displayValue string) tea.Cmd {
	if instanceID == "" || varName == "" {
		return func() tea.Msg { return errMsg{fmt.Errorf("missing instance or variable name")} }
	}
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env)
	return func() tea.Msg {
		if err := c.SetProcessInstanceVariable(instanceID, varName, value, valueType); err != nil {
			return errMsg{err}
		}
		return editSavedMsg{rowIndex: rowIndex, colIndex: colIndex, value: displayValue}
	}
}

// helper to create a command that triggers the flash indicator
func flashOnCmd() tea.Cmd {
	return func() tea.Msg { return flashOnMsg{} }
}

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

		// Handle modal-specific keys first
		if m.activeModal == ModalHelp {
			// Any key closes help screen
			m.activeModal = ModalNone
			return m, nil
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
				m.setEditColumn(m.editColumnPos + 1)
				return m, nil
			case "shift+tab", "backtab":
				m.setEditColumn(m.editColumnPos - 1)
				return m, nil
			case " ", "space":
				if inputType == "bool" {
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
				if row == nil || col == nil {
					m.editError = "No selection"
					return m, nil
				}
				if m.editTableKey == "process-variables" || m.editTableKey == "variables" || m.editTableKey == "variable-instance" || m.editTableKey == "variable-instances" {
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
					return m, tea.Batch(m.setVariableCmd(m.selectedInstanceID, varName, parsedValue, typeName, rowIndex, colIndex, displayValue), flashOnCmd())
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
				if m.pendingDeleteID != "" {
					return m, tea.Batch(m.terminateInstanceCmd(m.pendingDeleteID), flashOnCmd())
				}
			} else {
				// Cancel
				m.activeModal = ModalNone
				m.pendingDeleteID = ""
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
				m.footerError = ""
			} else {
				m.showRootPopup = false
			}
			return m, nil
		}

		switch s {
		case "?":
			// Show help screen
			m.activeModal = ModalHelp
			return m, nil
		case "ctrl+c":
			// Quit via <ctrl>+c only; do not exit on plain 'q'
			return m, tea.Quit
		case "ctrl+e":
			// Switch environment
			m.nextEnvironment()
			m.resetViews()
			// refetch definitions for new env and flash
			return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd())
		case "ctrl+r", "r":
			// Toggle auto-refresh
			m.autoRefresh = !m.autoRefresh
			if m.autoRefresh {
				return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} }))
			}
			return m, nil
		case "ctrl+d":
			// Delete/kill instance - show confirmation modal
			if m.viewMode == "instances" {
				row := m.table.SelectedRow()
				if len(row) > 0 {
					m.pendingDeleteID = fmt.Sprintf("%v", row[0])
					m.activeModal = ModalConfirmDelete
				}
			}
			return m, nil
		case "e":
			if m.showRootPopup {
				m.rootInput += s
				return m, nil
			}
			tableKey := m.currentTableKey()
			if len(m.editableColumnsFor(tableKey)) > 0 {
				m.startEdit(tableKey)
				return m, nil
			}
			m.footerError = "No editable columns"
			return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return clearErrorMsg{} })
		case "esc":
			if m.showRootPopup {
				m.showRootPopup = false
				m.rootInput = ""
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
				// only switch if exact match to a known root context
				for _, rc := range m.rootContexts {
					if rc == m.rootInput {
						m.currentRoot = rc
						m.showRootPopup = false
						m.rootInput = ""
						// clear any footer error
						m.footerError = ""
						// reset breadcrumb and header
						m.breadcrumb = []string{rc}
						m.contentHeader = rc
						return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd())
					}
				}
				// no exact match: ignore
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

				// resolve value from selected row (fall back to first cell)
				colName := chosen.Column
				if colName == "" {
					colName = "id"
				}
				idx := m.visibleColumnIndex(def, colName)
				val := ""
				if idx >= 0 && idx < len(row) {
					val = fmt.Sprintf("%v", row[idx])
				} else {
					val = fmt.Sprintf("%v", row[0])
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
					}
					m.navigationStack = append(m.navigationStack, currentState2)
					// definitions -> instances (expects a process key)
					m.selectedDefinitionKey = val
					m.viewMode = "instances"
					m.breadcrumb = []string{m.currentRoot, "process-instances"}
					m.contentHeader = fmt.Sprintf("%s(%s)", m.currentRoot, val)
					// reset cursor to row 0
					m.table.SetCursor(0)
					return m, tea.Batch(m.fetchInstancesCmd(val), flashOnCmd())
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
					// config declares a drill target but UI/client not implemented yet
					m.footerError = fmt.Sprintf("Drill target '%s' not supported in UI yet", chosen.Target)
					return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg { return clearErrorMsg{} })
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
				}
				m.navigationStack = append(m.navigationStack, currentState3)

				key := fmt.Sprintf("%v", row[0])
				m.selectedDefinitionKey = key
				m.viewMode = "instances"
				m.breadcrumb = []string{m.currentRoot, "process-instances"}
				m.contentHeader = fmt.Sprintf("%s(%s)", m.currentRoot, key)
				// reset cursor to row 0
				m.table.SetCursor(0)
				return m, tea.Batch(m.fetchInstancesCmd(key), flashOnCmd())
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
				}
				m.navigationStack = append(m.navigationStack, currentState4)

				id := fmt.Sprintf("%v", row[0])
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
			if m.showRootPopup && len(m.rootInput) > 0 {
				// complete to first match
				for _, rc := range m.rootContexts {
					if strings.HasPrefix(rc, m.rootInput) {
						m.rootInput = rc
						break
					}
				}
				return m, nil
			}
			return m, nil
		case "backspace":
			if m.showRootPopup {
				if len(m.rootInput) > 0 {
					m.rootInput = m.rootInput[:len(m.rootInput)-1]
				}
				return m, nil
			}
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
		// Reserve lines: compact header 3 rows + content header 1 row + context selection 1 line + footer 1 line
		headerLines := 4 // compactHeader (3) + content header (1)
		contextSelectionLines := 1
		footerLines := 1
		// reserve an extra safe line to avoid off-by-one overflow
		contentHeight := height - headerLines - contextSelectionLines - footerLines - 1
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
			return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} }))
		}
	case flashOnMsg:
		m.flashActive = true
		// schedule turning the flash off after 200ms
		return m, tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg { return flashOffMsg{} })
	case flashOffMsg:
		m.flashActive = false
	case definitionsLoadedMsg:
		m.applyDefinitions(msg.definitions)
	case instancesLoadedMsg:
		m.applyInstances(msg.instances)
	case variablesLoadedMsg:
		// show variables table
		m.applyVariables(msg.variables)
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
		m.footerError = "Saved"
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return clearErrorMsg{} })
	case dataLoadedMsg:
		// keep backward compatibility: only apply definitions to avoid auto-drilldown
		m.applyDefinitions(msg.definitions)
	case terminatedMsg:
		m.removeInstance(msg.id)
	case errMsg:
		// display error in footer
		m.footerError = msg.err.Error()
		// clear after 4 seconds
		return m, tea.Tick(4*time.Second, func(time.Time) tea.Msg { return clearErrorMsg{} })
	case clearErrorMsg:
		m.footerError = ""
	}

	var cmd tea.Cmd
	// update list/table internals as usual
	prevIndex := m.list.Index()
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	newIndex := m.list.Index()
	changed := prevIndex != newIndex || newIndex != m.lastListIndex
	// only set manualRefreshTriggered; do NOT fetch instances on selection change
	if changed && !m.autoRefresh {
		m.manualRefreshTriggered = true
	}
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

func (m model) asciiArt() string {
	return "  ____   ___  _  _\n / __ \\ / _ \\| \\| |\n| |  | | | | |  \\|\n| |  | | | | | . ` |\n| |__| | |_| | |\\  |\n \\____/ \\___/|_| \\_|\n" + "o8n"
}

func (m model) View() string {
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
		info := "v0.1.0"
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
	// Ensure compact header occupies exactly 3 rows so it remains visible
	compactHeader = lipgloss.Place(m.lastWidth, 3, lipgloss.Left, lipgloss.Center, compactHeader)

	// get border color
	color := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			color = env.UIColor
		}
	}

	// render context selection box (1 line)
	var contextSelectionBox string
	if m.showRootPopup {
		completion := ""
		if m.rootInput != "" {
			for _, rc := range m.rootContexts {
				if strings.HasPrefix(rc, m.rootInput) && rc != m.rootInput {
					completion = rc[len(m.rootInput):]
					break
				}
			}
		}

		inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		completionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		displayText := inputStyle.Render(m.rootInput) + completionStyle.Render(completion)

		boxStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(color)).
			Width(m.lastWidth-4).
			Padding(0, 1)
		contextSelectionBox = boxStyle.Render(displayText)
	} else {
		// do not render an empty boxed context selection when popup is inactive
		contextSelectionBox = ""
	}

	// main content boxed (table)
	mainBoxStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(color)).Padding(0, 1)
	pw := m.paneWidth
	if pw > m.lastWidth {
		pw = m.lastWidth
	}
	if pw < 10 {
		pw = 10
	}

	// render content header as a centered small box above the table
	headerStyle := lipgloss.NewStyle().Width(pw).Align(lipgloss.Center).Bold(true).Foreground(lipgloss.Color(color))
	headerBox := headerStyle.Render(m.contentHeader)

	// combine compact header and content header into a single header stack so it's always visible
	headerStack := lipgloss.JoinVertical(lipgloss.Center, compactHeader, headerBox)

	mainBox := mainBoxStyle.Width(pw).Height(m.paneHeight).Render(m.table.View())

	// Footer: breadcrumb on the left, remote indicator right
	// build breadcrumb render
	crumbs := make([]string, 0, len(m.breadcrumb))
	for i, c := range m.breadcrumb {
		style := lipgloss.NewStyle()
		if i < len(m.breadcrumbStyles) {
			style = m.breadcrumbStyles[i]
		}
		// prefix with index hint [1]
		hint := fmt.Sprintf("%d", i+1)
		crumbs = append(crumbs, style.Render(fmt.Sprintf("[%s] <%s>", hint, c)))
	}
	breadcrumbRendered := strings.Join(crumbs, " ")

	remote := " "
	if m.flashActive {
		remote = "⚡"
	}

	// place breadcrumb left and remote right within lastWidth
	totalW := m.lastWidth
	leftPart := breadcrumbRendered
	rightPart := remote
	// compute padding
	padW := totalW - lipgloss.Width(leftPart) - lipgloss.Width(rightPart)
	if padW < 1 {
		padW = 1
	}
	spacer := strings.Repeat(" ", padW)
	footerLine := leftPart + spacer + rightPart

	// Compose final vertical layout: compactHeader (3 rows), contextSelectionBox (1 row), headerBox, mainBox, footerLine (1 row)
	baseView := lipgloss.JoinVertical(lipgloss.Left, headerStack, contextSelectionBox, mainBox, footerLine)

	// If modal is active, overlay it
	if m.activeModal == ModalConfirmDelete {
		modalOverlay := m.renderConfirmDeleteModal(m.lastWidth, m.lastHeight)
		return modalOverlay
	} else if m.activeModal == ModalEdit {
		return m.renderEditModal(m.lastWidth, m.lastHeight)
	} else if m.activeModal == ModalHelp {
		return m.renderHelpScreen(m.lastWidth, m.lastHeight)
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
	// Optionally write last rendered view to ./debug/last-screen.txt when debug enabled
	if m.debugEnabled {
		go func(s string) {
			_ = os.MkdirAll("./debug", 0755)
			_ = os.WriteFile("./debug/last-screen.txt", []byte(s), 0644)
		}(baseView)
	}
	return lipgloss.Place(w, h, lipgloss.Left, lipgloss.Top, baseView)
}

func rowInstanceID(row table.Row) string {
	if len(row) > 0 {
		return row[0]
	}
	return ""
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
	switch last {
	case "process-definitions":
		m.viewMode = "definitions"
		m.contentHeader = last
		m.selectedDefinitionKey = ""
		m.selectedInstanceID = ""
		return tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd())
	case "process-instances":
		if m.selectedDefinitionKey == "" {
			m.footerError = "No definition selected to show instances"
			return nil
		}
		m.viewMode = "instances"
		m.contentHeader = fmt.Sprintf("%s(%s)", m.currentRoot, m.selectedDefinitionKey)
		return tea.Batch(m.fetchInstancesCmd(m.selectedDefinitionKey), flashOnCmd())
	case "variables":
		if m.selectedInstanceID == "" {
			m.footerError = "No instance selected to show variables"
			return nil
		}
		m.viewMode = "variables"
		m.contentHeader = fmt.Sprintf("process-instances(%s)", m.selectedInstanceID)
		return tea.Batch(m.fetchVariablesCmd(m.selectedInstanceID), flashOnCmd())
	default:
		// treat as root context switch
		m.currentRoot = last
		m.viewMode = "definitions"
		m.contentHeader = last
		m.selectedDefinitionKey = ""
		m.selectedInstanceID = ""
		return tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd())
	}
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

func main() {
	debugMode := false
	for _, a := range os.Args[1:] {
		if a == "--debug" {
			debugMode = true
			break
		}
	}
	// Load split config files (o8n-env.yaml + o8n-cfg.yaml). No legacy fallback.
	cfg, err := config.LoadSplitConfig()
	if err != nil {
		log.Printf("Configuration error: %v", err)
		log.Printf("Please create 'o8n-env.yaml' (see o8n-env.yaml.example) and 'o8n-cfg.yaml' (table definitions).")
		return
	}

	if len(cfg.Environments) == 0 {
		log.Println("No environments configured. Please create 'o8n-env.yaml' and define at least one environment.")
		return
	}

	m := newModel(cfg)
	if debugMode {
		// ensure internal client picks up debug mode
		_ = os.Setenv("O8N_DEBUG", "1")
		m.debugEnabled = true
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatalf("failed to run program: %v", err)
	}
}
