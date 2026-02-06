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

const refreshInterval = 5 * time.Second

type refreshMsg struct{}

type dataLoadedMsg struct {
	definitions []ProcessDefinition
	instances   []ProcessInstance
}

type terminatedMsg struct{ id string }

type errMsg struct{ err error }

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

	list  list.Model
	table table.Model

	style lipgloss.Style
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

	cols := []table.Column{
		{Title: "Instance ID", Width: 22},
		{Title: "Definition", Width: 18},
		{Title: "Business Key", Width: 16},
		{Title: "Started", Width: 14},
	}
	t := table.New(table.WithColumns(cols), table.WithFocused(true))

	current := ""
	if len(envNames) > 0 {
		current = envNames[0]
	}

	m := model{
		config:     cfg,
		envNames:   envNames,
		currentEnv: current,
		list:       l,
		table:      t,
	}
	m.applyStyle()
	return m
}

func (m *model) applyStyle() {
	color := ""
	if env, ok := m.config.Environments[m.currentEnv]; ok {
		color = env.UIColor
	}
	m.style = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(color)).
		Bold(true)

	listStyles := list.DefaultStyles()
	listStyles.Title = listStyles.Title.BorderForeground(lipgloss.Color(color))
	m.list.Styles = listStyles

	tStyles := table.DefaultStyles()
	tStyles.Header = tStyles.Header.BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(color)).
		BorderBottom(true).
		Bold(true)
	tStyles.Selected = tStyles.Selected.Foreground(lipgloss.Color(color)).Bold(true)
	m.table.SetStyles(tStyles)
}

func (m model) Init() tea.Cmd {
	if m.autoRefresh {
		return tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} })
	}
	return nil
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
}

func (m *model) resetViews() {
	m.list.SetItems([]list.Item{})
	m.table.SetRows([]table.Row{})
	m.selectedInstanceID = ""
}

func (m *model) applyData(defs []ProcessDefinition, instances []ProcessInstance) {
	items := make([]list.Item, 0, len(defs))
	for _, d := range defs {
		items = append(items, processDefinitionItem{definition: d})
	}
	m.list.SetItems(items)

	rows := make([]table.Row, 0, len(instances))
	for _, inst := range instances {
		rows = append(rows, table.Row{inst.ID, inst.DefinitionID, inst.BusinessKey, inst.StartTime})
	}
	m.table.SetRows(rows)

	m.selectedInstanceID = rowInstanceID(m.table.SelectedRow())
}

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

func (m model) fetchDataCmd() tea.Cmd {
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
			return m, nil
		case "r":
			m.autoRefresh = !m.autoRefresh
			if m.autoRefresh {
				return m, tea.Batch(m.fetchDataCmd(), tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} }))
			}
			return m, nil
		case "x":
			if m.selectedInstanceID != "" {
				m.showKillModal = true
			}
			return m, nil
		case "esc":
			m.showKillModal = false
			return m, nil
		case "y":
			if m.showKillModal && m.selectedInstanceID != "" {
				m.showKillModal = false
				return m, m.terminateInstanceCmd(m.selectedInstanceID)
			}
			return m, nil
		}
	case refreshMsg:
		if m.autoRefresh {
			return m, tea.Batch(m.fetchDataCmd(), tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshMsg{} }))
		}
	case dataLoadedMsg:
		m.applyData(msg.definitions, msg.instances)
	case terminatedMsg:
		m.removeInstance(msg.id)
	case errMsg:
		log.Printf("error: %v", msg.err)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	m.selectedInstanceID = rowInstanceID(m.table.SelectedRow())

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	refreshStatus := "OFF"
	if m.autoRefresh {
		refreshStatus = "ON"
	}
	header := m.style.Render(fmt.Sprintf("Environment: %s | Auto-refresh: %s", m.currentEnv, refreshStatus))
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.style.Render(m.list.View()),
		m.style.Render(m.table.View()),
	)

	if m.showKillModal {
		modal := m.style.Copy().
			BorderStyle(lipgloss.DoubleBorder()).
			Render("Terminate selected instance? (y/esc)")
		return lipgloss.JoinVertical(lipgloss.Left, header, content, modal)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func rowInstanceID(row table.Row) string {
	if len(row) > 0 {
		return row[0]
	}
	return ""
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
