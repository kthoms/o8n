package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
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
	"github.com/charmbracelet/x/ansi"

	"github.com/kthoms/o8n/internal/client"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/contentassist"
	"github.com/kthoms/o8n/internal/dao"
	"github.com/kthoms/o8n/internal/validation"
	"golang.org/x/term"
)

const (
	refreshInterval = 5 * time.Second
	appVersion      = "0.1.0"
)

// EnvironmentStatus represents connection status to an environment
type EnvironmentStatus string

const (
	StatusUnknown     EnvironmentStatus = "unknown"
	StatusOperational EnvironmentStatus = "operational"
	StatusUnreachable EnvironmentStatus = "unreachable"
)

// envStatusMsg reports health check result for an environment
type envStatusMsg struct {
	env    string
	status EnvironmentStatus
	err    error
}

// Package-level style constants used in View — cached here to avoid per-frame allocations.
var (
	// completionStyle renders the greyed-out autocomplete ghost text in the root popup.
	completionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	// flashBaseStyle is the fixed-width right-aligned base for the flash/remote indicator.
	flashBaseStyle = lipgloss.NewStyle().Width(3).Align(lipgloss.Right)
	// Feedback status styles
	errorFooterStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
	successFooterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50C878")).Bold(true)
	infoFooterStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00A8E1"))
	loadingFooterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	// Validation error style for edit modal
	validationErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
)

type refreshMsg struct{}

type dataLoadedMsg struct {
	definitions []config.ProcessDefinition
	instances   []config.ProcessInstance
}

type definitionsLoadedMsg struct {
	definitions []config.ProcessDefinition
	count       int
}
type instancesLoadedMsg struct{ instances []config.ProcessInstance }

// New: fetch variables for a given instance
type variablesLoadedMsg struct{ variables []config.Variable }

// genericLoadedMsg carries arbitrary collection data fetched from an endpoint
type genericLoadedMsg struct {
	root  string
	items []map[string]interface{}
}

type instancesWithCountMsg struct {
	instances []config.ProcessInstance
	count     int
}

type terminatedMsg struct{ id string }
type suspendedMsg struct{ id string }
type resumedMsg struct{ id string }
type retriedMsg struct{ id string }

// actionExecutedMsg is sent when a config-driven action completes successfully
type actionExecutedMsg struct {
	label string // the action label for feedback
}

type errMsg struct{ err error }

// Messages for the flash indicator
type flashOnMsg struct{}
type flashOffMsg struct{}

// message used to clear footer errors
type clearErrorMsg struct{}

// clearPendingGMsg resets the pendingG flag after timeout
type clearPendingGMsg struct{}

// footerStatusKind represents the type of feedback message in the footer
type footerStatusKind int

const (
	footerStatusNone    footerStatusKind = iota
	footerStatusError                    // red ✗
	footerStatusSuccess                  // green ✓
	footerStatusInfo                     // blue ℹ
	footerStatusLoading                  // yellow ⟳
)

// splash done message
type splashDoneMsg struct{}

// splash frame message for animation
type splashFrameMsg struct{ frame int }

const totalSplashFrames = 15

// setFooterStatus returns values to set a footer message with a kind and schedules auto-clear.
// clearAfter == 0 means no auto-clear (for loading states).
func setFooterStatus(kind footerStatusKind, msg string, clearAfter time.Duration) (string, footerStatusKind, tea.Cmd) {
	var cmd tea.Cmd
	if clearAfter > 0 {
		cmd = tea.Tick(clearAfter, func(time.Time) tea.Msg { return clearErrorMsg{} })
	}
	return msg, kind, cmd
}

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
	ModalSort
	ModalDetailView
	ModalEnvironment
)

// editFocusArea tracks keyboard focus within the edit modal
type editFocusArea int

