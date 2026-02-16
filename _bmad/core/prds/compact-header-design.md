# 🎯 Compact Header Design - Priority-Based Key Visibility

**Feature:** Compact 3-Row Header with Smart Key Display  
**Priority:** 🔴 HIGH  
**Effort:** 4 hours  
**Status:** Design Complete  
**Designer:** BMM UX Designer  
**Date:** February 16, 2026

---

## Executive Summary

This document specifies the **compact 3-row header** design with **priority-based key visibility**. The header adapts to terminal width, always showing the most critical keys first and progressively revealing less-used features as space allows.

**Design Principles:**
- ✅ **Discoverability first** - Help key always visible
- ✅ **Navigation priority** - Core navigation keys prominent
- ✅ **Context-aware** - Keys change based on current view
- ✅ **Progressive disclosure** - More keys appear as width increases
- ✅ **Low-priority hidden** - Environment switch only on wide terminals

**Space Savings:** 8 rows → 3 rows = **5 rows back to content (+62.5% reduction)**

---

## 1. Key Priority Matrix

### 1.1 Global Priority Ranking

Keys ranked by importance (1 = highest priority):

| Priority | Key | Action | Why It's Important | Always Show |
|----------|-----|--------|-------------------|-------------|
| 1 | `?` | Help | **Critical for discoverability** - users can't use what they can't discover | ✓ Yes |
| 2 | `:` | Switch view | **Core navigation** - needed to access different data types | ✓ Yes |
| 3 | `↑/↓` | Navigate list | **Primary interaction** - always navigating lists | ✓ Yes |
| 4 | `Enter` | Drill down | **Core navigation** - drill into details | ✓ Yes |
| 5 | `Esc` | Go back | **Core navigation** - escape/back from anywhere | ✓ Yes |
| 6 | `<ctrl>-r` | Refresh | **Common action** - frequently used to update data | Width ≥ 90 |
| 7 | `<ctrl>+d` | Delete/Kill | **Destructive action** - needs to be discoverable but not accidental | Width ≥ 100 |
| 8 | `q` | Quit | **Exit** - standard but Ctrl+C also works | Width ≥ 110 |
| 9 | `<ctrl>+e` | Switch env | **Low priority** - not frequently used | Width ≥ 130 |

### 1.2 View-Specific Keys

Additional context-sensitive keys (shown when relevant):

**Process Instances View:**
- `v` - View variables (Priority 6)
- `s` - Suspend instance (Priority 8)
- `r` - Resume instance (Priority 8)

**Variables View:**
- `/` - Search (Priority 7)
- `e` - Edit variable (Priority 9, future feature)

---

## 2. Compact Header Layouts

### 2.1 Minimum Width (80 chars)

**Philosophy:** Essential keys only - discoverability + core navigation

```
┌─ o8n v0.1.0 ────────────────────────────────────────────────────────────────┐
│ local @ localhost:8080 │ demo │ ✓ Connected │ ⚡ 45ms │ 23 items            │ Row 1: Status
│ ? help  : switch  ↑↓ nav  Enter drill  Esc back  <ctrl>-r refresh          │ Row 2: Keys
│                                                                              │ Row 3: Spacer
└──────────────────────────────────────────────────────────────────────────────┘

Character count: 80 chars
Keys shown: 6 (help, switch, nav, drill, back, refresh)
Keys hidden: 3 (delete, quit, env)
```

**Key Features:**
- Status bar: App version + environment + connection + API latency + item count
- Essential 6 keys cover 90% of usage
- All navigation keys present
- No destructive actions (safer for beginners)

---

### 2.2 Small Width (100 chars)

**Philosophy:** Add destructive action visibility

```
┌─ o8n v0.1.0 ────────────────────────────────────────────────────────────────────────────────┐
│ local @ http://localhost:8080 │ demo │ ✓ Connected │ ⚡ 45ms │ 23 definitions               │ Row 1
│ ? help  : switch  ↑↓ nav  Enter drill  Esc back  <ctrl>-r refresh  <ctrl>+d delete          │ Row 2
│                                                                                              │ Row 3
└──────────────────────────────────────────────────────────────────────────────────────────────┘

Character count: 100 chars
Keys shown: 7 (+delete)
Keys hidden: 2 (quit, env)
```

