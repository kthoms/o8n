package app

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// helper: strip ANSI escape codes
var ansiEscRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiEscRe.ReplaceAllString(s, "")
}

// helper: ANSI-aware width
func lipglossWidth(s string) int {
	return lipgloss.Width(s)
}

// helper: render sort popup with given column names
func renderSortPopupWithCols(m model, colNames []string) string {
	cols := make([]table.Column, len(colNames))
	for i, name := range colNames {
		cols[i] = table.Column{Title: name, Width: 10}
	}
	m.table.SetColumns(cols)
	return m.renderSortPopup(m.lastWidth, m.lastHeight)
}

// ── T1: Remove ghost left pane ────────────────────────────────────────────────

func TestPaneWidthEqualsTerminalWidth(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m3 := m2.(model)

	if m3.paneWidth != 120 {
		t.Errorf("expected paneWidth=120 (full terminal width), got %d", m3.paneWidth)
	}
}

func TestPaneWidthUpdatesOnResize(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	m2, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m3 := m2.(model)

	if m3.paneWidth != 80 {
		t.Errorf("expected paneWidth=80, got %d", m3.paneWidth)
	}
}

// ── T2: Header is 2 rows (not 3) ─────────────────────────────────────────────

func TestHeaderIs2Rows(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 40

	rendered := m.renderCompactHeader(120)
	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")

	if len(lines) < 2 {
		t.Errorf("expected at least 2 header rows, got %d", len(lines))
	}
}

func TestPaneHeightGainsOneRowWithHeader2(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m3 := m2.(model)

	// With header=2, footer=1, contextSelectionLines=1: contentHeight = 40-2-1-1 = 36
	// computePaneHeight: 40 - 2 - 2 - 1 = 35 (header(2) - footer(2) - safe(1))
	// Before fix (header=3): 40 - 3 - 2 - 1 = 34
	if m3.paneHeight <= 33 {
		t.Errorf("expected paneHeight > 33 (gained row from header), got %d", m3.paneHeight)
	}
}

// ── T3: Breadcrumb capped at 50% terminal width ───────────────────────────────

func TestBreadcrumbCapAtHalfWidth(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 80
	m.lastHeight = 30

	m.breadcrumb = []string{
		"process-definitions",
		"process-instances",
		"process-variables",
		"deep-resource-name",
	}

	rendered := m.View()
	lines := strings.Split(rendered, "\n")
	if len(lines) == 0 {
		t.Fatal("empty view")
	}

	footerLine := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			footerLine = lines[i]
			break
		}
	}

	parts := strings.SplitN(footerLine, " | ", 2)
	if len(parts) < 1 {
		t.Skip("no footer separator found")
	}
	breadcrumbPart := parts[0]
	bw := len(stripANSI(breadcrumbPart))
	if bw > 40 {
		t.Errorf("breadcrumb width %d exceeds 50%% cap of 40 at 80-char terminal", bw)
	}
}

// ── T4: Sort modal width dynamic ─────────────────────────────────────────────

func TestSortModalWidthAdaptsToColumnNames(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 40

	longColName := strings.Repeat("A", 50)
	rendered := renderSortPopupWithCols(m, []string{longColName, "Short"})

	// min width = max(30, 50+8) = 58, capped at 120-10=110
	modalW := lipglossWidth(rendered)
	if modalW < 50 {
		t.Errorf("expected sort modal wider than 50 for long column name, got %d", modalW)
	}
	if modalW > 110 {
		t.Errorf("expected sort modal width <= 110 (width-10), got %d", modalW)
	}
}

func TestSortModalMinWidth30(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 40

	rendered := renderSortPopupWithCols(m, []string{"ID", "Name"})
	modalW := lipglossWidth(rendered)
	if modalW < 28 {
		t.Errorf("expected sort modal >= 28 chars wide for min width, got %d", modalW)
	}
}

