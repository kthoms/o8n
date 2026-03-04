package app

// Hint represents a keyboard shortcut hint for display in the footer or a modal hint line.
// Priority follows the KeyHint convention: lower integer = higher priority (1 = always shown).
// MinWidth specifies the minimum terminal width required to display this hint; 0 = always show.
type Hint struct {
	Key      string
	Label    string
	MinWidth int // terminal columns required; 0 = always show
	Priority int // 1 = highest priority; shown first when space is tight
}
