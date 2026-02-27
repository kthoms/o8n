# Story: Task Complete Dialog UX Redesign

## Summary

Redesign the `ModalTaskComplete` dialog for better usability and clarity. The task name moves to the modal border title. The two-section INPUT/OUTPUT layout merges into a single unified variable list where read-only context variables and editable form fields coexist side-by-side. The content area becomes scrollable so all variables are reachable without hiding the buttons. Button focus styling switches from a layout-breaking border to an inverted-color highlight.

## Motivation

The current dialog has four compounding UX problems visible in the live screen:

1. The task name subtitle wraps across multiple lines inside the box, wasting vertical space and looking broken.
2. Every variable appears twice — once as a read-only input variable and again as an editable output field — because most form variables are pre-filled from matching process variables. This creates confusion about what the user is actually editing.
3. The `[Complete]` and `[Back]` buttons are not visible in the screenshot — the content overflows the terminal height, making the dialog unusable without knowing to scroll.
4. The focused button renders with an added border, which shifts surrounding layout and breaks visual alignment.

## Acceptance Criteria

### Border Title

- [ ] **AC-1:** The dialog border displays the task name as its title (Lipgloss `border.Title(...)`) instead of a subtitle line inside the box
- [ ] **AC-2:** The old `Complete Task` static title line and the task name subtitle line are removed from the dialog body
- [ ] **AC-3:** Long task names are truncated with `…` if they would exceed the border width

### Merged Variable List

- [ ] **AC-4:** The `INPUT VARIABLES` section, `OUTPUT VARIABLES` section, and the separator between them are removed
- [ ] **AC-5:** A single unified variable list is shown, sorted alphabetically by variable name
- [ ] **AC-6:** A variable present **only in input variables** (not a form field) is rendered read-only: `  name  :  value` in `FgMuted` style with no edit cursor
- [ ] **AC-7:** A variable present **in form variables** is rendered as an editable row: `  name  │ > <input.View()>` with the standard editable field style
- [ ] **AC-8:** If a form variable has a matching key in input variables it is pre-populated with the input variable's value (existing pre-fill behaviour preserved)
- [ ] **AC-9:** `Tab` and `Shift+Tab` skip read-only rows — only editable fields and the two buttons are in the focus cycle
- [ ] **AC-10:** If there are no variables at all (neither input nor form), show `  No variables` in `FgMuted` style
- [ ] **AC-11:** Inline validation errors (`⚠ message`) still appear below the affected editable field

### Scrollable Content / Pinned Buttons

- [ ] **AC-12:** The variable list area has a computed max height: `dialogHeight - fixedRows`, where `fixedRows` accounts for the border, separator, button row, and hint line
- [ ] **AC-13:** When the variable list exceeds `maxHeight` rows it is scrollable; a scroll offset is tracked in model state (`taskCompleteScrollOffset int`)
- [ ] **AC-14:** `↑` / `↓` arrow keys scroll the list one row at a time when focus is on a field (not on a button)
- [ ] **AC-15:** The focused editable field is always kept in view — scroll offset auto-adjusts when Tab moves focus to a field outside the visible window
- [ ] **AC-16:** The separator, button row (`[ Complete ]  [ Back ]`), and hint line are rendered **outside** the scrollable region and are always visible regardless of scroll position
- [ ] **AC-17:** A scroll indicator `↑` / `↓` / `↕` appears at the right edge of the separator when content is scrollable (↑ when not at top, ↓ when not at bottom, ↕ when both)

### Button Focus Styling

- [ ] **AC-18:** Focused buttons use **inverted colors** (accent background, dark foreground) with no border added — the button text is padded with one space each side: ` Complete `
- [ ] **AC-19:** Unfocused buttons render as `[ Complete ]` / `[ Back ]` in normal (dim) style — identical to current unfocused appearance
- [ ] **AC-20:** Disabled Complete button renders as `[ Complete ]` in `FgMuted` / `BtnDisabled` style — same as current
- [ ] **AC-21:** Button widths are fixed (not dependent on focus state) so the button row never shifts layout when focus changes

### Hint Line

- [ ] **AC-22:** Hint line reads: `Tab: next field  ↑↓: scroll  Space: toggle bool  Esc: back`

## Tasks

### Task 1: Model state — scroll offset

**Files:** `internal/app/model.go`

1. Add `taskCompleteScrollOffset int` field to the `model` struct
2. Reset `taskCompleteScrollOffset` to `0` in `closeTaskCompleteDialog()`

### Task 2: View — merged list and new layout

**Files:** `internal/app/view.go`

1. Replace `renderTaskCompleteModal` with a new implementation:
   - **Border title:** use `lipgloss.NewStyle().Border(...).Title(truncated(m.taskCompleteTaskName, width-4))` — remove old title/subtitle lines from body
   - **Build unified row slice:** iterate alphabetically over the union of input var keys and form field names; for each:
     - If the name is a form field key → `editableRow{name, field}`
     - Else → `readOnlyRow{name, value}` (muted, no tab stop)
   - **Compute visible window:** `maxVisible = dialogHeight - fixedRows`; slice `unifiedRows[scrollOffset : scrollOffset+maxVisible]`
   - **Render rows:** editable rows as `  name  │ > <input.View()>` + optional `  ⚠ error` line; read-only rows as `  name  :  value` in `FgMuted`
   - **Scroll indicator:** append `↑`, `↓`, or `↕` to the separator line when content overflows
   - **Pinned footer:** separator + button row + hint line rendered after the scrollable block, outside the clipped region
   - **Buttons:** focused = `lipgloss.NewStyle().Background(accentColor).Foreground(darkColor).Render(" Complete ")` ; unfocused = `[ Complete ]`; disabled = `BtnDisabled` style

