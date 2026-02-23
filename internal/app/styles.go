package app

import (
"time"

"github.com/charmbracelet/lipgloss"
)

const (
refreshInterval = 5 * time.Second
appVersion      = "0.1.0"
)

// EnvironmentStatus represents connection status to an environment
type EnvironmentStatus string

const (
StatusUnknown     EnvironmentStatus = "unknown"
StatusOperational EnvironmentStatus = "operational"
StatusUnreachable EnvironmentStatus = "unreachable"
)

// envStatusMsg reports health check result for an environment
type envStatusMsg struct {
env    string
status EnvironmentStatus
err    error
}

// spinnerFrames is the Braille dot spinner animation sequence.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// StyleSet holds all lipgloss styles for the application, built from a Skin.
// This is the ONLY place where lipgloss.Color values are constructed from skin roles.
// All view code uses m.styles.X — never literal color strings.
type StyleSet struct {
// Footer feedback
ErrorFooter   lipgloss.Style
SuccessFooter lipgloss.Style
InfoFooter    lipgloss.Style
LoadingFooter lipgloss.Style
PageCounter   lipgloss.Style

// Edit modal
ValidationError lipgloss.Style

// Popup / command palette
PopupCompletion lipgloss.Style // ghost-text autocomplete
PopupInput      lipgloss.Style // typed input text
PopupHint       lipgloss.Style // hint line
PopupBorder     lipgloss.Style // popup box border
PopupCursor     lipgloss.Style // highlighted cursor item
PopupItem       lipgloss.Style // non-selected list item

// Breadcrumb
CrumbNormal lipgloss.Style
CrumbActive lipgloss.Style

// Borders
BorderFg    lipgloss.Style
BorderFocus lipgloss.Style

// Table
TableHeader     lipgloss.Style
TableRow        lipgloss.Style
TableRowCursor  lipgloss.Style

// JSON detail view
JSONKey    lipgloss.Style
JSONValue  lipgloss.Style
JSONNumber lipgloss.Style
JSONBool   lipgloss.Style
JSONColon  lipgloss.Style

// Status indicators (env health)
StatusOperational lipgloss.Style
StatusUnreachable lipgloss.Style
StatusUnknown     lipgloss.Style

// Row status colors (process-instance / job / incident etc.)
RowRunning   lipgloss.Style
RowSuspended lipgloss.Style
RowFailed    lipgloss.Style
RowEnded     lipgloss.Style

// Edit modal buttons
BtnSave          lipgloss.Style
BtnSaveFocused   lipgloss.Style
BtnCancel        lipgloss.Style
BtnCancelFocused lipgloss.Style
BtnDisabled      lipgloss.Style

// Logo / splash
Logo lipgloss.Style
Info lipgloss.Style

// General
Accent  lipgloss.Style // accent-colored text (spinner, keys)
FgMuted lipgloss.Style // muted text
FlashBase lipgloss.Style

// Modal border base (no width — callers set width)
ModalBorderFocus lipgloss.Style
ModalBorderFg    lipgloss.Style
}

// col converts a skin color role to a lipgloss.Color.
// If the role resolves to "" we use lipgloss.NoColor (terminal default).
func col(skin *Skin, role string) lipgloss.Color {
v := skin.Color(role)
return lipgloss.Color(v)
}

