package app

import (
	"testing"

	"github.com/kthoms/o8n/internal/config"
)

// boolPtr returns a pointer to b, for use in ColumnDef.Visible.
func boolPtr(b bool) *bool { return &b }

// buildTestModel returns a minimal model with the given table definitions.
func buildTestModel(tables []config.TableDef) model {
	cfg := &config.Config{Tables: tables}
	return newModel(cfg)
}

// TestBuildColumnsDefaultTypeWidths verifies implicit widths from column types.
func TestBuildColumnsDefaultTypeWidths(t *testing.T) {
	cases := []struct {
		colType  string
		wantMin  int
		wantMax  int
	}{
		{"bool", 6, 6},
		{"int", 8, 8},
		{"float", 10, 10},
		{"datetime", 20, 20},
		{"id", 36, 36},
		{"string", 20, 20},
		{"", 20, 20}, // empty type → string default
	}

	for _, tc := range cases {
		got := config.DefaultTypeWidth(tc.colType)
		if got < tc.wantMin || got > tc.wantMax {
			t.Errorf("DefaultTypeWidth(%q) = %d, want %d–%d", tc.colType, got, tc.wantMin, tc.wantMax)
		}
	}
}

// TestBuildColumnsOrderPreserved checks that columns appear in config order.
func TestBuildColumnsOrderPreserved(t *testing.T) {
	m := buildTestModel([]config.TableDef{{
		Name: "t",
		Columns: []config.ColumnDef{
			{Name: "alpha"},
			{Name: "beta"},
			{Name: "gamma"},
		},
	}})
	cols := m.buildColumnsFor("t", 200)
	if len(cols) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(cols))
	}
	want := []string{"ALPHA", "BETA", "GAMMA"}
	for i, c := range cols {
		if c.Title != want[i] {
			t.Errorf("col[%d] title = %q, want %q", i, c.Title, want[i])
		}
	}
}

// TestBuildColumnsExplicitHiddenSkipped verifies Visible=false columns are never shown.
func TestBuildColumnsExplicitHiddenSkipped(t *testing.T) {
	m := buildTestModel([]config.TableDef{{
		Name: "t",
		Columns: []config.ColumnDef{
			{Name: "a"},
			{Name: "b", Visible: boolPtr(false)},
			{Name: "c"},
		},
	}})
	cols := m.buildColumnsFor("t", 200)
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if cols[0].Title != "A" || cols[1].Title != "C" {
		t.Errorf("unexpected titles: %v", cols)
	}
}

// TestBuildColumnsHideDefaultOrderLastFirst verifies that unspecified columns
// are hidden rightmost-first when space is tight.
func TestBuildColumnsHideDefaultOrderLastFirst(t *testing.T) {
	// Three 20-char columns = 60 chars total.
	// Provide only 42 chars so the last column (C) must be hidden.
	m := buildTestModel([]config.TableDef{{
		Name: "t",
		Columns: []config.ColumnDef{
			{Name: "a", Width: 20},
			{Name: "b", Width: 20},
			{Name: "c", Width: 20},
		},
	}})
	cols := m.buildColumnsFor("t", 42)
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns (a, b) with width=42, got %d: %v", len(cols), cols)
	}
	if cols[0].Title != "A" {
		t.Errorf("first column should be A, got %q", cols[0].Title)
	}
	if cols[1].Title != "B" {
		t.Errorf("second column should be B, got %q", cols[1].Title)
	}
}

// TestBuildColumnsHideDefaultBeforeConfigured verifies that columns without
// hide_order are hidden before columns with hide_order set.
func TestBuildColumnsHideDefaultBeforeConfigured(t *testing.T) {
	// Column layout: A (no hide_order), B (hide_order=1), C (no hide_order)
	// Width budget: forces hiding of one column. C is last with no hide_order → hidden first.
	m := buildTestModel([]config.TableDef{{
		Name: "t",
		Columns: []config.ColumnDef{
			{Name: "a", Width: 20},
			{Name: "b", Width: 20, HideOrder: 1},
			{Name: "c", Width: 20},
		},
	}})
	// 42 chars fits two 20-char columns. C has no hide_order and is last → hidden first.
	cols := m.buildColumnsFor("t", 42)
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns with width=42, got %d: %v", len(cols), cols)
	}
	titles := map[string]bool{}
	for _, col := range cols {
		titles[col.Title] = true
	}
	if !titles["A"] || !titles["B"] {
		t.Errorf("expected A and B to remain visible, got %v", cols)
	}
	if titles["C"] {
		t.Errorf("C should have been hidden, but it's still visible")
	}
}

// TestBuildColumnsHideOrderRespected verifies higher HideOrder stays visible longer.
func TestBuildColumnsHideOrderRespected(t *testing.T) {
	// A (hide_order=2), B (hide_order=1), C (hide_order=3)
	// All configured. B should be hidden first (lowest non-zero order), then A, then C.
	m := buildTestModel([]config.TableDef{{
		Name: "t",
		Columns: []config.ColumnDef{
			{Name: "a", Width: 20, HideOrder: 2},
			{Name: "b", Width: 20, HideOrder: 1},
			{Name: "c", Width: 20, HideOrder: 3},
		},
	}})
	// 42 chars fits two columns → B (lowest hide_order=1) is hidden first.
	cols := m.buildColumnsFor("t", 42)
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns with width=42, got %d: %v", len(cols), cols)
	}
	titles := map[string]bool{}
	for _, col := range cols {
		titles[col.Title] = true
	}
	if titles["B"] {
		t.Errorf("B (hide_order=1) should have been hidden first, but it's still visible")
	}
	if !titles["A"] || !titles["C"] {
		t.Errorf("expected A and C to remain, got %v", cols)
	}
}

