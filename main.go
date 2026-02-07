package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const refreshInterval = 5 * time.Second

type refreshMsg struct{}

type dataLoadedMsg struct {
	definitions []ProcessDefinition
	instances   []ProcessInstance
}

type definitionsLoadedMsg struct{ definitions []ProcessDefinition }
type instancesLoadedMsg struct{ instances []ProcessInstance }

// New: fetch variables for a given instance
type variablesLoadedMsg struct{ variables []Variable }

type terminatedMsg struct{ id string }

type errMsg struct{ err error }

// Messages for the flash indicator
type flashOnMsg struct{}
type flashOffMsg struct{}

type processDefinitionItem struct {
	definition ProcessDefinition
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

type model struct {
	config *Config

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
	cachedDefinitions []ProcessDefinition

	// view mode: "definitions" or "instances"
	viewMode string

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
	envConfig *EnvConfig
	appConfig *AppConfig

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
}

func newModel(cfg *Config) model {
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
		config:     cfg,
		envNames:   envNames,
		currentEnv: current,
		list:       l,
		table:      t,
		viewMode:   "definitions",
	}
	m.applyStyle()
	// initialize lastListIndex
	m.lastListIndex = m.list.Index()

	// sensible defaults so the header is visible immediately
	m.lastWidth = 80
	m.paneWidth = m.lastWidth - 4
	m.paneHeight = 12

	// set root contexts and currentRoot
	m.rootContexts = loadRootContexts("resources/operaton-rest-api.json")
	if len(m.rootContexts) > 0 {
		m.currentRoot = "process-definitions"
	} else {
		m.currentRoot = "process-definitions"
	}

	return m
}

func newModelEnvApp(envCfg *EnvConfig, appCfg *AppConfig) model {
	// Build compatibility Config from split configs
	cfg := &Config{
		Environments: make(map[string]Environment),
		Active:       "",
		Tables:       appCfg.Tables,
	}
	for k, v := range envCfg.Environments {
		cfg.Environments[k] = Environment{
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

func (m *model) applyStyle() {
	color := ""
	if env, ok := m.config.Environments[m.currentEnv]; ok {
		color = env.UIColor
	}
	m.style = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(color)).Bold(true)

	listStyles := list.DefaultStyles()
	listStyles.Title = listStyles.Title.BorderForeground(lipgloss.Color(color))
	m.list.Styles = listStyles

	// Table header color is white as requested
	tStyles := table.DefaultStyles()
	tStyles.Header = tStyles.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(color)).BorderBottom(true).Foreground(lipgloss.Color("white")).Bold(true)
	tStyles.Selected = tStyles.Selected.Foreground(lipgloss.Color(color)).Bold(true)
	m.table.SetStyles(tStyles)
}

func (m model) Init() tea.Cmd {
	// fetch definitions at start and flash
	return tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd())
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
		if err := SaveConfig("config.yaml", m.config); err != nil {
			log.Printf("warning: failed to save config active environment: %v", err)
		}
	}
}

func (m *model) resetViews() {
	m.list.SetItems([]list.Item{})
	m.table.SetRows([]table.Row{})
	m.selectedInstanceID = ""
}

