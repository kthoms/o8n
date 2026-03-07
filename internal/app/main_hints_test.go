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
	h, ok := findHint(hints, "Ctrl+Shift+r", "refresh")
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

// --- Per-view dispatch tests (AC 2–5) ---

func TestCurrentViewHints_ModalContextSwitcherActive(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalContextSwitcher

	hints := currentViewHints(m)

	if _, ok := findHint(hints, "Enter", "select"); !ok {
		t.Fatal("expected Enter select hint for context switcher")
	}
	if _, ok := findHint(hints, "↑↓", "nav"); !ok {
		t.Fatal("expected ↑↓ nav hint for context switcher")
	}
	if _, ok := findHint(hints, "Esc", "close"); !ok {
		t.Fatal("expected Esc close hint for context switcher")
	}
	// Table hints must NOT appear when context switcher is open
	if _, ok := findHint(hints, "?", "help"); ok {
		t.Fatal("did not expect table help hint while context switcher is active")
	}
}

func TestCurrentViewHints_SearchPopup(t *testing.T) {
	m := newTestModel(t)
	m.popup.mode = popupModeSearch

	hints := currentViewHints(m)

	if _, ok := findHint(hints, "Enter", "jump"); !ok {
		t.Fatal("expected Enter jump hint for search popup")
	}
	if _, ok := findHint(hints, "↑↓", "select"); !ok {
		t.Fatal("expected ↑↓ select hint for search popup")
	}
	if _, ok := findHint(hints, "Esc", "cancel"); !ok {
		t.Fatal("expected Esc cancel hint for search popup")
	}
	if _, ok := findHint(hints, "?", "help"); ok {
		t.Fatal("did not expect table help hint while search popup is active")
	}
}

func TestCurrentViewHints_SkinPickerPopup(t *testing.T) {
	m := newTestModel(t)
	m.popup.mode = popupModeSkin

	hints := currentViewHints(m)

	if _, ok := findHint(hints, "Enter", "apply"); !ok {
		t.Fatal("expected Enter apply hint for skin picker")
	}
	if _, ok := findHint(hints, "↑↓", "preview"); !ok {
		t.Fatal("expected ↑↓ preview hint for skin picker")
	}
	if _, ok := findHint(hints, "Esc", "revert"); !ok {
		t.Fatal("expected Esc revert hint for skin picker")
	}
	if _, ok := findHint(hints, "?", "help"); ok {
		t.Fatal("did not expect table help hint while skin picker is active")
	}
}

func TestCurrentViewHints_ModalHelpActive(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalHelp

	hints := currentViewHints(m)

	if _, ok := findHint(hints, "↑↓", "scroll"); !ok {
		t.Fatal("expected ↑↓ scroll hint for ModalHelp")
	}
	if _, ok := findHint(hints, "q/?/Esc", "close"); !ok {
		t.Fatal("expected q/?/Esc close hint for ModalHelp")
	}
	if _, ok := findHint(hints, "?", "help"); ok {
		t.Fatal("did not expect table help hint while ModalHelp is active")
	}
}

func TestCurrentViewHints_ModalEditActive(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalEdit

	hints := currentViewHints(m)

	if _, ok := findHint(hints, "Tab", "switch"); !ok {
		t.Fatal("expected Tab switch hint for ModalEdit")
	}
	if _, ok := findHint(hints, "Enter", "save"); !ok {
		t.Fatal("expected Enter save hint for ModalEdit")
	}
	if _, ok := findHint(hints, "Esc", "cancel"); !ok {
		t.Fatal("expected Esc cancel hint for ModalEdit")
	}
	if _, ok := findHint(hints, "?", "help"); ok {
		t.Fatal("did not expect table help hint while ModalEdit is active")
	}
}

func TestCurrentViewHints_ModalFirstRunNoEsc(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalFirstRun

	hints := currentViewHints(m)

	if _, ok := findHint(hints, "↑↓", "nav"); !ok {
		t.Fatal("expected ↑↓ nav hint for ModalFirstRun")
	}
	if _, ok := findHint(hints, "Enter", "select"); !ok {
		t.Fatal("expected Enter select hint for ModalFirstRun")
	}
	// Esc MUST NOT appear — documented exception: selection is required (Story 1.5 AC2)
	for _, h := range hints {
		if h.Key == "Esc" {
			t.Fatalf("Esc hint must not appear in ModalFirstRun (selection required): got %+v", h)
		}
	}
}

func TestCurrentViewHints_ModalConfirmDeleteActive(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalConfirmDelete

	hints := currentViewHints(m)

	if _, ok := findHint(hints, "Enter", "confirm"); !ok {
		t.Fatal("expected Enter confirm hint for ModalConfirmDelete")
	}
	if _, ok := findHint(hints, "Esc", "cancel"); !ok {
		t.Fatal("expected Esc cancel hint for ModalConfirmDelete")
	}
	if _, ok := findHint(hints, "?", "help"); ok {
		t.Fatal("did not expect table help hint while ModalConfirmDelete is active")
	}
}

func TestCurrentViewHints_MainTableUnchanged(t *testing.T) {
	m := newTestModel(t)
	// No modal, no popup — must dispatch to tableViewHints

	hints := currentViewHints(m)

	if _, ok := findHint(hints, "?", "help"); !ok {
		t.Fatal("expected help hint in main table view")
	}
	if _, ok := findHint(hints, ":", "switch"); !ok {
		t.Fatal("expected : switch hint in main table view")
	}
	if _, ok := findHint(hints, "↑↓", "nav"); !ok {
		t.Fatal("expected ↑↓ nav hint in main table view")
	}
}

func TestCurrentViewHints_ModalActionMenuActive(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalActionMenu

	hints := currentViewHints(m)

	if _, ok := findHint(hints, "↑↓", "nav"); !ok {
		t.Fatal("expected ↑↓ nav hint for ModalActionMenu")
	}
	if _, ok := findHint(hints, "Enter", "run"); !ok {
		t.Fatal("expected Enter run hint for ModalActionMenu")
	}
	if _, ok := findHint(hints, "Esc", "close"); !ok {
		t.Fatal("expected Esc close hint for ModalActionMenu")
	}
	if _, ok := findHint(hints, "?", "help"); ok {
		t.Fatal("did not expect table help hint while ModalActionMenu is active")
	}
}
