# o8n Quick Wins - Implementation Guide

**Goal:** 4 high-impact UX improvements that can be done in ~6-8 hours
**Result:** A-grade user experience with minimal code changes

---

## 🥇 Win #1: API Latency Display (30 mins)

**Current:**
```
[process-definitions] | Loading... | ⚡
```

**Better:**
```
[process-definitions] | Auto-refresh: ON ⚙ | ⚡ 42ms
```

### Code Changes

**File:** main.go

1. Add to model struct (~line 280):
```go
lastAPILatency time.Duration
apiCallStarted time.Time
```

2. When fetch starts (~line 1970+):
```go
m.apiCallStarted = time.Now()
```

3. When data arrives (~line 2450+):
```go
if !m.apiCallStarted.IsZero() {
    m.lastAPILatency = time.Since(m.apiCallStarted)
    m.apiCallStarted = time.Time{} // reset
}
```

4. In footer rendering (~line 2800+):
```go
latencyStr := ""
if m.lastAPILatency > 0 {
    latencyStr = fmt.Sprintf(" | ⚡ %dms", m.lastAPILatency.Milliseconds())
}
statusMessage = statusMessage + latencyStr
```

**Testing:**
```bash
go test ./... # Should still pass
./o8n --no-splash # Watch footer show latency after each fetch
```

**Impact:** ⭐⭐⭐⭐⭐ Users see system responsiveness; builds confidence

---

## 🥈 Win #2: Context-Aware Key Hints (45 mins)

**Current:** Same hints everywhere
**Better:** Hints change based on what you can do

### Code Changes

**File:** main.go

1. Add helper function (~line 670):
```go
func (m model) getContextKeyHints() []KeyHint {
    hints := []KeyHint{
        {Key: "?", Description: "Help", Priority: 1},
        {Key: "q", Description: "Quit", Priority: 1},
    }

    // Context-specific hints
    switch m.viewMode {
    case "definitions":
        hints = append(hints, KeyHint{Key: "e", Description: "Edit", Priority: 2})
        hints = append(hints, KeyHint{Key: ":", Description: "Switch View", Priority: 3})
    case "instances":
        hints = append(hints, KeyHint{Key: "e", Description: "Edit Variables", Priority: 2})
        hints = append(hints, KeyHint{Key: "ctrl+d", Description: "Terminate", Priority: 2})
    }

    if m.autoRefresh {
        hints = append(hints, KeyHint{Key: "r", Description: "Refresh: ON", Priority: 2})
    }

    return hints
}
```

2. Update header rendering (~line 670):
```go
// Change from getKeyHints() to:
hints := m.getContextKeyHints()
```

**Testing:**
```bash
go test ./...
./o8n --no-splash # Watch key hints change as you switch views
```

**Impact:** ⭐⭐⭐⭐ Users never wonder "what can I do here?"

---

## 🥉 Win #3: Inline Edit Validation (1 hour)

**Current:** Validation happens on Enter, user waits for feedback
**Better:** Show validation errors as they type

### Code Changes

**File:** main.go

1. Add validation display to editInput (~line 760):
```go
// In renderEditModal, after the input box:
validationMsg := ""
if m.editError != "" {
    validationMsg = errorFooterStyle.Render("⚠ " + m.editError)
}
```

2. Validate on each keystroke (~line 2100+, in Update method):
```go
case tea.KeyMsg:
    // ... existing code ...
    // After handling key input to editInput:
    m.editError = "" // clear previous error
    if m.activeModal == ModalEdit {
        // Get current field definition
        currentCol := m.editColumns[m.editColumnPos]
        inputVal := m.editInput.Value()

        // Validate immediately
        if _, err := validation.ValidateAndParse(
            currentCol.def.InputType,
            inputVal,
        ); err != nil {
            m.editError = err.Error()
        }
    }
```

3. Style the error box red:
```go
// In main.go, near style definitions:
validationErrorStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FF6B6B")).
    Bold(true)
```

**Testing:**
```bash
go test ./...
./o8n --no-splash
# Navigate to variables, press 'e' to edit
# Type invalid JSON - should see error immediately
```

**Impact:** ⭐⭐⭐⭐ Fewer failed API calls, faster workflow

---

## 🏆 Win #4: Pagination Status in Footer (30 mins)

**Current:** No indication of pagination
**Better:** Show "Page 1/3 (10 of 24 items)"

### Code Changes

**File:** main.go

1. Add to footer rendering (~line 2800):
```go
// After status message calculation:
paginationStr := ""
if total, ok := m.pageTotals[m.currentRoot]; ok && total > 0 {
    currentPage := (m.pageOffsets[m.currentRoot] / m.getPageSize()) + 1
    totalPages := (total + m.getPageSize() - 1) / m.getPageSize()
    visibleItems := len(m.table.Rows())
    paginationStr = fmt.Sprintf(" [%d/%d]", currentPage, totalPages)
}

// Append to status message before truncation:
statusMessage = statusMessage + paginationStr
```

