package app

import "testing"

func findHint(hints []Hint, key, label string) (Hint, bool) {
	for _, h := range hints {
		if h.Key == key && h.Label == label {
			return h, true
		}
	}
	return Hint{}, false
}

func TestFilterHints_MinWidthZeroAlwaysVisible(t *testing.T) {
	out := filterHints([]Hint{{Key: "?", Label: "help", MinWidth: 0, Priority: 1}}, 1)
	if len(out) != 1 {
		t.Fatalf("expected 1 hint, got %d", len(out))
	}
}

func TestFilterHints_ExcludesWhenMinWidthGreaterThanWidth(t *testing.T) {
	out := filterHints([]Hint{{Key: "Ctrl+r", Label: "refresh", MinWidth: 90, Priority: 6}}, 80)
	if len(out) != 0 {
		t.Fatalf("expected 0 hints, got %d", len(out))
	}
}

func TestFilterHints_IncludesWhenMinWidthMet(t *testing.T) {
	out := filterHints([]Hint{{Key: "Ctrl+r", Label: "refresh", MinWidth: 90, Priority: 6}}, 90)
	if len(out) != 1 {
		t.Fatalf("expected 1 hint, got %d", len(out))
	}
}

func TestFilterHints_SortsByPriorityAscending(t *testing.T) {
	out := filterHints([]Hint{
		{Key: "c", Label: "c", MinWidth: 0, Priority: 3},
		{Key: "a", Label: "a", MinWidth: 0, Priority: 1},
		{Key: "b", Label: "b", MinWidth: 0, Priority: 2},
	}, 80)
	if len(out) != 3 {
		t.Fatalf("expected 3 hints, got %d", len(out))
	}
	if out[0].Priority != 1 || out[1].Priority != 2 || out[2].Priority != 3 {
		t.Fatalf("expected sorted priorities [1,2,3], got [%d,%d,%d]", out[0].Priority, out[1].Priority, out[2].Priority)
	}
}

func TestFilterHints_EmptyInput(t *testing.T) {
	out := filterHints(nil, 80)
	if len(out) != 0 {
		t.Fatalf("expected empty result, got %d", len(out))
	}
}

func TestCurrentViewHints_AlwaysIncludesHelp(t *testing.T) {
	m := newTestModel(t)
	hints := currentViewHints(m)
	h, ok := findHint(hints, "?", "help")
	if !ok {
		t.Fatal("expected help hint")
	}
	if h.Priority != 1 || h.MinWidth != 0 {
		t.Fatalf("expected help hint priority=1,minWidth=0; got priority=%d,minWidth=%d", h.Priority, h.MinWidth)
	}
}

func TestCurrentViewHints_EscBackOnlyWhenNavigationStackNonEmpty(t *testing.T) {
	m := newTestModel(t)
	hints := currentViewHints(m)
	if _, ok := findHint(hints, "Esc", "back"); ok {
		t.Fatal("did not expect Esc back hint without navigation stack")
	}

	m.navigationStack = append(m.navigationStack, viewState{})
	hints = currentViewHints(m)
	if _, ok := findHint(hints, "Esc", "back"); !ok {
		t.Fatal("expected Esc back hint with navigation stack")
	}
}

func TestCurrentViewHints_EnterDrillOnlyWhenTableHasDrilldown(t *testing.T) {
	m := newTestModel(t)
	m.currentRoot = "process-definition"
	hints := currentViewHints(m)
	if _, ok := findHint(hints, "Enter", "drill"); !ok {
		t.Fatal("expected Enter drill hint for root with drilldown")
	}

	m.currentRoot = "process-variables"
	hints = currentViewHints(m)
	if _, ok := findHint(hints, "Enter", "drill"); ok {
		t.Fatal("did not expect Enter drill hint for root without drilldown")
	}
}

func TestCurrentViewHints_RefreshHintMinWidthNinety(t *testing.T) {
	m := newTestModel(t)
	hints := currentViewHints(m)
	h, ok := findHint(hints, "Ctrl+r", "refresh")
	if !ok {
		t.Fatal("expected refresh hint")
	}
	if h.MinWidth != 90 {
		t.Fatalf("expected refresh MinWidth=90, got %d", h.MinWidth)
	}
}

func TestCurrentViewHints_EditHintOnlyWhenEditableColumns(t *testing.T) {
	m := newTestModel(t)
	// process-definition has no editable columns
	m.currentRoot = "process-definition"
	m.breadcrumb = []string{"process-definition"}
	hints := currentViewHints(m)
	if _, ok := findHint(hints, "e", "edit"); ok {
		t.Fatal("did not expect edit hint for table without editable columns")
	}

	// process-variables has value column with Editable: true
	m.currentRoot = "process-variables"
	m.breadcrumb = []string{"process-variables"}
	hints = currentViewHints(m)
	if _, ok := findHint(hints, "e", "edit"); !ok {
		t.Fatal("expected edit hint for table with editable columns")
	}
}