const (
	editFocusInput  editFocusArea = iota // text input focused
	editFocusSave                         // Save button focused
	editFocusCancel                       // Cancel button focused
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

// actionItem represents a context-specific action in the actions menu
type actionItem struct {
	key   string // single-key shortcut
	label string // display label
	cmd   func(m *model) tea.Cmd
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
	genericParams         map[string]string // drilldown filter params active at this level
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

	// Pending config-driven action awaiting confirmation
	pendingAction     *config.ActionDef // action definition awaiting confirm
	pendingActionID   string            // resolved ID for the pending action
	pendingActionPath string            // resolved path for the pending action

	// Edit modal state
	editInput     textinput.Model
	editColumns   []editableColumn
	editColumnPos int
	editRowIndex  int
	editTableKey  string
	editError     string
	editFocus     editFocusArea

	// Search/filter state
	searchMode      bool
	searchInput     textinput.Model
	searchTerm      string
	filteredRows    []table.Row
	originalRows    []table.Row

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
	// footer status kind (error, success, info, loading)
	footerStatusKind footerStatusKind
	// isLoading indicates that an async API call is in progress
	isLoading bool
	// lastAPILatency tracks the most recent API call duration
	lastAPILatency time.Duration
	// apiCallStarted records when the current API call started
	apiCallStarted time.Time

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

	// pagination state per root collection: first result offset and known total count
	pageOffsets map[string]int
	pageTotals  map[string]int
	// requested cursor position to restore after page load (keeps selection stable across pages)
	pendingCursorAfterPage int

	// active filter params for the current generic collection view (e.g. drilldown filters)
	genericParams map[string]string

	// Version number
	version      string
	debugEnabled bool

	// Environment connection status tracking
	envStatus map[string]EnvironmentStatus

	// Vim navigation: tracks first 'g' press for gg sequence
	pendingG bool

	// Sort state
	sortColumn      int // index into visible columns, -1 = unsorted
	sortAscending   bool
	sortPopupCursor int

	// Actions menu state
	showActionsMenu   bool
	actionsMenuItems  []actionItem
	actionsMenuCursor int

	// Detail viewer state
	detailContent   string
	detailScroll    int
	detailMaxScroll int

	// Environment popup state
	showEnvPopup   bool
	envPopupCursor int

	activeSkin string
	skin       *Skin
}

// getKeyHints returns context-aware keyboard hints based on current view and terminal width
func (m *model) getKeyHints(width int) []KeyHint {
	hints := []KeyHint{}

	// Global hints (always relevant)
	hints = append(hints,
		KeyHint{"?", "help", 1},
		KeyHint{":", "switch", 2},
	)

	// Context-specific hints based on viewMode
	if m.viewMode == "definitions" {
		hints = append(hints,
			KeyHint{"↑↓", "nav", 3},
			KeyHint{"Enter", "drill", 4},
		)
		// Drill-down drilldown hint
		if width >= 85 {
			hints = append(hints, KeyHint{"e", "Edit def", 4})
		}
	} else if m.viewMode == "instances" {
		hints = append(hints,
			KeyHint{"Esc", "back", 5},
			KeyHint{"↑↓", "nav", 3},
			KeyHint{"Enter", "vars", 4},
		)
		// Terminate instance hint in instances view
		if width >= 100 {
			hints = append(hints, KeyHint{"<ctrl>+d", "terminate", 7})
		}
	} else if m.viewMode == "variables" {
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
	if width >= 90 {
		hints = append(hints, KeyHint{"<ctrl>-r", "refresh", 6})
	}
	hints = append(hints, KeyHint{"PgDn/PgUp", "page", 3})
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

	// Row 1: Status line (environment with status indicator instead of URL/user)
	status, ok := m.envStatus[m.currentEnv]
	if !ok {
		status = StatusUnknown
	}

	// Determine status symbol and color
	statusSymbol := "●"
	statusColor := lipgloss.Color("#FFAA00") // yellow/orange for unknown
	switch status {
	case StatusOperational:
		statusSymbol = "●"
		statusColor = lipgloss.Color("#00FF00") // green
	case StatusUnreachable:
		statusSymbol = "✗"
		statusColor = lipgloss.Color("#FF0000") // red
	case StatusUnknown:
		statusSymbol = "○"
		statusColor = lipgloss.Color("#FFAA00") // yellow
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	envInfo := fmt.Sprintf("%s %s", m.currentEnv, statusStyle.Render(statusSymbol))

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

	// Render header using default terminal colors (no forced background/foreground).
	headerStyle := lipgloss.NewStyle().Width(width).Padding(0, 1).Bold(true)
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
		config:                 cfg,
		envNames:               envNames,
		currentEnv:             current,
		list:                   l,
		table:                  t,
		viewMode:               "definitions",
		splashActive:           true,
		splashFrame:            1,
		activeModal:            ModalNone,
		version:                appVersion,
		variablesByName:        map[string]config.Variable{},
		debugEnabled:           false,
		pageOffsets:            make(map[string]int),
		pageTotals:             make(map[string]int),
		pendingCursorAfterPage: -1,
		genericParams:          make(map[string]string),
		envStatus:              make(map[string]EnvironmentStatus),
		sortColumn:             -1,
	}

	// edit input defaults
	editInput := textinput.New()
	editInput.Placeholder = "value"
	editInput.Prompt = "> "
	editInput.CharLimit = 0
	editInput.Width = 40
	m.editInput = editInput

	// Initialize search input
	searchInput := textinput.New()
	searchInput.Placeholder = "search..."
	searchInput.Prompt = "/"
	searchInput.CharLimit = 0
	searchInput.Width = 40
	m.searchInput = searchInput

	m.applyStyle()
	// initialize lastListIndex
	m.lastListIndex = m.list.Index()

	// initialize breadcrumb: start with current root
	m.breadcrumb = []string{dao.ResourceProcessDefinitions}
	m.contentHeader = dao.ResourceProcessDefinitions

	// initialize environment status to unknown
	for _, envName := range envNames {
		m.envStatus[envName] = StatusUnknown
	}

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
	footerLines := 2  // breadcrumb line + status line
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

func newModelEnvApp(envCfg *config.EnvConfig, appCfg *config.AppConfig, skinName string) model {
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
	cfg.Skin = envCfg.Skin

	m := newModel(cfg)
	m.envConfig = envCfg
	m.appConfig = appCfg
	m.activeSkin = skinName

	skin, err := loadSkin(skinName)
	if err != nil {
		log.Printf("Falling back to stock skin. Could not load skin: %v", err)
		skin, _ = loadSkin("stock.yaml")
	}
	m.skin = skin

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

	// Return just the styled box — overlayCenter handles centering
	return modal
}

// renderHelpScreen renders the help screen modal
func (m *model) renderHelpScreen(width, height int) string {
	helpContent := `o8n Help

NAVIGATION              │  ACTIONS                │  GLOBAL
─────────────────────   │  ────────────────────   │  ──────────────────
↑/↓/j/k  Navigate list  │  <ctrl>+e  Switch env   │  ?     This help
PgUp/Dn  Page up/down   │  <ctrl>-r  Auto-refresh │  :     Switch view
gg/G     Top/bottom     │  <ctrl>+d  Delete item  │  <ctrl>+c Quit
Ctrl+d/u Half-page      │  <ctrl>-R  Force refresh│
Enter    Drill down     │  Space     Actions menu │  CONTEXT
Esc      Go back        │  VIEW SPECIFIC          │  ───────────────────
                        │  (varies by view)       │  Tab      Complete
SEARCH                  │                         │  Enter    Confirm
─────────────────────   │  In Process Instances:  │  Esc      Cancel
/        Search/filter  │  v  View variables      │  s        Sort
Esc      Clear filter   │  <ctrl>+d Kill instance │  y        Detail view
Enter    Lock filter    │  s  Suspend instance    │
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
		columnsLine = strings.Join(parts, " ") + "\n\n"
	}

	errorLine := ""
	if m.editError != "" {
		errorLine = validationErrorStyle.Render("⚠ " + m.editError) + "\n\n"
	}

	// Build body (without header)
	body := columnsLine + m.editInput.View() + "\n\n" + errorLine

	// Determine button styles from config or defaults
	saveBg := "#00A8E1"
	saveFg := "#FFFFFF"
	cancelBg := "#666666"
	cancelFg := "#FFFFFF"
	if m.config != nil && m.config.UI != nil && m.config.UI.EditModal != nil && m.config.UI.EditModal.Buttons != nil {
		if b := m.config.UI.EditModal.Buttons.Save; b != nil {
			if b.Background != "" {
				saveBg = b.Background
			}
			if b.Foreground != "" {
				saveFg = b.Foreground
			}
		}
		if b := m.config.UI.EditModal.Buttons.Cancel; b != nil {
			if b.Background != "" {
				cancelBg = b.Background
			}
			if b.Foreground != "" {
				cancelFg = b.Foreground
			}
		}
	}

	saveStyle := lipgloss.NewStyle().Background(lipgloss.Color(saveBg)).Foreground(lipgloss.Color(saveFg)).Padding(0, 1).Bold(true)
	cancelStyle := lipgloss.NewStyle().Background(lipgloss.Color(cancelBg)).Foreground(lipgloss.Color(cancelFg)).Padding(0, 1)
	disabledSaveStyle := lipgloss.NewStyle().Background(lipgloss.Color("#777777")).Foreground(lipgloss.Color("#DDDDDD")).Padding(0, 1)

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
				errorLine = validationErrorStyle.Render("⚠ " + err.Error()) + "\n\n"
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
	body = columnsLine + m.editInput.View() + "\n\n" + errorLine + suggestionLine

	// Render buttons with focus styles
	savedFocusedStyle := saveStyle.Copy().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#FFFFFF"))
	cancelFocusedStyle := cancelStyle.Copy().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#FFFFFF"))

	var saveBtn, cancelBtn string
	switch m.editFocus {
	case editFocusSave:
		if saveDisabled {
			saveBtn = disabledSaveStyle.Copy().Border(lipgloss.NormalBorder()).Render(saveLabel)
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

	// Border color: prefer environment color if available
	borderColor := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			borderColor = env.UIColor
		}
		if m.config.UI != nil && m.config.UI.EditModal != nil && m.config.UI.EditModal.BorderColor != "" {
			borderColor = m.config.UI.EditModal.BorderColor
		}
	}

	modalStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(borderColor)).Padding(1, 2)
	if m.config != nil && m.config.UI != nil && m.config.UI.EditModal != nil && m.config.UI.EditModal.Width > 0 {
		modalStyle = modalStyle.Width(m.config.UI.EditModal.Width)
	} else {
		modalStyle = modalStyle.Width(60)
	}

	modal := modalStyle.Render(modalBody)
	// Return just the styled box — overlayCenter handles centering
	return modal
}

func (m *model) applyStyle() {
	log.Printf("DEBUG: applyStyle called. m.skin is nil: %t", m.skin == nil)
	if m.skin != nil {
		m.style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.skin.O8n.Body.FgColor)).
			Background(lipgloss.Color(m.skin.O8n.Body.BgColor))

		listStyles := list.DefaultStyles()
		listStyles.Title = listStyles.Title.BorderForeground(lipgloss.Color(m.skin.O8n.Frame.Border.FocusColor))
		m.list.Styles = listStyles

		tStyles := table.DefaultStyles()
		tStyles.Header = tStyles.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(m.skin.O8n.Frame.Border.FgColor)).
			BorderBottom(false).
			Foreground(lipgloss.Color(m.skin.O8n.Body.FgColor)).
			Bold(true)
		tStyles.Selected = tStyles.Selected.
			Foreground(lipgloss.Color(m.skin.O8n.Body.BgColor)).
			Background(lipgloss.Color(m.skin.O8n.Frame.Border.FocusColor)).
			Bold(true)
		m.table.SetStyles(tStyles)

		m.splashLogoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(m.skin.O8n.Body.LogoColor)).Bold(true).Align(lipgloss.Center)
		m.splashInfoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(m.skin.O8n.Body.LogoColor)).Align(lipgloss.Center)

		m.breadcrumbStyles = []lipgloss.Style{
			lipgloss.NewStyle().Background(lipgloss.Color(m.skin.O8n.Frame.Border.FocusColor)).Foreground(lipgloss.Color(m.skin.O8n.Body.BgColor)).Padding(0, 1),
			lipgloss.NewStyle().Background(lipgloss.Color("#e6e6fa")).Foreground(lipgloss.Color("black")).Padding(0, 1),
			lipgloss.NewStyle().Background(lipgloss.Color("#f0fff0")).Foreground(lipgloss.Color("black")).Padding(0, 1),
			lipgloss.NewStyle().Background(lipgloss.Color("#fffaf0")).Foreground(lipgloss.Color("black")).Padding(0, 1),
		}
	} else {
		// Fallback to old styling if skin is not loaded
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

		tStyles := table.DefaultStyles()
		tStyles.Header = tStyles.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(color)).BorderBottom(false).Foreground(lipgloss.Color("white")).Bold(true)
		bgColor := m.deriveFocusBackgroundColor(color)
		tStyles.Selected = tStyles.Selected.
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color(bgColor)).
			Bold(true)
		m.table.SetStyles(tStyles)

		m.splashLogoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Align(lipgloss.Center)
		m.splashInfoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Align(lipgloss.Center)

		m.breadcrumbStyles = []lipgloss.Style{
			lipgloss.NewStyle().Background(lipgloss.Color(color)).Foreground(lipgloss.Color("black")).Padding(0, 1),
			lipgloss.NewStyle().Background(lipgloss.Color("#e6e6fa")).Foreground(lipgloss.Color("black")).Padding(0, 1),
			lipgloss.NewStyle().Background(lipgloss.Color("#f0fff0")).Foreground(lipgloss.Color("black")).Padding(0, 1),
			lipgloss.NewStyle().Background(lipgloss.Color("#fffaf0")).Foreground(lipgloss.Color("black")).Padding(0, 1),
		}
	}
}
func (m model) Init() tea.Cmd {
	// fetch definitions at start and flash, and start splash animation (150ms per frame, total 1.5s)
	// we start with frame 1 already set; schedule frame 2 after 150ms
	firstTick := tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return splashFrameMsg{frame: 2} })

	// Check health of all environments
	cmds := []tea.Cmd{m.fetchDefinitionsCmd(), flashOnCmd(), firstTick}
	for _, envName := range m.envNames {
		cmds = append(cmds, m.checkEnvironmentHealthCmd(envName))
	}

	return tea.Batch(cmds...)
}

func (m *model) nextEnvironment() tea.Cmd {
	if len(m.envNames) == 0 {
		return nil
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
	// Check health of the newly selected environment
	return m.checkEnvironmentHealthCmd(m.currentEnv)
}

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

// executeActionCmd creates a command that performs a config-driven REST action.
func (m model) executeActionCmd(action config.ActionDef, resolvedPath string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env)
	label := action.Label
	return func() tea.Msg {
		if err := c.ExecuteAction(action.Method, resolvedPath, action.Body); err != nil {
			return errMsg{err}
		}
		return actionExecutedMsg{label: label}
	}
}

// buildActionsForRoot returns context-specific action items for the current root resource.
// Actions are loaded from the config file's table definitions. A "View as JSON" action
// is always appended as the last item.
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

// suspendInstanceCmd creates a command to suspend/resume a process instance.
func (m model) suspendInstanceCmd(id string, suspend bool) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env)
	return func() tea.Msg {
		if err := c.SuspendProcessInstance(id, suspend); err != nil {
			return errMsg{err}
		}
		if suspend {
			return suspendedMsg{id: id}
		}
		return resumedMsg{id: id}
	}
}

// setJobRetriesCmd creates a command to set retries on a job.
func (m model) setJobRetriesCmd(id string, retries int) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	c := client.NewClient(env)
	return func() tea.Msg {
		if err := c.SetJobRetries(id, retries); err != nil {
			return errMsg{err}
		}
		return retriedMsg{id: id}
	}
}

// switchToEnvironment switches to the named environment (extracted from cycling logic).
func (m *model) switchToEnvironment(name string) {
	m.currentEnv = name
	m.applyStyle()
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
		if !c.IsVisible() {
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
		return []table.Column{{Title: "COL", Width: 20}, {Title: "COL2", Width: 20}}
	}

	// collect all columns that are not explicitly hidden, preserving config order
	type colEntry struct {
		def      config.ColumnDef
		title    string
		width    int // desired width
		minWidth int // minimum width (= header title length if not configured)
	}

	entries := make([]colEntry, 0, len(def.Columns))
	for _, c := range def.Columns {
		if !c.IsVisible() {
			continue
		}
		title := strings.ToUpper(c.Name)
		if c.Editable {
			title = title + " 🖍️"
		}
		headerLen := lipgloss.Width(title)

		desired := c.Width
		if desired == 0 {
			desired = config.DefaultTypeWidth(c.Type)
		}

		minW := c.MinWidth
		if minW == 0 {
			minW = headerLen // implicit minimum = column header title width
		}
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
	for _, d := range defs {
		items = append(items, processDefinitionItem{definition: d})
		// Add drilldown prefix to first column for focus indicator
		rows = append(rows, table.Row{"▶ " + d.Key, d.Name, fmt.Sprintf("%d", d.Version), d.Resource})
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
	m.table.SetColumns(cols)
	m.table.SetRows(normRows)
	m.viewMode = "definitions"
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
	for _, inst := range instances {
		// Add drilldown prefix to first column for focus indicator
		rows = append(rows, table.Row{"▶ " + inst.ID, inst.DefinitionID, inst.BusinessKey, inst.StartTime})
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
	m.table.SetColumns(cols)
	m.table.SetRows(normRows)
	m.viewMode = "instances"
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
	normRows := normalizeRows(rows, colsCount)
	m.table.SetColumns(cols)
	m.table.SetRows(normRows)
	m.viewMode = "variables"
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
		// Fetch count separately (non-fatal if it fails)
		count := 0
		if countVal, err := c.FetchProcessDefinitionsCount(); err == nil {
			count = countVal
		}
		return definitionsLoadedMsg{definitions: defs, count: count}
	}
}

func (m model) fetchInstancesCmd(paramName, paramValue string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	// Use paged HTTP fetch similar to generic fetch so we can pass firstResult/maxResults
	return func() tea.Msg {
		// If caller asked for a definition id but provided a key, resolve it from cache
		value := paramValue
		if paramName == "processDefinitionId" {
			// try to map key -> id when possible
			for _, d := range m.cachedDefinitions {
				if d.Key == paramValue {
					value = d.ID
					break
				}
			}
		}

		base := strings.TrimRight(env.URL, "/")
		// API uses singular path for instances endpoint
		urlStr := base + "/process-instance"
		// add filter param if provided
		q := ""
		if paramName != "" && value != "" {
			q = fmt.Sprintf("%s=%s", paramName, value)
		}
		offset := 0
		if v, ok := m.pageOffsets[dao.ResourceProcessInstances]; ok {
			offset = v
		}
		limit := m.getPageSize()
		if q != "" {
			urlStr = urlStr + "?" + q + fmt.Sprintf("&firstResult=%d&maxResults=%d", offset, limit)
		} else {
			urlStr = urlStr + fmt.Sprintf("?firstResult=%d&maxResults=%d", offset, limit)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			return errMsg{err}
		}
		req.Header.Set("Accept", "application/json")
		if env.Username != "" {
			req.SetBasicAuth(env.Username, env.Password)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			data, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("failed to fetch instances: %s", string(data))}
		}
		var items []map[string]interface{}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&items); err != nil {
			return errMsg{err}
		}

		// Try to load count
		count := -1
		countURL := base + "/process-instance/count"
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		req2, err2 := http.NewRequestWithContext(ctx2, http.MethodGet, countURL, nil)
		if err2 == nil {
			req2.Header.Set("Accept", "application/json")
			if env.Username != "" {
				req2.SetBasicAuth(env.Username, env.Password)
			}
			if resp2, err2 := http.DefaultClient.Do(req2); err2 == nil {
				defer resp2.Body.Close()
				if resp2.StatusCode < 400 {
					var cntBody map[string]interface{}
					dec2 := json.NewDecoder(resp2.Body)
					if err3 := dec2.Decode(&cntBody); err3 == nil {
						if v, ok := cntBody["count"]; ok {
							if n, ok := v.(float64); ok {
								count = int(n)
							}
						}
					}
				}
			}
		}

		// Convert items to typed instances
		instances := make([]config.ProcessInstance, 0, len(items))
		for _, it := range items {
			pi := config.ProcessInstance{}
			if v, ok := it["id"]; ok {
				pi.ID = fmt.Sprintf("%v", v)
			}
			if v, ok := it["processDefinitionId"]; ok {
				pi.DefinitionID = fmt.Sprintf("%v", v)
			}
			if v, ok := it["businessKey"]; ok {
				pi.BusinessKey = fmt.Sprintf("%v", v)
			}
			if v, ok := it["startTime"]; ok {
				pi.StartTime = fmt.Sprintf("%v", v)
			}
			instances = append(instances, pi)
		}

		// return typed instances message (tests expect this path)
		// Note: count handling for instances is not propagated in this message.
		return instancesWithCountMsg{instances: instances, count: count}
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
			instances, err = c.FetchInstances("processDefinitionKey", selectedKey)
			if err != nil {
				return errMsg{err}
			}
		}

		return dataLoadedMsg{definitions: defs, instances: instances}
	}
}

// fetchForRoot returns a command that fetches data for the given root resource.
// It maps known root names to their corresponding fetch commands.
func (m model) fetchForRoot(root string) tea.Cmd {
	// If we know the common ones, handle them specially
	switch root {
	case dao.ResourceProcessDefinitions:
		return m.fetchDefinitionsCmd()
	case dao.ResourceProcessInstances:
		// fetch all instances (no param)
		return m.fetchInstancesCmd("", "")
	case dao.ResourceProcessVariables:
		// no sensible root-level fetch for variables without an instance id
		return nil
	default:
		// If we have a table definition, attempt a generic fetch of the collection
		if def := m.findTableDef(root); def != nil {
			return m.fetchGenericCmd(root)
		}
		// fallback: try definitions
		return m.fetchDefinitionsCmd()
	}
}

// fetchGenericCmd performs a GET to the environment server for the provided
// collection resource (root) and returns a genericLoadedMsg with the parsed
// JSON array of objects.
func (m model) fetchGenericCmd(root string) tea.Cmd {
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	// ensure paging defaults
	if m.pageOffsets == nil {
		m.pageOffsets = make(map[string]int)
	}
	if m.pageTotals == nil {
		m.pageTotals = make(map[string]int)
	}

	// Resolve the API path: rootContexts pluralizes names (e.g. "task" → "tasks")
	// but the Operaton REST API uses the singular form. Use the table def name when
	// available since it mirrors the actual API path segment.
	apiPath := root
	if def := m.findTableDef(root); def != nil {
		apiPath = def.Name
	}

	// Copy active filter params for thread-safe use inside the goroutine.
	paramsCopy := make(map[string]string, len(m.genericParams))
	for k, v := range m.genericParams {
		paramsCopy[k] = v
	}

	return func() tea.Msg {
		base := strings.TrimRight(env.URL, "/")
		offset := 0
		if v, ok := m.pageOffsets[root]; ok {
			offset = v
		}
		limit := m.getPageSize()
		urlStr := base + "/" + strings.TrimLeft(apiPath, "/")
		// append drilldown filter params, then paging params
		for k, v := range paramsCopy {
			if strings.Contains(urlStr, "?") {
				urlStr = fmt.Sprintf("%s&%s=%s", urlStr, k, v)
			} else {
				urlStr = fmt.Sprintf("%s?%s=%s", urlStr, k, v)
			}
		}
		if limit > 0 {
			if strings.Contains(urlStr, "?") {
				urlStr = fmt.Sprintf("%s&firstResult=%d&maxResults=%d", urlStr, offset, limit)
			} else {
				urlStr = urlStr + fmt.Sprintf("?firstResult=%d&maxResults=%d", offset, limit)
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			return errMsg{err}
		}
		req.Header.Set("Accept", "application/json")
		if env.Username != "" {
			req.SetBasicAuth(env.Username, env.Password)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			data, _ := io.ReadAll(resp.Body)
			return errMsg{fmt.Errorf("failed to fetch %s: %s", root, string(data))}
		}
		var items []map[string]interface{}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&items); err != nil {
			return errMsg{err}
		}

		// Try to load count
		count := -1
		countURL := base + "/process-instance/count"
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		req2, err2 := http.NewRequestWithContext(ctx2, http.MethodGet, countURL, nil)
		if err2 == nil {
			req2.Header.Set("Accept", "application/json")
			if env.Username != "" {
				req2.SetBasicAuth(env.Username, env.Password)
			}
			if resp2, err2 := http.DefaultClient.Do(req2); err2 == nil {
				defer resp2.Body.Close()
				if resp2.StatusCode < 400 {
					var cntBody map[string]interface{}
					dec2 := json.NewDecoder(resp2.Body)
					if err3 := dec2.Decode(&cntBody); err3 == nil {
						if v, ok := cntBody["count"]; ok {
							if n, ok := v.(float64); ok {
								count = int(n)
							}
						}
					}
				}
			}
		}

		msg := genericLoadedMsg{root: root, items: items}
		if count >= 0 {
			// attach count via tableTotals map by returning a special msg
			// reuse genericLoadedMsg and set global map after sending message
			// We'll include count as an extra field by wrapping in errMsg on channel is not ideal,
			// instead set pageTotals directly on model via a closure side-effect is not possible here
			// So include count by encoding into items as a special item _meta_count
			if msg.items == nil {
				msg.items = []map[string]interface{}{}
			}
			meta := map[string]interface{}{"_meta_count": count}
			msg.items = append([]map[string]interface{}{meta}, msg.items...)
		}
		return msg
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
				return m, nil
			case "enter":
				m.searchMode = false
				m.searchInput.Blur()
				m.searchTerm = m.searchInput.Value()
				// keep filtered rows as current table rows
				m.originalRows = nil
				m.filteredRows = nil
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
				if m.sortPopupCursor > 0 {
					m.sortPopupCursor--
				}
				return m, nil
			case "down", "j":
				cols := m.table.Columns()
				if m.sortPopupCursor < len(cols)-1 {
					m.sortPopupCursor++
				}
				return m, nil
			case "enter":
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
			switch s {
			case "esc", "q", "y":
				m.activeModal = ModalNone
				m.detailContent = ""
				return m, nil
			case "j", "down":
				if m.detailScroll < m.detailMaxScroll {
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
				if m.detailScroll > m.detailMaxScroll {
					m.detailScroll = m.detailMaxScroll
				}
				return m, nil
			case "ctrl+u":
				m.detailScroll -= 10
				if m.detailScroll < 0 {
					m.detailScroll = 0
				}
				return m, nil
			case "G":
				m.detailScroll = m.detailMaxScroll
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
						return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), m.checkEnvironmentHealthCmd(targetEnv))
					}
				}
				return m, nil
			}
			return m, nil
		}

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
				switch m.editFocus {
				case editFocusInput:
					if len(m.editColumns) > 1 && m.editColumnPos < len(m.editColumns)-1 {
						m.setEditColumn(m.editColumnPos + 1)  // still on input, next column
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
						m.setEditColumn(m.editColumnPos - 1)  // still on input, prev column
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
				m.footerError = ""
			} else {
				m.showRootPopup = false
			}
			return m, nil
		}

		switch s {
		case "q", "Q":
			// Ignore plain 'q'/'Q' to avoid accidental quit; only ctrl+c quits.
			return m, nil
		case "?":
			// Show help screen
			m.activeModal = ModalHelp
			return m, nil
		case "/":
			// Enter search/filter mode
			if !m.showRootPopup && m.activeModal == ModalNone {
				m.searchMode = true
				m.originalRows = append([]table.Row{}, m.table.Rows()...)
				m.searchInput.SetValue("")
				m.searchTerm = ""
				m.searchInput.Focus()
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
			return m, tea.Quit
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
				return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} }))
			}
			return m, nil
		case "s":
			// Open sort popup
			if m.showRootPopup {
				m.rootInput += s
				return m, nil
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
				m.genericParams = prevState.genericParams // restore drilldown filter params

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
						// clear drilldown filter params when switching root context
						m.genericParams = make(map[string]string)
						// reset breadcrumb and header
						m.breadcrumb = []string{rc}
						m.contentHeader = rc
						// If we have a TableDef for this root, set columns and trigger the appropriate fetch
						if def := m.findTableDef(rc); def != nil {
							cols := m.buildColumnsFor(rc, m.paneWidth-4)
							if len(cols) > 0 {
								m.table.SetRows(normalizeRows(nil, len(cols)))
								m.table.SetColumns(cols)
							}
							m.isLoading = true
							m.apiCallStarted = time.Now()
							return m, tea.Batch(m.fetchForRoot(rc), flashOnCmd())
						}
						// fallback to definitions fetch
						m.isLoading = true
					m.apiCallStarted = time.Now()
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
				}
				m.navigationStack = append(m.navigationStack, currentState3)

				key := fmt.Sprintf("%v", row[0])
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
				}
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
				m.rootInput += s
			}
			return m, nil
		case "k":
			// Vim up — only when not in popup/modal/search
			if !m.showRootPopup && m.activeModal == ModalNone {
				m.table.MoveUp(1)
				return m, nil
			}
			if m.showRootPopup {
				m.rootInput += s
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
		// Reserve lines: compact header is 4 rows (with top spacer), context selection 1 line (when active), footer 1 line
		headerLines := 4 // compactHeader placed at 4 rows
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
			return m, tea.Batch(m.fetchDefinitionsCmd(), flashOnCmd(), tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} }))
		}
	case flashOnMsg:
		m.flashActive = true
		// schedule turning the flash off after 200ms
		return m, tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg { return flashOffMsg{} })
	case flashOffMsg:
		m.flashActive = false
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
		// Build rows from items
		rows := make([]table.Row, 0, len(msg.items))
		for _, it := range msg.items {
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
			rows = append(rows, r)
		}

		if len(cols) > 0 {
			m.table.SetColumns(cols)
		}
		normalized := normalizeRows(rows, len(cols))
		colorized := colorizeRows(msg.root, normalized, cols)
		m.table.SetRows(colorized)
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
		msg2, kind, cmd := setFooterStatus(footerStatusSuccess, "✓ Saved", 2*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		return m, cmd
	case dataLoadedMsg:
		// keep backward compatibility: only apply definitions to avoid auto-drilldown
		m.applyDefinitions(msg.definitions)
	case terminatedMsg:
		m.removeInstance(msg.id)
	case suspendedMsg:
		msg2, kind, cmd := setFooterStatus(footerStatusSuccess, fmt.Sprintf("✓ Suspended %s", msg.id), 3*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		return m, cmd
	case resumedMsg:
		msg2, kind, cmd := setFooterStatus(footerStatusSuccess, fmt.Sprintf("✓ Resumed %s", msg.id), 3*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		return m, cmd
	case retriedMsg:
		msg2, kind, cmd := setFooterStatus(footerStatusSuccess, fmt.Sprintf("✓ Retried %s", msg.id), 3*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		return m, cmd
	case actionExecutedMsg:
		msg2, kind, cmd := setFooterStatus(footerStatusSuccess, fmt.Sprintf("✓ %s", msg.label), 3*time.Second)
		m.footerError = msg2
		m.footerStatusKind = kind
		return m, cmd
	case envStatusMsg:
		// Update environment status
		m.envStatus[msg.env] = msg.status
	case errMsg:
		// display error in footer with 8s auto-clear
		msg2, kind, cmd := setFooterStatus(footerStatusError, msg.err.Error(), 8*time.Second)
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
	return "   ____\n____  ( __ )\n / __ \\/ __  / __ \\\n/ /_/ / /_/ / / / /\n\\____/\\____/_/ /_/\n" + "o8n"
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
	// Ensure compact header occupies exactly 4 rows (one extra top spacer)
	compactHeader = lipgloss.Place(m.lastWidth, 4, lipgloss.Left, lipgloss.Center, compactHeader)

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
	title := m.contentHeader
	if total, ok := m.pageTotals[m.currentRoot]; ok && total >= 0 {
		title = fmt.Sprintf("%s — %d items", m.contentHeader, total)
	}
	// Add search filter indicator to title
	if m.searchTerm != "" {
		title = fmt.Sprintf("%s [/%s/]", title, m.searchTerm)
	}

	// Render search bar when in search mode
	searchBar := ""
	if m.searchMode {
		searchStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(color)).
			Width(pw - 4).
			Padding(0, 1)
		searchBar = searchStyle.Render(m.searchInput.View())
	}

	// Apply sort indicator to column headers if sorted
	if m.sortColumn >= 0 {
		cols := m.table.Columns()
		if m.sortColumn < len(cols) {
			// Make a copy to avoid mutating the original
			newCols := make([]table.Column, len(cols))
			copy(newCols, cols)
			indicator := " ▲"
			if !m.sortAscending {
				indicator = " ▼"
			}
			// Remove any previous indicator from all columns
			for i := range newCols {
				newCols[i].Title = strings.TrimSuffix(strings.TrimSuffix(newCols[i].Title, " ▲"), " ▼")
			}
			newCols[m.sortColumn].Title = newCols[m.sortColumn].Title + indicator
			m.table.SetColumns(newCols)
		}
	}

	mainBox := renderBoxWithTitle(m.table.View(), pw, m.paneHeight, title, color)

	// Footer (1 row with 3 columns per specification):
	// Column 1: Context tag with breadcrumb navigation hints
	// Column 2: Status message (error/success/loading/info)
	// Column 3: Remote activity indicator (⚡)
	// Columns separated by " | "

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

	// Render remote flash as a fixed-width symbol on the right, plus latency and pagination if available
	remoteSymbol := " "
	rpStyle := flashBaseStyle
	if m.flashActive {
		remoteSymbol = "⚡"
		rpStyle = rpStyle.Foreground(lipgloss.Color("#00FF00")).Bold(true)
	} else {
		remoteSymbol = " "
		rpStyle = rpStyle.Foreground(lipgloss.Color("#666666"))
	}
	latencyStr := ""
	if m.lastAPILatency > 0 {
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
	rightPart := rpStyle.Render(remoteSymbol + latencyStr + paginationStr)

	// Build status message with icon based on footerStatusKind
	statusMessage := ""
	if m.footerError != "" {
		icon := ""
		style := lipgloss.NewStyle()
		switch m.footerStatusKind {
		case footerStatusError:
			icon, style = "✗ ", errorFooterStyle
		case footerStatusSuccess:
			icon, style = "✓ ", successFooterStyle
		case footerStatusLoading:
			icon, style = "⟳ ", loadingFooterStyle
		default: // footerStatusInfo
			icon, style = "ℹ ", infoFooterStyle
		}
		statusMessage = style.Render(icon + m.footerError)
	}

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

	// Truncate or pad status message to fit available space
	statusWidth := lipgloss.Width(statusMessage)
	if statusWidth > middleW && middleW > 0 {
		// Simple truncation: remove from the end until it fits
		// Note: This is a simplification; proper UTF-8 aware truncation would be more complex
		truncMsg := statusMessage
		for lipgloss.Width(truncMsg) > middleW && len(truncMsg) > 0 {
			truncMsg = truncMsg[:len(truncMsg)-1]
		}
		statusMessage = truncMsg
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
		return m.renderHelpScreen(m.lastWidth, m.lastHeight)
	} else if m.activeModal == ModalSort {
		return m.renderSortPopup(m.lastWidth, m.lastHeight)
	} else if m.activeModal == ModalDetailView {
		return m.renderDetailView(m.lastWidth, m.lastHeight)
	} else if m.activeModal == ModalEnvironment {
		return m.renderEnvPopup(m.lastWidth, m.lastHeight)
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
	// Optionally write last rendered view to ./debug/last-screen.txt when debug enabled
	if m.debugEnabled {
		go func(s string) {
			_ = os.MkdirAll("./debug", 0755)
			_ = os.WriteFile("./debug/last-screen.txt", []byte(s), 0644)
		}(baseView)
	}
	return lipgloss.Place(w, h, lipgloss.Left, lipgloss.Top, baseView)
}

// stripFocusIndicatorPrefix removes the drilldown indicator (▶ ) from the beginning of a string.
// Used to extract clean IDs/names from table rows that have the visual prefix for focus indication.
func stripFocusIndicatorPrefix(s string) string {
	if strings.HasPrefix(s, "▶ ") {
		return strings.TrimPrefix(s, "▶ ")
	}
	return s
}

func rowInstanceID(row table.Row) string {
	if len(row) > 0 {
		return stripFocusIndicatorPrefix(row[0])
	}
	return ""
}

// renderBoxWithTitle draws a simple single-line-border box of the given
// totalWidth/totalHeight and embeds title centered into the top border.
// The content is clipped/padded to fit the inner area. If color is non-empty
// the entire box text is colorized using that foreground color.
func renderBoxWithTitle(content string, totalWidth, totalHeight int, title, color string) string {
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
	// Prepare border style if color provided
	var borderStyle lipgloss.Style
	useBorderColor := color != ""
	if useBorderColor {
		borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	}

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
		return tea.Batch(m.fetchInstancesCmd("processDefinitionKey", m.selectedDefinitionKey), flashOnCmd())
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

// filterRows returns rows where any cell contains the search term (case-insensitive).
func filterRows(rows []table.Row, term string) []table.Row {
	if term == "" {
		return rows
	}
	lower := strings.ToLower(term)
	var result []table.Row
	for _, row := range rows {
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), lower) {
				result = append(result, row)
				break
			}
		}
	}
	return result
}

// Status color styles for row colorization
var (
	statusRunningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#50C878"))
	statusSuspendedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	statusFailedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	statusEndedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
)

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

// colorizeRows applies status-based color indicators to the first column of each row.
func colorizeRows(root string, rows []table.Row, columns []table.Column) []table.Row {
	if len(rows) == 0 || len(columns) == 0 {
		return rows
	}
	result := make([]table.Row, len(rows))
	for i, row := range rows {
		status := detectRowStatus(root, row, columns)
		newRow := make(table.Row, len(row))
		copy(newRow, row)

		// Apply color to all cells in the row
		var style lipgloss.Style
		switch status {
		case "running":
			style = statusRunningStyle
		case "suspended":
			style = statusSuspendedStyle
		case "failed":
			style = statusFailedStyle
		case "ended":
			style = statusEndedStyle
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

// renderSortPopup renders the column sort selection popup.
func (m *model) renderSortPopup(width, height int) string {
	cols := m.table.Columns()

	color := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			color = env.UIColor
		}
	}

	var b strings.Builder
	for i, col := range cols {
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
		b.WriteString(fmt.Sprintf("%s%s%s\n", cursor, col.Title, indicator))
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Padding(0, 1).
		Width(30)

	title := "Sort by Column"
	content := b.String()
	modal := modalStyle.Render(title + "\n" + content + "\nEnter: Select  Esc: Close")

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}

// renderDetailView renders the YAML/JSON detail viewer overlay.
func (m *model) renderDetailView(width, height int) string {
	color := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			color = env.UIColor
		}
	}

	// Calculate visible area
	viewHeight := height - 6
	if viewHeight < 3 {
		viewHeight = 3
	}

	lines := strings.Split(m.detailContent, "\n")
	m.detailMaxScroll = len(lines) - viewHeight
	if m.detailMaxScroll < 0 {
		m.detailMaxScroll = 0
	}
	if m.detailScroll > m.detailMaxScroll {
		m.detailScroll = m.detailMaxScroll
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
		b.WriteString(fmt.Sprintf("%4d │ %s\n", lineNum, syntaxHighlightJSON(line)))
	}

	scrollInfo := fmt.Sprintf("[%d/%d]", m.detailScroll+1, len(lines))

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Padding(1, 2).
		Width(width - 4).
		Height(viewHeight + 2)

	content := b.String()
	title := "Detail View  " + scrollInfo + "  (j/k scroll, q/Esc close)"
	modal := modalStyle.Render(title + "\n" + content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}

// syntaxHighlightJSON applies basic syntax highlighting to a JSON line.
func syntaxHighlightJSON(line string) string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00BCD4"))
	stringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#50C878"))
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	boolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF69B4"))

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
	color := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			color = env.UIColor
		}
	}

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

		// Environment color
		envColor := ""
		if env, ok := m.config.Environments[name]; ok {
			envColor = env.UIColor
		}
		envStyle := lipgloss.NewStyle()
		if envColor != "" {
			envStyle = envStyle.Foreground(lipgloss.Color(envColor))
		}

		url := ""
		if env, ok := m.config.Environments[name]; ok {
			url = env.URL
		}

		line := fmt.Sprintf("%s%s %s  %s", cursor, statusIcon, envStyle.Render(name), url)
		b.WriteString(line + "\n")
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Padding(0, 1).
		Width(50)

	content := "Select Environment\n\n" + b.String() + "\nEnter: Select  Esc: Close"
	modal := modalStyle.Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}

// renderActionsMenu renders the context actions menu popup.
func (m *model) renderActionsMenu(width, height int) string {
	color := ""
	if m.config != nil {
		if env, ok := m.config.Environments[m.currentEnv]; ok {
			color = env.UIColor
		}
	}

	var b strings.Builder
	root := m.currentRoot
	b.WriteString(fmt.Sprintf("Actions: %s\n\n", root))
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
		BorderForeground(lipgloss.Color(color)).
		Padding(0, 1).
		Width(35)

	modal := modalStyle.Render(b.String())
	return overlayCenter(lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, ""), modal)
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

// checkEnvironmentHealthCmd performs a simple health check on the given environment
func (m *model) checkEnvironmentHealthCmd(envName string) tea.Cmd {
	return func() tea.Msg {
		if m.config == nil {
			return envStatusMsg{env: envName, status: StatusUnknown}
		}

		env, ok := m.config.Environments[envName]
		if !ok {
			return envStatusMsg{env: envName, status: StatusUnknown}
		}

		// Make a simple HTTP GET request to check connectivity
		httpClient := &http.Client{
			Timeout: 2 * time.Second,
		}

		// Try a simple identity endpoint to check health
		req, err := http.NewRequest("GET", env.URL+"/identity/current", nil)
		if err != nil {
			return envStatusMsg{env: envName, status: StatusUnreachable, err: err}
		}

		// Add basic auth
		req.SetBasicAuth(env.Username, env.Password)

		resp, err := httpClient.Do(req)
		if err != nil {
			return envStatusMsg{env: envName, status: StatusUnreachable, err: err}
		}
		defer resp.Body.Close()

		// Any 2xx or 3xx response = operational (ignore auth errors for now)
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			return envStatusMsg{env: envName, status: StatusOperational}
		}

		// 4xx/5xx = unreachable/unhealthy
		return envStatusMsg{env: envName, status: StatusUnreachable}
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

// deriveFocusBackgroundColor returns a darker shade of the accent color for focus indicator background.
// Maps common hex colors to darker 256-color codes or returns a default dark color.
func (m *model) deriveFocusBackgroundColor(accentColor string) string {
	// Map of accent colors to dark background colors (256-color codes)
	colorMap := map[string]string{
		"#FFA500": "94",   // Orange → dark orange
		"#00A8E1": "23",   // Blue → dark blue
		"#00D7FF": "23",   // Cyan → dark cyan
		"#50C878": "22",   // Green → dark green
		"#FF6B6B": "52",   // Red → dark red
		"#FFD700": "136",  // Gold → dark gold
	}

	if darkShade, ok := colorMap[accentColor]; ok {
		return darkShade
	}
	// Default to dark blue if color not in map
	return "23"
}

// validateConfigFiles verifies that critical config files exist and are not empty.
// Returns an error if config files appear corrupted or missing.
func validateConfigFiles() error {
	criticalFiles := map[string]int{
		"o8n-cfg.yaml": 700,   // Config should have ~760 lines
		"o8n-env.yaml": 5,     // Env should have ~10-11 lines (absolute minimum)
	}

	for file, minLines := range criticalFiles {
		// Read file to count lines and verify it exists
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("critical file missing or unreadable: %s", file)
		}

		// Check if file appears empty or too small
		size := len(data)
		if size < 100 {
			return fmt.Errorf("critical file appears corrupted or empty: %s (%d bytes). Expected ~%d lines minimum. Restore with: git show HEAD~2:%s > %s",
				file, size, minLines, file, file)
		}

		lineCount := strings.Count(string(data), "\n")
		if lineCount < minLines {
			return fmt.Errorf("%s appears corrupted (%d lines, expected ~%d minimum). Restore with: git show HEAD~2:%s > %s",
				file, lineCount, minLines, file, file)
		}
	}

	return nil
}

func main() {
	var debug = flag.Bool("debug", false, "enable debug logging")
	var skin = flag.String("skin", "", "skin to use")
	var noSplash = flag.Bool("no-splash", false, "disable splash screen")
	flag.Parse()

	// If debug is enabled, create a log file
	if *debug {
		os.Mkdir("debug", 0755)
		f, err := os.OpenFile("debug/o8n.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
		log.Println("--- o8n debug session started ---")
	} else {
		// a nil writer discards all log output
		log.SetOutput(io.Discard)
	}

	// Verify critical config files exist and are not corrupted
	if err := validateConfigFiles(); err != nil {
		log.Printf("CRITICAL: %v", err)
		return
	}

	// Load split config files (o8n-env.yaml + o8n-cfg.yaml). No legacy fallback.
	envCfg, err := config.LoadEnvConfig("o8n-env.yaml")
	if err != nil {
		fmt.Printf("Error loading o8n-env.yaml: %v\n", err)
		fmt.Println("Please create o8n-env.yaml from the example.")
		os.Exit(1)
	}

	appCfg, err := config.LoadAppConfig("o8n-cfg.yaml")
	if err != nil {
		fmt.Printf("Error loading o8n-cfg.yaml: %v\n", err)
		os.Exit(1)
	}

	if len(envCfg.Environments) == 0 {
		log.Println("No environments configured. Please create 'o8n-env.yaml' and define at least one environment.")
		return
	}

	skinName := *skin
	if skinName == "" && envCfg.Skin != "" {
		skinName = envCfg.Skin
	}
	log.Printf("DEBUG: skinName resolved to: %s", skinName)

	m := newModelEnvApp(envCfg, appCfg, skinName)
	m.debugEnabled = *debug
	if *noSplash {
		m.splashActive = false
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatalf("failed to run program: %v", err)
	}
}
