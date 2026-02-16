# 📐 o8n Layout Design - Optimized Responsive System

**Feature:** Screen Layout Optimization  
**Priority:** 🟡 MEDIUM  
**Effort:** 8 hours  
**Status:** Design Complete  
**Designer:** BMM UX Designer  
**Date:** February 16, 2026

---

## Executive Summary

The current o8n layout is functional but **wastes valuable vertical space** (33% on header in minimum 80x24 terminal). This design provides a **responsive layout system** that maximizes content visibility while maintaining usability across terminal sizes:

**Key Improvements:**
- ✅ **Header: 8 rows → 3 rows** (saves 5 rows for content, 63% reduction)
- ✅ **Smart responsive behavior** - adapts from 80x24 to 200x60
- ✅ **Progressive disclosure** - show more info as space increases
- ✅ **Mobile-first approach** - optimized for minimum size first

**Impact:**
- More data visible without scrolling
- Professional, compact appearance
- Better use of premium screen real estate

---

## 1. Current Layout Analysis

### 1.1 Current Structure (80x24 terminal)

```
Row Distribution:
┌─ Header (rows 1-8) ───────────── 33% ────────────────┐
│ Column 1: Env info (8 lines)                         │
│ Column 2: Key bindings (8 lines)                     │
│ Column 3: ASCII logo (8 lines, 25 chars)             │
├─ Context Selector (row 9) ───────────────────────────┤
│ : [input________________]                            │
├─ Content Header (row 10) ─────────────────────────────┤
│ process-definitions                                  │
├─ Main Content (rows 11-23) ───── 54% ────────────────┤
│ KEY        │ NAME │ VERSION │ INSTANCES              │
│ ──────────────────────────────────────────────────   │
│ ▶ invoice  │ Rev  │ 2       │ 15                     │
│   order    │ Proc │ 3       │ 7                      │
│   shipping │ Ship │ 1       │ 0                      │
│   [10 more rows available...]                        │ ← Only 13 content rows!
├─ Footer (row 24) ────────────────────────────────────┤
│ [1] <process-definitions> | ⚡                        │
└───────────────────────────────────────────────────────┘

Measurements:
- Total height: 24 rows
- Header: 8 rows (33%)
- Context selector: 1 row (4%)
- Content header: 1 row (4%)
- Content: 13 rows (54%) ← THE PROBLEM
- Footer: 1 row (4%)
```

**Critical Issue:** Only **13 rows** for actual data in a 24-row terminal!

---

### 1.2 Space Waste Analysis

| Section | Current Rows | Waste | Improvement Potential |
|---------|--------------|-------|----------------------|
| Environment info | 3 lines | 2 rows | Combine into 1 row |
| API URL | 1 line | 0 | Keep |
| Username | 1 line | 0.5 | Combine with URL |
| Key bindings | 8 lines | 6 rows | Context-sensitive 1 row |
| ASCII logo | 8 rows | 8 rows | Remove (splash only) |
| **TOTAL WASTE** | **8 rows** | **16.5 rows** | **21% → 12.5% header** |

**Savings:** Reclaim 5-6 rows = **38% more content visible**

---

## 2. Optimized Layout Design

### 2.1 Compact Layout (80x24 minimum)

**Target:** Maximize content, maintain usability

```
Row Distribution (New):
┌─ Header (rows 1-3) ───────────── 12.5% ──────────────┐
│ o8n v0.1.0 │ local │ demo@localhost:8080 │ ⚡ 45ms   │ ← Row 1: Status bar
│ ? help  : switch  <ctrl>+e env  <ctrl>-r refresh     │ ← Row 2: Key hints
│                                                       │ ← Row 3: Spacer
├─ Context (row 4) ────────────────────────────────────┤
│ : [input________________]                            │
├─ Content (rows 5-23) ─────────── 79% ────────────────┤
│ ┌─ process-definitions ───────────────────────────┐  │ ← Integrated header
│ │ KEY        │ NAME │ VERSION │ INSTANCES         │  │
│ │ ──────────────────────────────────────────────  │  │
│ │ ▶ invoice  │ Rev  │ 2       │ 15                │  │
│ │   order    │ Proc │ 3       │ 7                 │  │
│ │   shipping │ Ship │ 1       │ 0                 │  │
│ │   payment  │ Pay  │ 4       │ 23                │  │
│ │   fulfill  │ Ful  │ 1       │ 8                 │  │
│ │   audit    │ Aud  │ 2       │ 0                 │  │
│ │   [13 more rows...]                             │  │ ← 19 content rows!
│ └─────────────────────────────────────────────────┘  │
├─ Footer (row 24) ─────────────────────────────────────┤
│ [1] <process-definitions> [2] <invoice-v2> | ⚡        │
└───────────────────────────────────────────────────────┘

Measurements:
- Total: 24 rows
- Header: 3 rows (12.5%)
- Context: 1 row (4%)
- Content: 19 rows (79%) ← 46% improvement!
- Footer: 1 row (4%)
```

