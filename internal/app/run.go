package app

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kthoms/o8n/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

// Run is the application entry point called from main.
func Run() {
	var debug = flag.Bool("debug", false, "enable debug logging")
	var skin = flag.String("skin", "", "skin to use")
	var noSplash = flag.Bool("no-splash", false, "disable splash screen")
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