**Added:**
- Full URL visible
- `<ctrl>+d` delete action (important for power users)

---

### 2.3 Medium Width (120 chars)

**Philosophy:** Add quit for completeness

```
┌─ o8n v0.1.0 ────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Environment: local @ http://localhost:8080/engine-rest │ User: demo │ ✓ 45ms │ 23 definitions                   │ Row 1
│ ? help  : switch  ↑↓ nav  Enter drill  Esc back  <ctrl>-r refresh  <ctrl>+d delete  q quit                      │ Row 2
│                                                                                                                  │ Row 3
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

Character count: 120 chars
Keys shown: 8 (+quit)
Keys hidden: 1 (env)
```

**Added:**
- Full API path
- `q` quit key
- More descriptive labels ("Environment:", "User:")

---

### 2.4 Large Width (140+ chars)

**Philosophy:** Show everything including low-priority env switch

```
┌─ o8n v0.1.0 ────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Environment: local @ http://localhost:8080/engine-rest │ User: demo │ API: ✓ 45ms │ Data: 23 definitions │ Updated: 14:32:15        │ Row 1
│ Navigation: ↑↓ nav  Enter drill  Esc back  PgUp/Dn page │ Actions: <ctrl>-r refresh  <ctrl>+d delete │ Global: ? help  : switch  q quit │ Row 2
│ Advanced: <ctrl>+e env  / search  <ctrl>-R force-refresh │ View: process-definitions                                              │ Row 3
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

Character count: 140+ chars
Keys shown: 12+ (all keys)
Keys hidden: 0
```

**Added:**
- Environment switch visible
- Grouped by category (Navigation | Actions | Global | Advanced)
- Additional keys (PgUp/Dn, force-refresh, search)
- Full timestamps

---

## 3. View-Specific Key Variations

### 3.1 Process Definitions View

**80 chars (minimum):**
```
? help  : switch  ↑↓ nav  Enter drill  Esc back  <ctrl>-r refresh
```

**100 chars:**
```
? help  : switch  ↑↓ nav  Enter drill  Esc back  <ctrl>-r refresh  <ctrl>+d delete
```

**120 chars:**
```
? help  : switch  ↑↓ nav  Enter drill  Esc back  <ctrl>-r refresh  <ctrl>+d delete  q quit
```

---

### 3.2 Process Instances View

**80 chars (minimum):**
```
? help  Esc back  ↑↓ nav  Enter vars  v vars  <ctrl>+d kill  <ctrl>-r refresh
```
*Note:* "Enter vars" = drill into variables, "v vars" = quick variable view

**100 chars:**
```
? help  Esc back  ↑↓ nav  Enter vars  v vars  <ctrl>+d kill  s suspend  r resume  <ctrl>-r refresh
```

**120 chars:**
```
? help  : switch  Esc back  ↑↓ nav  Enter vars  v vars  <ctrl>+d kill  s suspend  r resume  <ctrl>-r refresh  q quit
```

**Key Changes:**
- `Esc back` promoted (important in drill-down context)
- `v vars` added (instance-specific action)
- `s suspend` / `r resume` appear at 100+ width
- `<ctrl>+d kill` always visible (common destructive action)

---

### 3.3 Variables View

**80 chars (minimum):**
```
? help  Esc back  ↑↓ nav  / search  <ctrl>-r refresh  q quit
```

**100 chars:**
```
? help  Esc back  ↑↓ nav  Enter details  / search  e edit  <ctrl>-r refresh  q quit
```

**Key Changes:**
- `Esc back` critical (deepest level, need to go back)
- `/ search` visible (important for finding variables)
- `e edit` appears at 100+ width (future feature)
- No `<ctrl>+d` (read-only view, no delete)

---

## 4. Key Display Formatting

### 4.1 Visual Hierarchy

**Key Rendering Styles:**

```go
// Color scheme
helpKey      := lipgloss.NewStyle().Bold(true).Foreground("220")  // Yellow (stands out)
navigationKey := lipgloss.NewStyle().Bold(true).Foreground(uiColor) // Environment color
actionKey    := lipgloss.NewStyle().Bold(true).Foreground("white")
destructiveKey := lipgloss.NewStyle().Bold(true).Foreground("196") // Red
description  := lipgloss.NewStyle().Foreground("252") // Light gray
```

