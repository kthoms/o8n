# 🎯 Modal Confirmation Design - Complete Specification

**Feature:** Modal Confirmation Dialogs  
**Priority:** 🟡 HIGH  
**Effort:** 6 hours  
**Status:** Design Complete, Ready for Implementation  
**Designer:** BMM UX Designer  
**Date:** February 16, 2026

---

## Executive Summary

Modal confirmations are **critical for preventing accidental destructive actions** in o8n. The current two-key sequence (`x` + `y`) lacks visual feedback and is not discoverable. This design implements proper modal dialogs following TUI best practices from k9s, lazygit, and other successful terminal applications.

**Key Principles:**
- ✅ **Clear intent** - User knows exactly what will happen
- ✅ **Explicit confirmation** - Requires deliberate action, not accidental key press
- ✅ **Escapable** - Easy to cancel without consequence
- ✅ **Informative** - Shows relevant context and warnings
- ✅ **Consistent** - Same pattern across all destructive actions

---

## 1. When to Use Modals

### ✅ Required Confirmations (Destructive Actions)

| Action | Trigger Key | Modal Type | Reason |
|--------|-------------|------------|--------|
| Delete/Kill Instance | `<ctrl>+d` | Destructive | Cannot be undone, terminates running process |
| Delete Deployment | `<ctrl>+d` | Destructive | Removes workflow from engine |
| Quit Application | `<ctrl>+c` | Confirmation | Exits application (safe, but explicit intent) |
| Batch Delete | `<ctrl>+d` | Destructive | Multiple items affected |
| Modify Variable (future) | `<ctrl>+e` | Warning | Changes process state |

### ❌ No Modal Needed

| Action | Trigger Key | Why No Modal |
|--------|-------------|--------------|
| View drill-down | `Enter` | Navigation only, non-destructive |
| Back navigation | `Esc` | Standard navigation pattern |
| Refresh | `<ctrl>-r` | Safe, read-only operation |
| Toggle auto-refresh | `<ctrl>-R` | Non-destructive, reversible |
| Switch environment | `<ctrl>+e` | Safe context switch (could add warning if unsaved changes) |
| Switch view | `:` | Navigation only |

---

## 2. Visual Design - Modal Types

### 2.1 Destructive Confirmation Modal

**Use Case:** Delete instance, delete deployment, kill process

```
┌─ Process Definitions ────────────────────────────────────────────────────────┐
│                                                                               │
│ KEY        │ NAME             │ VERSION │ INSTANCES │ DEPLOYED               │
│ ──────────────────────────────────────────────────────────────────────────   │
│ invoice-v2 │ Invoice Review   │ 2       │ 15        │ 2026-02-15 14:30       │
│ ▶ order-v3 │ Order Processing │ 3       │ 7         │ 2026-02-16 09:15       │ ← Selected
│ shipping   │ Ship Product     │ 1       │ 0         │ 2026-02-10 11:00       │
│                                                                               │
│                ┌─ ⚠️  DELETE PROCESS INSTANCE ───────────────┐              │
│                │                                              │              │
│                │  You are about to DELETE this instance:     │              │
│                │                                              │              │
│                │  Instance ID:   order-v3-inst-abc123        │              │
│                │  Definition:    Order Processing v3         │              │
│                │  Status:        ACTIVE (running)            │              │
│                │  Started:       2026-02-16 10:45:32         │              │
│                │                                              │              │
│                │  ⚠️  WARNING: This action CANNOT be undone! │              │
│                │  The instance will be terminated.           │              │
│                │                                              │              │
│                │  <ctrl>+d  Confirm Delete    Esc  Cancel    │              │
│                │                                              │              │
│                └──────────────────────────────────────────────┘              │
│                                                                               │
│ Definitions > Order Processing v3 > Instances | local @ localhost:8080      │
└───────────────────────────────────────────────────────────────────────────────┘
```