**Key Changes:**
1. **Status bar** - Single line with all essential info
2. **Context-sensitive keys** - Only show relevant shortcuts
3. **No ASCII logo** - Shown on splash only
4. **Integrated content header** - Part of table box, not separate
5. **Result:** +6 rows for content (13 → 19 rows = +46%)

---

### 2.2 Medium Layout (120x30 terminal)

**Target:** Add more context without bloat

```
┌─ Header (rows 1-3) ───────────────────────────────────────────────────────────┐
│ o8n v0.1.0 │ local @ http://localhost:8080 │ demo │ ⚡ 45ms │ 23 definitions  │
│ ? help  : switch  <ctrl>+e env  <ctrl>-r refresh  <ctrl>+d delete  q quit     │
│                                                                                │
├─ Context (row 4) ─────────────────────────────────────────────────────────────┤
│ : [input____________________________________]                                 │
├─ Content (rows 5-29) ─────────────────────────────────────────────────────────┤
│ ┌─ process-definitions ──────────────────────────────────────────────────────┐│
│ │ KEY           │ NAME                    │ VER │ INST │ DEPLOYED           ││
│ │ ─────────────────────────────────────────────────────────────────────────  ││
│ │ ▶ invoice-v2  │ Invoice Review Process  │ 2   │ 15   │ 2026-02-15 14:30  ││
│ │   order-v3    │ Order Processing        │ 3   │ 7    │ 2026-02-16 09:15  ││
│ │   shipping    │ Ship Product            │ 1   │ 0    │ 2026-02-10 11:00  ││
│ │   [20 more rows...]                                                        ││
│ └────────────────────────────────────────────────────────────────────────────┘│
├─ Footer (row 30) ─────────────────────────────────────────────────────────────┤
│ [1] <process-definitions> [2] <invoice-v2> [3] <inst-abc123> | Last: 14:32 ⚡ │
└───────────────────────────────────────────────────────────────────────────────┘

Measurements:
- Total: 30 rows
- Header: 3 rows (10%)
- Context: 1 row (3%)
- Content: 25 rows (83%)
- Footer: 1 row (3%)
```

**Enhancements at 120x30:**
- Full URL visible
- More key hints (delete, quit)
- Stats in status bar (definition count)
- Longer column names
- Deployment timestamp column
- More breadcrumb visible

---

### 2.3 Large Layout (160x40+ terminal)

**Target:** Maximum information density

```
┌─ Header (rows 1-3) ───────────────────────────────────────────────────────────────────────────────────────────────┐
│ o8n v0.1.0 │ Environment: local @ http://localhost:8080/engine-rest │ User: demo │ API: ✓ 45ms │ Data: 23 definitions│
│ Navigation: ↑↓ move  Enter drill  Esc back  PgUp/Dn page │ Actions: <ctrl>+e env  <ctrl>-r refresh  <ctrl>+d delete │
│ Global: ? help  : switch  <ctrl>+c quit │ Active: Auto-refresh ON (5s) │ View: Definitions                         │
├─ Context (row 4) ─────────────────────────────────────────────────────────────────────────────────────────────────┤
│ : [input________________________________________________________________]                                         │
├─ Split View: List + Detail (rows 5-39) ──────────────────────────────────────────────────────────────────────────┤
│ ┌─ Definitions (70%) ──────────────────────┐ ┌─ Details (30%) ────────────────────────────────────────────────┐ │
│ │ KEY           │ NAME         │ VER │ INST │ │ Definition: invoice-v2                                        │ │
│ │ ───────────────────────────────────────  │ │ ───────────────────────────────────────────────────────────── │ │
│ │ ▶ invoice-v2  │ Invoice Rev  │ 2   │ 15  │ │ Key:           invoice-v2                                     │ │
│ │   order-v3    │ Order Proc   │ 3   │ 7   │ │ Name:          Invoice Review Process                        │ │
│ │   shipping    │ Ship Prod    │ 1   │ 0   │ │ Version:       2                                              │ │
│ │   payment-v4  │ Payment      │ 4   │ 23  │ │ ID:            invoice-v2:2:abc123-def456                     │ │
│ │   [30 more rows...]                       │ │ Instances:     15 active, 3 suspended                        │ │
│ │                                           │ │ Deployment:    2026-02-15 14:30:45                           │ │
│ │                                           │ │ Deployed by:   admin                                          │ │
│ │                                           │ │ Tenant ID:     (none)                                         │ │
│ │                                           │ │ Resource:      invoice-review-v2.bpmn                         │ │
│ │                                           │ │                                                               │ │
│ │                                           │ │ Description:                                                  │ │
│ │                                           │ │ Handles invoice approval workflow with three-stage review.   │ │
│ │                                           │ │ Includes automatic escalation after 48 hours.                │ │
│ │                                           │ │                                                               │ │
│ │                                           │ │ Actions:                                                      │ │
│ │                                           │ │ Enter    View instances                                       │ │
│ │                                           │ │ <ctrl>+d Delete deployment                                    │ │
│ └───────────────────────────────────────────┘ └───────────────────────────────────────────────────────────────┘ │
├─ Footer (row 40) ─────────────────────────────────────────────────────────────────────────────────────────────────┤
│ [1] <process-definitions> [2] <invoice-v2> [3] <instances> [4] <inst-abc123> [5] <variables> │ Updated: 14:32:15 ⚡│
└───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

Measurements:
- Total: 40 rows
- Header: 3 rows (7.5%)
- Context: 1 row (2.5%)
- Content: 35 rows (87.5%)
- Footer: 1 row (2.5%)
```

