package app

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kthoms/o8n/internal/config"
	"github.com/kthoms/o8n/internal/dao"
	"golang.org/x/term"
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
	label           string // the action label for feedback
	closeTaskDialog bool   // when true, close ModalTaskComplete and clear its state
}

type errMsg struct{ err error }

type healthTickMsg struct{}

// Messages for the flash indicator
type flashOnMsg struct{}
type flashOffMsg struct{}

// message used to clear footer errors
type clearErrorMsg struct{}

// clearPendingGMsg resets the pendingG flag after timeout
type clearPendingGMsg struct{}

type spinnerTickMsg struct{}

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

func spinnerTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return spinnerTickMsg{} })
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
	ModalTaskComplete
)

// taskCompleteFocusArea tracks keyboard focus within the task completion modal
type taskCompleteFocusArea int

const (
	focusTaskField    taskCompleteFocusArea = iota // a form field is focused
	focusTaskComplete                              // [Complete] button focused
	focusTaskBack                                  // [Back] button focused
)

// taskCompleteField represents one editable output variable in the completion dialog
type taskCompleteField struct {
	name     string
	varType  string // lowercased: "string", "boolean", "integer", "double", "json"
	origType string // original casing from API for submission (e.g. "String", "Boolean")
	input    textinput.Model
	error    string
}

// variableValue holds a variable's value and type name as returned by the API
type variableValue struct {
	Value    interface{}
	TypeName string // original casing from API (e.g. "String", "Boolean", "Integer")
}

// taskVariablesLoadedMsg is sent when both task variable fetches complete
type taskVariablesLoadedMsg struct {
	taskID    string
	taskName  string
	inputVars map[string]variableValue // from GET /task/{id}/variables
	formVars  map[string]variableValue // from GET /task/{id}/form-variables
}

// popupMode identifies the active command palette mode.
type popupMode int

const (
	popupModeNone    popupMode = iota
	popupModeContext           // : key — switch resource context
	popupModeSkin              // Ctrl+T key — switch skin/theme
	popupModeSearch            // / key — filter table rows
)

// popupState holds the runtime state of the generic command palette popup.
type popupState struct {
	mode        popupMode
	input       string
	cursor      int
	offset      int      // scroll offset: index of first visible item in the list
	items       []string // filtered/available items for current mode
	title       string   // "context" | "skin"
	hint        string   // hint line shown below input
	previewSkin string   // skin name before preview started (for revert on Esc)
}

// skinsLoadedMsg carries the list of available skin names.
type skinsLoadedMsg struct{ names []string }

// editFocusArea tracks keyboard focus within the edit modal
type editFocusArea int

const (
	editFocusInput  editFocusArea = iota // text input focused
	editFocusSave                        // Save button focused
	editFocusCancel                      // Cancel button focused
)

type editSavedMsg struct {
	rowIndex int
	colIndex int
	value    string
	dataKey  string // API field name to update in rowData
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
	genericParams         map[string]string        // drilldown filter params active at this level
	rowData               []map[string]interface{} // raw API data per row for drilldown column lookup
}