func (m *model) findTableDef(name string) *TableDef {
	if m.config == nil {
		return nil
	}
	for _, t := range m.config.Tables {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

// buildColumnsFor builds table.Column slice for a named table using the config definitions
// totalWidth is the available characters for the table content; if zero, returns reasonable defaults
func (m *model) buildColumnsFor(tableName string, totalWidth int) []table.Column {
	def := m.findTableDef(tableName)
	if def == nil {
		// fallback: reasonable default column
		return []table.Column{{Title: "COL", Width: 20}, {Title: "COL2", Width: 20}}
	}

	// collect visible columns
	visible := make([]ColumnDef, 0, len(def.Columns))
	for _, c := range def.Columns {
		if c.Visible {
			visible = append(visible, c)
		}
	}
	n := len(visible)
	if n == 0 {
		return []table.Column{{Title: "EMPTY", Width: 20}}
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

	// if totalWidth known, compute column widths
	cols := make([]table.Column, 0, n)
	if totalWidth <= 0 {
		// fallback widths
		for _, c := range visible {
			cols = append(cols, table.Column{Title: strings.ToUpper(c.Name), Width: 20})
		}
		return cols
	}

	used := 0
	for i, c := range visible {
		w := totalWidth * percentCols[i] / 100
		if w < 3 {
			w = 3
		}
		used += w
		cols = append(cols, table.Column{Title: strings.ToUpper(c.Name), Width: w})
	}
	// adjust for rounding differences
	if used < totalWidth {
		cols[len(cols)-1].Width += totalWidth - used
	}
	return cols
}

// Update applyDefinitions and applyInstances to recover from panics and show footer error
func (m *model) applyDefinitions(defs []ProcessDefinition) {
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
	cols := m.buildColumnsFor("process-definitions", m.paneWidth-4)
	if len(cols) == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), m.paneWidth-4)
	}
	colsCount := len(cols)
	if colsCount == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), m.paneWidth-4)
		colsCount = len(cols)
	}
	normRows := normalizeRows(rows, colsCount)
	m.table.SetColumns(cols)
	m.table.SetRows(normRows)
	m.viewMode = "definitions"
}

func (m *model) applyInstances(instances []ProcessInstance) {
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
	cols := m.buildColumnsFor("process-instances", m.paneWidth-4)
	if len(cols) == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), m.paneWidth-4)
	}
	colsCount := len(cols)
	if colsCount == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), m.paneWidth-4)
		colsCount = len(cols)
	}
	normRows := normalizeRows(rows, colsCount)
	m.table.SetColumns(cols)
	m.table.SetRows(normRows)
	m.viewMode = "instances"
}

// New: variables table
func (m *model) applyVariables(vars []Variable) {
	defer func() {
		if r := recover(); r != nil {
			m.footerError = fmt.Sprintf("Error loading variables: %v", r)
			log.Printf("variables panic recovered: %v", r)
		}
	}()

	rows := make([]table.Row, 0, len(vars))
	for _, v := range vars {
		rows = append(rows, table.Row{v.Name, v.Value})
	}
	cols := m.buildColumnsFor("process-variables", m.paneWidth-4)
	if len(cols) == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), m.paneWidth-4)
	}
	colsCount := len(cols)
	if colsCount == 0 && len(rows) > 0 {
		cols = defaultColumns(len(rows[0]), m.paneWidth-4)
		colsCount = len(cols)
	}
	normRows := normalizeRows(rows, colsCount)
	m.table.SetColumns(cols)
	m.table.SetRows(normRows)
	m.viewMode = "variables"
}

