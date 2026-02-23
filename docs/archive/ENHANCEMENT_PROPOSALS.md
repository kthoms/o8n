# o8n Terminal UI Enhancement Proposals

**Date:** 2026-02-20
**Based On:** Existing UX Review + Critical Audit + Claude Code UI Patterns
**Goal:** Elevate o8n from B+ to A-grade TUI experience

---

## 🎯 Strategic Vision

o8n should feel like the **"Claude Code for Operaton"** — a professional, responsive, keyboard-first terminal UI that makes managing workflow instances enjoyable and efficient.

### Inspiration from Claude Code UI
- **Streaming Response Handling** → Real-time data updates with visual spinners
- **Context Display** → Current state/operation always visible
- **Smart Defaults** → Minimal configuration needed
- **Rich Status Indicators** → Animated loading, progress bars
- **Keyboard Shortcuts Help** → Always available, context-aware
- **Error Messages** → Helpful, actionable feedback

---

## 🔥 High-Impact Quick Wins (1-2 hours each)

### 1. Real-time Auto-Refresh Indicator
**Status:** Partially implemented (flash indicator exists, but incomplete UX)

**Enhancement:**
```
Current: [breadcrumb] | Loading... | ⚡
Better:  [breadcrumb] | Auto-refresh: ON ⟳  | ⚡ API: 145ms
         (shows refresh status + API latency)
```

**Implementation:**
- Track API call latency (started vs completed)
- Display latency in footer when available
- Show refresh toggle state clearly
- Animate the ⟳ symbol when data is being fetched

**Files to Update:**
- main.go: footer rendering (line 2800+)
- model struct: add `lastAPILatency time.Duration`
- Flash indicator: make it animated ⟳ during fetch

**Impact:** Users know if auto-refresh is ON and get confidence about API health

---

### 2. Streaming Indicator for Large Datasets
**Status:** Not implemented

**Enhancement:**
When fetching 100+ items, show progress:
```
Loading process instances (23/100)... ⟳
Loading process instances (45/100)... ⟳
Loading process instances (100/100) ✓
```

**Implementation:**
- Add `loadProgress struct { current, total int }` to model
- Create new message type: `dataProgressMsg`
- Fetch in batches and report progress
- Render progress bar or percentage in footer

**Example CLI Tool:** kubectl shows `Retrieving workloads... 234/512`

**Impact:** No more wondering "is it frozen?" — shows responsiveness

---

### 3. Inline Data Validation Feedback
**Status:** Partially implemented (edit modal has validation)

**Enhancement:**
Show validation errors immediately as user types:
```
Variable Name: my_var ✓
Variable Value: [invalid json
                 ↑ Error: Missing closing brace
Variable Type: JSON
```

**Implementation:**
- Add real-time validation display in edit modal
- Color code field validity (red border = invalid)
- Show helpful error message inline
- Use existing `validation` package

**Files to Update:**
- main.go: renderEditModal() (~line 770)
- textinput model: add validation state display

**Impact:** Users fix errors faster, fewer failed API calls

---

### 4. Context-Sensitive Key Hint Bar
**Status:** Implemented but could be smarter

**Enhancement:**
Show only relevant keys for current context:
```
At process-definitions level:
[e] Edit  [r] Refresh  [:] Switch View  [?] Help  [q] Quit

At variables level (editable):
[e] Edit Variable  [Tab] Next Field  [Enter] Save  [Esc] Cancel  [?] Help

When in edit modal:
[Enter] Save  [Esc] Cancel  [Tab] Next Column  [Shift+Tab] Prev  [?] Help
```

**Implementation:**
- Add `contextKeyHints []KeyHint` to model
- Update getKeyHints() to be context-aware
- Show hints based on `activeModal` + `viewMode`
- Maximum 6 hints per line (fits 80+ char terminals)

**Files to Update:**
- main.go: renderKeyHints() (~line 670)
- Add context detection logic

**Impact:** Users never wonder "what can I do here?" — hints change with context

---

### 5. Auto-Complete Suggestions in Edit Modal
**Status:** User suggestions for some fields exist, but no general search/filter

