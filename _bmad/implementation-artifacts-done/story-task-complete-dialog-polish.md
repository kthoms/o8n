# Story: Task Complete Dialog Polish

## Summary

Four follow-up fixes for `ModalTaskComplete` identified from live usage:
remove the redundant hint line, make dialog height content-driven (eliminates blank rows and makes the bottom border visible), surface API errors inside the dialog (they were invisible behind the overlay), and allow Space to activate focused buttons.

## Motivation

Live screenshot (`debug/screen.txt`) of the "Assign Reviewer" task shows:

1. **Hint line** — `Tab: next field  ↑↓: scroll  Space: toggle bool  Esc: back` appears below the buttons and wastes a row; users already know these keys from the help screen.
2. **Blank rows and missing bottom border** — `dialogH = height - 4` always allocates maximum terminal height. With 5 fields the dialog needs ~13 rows but gets 25+, resulting in 15+ blank rows in the scroll region and the bottom `╰╯` border pushed off-screen.
3. **API errors invisible** — When the Complete button triggers an API call that fails, `errMsg` sets the footer error, but the dialog overlay covers the footer completely. The user sees no feedback and concludes the button is broken.
4. **Space doesn't activate buttons** — Space only toggles bool fields; it does nothing when a button (`[ Complete ]` / `[ Back ]`) has focus, which is inconsistent with common TUI conventions.

## Acceptance Criteria

### Remove Hint Line

- [x] **AC-1:** The hint line (`Tab: next field  ↑↓: scroll  Space: toggle bool  Esc: back`) is removed from the dialog body.
- [x] **AC-2:** The fixed-row count used for height calculation is updated to reflect the removed line (5 fixed content lines → 4: `blank + sep + blank + buttons`).

### Content-Driven Height

- [x] **AC-3:** `dialogH` is derived from the actual row count: `totalRows + 8` (2 border + 1 top-blank + totalRows + 1 blank + 1 sep + 1 blank + 1 buttons), where `totalRows = len(unifiedRows)` before any scroll clipping.
- [x] **AC-4:** `dialogH` is capped at `height - 4` (never exceeds terminal) and floored at `10 + errorRows` (minimum usable size with error line accounting).
- [x] **AC-5:** When `totalRows` exceeds the available space (`height - 12`), the dialog reaches max height and the scroll region activates — same scrolling behaviour as before, no regression.
- [x] **AC-6:** For the 5-field "Assign Reviewer" case the dialog is approximately 13 rows tall, shows no blank padding rows, and the bottom `╰╯` border is fully visible.
- [x] **AC-7:** `taskCompleteMaxVisible()` in `update.go` is kept consistent with the new formula so scroll key handling remains correct.

### Error Display Inside Dialog

- [x] **AC-8:** The model gains a `taskCompleteError string` field (reset to `""` in `closeTaskCompleteDialog()` and on each new dialog open via `taskVariablesLoadedMsg`).
- [x] **AC-9:** When `errMsg` is received and `m.activeModal == ModalTaskComplete`, the error text is stored in `m.taskCompleteError` (in addition to the normal footer handling). The footer still shows the error for consistency.
- [x] **AC-10:** `renderTaskCompleteModal` renders `taskCompleteError` as a single red/muted line below the button row when non-empty: `  ⚠ <error text>` in `ValidationError` style. When empty the line is not rendered (no blank placeholder).
- [x] **AC-11:** `taskCompleteError` is cleared (set to `""`) when the user makes any edit to a field (on the next `tea.KeyMsg` that reaches a field) OR when the dialog is re-opened.

### Space Activates Focused Buttons

- [x] **AC-12:** In the `ModalTaskComplete` key handler, the `" "` / `"space"` case is extended: when `taskCompleteFocus == focusTaskComplete` it calls `submitTaskComplete()` (same as Enter); when `taskCompleteFocus == focusTaskBack` it calls `closeTaskCompleteDialog()` (same as Enter).

## Tasks

### Task 1: Model — new field

**Files:** `internal/app/model.go`

1. [x] Add `taskCompleteError string` to the `model` struct.

