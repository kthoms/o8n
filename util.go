package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
)

// truncateString truncates s to at most n visible characters (runes), safe for Unicode.
func truncateString(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

// friendlyError translates raw Go network errors into user-friendly messages.
func friendlyError(env string, err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "connection refused"):
		return fmt.Sprintf("Cannot connect to %s — is the engine running?", env)
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded"):
		return fmt.Sprintf("Request timed out — %s may be slow or unreachable", env)
	case strings.Contains(msg, "certificate") || strings.Contains(msg, "x509"):
		return fmt.Sprintf("TLS/certificate error — check HTTPS config for %s", env)
	case strings.Contains(msg, "no such host"):
		return fmt.Sprintf("Unknown host for %s — check the base URL in config", env)
	default:
		return msg
	}
}

// stripFocusIndicatorPrefix removes the drilldown indicator (▶ ) from the beginning of a string.
// Used to extract clean IDs/names from table rows that have the visual prefix for focus indication.
func stripFocusIndicatorPrefix(s string) string {
	if strings.HasPrefix(s, "▶ ") {
		return strings.TrimPrefix(s, "▶ ")
	}
	return s
}

func rowInstanceID(row table.Row) string {
	if len(row) > 0 {
		return stripFocusIndicatorPrefix(row[0])
	}
	return ""
}

// deriveFocusBackgroundColor returns a darker shade of the accent color for focus indicator background.
// Maps common hex colors to darker 256-color codes or returns a default dark color.
func (m *model) deriveFocusBackgroundColor(accentColor string) string {
	// Map of accent colors to dark background colors (256-color codes)
	colorMap := map[string]string{
		"#FFA500": "94",  // Orange → dark orange
		"#00A8E1": "23",  // Blue → dark blue
		"#00D7FF": "23",  // Cyan → dark cyan
		"#50C878": "22",  // Green → dark green
		"#FF6B6B": "52",  // Red → dark red
		"#FFD700": "136", // Gold → dark gold
	}

	if darkShade, ok := colorMap[accentColor]; ok {
		return darkShade
	}
	// Default to dark blue if color not in map
	return "23"
}

// validateConfigFiles verifies that critical config files exist and are not empty.
// Returns an error if config files appear corrupted or missing.
func validateConfigFiles() error {
	criticalFiles := map[string]int{
		"o8n-cfg.yaml": 700, // Config should have ~760 lines
		"o8n-env.yaml": 5,   // Env should have ~10-11 lines (absolute minimum)
	}

	for file, minLines := range criticalFiles {
		// Read file to count lines and verify it exists
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("critical file missing or unreadable: %s", file)
		}

		// Check if file appears empty or too small
		size := len(data)
		if size < 100 {
			return fmt.Errorf("critical file appears corrupted or empty: %s (%d bytes). Expected ~%d lines minimum. Restore with: git show HEAD~2:%s > %s",
				file, size, minLines, file, file)
		}

		lineCount := strings.Count(string(data), "\n")
		if lineCount < minLines {
			return fmt.Errorf("%s appears corrupted (%d lines, expected ~%d minimum). Restore with: git show HEAD~2:%s > %s",
				file, lineCount, minLines, file, file)
		}
	}

	return nil
}
