package main

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type refreshMsg struct{}

type fetchResultMsg struct {
	defs []ProcessDefinition
	err  error
}

type terminateResultMsg struct {
	err error
}

type processDefinitionItem struct {
	name string
	desc string
}

func (p processDefinitionItem) Title() string       { return p.name }
func (p processDefinitionItem) Description() string { return p.desc }
func (p processDefinitionItem) FilterValue() string { return p.name }

type model struct {
	list               list.Model
	table              table.Model
	config             *Config
	envKeys            []string
	currentEnv         string
	autoRefresh        bool
	showKillModal      bool
	selectedInstanceID string
	styles             lipgloss.Style
	tick               func(time.Duration) tea.Cmd
	clientFactory      func(Environment) *Client
}

func newModel(cfg *Config) model {
	envKeys := make([]string, 0, len(cfg.Environments))
	for key := range cfg.Environments {
		envKeys = append(envKeys, key)
	}
	sort.Strings(envKeys)

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	t := table.New(table.WithColumns([]table.Column{
		{Title: "Instance ID", Width: 20},
		{Title: "Definition ID", Width: 20},
	}))

	m := model{
		list:          l,
		table:         t,
		config:        cfg,
		envKeys:       envKeys,
		autoRefresh:   false,
		showKillModal: false,
		tick: func(d time.Duration) tea.Cmd {
			return tea.Tick(d, func(time.Time) tea.Msg { return refreshMsg{} })
		},
		clientFactory: NewClient,
	}

	if len(envKeys) > 0 {
		m.currentEnv = envKeys[0]
	}
	m.updateStyles()
	return m
}

func (m *model) updateStyles() {
	color := ""
	if env, ok := m.config.Environments[m.currentEnv]; ok && env.UIColor != "" {
		color = env.UIColor
	}
	m.styles = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(color)).
		Bold(true)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "e":
			if len(m.envKeys) > 0 {
				m.currentEnv = m.nextEnv()
				m.list.SetItems([]list.Item{})
				m.table.SetRows([]table.Row{})
				m.selectedInstanceID = ""
				m.showKillModal = false
				m.updateStyles()
			}
		case "r":
			m.autoRefresh = !m.autoRefresh
			if m.autoRefresh {
				return m, m.tick(5 * time.Second)
			}
		case "x":
			if m.selectedInstanceID != "" {
				m.showKillModal = true
			}
		case "esc":
			if m.showKillModal {
				m.showKillModal = false
			}
		case "y":
			if m.showKillModal && m.selectedInstanceID != "" {
				cmd := m.terminateInstanceCmd()
				m.showKillModal = false
				return m, cmd
			}
		}
	case refreshMsg:
		if m.autoRefresh {
			return m, tea.Batch(m.fetchDataCmd(), m.tick(5*time.Second))
		}
	case fetchResultMsg:
		if msg.err == nil {
			items := make([]list.Item, len(msg.defs))
			for i, def := range msg.defs {
				name := def.Name
				if name == "" {
					name = def.Key
				}
				items[i] = processDefinitionItem{
					name: name,
					desc: def.ID,
				}
			}
			m.list.SetItems(items)
		}
	case terminateResultMsg:
		if msg.err != nil {
			log.Printf("failed to terminate instance %s: %v", m.selectedInstanceID, msg.err)
		}
	}

	var cmds []tea.Cmd

	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	if listCmd != nil {
		cmds = append(cmds, listCmd)
	}

	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)
	if tableCmd != nil {
		cmds = append(cmds, tableCmd)
	}

	if row := m.table.SelectedRow(); len(row) > 0 {
		m.selectedInstanceID = row[0]
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	header := fmt.Sprintf("Env: %s | Auto-refresh: %t", m.currentEnv, m.autoRefresh)
	body := m.styles.Render(header)
	if m.showKillModal {
		modal := m.styles.Copy().
			Border(lipgloss.RoundedBorder()).
			Render("Terminate instance? (y / esc)")
		return body + "\n\n" + modal
	}
	return body
}

func (m model) nextEnv() string {
	if len(m.envKeys) == 0 {
		return ""
	}
	for i, key := range m.envKeys {
		if key == m.currentEnv {
			return m.envKeys[(i+1)%len(m.envKeys)]
		}
	}
	return m.envKeys[0]
}

func (m model) fetchDataCmd() tea.Cmd {
	if m.config == nil || m.currentEnv == "" {
		return nil
	}
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	return func() tea.Msg {
		client := m.clientFactory(env)
		defs, err := client.FetchProcessDefinitions()
		return fetchResultMsg{defs: defs, err: err}
	}
}

func (m model) terminateInstanceCmd() tea.Cmd {
	if m.config == nil || m.currentEnv == "" || m.selectedInstanceID == "" {
		return nil
	}
	env, ok := m.config.Environments[m.currentEnv]
	if !ok {
		return nil
	}
	instanceID := m.selectedInstanceID
	return func() tea.Msg {
		client := m.clientFactory(env)
		err := client.TerminateInstance(instanceID)
		return terminateResultMsg{err: err}
	}
}

func main() {
	fmt.Println("o8n - Terminal UI for Operaton")
	fmt.Println("-------------------------------")

	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Note: Could not load config.yaml: %v", err)
		log.Printf("You can create a config.yaml based on config.yaml.example")
		return
	}

	if len(config.Environments) == 0 {
		log.Printf("No environments configured; exiting.")
		return
	}

	if _, err := tea.NewProgram(newModel(config)).Run(); err != nil {
		log.Fatalf("failed to start TUI: %v", err)
	}
}