**Key Features:**
1. **⚠️ Warning icon** - Immediate visual alert
2. **Modal title** - Clear action name in CAPS
3. **Context information** - Shows what will be deleted
4. **Warning message** - Explicit consequence statement
5. **Key hints** - Shows available actions at bottom
6. **Semi-transparent overlay** - Content behind is slightly visible (context preserved)

**Dimensions:**
- Width: 54 chars (centered)
- Height: 13 rows
- Position: Vertically centered
- Border: Double line (`═`, `║`) for emphasis

---

### 2.2 Simple Confirmation Modal (Less Critical)

**Use Case:** Quit application, switch environment with active refresh

```
┌─ Process Definitions ────────────────────────────────────────────────────────┐
│                                                                               │
│ KEY        │ NAME             │ VERSION │ INSTANCES │ DEPLOYED               │
│ ──────────────────────────────────────────────────────────────────────────   │
│ invoice-v2 │ Invoice Review   │ 2       │ 15        │ 2026-02-15 14:30       │
│                                                                               │
│                   ┌─ Confirm Quit ───────────────┐                           │
│                   │                               │                           │
│                   │  Quit o8n application?        │                           │
│                   │                               │                           │
│                   │  <ctrl>+c  Yes    Esc  No     │                           │
│                   │                               │                           │
│                   └───────────────────────────────┘                           │
│                                                                               │
│ Definitions > Order Processing v3 > Instances | local @ localhost:8080      │
└───────────────────────────────────────────────────────────────────────────────┘
```

**Key Features:**
1. **Smaller size** - 35 chars wide, 7 rows tall
2. **Simple message** - Direct question
3. **Binary choice** - Yes/No with clear shortcuts
4. **Single line border** - Less emphasis than destructive modal

---

### 2.3 Warning Modal (Non-destructive but Important)

**Use Case:** Switching environment, disabling auto-refresh with active monitoring

```
┌─ Process Definitions ────────────────────────────────────────────────────────┐
│                                                                               │
│                ┌─ ⓘ  Switch Environment ──────────────────┐                 │
│                │                                           │                 │
│                │  Switch from 'local' to 'dev'?            │                 │
│                │                                           │                 │
│                │  Current:  local @ localhost:8080         │                 │
│                │  Target:   dev @ dev.operaton.example.com │                 │
│                │                                           │                 │
│                │  Auto-refresh will be disabled.           │                 │
│                │                                           │                 │
│                │  <ctrl>+e  Switch    Esc  Cancel          │                 │
│                │                                           │                 │
│                └───────────────────────────────────────────┘                 │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

**Key Features:**
1. **ⓘ Info icon** - Informational tone (not warning)
2. **Context details** - Shows before/after state
3. **Side effects** - Lists what else will happen
4. **Optional** - Could be skippable with setting

---

## 3. Interaction Flow

### 3.1 Destructive Action Flow (Delete Instance)

```
State 1: Normal View
┌─────────────────────────────┐
│ User browsing instances     │
│ Arrow keys to navigate      │
│ Instance selected (▶)       │
└─────────────────────────────┘
                ↓
         User presses <ctrl>+d
                ↓
State 2: Modal Appears
┌─────────────────────────────┐
│ Modal overlays content      │
│ Background dimmed/visible   │
│ Shows instance details      │
│ Warning message visible     │
└─────────────────────────────┘
                ↓
    ┌─────────────────────┐
    │                     │
    ↓                     ↓
Press <ctrl>+d       Press Esc
(confirm delete)     (any time)
    ↓                     ↓
State 4a: Deleting    State 4b: Cancelled
┌─────────────────┐   ┌──────────────────┐
│ Modal shows:    │   │ Modal disappears │
│ "Deleting..."   │   │ Return to list   │
│ Spinner active  │   │ Selection intact │
│                 │   │ Footer message:  │
└─────────────────┘   │ "Cancelled"      │
    ↓                 └──────────────────┘
API call executes
    ↓
    ┌────────────────┐
    │                │
    ↓                ↓
Success          Failure
    ↓                ↓
