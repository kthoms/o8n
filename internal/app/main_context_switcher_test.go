package app

import (
	"strings"
	"testing"
)

// TestColonKey_OpensContextSwitcherModal verifies that pressing `:` when no modal
// is open activates ModalContextSwitcher and resets popup state.
func TestColonKey_OpensContextSwitcherModal(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false

	m2, _ := sendKeyString(m, ":")

	if m2.activeModal != ModalContextSwitcher {
		t.Fatalf("expected ModalContextSwitcher after ':', got %v", m2.activeModal)
	}
	if m2.popup.input != "" {
		t.Fatalf("expected popup.input cleared, got %q", m2.popup.input)
	}
	if m2.popup.cursor != -1 {
		t.Fatalf("expected popup.cursor -1, got %d", m2.popup.cursor)
	}
}

// TestColonKey_WhenModalOpenAppendsToFilter verifies that pressing `:` while
// ModalContextSwitcher is active appends `:` to the filter input (not toggle-close),
// because the rune handler captures it before the colon-toggle logic.
func TestColonKey_WhenModalOpenAppendsToFilter(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.activeModal = ModalContextSwitcher
	m.popup.input = "proc"

	m2, _ := sendKeyString(m, ":")

	if m2.activeModal != ModalContextSwitcher {
		t.Fatalf("expected ModalContextSwitcher to remain open, got %v", m2.activeModal)
	}
	if m2.popup.input != "proc:" {
		t.Fatalf("expected popup.input 'proc:' after typing colon, got %q", m2.popup.input)
	}
}

// TestContextSwitcher_EscClosesModal verifies that Esc dismisses ModalContextSwitcher
// and resets popup input, cursor, and offset.
func TestContextSwitcher_EscClosesModal(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalContextSwitcher
	m.popup.input = "proc"
	m.popup.cursor = 2
	m.popup.offset = 4

	m2, _ := sendKeyString(m, "esc")

	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after Esc, got %v", m2.activeModal)
	}
	if m2.popup.input != "" {
		t.Fatalf("expected popup.input cleared, got %q", m2.popup.input)
	}
	if m2.popup.cursor != -1 {
		t.Fatalf("expected popup.cursor -1, got %d", m2.popup.cursor)
	}
	if m2.popup.offset != 0 {
		t.Fatalf("expected popup.offset 0 after Esc, got %d", m2.popup.offset)
	}
}

// TestContextSwitcher_TypingFiltersInput verifies that typing rune characters
// appends to popup.input and resets cursor to -1.
func TestContextSwitcher_TypingFiltersInput(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalContextSwitcher
	m.popup.input = ""
	m.popup.cursor = 3

	m2, _ := sendKeyString(m, "p")

	if m2.popup.input != "p" {
		t.Fatalf("expected popup.input 'p' after typing, got %q", m2.popup.input)
	}
	if m2.popup.cursor != -1 {
		t.Fatalf("expected cursor reset to -1 after typing, got %d", m2.popup.cursor)
	}

	m3, _ := sendKeyString(m2, "r")
	if m3.popup.input != "pr" {
		t.Fatalf("expected popup.input 'pr' after second char, got %q", m3.popup.input)
	}
}

// TestContextSwitcher_BackspaceRemovesLastChar verifies that Backspace trims the
// last character from popup.input and resets cursor.
func TestContextSwitcher_BackspaceRemovesLastChar(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalContextSwitcher
	m.popup.input = "proc"
	m.popup.cursor = 1

	m2, _ := sendKeyString(m, "backspace")

	if m2.popup.input != "pro" {
		t.Fatalf("expected popup.input 'pro' after backspace, got %q", m2.popup.input)
	}
	if m2.popup.cursor != -1 {
		t.Fatalf("expected cursor reset to -1 after backspace, got %d", m2.popup.cursor)
	}
}

// TestContextSwitcher_BackspaceOnEmptyInput verifies that Backspace on empty input is a no-op.
func TestContextSwitcher_BackspaceOnEmptyInput(t *testing.T) {
	m := newTestModel(t)
	m.activeModal = ModalContextSwitcher
	m.popup.input = ""

	m2, _ := sendKeyString(m, "backspace")

	if m2.popup.input != "" {
		t.Fatalf("expected popup.input still empty after backspace, got %q", m2.popup.input)
	}
}