### Task 2: Update — error capture, Space on buttons, scroll formula sync

**Files:** `internal/app/update.go`

1. [x] In `closeTaskCompleteDialog()`, add `m.taskCompleteError = ""`.
2. [x] In `taskVariablesLoadedMsg` handler, add `m.taskCompleteError = ""`.
3. [x] In `errMsg` handler: add a branch — if `m.activeModal == ModalTaskComplete`, set `m.taskCompleteError = friendlyError(m.currentEnv, msg.err)`.
4. [x] In the `ModalTaskComplete` key handler, extend the `" "` / `"space"` case:
   ```go
   case " ", "space":
       switch m.taskCompleteFocus {
       case focusTaskComplete:
           if cmd := m.submitTaskComplete(); cmd != nil {
               return m, cmd
           }
       case focusTaskBack:
           m.closeTaskCompleteDialog()
       case focusTaskField:
           // existing bool toggle logic
       }
       return m, nil
   ```
5. [x] In the default key case (printable keystroke fed to field), clear `m.taskCompleteError = ""` before feeding the key.
6. [x] Update `taskCompleteMaxVisible()` to use the new formula: `maxVis = dialogH - 2 - 1 - 4` where `dialogH` is derived as in Task 3 below, to stay consistent with the view.
   - Simpler: `maxVis = min(totalRows, height - 12)` capped at `≥ 3`. Where `totalRows = m.taskCompleteTotalRows()` and `height = m.lastHeight`.

### Task 3: View — remove hint, content height, error line

**Files:** `internal/app/view.go`

1. [x] **Remove hint line:** deleted `contentLines = append(contentLines, m.styles.FgMuted.Render("Tab: next field..."))` and the blank line before it.
2. [x] **Update fixed-row comment and `maxVisible` formula:**
   - Fixed lines inside border (below scroll region): `blank(1) + sep(1) + blank(1) + buttons(1)` = **4**
   - Plus blank top(1) = **5** total fixed content lines
   - `maxVisible = contentH - 5 - errorRows` (was `contentH - 7`)
3. [x] **Content-driven `dialogH`:**
   - Row-list build moved to BEFORE the `dialogH` calculation so `totalRows = len(rows)` is available.
   - `dialogH = totalRows + 8 + errorRows`  — 2 borders + 1 top-blank + totalRows + 4 footer lines + optional error line
   - Cap: `if dialogH > height-4 { dialogH = height-4 }`
   - Floor: `if dialogH < 10+errorRows { dialogH = 10+errorRows }` (ensures `maxVisible >= 3`)
   - Recomputed `contentH = dialogH - 2` and `maxVisible = contentH - 5 - errorRows`.
4. [x] **Error line:** after the button row, if `m.taskCompleteError != ""` append:
   ```go
   contentLines = append(contentLines, m.styles.ValidationError.Render("  ⚠ "+m.taskCompleteError))
   ```
   Error line is included in `dialogH` via `errorRows` so it is never clipped.

### Task 4: Tests

**Files:** `internal/app/claim_complete_task_test.go`

Update or add:

1. **No hint line:** `renderTaskCompleteModal` output does NOT contain `"Tab: next field"`.
2. **Content-driven height:** for 3 fields + 0 input vars, the rendered output has fewer lines than `height - 4`; specifically `dialogH = 3 + 8 = 11` lines (count newlines).
3. **Error line:** when `taskCompleteError` is set, `renderTaskCompleteModal` output contains `"⚠"` and the error text.
4. **No error line when empty:** when `taskCompleteError` is `""`, output does NOT contain `"⚠"`.
5. **Space on Complete:** when `taskCompleteFocus == focusTaskComplete`, sending `" "` key dispatches `completeTaskCmd` (cmd non-nil) — same as Enter.
6. **Space on Back:** when `taskCompleteFocus == focusTaskBack`, sending `" "` closes the dialog (`activeModal == ModalNone`).
7. **Error cleared on field edit:** after setting `taskCompleteError = "some error"`, sending a printable key while `focusTaskField` clears it to `""`.
8. **Existing `TestRenderTaskCompleteModal`** — verify it still passes (no hint line check needed since we removed the hint).

