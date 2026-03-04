package app

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SizeHint controls the size class of a modal overlay.
type SizeHint int

const (
	// OverlayCenter is a compact dialog (~50% terminal width, auto-height), centered.
	OverlayCenter SizeHint = iota
	// OverlayLarge is a rich-content modal (~80% terminal width × ~80% terminal height),
	// centered with background content visible behind it. Requires HintLine.
	OverlayLarge
	// FullScreen occupies the entire terminal viewport. Requires HintLine.
	FullScreen
)

// ModalConfig describes all rendering parameters for a modal type.
// Register configs with registerModal; renderModal dispatches via modalRegistry.
//
// Body renderer conventions:
//   - OverlayCenter: returns a complete styled modal box (border+padding already applied)
//   - OverlayLarge: returns raw inner content (no border); factory applies border + HintLine
//   - FullScreen: returns a complete custom-layout string; factory returns it as-is
type ModalConfig struct {
	SizeHint     SizeHint
	Title        string
	BodyRenderer func(m model) string
	ConfirmLabel string
	CancelLabel  string
	HintLine     []Hint // rendered at modal bottom; required for OverlayLarge and FullScreen
}

// modalRegistry maps each ModalType to its rendering ModalConfig.
var modalRegistry = map[ModalType]ModalConfig{}

// registerModal registers a ModalConfig for a given ModalType.
func registerModal(t ModalType, cfg ModalConfig) {
	modalRegistry[t] = cfg
}

func init() {
	registerModal(ModalConfirmDelete, ModalConfig{
		SizeHint: OverlayCenter,
		BodyRenderer: func(m model) string {
			return m.renderConfirmDeleteModal(m.lastWidth, m.lastHeight)
		},
		ConfirmLabel: "Delete",
		CancelLabel:  "Cancel",
	})

	registerModal(ModalConfirmQuit, ModalConfig{
		SizeHint: OverlayCenter,
		BodyRenderer: func(m model) string {
			return m.renderConfirmQuitModal(m.lastWidth, m.lastHeight)
		},
		ConfirmLabel: "Quit",
		CancelLabel:  "Cancel",
	})

	registerModal(ModalSort, ModalConfig{
		SizeHint: OverlayCenter,
		BodyRenderer: func(m model) string {
			return m.renderSortPopup(m.lastWidth, m.lastHeight)
		},
	})

	registerModal(ModalEnvironment, ModalConfig{
		SizeHint: OverlayCenter,
		BodyRenderer: func(m model) string {
			return m.renderEnvPopup(m.lastWidth, m.lastHeight)
		},
	})

	registerModal(ModalEdit, ModalConfig{
		SizeHint: OverlayCenter,
		BodyRenderer: func(m model) string {
			return m.renderEditModal(m.lastWidth, m.lastHeight)
		},
		ConfirmLabel: "Save",
		CancelLabel:  "Cancel",
	})

	registerModal(ModalHelp, ModalConfig{
		SizeHint: OverlayLarge,
		BodyRenderer: func(m model) string {
			return m.modalHelpBody()
		},
		HintLine: []Hint{
			{Key: "↑↓", Label: "scroll", Priority: 1},
			{Key: "q/Esc", Label: "close", Priority: 1},
		},
	})

	registerModal(ModalDetailView, ModalConfig{
		SizeHint: OverlayLarge,
		BodyRenderer: func(m model) string {
			return m.modalDetailViewBody()
		},
		HintLine: []Hint{
			{Key: "↑↓", Label: "scroll", Priority: 1},
			{Key: "q/Esc", Label: "close", Priority: 1},
		},
	})

	registerModal(ModalTaskComplete, ModalConfig{
		SizeHint: FullScreen,
		BodyRenderer: func(m model) string {
			return m.renderTaskCompleteModal(m.lastWidth, m.lastHeight)
		},
		HintLine: []Hint{
			{Key: "Tab", Label: "switch", Priority: 1},
			{Key: "Enter", Label: "confirm", Priority: 1},
			{Key: "Esc", Label: "cancel", Priority: 2},
		},
	})
}

// renderModal is the factory entry point for all modal rendering.
// It looks up the ModalConfig for m.activeModal, calls the body renderer,
// applies uniform sizing per size class, and returns the overlay string.
//
// OverlayCenter: body renderer returns a complete styled modal; returned as-is.
// OverlayLarge:  body renderer returns raw text; factory applies RoundedBorder at ~80%
//
//	terminal width and renders HintLine below the content.
//
// FullScreen:    body renderer returns complete custom layout; returned as-is.
func renderModal(m model, cfg ModalConfig) string {
	body := cfg.BodyRenderer(m)
	if body == "" {
		return ""
	}

	switch cfg.SizeHint {
	case FullScreen:
		return body

	case OverlayLarge:
		targetW := int(float64(m.lastWidth) * 0.80)
		targetH := int(float64(m.lastHeight) * 0.80)
		if targetW < 60 {
			targetW = 60
		}
		if targetH < 8 {
			targetH = 8
		}
		if m.lastWidth > 4 && targetW > m.lastWidth-4 {
			targetW = m.lastWidth - 4
		}
		if m.lastHeight > 4 && targetH > m.lastHeight-4 {
			targetH = m.lastHeight - 4
		}
		content := body
		hintStr := renderModalHintLine(m, cfg.HintLine)
		if hintStr != "" {
			content = body + "\n\n" + hintStr
		}
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(col(m.skin, "borderFocus")).
			Padding(1, 2).
			Width(targetW).
			Height(targetH).
			Render(content)

	default: // OverlayCenter
		targetW := int(float64(m.lastWidth) * 0.50)
		if targetW < 44 {
			targetW = 44
		}
		if m.lastWidth > 4 && targetW > m.lastWidth-4 {
			targetW = m.lastWidth - 4
		}
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(col(m.skin, "borderFocus")).
			Padding(1, 2).
			Width(targetW).
			Render(body)
	}
}

// renderModalHintLine formats a []Hint slice into a single-line display string
// using the FgMuted style. Used by the factory for OverlayLarge modals.
func renderModalHintLine(m model, hints []Hint) string {
	if len(hints) == 0 {
		return ""
	}
	visible := make([]Hint, 0, len(hints))
	for _, h := range hints {
		if h.MinWidth > 0 && m.lastWidth < h.MinWidth {
			continue
		}
		visible = append(visible, h)
	}
	if len(visible) == 0 {
		return ""
	}
	sort.SliceStable(visible, func(i, j int) bool {
		return visible[i].Priority < visible[j].Priority
	})
	parts := make([]string, 0, len(visible))
	for _, h := range visible {
		parts = append(parts, h.Key+"  "+h.Label)
	}
	return m.styles.FgMuted.Render(strings.Join(parts, "   "))
}

// overlayLarge places fg centered over bg for an OverlayLarge modal.
// The centering logic is identical to overlayCenter; the size distinction
// is enforced by the content width set in renderModal.
func overlayLarge(bg, fg string) string {
	return overlayCenter(bg, fg)
}

// overlayFullscreen replaces the background with fg placed at full terminal dimensions.
func overlayFullscreen(bg, fg string, termW, termH int) string {
	_ = bg
	return lipgloss.Place(termW, termH, lipgloss.Left, lipgloss.Top, fg)
}
