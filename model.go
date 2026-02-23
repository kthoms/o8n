package main

import (
	"fmt"
	"log"
	"os"
	"sort"
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
	label string // the action label for feedback
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
)

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

	// colon popup
	showRootPopup   bool
	rootPopupCursor int
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

	// raw API data per visible row; used to resolve drilldown column values
	// including columns that are not displayed (hidden) in the table
	rowData []map[string]interface{}

	// Version number
	version      string
	debugEnabled bool
	debugCh      chan string

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

	// Environment popup state
	showEnvPopup   bool
	envPopupCursor int

	activeSkin string
	skin       *Skin
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
		debugCh:                make(chan string, 1),
		pageOffsets:            make(map[string]int),
		pageTotals:             make(map[string]int),
		pendingCursorAfterPage: -1,
		genericParams:          make(map[string]string),
		envStatus:              make(map[string]EnvironmentStatus),
		sortColumn:             -1,
		rootPopupCursor:        -1,
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
	// compactHeader (2 lines) + content header (1 line) = 3 header lines total
	headerLines := 3
	footerLines := 2 // breadcrumb line + status line
	searchBarLines := 0
	if m.searchMode {
		searchBarLines = 1
	}
	// reserve an extra safe line to avoid off-by-one overflow
	contentHeight := m.lastHeight - headerLines - m.contextPopupHeight() - searchBarLines - footerLines - 1
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

	// start single debug writer goroutine (runs for process lifetime)
	go func(ch chan string) {
		_ = os.MkdirAll("./debug", 0755)
		for s := range ch {
			_ = os.WriteFile("./debug/last-screen.txt", []byte(s), 0644)
		}
	}(m.debugCh)

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
	cmds = append(cmds, tea.Tick(60*time.Second, func(time.Time) tea.Msg { return healthTickMsg{} }))

	return tea.Batch(cmds...)
}
