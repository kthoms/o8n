# Story: Layout Optimization

## Summary

Remove the ghost left pane allocation that wastes 25% of terminal width, eliminate the header spacer row to reclaim 1 content row, cap breadcrumb width to stabilize the footer on narrow terminals, align popup/box widths for visual consistency, and make the sort modal width dynamic.

## Motivation

The current layout has a `leftW = width / 4` allocation in the resize handler that sizes a `list.Model` which is never rendered. At a 120-char terminal, this steals 30 characters from the content table — a 36% width loss. Combined with 1-2 wasted spacer rows in the header, o8n shows significantly less data than it could.

After these fixes, o8n matches k9s's layout efficiency: 5 fixed overhead rows, full terminal width for content.

## Acceptance Criteria

### Fix 1: Remove ghost left pane (highest impact)

- [ ] **AC-1:** `m.paneWidth` equals `m.lastWidth` (full terminal width, no left column deduction)
- [ ] **AC-2:** The `leftW := width / 4` calculation is removed from the `WindowSizeMsg` handler
- [ ] **AC-3:** `m.list.SetSize(...)` call is removed (or the list model itself if unused elsewhere)
- [ ] **AC-4:** `m.table.SetWidth()` receives `m.lastWidth - 4` (full width minus border/padding)
- [ ] **AC-5:** Content box renders at full terminal width
- [ ] **AC-6:** All content (header, popup, content box, footer) aligns to the same full width

### Fix 2: Remove header spacer row

- [ ] **AC-7:** Header renders 2 content rows (status line + key hints), not 3
- [ ] **AC-8:** `lipgloss.Place` for header uses height 2 instead of 3
- [ ] **AC-9:** `headerLines` constant updated from 3 to 2 in all pane height calculations (model.go, nav.go, update.go)
- [ ] **AC-10:** Content table gains 1 additional visible row

### Fix 3: Cap breadcrumb width

- [ ] **AC-11:** Breadcrumb rendering in footer is capped at 50% of terminal width
- [ ] **AC-12:** When breadcrumb exceeds cap, crumb labels are truncated with `…` (rightmost/current crumb preserved in full, ancestors truncated first)
- [ ] **AC-13:** Status message column (center) is guaranteed a minimum of 20 characters
- [ ] **AC-14:** Footer does not visually break on narrow terminals (80 chars) with deep breadcrumbs

### Fix 4: Align popup and content box widths

- [ ] **AC-15:** Context popup width matches content box width exactly (same left/right margins)
- [ ] **AC-16:** Search bar width matches content box width exactly
- [ ] **AC-17:** No visual misalignment between popup borders and content box borders

### Fix 5: Dynamic sort modal width

- [ ] **AC-18:** Sort modal width adapts to the longest column name: `min(max(30, longestColumnName + 8), terminalWidth - 10)`
- [ ] **AC-19:** Sort modal never overflows the terminal width
- [ ] **AC-20:** Sort modal minimum width remains 30 characters

## Tasks

### Task 1: Remove ghost left pane allocation

**Files:** `update.go` (WindowSizeMsg handler)

1. Remove `leftW := width / 4` and the minimum clamp
2. Set `m.paneWidth = width` (or remove the field if it always equals `m.lastWidth`)
3. Remove `m.list.SetSize(leftW-2, contentHeight-1)` call
4. Update `m.table.SetWidth()` to use `width - 4` (accounting for box borders + padding)
5. Verify `m.list` model is not used elsewhere — if not, remove the field from the model struct
6. Update `renderBoxWithTitle` call to pass `m.lastWidth` as width

### Task 2: Remove header spacer row

**Files:** `view.go` (renderCompactHeader), `model.go`, `nav.go`, `update.go`

1. In `renderCompactHeader`: change `lipgloss.Place(m.lastWidth, 3, ...)` to `lipgloss.Place(m.lastWidth, 2, ...)`
2. In all pane height calculations, update `headerLines` from `3` to `2`:
   - `model.go:508` (initial setup)
   - `nav.go:270` (computePaneHeight)
   - `update.go` (WindowSizeMsg handler)
3. Test that content table gains 1 extra visible row

### Task 3: Cap breadcrumb width in footer

**Files:** `view.go` (footer rendering, ~line 644-753)

1. Calculate `maxBreadcrumbW := m.lastWidth / 2`
2. If rendered breadcrumb width exceeds `maxBreadcrumbW`, truncate ancestor crumb labels:
   - Start truncating from the leftmost (oldest) ancestor
   - Replace label text with first 8 chars + `…`
   - Keep current (rightmost) crumb label intact
3. Ensure `middleW >= 20` — if not, further truncate breadcrumb
4. Test with 4-level deep breadcrumb at 80-char terminal

### Task 4: Align popup and content box widths

**Files:** `view.go` (popup rendering, search bar rendering)

1. After Task 1, popup width should use the same base as content box
2. Replace `m.lastWidth - 4` with the same width used for `renderBoxWithTitle`
3. Verify visual alignment by testing at 80, 120, and 160 char widths
4. Ensure border characters line up vertically between popup and content box

### Task 5: Dynamic sort modal width

**Files:** `view.go` (sort modal rendering)

1. Before rendering sort modal, scan all column names in the current sort list
2. Calculate `longestName := max length of column display names`
3. Set modal width: `sortWidth := min(max(30, longestName + 8), m.lastWidth - 10)`
4. Apply to sort modal style

### Task 6: Update tests

1. Test pane width equals terminal width (no left column deduction)
2. Test pane height calculation with reduced header (2 rows instead of 3)
3. Test breadcrumb truncation at various terminal widths
4. Test sort modal width adapts to column name length
5. Verify all existing layout tests pass with updated constants

## Technical Notes

- The `m.list` (bubbles `list.Model`) appears to be a remnant from an earlier list-based view before the table model was adopted. Verify it's unused before removing.
- The header height of 3 was set via `lipgloss.Place(width, 3, ...)` — reducing to 2 means the header content (2 rows) fills exactly without padding. If visual spacing is needed, a border-bottom can provide separation without wasting a full row.
- Breadcrumb truncation should use `lipgloss.Width()` for ANSI-aware measurement, not `len()`, since crumb labels are styled with colors.
- After removing the ghost pane, the `paneWidth` field may become redundant (always equals `lastWidth`). Consider removing it to reduce model complexity, or keep it if future split-pane features are planned.

## Before / After

```
BEFORE (120-char terminal, 40 rows):
  Header: 3 rows + 1 spacer = 4 rows
  Ghost left pane: 30 chars wasted
  Table width: 82 chars
  Content rows: 32

AFTER (120-char terminal, 40 rows):
  Header: 2 rows = 2 rows
  Full width: no waste
  Table width: 116 chars
  Content rows: 35

  Net gain: +3 content rows, +34 chars table width
```