// TestContextSwitcher_DownKeyAdvancesCursor verifies that Down moves popup.cursor
// from -1 to 0 when items are available.
func TestContextSwitcher_DownKeyAdvancesCursor(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.rootContexts = []string{"process-definition", "process-instance", "task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = ""
	m.popup.cursor = -1

	m2, _ := sendKeyString(m, "down")

	if m2.popup.cursor != 0 {
		t.Fatalf("expected cursor 0 after Down from -1, got %d", m2.popup.cursor)
	}

	m3, _ := sendKeyString(m2, "down")

	if m3.popup.cursor != 1 {
		t.Fatalf("expected cursor 1 after second Down, got %d", m3.popup.cursor)
	}
}

// TestContextSwitcher_DownKeyWrapsAround verifies that Down wraps cursor from last
// item back to 0.
func TestContextSwitcher_DownKeyWrapsAround(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.rootContexts = []string{"process-definition", "process-instance"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = ""
	m.popup.cursor = 1 // last item

	m2, _ := sendKeyString(m, "down")

	if m2.popup.cursor != 0 {
		t.Fatalf("expected cursor wrap to 0, got %d", m2.popup.cursor)
	}
}

// TestContextSwitcher_UpKeyWrapsAround verifies that Up wraps cursor from 0
// back to last item.
func TestContextSwitcher_UpKeyWrapsAround(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.rootContexts = []string{"process-definition", "process-instance", "task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = ""
	m.popup.cursor = 0

	m2, _ := sendKeyString(m, "up")

	if m2.popup.cursor != 2 {
		t.Fatalf("expected cursor wrap to 2 (last), got %d", m2.popup.cursor)
	}
}

// TestPopupItems_ContextSwitcher_FiltersByPrefix verifies that popupItems returns only
// items whose names contain popup.input when ModalContextSwitcher is active.
func TestPopupItems_ContextSwitcher_FiltersByPrefix(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition", "process-instance", "task", "incident"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "process"

	items := m.popupItems()

	if len(items) != 2 {
		t.Fatalf("expected 2 items matching 'process', got %d: %v", len(items), items)
	}
	for _, item := range items {
		if !strings.Contains(item, "process") {
			t.Errorf("expected item containing 'process', got %q", item)
		}
	}
}

// TestPopupItems_ContextSwitcher_FiltersByContains verifies that popupItems uses
// contains-matching so partial middle/suffix terms find resources.
func TestPopupItems_ContextSwitcher_FiltersByContains(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition", "process-instance", "task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "instance" // not a prefix of "process-instance"

	items := m.popupItems()

	if len(items) != 1 || items[0] != "process-instance" {
		t.Fatalf("expected ['process-instance'] via contains match, got %v", items)
	}
}

// TestPopupItems_ContextSwitcher_EmptyInputReturnsAll verifies that an empty filter
// returns all rootContexts.
func TestPopupItems_ContextSwitcher_EmptyInputReturnsAll(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition", "task", "incident"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = ""

	items := m.popupItems()

	if len(items) != 3 {
		t.Fatalf("expected all 3 items with empty input, got %d: %v", len(items), items)
	}
}

// TestPopupItems_ContextSwitcher_NoMatchReturnsEmpty verifies that a non-matching
// prefix returns an empty slice.
func TestPopupItems_ContextSwitcher_NoMatchReturnsEmpty(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition", "task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "xyz"

	items := m.popupItems()

	if len(items) != 0 {
		t.Fatalf("expected no items for 'xyz', got %d: %v", len(items), items)
	}
}

// TestContextSwitcher_SelectViaCursor verifies that Enter with popup.cursor set selects
// the item at the cursor position (even when popup.input doesn't match exactly).
func TestContextSwitcher_SelectViaCursor(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.rootContexts = []string{"process-definition", "process-instance", "task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = ""
	m.popup.cursor = 2 // "task"

	m2, _ := sendKeyString(m, "enter")

	if m2.currentRoot != "task" {
		t.Fatalf("expected currentRoot 'task' (cursor=2), got %q", m2.currentRoot)
	}
	if m2.activeModal != ModalNone {
		t.Fatalf("expected ModalNone after selection, got %v", m2.activeModal)
	}
}

// TestContextSwitcher_EnterWithNoMatchIsNoop verifies that Enter when popup items
// are empty (no match) does nothing.
func TestContextSwitcher_EnterWithNoMatchIsNoop(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.rootContexts = []string{"process-definition"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "xyz" // no match
	m.popup.cursor = -1
	m.currentRoot = "process-definition"

	m2, _ := sendKeyString(m, "enter")

	// Modal should stay open, root should not change.
	if m2.activeModal != ModalContextSwitcher {
		t.Fatalf("expected ModalContextSwitcher to stay open with no match, got %v", m2.activeModal)
	}
	if m2.currentRoot != "process-definition" {
		t.Fatalf("expected currentRoot unchanged, got %q", m2.currentRoot)
	}
}

// TestRenderContextSwitcherBody_ContainsTitleAndInput verifies that
// renderContextSwitcherBody includes the title text and filter input line.
func TestRenderContextSwitcherBody_ContainsTitleAndInput(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition", "task"}
	m.popup.input = "proc"

	body := m.renderContextSwitcherBody()

	if !strings.Contains(body, "Switch Resource Context") {
		t.Error("expected body to contain title 'Switch Resource Context'")
	}
	if !strings.Contains(body, "proc") {
		t.Error("expected body to contain filter input 'proc'")
	}
}

// TestRenderContextSwitcherBody_ShowsNoMatchMessage verifies that when filter
// produces no results, the body contains a "No matching" message.
func TestRenderContextSwitcherBody_ShowsNoMatchMessage(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "xyz" // no match

	body := m.renderContextSwitcherBody()

	if !strings.Contains(body, "No matching") {
		t.Errorf("expected 'No matching' message, got body: %q", body)
	}
}

// TestContextSwitcher_TransitionFullClearsNavigationStack verifies that selecting
// a context via ModalContextSwitcher calls prepareStateTransition(TransitionFull),
// clearing any prior navigation stack entries.
func TestContextSwitcher_TransitionFullClearsNavigationStack(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.rootContexts = []string{"process-definition", "process-instance", "task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "process-instance"
	m.popup.cursor = -1
	// Simulate being drilled into a child view.
	m.navigationStack = []viewState{{viewMode: "process-definition"}}
	m.searchTerm = "invoice"

	m2, _ := sendKeyString(m, "enter")

	if len(m2.navigationStack) != 0 {
		t.Fatalf("expected navigationStack cleared by TransitionFull, got %d entries", len(m2.navigationStack))
	}
	if m2.searchTerm != "" {
		t.Fatalf("expected searchTerm cleared by TransitionFull, got %q", m2.searchTerm)
	}
}

// TestContextSwitcher_BreadcrumbUpdatedOnSelection verifies that selecting a context
// updates breadcrumb to a single-element slice with the selected resource name.
func TestContextSwitcher_BreadcrumbUpdatedOnSelection(t *testing.T) {
	m := newTestModel(t)
	m.splashActive = false
	m.rootContexts = []string{"process-definition", "task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "task"
	m.popup.cursor = -1
	m.breadcrumb = []string{"process-definition", "process-instance"}

	m2, _ := sendKeyString(m, "enter")

	if len(m2.breadcrumb) != 1 || m2.breadcrumb[0] != "task" {
		t.Fatalf("expected breadcrumb=[task], got %v", m2.breadcrumb)
	}
}

// TestContextSwitcher_TabCompletesToFirstMatch verifies that Tab with cursor=-1 and
// matching items completes popup.input to the first item in the filtered list.
func TestContextSwitcher_TabCompletesToFirstMatch(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition", "process-instance", "task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "proc"
	m.popup.cursor = -1

	m2, _ := sendKeyString(m, "tab")

	if m2.popup.input != "process-definition" {
		t.Fatalf("expected popup.input 'process-definition' after Tab, got %q", m2.popup.input)
	}
	if m2.popup.cursor != -1 {
		t.Fatalf("expected cursor -1 after Tab, got %d", m2.popup.cursor)
	}
	// Tab should not close the modal
	if m2.activeModal != ModalContextSwitcher {
		t.Fatalf("expected modal still open after Tab, got %v", m2.activeModal)
	}
}

// TestContextSwitcher_TabCompletesToCursorItem verifies that Tab with a valid cursor
// completes to the cursor's item rather than always the first item.
func TestContextSwitcher_TabCompletesToCursorItem(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition", "process-instance", "task"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "proc"
	m.popup.cursor = 1 // "process-instance"

	m2, _ := sendKeyString(m, "tab")

	if m2.popup.input != "process-instance" {
		t.Fatalf("expected popup.input 'process-instance' after Tab on cursor=1, got %q", m2.popup.input)
	}
}

// TestContextSwitcher_TabNoopOnEmptyInput verifies that Tab with empty input and
// no items is a no-op.
func TestContextSwitcher_TabNoopOnEmptyInput(t *testing.T) {
	m := newTestModel(t)
	m.rootContexts = []string{"process-definition"}
	m.activeModal = ModalContextSwitcher
	m.popup.input = "" // empty input → items returned, but Tab requires len(input) > 0
	m.popup.cursor = -1

	m2, _ := sendKeyString(m, "tab")

	if m2.popup.input != "" {
		t.Fatalf("expected popup.input unchanged (empty) after Tab with empty input, got %q", m2.popup.input)
	}
}