State 5a:        State 5b:
┌──────────────┐ ┌──────────────────┐
│ Modal closes │ │ Error modal      │
│ Item removed │ │ Shows API error  │
│ Footer msg:  │ │ "Retry" option   │
│ "Deleted ✓"  │ │ or "Cancel"      │
└──────────────┘ └──────────────────┘
```

**Key Decision Points:**
1. **Entry:** `<ctrl>+d` key press (mode-aware)
2. **Display:** Show instance details and warning
3. **Confirmation:** Press `<ctrl>+d` again to confirm
4. **Cancellation:** `Esc` at any time (no side effects)
5. **Feedback:** Loading state + result message

**Note:** The two-step `<ctrl>+d` process (trigger → confirm) provides safety without tedious typing.

---

### 3.2 Simple Confirmation Flow (Quit)

```
User presses <ctrl>+c
    ↓
Modal appears: "Quit o8n application?"
    ↓
    ┌─────────────────────┐
    │                     │
    ↓                     ↓
Press <ctrl>+c       Press Esc
    ↓                     ↓
Quit immediately     Cancel, return to app
```

**Simplified:** No input validation, binary choice only

---

## 4. Keyboard Interactions

### 4.1 Modal-Specific Keys

| Key | State | Action | Notes |
|-----|-------|--------|-------|
| `Esc` | Any | Cancel and close modal | Always available |
| `<ctrl>+d` | Modal displayed | Confirm destructive action | Same key confirms |
| `<ctrl>+c` | Simple confirm | Confirm and proceed | Context-dependent |
| Any other key | Modal displayed | Ignored | Prevents accidental triggers |

### 4.2 Key Behavior Changes When Modal Active

**Background Interaction:**
- ❌ **Blocked:** All list navigation keys (↑/↓, PgUp/Dn)
- ❌ **Blocked:** View switching (`:`, `e`, `r`)
- ❌ **Blocked:** Other actions (`d`, `v`, `s`)
- ✅ **Allowed:** `Esc` to cancel
- ✅ **Allowed:** Confirmation keys (`<ctrl>+d`, `<ctrl>+c`)

**Focus Management:**
- Modal captures all keyboard input
- Tab focus cycles within modal (if multiple inputs)
- No focus visible on background content

---

## 5. Visual States & Styling

### 5.1 Modal Color Themes (Environment-Aware)

Modals should adapt to the environment color scheme:

```go
// Modal styles based on environment
type ModalStyle struct {
    BorderColor    string // Environment UI color
    TitleBG        string // Slightly lighter than border
    WarningColor   string // Red (#FF5733)
    InfoColor      string // Blue (#00A8E1)
    SuccessColor   string // Green (#00FF00)
    ErrorColor     string // Red (#FF0000)
    InputBG        string // Dark gray (#2D2D2D)
    InputFocusBG   string // Lighter gray (#404040)
    InputValidBG   string // Green tint (#1A4D2E)
}
```

**Example Rendering:**

```
// Local environment (blue theme)
BorderColor: #00A8E1 (cyan/blue)
┌─ ⚠️  DELETE... ─┐  ← Title bar in blue

// Dev environment (orange theme)
BorderColor: #FFA500 (orange)
┌─ ⚠️  DELETE... ─┐  ← Title bar in orange

// Prod environment (red theme)
BorderColor: #FF5733 (red)
┌─ ⚠️  DELETE... ─┐  ← Title bar in red
```

---

### 5.2 Background Dimming

**Technique:** Overlay with semi-transparent effect

```
Before modal:
┌─────────────────────────────┐
│ KEY      │ NAME    │ VERSION│  ← Full brightness
│ invoice  │ Review  │ 2      │
│ ▶ order  │ Process │ 3      │  ← Selected (highlighted)
└─────────────────────────────┘