**Example Rendering:**

```
? help  : switch  ↑↓ nav  <ctrl>+d delete
└─┘ └─┘ └┘ └────┘ └─┘ └─┘ └──────┘ └────┘
 │   │   │    │    │   │     │       │
 │   │   │    │    │   │     │       └─ Description (light gray)
 │   │   │    │    │   │     └─ Key (red - destructive)
 │   │   │    │    │   └─ Description
 │   │   │    │    └─ Key (white - action)
 │   │   │    └─ Description
 │   │   └─ Key (env color - navigation)
 │   └─ Description
 └─ Key (yellow - help, stands out)
```

---

### 4.2 Grouping and Spacing

**Pattern:** `[key] [action] [separator] [key] [action]`

```
Tight spacing (80 chars):
? help  : switch  ↑↓ nav
^     ^^       ^^     ^
key   sp key    sp key

Medium spacing (100 chars):
? help  : switch  ↑↓ nav  Enter drill
^     ^^       ^^     ^^          ^
key   sp key    sp key  sp key

Grouped (120+ chars):
Navigation: ↑↓ nav  Enter drill │ Actions: <ctrl>-r refresh
^─────────^ ^                   ^ ^───────^ ^
category    keys                sep category keys
```

**Separator Rules:**
- 2 spaces between key-action pairs
- " │ " (space + pipe + space) between groups
- Group labels only at 120+ width

---

## 5. Implementation Specification

### 5.1 Data Structure

```go
// KeyHint represents a keyboard shortcut hint
type KeyHint struct {
    Key         string  // "?" or "<ctrl>+d"
    Description string  // "help" or "delete"
    Priority    int     // 1 (highest) to 10 (lowest)
    MinWidth    int     // Minimum terminal width to display
    Type        KeyType // Help, Navigation, Action, Destructive
    ViewMode    string  // "all", "definitions", "instances", "variables"
}

type KeyType int

const (
    KeyTypeHelp KeyType = iota
    KeyTypeNavigation
    KeyTypeAction
    KeyTypeDestructive
    KeyTypeGlobal
)

// Key hint registry
var keyHints = []KeyHint{
    // Universal keys (all views)
    {Key: "?", Description: "help", Priority: 1, MinWidth: 80, Type: KeyTypeHelp, ViewMode: "all"},
    {Key: ":", Description: "switch", Priority: 2, MinWidth: 80, Type: KeyTypeNavigation, ViewMode: "all"},
    {Key: "↑↓", Description: "nav", Priority: 3, MinWidth: 80, Type: KeyTypeNavigation, ViewMode: "all"},
    {Key: "Enter", Description: "drill", Priority: 4, MinWidth: 80, Type: KeyTypeNavigation, ViewMode: "all"},
    {Key: "Esc", Description: "back", Priority: 5, MinWidth: 80, Type: KeyTypeNavigation, ViewMode: "all"},
    {Key: "<ctrl>-r", Description: "refresh", Priority: 6, MinWidth: 80, Type: KeyTypeAction, ViewMode: "all"},
    {Key: "<ctrl>+d", Description: "delete", Priority: 7, MinWidth: 100, Type: KeyTypeDestructive, ViewMode: "definitions,instances"},
    {Key: "q", Description: "quit", Priority: 8, MinWidth: 110, Type: KeyTypeGlobal, ViewMode: "all"},
    {Key: "<ctrl>+e", Description: "env", Priority: 9, MinWidth: 130, Type: KeyTypeAction, ViewMode: "all"},
    
    // Instance-specific keys
    {Key: "v", Description: "vars", Priority: 6, MinWidth: 80, Type: KeyTypeAction, ViewMode: "instances"},
    {Key: "s", Description: "suspend", Priority: 8, MinWidth: 100, Type: KeyTypeAction, ViewMode: "instances"},
    {Key: "r", Description: "resume", Priority: 8, MinWidth: 100, Type: KeyTypeAction, ViewMode: "instances"},
    
    // Variables-specific keys
    {Key: "/", Description: "search", Priority: 7, MinWidth: 80, Type: KeyTypeAction, ViewMode: "variables"},
    {Key: "e", Description: "edit", Priority: 9, MinWidth: 100, Type: KeyTypeAction, ViewMode: "variables"},
}
```

