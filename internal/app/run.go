package app

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kthoms/o8n/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

const statePath = "o8n-stat.yml"

// Run is the application entry point called from main.
func Run() {
	var debug = flag.Bool("debug", false, "enable debug logging")
	var skin = flag.String("skin", "", "skin to use")
	var noSplash = flag.Bool("no-splash", false, "disable splash screen")
	var vimFlag = flag.Bool("vim", false, "enable vim keybindings (j/k/gg/G/Ctrl+U/Ctrl+D)")
	flag.Parse()

	// Always open debug/o8n.log for error logging.
	// Verbose debug output additionally requires --debug.
	os.Mkdir("debug", 0755)
	logFile, err := os.OpenFile("debug/o8n.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()
	if *debug {
		log.SetOutput(logFile)
		log.Println("--- o8n debug session started ---")
	} else {
		// Non-debug mode: only error-level messages go to the log file.
		// Use a custom writer that prefixes every line so it's easy to grep.
		log.SetOutput(logFile)
		log.SetPrefix("[ERROR] ")
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

	// Load persisted runtime state (o8n-stat.yml). Missing file is not an error.
	appState, err := config.LoadAppState(statePath)
	if err != nil {
		log.Printf("Warning: could not load state file: %v", err)
		appState = &config.AppState{}
	}

	// Resolve skin: CLI flag > state file > env config (legacy) > ""
	skinName := *skin
	if skinName == "" {
		skinName = appState.Skin
	}
	if skinName == "" {
		skinName = envCfg.Skin
	}
	log.Printf("DEBUG: skinName resolved to: %s", skinName)

	m := newModelEnvApp(envCfg, appCfg, skinName)
	m.debugEnabled = *debug
	m.vimMode = *vimFlag || appCfg.VimMode
	m.statePath = statePath
	m.showLatency = appState.ShowLatency

	// Restore active environment from state (falls back to env config / default).
	if appState.ActiveEnv != "" {
		if _, ok := m.config.Environments[appState.ActiveEnv]; ok {
			m.currentEnv = appState.ActiveEnv
			m.applyStyle()
		}
	}

	// Restore last navigation position (root resource + drilldown path).
	m.restoreNavState(appState.Navigation)

	if *noSplash {
		m.splashActive = false
	}

	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		log.Fatalf("failed to run program: %v", err)
	}

	// Persist state on clean exit using the final model, not the initial one.
	if fm, ok := finalModel.(model); ok {
		_ = config.SaveAppState(statePath, fm.currentAppState())
	}
}
