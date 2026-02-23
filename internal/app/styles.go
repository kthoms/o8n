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

// Package-level style constants used in View — cached here to avoid per-frame allocations.
var (
	// completionStyle renders the greyed-out autocomplete ghost text in the root popup.
	completionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	// flashBaseStyle is the fixed-width right-aligned base for the flash/remote indicator.
	flashBaseStyle = lipgloss.NewStyle().Width(3).Align(lipgloss.Right)
	// Feedback status styles
	errorFooterStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
	successFooterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50C878")).Bold(true)
	infoFooterStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00A8E1"))
	loadingFooterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	// Validation error style for edit modal
	validationErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
	// Page counter style (neutral, separate from flash color)
	pageCounterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)