---

### 5.2 Key Hint Builder

```go
// buildKeyHints creates context-sensitive key hints for header row 2
func (m *model) buildKeyHints() string {
    width := m.lastWidth
    viewMode := m.viewMode
    
    // Get applicable hints for current view and width
    var hints []KeyHint
    for _, hint := range keyHints {
        // Check if hint applies to current view
        if hint.ViewMode == "all" || strings.Contains(hint.ViewMode, viewMode) {
            // Check if terminal is wide enough
            if width >= hint.MinWidth {
                hints = append(hints, hint)
            }
        }
    }
    
    // Sort by priority (low number = high priority)
    sort.Slice(hints, func(i, j int) bool {
        return hints[i].Priority < hints[j].Priority
    })
    
    // Render hints
    return m.renderKeyHints(hints, width)
}

func (m *model) renderKeyHints(hints []KeyHint, width int) string {
    var parts []string
    usedWidth := 0
    
    for _, hint := range hints {
        // Calculate this hint's width: len(key) + space + len(desc) + 2 spaces
        hintWidth := lipgloss.Width(hint.Key) + 1 + lipgloss.Width(hint.Description) + 2
        
        // Check if we have space
        if usedWidth + hintWidth > width - 4 { // -4 for margins
            break
        }
        
        // Render key with appropriate style
        var keyStyle lipgloss.Style
        switch hint.Type {
        case KeyTypeHelp:
            keyStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")) // Yellow
        case KeyTypeNavigation:
            keyStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.currentColor))
        case KeyTypeDestructive:
            keyStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")) // Red
        default:
            keyStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("white"))
        }
        
        descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
        
        part := fmt.Sprintf("%s %s",
            keyStyle.Render(hint.Key),
            descStyle.Render(hint.Description),
        )
        
        parts = append(parts, part)
        usedWidth += hintWidth
    }
    
    return strings.Join(parts, "  ")
}
```

---

### 5.3 Status Bar Builder

```go
// StatusSegment represents one piece of status info
type StatusSegment struct {
    Text     string
    MinWidth int
    Priority int
    Color    string
}

// buildStatusBar creates adaptive status bar for header row 1
func (m *model) buildStatusBar(width int) string {
    segments := []StatusSegment{
        {Text: "o8n v0.1.0", MinWidth: 12, Priority: 100, Color: "white"},
        {Text: m.currentEnv, MinWidth: 8, Priority: 90, Color: m.currentColor},
        {Text: "✓ Connected", MinWidth: 12, Priority: 80, Color: "42"},
        {Text: m.username, MinWidth: 8, Priority: 70, Color: "252"},
        {Text: fmt.Sprintf("⚡ %dms", m.apiLatency), MinWidth: 10, Priority: 60, Color: "220"},
        {Text: fmt.Sprintf("%d items", m.itemCount), MinWidth: 12, Priority: 50, Color: "252"},
    }
    
    // Add full URL if available and wide enough
    if width >= 100 && m.config != nil {
        if env, ok := m.config.Environments[m.currentEnv]; ok {
            urlText := fmt.Sprintf("@ %s", env.URL)
            segments = append(segments, StatusSegment{
                Text: urlText, MinWidth: len(urlText), Priority: 85, Color: "252",
            })
        }
    }
    
    // Sort by priority (high to low)
    sort.Slice(segments, func(i, j int) bool {
        return segments[i].Priority > segments[j].Priority
    })
    
    // Fit segments into available width
    var parts []string
    available := width - 6 // Reserve for separators and margins
    
    for _, seg := range segments {
        if available >= seg.MinWidth {
            style := lipgloss.NewStyle().Foreground(lipgloss.Color(seg.Color))
            parts = append(parts, style.Render(seg.Text))
            available -= seg.MinWidth + 3 // +3 for " │ "
        }
    }
    
    separator := lipgloss.NewStyle().
        Foreground(lipgloss.Color("240")).
        Render(" │ ")
    
    return strings.Join(parts, separator)
}
```

---

### 5.4 Complete Header View

