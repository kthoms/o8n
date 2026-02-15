# 🎨 o8n Terminal UI - Comprehensive UX Review

**Reviewer:** BMM UX Designer  
**Date:** February 15, 2026  
**Application:** o8n - Operaton Terminal UI  
**Version:** 0.1.0  
**Status:** Post Sprint 1 Phase 1 Refactoring

---

## Executive Summary

**Overall UX Grade: B+** (Good with room for polish)

Your o8n terminal UI shows strong fundamentals inspired by k9s. The keyboard-driven workflow, drill-down navigation, and responsive layout are well-designed. However, there are several UX improvements that would elevate the experience from "good" to "excellent."

**Strengths:**
- ✅ Clear vertical layout hierarchy
- ✅ Keyboard-first design
- ✅ Context awareness (breadcrumbs, environment display)
- ✅ Visual feedback (flash indicator for API calls)
- ✅ Responsive to terminal resize

**Priority Issues:**
- 🔴 **Critical:** Help screen (`?`) not implemented - discoverability issue
- 🟡 **High:** Modal confirmations missing (using key sequence instead)
- 🟡 **High:** Column hiding not implemented (responsive design incomplete)
- 🟡 **Medium:** No search/filter functionality
- 🟢 **Low:** Key bindings could be more mnemonic

---

## 1. Information Architecture 📊

### Current State
```
┌─ Header (8 rows, unboxed) ──────────────────────────┐
│ Column 1: Environment info                          │
│ Column 2: Key bindings                              │
│ Column 3: ASCII logo (right-aligned, 25 chars)      │
├─ Context Selection (dynamic, 1 row boxed) ──────────┤
│ : [input with completion]                           │
├─ Main Content (boxed table, fills remaining) ───────┤
│ HEADER1    HEADER2    HEADER3    HEADER4            │
│ ───────────────────────────────────────────────     │
│ ▶ value1   value2     value3     value4             │
│   value1   value2     value3     value4             │
├─ Footer (1 row) ─────────────────────────────────────┤
│ <breadcrumb1> <breadcrumb2> | message | ⚡          │
└──────────────────────────────────────────────────────┘
```

### ✅ Strengths

1. **Clear Hierarchy**
   - 4 distinct sections with clear purposes
   - Fixed header/footer, flexible content (correct!)
   - Visual separation through boxing

2. **Context Awareness**
   - Environment shown in header (good!)
   - Breadcrumbs in footer (excellent!)
   - Content header shows drill-down context

3. **Information Density**
   - Header: Essential context without clutter
   - Content: Table format maximizes data visibility
   - Footer: Status at a glance

### 🟡 Issues & Recommendations

#### Issue 1.1: Header Takes Too Much Space (Medium Priority)
**Problem:** 8 rows for header is significant (33% of minimum 24-row terminal)

**Impact:** Less space for actual content

**Recommendation:**
```
┌─ Header (3-4 rows) ──────────────────────────────────┐
│ o8n v0.1.0 | local @ http://localhost:8080 | demo    │
│ ? help | : switch | e env | r refresh | q quit       │
│                                                       │
├─ ...
```

**Benefits:**
- Saves 4-5 rows for content
- Still shows essential information
- More compact = more professional

**Implementation:**
- Reduce to 3 rows max
- First row: App name, env, URL, user (single line)
- Second row: Key help (single line, context-sensitive)
- Third row: Spacer/logo (optional, can be removed)

#### Issue 1.2: ASCII Logo Not Functional (Low Priority)
**Problem:** Logo takes 25 characters but provides no value during use

**Recommendation:**
- Show logo only on splash screen (already done!)
- Use logo space for more key help or stats
- Example: Show API response time, data freshness

**Alternative Design:**
```
┌─ Header (2 rows) ────────────────────────────────────┐
│ o8n v0.1.0 │ local │ demo@localhost:8080 │ ⚡ 45ms   │
│ ? help  : context  e env  r refresh  q quit          │
├─ ...
```

---

## 2. Navigation & Interaction Patterns 🧭

### Current Interaction Flow