With modal:
┌─────────────────────────────┐
│ KEY      │ NAME    │ VERSION│  ← 40% dimmed
│ invoice  │ Review  │ 2      │  ← Gray overlay
│ ▶ order  │ Process │ 3      │  ← Still visible but muted
│     ┌─ Modal Here ─┐        │  ← Full brightness
│     │               │        │
│     └───────────────┘        │
└─────────────────────────────┘
```

**Implementation:** Apply `lipgloss.Faint(true)` to background content

---

## 6. Modal Copy & Messaging

### 6.1 Tone & Voice Guidelines

**Destructive Actions:**
- ⚠️ Start with warning icon
- Use ALL CAPS for critical words (DELETE, CANNOT BE UNDONE)
- State consequences explicitly
- Be direct, not cute

**Good Examples:**
```
✅ "You are about to DELETE this instance."
✅ "This action CANNOT be undone."
✅ "The instance will be terminated immediately."
```

**Bad Examples:**
```
❌ "Are you sure?" (too vague)
❌ "This will delete the thing." (unclear)
❌ "Careful! This is permanent-ish." (unprofessional)
```

---

### 6.2 Title Patterns

| Action Type | Title Format | Example |
|-------------|--------------|---------|
| Destructive | `⚠️  [VERB] [NOUN]` | `⚠️  DELETE PROCESS INSTANCE` |
| Confirmation | `Confirm [Action]` | `Confirm Quit` |
| Warning | `ⓘ  [Context]` | `ⓘ  Switch Environment` |
| Error | `❌ [Error Type]` | `❌ Delete Failed` |
| Success | `✓ [Action] Complete` | `✓ Instance Deleted` |

---

### 6.3 Button Labels

**Standard Patterns:**

```
Destructive:
  <ctrl>+d  Confirm Delete    Esc  Cancel

Confirmation:
  <ctrl>+c  Yes    Esc  No

Warning:
  <ctrl>+e  Switch    Esc  Cancel

Error (with retry):
  <ctrl>+r  Retry    Esc  Close
```

**Rules:**
- Primary action first (left side)
- Cancel/escape always on right
- Show actual key combinations
- Use verb for primary action

---

## 7. Accessibility Considerations

### 7.1 Color-Independent Design

**Don't Rely on Color Alone:**

❌ **Bad:** Red text only for errors
```
Wrong ID  ← Just red, no other indicator
```

✅ **Good:** Icon + color + position
```
✗ Wrong ID  ← Icon + red + error styling
```

**State Indicators:**

| State | Visual | Symbol | Color |
|-------|--------|--------|-------|
| Valid | Border highlight | `✓` | Green |
| Invalid | Dashed border | `✗` | Red |
| Warning | Bold border | `⚠️` | Yellow |
| Info | Thin border | `ⓘ` | Blue |

---

### 7.2 Screen Reader Friendly

**Semantic Structure:**
```
Modal:
- Title: "Warning: Delete Process Instance"
- Content: "You are about to delete instance ABC123..."
- Input: "Type instance ID to confirm, currently empty"
- Button 1: "Confirm Delete, Control+D"
- Button 2: "Cancel, Escape"
```

**Implementation Notes:**
- Use clear, linear layout (top to bottom)
- Logical reading order
- Action buttons last (natural tab order)

---

## 8. Error Handling

### 8.1 API Error Modal

When deletion fails:

```
┌─ ❌ Delete Failed ────────────────────────────┐
│                                               │
│  Could not delete instance:                   │
│  order-v3-inst-abc123                         │
│                                               │
│  Error: 404 Not Found                         │
│  The instance may have already been deleted.  │
│                                               │
│  <ctrl>+r  Retry    Esc  Close                │
│                                               │
└───────────────────────────────────────────────┘
```

**Features:**
1. Clear error title with ❌ icon
2. Shows what was attempted (context)
3. Displays actual API error
4. Provides human-friendly interpretation
5. Offers retry option (if applicable)

---

### 8.2 Network Error

```
┌─ ❌ Network Error ────────────────────┐
│                                       │
│  Cannot reach Operaton server:        │
│  http://localhost:8080                │
│                                       │
│  Error: Connection refused            │
│                                       │
│  Check that the server is running.    │
│                                       │
│  <ctrl>+r  Retry    Esc  Close        │
│                                       │
└───────────────────────────────────────┘
```

---

### 8.3 Validation Error

```
┌─ ⚠️  Invalid Input ───────────────────┐
│                                       │
│  Instance ID must match exactly:      │
│  order-v3-inst-abc123                 │
│                                       │
│  You entered:                         │
│  order-inst-abc123                    │
│                  ↑                    │
│              Missing "v3"             │
│                                       │
│  Esc  Try Again                       │
│                                       │
└───────────────────────────────────────┘
```

---

## 9. Implementation Guide

### 9.1 Component Structure

```go
// Modal types
type ModalType int

