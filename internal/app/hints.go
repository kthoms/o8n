package app

import (
	"fmt"
	"sort"
)

// Hint represents a keyboard shortcut hint for display in the footer or a modal hint line.
// Priority follows the existing convention: lower integer = higher priority (1 = always shown).
// MinWidth specifies the minimum terminal width required to display this hint; 0 = always show.
type Hint struct {
	Key      string
	Label    string
	MinWidth int // terminal columns required; 0 = always show
	Priority int // 1 = highest priority; shown first when space is tight
}

func filterHints(hints []Hint, width int) []Hint {
	sorted := append([]Hint(nil), hints...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})

	filtered := make([]Hint, 0, len(sorted))
	for _, h := range sorted {
		if h.MinWidth == 0 || width >= h.MinWidth {
			filtered = append(filtered, h)
		}
	}
	return filtered
}

func currentViewHints(m model) []Hint {
	// Active modal → use the modal's HintLine (nil if not configured)
	if m.activeModal != ModalNone {
		if cfg, ok := modalRegistry[m.activeModal]; ok && len(cfg.HintLine) > 0 {
			return cfg.HintLine
		}
		return nil
	}
	// Active popup mode → per-popup hints
	switch m.popup.mode {
	case popupModeSearch:
		return searchPopupHints()
	case popupModeSkin:
		return skinPickerHints()
	}
	// Default: main table view hints
	return tableViewHints(m)
}

func tableViewHints(m model) []Hint {
	hints := []Hint{
		{Key: "?", Label: "help", MinWidth: 0, Priority: 1},
		{Key: ":", Label: "switch", MinWidth: 0, Priority: 2},
		{Key: "↑↓", Label: "nav", MinWidth: 0, Priority: 3},
		{Key: "/", Label: "filter", MinWidth: 0, Priority: 3},
		{Key: "PgDn/PgUp", Label: "page", MinWidth: 0, Priority: 3},
		{Key: "s", Label: "sort", MinWidth: 88, Priority: 5},
		{Key: "Ctrl+Shift+r", Label: "refresh", MinWidth: 90, Priority: 6},
		{Key: "Ctrl+t", Label: "skin", MinWidth: 90, Priority: 6},
		{Key: "Ctrl+e", Label: "env", MinWidth: 90, Priority: 6},
		{Key: "Ctrl+Space", Label: "actions", MinWidth: 100, Priority: 6},
		{Key: "J", Label: "json", MinWidth: 112, Priority: 6},
		{Key: "Ctrl+c", Label: "quit", MinWidth: 110, Priority: 8},
	}

	if def := m.findTableDef(m.currentRoot); def != nil && def.Drilldown != nil {
		hints = append(hints, Hint{Key: "Enter", Label: "drill", MinWidth: 0, Priority: 4})
	}
	if m.hasEditableColumns() {
		hints = append(hints, Hint{Key: "e", Label: "edit", MinWidth: 0, Priority: 4})
	}

	// Append hints from table actions (resource-specific hints)
	if def := m.findTableDef(m.currentRoot); def != nil {
		for _, action := range def.Actions {
			// Skip complex key combinations and special keys as hints (only single-char common keys)
			if len(action.Key) == 1 && action.Label != "" && action.Key != "?" && action.Key != ":" {
				hints = append(hints, Hint{Key: action.Key, Label: action.Label, MinWidth: 0, Priority: 4})
			}
		}
	}

	if len(m.navigationStack) > 0 {
		hints = append(hints, Hint{Key: "Esc", Label: "back", MinWidth: 0, Priority: 5})
	}
	if len(m.breadcrumb) > 1 {
		hints = append(hints, Hint{Key: fmt.Sprintf("1–%d", len(m.breadcrumb)-1), Label: "back", MinWidth: 0, Priority: 5})
	}

	return hints
}

func searchPopupHints() []Hint {
	return []Hint{
		{Key: "↑↓", Label: "select", MinWidth: 0, Priority: 1},
		{Key: "Enter", Label: "jump", MinWidth: 0, Priority: 1},
		{Key: "Esc", Label: "cancel", MinWidth: 0, Priority: 2},
	}
}

func skinPickerHints() []Hint {
	return []Hint{
		{Key: "↑↓", Label: "preview", MinWidth: 0, Priority: 1},
		{Key: "Enter", Label: "apply", MinWidth: 0, Priority: 1},
		{Key: "Esc", Label: "revert", MinWidth: 0, Priority: 2},
	}
}