// TestBuildColumnsMinWidthEnforced verifies that a column is hidden when totalWidth
// would force it below its header title width (implicit minimum).
func TestBuildColumnsMinWidthEnforced(t *testing.T) {
	// Column "longname" has a 8-char header. If we give exactly 8 chars per col
	// but total budget forces hiding, it should be hidden rather than shrunk.
	m := buildTestModel([]config.TableDef{{
		Name: "t",
		Columns: []config.ColumnDef{
			{Name: "id", Width: 36},     // needs 36 chars
			{Name: "longname", Width: 8}, // needs 8 chars (header = "LONGNAME" = 8)
		},
	}})
	// 36 chars budget: only the "id" column fits; "longname" must be hidden.
	cols := m.buildColumnsFor("t", 36)
	if len(cols) != 1 {
		t.Fatalf("expected 1 column at width=36, got %d: %v", len(cols), cols)
	}
	if cols[0].Title != "ID" {
		t.Errorf("expected ID column to remain, got %q", cols[0].Title)
	}
}

// TestBuildColumnsLastColumnStretches verifies that the last visible column
// stretches to fill any remaining horizontal space.
func TestBuildColumnsLastColumnStretches(t *testing.T) {
	m := buildTestModel([]config.TableDef{{
		Name: "t",
		Columns: []config.ColumnDef{
			{Name: "a", Width: 10},
			{Name: "b", Width: 10},
		},
	}})
	cols := m.buildColumnsFor("t", 50)
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	// a stays at 10, b stretches from 10 to 40
	if cols[0].Width != 10 {
		t.Errorf("first column width = %d, want 10", cols[0].Width)
	}
	if cols[1].Width != 40 {
		t.Errorf("last column width = %d, want 40 (stretched)", cols[1].Width)
	}
}

// TestBuildColumnsExplicitWidthOverridesType verifies that an explicit Width
// field takes precedence over the type default.
func TestBuildColumnsExplicitWidthOverridesType(t *testing.T) {
	m := buildTestModel([]config.TableDef{{
		Name: "t",
		Columns: []config.ColumnDef{
			{Name: "x", Type: "id", Width: 12}, // id default=36, explicit=12
		},
	}})
	cols := m.buildColumnsFor("t", 200)
	if len(cols) != 1 {
		t.Fatalf("expected 1 column, got %d", len(cols))
	}
	if cols[0].Width != 200 {
		t.Errorf("width = %d, want 200 (explicit 12, stretched to fill totalWidth)", cols[0].Width)
	}
}

// TestHideSequenceUnspecifiedRightmostFirst verifies the HideSequence helper
// returns unspecified columns from rightmost to leftmost.
func TestHideSequenceUnspecifiedRightmostFirst(t *testing.T) {
	cols := []config.ColumnDef{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
	}
	seq := config.HideSequence(cols)
	if len(seq) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(seq))
	}
	want := []int{2, 1, 0} // rightmost first
	for i, idx := range seq {
		if idx != want[i] {
			t.Errorf("seq[%d] = %d, want %d", i, idx, want[i])
		}
	}
}

// TestHideSequenceConfiguredAfterUnspecified verifies that configured (HideOrder>0)
// columns appear after unspecified ones in the hide sequence.
func TestHideSequenceConfiguredAfterUnspecified(t *testing.T) {
	cols := []config.ColumnDef{
		{Name: "a", HideOrder: 2}, // configured, hide later
		{Name: "b"},               // unspecified, hide first
		{Name: "c", HideOrder: 1}, // configured, hide before a
	}
	seq := config.HideSequence(cols)
	// expected: b(idx=1) first, then c(idx=2, order=1), then a(idx=0, order=2)
	want := []int{1, 2, 0}
	for i, idx := range seq {
		if idx != want[i] {
			t.Errorf("seq[%d] = %d, want %d (full seq: %v)", i, idx, want[i], seq)
		}
	}
}

// TestColumnDefIsVisible verifies the IsVisible method semantics.
func TestColumnDefIsVisible(t *testing.T) {
	visible := true
	hidden := false

	cases := []struct {
		col  config.ColumnDef
		want bool
	}{
		{config.ColumnDef{Name: "a"}, true},             // nil Visible = visible by default
		{config.ColumnDef{Name: "b", Visible: &visible}, true},  // explicit true
		{config.ColumnDef{Name: "c", Visible: &hidden}, false},  // explicit false
	}
	for _, tc := range cases {
		got := tc.col.IsVisible()
		if got != tc.want {
			t.Errorf("IsVisible for %q: got %v, want %v", tc.col.Name, got, tc.want)
		}
	}
}