**Enhancements at 160x40:**
- **Split-pane view** - List (70%) + Detail panel (30%)
- **Detailed info panel** - Shows full metadata for selected item
- **Multi-line key hints** - Grouped by category
- **Extended breadcrumb** - Shows full navigation path
- **Timestamp in footer** - Last update time
- **Status indicators** - Auto-refresh state, API health

---

## 3. Component Specifications

### 3.1 Status Bar (Header Row 1)

**Purpose:** Show essential context at a glance  
**Height:** 1 row  
**Always visible:** Yes

#### Layout Calculation:

```go
// Status bar segments (left to right)
type StatusBarSegment struct {
    Text     string
    MinWidth int
    Priority int  // Higher = more important
    Style    lipgloss.Style
}

// Adaptive status bar builder
func buildStatusBar(width int, segments []StatusBarSegment) string {
    // Sort by priority (high to low)
    sort.Slice(segments, func(i, j int) bool {
        return segments[i].Priority > segments[j].Priority
    })
    
    // Calculate available space
    available := width - (len(segments) - 1) * 3  // " │ " separators
    
    // Add segments until space runs out
    var parts []string
    for _, seg := range segments {
        if available >= seg.MinWidth {
            parts = append(parts, seg.Style.Render(seg.Text))
            available -= seg.MinWidth
        }
    }
    
    return strings.Join(parts, " │ ")
}
```

#### Segment Priority (High → Low):

| Priority | Segment | MinWidth | Show When |
|----------|---------|----------|-----------|
| 100 | App version | 12 | Always |
| 90 | Environment name | 8 | Always |
| 80 | Connection status | 10 | Always |
| 70 | Username | 8 | Width ≥ 100 |
| 60 | API response time | 10 | Width ≥ 110 |
| 50 | Item count | 15 | Width ≥ 130 |
| 40 | Full URL | 30 | Width ≥ 140 |
| 30 | Last update time | 12 | Width ≥ 160 |

#### Examples by Width:

**80 chars:**
```
o8n v0.1.0 │ local │ ✓ Connected │ ⚡ 45ms
```

**120 chars:**
```
o8n v0.1.0 │ local @ http://localhost:8080 │ demo │ ✓ 45ms │ 23 items
```

**160 chars:**
```
o8n v0.1.0 │ Environment: local @ http://localhost:8080/engine-rest │ User: demo │ API: ✓ 45ms │ Data: 23 definitions │ Updated: 14:32
```

---

### 3.2 Key Hints Bar (Header Row 2)

**Purpose:** Context-sensitive keyboard shortcuts  
**Height:** 1 row  
**Dynamic:** Changes based on view mode

#### Design Pattern:

```
[icon] [key] [action]  [key] [action]  │  [icon] [key] [action]
└─ Navigation group ─────────────┘         └─ Action group ──┘
```

#### View-Specific Hints:

**Definitions View (80 char):**
```
? help  : switch  <ctrl>+e env  <ctrl>-r refresh  q quit
```

**Definitions View (120 char):**
```
? help  : switch  <ctrl>+e env  <ctrl>-r refresh  <ctrl>+d delete  Enter drill  q quit
```

**Instances View (80 char):**
```
? help  Esc back  <ctrl>+d kill  v variables  <ctrl>-r refresh
```