2. Optional: Show navigation hints:
```go
// In getContextKeyHints(), add when in main view:
if m.currentRoot != "" && total > pageSize {
    hints = append(hints, KeyHint{
        Key: "PgDn/PgUp",
        Description: "Next/Prev Page",
        Priority: 2,
    })
}
```

**Testing:**
```bash
go test ./...
./o8n --no-splash
# Navigate to process-instances (8 items)
# With default page size, should show Page 1/1 or 1/2 depending on height
```

**Impact:** ⭐⭐⭐ Users always know where they are in dataset

---

## 📋 Implementation Checklist

### Phase 1: Latency Display (30 mins)
- [ ] Add model fields (`lastAPILatency`, `apiCallStarted`)
- [ ] Set `apiCallStarted` when fetch begins
- [ ] Calculate latency when data arrives
- [ ] Display in footer
- [ ] Test: latency appears and updates

### Phase 2: Context-Aware Hints (45 mins)
- [ ] Create `getContextKeyHints()` function
- [ ] Add view-specific hint logic
- [ ] Add auto-refresh state hint
- [ ] Update header rendering to use new function
- [ ] Test: hints change by view, by state

### Phase 3: Inline Validation (1 hour)
- [ ] Add `validationErrorStyle` to styles
- [ ] Add validation on keystroke in edit modal
- [ ] Display validation message inline
- [ ] Style error box
- [ ] Test: validation error appears immediately, clears on fix

### Phase 4: Pagination Status (30 mins)
- [ ] Add pagination display logic
- [ ] Calculate current page and total pages
- [ ] Format pagination string
- [ ] Add optional page navigation hints
- [ ] Test: page info displays correctly

---

## 🧪 Testing All 4 Changes

```bash
# Run tests
go test ./... -v

# Build binary
go build -o execs/o8n

# Manual testing workflow
./execs/o8n --no-splash

# Test scenarios:
# 1. Switch between views → see context hints change
# 2. Go to process-instances → see "Page 1/X" in footer
# 3. Press 'e' to edit → see latency in footer
# 4. Type invalid data → see validation error immediately
# 5. Press Ctrl+C → quit cleanly
```

---

## 📊 Expected Results

### Before
```
┌─────────────────────────────────────────┐
│ Header (3-4 rows)                       │
├─ Context box (1 row)                    │
├─ Key hints: Same everywhere             │
├─ Content (table)                        │
│ [process-definitions]  │  Loading...    │ ⚡
│                                          │
│ ID           KEY         VERSION         │
│ ─────────────────────────────────────── │
│ def1         invoice     2               │
│ def2         review      1               │
└─────────────────────────────────────────┘
```

### After
```
┌─────────────────────────────────────────┐
│ Header (3-4 rows)                       │
├─ Key hints: [e] Edit  [:] Switch  [?] Help
├─ Content (table)                        │
│ ID           KEY         VERSION         │
│ ─────────────────────────────────────── │
│ def1         invoice     2               │
│ def2         review      1               │
├─[process-definitions] | Page 1/1 | ⚡ 42ms
│ Validation: ✓ All fields valid           │
└─────────────────────────────────────────┘
```

---

## ⏱️ Time Breakdown

| Task | Time | Notes |
|------|------|-------|
| Win #1 (Latency) | 30m | Straightforward, low risk |
| Win #2 (Hints) | 45m | Small refactor, moderate risk |
| Win #3 (Validation) | 60m | Most complex, moderate risk |
| Win #4 (Pagination) | 30m | Easy, low risk |
| Testing | 30m | Manual testing and fixes |
| **Total** | **3 hours** | Can be done in one session |

---

## 🎓 Learning Outcomes

After implementing these changes, you'll understand:
1. ✅ How to track timing/performance metrics
2. ✅ How to make UI context-aware
3. ✅ How to provide real-time validation feedback
4. ✅ How to calculate and display pagination state
5. ✅ How to use the existing validation package
6. ✅ Bubble Tea event loop and message handling

---

## 🚀 Next Level (After Quick Wins)

Once these 4 are done:
1. **Search/Filter** (3-4 hours) - Major feature
2. **Column Customization** (2-3 hours) - Power user feature
3. **Streaming Indicator** (2-3 hours) - Performance feature
4. **Auto-Complete** (2-3 hours) - Productivity feature

See `ENHANCEMENT_PROPOSALS.md` for full details.

---

## 📞 Questions?

### "Will this break existing tests?"
**No.** These are UI-only changes. All tests should still pass. Model struct additions are backward-compatible.

### "Can I implement these in any order?"
**Yes.** They're independent. But Win #1 (Latency) is easiest to start with.

### "How much code is this really?"
**~100 lines total** across all changes:
- Latency: ~20 lines
- Context Hints: ~30 lines
- Validation: ~25 lines
- Pagination: ~20 lines

Most changes are in main.go's View() and Update() methods.

### "Will it impact performance?"
**No.** All changes are O(1) operations (no loops added).