type model struct {
	config *config.Config

	envNames []string

	currentEnv         string
	autoRefresh        bool
	showKillModal      bool
	selectedInstanceID string

	list  list.Model
	table table.Model

	// cached definitions for drilldown/back
	cachedDefinitions []config.ProcessDefinition

	// navigation history stack for back/forward
	navigationStack []viewState

	// view mode: "definitions" or "instances"
	viewMode string

	// Modal state
	activeModal        ModalType
	modalConfirmKey    string // The key to press to confirm (e.g., "ctrl+d")
	pendingDeleteID    string // ID pending deletion confirmation
	pendingDeleteLabel string // Display label for the item pending deletion

	confirmFocusedBtn int // 0=confirm button focused, 1=cancel button focused (default 1=cancel)

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
	searchMode   bool
	searchInput  textinput.Model
	searchTerm   string
	filteredRows []table.Row
	originalRows []table.Row

	// lastListIndex stores the last-known list index so we can detect
	// selection changes even when list.Update doesn't change the index in tests.
	lastListIndex int

	// flashActive indicates the footer flash should be shown
	flashActive bool

	spinnerFrame int

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

	// generic command palette popup (context switch, skin picker, etc.)
	popup popupState
	// available root contexts (computed from API spec)
	rootContexts []string

	// available skins (populated at startup)
	availableSkins []string

	// current root context
	currentRoot string

	// skin-driven style set — rebuilt on skin switch
	styles StyleSet

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

	// raw API data per visible row; used to resolve drilldown column values
	// including columns that are not displayed (hidden) in the table
	rowData []map[string]interface{}

	// Version number
	version      string
	debugEnabled bool
	vimMode      bool // true when vim keybindings are active (--vim flag or vim_mode config)
	quitting     bool // set true before tea.Quit so View() returns "" to clear screen

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
	detailContent string
	detailScroll  int

	// Help scroll offset
	helpScroll int

	// Task completion dialog state
	taskCompleteTaskID   string
	taskCompleteTaskName string
	taskInputVars        map[string]variableValue // from GET /task/{id}/variables (for pre-fill)
	taskCompleteFields   []taskCompleteField      // editable output fields (form variables)
	taskCompletePos      int                      // index of focused field (when focusTaskField)
	taskCompleteFocus    taskCompleteFocusArea

	// Environment popup state
	showEnvPopup   bool
	envPopupCursor int

	activeSkin  string
	skin        *Skin
	showLatency bool   // toggleable latency display (default off)
	statePath   string // path to o8n-stat.yml
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
		viewMode:               "process-definition",
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
		popup:                  popupState{cursor: -1},
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
	// full terminal width — no ghost left pane
	m.paneWidth = m.lastWidth
	if m.paneWidth < 10 {
		m.paneWidth = 10
	}
	// compute content height reserving header/footer lines
	// compactHeader 2 rows + contextSelection 1 row + footer 1 row
	headerLines := 2
	footerLines := 1
	contextLines := 1
	searchBarLines := 0
	if m.searchMode {
		searchBarLines = 1
	}
	contentHeight := m.lastHeight - headerLines - contextLines - searchBarLines - footerLines
	if contentHeight < 3 {
		contentHeight = 3
	}
	m.paneHeight = contentHeight

	// initialize table sizes to match detected terminal size
	tableInner := m.paneWidth - 4
	if tableInner < 10 {
		tableInner = 10
	}
	m.table.SetWidth(tableInner)
	m.table.SetHeight(contentHeight - 1)

	// set root contexts and currentRoot
	m.rootContexts = loadRootContexts("resources/operaton-rest-api.json")
	// Filter to only contexts that have a TableDef in config — prevents broken contexts
	filtered := m.rootContexts[:0]
	for _, rc := range m.rootContexts {
		if m.findTableDef(rc) != nil {
			filtered = append(filtered, rc)
		}
	}
	m.rootContexts = filtered
	m.currentRoot = dao.ResourceProcessDefinitions

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
		}
	}
	cfg.Active = envCfg.Active
	cfg.Skin = envCfg.Skin

	m := newModel(cfg)
	m.envConfig = envCfg
	m.appConfig = appCfg
	m.activeSkin = skinName

	// Normalize skin filename: accept either bare name (e.g. "narsingh") or full filename ("narsingh.yaml").
	skinFile := skinName
	if skinFile != "" && !strings.HasSuffix(skinFile, ".yaml") {
		skinFile = skinFile + ".yaml"
	}
	skin, err := loadSkin(skinFile)
	if err != nil {
		log.Printf("Falling back to stock skin. Could not load skin: %v", err)
		skin, _ = loadSkin("stock.yaml")
	}
	m.skin = skin
	m.styles = buildStyleSet(skin)
	// Apply the loaded skin to update styles, list/table styles, and other derived styles.
	m.applyStyle()

	// copy tables into m.config for backward compatibility
	m.config = cfg
	return m
}

func (m *model) applyStyle() {
	if m.skin == nil {
		return
	}
	m.styles = buildStyleSet(m.skin)

	m.style = lipgloss.NewStyle().
		Foreground(col(m.skin, "fg")).
		Background(col(m.skin, "bg"))

	listStyles := list.DefaultStyles()
	listStyles.Title = listStyles.Title.BorderForeground(col(m.skin, "borderFocus"))
	m.list.Styles = listStyles

	tStyles := table.DefaultStyles()
	tStyles.Header = tStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(col(m.skin, "borderFg")).
		BorderBottom(false).
		Foreground(col(m.skin, "fg")).
		Bold(true)
	tStyles.Selected = tStyles.Selected.
		Foreground(col(m.skin, "bg")).
		Background(col(m.skin, "borderFocus")).
		Bold(true).
		Width(m.paneWidth)
	m.table.SetStyles(tStyles)

	m.splashLogoStyle = m.styles.Logo
	m.splashInfoStyle = m.styles.Info

	m.breadcrumbStyles = []lipgloss.Style{
		m.styles.CrumbActive,
		m.styles.CrumbNormal,
		m.styles.CrumbNormal,
		m.styles.CrumbNormal,
	}
}

// skinPopupItems returns the filtered skin names for the current popup input.
func (m *model) skinPopupItems() []string {
	var items []string
	for _, name := range m.availableSkins {
		if m.popup.input == "" || strings.HasPrefix(name, m.popup.input) {
			items = append(items, name)
		}
	}
	return items
}

// previewSkinByName loads a skin and applies it for live preview without committing.
func (m *model) previewSkinByName(name string) {
	skin, err := loadSkin(name + ".yaml")
	if err != nil {
		return
	}
	m.skin = skin
	m.activeSkin = name
	m.applyStyle()
}

func (m model) Init() tea.Cmd {
	firstTick := tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return splashFrameMsg{frame: 2} })

	// Determine the initial fetch based on restored navigation state.
	var initialFetch tea.Cmd
	if m.currentRoot != "" && len(m.breadcrumb) > 0 {
		// Restore last viewed resource — re-fetch fresh data for it.
		last := m.breadcrumb[len(m.breadcrumb)-1]
		// Use genericParams already set by restoreNavState to carry drilldown filters.
		initialFetch = m.fetchForRoot(last)
		if initialFetch == nil {
			initialFetch = m.fetchForRoot(m.currentRoot)
		}
	} else {
		initialFetch = m.fetchForRoot("process-definition")
	}

	// Check health of all environments
	cmds := []tea.Cmd{initialFetch, flashOnCmd(), firstTick, listSkinsCmd()}
	for _, envName := range m.envNames {
		cmds = append(cmds, m.checkEnvironmentHealthCmd(envName))
	}
	cmds = append(cmds, tea.Tick(60*time.Second, func(time.Time) tea.Msg { return healthTickMsg{} }))

	return tea.Batch(cmds...)
}