func (m model) fetchDefinitionsCmd() tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	client := NewClient(env)
	return func() tea.Msg {
		defs, err := client.FetchProcessDefinitions()
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
	client := NewClient(env)
	return func() tea.Msg {
		instances, err := client.FetchInstances(processKey)
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
	client := NewClient(env)
	return func() tea.Msg {
		vars, err := client.FetchVariables(instanceID)
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

	client := NewClient(env)

	return func() tea.Msg {
		defs, err := client.FetchProcessDefinitions()
		if err != nil {
			return errMsg{err}
		}

		var instances []ProcessInstance
		if selectedKey != "" {
			instances, err = client.FetchInstances(selectedKey)
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
	client := NewClient(env)

	return func() tea.Msg {
		if err := client.TerminateInstance(id); err != nil {
			return errMsg{err}
		}
		return terminatedMsg{id: id}
	}
}

// helper to create a command that triggers the flash indicator
func flashOnCmd() tea.Cmd {
	return func() tea.Msg { return flashOnMsg{} }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "e":
			m.nextEnvironment()
			m.resetViews()
			// refetch definitions for new env and flash
			return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd())
		case "r":
			m.autoRefresh = !m.autoRefresh
			if m.autoRefresh {
				return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} }))
			}
			return m, nil
		case "x":
			if m.selectedInstanceID != "" {
				m.showKillModal = true
			}
			return m, nil
		case "esc":
			if m.showRootPopup {
				m.showRootPopup = false
				m.rootInput = ""
				return m, nil
			}
			if m.viewMode == "variables" {
				// back to instances
				m.viewMode = "instances"
				return m, nil
			}
			if m.viewMode == "instances" {
				// go back to definitions view and flash
				m.viewMode = "definitions"
				return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd())
			}
			m.showKillModal = false
			return m, nil
		case "y":
			if m.showKillModal && m.selectedInstanceID != "" {
				m.showKillModal = false
				return m, tea.Batch(m.terminateInstanceCmd(m.selectedInstanceID), flashOnCmd())
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
						return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd())
					}
				}
				// no exact match: ignore
				return m, nil
			}
			if m.viewMode == "definitions" {
				// get selected definition key from table
				row := m.table.SelectedRow()
				if len(row) > 0 {
					key := fmt.Sprintf("%v", row[0])
					m.viewMode = "instances"
					return m, tea.Batch(m.fetchInstancesCmd(key), flashOnCmd())
				}
			} else if m.viewMode == "instances" {
				// drill down into variables for the selected instance
				row := m.table.SelectedRow()
				if len(row) > 0 {
					id := fmt.Sprintf("%v", row[0])
					m.selectedInstanceID = id
					m.viewMode = "variables"
					return m, tea.Batch(m.fetchVariablesCmd(id), flashOnCmd())
				}
			} else if m.viewMode == "variables" {
				// no deeper drilldown
			}
			return m, nil
		case ":":
			if !m.showRootPopup {
				m.showRootPopup = true
				m.rootInput = ""
				m.footerError = ""
			} else {
				m.showRootPopup = false
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
		default:
			// typing into the root input when popup active
			if m.showRootPopup {
				r := msg.String()
				// append single rune only
				if len(r) == 1 {
					m.rootInput += r
				}
				return m, nil
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		// Resize UI to take full terminal size
		width := msg.Width
		height := msg.Height
		// store terminal size for View footer alignment
		m.lastWidth = width
		m.lastHeight = height
		// Reserve lines: header area (3 columns) 8 rows + footer 1 line + context selection 1 line
		headerLines := 8
		contextSelectionLines := 1
		footerLines := 1
		contentHeight := height - headerLines - contextSelectionLines - footerLines
		if contentHeight < 3 {
			contentHeight = 3
		}
		// pane widths: full width minus some margins
		m.paneWidth = width - 4
		// reduce to avoid table overflow
		m.paneHeight = contentHeight
		m.list.SetSize((width/4)-2, contentHeight-1)
		m.table.SetWidth(width - (width / 4) - 4)
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

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	m.selectedInstanceID = rowInstanceID(m.table.SelectedRow())

	return m, tea.Batch(cmds...)
}

// Backwards-compatible wrapper used by tests: applies both definitions and instances
func (m *model) applyData(defs []ProcessDefinition, instances []ProcessInstance) {
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
	return `  ____   ___  _  _\n / __ \ / _ \| \| |\n| |  | | | | |  \|\n| |  | | | | | . ` + "`" + ` |\n| |__| | |_| | |\  |\n \____/ \___/|_| \_|\n` + "o8n"
}

func (m model) View() string {
	// top unboxed area with 3 columns: context, keybindings, ascii art (8 rows)
	envInfo := "Environment: " + m.currentEnv
	apiURL := ""
	username := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			apiURL = env.URL
			username = env.Username
		}
	}
	ctxCol := fmt.Sprintf("%s\nURL: %s\nUser: %s", envInfo, apiURL, username)

	helpCol := "KEYS:\nq - quit\ne - switch env\nr - toggle auto-refresh\nenter - drill into instances\nesc - back\nx - kill instance\n"

	artCol := m.asciiArt()

	// compute logo and column widths with clamping
	totalW := m.lastWidth
	if totalW < 40 {
		totalW = 40
	}
	logoW := 25
	if totalW < 80 {
		logoW = totalW / 4
		if logoW < 12 {
			logoW = 12
		}
	}
	rem := totalW - logoW
	if rem < 20 {
		rem = 20
	}
	colW := rem / 2
	if colW < 12 {
		colW = 12
	}

	left := lipgloss.Place(colW, 8, lipgloss.Left, lipgloss.Top, ctxCol)
	middle := lipgloss.Place(colW, 8, lipgloss.Left, lipgloss.Top, helpCol)
	right := lipgloss.Place(logoW, 8, lipgloss.Right, lipgloss.Top, artCol)

	top := lipgloss.JoinHorizontal(lipgloss.Top, left, middle, right)
	// ensure top is visible as 8 rows even if no WindowSizeMsg yet
	topBox := lipgloss.NewStyle().Width(m.lastWidth).Height(8).Render(top)

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
		boxStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(color)).
			Width(m.lastWidth-4).
			Padding(0, 1)
		contextSelectionBox = boxStyle.Render(" ")
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
	mainBox := mainBoxStyle.Width(pw).Height(m.paneHeight).Render(m.table.View())

	// Footer: three columns: context | message | remote (1 char)
	// compute column widths
	sep := " | "
	remote := " "
	if m.flashActive {
		remote = "⚡"
	}

	// message from footerError otherwise empty
	message := m.footerError

	// reserve remote column + separators
	remWidth := m.lastWidth - 3 - 1 // two separators (" | ") occupy 6? but separators length total 6 characters; we will compute precisely
	// simpler: compute contextCol as 60% of width
	ctxColW := int(float64(m.lastWidth) * 0.6)
	if ctxColW < len(m.currentRoot)+2 {
		ctxColW = len(m.currentRoot) + 2
	}
	// ensure we have at least space for message and remote
	if ctxColW > m.lastWidth-10 {
		ctxColW = m.lastWidth - 10
	}
	// remaining for message and remote and separators
	remaining := m.lastWidth - ctxColW - len(sep) - 1 - len(sep)
	if remaining < 10 {
		remaining = 10
	}
	msgColW := remaining

	// truncate context and message
	ctxDisplay := fmt.Sprintf("<%s>", m.currentRoot)
	if lipgloss.Width(ctxDisplay) > ctxColW {
		// truncate
		ctxDisplay = ctxDisplay[:ctxColW-1]
	}
	if len(message) > msgColW {
		message = message[:msgColW-1]
	}

	footerLine := ctxDisplay + sep + message + sep + remote

	// Compose final vertical layout: topBox (8 rows), contextSelectionBox (1 row), mainBox, footerLine (1 row)
	return lipgloss.JoinVertical(lipgloss.Left, topBox, contextSelectionBox, mainBox, footerLine)
}

func rowInstanceID(row table.Row) string {
	if len(row) > 0 {
		return row[0]
	}
	return ""
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
		return []string{"process-definitions", "process-instances", "process-variables", "task", "job", "external-task"}
	}
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return []string{"process-definitions", "process-instances", "process-variables", "task", "job", "external-task"}
	}
	pathsI, ok := doc["paths"]
	if !ok {
		return []string{"process-definitions", "process-instances", "process-variables", "task", "job", "external-task"}
	}
	pathsMap, ok := pathsI.(map[string]interface{})
	if !ok {
		return []string{"process-definitions", "process-instances", "process-variables", "task", "job", "external-task"}
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
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Note: Could not load config.yaml: %v", err)
		log.Printf("You can create a config.yaml based on config.yaml.example")
		return
	}

	if len(cfg.Environments) == 0 {
		log.Println("No environments configured in config.yaml")
		return
	}

	if _, err := tea.NewProgram(newModel(cfg)).Run(); err != nil {
		log.Fatalf("failed to run program: %v", err)
	}
}