const (
    ModalNone ModalType = iota
    ModalDestructiveConfirm
    ModalSimpleConfirm
    ModalWarning
    ModalError
    ModalInfo
)

// Modal state
type Modal struct {
    Type          ModalType
    Title         string
    Message       string
    Icon          string          // "⚠️", "ⓘ", "✓", "❌"
    
    // Context
    ContextLines  []string       // Key-value pairs to show
    
    // Actions
    PrimaryKey    string         // "<ctrl>+d"
    PrimaryLabel  string         // "Confirm Delete"
    SecondaryKey  string         // "Esc"
    SecondaryLabel string        // "Cancel"
    
    // Callbacks
    OnConfirm     func() error
    OnCancel      func()
    
    // State
    IsLoading     bool
    LoadingMsg    string
    Error         error
}

// Add to main model
type model struct {
    // ... existing fields ...
    
    modal *Modal
    showModal bool
}
```

---

### 9.2 Update Function Changes

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Priority 1: Check if modal is active
    if m.showModal && m.modal != nil {
        return m.handleModalInput(msg)
    }
    
    // Priority 2: Regular input handling
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+d":
            // Open delete confirmation modal
            return m.showDeleteModal()
            
        case "ctrl+c":
            // Open quit confirmation modal
            return m.showQuitModal()
            
        // ... other keys
        }
    }
    
    return m, nil
}

// Modal input handler
func (m model) handleModalInput(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "esc":
            // Always cancel on Esc
            m.showModal = false
            if m.modal.OnCancel != nil {
                m.modal.OnCancel()
            }
            return m, nil
            
        case "ctrl+d":
            // Destructive confirmation
            if m.modal.Type == ModalDestructiveConfirm {
                return m.executeModalAction()
            }
            
        case "ctrl+c":
            // Simple confirmation
            if m.modal.Type == ModalSimpleConfirm {
                return m.executeModalAction()
            }
            
        default:
            // Ignore other keys to prevent accidental actions
            return m, nil
        }
    }
    
    return m, nil
}
```

---

### 9.3 View Function

```go
func (m model) View() string {
    var b strings.Builder
    
    // Render main content (dimmed if modal active)
    mainContent := m.renderMainView()
    if m.showModal {
        mainContent = lipgloss.NewStyle().
            Foreground(lipgloss.Color("240")).  // Dim it
            Render(mainContent)
    }
    
    b.WriteString(mainContent)
    
    // Overlay modal if active
    if m.showModal && m.modal != nil {
        modalView := m.renderModal()
        // Position modal at center
        overlayed := m.overlayModal(mainContent, modalView)
        return overlayed
    }
    
    return b.String()
}

func (m model) renderModal() string {
    modal := m.modal
    
    // Build modal box
    var content strings.Builder
    
    // Title with icon
    title := fmt.Sprintf("%s  %s", modal.Icon, modal.Title)
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color(m.uiColor)).
        Width(56).
        Align(lipgloss.Center)
    
    content.WriteString(titleStyle.Render(title))
    content.WriteString("\n\n")
    
    // Message
    content.WriteString(modal.Message)
    content.WriteString("\n\n")
    
    // Context lines (if any)
    for _, line := range modal.ContextLines {
        content.WriteString("  " + line + "\n")
    }
    content.WriteString("\n")
    
    // Action buttons
    primaryStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color(m.uiColor))
    
    actions := fmt.Sprintf("  %s  %s    %s  %s",
        modal.PrimaryKey, modal.PrimaryLabel,
        modal.SecondaryKey, modal.SecondaryLabel,
    )
    content.WriteString(actions)
    
    // Wrap in bordered box
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.DoubleBorder()).
        BorderForeground(lipgloss.Color(m.uiColor)).
        Padding(1, 2).
        Width(60)
    
    return boxStyle.Render(content.String())
}
```