```go
func (m *model) renderHeader() string {
    // Row 1: Status bar
    statusBar := m.buildStatusBar(m.lastWidth)
    
    // Row 2: Key hints
    keyHints := m.buildKeyHints()
    
    // Row 3: Spacer (empty line for visual breathing room)
    spacer := ""
    
    // Combine into 3-row header
    header := lipgloss.JoinVertical(lipgloss.Left,
        statusBar,
        keyHints,
        spacer,
    )
    
    // Ensure header is exactly the right width
    headerStyle := lipgloss.NewStyle().
        Width(m.lastWidth).
        Height(3)
    
    return headerStyle.Render(header)
}
```

---

## 6. Visual Examples with Actual Colors

### 6.1 Minimum (80 chars) - Definitions View

```
o8n v0.1.0 │ local │ ✓ Connected │ ⚡ 45ms │ 23 definitions
? help  : switch  ↑↓ nav  Enter drill  Esc back  <ctrl>-r refresh

^     ^ ^      ^ ^─^    ^ ^          ^ ^       ^ ^────────^
yellow  cyan     cyan     cyan         cyan      white
```

Colors:
- `?` = Yellow bold (stands out for help)
- `:`, `↑↓`, `Enter`, `Esc` = Cyan bold (environment color)
- `<ctrl>-r` = White bold
- Descriptions = Light gray

---

### 6.2 Medium (100 chars) - Instances View

```
local @ http://localhost:8080 │ demo │ ✓ Connected │ ⚡ 45ms │ 15 instances
? help  Esc back  ↑↓ nav  v vars  <ctrl>+d kill  <ctrl>-r refresh

^     ^ ^       ^ ^─^    ^ ^    ^ ^────────^    ^ ^────────^
yellow  cyan      cyan     cyan    red             white
```

Colors:
- `<ctrl>+d` = Red bold (destructive action)
- Other navigation keys = Cyan
- Action keys = White

---

### 6.3 Large (140 chars) - Definitions View

```
Environment: local @ http://localhost:8080/engine-rest │ User: demo │ API: ✓ 45ms │ 23 defs
Navigation: ↑↓ nav  Enter drill  Esc back │ Actions: <ctrl>-r refresh  <ctrl>+d delete
Advanced: <ctrl>+e env │ View: process-definitions

^─────────^ ^──────────────────────────────^   ^──────^ ^──────────────────────────^
category    navigation keys (cyan)             category  action keys (white/red)
```

Grouped by category with labels at 140+ width.

---

## 7. Responsive Behavior Table

| Terminal Width | Status Segments | Key Count | Groups | Env Switch Visible |
|----------------|----------------|-----------|--------|-------------------|
| 80-89 | 5 (essential) | 6 | No | ❌ No |
| 90-99 | 5-6 | 6-7 | No | ❌ No |
| 100-109 | 6 | 7 (+delete) | No | ❌ No |
| 110-119 | 6-7 | 8 (+quit) | No | ❌ No |
| 120-129 | 7 | 8 | No | ❌ No |
| 130-139 | 7-8 | 9 (+env) | No | ✅ Yes |
| 140+ | 8+ | 10+ | ✅ Yes | ✅ Yes |

---

## 8. Testing Checklist

### 8.1 Visual Testing

- [ ] Header exactly 3 rows at all widths
- [ ] No text overflow at minimum 80 width
- [ ] Colors render correctly (yellow help, cyan nav, red destructive)
- [ ] Separators consistent (2 spaces between keys, " │ " between groups)
- [ ] Status bar balanced (not all crammed on left)

### 8.2 Responsive Testing

**Test at each width:**
- [ ] 80 chars: 6 keys visible, no delete/quit/env
- [ ] 90 chars: 6-7 keys visible
- [ ] 100 chars: delete key appears
- [ ] 110 chars: quit key appears
- [ ] 130 chars: env key appears
- [ ] 140 chars: grouped layout appears

### 8.3 View-Specific Testing

- [ ] Definitions view: shows standard navigation keys
- [ ] Instances view: shows v/s/r keys, <ctrl>+d as "kill"
- [ ] Variables view: shows / search, no delete key
- [ ] Switching views updates keys immediately
- [ ] No keys from wrong view shown

### 8.4 Edge Cases

