package app

// layout_validation_test.go — Story 4.1: 120×20 Rendering Validation
//
// These tests verify that the layout engine correctly allocates terminal space
// at the 120×20 minimum viewport required for VSCode and IntelliJ IDEA targets.
//
// MANUAL VALIDATION NOTE (AC: 5):
// Run `./o6n --no-splash` in a terminal resized to exactly 120×20 in:
//   - VSCode Integrated Terminal (View → Terminal, resize to 120×20)
//   - IntelliJ IDEA Terminal (resize terminal pane to 120×20)
// Confirm: header visible (2 rows), table body visible (~16 rows), footer visible (1 row).
// No text should overflow, wrap, or appear truncated for critical content.

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// ── Task 1: Viewport Height Calibration ──────────────────────────────────────

// TestWindowSizeMsg_120x20_SetsCorrectPaneDimensions verifies that receiving a
// WindowSizeMsg at 120×20 sets the model's layout state to the correct values.
// Row budget: 20 - 2(header) - 1(contextSelection) - 1(footer) = 16 (paneHeight).
func TestWindowSizeMsg_120x20_SetsCorrectPaneDimensions(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	res, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 20})
	m2 := res.(model)

	if m2.lastWidth != 120 {
		t.Errorf("expected lastWidth=120, got %d", m2.lastWidth)
	}
	if m2.lastHeight != 20 {
		t.Errorf("expected lastHeight=20, got %d", m2.lastHeight)
	}
	if m2.paneWidth != 120 {
		t.Errorf("expected paneWidth=120, got %d", m2.paneWidth)
	}
	// contentHeight = 20 - 2 - 1 - 1 = 16
	if m2.paneHeight != 16 {
		t.Errorf("expected paneHeight=16 (20-2-1-1), got %d", m2.paneHeight)
	}
}

// TestWindowSizeMsg_120x20_TableDimensionsCorrect verifies that the table widget
// is configured with the correct inner width and height after a 120×20 resize.
// tableInner = paneWidth - 4 (border 2 + padding 2) = 116
// table.Height = paneHeight - 1 = 15
func TestWindowSizeMsg_120x20_TableDimensionsCorrect(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	res, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 20})
	m2 := res.(model)

	tableWidth := m2.table.Width()
	if tableWidth > 116 {
		t.Errorf("expected table.Width <= 116 (paneWidth-4), got %d", tableWidth)
	}
	if tableWidth < 10 {
		t.Errorf("expected table.Width >= 10 (minimum), got %d", tableWidth)
	}
	// bubbles table.SetHeight(h) internally subtracts the header row, so Height() = h-1.
	// SetHeight(contentHeight-1) = SetHeight(15), so Height() = 14.
	tableHeight := m2.table.Height()
	if tableHeight != 14 {
		t.Errorf("expected table.Height=14 (paneHeight-2=16-2), got %d", tableHeight)
	}
}

// TestWindowSizeMsg_SmallTerminal_DoesNotPanic verifies that very small terminals
// hit the safety minimums without panicking.
func TestWindowSizeMsg_SmallTerminal_DoesNotPanic(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	// Should not panic even at very small sizes
	res, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	m2 := res.(model)

	if m2.paneHeight < 3 {
		t.Errorf("expected paneHeight >= 3 (minimum), got %d", m2.paneHeight)
	}
	if m2.paneWidth < 10 {
		t.Errorf("expected paneWidth >= 10 (minimum), got %d", m2.paneWidth)
	}
}

// TestRenderCompactHeader_120Wide_IsExactly2Lines verifies the header occupies
// exactly 2 rows at 120 column width. lipgloss.Place enforces this constraint.
func TestRenderCompactHeader_120Wide_IsExactly2Lines(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120
	m.lastHeight = 20
	m.splashActive = false
	m.currentEnv = "local"

	header := m.renderCompactHeader(120)
	lines := strings.Split(header, "\n")
	// Strip trailing blank lines that lipgloss.Place may add
	for len(lines) > 0 && strings.TrimSpace(stripANSIForTest(lines[len(lines)-1])) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) != 2 {
		t.Errorf("expected exactly 2 header lines at width=120, got %d: %q", len(lines), header)
	}
}