**Enhancement:**
When editing common fields, show suggestions:
```
Variable Name: process_   [▼ Suggestions]
               ├─ process_id
               ├─ process_key
               ├─ process_status
               └─ process_priority

Variable Value: [use contentassist package]
```

**Implementation:**
- Extend `contentassist` package to support process field names
- Add `suggestions []string` to edit modal
- Arrow keys navigate suggestions, Tab/Enter select
- Already have user name suggestions, extend pattern

**Files to Update:**
- renderEditModal(): add suggestion box (20 lines)
- Integrate with existing contentassist API

**Impact:** Faster data entry, fewer typos, discoverability of valid values

---

## 💎 Medium-Effort Polish (2-4 hours each)

### 6. Search/Filter Functionality
**Status:** Not implemented (planned in specification)

**Enhancement:**
Press `/` to search current view:
```
Search process definitions for:
[in_voice                          ]  ← fuzzy search
Match 1 of 3:
├─ ReviewInvoice (matches "invoice")
├─ ApproveInvoice (matches "invoice")  ← highlighted
└─ PaymentInvoice (matches "invoice")

Navigation: ↑↓ Select  Enter Drill Down  Esc Cancel
```

**Implementation:**
- New modal type: `ModalSearch`
- Implement fuzzy search (use existing libraries or simple string matching)
- Filter table rows in real-time
- Save last search per view

**Files to Update:**
- main.go: Add ModalSearch case, implement search logic (~100 lines)
- model struct: add `searchQuery string`, `searchResults []int`

**Key Insights from k9s:**
- k9s shows "Nodes(5) ↔ Search" when search is active
- Search survives pagination
- Can combine with filters

**Impact:** Users can find instances fast without scrolling

---

### 7. Saved Views / Quick Filters
**Status:** Not implemented

**Enhancement:**
Save frequently-used filters:
```
Saved Views:
[1] Active Invoices (status=ACTIVE)
[2] Pending Approval (status=PENDING)
[3] Failed (state=ERROR)
[4] Today's Processes
```

**Implementation:**
- Add `savedViews map[string]View` to config
- Quick access with `Ctrl+1`, `Ctrl+2`, etc.
- Display in header or context menu
- Save to o8n-cfg.yaml

**Files to Update:**
- config.go: add SavedView type and persistence
- main.go: add view switching logic

**Impact:** Power users can jump to common scenarios instantly

---

### 8. Table Column Customization
**Status:** Configured in YAML but not interactive

**Enhancement:**
Interactive column visibility toggle:
```
Press 'c' to customize columns:
[ ✓ ] ID
[ ✓ ] Key
[ ✓ ] Version
[   ] Category       (hidden by default, 3 more options...)
[ ✓ ] TenantID
[   ] DeploymentID

Use arrows to navigate, Space to toggle, Enter to apply
```

**Implementation:**
- New modal: `ModalColumnCustomization`
- Read/write column visibility config
- Save to o8n-cfg.yaml for persistence
- Remember per table

**Files to Update:**
- main.go: Add column customization modal and logic (~80 lines)
- config.go: already has column definitions

**Impact:** Users can focus on the data they care about

---

### 9. Pagination Info Display
**Status:** Footer shows counts (fixed in critical audit), but pagination is hidden

**Enhancement:**
Show pagination status clearly:
```
Viewing 10 of 24 items (Page 1 of 3)
[◀ Prev]  [▶ Next]  [Jump to page...]
```

Or in footer:
```
[process-instances] | Page 1/3 (10 of 24) ▶ | ⚡
```

**Implementation:**
- Calculate total pages from pageTotals
- Display in footer or header
- Allow direct jump to page via modal
- Show items per page (configurable)

**Files to Update:**
- main.go: footer rendering (~10 lines)
- Add pagination modal if needed

**Impact:** Users always know where they are in large datasets

---

## 🚀 Advanced Features (4+ hours each)

### 10. Process Visualizer
**Status:** Not implemented (would require Graphviz or custom rendering)