**Variables View (80 char):**
```
? help  Esc back  / search  e edit  q quit
```

#### Implementation:

```go
type KeyHint struct {
    Key         string
    Description string
    MinWidth    int
    Group       string  // "navigation", "action", "global"
}

// Build context-sensitive key hints
func (m *model) buildKeyHints() string {
    var hints []KeyHint
    
    // Always show help
    hints = append(hints, KeyHint{"?", "help", 8, "global"})
    
    switch m.viewMode {
    case "definitions":
        hints = append(hints,
            KeyHint{":", "switch", 10, "navigation"},
            KeyHint{"<ctrl>+e", "env", 11, "action"},
            KeyHint{"<ctrl>-r", "refresh", 15, "action"},
        )
        if m.lastWidth >= 100 {
            hints = append(hints,
                KeyHint{"<ctrl>+d", "delete", 14, "action"},
                KeyHint{"Enter", "drill", 12, "navigation"},
            )
        }
    case "instances":
        hints = append(hints,
            KeyHint{"Esc", "back", 10, "navigation"},
            KeyHint{"<ctrl>+d", "kill", 11, "action"},
            KeyHint{"v", "variables", 13, "navigation"},
        )
    case "variables":
        hints = append(hints,
            KeyHint{"Esc", "back", 10, "navigation"},
            KeyHint{"/", "search", 10, "action"},
        )
    }
    
    // Always add quit
    hints = append(hints, KeyHint{"q", "quit", 7, "global"})
    
    return renderKeyHints(hints, m.lastWidth)
}
```

---

### 3.3 Content Area Optimization

**Goal:** Maximize data visibility  

#### Current Issue:
```
┌─ Content Header ───────┐  ← Separate box (1 row + borders = 3 rows)
│ process-definitions    │
└────────────────────────┘

┌─ Table ────────────────┐  ← Main content box
│ KEY    │ NAME │ VER    │
│ ───────────────────────│
│ data   │ data │ data   │
└────────────────────────┘
```
**Waste:** 3 rows for header when it could be integrated

#### Optimized Design:
```
┌─ process-definitions ──┐  ← Header integrated into table border
│ KEY    │ NAME │ VER    │
│ ───────────────────────│
│ ▶ data │ data │ data   │
│   data │ data │ data   │
│   data │ data │ data   │
└────────────────────────┘
```
**Savings:** 2 rows back to content

#### Implementation:

```go
func (m model) renderContent() string {
    // Use table title instead of separate header
    tableStyle := table.DefaultStyles()
    
    // Set border with title
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(m.currentColor)).
        BorderTop(true).
        BorderTitle(m.contentHeader).  // ← Integrated!
        BorderTitleAlign(lipgloss.Left)
    
    return boxStyle.Render(m.table.View())
}
```

---

### 3.4 Footer Optimization

**Current:**
```
[1] <process-definitions> | ⚡
```
**Width used:** ~30 chars (waste in 160-char terminal)

**Optimized (80 char):**
```
[1] <defs> [2] <invoice-v2> | ⚡
```

**Optimized (120 char):**
```
[1] <process-definitions> [2] <invoice-v2> [3] <inst-123> | Last: 14:32 ⚡
```

**Optimized (160 char):**
```
[1] <process-definitions> [2] <invoice-v2> [3] <instances> [4] <inst-abc123> [5] <variables> | Updated: 14:32:15 | Refresh: ON (5s) ⚡
```

#### Footer Segments:

| Priority | Segment | MinWidth | Show When |
|----------|---------|----------|-----------|
| 100 | Breadcrumb[0] | 15 | Always |
| 90 | Flash indicator | 2 | Always |
| 80 | Breadcrumb[1] | 15 | Width ≥ 90 |
| 70 | Breadcrumb[2] | 15 | Width ≥ 110 |
| 60 | Last update | 12 | Width ≥ 130 |
| 50 | Refresh status | 15 | Width ≥ 150 |
| 40 | Extra crumbs | 15/each | Width ≥ 160 |

---

## 4. Responsive Behavior Matrix

### 4.1 Breakpoint System

| Size Class | Width Range | Height Range | Layout Mode | Notes |
|------------|-------------|--------------|-------------|-------|
| Minimum | 80-99 | 24-29 | Compact | Essential only |
| Small | 100-119 | 24-29 | Compact+ | Add username, stats |
| Medium | 120-139 | 30-39 | Standard | Full URL, more hints |
| Large | 140-159 | 30-39 | Enhanced | Detail pane option |
| XLarge | 160+ | 40+ | Split-view | List + detail panels |