2. Remove `sortStrings` helper if it is no longer used elsewhere (it was only used by the old two-section layout)

### Task 3: Key handling — scroll and tab skip

**Files:** `internal/app/update.go`

1. In the `ModalTaskComplete` key handler, add `↑` / `↓` cases:
   - `↑`: decrement `taskCompleteScrollOffset` (min 0)
   - `↓`: increment `taskCompleteScrollOffset` (max `len(unifiedRows) - maxVisible`, if positive)
2. Modify `taskCompleteTabForward` / `taskCompleteTabBackward` to skip read-only rows when advancing through `focusTaskField` positions — only editable field indices are valid tab stops
3. After Tab moves focus, call a `taskCompleteEnsureVisible()` helper that adjusts `taskCompleteScrollOffset` so the newly focused field is within the visible window
4. Add `taskCompleteEnsureVisible(visibleHeight int)` helper method

### Task 4: Tests

**Files:** `internal/app/claim_complete_task_test.go`

Update or add:

1. **Merged list rendering:** `renderTaskCompleteModal` output contains no `INPUT VARIABLES` or `OUTPUT VARIABLES` heading text
2. **Read-only row style:** a pure input variable (not a form field) appears as `name  :  value` and is not a tab stop
3. **Tab skips read-only:** Tab from field[0] skips to field[1] (not to a read-only row) when a read-only row sits between them
4. **Scroll offset:** when `taskCompleteScrollOffset > 0` and the list has more rows than the visible height, the first row in the render output is not the first row of the unified list
5. **Ensure visible:** after Tab moves to a field outside the visible window, `taskCompleteScrollOffset` adjusts so the field is visible
6. **Button focus style:** focused Complete button render does NOT contain `[` or `]` characters; unfocused does
7. **Border title:** `renderTaskCompleteModal` output contains the task name in the border, not as a body line

## Dev Notes

### Unified row model

The simplest implementation builds a `[]unifiedRow` slice (a local struct or pair of slices used only inside `renderTaskCompleteModal`). Each entry carries: `name string`, `isEditable bool`, `fieldIdx int` (index into `taskCompleteFields` if editable, -1 if read-only), `value string` (for read-only display).

Tab navigation needs to know which `fieldIdx` values exist and in what order — derive a `tabStops []int` slice (field indices of editable rows in display order) and use it in `taskCompleteTabForward/Backward` instead of iterating `taskCompleteFields` directly.

### Lipgloss border title

```go
style := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(accentColor).
    Title(" " + truncated + " ").
    TitleAlign(lipgloss.Left).
    Width(dialogWidth - 2).
    Padding(1, 2)
```

The `Title()` method is available in Lipgloss v0.10+. Check `go.mod` to confirm the version — if `Title()` is unavailable, render the title manually by replacing the top border string.

### Scroll region

Compute `fixedRows` as: `1 (separator) + 1 (blank) + 1 (button row) + 1 (blank) + 1 (hint line)` = 5 rows below the variable list, plus `1 (top padding)` above = 6 total fixed rows inside the border. `maxVisible = innerHeight - 6`.

### Button fixed width

Pick a fixed button width (e.g. 12 chars including padding) and apply it to both focused and unfocused variants using `lipgloss.NewStyle().Width(12).Align(lipgloss.Center)`. This ensures the button row width is stable regardless of focus state.

## File List

- `internal/app/model.go` — add `taskCompleteScrollOffset`
- `internal/app/update.go` — scroll key handlers, tab-skip read-only, `taskCompleteEnsureVisible`
- `internal/app/view.go` — new `renderTaskCompleteModal`, remove old two-section layout, button style
- `internal/app/claim_complete_task_test.go` — update/add tests per Task 4

## Dev Agent Record

**Agent:** Amelia (dev)
**Date:** 2026-02-27

### Implementation Notes

- Manual rounded border rendering used (╭╮╰╯ built by hand) instead of Lipgloss `.Border()` so the task title can be injected into the top border line before ANSI codes are applied.
- `sortStrings` helper retained in view.go — still used by the unified row list builder.
- `taskCompleteMaxVisible()` formula: `dialogH - 9` (where `dialogH = lastHeight - 4`), matching `contentH - 7` used in the view.
- Button widths: `completeW = 12`, `backW = 8` — stable regardless of focus state via `centerPad()`.
- `taskCompleteEnsureVisible()` computes the focused field's virtual row index using the same `sort.Strings` order as the view to guarantee alignment.

### Files Modified

- `internal/app/model.go` — added `taskCompleteScrollOffset int`
- `internal/app/view.go` — replaced `renderTaskCompleteModal`, added `centerPad` helper
- `internal/app/update.go` — scroll key handlers, `closeTaskCompleteDialog` reset, `taskCompleteEnsureVisible` calls in tab methods, 4 new helper methods
- `internal/app/claim_complete_task_test.go` — updated `TestRenderTaskCompleteModal`, added 7 new tests

## Change Log

- 2026-02-27: Story created (UX redesign — border title, merged list, scroll, button style)
- 2026-02-27: Implementation complete — all 4 tasks done, all tests pass

## Status

done