**Enhancement:**
Visualize workflow process when drilling into an instance:
```
┌─── reviewInvoice (v2) ───┐
│                          │
│  ○ Review Document       │
│   │                      │
│   ├→ ○ Approve [10]     │  ← 10 instances in this step
│   │   │                  │
│   │   └→ ○ Payment       │
│   │                      │
│   └→ ○ Reject            │
│       │                  │
│       └→ ○ Close         │
│                          │
└──────────────────────────┘
```

**Implementation:**
- Query process definition XML/BPMN if available via API
- Render simple ASCII diagram
- Highlight current instance's step
- Show count of instances in each step

**Files Needed:**
- internal/visualization/diagram.go (new)
- main.go: add diagram rendering to view

**Impact:** Users understand workflow structure without external tools

---

### 11. Historical Trend Display
**Status:** Not implemented

**Enhancement:**
Show instance count trends:
```
invoice instances per hour (last 24h):
│                    ╱╲
│                ╱╲╱╲╱ ╲
│            ╱╲╱╲╱╲ ╲   ╲
│        ╱╲╱╲╱╲    ╲ ╲   ╲
│    ╱╲╱╲╱╲╱      ╲ ╲   ╲
│╱╲╱╲╱           ╲ ╲   ╲
└─────────────────────── Now

Min: 2 | Avg: 18 | Max: 34 | Trending: ↑
```

**Implementation:**
- Fetch historical data from API (if available)
- Render ASCII sparkline
- Cache data for performance
- Update periodically

**Impact:** Operators get insights into system load patterns

---

### 12. Multi-View Dashboard
**Status:** Not implemented

**Enhancement:**
Split screen view showing multiple resources simultaneously:
```
┌─ Definitions (left 40%) ──┬─ Recent Instances (right 60%) ──┐
│ reviewInvoice      ► │    │ inst-001: RUNNING      15m ago    │
│ paymentProcess    ► │    │ inst-042: COMPLETED    23m ago    │
│ approvalFlow      ► │    │ inst-033: ERROR        2h ago    │
│                     │    │ inst-005: RUNNING      3h ago    │
│                     │    │ ...                               │
└─────────────────────┴──────────────────────────────────────┘
```