### 4.2 Feature Visibility by Size

| Feature | Min (80x24) | Small (100x24) | Medium (120x30) | Large (160x40) |
|---------|-------------|----------------|-----------------|----------------|
| App version | ✓ | ✓ | ✓ | ✓ |
| Environment | ✓ (short) | ✓ (short) | ✓ (full URL) | ✓ (full) |
| Username | ✗ | ✓ | ✓ | ✓ |
| API response time | ✓ | ✓ | ✓ | ✓ |
| Item count | ✗ | ✓ | ✓ | ✓ |
| Key hints (basic) | 5 | 6 | 8 | 10+ |
| Breadcrumb depth | 2 | 3 | 4 | 5+ |
| Detail panel | ✗ | ✗ | ✗ | ✓ |
| Extended columns | ✗ | ✗ | ✓ | ✓ |
| Timestamps | ✗ | ✗ | ✓ | ✓ (extended) |

---

## 5. Implementation Roadmap

### Phase 1: Core Optimization (4 hours) 🔴 HIGH

**Goal:** Compact header (8 → 3 rows)

**Tasks:**
1. ✅ Implement status bar builder with adaptive segments
2. ✅ Implement context-sensitive key hints
3. ✅ Integrate content header into table border
4. ✅ Remove ASCII logo from main view (splash only)
5. ✅ Test on 80x24 minimum size

**Files to modify:**
- `main.go` - `View()` function
- `main.go` - Add `buildStatusBar()`, `buildKeyHints()`

**Expected result:**
```
Before: 13 content rows (54%)
After:  19 content rows (79%)
Improvement: +46%
```

---

### Phase 2: Responsive Enhancement (2 hours) 🟡 MEDIUM

**Goal:** Adapt layout to terminal size

**Tasks:**
1. ✅ Add breakpoint detection based on width/height
2. ✅ Implement segment priority system
3. ✅ Progressive footer expansion
4. ✅ Test on 80x24, 120x30, 160x40

**Files to modify:**
- `main.go` - Add `detectLayoutMode()` function
- `main.go` - Update `WindowSizeMsg` handler

**Expected result:**
- Small terminals: Clean, uncluttered
- Large terminals: Rich information without bloat

---

### Phase 3: Split-View Mode (2 hours) 🟢 LOW

**Goal:** Detail panel for large terminals

**Tasks:**
1. ✅ Detect when width ≥ 160
2. ✅ Implement horizontal split (70%/30%)
3. ✅ Render detail panel for selected item
4. ✅ Handle resize smoothly (collapse panel if shrinks)

**Files to modify:**
- `main.go` - Add `renderDetailPanel()`
- `main.go` - Modify `View()` for split layout

**Expected result:**
- 160+ width: Show list + details side-by-side
- No extra key required - automatic
- Detail shows full metadata, actions available

---

## 6. Code Samples

### 6.1 Status Bar Builder

```go
// StatusSegment represents one piece of status bar info
type StatusSegment struct {
    Text     string
    MinWidth int
    Priority int
    Color    string
}

// buildStatusBar creates adaptive status bar based on available width
func (m *model) buildStatusBar(width int) string {
    segments := []StatusSegment{
        {"o8n v0.1.0", 12, 100, "white"},
        {m.currentEnv, 8, 90, m.currentColor},
        {"✓ Connected", 12, 80, "42"},  // Green
        {m.username, 8, 70, "gray"},
        {fmt.Sprintf("⚡ %dms", m.apiLatency), 10, 60, "yellow"},
        {fmt.Sprintf("%d items", m.itemCount), 12, 50, "cyan"},
    }
    
    // Sort by priority
    sort.Slice(segments, func(i, j int) bool {
        return segments[i].Priority > segments[j].Priority
    })
    
    var parts []string
    available := width - 6  // Reserve space for separators
    
    for _, seg := range segments {
        if available >= seg.MinWidth {
            style := lipgloss.NewStyle().Foreground(lipgloss.Color(seg.Color))
            parts = append(parts, style.Render(seg.Text))
            available -= seg.MinWidth + 3  // +3 for " │ "
        }
    }
    
    separator := lipgloss.NewStyle().
        Foreground(lipgloss.Color("240")).
        Render(" │ ")
    
    return strings.Join(parts, separator)
}
```

---

### 6.2 Context-Sensitive Key Hints