// TestRenderCompactHeader_120Wide_LinesDoNotExceedWidth verifies no line in the
// header exceeds 120 visible characters (ANSI-stripped).
func TestRenderCompactHeader_120Wide_LinesDoNotExceedWidth(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120
	m.lastHeight = 20
	m.splashActive = false
	m.currentEnv = "local"

	header := m.renderCompactHeader(120)
	for i, line := range strings.Split(header, "\n") {
		visibleWidth := len([]rune(stripANSIForTest(line)))
		if visibleWidth > 120 {
			t.Errorf("header line %d exceeds 120 chars: width=%d", i+1, visibleWidth)
		}
	}
}

// TestRenderBoxWithTitle_FixedHeight verifies that renderBoxWithTitle always
// produces exactly totalHeight rows — no vertical overflow possible.
func TestRenderBoxWithTitle_FixedHeight(t *testing.T) {
	m := newTestModel(t)

	// With paneHeight=16 at 120×20, the mainBox must be exactly 16 rows.
	box := renderBoxWithTitle("some content\nline2\nline3", 120, 16, "Test Title", m.styles.BorderFocus)
	lines := strings.Split(box, "\n")
	// Trim trailing empty line from strings.Split artifact
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) != 16 {
		t.Errorf("expected renderBoxWithTitle to produce exactly 16 rows, got %d", len(lines))
	}
}

// ── Task 2: Horizontal Space Management ──────────────────────────────────────

// TestBuildColumnsFor_AtTableWidth116_NoOverflow verifies that for all resource
// types defined in the test config, total column widths never exceed 116 chars
// (the available inner table width at paneWidth=120, minus border and padding).
func TestBuildColumnsFor_AtTableWidth116_NoOverflow(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120

	tableNames := []string{"process-definition", "process-instance", "process-variables"}
	for _, name := range tableNames {
		cols := m.buildColumnsFor(name, 116)
		if len(cols) == 0 {
			t.Errorf("buildColumnsFor(%q, 116): got zero columns", name)
			continue
		}
		total := 0
		for _, c := range cols {
			total += c.Width
		}
		if total > 116 {
			t.Errorf("buildColumnsFor(%q, 116): total column width %d > 116", name, total)
		}
	}
}

// TestBuildColumnsFor_AtTableWidth116_AtLeastOneColumn verifies that column
// hiding never removes ALL columns — at least one column always remains visible.
func TestBuildColumnsFor_AtTableWidth116_AtLeastOneColumn(t *testing.T) {
	m := newTestModel(t)

	tableNames := []string{"process-definition", "process-instance", "process-variables"}
	for _, name := range tableNames {
		cols := m.buildColumnsFor(name, 116)
		if len(cols) < 1 {
			t.Errorf("buildColumnsFor(%q, 116): expected at least 1 visible column, got 0", name)
		}
	}
}

// TestBuildColumnsFor_VeryNarrowWidth_AtLeastOneColumn verifies that even at
// very narrow widths the algorithm leaves at least one column rather than
// hiding everything and returning empty.
func TestBuildColumnsFor_VeryNarrowWidth_AtLeastOneColumn(t *testing.T) {
	m := newTestModel(t)

	cols := m.buildColumnsFor("process-definition", 20)
	if len(cols) < 1 {
		t.Errorf("expected at least 1 column even at very narrow width=20, got 0")
	}
}

// TestBuildColumnsFor_LastColumnStretchesToFillWidth verifies that the last
// visible column is stretched so the total fills the available width exactly.
func TestBuildColumnsFor_LastColumnStretchesToFillWidth(t *testing.T) {
	m := newTestModel(t)

	cols := m.buildColumnsFor("process-definition", 116)
	total := 0
	for _, c := range cols {
		total += c.Width
	}
	// After stretching, total must equal exactly 116
	if total != 116 {
		t.Errorf("expected columns to stretch to fill exactly 116, got total=%d", total)
	}
}

// ── Task 2: Hint Visibility at 120 Columns ───────────────────────────────────

// TestFilterHints_At120Cols_AlwaysShowHintsPresent verifies that hints with
// MinWidth=0 (always-show) are included in the filtered output at width=120.
func TestFilterHints_At120Cols_AlwaysShowHintsPresent(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120

	hints := filterHints(tableViewHints(m), 120)
	keys := make(map[string]bool)
	for _, h := range hints {
		keys[h.Key] = true
	}

	alwaysShowKeys := []string{"?", ":", "↑↓", "/", "PgDn/PgUp"}
	for _, key := range alwaysShowKeys {
		if !keys[key] {
			t.Errorf("expected always-show hint %q to be present at width=120", key)
		}
	}
}