**Implementation:**
- Toggle with `\` key or modal
- Synchronized navigation (select definition, see its instances)
- Configurable split ratio (30/70, 50/50, etc.)
- Remember user preference

**Files to Update:**
- main.go: major refactor to support multiple tables
- model struct: add `dashboardMode bool`, `splitRatio int`

**Impact:** Power users can see relationships without drilling down

---

## 📊 Implementation Priority Matrix

| Feature | Impact | Effort | Priority | Timeline |
|---------|--------|--------|----------|----------|
| Real-time Latency Display | High | 1-2h | 🔴 Week 1 | Sprint 2a |
| Streaming Indicator | Medium | 2-3h | 🟡 Week 2 | Sprint 2b |
| Inline Validation | High | 1-2h | 🔴 Week 1 | Sprint 2a |
| Context-Sensitive Hints | High | 2h | 🔴 Week 1 | Sprint 2a |
| Auto-Complete Suggestions | Medium | 2-3h | 🟡 Week 2 | Sprint 2b |
| Search/Filter | High | 3-4h | 🔴 Week 2 | Sprint 2b |
| Saved Views | Medium | 2-3h | 🟡 Week 3 | Sprint 2c |
| Column Customization | Medium | 2-3h | 🟡 Week 3 | Sprint 2c |
| Pagination Info | Low | 1h | 🟢 Quick | Sprint 2a |
| Process Visualizer | Low | 4-5h | 🟢 Sprint 3 | Sprint 3 |
| Historical Trends | Low | 3-4h | 🟢 Sprint 3 | Sprint 3 |
| Multi-View Dashboard | Low | 5-6h | 🟢 Sprint 3 | Sprint 3 |

---

## 🎨 Design System Updates

### Color Palette Enhancements
Current: Single accent color per environment

**Proposed:**
- 🔴 Error/Critical: Red (#FF6B6B)
- 🟢 Success/Healthy: Green (#50C878)
- 🟡 Warning/Pending: Yellow (#FFD700)
- 🔵 Info/Loading: Blue (#00A8E1)
- ⚪ Neutral/Disabled: Gray (#999999)

Already partially implemented! (See AUDIT.md)

### Animation Principles
- **Loading:** Slow rotation (⟳ → \\ → - → /)
- **Transitions:** 100-200ms slide/fade
- **Feedback:** Instant (key press → visual change)
- **Emphasis:** Brief highlight/color flash on success

### Typography & Spacing
- Keep monospace font (standard for TUI)
- Maintain 2-space indentation
- Use box-drawing characters for clarity
- Consistent 1-space padding inside boxes

---

## 📋 Claude Code UI Patterns to Adopt

### 1. **Smart Defaults**
✅ Environment auto-detection: "local" or first available
❌ Could add: Remember last view, remember last environment switch

### 2. **Contextual Help**
✅ Help screen exists (been implemented since UX review)
❌ Could add: Inline tips for first-time users, tutorial mode

### 3. **Status Indicators**
✅ Loading state exists
❌ Could add: Animation during loading, progress percentage

### 4. **Streaming UX**
❌ Not implemented
✅ Recommended: Show data as it arrives (partial results while loading)

### 5. **Error Recovery**
✅ Error messages display and auto-clear
❌ Could add: "Retry" button for failed API calls

### 6. **Keyboard Mastery**
✅ Vim-like navigation (hjkl alternative not yet added)
✅ Single-key commands (e, r, q)
❌ Could add: Chord keys (Ctrl+A = select all, etc.)

### 7. **Theme Consistency**
✅ Color themes in `/skins` directory
✅ Environment-specific colors
❌ Could add: High-contrast mode for accessibility

---

## 🧪 Testing Strategy for Enhancements

### Unit Tests
- Validation logic (already good)
- Search/filter functionality
- Pagination calculations

### Integration Tests
- API calls with progress reporting
- Multi-view synchronization
- Config persistence (saved views)

### E2E Tests (Manual)
- Full workflows with new features
- Terminal resize handling
- Keyboard shortcuts with new features

---

## 📝 Next Steps

### Week 1 (Quick Wins)
1. [ ] Add API latency display to footer
2. [ ] Implement inline validation feedback in edit modal
3. [ ] Make key hint bar context-aware
4. [ ] Verify pagination counts work correctly (already done in audit)

### Week 2 (Medium Features)
1. [ ] Implement streaming indicator for large datasets
2. [ ] Build search/filter functionality
3. [ ] Add auto-complete suggestions for variable names

### Week 3+ (Polish & Advanced)
1. [ ] Column visibility customization
2. [ ] Saved views / quick filters
3. [ ] Process visualizer (if BPMN available)
4. [ ] Dashboard split-view mode

---

## 💡 Ideas for Future Sprints

- **Voice Commands:** Use terminal speech recognition (experimental)
- **Vim Mode Toggle:** hjkl + :w :q commands
- **Theme Generator:** Auto-generate themes from terminal colors
- **Plugin System:** Allow custom command handlers
- **Webhook Notifications:** Alert on critical events
- **API Recording/Replay:** Debug mode showing all API calls

---

## 🎯 Success Metrics

| Metric | Current | Target | Timeline |
|--------|---------|--------|----------|
| Time to Find Instance | ~10s (scroll) | <2s (search) | Sprint 2b |
| User Errors (typos) | 15% | <5% (auto-complete) | Sprint 2a |
| Abandoned Sessions | Unknown | Track with analytics | Sprint 3 |
| Feature Discovery Rate | Low | 80%+ (better help) | Sprint 2a |
| Keyboard Efficiency | Good | Excellent | Sprint 2c |

---

## 🙏 Acknowledgments

- UX Review by BMM UX Designer (Feb 15, 2026)
- Specification documents (layout, modal, help screen designs)
- Claude Code UI patterns (streaming, status, context)
- k9s project (keyboard-first TUI patterns)