```go
// KeyHint represents a keyboard shortcut hint
type KeyHint struct {
    Key  string
    Desc string
}

// buildKeyHints returns context-aware key hints for current view
func (m *model) buildKeyHints() string {
    var hints []KeyHint
    
    // Global hints (always shown)
    hints = append(hints, KeyHint{"?", "help"})
    
    // View-specific hints
    switch m.viewMode {
    case "definitions":
        hints = append(hints,
            KeyHint{":", "switch"},
            KeyHint{"<ctrl>+e", "env"},
            KeyHint{"<ctrl>-r", "refresh"},
        )
        if m.lastWidth >= 100 {
            hints = append(hints,
                KeyHint{"<ctrl>+d", "delete"},
                KeyHint{"Enter", "drill"},
            )
        }
        
    case "instances":
        hints = append(hints,
            KeyHint{"Esc", "back"},
            KeyHint{"<ctrl>+d", "kill"},
            KeyHint{"v", "vars"},
        )
        if m.lastWidth >= 110 {
            hints = append(hints,
                KeyHint{"s", "suspend"},
                KeyHint{"r", "resume"},
            )
        }
        
    case "variables":
        hints = append(hints,
            KeyHint{"Esc", "back"},
            KeyHint{"/", "search"},
        )
    }
    
    // Always add quit
    hints = append(hints, KeyHint{"q", "quit"})
    
    // Render with adaptive spacing
    return m.renderHints(hints)
}

func (m *model) renderHints(hints []KeyHint) string {
    var parts []string
    for _, h := range hints {
        keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.currentColor))
        descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("white"))
        
        part := fmt.Sprintf("%s %s",
            keyStyle.Render(h.Key),
            descStyle.Render(h.Desc),
        )
        parts = append(parts, part)
    }
    
    return strings.Join(parts, "  ")
}
```

---

### 6.3 Integrated Content Header

```go
func (m *model) renderContent() string {
    // Create border style with integrated title
    borderStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(m.currentColor)).
        BorderTop(true).
        BorderTitle(m.contentHeader).  // ← Title in border!
        BorderTitleAlign(lipgloss.Left).
        Padding(0, 1).
        Width(m.paneWidth).
        Height(m.paneHeight)
    
    return borderStyle.Render(m.table.View())
}
```

**Before:**
```
┌────────────────────────┐
│ process-definitions    │  ← 3 rows wasted
└────────────────────────┘
┌────────────────────────┐
│ KEY │ NAME │ VERSION   │
└────────────────────────┘
```

**After:**
```
┌─ process-definitions ──┐
│ KEY │ NAME │ VERSION   │  ← 2 rows saved!
└────────────────────────┘
```

---

### 6.4 Optimized View() Function

```go
func (m model) View() string {
    // Splash screen (unchanged)
    if m.splashActive {
        return m.renderSplash()
    }
    
    // Detect layout mode based on size
    mode := m.detectLayoutMode()
    
    // Build components
    statusBar := m.buildStatusBar(m.lastWidth)
    keyHints := m.buildKeyHints()
    spacer := ""  // Optional spacer for breathing room
    
    // Header (3 rows total: status + hints + spacer)
    header := lipgloss.JoinVertical(lipgloss.Left,
        statusBar,
        keyHints,
        spacer,
    )
    
    // Context selector (unchanged, 1 row)
    contextBox := m.renderContextSelector()
    
    // Main content (integrated header)
    var content string
    if mode == "split" && m.lastWidth >= 160 {
        // Split view: list + detail panel
        content = m.renderSplitView()
    } else {
        // Standard view: list only
        content = m.renderContent()
    }
    
    // Footer (1 row)
    footer := m.renderFooter()
    
    // Compose final layout
    return lipgloss.JoinVertical(lipgloss.Left,
        header,
        contextBox,
        content,
        footer,
    )
}

func (m *model) detectLayoutMode() string {
    w := m.lastWidth
    h := m.lastHeight
    
    if w >= 160 && h >= 40 {
        return "split"
    } else if w >= 120 && h >= 30 {
        return "standard"
    } else if w >= 100 {
        return "compact+"
    }
    return "compact"
}
```

---

### 6.5 Split View (Large Terminals)