```
State Machine:
┌─────────────────┐
│ Definitions     │ ◀─── Start here
│ (list view)     │
└────────┬────────┘
         │ Enter
         ▼
┌─────────────────┐
│ Instances       │
│ (filtered list) │
└────────┬────────┘
         │ Enter
         ▼
┌─────────────────┐
│ Variables       │
│ (key-value)     │
└─────────────────┘
         │
    Esc to go back at any level
```

### ✅ Strengths

1. **Clear Mental Model**
   - Definitions → Instances → Variables (natural hierarchy)
   - Enter drills down, Esc goes back (standard pattern)
   - Breadcrumbs show current location

2. **Keyboard Efficiency**
   - Arrow keys for navigation (standard)
   - Single key for common actions (e, r, q)
   - No unnecessary modifier keys

3. **Context Switching**
   - `:` for switching (vim-inspired, good!)
   - Tab completion (excellent!)
   - Enter only on exact match (safe!)

### 🔴 Critical Issues

#### Issue 2.1: Help Not Discoverable (CRITICAL)
**Problem:** Specification mentions `?` for help, but it's NOT IMPLEMENTED

**Impact:** 
- Users can't discover features
- New users are lost
- Reduces adoption

**User Quote:** _"I pressed ? and nothing happened. How do I know what keys to press?"_

**Recommendation:** IMPLEMENT HELP SCREEN IMMEDIATELY

**Design:**
```
┌─ Help ───────────────────────────────────────────────┐
│                     o8n - Help                        │
│                                                       │
│ NAVIGATION                    ACTIONS                │
│ ↑/↓/PgUp/PgDn  Navigate      e  Switch environment   │
│ Enter          Drill down    r  Toggle auto-refresh  │
│ Esc            Go back       x  Kill instance        │
│ :              Switch view   /  Search (future)      │
│                                                       │
│ GLOBAL                       VIEWS                   │
│ ?              Help          :  Context switcher     │
│ q/Ctrl+C       Quit          Tab Complete            │
│                                                       │
│                Press any key to close                │
└───────────────────────────────────────────────────────┘
```

**Implementation Priority:** 🔴 **URGENT** - This is a standard TUI feature

#### Issue 2.2: Kill Instance Uses Key Sequence (High Priority)
**Problem:** `x` + `y` sequence instead of modal confirmation

**Current:**
1. Press `x` (marks for kill)
2. Press `y` (confirms kill)
3. No visual confirmation step

**Issues:**
- Not discoverable (how do users know about `y`?)
- Easy to accidentally kill (muscle memory)
- No clear cancellation path
- Not standard TUI pattern

**Recommendation:** Use modal dialog (k9s pattern)

**Design:**
```
┌─ Content ─────────────────────────────────────────────┐
│ ID          KEY         VERSION                       │
│ ───────────────────────────────────────────────────   │
│ inst-123    invoice     2                             │
│ inst-456    review      1                             │
│                                                        │
│   ┌─ Confirm Kill Instance ─────────────────┐        │
│   │                                          │        │
│   │ Kill instance: inst-123 (invoice v2)?   │        │
│   │                                          │        │
│   │ This action cannot be undone.           │        │
│   │                                          │        │
│   │      [Y] Yes    [N] No                  │        │
│   └──────────────────────────────────────────┘        │
└───────────────────────────────────────────────────────┘
```

**Benefits:**
- Clear confirmation required
- Shows what will be killed
- Easy to cancel (N or Esc)
- Standard pattern users expect

### 🟡 Medium Priority Issues

#### Issue 2.3: No Search/Filter Functionality
**Problem:** Large lists are hard to navigate

**Recommendation:** Add `/` key for search (vim pattern)

**Design:**
```
┌─ Footer ──────────────────────────────────────────────┐
│ /invoice_                                       ⚡     │
└───────────────────────────────────────────────────────┘
        ↑
    Search mode: type to filter, Enter to select
```

**Benefits:**
- Faster navigation in large datasets
- Standard vim pattern (discoverable)
- Non-destructive (Esc clears filter)

#### Issue 2.4: Context Selection Input Not Clear
**Problem:** When `:` is pressed, input focus not obvious