// buildStyleSet constructs a StyleSet from the provided Skin.
// This is the single authoritative location for all color assignments.
func buildStyleSet(skin *Skin) StyleSet {
if skin == nil {
skin = &Skin{}
}
s := StyleSet{}

// ── Footer ──────────────────────────────────────────────────────────────
s.ErrorFooter = lipgloss.NewStyle().Foreground(col(skin, "danger")).Bold(true)
s.SuccessFooter = lipgloss.NewStyle().Foreground(col(skin, "success")).Bold(true)
s.InfoFooter = lipgloss.NewStyle().Foreground(col(skin, "info"))
s.LoadingFooter = lipgloss.NewStyle().Foreground(col(skin, "warning"))
s.PageCounter = lipgloss.NewStyle().Foreground(col(skin, "fgMuted"))

// ── Edit modal ───────────────────────────────────────────────────────────
s.ValidationError = lipgloss.NewStyle().Foreground(col(skin, "danger")).Bold(true)

// ── Popup / command palette ──────────────────────────────────────────────
s.PopupCompletion = lipgloss.NewStyle().Foreground(col(skin, "fgMuted"))
s.PopupInput = lipgloss.NewStyle().Foreground(col(skin, "accent"))
s.PopupHint = lipgloss.NewStyle().Foreground(col(skin, "fgMuted"))
s.PopupBorder = lipgloss.NewStyle().Foreground(col(skin, "borderFocus"))
s.PopupCursor = lipgloss.NewStyle().Foreground(col(skin, "accent")).Bold(true)
s.PopupItem = lipgloss.NewStyle().Foreground(col(skin, "fg"))

// ── Breadcrumb ───────────────────────────────────────────────────────────
s.CrumbNormal = lipgloss.NewStyle().
Foreground(col(skin, "crumbFg")).
Background(col(skin, "crumbBg"))
s.CrumbActive = lipgloss.NewStyle().
Foreground(col(skin, "crumbActiveFg")).
Background(col(skin, "crumbActiveBg")).
Bold(true)

// ── Borders ──────────────────────────────────────────────────────────────
s.BorderFg = lipgloss.NewStyle().Foreground(col(skin, "borderFg"))
s.BorderFocus = lipgloss.NewStyle().Foreground(col(skin, "borderFocus"))

// ── Table ────────────────────────────────────────────────────────────────
s.TableHeader = lipgloss.NewStyle().
Foreground(col(skin, "fg")).
Background(col(skin, "surface")).
Bold(true)
s.TableRow = lipgloss.NewStyle().Foreground(col(skin, "fg"))
s.TableRowCursor = lipgloss.NewStyle().
Foreground(col(skin, "fg")).
Background(col(skin, "surfaceAlt")).
Bold(true)

// ── JSON detail ──────────────────────────────────────────────────────────
s.JSONKey = lipgloss.NewStyle().Foreground(col(skin, "jsonKey"))
s.JSONValue = lipgloss.NewStyle().Foreground(col(skin, "jsonValue"))
s.JSONNumber = lipgloss.NewStyle().Foreground(col(skin, "jsonNumber"))
s.JSONBool = lipgloss.NewStyle().Foreground(col(skin, "jsonBool"))
s.JSONColon = lipgloss.NewStyle().Foreground(col(skin, "fgMuted"))

// ── Status indicators ────────────────────────────────────────────────────
s.StatusOperational = lipgloss.NewStyle().Foreground(col(skin, "success"))
s.StatusUnreachable = lipgloss.NewStyle().Foreground(col(skin, "danger"))
s.StatusUnknown = lipgloss.NewStyle().Foreground(col(skin, "warning"))

// ── Row status colors (table rows by resource state) ─────────────────────
s.RowRunning = lipgloss.NewStyle().Foreground(col(skin, "success"))
s.RowSuspended = lipgloss.NewStyle().Foreground(col(skin, "warning"))
s.RowFailed = lipgloss.NewStyle().Foreground(col(skin, "danger"))
s.RowEnded = lipgloss.NewStyle().Foreground(col(skin, "fgMuted"))

// ── Edit modal buttons ────────────────────────────────────────────────────
s.BtnSave = lipgloss.NewStyle().
Background(col(skin, "btnPrimaryBg")).
Foreground(col(skin, "btnPrimaryFg")).
Padding(0, 1).Bold(true)
s.BtnSaveFocused = s.BtnSave.Copy().
Border(lipgloss.NormalBorder()).
BorderForeground(col(skin, "borderFocus"))
s.BtnCancel = lipgloss.NewStyle().
Background(col(skin, "btnSecondaryBg")).
Foreground(col(skin, "btnSecondaryFg")).
Padding(0, 1)
s.BtnCancelFocused = s.BtnCancel.Copy().
Border(lipgloss.NormalBorder()).
BorderForeground(col(skin, "borderFocus"))
s.BtnDisabled = lipgloss.NewStyle().
Background(col(skin, "surface")).
Foreground(col(skin, "fgMuted")).
Padding(0, 1)

// ── Logo / splash ─────────────────────────────────────────────────────────
s.Logo = lipgloss.NewStyle().Foreground(col(skin, "accentAlt")).Bold(true).Align(lipgloss.Center)
s.Info = lipgloss.NewStyle().Foreground(col(skin, "accentAlt")).Align(lipgloss.Center)

// ── General ──────────────────────────────────────────────────────────────
s.Accent = lipgloss.NewStyle().Foreground(col(skin, "accent"))
s.FgMuted = lipgloss.NewStyle().Foreground(col(skin, "fgMuted"))
s.FlashBase = lipgloss.NewStyle().Width(3).Align(lipgloss.Right)

// ── Modal borders ─────────────────────────────────────────────────────────
s.ModalBorderFocus = lipgloss.NewStyle().
Border(lipgloss.RoundedBorder()).
BorderForeground(col(skin, "borderFocus"))
s.ModalBorderFg = lipgloss.NewStyle().
Border(lipgloss.RoundedBorder()).
BorderForeground(col(skin, "borderFg"))

return s
}