// TestFilterHints_At120Cols_WidthGatedHintsPresent verifies that all hints with
// MinWidth ≤ 120 are included (at 120 wide, all standard table hints should show).
func TestFilterHints_At120Cols_WidthGatedHintsPresent(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 120

	hints := filterHints(tableViewHints(m), 120)
	keys := make(map[string]bool)
	for _, h := range hints {
		keys[h.Key] = true
	}

	// All hints in tableViewHints have MaxWidth <= 112, which is <= 120.
	// So all standard hints should be visible at 120 columns.
	expectedVisible := []string{"s", "Ctrl+Shift+r", "Ctrl+t", "Ctrl+e", "Ctrl+Space", "J", "Ctrl+c"}
	for _, key := range expectedVisible {
		if !keys[key] {
			t.Errorf("expected hint %q (MinWidth<=120) to be visible at width=120", key)
		}
	}
}

// TestFilterHints_At80Cols_NarrowHintsHidden verifies that width-gated hints are
// correctly hidden when the terminal is narrower than their MinWidth threshold.
func TestFilterHints_At80Cols_NarrowHintsHidden(t *testing.T) {
	m := newTestModel(t)
	m.lastWidth = 80

	hints := filterHints(tableViewHints(m), 80)
	keys := make(map[string]bool)
	for _, h := range hints {
		keys[h.Key] = true
	}

	// s (MinWidth=88), Ctrl+Shift+r (90), Ctrl+Space (100), J (112) should be hidden at 80.
	hiddenKeys := []string{"s", "Ctrl+Shift+r", "Ctrl+t", "Ctrl+e", "Ctrl+Space", "J", "Ctrl+c"}
	for _, key := range hiddenKeys {
		if keys[key] {
			t.Errorf("expected hint %q to be hidden at width=80, but it was visible", key)
		}
	}
	// Always-show hints must still be present.
	alwaysShowKeys := []string{"?", ":", "↑↓", "/"}
	for _, key := range alwaysShowKeys {
		if !keys[key] {
			t.Errorf("expected always-show hint %q to still be present at width=80", key)
		}
	}
}

// TestFilterHints_PriorityOrdering verifies that filterHints returns hints in
// ascending priority order (Priority 1 = highest, shown first).
func TestFilterHints_PriorityOrdering(t *testing.T) {
	m := newTestModel(t)

	hints := filterHints(tableViewHints(m), 120)
	for i := 1; i < len(hints); i++ {
		if hints[i].Priority < hints[i-1].Priority {
			t.Errorf("filterHints result not in priority order at index %d: priority %d before %d",
				i, hints[i-1].Priority, hints[i].Priority)
		}
	}
}

// ── Task 3: View Composition at 120×20 ───────────────────────────────────────

// TestView_120x20_HeightFitsTerminal verifies that the rendered View() output
// does not exceed 20 lines when the terminal is 120×20 with no popup active.
// lipgloss.Place(120, 20, ...) is used in View(), ensuring the output is padded
// to exactly 20 rows.
func TestView_120x20_HeightFitsTerminal(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 20
	m.paneWidth = 120
	m.paneHeight = 16
	m.currentRoot = "process-definition"
	m.currentEnv = "local"
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 60}, {Title: "NAME", Width: 56}})
	m.table.SetRows([]table.Row{})
	m.activeModal = ModalNone

	rendered := m.View()
	lines := strings.Split(rendered, "\n")
	// lipgloss.Place pads to height; trailing blank line from Split is expected
	// Remove final empty string from Split artifact
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > 20 {
		t.Errorf("expected View() height <= 20 rows at 120×20, got %d rows", len(lines))
	}
}

// TestView_120x20_NoLineExceedsWidth verifies that no rendered line in View()
// exceeds 120 visible characters (ANSI-stripped) at 120×20.
func TestView_120x20_NoLineExceedsWidth(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.lastWidth = 120
	m.lastHeight = 20
	m.paneWidth = 120
	m.paneHeight = 16
	m.currentRoot = "process-definition"
	m.currentEnv = "local"
	m.table.SetColumns([]table.Column{{Title: "ID", Width: 60}, {Title: "NAME", Width: 56}})
	m.table.SetRows([]table.Row{})
	m.activeModal = ModalNone

	rendered := m.View()
	for i, line := range strings.Split(rendered, "\n") {
		visibleWidth := len([]rune(stripANSIForTest(line)))
		if visibleWidth > 120 {
			t.Errorf("line %d exceeds 120 visible chars (width=%d): %q", i+1, visibleWidth, stripANSIForTest(line)[:min(80, len(stripANSIForTest(line)))])
		}
	}
}