**Current:**
```
├─ Context Selection ──────────────────────────────────┤
│ : proc_                                              │
│      ↑ cursor here, but is it obvious?               │
└──────────────────────────────────────────────────────┘
```

**Recommendation:** Make active input more prominent

**Design:**
```
├─ Context Selection ──────────────────────────────────┤
│ : proc█ess-definitions                               │
│      ↑ block cursor + completion hint                │
└──────────────────────────────────────────────────────┘
```

Or with color:
```
│ : [proc]ess-definitions                              │
│    ^^^^^ typed (bright) | ^^^^^^^^^^^^^ hint (gray)  │
```

---

## 3. Visual Design & Hierarchy 🎨

### Current Color Usage

From your spec:
- Borders: Environment UI color (`ui_color` from config)
- Header text: White (uppercase for emphasis)
- Selected row: Inverted colors
- Completion: Gray (#666666)
- Error: Red (#FF0000) bold
- Breadcrumb: Different background per level
- Flash indicator: ⚡ (appears for 0.2s)

### ✅ Strengths

1. **Consistent Theming**
   - Configurable color schemes (skins folder)
   - Environment-specific colors
   - Good contrast awareness

2. **Visual Feedback**
   - Flash on API calls (excellent!)
   - Inverted selection (clear!)
   - Error messages in red (obvious!)

3. **Progressive Enhancement**
   - Colors add information
   - Still functional without color (breadcrumbs use text)
   - Good accessibility baseline

### 🟡 Issues & Recommendations

#### Issue 3.1: Table Headers Could Be More Prominent
**Problem:** Uppercase helps but more can be done

**Current:**
```
│ KEY         NAME              VERSION    RESOURCE     │
│ ─────────────────────────────────────────────────── │
│ invoice     Invoice Receipt   2          invoice.bpmn │
```

**Recommendation:** Add background color to headers

**Design:**
```
│█KEY█████████NAME██████████████VERSION███RESOURCE████│
│ invoice     Invoice Receipt   2         invoice.bpmn │
│ review      Review Invoice    1         review.bpmn  │
```

**Implementation:**
```go
// In your table styles
tStyles.Header = tStyles.Header.
    Background(lipgloss.Color(headerBgColor)).
    Foreground(lipgloss.Color("white")).
    Bold(true)
```

#### Issue 3.2: No Visual Distinction for Context Title
**Problem:** Context title on table border hard to read

**Current (from spec):**
```
├─ Main Content (process-definitions(ReviewInvoice)) ──┤
│     ↑ small, in border, easy to miss                 │
```

**Recommendation:** Make context more prominent

**Design Option 1 - Dedicated Line:**
```
├─ Main Content ───────────────────────────────────────┤
│ Context: process-definitions > ReviewInvoice         │
│ ────────────────────────────────────────────────────  │
│ KEY         NAME              VERSION                │
```

**Design Option 2 - Styled Title:**
```
├─ [Process Definitions › ReviewInvoice] ──────────────┤
│  ^^^^^ breadcrumb style in border                    │
```

#### Issue 3.3: Footer Could Show More Status
**Problem:** Footer has space but only shows context and flash

**Current:**
```
│ <process-definitions> | error message here | ⚡      │
```

**Recommendation:** Add useful stats

**Design:**
```
│ <definitions> <instances> | 23 items | ⚡ 45ms | ●   │
│ ^^^^^^^^^^^^^ breadcrumb  | count   | API  | conn │
```

**Information to show:**
- Item count (how many in current view)
- API response time (performance feedback)
- Connection status (● green = ok, ● red = error)
- Auto-refresh status (⟳ spinning = on)

---

## 4. Responsive Design 📱

### Terminal Size Handling

**Minimum:** 80x24 (standard)
**Target:** 120x40+ (common modern terminals)

### ✅ Strengths

1. **Layout Adapts**
   - WindowSizeMsg handled
   - Content section flexible
   - Header/footer fixed

2. **Column Width Calculation**
   - Percentage-based (good!)
   - Minimum widths defined (smart!)

### 🟡 Issues & Recommendations

#### Issue 4.1: Column Hiding Not Implemented
**Problem:** Spec says columns can hide when space constrained, but not implemented

**Impact:**
- Horizontal scrolling or truncation
- Poor experience on narrow terminals
- Data becomes unreadable

**Recommendation:** Implement column priority system

**Design:**
```go
// In config
type ColumnDef struct {
    Name     string
    Visible  bool
    Width    string
    Align    string
    Priority int  // 1 = must show, 5 = hide first
}
```

**Example:**
```yaml
columns:
  - name: id
    priority: 3  # Hide this first if space tight
  - name: key
    priority: 1  # Always show
  - name: name
    priority: 1  # Always show
  - name: version
    priority: 2  # Show if possible
  - name: resource
    priority: 3  # Hide if space tight
```

**Algorithm:**
```
1. Calculate required width for all priority 1 columns
2. Add priority 2 if space available
3. Add priority 3+ if space available
4. Update table columns dynamically
```

#### Issue 4.2: Ellipsis Truncation Not Clear
**Problem:** "..." truncation specified but user doesn't know full value

**Current:**
```
│ very-long-process-defin...  │  ← what was the rest?
```

**Recommendation:** Show full value on hover OR in footer

**Design - Footer approach:**
```
┌─ Content ──────────────────────────────────────────┐
│ very-long-process-defin...                        │ ← selected
├─ Footer ───────────────────────────────────────────┤
│ <definitions> | very-long-process-definition-name │
│                 ^^^^^^^ show full selected value   │
```

---

## 5. Keyboard Shortcuts ⌨️

### Current Key Map

| Key | Action | Mnemonic? | Standard? |
|-----|--------|-----------|-----------|
| `q` | Quit | ✅ Yes | ✅ Standard |
| `Ctrl+C` | Quit | ✅ Yes | ✅ Standard |
| `e` | Switch env | ✅ Yes (Environment) | ⚠️ Not standard |
| `r` | Toggle refresh | ✅ Yes (Refresh) | ✅ Standard |
| `:` | Context switch | ✅ Yes (vim) | ✅ Standard (vim) |
| `Enter` | Drill down | ✅ Yes | ✅ Standard |
| `Esc` | Go back | ✅ Yes | ✅ Standard |
| `x` | Kill | ⚠️ Partial | ⚠️ Usually `d` (delete) |
| `y` | Confirm | ❌ Hidden | ❌ Not discoverable |
| `?` | Help | ✅ Yes | ✅ Standard |
| `↑↓` | Navigate | ✅ Yes | ✅ Standard |

### ✅ Strengths

1. **Mostly Mnemonic**
   - q = quit, r = refresh (obvious)
   - : = context (vim users know this)

2. **No Conflicts**
   - Avoids shell bindings (Ctrl+A, Ctrl+E, etc.)
   - Works in most terminals

### 🟡 Recommendations

#### Recommendation 5.1: Change Kill Key to `d`
**Reason:** In k9s and most TUIs, `d` = delete/destroy

**Current:** `x` + `y`  
**Proposed:** `d` → modal

**Benefits:**
- More standard (k9s uses `d`)
- Single key + modal = better UX
- `x` could be used for something else (export?)

#### Recommendation 5.2: Add More Power User Keys

**Missing Standard Keys:**
| Key | Proposed Action | Why |
|-----|----------------|-----|
| `/` | Search/filter | vim standard |
| `g` | Go to top | vim standard |
| `G` | Go to bottom | vim standard |
| `Ctrl+R` | Force refresh | standard |
| `n` | Next match | vim (after search) |
| `N` | Previous match | vim (after search) |
| `1-9` | Quick env switch | numbered items |
| `Tab` | Next pane | tmux pattern |

#### Recommendation 5.3: Context-Sensitive Keys

**Example:** In process instances view:
- `l` = view logs
- `v` = view variables
- `r` = restart instance
- `s` = suspend instance

**Show context keys in help:**
```
┌─ Help (Process Instances) ───────────────────────────┐
│ INSTANCE ACTIONS                                     │
│ d        Delete instance                             │
│ r        Restart instance                            │
│ s        Suspend instance                            │
│ v        View variables                              │
│ l        View logs (if available)                    │
└──────────────────────────────────────────────────────┘
```

---

## 6. Feedback & Error Handling 💬

### Current Feedback Mechanisms

1. **API Calls:** ⚡ flash for 0.2s (excellent!)
2. **Errors:** Red text in footer message column
3. **Selection:** Inverted colors on selected row
4. **Loading:** (Not mentioned - assumed none?)

### ✅ Strengths

1. **Flash Indicator**
   - Immediate feedback
   - Subtle but noticeable
   - Good performance perception

2. **Error Display**
   - Red color (obvious)
   - In footer (doesn't block content)
   - Clears after 4 seconds

### 🟡 Issues & Recommendations

#### Issue 6.1: No Loading Indicator
**Problem:** Long API calls have only flash, no progress

**Impact:**
- User doesn't know if app is frozen
- Poor experience on slow connections

**Recommendation:** Show spinner during loads

**Design:**
```
├─ Footer ───────────────────────────────────────────┤
│ <definitions> | Loading... ⟳ | ⚡                  │
│                          ^^^^ spinner              │
└─────────────────────────────────────────────────────┘
```

**Or in content:**
```
┌─ Content ────────────────────────────────────────────┐
│                                                      │
│                  ⟳ Loading data...                   │
│                                                      │
│          (This may take a few seconds)               │
│                                                      │
└──────────────────────────────────────────────────────┘
```

#### Issue 6.2: Error Messages Could Be More Actionable
**Current:** "Error loading variables: 404 Not Found"

**Better:** "Cannot load variables: Instance not found. Press Esc to go back."

**Principle:** Tell users:
1. What went wrong (clear description)
2. Why it happened (context)
3. What to do next (actionable)

**Examples:**
```
❌ "Failed to fetch instances"
✅ "Cannot load instances: API unreachable. Check network and press r to retry."

❌ "Error: 403"
✅ "Access denied: Invalid credentials. Press e to switch environment."

❌ "Timeout"
✅ "Request timed out after 10s. Press r to retry or q to quit."
```

#### Issue 6.3: Success Feedback Missing
**Problem:** After killing instance, no confirmation

**Current:**
- Flash appears (⚡)
- Instance disappears from list
- But was it successful?

**Recommendation:** Show brief success message

**Design:**
```
├─ Footer ───────────────────────────────────────────┤
│ <instances> | ✓ Instance inst-123 killed | ⚡     │
│               ^^^^^ green checkmark                │
└─────────────────────────────────────────────────────┘
```

**Auto-clear after 2-3 seconds**

---

## 7. Accessibility ♿

### Current State
- ✅ Keyboard-only navigation (good!)
- ✅ Color not sole information channel
- ⚠️ No screen reader considerations (TUI limitation)
- ✅ Configurable colors (themes)

### 🟡 Recommendations

#### Recommendation 7.1: Color-Blind Friendly Themes
**Action:** Test your themes with color-blind simulator

**Critical combinations to test:**
- Red/Green (most common color blindness)
- Blue/Yellow (less common but important)

**Safe patterns:**
- Use symbols + color (⚡ icon + color)
- Use position + color (breadcrumb text + color)
- Use brightness + color (dark/light + hue)

**Example safe error design:**
```
│ ✗ Error | Cannot connect to API | ● RED    │
│ ^^^^     ^^^^^^^^^^^^^^^^^^^^^   ^^^^^^^   │
│ symbol   text description        indicator │
```

Even without color:
```
│ ✗ Error | Cannot connect to API | ●        │
```
Still clear!

#### Recommendation 7.2: Contrast Ratios
**Minimum:** 4.5:1 for normal text, 3:1 for large text (WCAG AA)

**Check your themes:**
```yaml
# Example theme audit
background: "#1e1e1e"
foreground: "#d4d4d4"  # Check: 11.6:1 ✅ Excellent

error: "#ff0000"       # Check on dark bg: 5.5:1 ✅ Good
border: "#00a8e1"      # Check on dark bg: 4.2:1 ⚠️ Borderline
```

**Tool:** Use online contrast checker or add to your CI

---

## 8. Performance Perception ⚡

### Current Performance Features
- ✅ Flash indicator (shows API activity)
- ✅ Auto-refresh toggle (user control)
- ⚠️ No caching mentioned
- ⚠️ No pagination mentioned

### 🟡 Recommendations

#### Recommendation 8.1: Show Response Times
**Why:** Users want to know if slowness is them or the API

**Design:**
```
├─ Footer ────────────────────────────────────────────┤
│ <definitions> | 23 items | ⚡ 145ms | ●             │
│                            ^^^^^^^ response time    │
└──────────────────────────────────────────────────────┘
```

**Color code:**
- < 100ms: Green (●)
- 100-500ms: Yellow (●)
- > 500ms: Orange (●)
- Error: Red (●)

#### Recommendation 8.2: Pagination for Large Datasets
**Problem:** Loading 10,000 instances is slow

**Recommendation:** Implement pagination

**Design:**
```
┌─ Content ────────────────────────────────────────────┐
│ Showing 1-50 of 1,247 instances                      │
│ ────────────────────────────────────────────────────  │
│ ID          KEY         STATUS                       │
│ inst-001    invoice     running                      │
│ ...                                                  │
├─ Footer ─────────────────────────────────────────────┤
│ <instances> | Page 1/25 | [/n]ext [/p]rev | ⚡      │
└──────────────────────────────────────────────────────┘
```

**Keys:**
- `n` or `]` = next page
- `p` or `[` = previous page
- `1-9` = jump to page
- `gg` = first page (vim)
- `G` = last page (vim)

---

## 9. Comparison with k9s (Your Inspiration) 🎯

### What k9s Does Well (That You Should Adopt)

| Feature | k9s | o8n Current | Recommendation |
|---------|-----|-------------|----------------|
| **Help Screen** | `?` comprehensive | ❌ Missing | 🔴 MUST ADD |
| **Search** | `/` filter | ❌ Missing | 🟡 Should add |
| **Delete Confirm** | Modal dialog | Key sequence | 🟡 Should improve |
| **Logs View** | Integrated | N/A | Maybe future |
| **Context Switch** | `:` with completion | ✅ Implemented | Keep! |
| **Describe** | `d` key | ❌ Missing | Consider |
| **Port Forward** | Built-in | N/A | Not applicable |
| **Sort/Filter** | Header clicks | ❌ Missing | Consider |

### What You're Doing Better Than k9s

1. **Simpler Model**
   - k9s has many views, yours is focused
   - Good for Operaton-specific workflow

2. **Breadcrumb Navigation**
   - Your breadcrumb design is clearer
   - Shows hierarchy well

3. **Environment Switching**
   - Your `e` key is simpler
   - k9s requires `:ctx` command

### k9s Patterns to Study

1. **Logo Placement**
   - k9s shows logo only in help, not in main view
   - Saves space (you do this in splash, good!)

2. **Context-Sensitive Help**
   - k9s shows different keys per view
   - Very helpful for complex apps

3. **Status Bar**
   - k9s shows cluster info, pod count, etc.
   - Your footer could show similar stats

---

## 10. Recommendations Summary 📋

### 🔴 Critical (Must Fix)

1. **Implement Help Screen (`?` key)**
   - Priority: URGENT
   - Effort: 4 hours
   - Impact: VERY HIGH
   - **Blocking adoption** - users can't discover features

2. **Fix Column Hiding (Responsive)**
   - Priority: HIGH
   - Effort: 6 hours
   - Impact: HIGH
   - Poor experience on narrow terminals

### 🟡 High Priority (Should Fix)

3. **Modal Confirmation Dialogs**
   - Replace `x`+`y` with modal
   - Priority: HIGH
   - Effort: 4 hours
   - Impact: HIGH
   - Better UX, safer actions

4. **Add Search/Filter (`/` key)**
   - Priority: MEDIUM-HIGH
   - Effort: 8 hours
   - Impact: HIGH
   - Essential for large datasets

5. **Loading Indicators**
   - Show spinner during API calls
   - Priority: MEDIUM
   - Effort: 2 hours
   - Impact: MEDIUM
   - Better perceived performance

### 🟢 Nice to Have

6. **Reduce Header to 3-4 Rows**
   - More content space
   - Effort: 2 hours
   - Impact: MEDIUM

7. **Add Response Time Display**
   - Show API performance
   - Effort: 2 hours
   - Impact: LOW-MEDIUM

8. **Better Error Messages**
   - More actionable errors
   - Effort: 4 hours (review all errors)
   - Impact: MEDIUM

9. **Context-Sensitive Keys**
   - Different keys per view
   - Effort: 6 hours
   - Impact: LOW-MEDIUM

10. **Pagination**
    - For large datasets
    - Effort: 8 hours
    - Impact: LOW (nice to have)

---

## 11. Implementation Roadmap 🛣️

### Sprint 1 Phase 2 (UX Polish - 1 Week)

**Day 1-2: Help Screen**
- Design help layout
- Implement `?` key handler
- Add context-sensitive help
- Test discoverability

**Day 3-4: Modal Confirmations**
- Design modal component
- Replace kill confirmation
- Add to other destructive actions
- Test usability

**Day 5: Responsive Columns**
- Implement priority system
- Add column hiding logic
- Test on various terminal sizes
- Update config schema

### Sprint 2 (Advanced Features - 1 Week)

**Day 1-2: Search/Filter**
- Design search UI
- Implement `/` key
- Add fuzzy matching
- Test on large datasets

**Day 3-4: Loading & Performance**
- Add loading spinners
- Show response times
- Add connection status
- Performance tuning

**Day 5: Polish**
- Error message review
- Header compaction
- Footer enhancements
- Final testing

---

## 12. Quick Wins (Do These First) ⚡

These are high-impact, low-effort improvements:

### 1. Fix Error Messages (2 hours)
```go
// Before
return fmt.Errorf("failed to fetch: %w", err)

// After
return fmt.Errorf("cannot load instances: %w. Press r to retry", err)
```

### 2. Add Success Feedback (2 hours)
```go
// After successful action
m.footerError = "✓ Instance killed successfully"
go func() {
    time.Sleep(2 * time.Second)
    // Clear message
}()
```

### 3. Show Item Count (1 hour)
```go
// In footer
footerText := fmt.Sprintf("<%s> | %d items", context, len(rows))
```

### 4. Better Input Cursor (1 hour)
```go
// In context selection
if m.showRootPopup {
    input := lipgloss.NewStyle().
        Foreground(lipgloss.Color("white")).
        Background(lipgloss.Color("blue")).
        Render(m.rootInput + "█")
}
```

---

## 13. Testing Checklist ✅

When implementing UX changes, test:

### Functionality
- [ ] All keys work as documented
- [ ] Help screen shows correct info
- [ ] Modal dialogs can be cancelled
- [ ] Search filters correctly
- [ ] Errors are actionable

### Responsiveness
- [ ] Works at 80x24 (minimum)
- [ ] Works at 120x40 (common)
- [ ] Works at 200x60 (large)
- [ ] Columns hide appropriately
- [ ] No horizontal scrolling

### Accessibility
- [ ] Works without color
- [ ] High contrast is readable
- [ ] Error states are clear
- [ ] Focus indicators visible

### Performance
- [ ] Loading large lists (1000+ items)
- [ ] Rapid key presses
- [ ] Terminal resize while loading
- [ ] Multiple environment switches

---

## Final Recommendations 🎯

Your o8n terminal UI has a solid foundation. The keyboard-driven workflow, clear hierarchy, and responsive design are well thought out. To elevate it to "excellent," focus on:

1. **🔴 Add help screen** - This is blocking adoption
2. **🟡 Implement modals** - Standard TUI pattern
3. **🟡 Fix column hiding** - Complete responsive design
4. **🟢 Add search** - Essential for large datasets

With these improvements, o8n will match or exceed k9s in UX quality for its specific domain.

**Estimated effort to reach "excellent":** 2-3 weeks

**Current Grade:** B+ (Good)  
**Potential Grade:** A (Excellent)

---

**End of UX Review**

Would you like me to:
- **A.** Design the help screen in detail (mockups + implementation plan)?
- **B.** Create modal dialog component design?
- **C.** Design the search/filter interface?
- **D.** Something else?

What would you like to tackle first? 🎨