```go
func (m *model) renderSplitView() string {
    // Calculate widths: 70% list, 30% detail
    totalWidth := m.paneWidth
    listWidth := int(float64(totalWidth) * 0.70)
    detailWidth := totalWidth - listWidth - 2  // -2 for gap
    
    // Render list (left pane)
    listStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(m.currentColor)).
        BorderTitle("Definitions").
        Width(listWidth).
        Height(m.paneHeight)
    
    listPane := listStyle.Render(m.table.View())
    
    // Render detail (right pane)
    detailContent := m.buildDetailPanel()
    
    detailStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(m.currentColor)).
        BorderTitle("Details").
        Width(detailWidth).
        Height(m.paneHeight)
    
    detailPane := detailStyle.Render(detailContent)
    
    // Join horizontally
    return lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)
}

func (m *model) buildDetailPanel() string {
    selectedRow := m.table.SelectedRow()
    if len(selectedRow) == 0 {
        return "No selection"
    }
    
    var details strings.Builder
    
    switch m.viewMode {
    case "definitions":
        // Show process definition details
        details.WriteString("Definition: " + selectedRow[1] + "\n\n")
        details.WriteString("Key:        " + selectedRow[0] + "\n")
        details.WriteString("Version:    " + selectedRow[2] + "\n")
        details.WriteString("Instances:  " + selectedRow[3] + "\n")
        details.WriteString("\nActions:\n")
        details.WriteString("Enter    View instances\n")
        details.WriteString("<ctrl>+d Delete deployment\n")
        
    case "instances":
        // Show instance details
        details.WriteString("Instance: " + selectedRow[0] + "\n\n")
        details.WriteString("Status:     " + selectedRow[1] + "\n")
        details.WriteString("Started:    " + selectedRow[2] + "\n")
        details.WriteString("\nActions:\n")
        details.WriteString("Enter    View variables\n")
        details.WriteString("<ctrl>+d Kill instance\n")
        details.WriteString("s        Suspend\n")
        details.WriteString("r        Resume\n")
    }
    
    return details.String()
}
```

---

## 7. Testing Matrix

### 7.1 Size Testing Checklist

Test on each size class to ensure proper adaptation:

| Terminal Size | Status Bar | Key Hints | Content Rows | Footer | Detail Panel |
|---------------|------------|-----------|--------------|--------|--------------|
| 80x24 | ✓ Essential (3 parts) | ✓ 5 hints | ✓ 19 rows | ✓ 2 crumbs | ✗ |
| 100x26 | ✓ + username | ✓ 6 hints | ✓ 21 rows | ✓ 3 crumbs | ✗ |
| 120x30 | ✓ + full URL | ✓ 8 hints | ✓ 25 rows | ✓ 4 crumbs | ✗ |
| 140x35 | ✓ + stats | ✓ 10 hints | ✓ 30 rows | ✓ 5 crumbs | ✗ |
| 160x40 | ✓ Full | ✓ 12 hints | ✓ 35 rows | ✓ Full | ✓ Yes |

### 7.2 Resize Testing

**Test scenario:** Start at 160x40, resize down to 80x24, resize back up

**Expected behavior:**
1. Split view → Standard view (smooth transition)
2. Extended status → Compact status (no overflow)
3. All key hints → Essential hints (no wrapping)
4. Footer contracts (no flickering)
5. Resize up: Features progressively reappear

### 7.3 Edge Cases

- [ ] Terminal smaller than 80x24 (graceful degradation)
- [ ] Terminal wider than 200 chars (reasonable max width)
- [ ] Rapid resize events (debouncing)
- [ ] Very long breadcrumb (truncation)
- [ ] Very long instance IDs (column overflow)
- [ ] Unicode characters in status bar (width calculation)

---

## 8. Success Metrics

### Before vs After Comparison

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Header rows (80x24) | 8 | 3 | -62.5% |
| Content rows (80x24) | 13 | 19 | +46% |
| Content % of screen | 54% | 79% | +25pp |
| Wasted space | 33% | 12.5% | -20.5pp |
| Items visible (avg) | 10 | 16 | +60% |

### User Experience Goals

| Goal | Target | Measurement |
|------|--------|-------------|
| More data visible | +40% | Count rows before/after |
| Faster scanning | -30% time | Time to find item in list |
| Professional look | 4.5/5 rating | User survey |
| Responsive feel | < 16ms resize | Frame time measurement |
| No information loss | 100% | All features accessible |

---

## 9. Migration Plan

### Rollout Strategy

**Phase 1 (Safe):** Make header changes only
- Low risk, high impact
- Easy to rollback
- Users see immediate benefit

**Phase 2 (Enhancement):** Add responsive behavior
- Progressive enhancement
- No breaking changes
- Enhances large terminals

**Phase 3 (Feature):** Split-view mode
- New capability
- Optional (automatic on large screens)
- Can be disabled if issues arise

### Rollback Plan

If issues arise, revert in reverse order:
1. Disable split-view (keep optimized header)
2. Disable responsive features (keep compact header)
3. Full rollback to 8-row header (last resort)

---

## Appendix A: Layout Calculation Formulas

### Available Content Height

```
contentHeight = terminalHeight - headerRows - contextRows - footerRows - borders
```