## Dev Notes

### Height formula derivation

```
dialogH  = 2 (top/bottom border)
         + 1 (blank top padding)
         + totalRows
         + 1 (blank before separator)
         + 1 (separator)
         + 1 (blank before buttons)
         + 1 (buttons)
         = totalRows + 8
```

`maxVisible` inside the content area:
```
contentH = dialogH - 2  = totalRows + 6
maxVisible = contentH - 5  (5 fixed content lines)
           = totalRows + 6 - 5
           = totalRows + 1   ← but rows capped at totalRows, so effectively all rows fit
```

When `dialogH` would exceed terminal:
```
cap at height - 4
→ contentH = height - 6
→ maxVisible = height - 11   (same floor ≥ 3)
```

### taskCompleteMaxVisible sync

The update.go helper `taskCompleteMaxVisible()` must return the same value that the view uses. The simplest sync: compute it as `min(m.taskCompleteTotalRows(), m.lastHeight - 12)` with a floor of 3. This mirrors the capped case above.

### Error line and border height

The error line is appended to `contentLines` AFTER the buttons. The border rendering loop iterates `contentH` lines and pads/truncates. Since the error line is an extra line beyond `contentH`, it will naturally be clipped if the terminal is very small — acceptable behaviour.

Alternatively (safer): include the error line in `contentH` by adding 1 to `dialogH` when `taskCompleteError != ""`. This avoids clipping. The simpler approach (let it clip) is acceptable given the error is also shown in the footer.

## File List

- `internal/app/model.go` — add `taskCompleteError string`
- `internal/app/update.go` — error capture, Space on buttons, field-edit error clear, `taskCompleteMaxVisible` sync
- `internal/app/view.go` — remove hint line, content-driven height, error line rendering
- `internal/app/claim_complete_task_test.go` — new and updated tests per Task 4

## Dev Agent Record

**Agent:** Amelia (bmm-dev)
**Date:** 2026-02-27
**Status:** All 4 tasks complete + code review fixes applied, all tests passing.

### Implementation Notes

- Introduced `errorRows` variable (0 or 1) in `view.go` to account for the optional error line in both `dialogH` and `maxVisible`. This avoids clipping when the error line is shown.
- Floor for `dialogH` changed from `9` to `10 + errorRows` to guarantee `maxVisible >= 3` in the minimum case.
- Scroll tests updated to use 10 total rows (was 8) — at height=20, need `totalRows >= 9` for `totalRows+8 > height-4` (cap to trigger), giving `maxOffset = 1`.
- `taskCompleteMaxVisible()` in `update.go` mirrors the view's capped formula: `min(total, height-11)` with floor 3. The `errorRows` offset is not needed in `update.go` because that helper only governs scroll bounds, not rendering.

### Tests Added

7 new tests in `claim_complete_task_test.go`:
- `TestNoHintLineInModal`
- `TestContentDrivenHeight`
- `TestErrorLineShownInDialog`
- `TestNoErrorLineWhenErrorEmpty`
- `TestSpaceActivatesCompleteButton`
- `TestSpaceActivatesBackButton`
- `TestErrorClearedOnFieldEdit`

2 existing scroll tests updated: `TestScrollOffsetChangesVisibleRows`, `TestEnsureVisibleScrollsToFocusedField`.

### Code Review Fixes (2026-02-27)

**M-1 fixed** — `taskCompleteMaxVisible()` now accounts for `errorRows`:
- Capping threshold changed from `total+8 > height-4` to `total+8+errorRows > height-4`
- Capped value changed from `height-11` to `height-11-errorRows`
- Stale comment updated (was "floor = 9"; now documents actual floors)
- Regression test added: `TestScrollMaxVisibleAccountsForErrorRow`

## Change Log

- 2026-02-27: Story created (polish — hint removal, content height, error in dialog, Space on buttons)
- 2026-02-27: Implementation complete — all ACs satisfied, all tests passing
- 2026-02-27: Code review — 1 MEDIUM fixed (taskCompleteMaxVisible errorRows), 1 LOW fixed (stale comment), regression test added

## Status

done