---

### 9.4 Helper Functions

```go
// Show delete confirmation modal
func (m model) showDeleteModal() (model, tea.Cmd) {
    selectedItem := m.getCurrentSelectedItem()
    
    m.modal = &Modal{
        Type:         ModalDestructiveConfirm,
        Title:        "DELETE PROCESS INSTANCE",
        Icon:         "⚠️",
        Message:      "You are about to DELETE this instance:",
        ContextLines: []string{
            fmt.Sprintf("Instance ID:   %s", selectedItem.ID),
            fmt.Sprintf("Definition:    %s v%d", selectedItem.Name, selectedItem.Version),
            fmt.Sprintf("Status:        %s", selectedItem.Status),
            "",
            "⚠️  WARNING: This action CANNOT be undone!",
            "The instance will be terminated immediately.",
        },
        PrimaryKey:     "<ctrl>+d",
        PrimaryLabel:   "Confirm Delete",
        SecondaryKey:   "Esc",
        SecondaryLabel: "Cancel",
        OnConfirm: func() error {
            return m.deleteInstance(selectedItem.ID)
        },
    }
    
    m.showModal = true
    return m, nil
}

// Show quit confirmation modal
func (m model) showQuitModal() (model, tea.Cmd) {
    m.modal = &Modal{
        Type:           ModalSimpleConfirm,
        Title:          "Confirm Quit",
        Icon:           "",
        Message:        "Quit o8n application?",
        PrimaryKey:     "<ctrl>+c",
        PrimaryLabel:   "Yes",
        SecondaryKey:   "Esc",
        SecondaryLabel: "No",
        OnConfirm: func() error {
            return nil // Will quit in Update
        },
    }
    
    m.showModal = true
    return m, nil
}
```

---

## 10. Testing Checklist

### 10.1 Visual Testing

- [ ] Modal centers correctly on different terminal sizes
- [ ] Background dims appropriately
- [ ] Border colors match environment theme
- [ ] Modal doesn't overflow on small terminals (80x24)
- [ ] Text wraps correctly within modal
- [ ] Input field visible and properly styled
- [ ] Icons render correctly (⚠️, ✓, ❌, ⓘ)

### 10.2 Interaction Testing

- [ ] `Esc` cancels modal at any time
- [ ] Background navigation blocked when modal open
- [ ] `<ctrl>+d` confirms immediately (no input required)
- [ ] Other keys ignored to prevent accidents
- [ ] Modal closes after successful action
- [ ] Error modal shows on API failure
- [ ] Retry works after error
- [ ] Loading state displays during API call

### 10.3 Edge Cases

- [ ] Very long instance IDs in display (truncation)
- [ ] Network timeout during delete
- [ ] Delete already-deleted instance (404)
- [ ] Permission denied (403)
- [ ] Rapid key presses (debounce `<ctrl>+d`)
- [ ] Terminal resize while modal open
- [ ] Modal displayed while auto-refresh active

### 10.4 Accessibility

- [ ] Color-blind users: icons not color-only
- [ ] Logical tab order
- [ ] Clear visual focus indicator
- [ ] Consistent keyboard shortcuts
- [ ] Clear error messages
- [ ] Timeout warnings (if applicable)

---

## 11. Future Enhancements

### Phase 2 Features