**Examples:**
```
80x24:  24 - 3 - 1 - 1 - 2 = 17 inner rows + 2 border = 19 visible
120x30: 30 - 3 - 1 - 1 - 2 = 23 inner rows + 2 border = 25 visible
160x40: 40 - 3 - 1 - 1 - 2 = 33 inner rows + 2 border = 35 visible
```

### Column Width Distribution

**Simple (no detail panel):**
```
tableWidth = terminalWidth - 4  // 2 chars padding on each side
```

**Split view:**
```
listWidth = (terminalWidth * 0.70) - 2
detailWidth = (terminalWidth * 0.30) - 2
```

### Status Bar Segment Fitting

```go
// Greedy algorithm: fit highest priority segments first
func fitSegments(segments []Segment, availableWidth int) []Segment {
    sort.ByPriority(segments)  // High to low
    
    var fitted []Segment
    remaining := availableWidth
    
    for _, seg := range segments {
        if remaining >= seg.MinWidth + separatorWidth {
            fitted = append(fitted, seg)
            remaining -= seg.MinWidth + separatorWidth
        }
    }
    
    return fitted
}
```

---

## Appendix B: Visual Comparison

### Before (Current Layout)

```
 1  ┌─────────────────────────────────────┐  ▲
 2  │ Environment: local                  │  │
 3  │ URL: http://localhost:8080          │  │
 4  │ User: demo                          │  │ Header
 5  │ KEYS:                               │  │ 8 rows
 6  │ q - quit                            │  │ (33%)
 7  │ e - switch env                      │  │
 8  │ r - toggle auto-refresh             │  ▼
 9  │ : [input__________________]         │  Context (4%)
10  │ process-definitions                 │  Header (4%)
11  ┌───────────────────────────────────┐ │  ▲
12  │ KEY    │ NAME      │ VERSION      │ │  │
13  │ ───────────────────────────────── │ │  │
14  │ ▶ inv  │ Invoice   │ 2            │ │  │
15  │   ord  │ Order     │ 3            │ │  │ Content
16  │   shi  │ Shipping  │ 1            │ │  │ 13 rows
17  │   pay  │ Payment   │ 4            │ │  │ (54%)
18  │   ful  │ Fulfill   │ 1            │ │  │
19  │   aud  │ Audit     │ 2            │ │  │
20  │   not  │ Notify    │ 1            │ │  │
21  │   rep  │ Report    │ 3            │ │  │
22  │   arc  │ Archive   │ 1            │ │  │
23  │   [...]                           │ │  ▼
24  │ [1] <process-definitions> | ⚡     │  Footer (4%)
    └───────────────────────────────────┘
```

### After (Optimized Layout)

```
 1  o8n v0.1.0 │ local │ demo@localhost:8080 │ ⚡ 45ms    ▲ Header
 2  ? help  : switch  <ctrl>+e env  <ctrl>-r refresh  q quit  │ 3 rows
 3                                                             ▼ (12.5%)
 4  : [input__________________________]                         Context (4%)
 5  ┌─ process-definitions ──────────────────────────────┐    ▲
 6  │ KEY       │ NAME              │ VER │ INSTANCES   │    │
 7  │ ──────────────────────────────────────────────────│    │
 8  │ ▶ invoice │ Invoice Review    │ 2   │ 15          │    │
 9  │   order   │ Order Processing  │ 3   │ 7           │    │
10  │   ship    │ Ship Product      │ 1   │ 0           │    │
11  │   payment │ Payment Gateway   │ 4   │ 23          │    │
12  │   fulfill │ Fulfillment       │ 1   │ 8           │    │ Content
13  │   audit   │ Audit Trail       │ 2   │ 0           │    │ 19 rows
14  │   notify  │ Notification      │ 1   │ 12          │    │ (79%)
15  │   report  │ Report Generator  │ 3   │ 4           │    │
16  │   archive │ Archive Process   │ 1   │ 0           │    │
17  │   approve │ Approval Workflow │ 2   │ 6           │    │
18  │   review  │ Review Process    │ 1   │ 9           │    │
19  │   escalat │ Escalation        │ 1   │ 2           │    │
20  │   [6 more rows...]                                │    │
21  │   [...]                                           │    │
22  │   [...]                                           │    │
23  └───────────────────────────────────────────────────┘    ▼
24  [1] <process-definitions> [2] <invoice-v2> | ⚡            Footer (4%)
```

**Visual Impact:**
- ✅ 6 more data rows visible (+60%)
- ✅ Cleaner, more professional appearance
- ✅ Essential info still accessible
- ✅ More context in footer (breadcrumb depth)

---

**End of Layout Design Specification**

Ready for implementation! 🚀