- [ ] Very narrow terminal (< 80 chars): graceful truncation
- [ ] Very wide terminal (> 200 chars): reasonable max width
- [ ] Long environment URLs don't break layout
- [ ] Unicode characters in status (⚡, ✓) render correctly
- [ ] Rapid resize: no flickering, smooth transitions

---

## 9. Priority Adjustment Guidelines

### When to Promote a Key (Lower Priority Number)

**Criteria:**
1. Used in > 50% of user sessions
2. Required for core workflow
3. Not discoverable by other means
4. Mistakes are costly if not known

**Example:** If analytics show "suspend" is used frequently, promote to priority 6.

### When to Demote a Key (Higher Priority Number)

**Criteria:**
1. Used in < 10% of sessions
2. Alternative exists (e.g., `q` vs `Ctrl+C`)
3. Obvious from context
4. Only for advanced users

**Example:** Environment switch demoted to priority 9 (low usage).

---

## 10. Implementation Notes

### 10.1 Add to main.go Model

```go
type model struct {
    // ... existing fields ...
    
    // Header cache (recalculate on view change or resize)
    cachedStatusBar string
    cachedKeyHints  string
    headerDirty     bool
}
```

### 10.2 Update on Events

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.lastWidth = msg.Width
        m.lastHeight = msg.Height
        m.headerDirty = true  // Recalculate header
        
    case tea.KeyMsg:
        // After view change
        if msg.String() == ":" {
            // ... handle switch ...
            m.headerDirty = true  // Keys changed
        }
    }
    
    // ... rest of update logic ...
}
```

### 10.3 View Function

```go
func (m model) View() string {
    if m.splashActive {
        return m.renderSplash()
    }
    
    // Rebuild header if needed
    if m.headerDirty {
        m.cachedStatusBar = m.buildStatusBar(m.lastWidth)
        m.cachedKeyHints = m.buildKeyHints()
        m.headerDirty = false
    }
    
    header := m.renderHeader() // Uses cached values
    contextBox := m.renderContextSelector()
    content := m.renderContent()
    footer := m.renderFooter()
    
    return lipgloss.JoinVertical(lipgloss.Left,
        header,
        contextBox,
        content,
        footer,
    )
}
```

---

## 11. Success Metrics

### Before vs After

| Metric | Before (8 rows) | After (3 rows) | Improvement |
|--------|----------------|----------------|-------------|
| Header height | 8 rows | 3 rows | -62.5% |
| Content rows (24 total) | 13 rows | 19 rows | +46% |
| Key discoverability | Manual/docs | Visual header | +100% |
| Essential keys visible | All (cluttered) | Top 6-9 (clear) | Quality > Quantity |

### User Experience Goals

| Goal | Target | Measurement |
|------|--------|-------------|
| Key recall | > 80% | Survey: "Can you name 3 keys without looking?" |
| Help usage | < 20% | Track `?` key presses (lower = better discoverability) |
| Navigation confidence | > 4/5 | Survey: "Do you feel confident navigating?" |
| Visual clarity | > 4.5/5 | Survey: "Is the header clear and uncluttered?" |

---

## Appendix A: Quick Reference

### Minimum 80 Char Template (Copy-Paste)

```go
// Status bar template
"o8n v0.1.0 │ %s │ ✓ Connected │ ⚡ %dms │ %d items"

// Key hints template (definitions)
"? help  : switch  ↑↓ nav  Enter drill  Esc back  <ctrl>-r refresh"

// Key hints template (instances)
"? help  Esc back  ↑↓ nav  Enter vars  v vars  <ctrl>+d kill  <ctrl>-r refresh"

// Key hints template (variables)
"? help  Esc back  ↑↓ nav  / search  <ctrl>-r refresh  q quit"
```

### Color Constants

```go
const (
    ColorHelp        = "220"  // Yellow - help key
    ColorNavigation  = "%s"   // Environment color - navigation keys
    ColorAction      = "white" // White - action keys
    ColorDestructive = "196"  // Red - delete/kill keys
    ColorDescription = "252"  // Light gray - descriptions
    ColorSeparator   = "240"  // Dark gray - │ separators
)
```

---

**End of Compact Header Design**

Ready for implementation! 🚀