1. **Batch Confirmations**
   - Multi-select in list
   - Modal shows count: "Delete 5 instances?"
   - Option to review list before confirming

2. **Confirmation History**
   - "Undo" for recent deletions
   - Modal: "Instance deleted 2s ago - Undo?"
   - Limited time window (30s)

3. **Customizable Confirmation Level**
   - Settings: "Always confirm", "Confirm destructive only", "Never confirm"
   - Environment-specific settings (e.g., prod always confirms)

4. **Smart Defaults**
   - Option to skip confirmation for certain actions
   - Remember "Don't ask again" preference
   - Reset in settings

5. **Modal Stacking**
   - Multiple modals (e.g., error during confirmation)
   - Breadcrumb for modal stack
   - Each Esc closes one modal

---

## 12. Success Metrics

### User Experience Goals

| Metric | Target | Measurement |
|--------|--------|-------------|
| Accidental deletions | < 1% | Track delete cancellations vs completions |
| Time to confirm | < 2 seconds | Measure modal open to confirm/cancel |
| Confirmation success rate | > 99% | All confirms work (no input errors) |
| Modal discoverability | > 90% | Survey: "Did you notice confirmation prompt?" |
| User confidence | > 4/5 rating | Survey: "How confident deleting instances?" |

### Technical Goals

| Metric | Target | Notes |
|--------|--------|-------|
| Modal render time | < 50ms | No visible lag |
| Input responsiveness | < 10ms | Validation real-time |
| Memory overhead | < 1MB | Modal state minimal |
| No UI flicker | 0% | Smooth overlay |

---

## Appendix A: Complete Interaction Matrix

| Current View | Action Key | Modal Type | Confirmation Key | Notes |
|--------------|------------|------------|------------------|-------|
| Process Definitions | `<ctrl>+d` | Destructive | `<ctrl>+d` | Same key confirms |
| Process Instances | `<ctrl>+d` | Destructive | `<ctrl>+d` | Same key confirms |
| Variables | N/A | N/A | N/A | Read-only |
| Any view | `<ctrl>+c` | Simple Confirm | `<ctrl>+c` | Quick yes/no |
| Any view | `<ctrl>+e` | Warning | `<ctrl>+e` | Environment switch |

---

## Appendix B: Copy Template Library

### Destructive Delete - Instance
```
Title: ⚠️  DELETE PROCESS INSTANCE
Message: You are about to DELETE this instance:
Context:
  Instance ID:   {id}
  Definition:    {name} v{version}
  Status:        {status}
  Started:       {timestamp}
  
  ⚠️  WARNING: This action CANNOT be undone!
  The instance will be terminated immediately.
  
Actions: <ctrl>+d  Confirm Delete    Esc  Cancel
```

### Destructive Delete - Deployment
```
Title: ⚠️  DELETE DEPLOYMENT
Message: You are about to DELETE this deployment:
Context:
  Deployment ID:     {id}
  Process Def:       {name} v{version}
  Deployment Time:   {timestamp}
  Active Instances:  {count}
  
  ⚠️  WARNING: This does NOT delete running instances!
  Only the deployment record will be removed.
  
Actions: <ctrl>+d  Confirm Delete    Esc  Cancel
```

### Simple Confirm - Quit
```
Title: Confirm Quit
Message: Quit o8n application?
Actions: <ctrl>+c  Yes    Esc  No
```

### Warning - Switch Environment
```
Title: ⓘ  Switch Environment
Message: Switch from '{current}' to '{target}'?
Context:
  Current:  {current} @ {current_url}
  Target:   {target} @ {target_url}
  
  Auto-refresh will be disabled.
Actions: <ctrl>+e  Switch    Esc  Cancel
```

### Error - Delete Failed
```
Title: ❌ Delete Failed
Message: Could not delete instance:
Context:
  {instance_id}
  
  Error: {api_error_message}
  {human_friendly_explanation}
Actions: <ctrl>+r  Retry    Esc  Close
```

---

**End of Specification**

Ready for implementation! 🚀
