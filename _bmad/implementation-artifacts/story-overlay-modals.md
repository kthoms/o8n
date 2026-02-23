# Story: Overlay Modals, Generic Confirm Dialog, Remove j/k

**Status**: in-progress
**Priority**: High

---

## Story

As a user of o8n, I want all dialogs to appear as overlays over the current view so that I retain context while interacting with modals. Confirmation dialogs should be visually consistent, have opaque Tab-navigatable buttons, and the application should not support vim-style j/k navigation (arrow keys only).

---

## Acceptance Criteria

- **AC1**: All modal dialogs (confirm-delete, confirm-quit, help, sort, detail, env, actions) render as overlays with the background UI visible.
- **AC2**: A generic `confirmDialog` struct drives all confirmation modals — title, message, confirm label, cancel label, confirm key, confirm action.
- **AC3**: Confirm dialog renders two opaque styled buttons (Confirm / Cancel) that can be navigated with Tab and activated with Enter or the bound key.
- **AC4**: The active/focused button is visually distinct (focused border via `BtnSaveFocused` / `BtnCancelFocused` styles).
- **AC5**: `j` and `k` are no longer navigation keys for table scrolling — arrow keys only.
- **AC6**: The help screen text and README keyboard shortcuts section reflect all changes.
- **AC7**: No regressions — all existing tests pass.

---

## Tasks / Subtasks

- [ ] **T1**: Make all non-overlay modals use `overlayCenter`
  - [ ] T1a: `renderHelpScreen` — wrap result with `overlayCenter(baseView, ...)`
  - [ ] T1b: `renderSortPopup` — wrap with `overlayCenter`
  - [ ] T1c: `renderDetailView` — wrap with `overlayCenter`
  - [ ] T1d: `renderEnvPopup` — wrap with `overlayCenter`
  - [ ] T1e: `renderActionsMenu` — wrap with `overlayCenter`

- [ ] **T2**: Generic `confirmDialog` struct + rendering
  - [ ] T2a: Define `confirmDialog` struct in `model.go` with fields: `title`, `message`, `confirmLabel`, `cancelLabel`, `confirmKey`, `focusedBtn` (0=confirm, 1=cancel)
  - [ ] T2b: Add `m.confirmDialog *confirmDialog` field to `model` struct
  - [ ] T2c: Implement `renderConfirmDialog(bg string) string` method that renders opaque buttons and overlays onto `bg`
  - [ ] T2d: Replace `renderConfirmDeleteModal` and `renderConfirmQuitModal` with `confirmDialog` instances in `update.go`
  - [ ] T2e: Handle Tab key to toggle `focusedBtn` inside confirm dialog
  - [ ] T2f: Handle Enter to activate focused button inside confirm dialog
  - [ ] T2g: Handle Esc to cancel confirm dialog

- [ ] **T3**: Remove j/k navigation bindings
  - [ ] T3a: Remove `"j"` and `"k"` standalone cases from all key handlers in `update.go`
  - [ ] T3b: Remove `"j"` and `"k"` from combined `"j", "down"` / `"k", "up"` cases (keep arrow keys)
  - [ ] T3c: Verify `"j"` still works as text input in popup/search contexts (it should — those are caught earlier)

- [ ] **T4**: Update help text and README
  - [ ] T4a: Update `renderHelpContentForLineCount` and actual help content in `view.go` — remove j/k references
  - [ ] T4b: Update README keyboard shortcuts section

- [ ] **T5**: Tests
  - [ ] T5a: `TestConfirmDialogTabTogglesFocus` — Tab toggles between confirm/cancel
  - [ ] T5b: `TestConfirmDialogEnterActivatesConfirm` — Enter on focused=0 triggers confirm action
  - [ ] T5c: `TestConfirmDialogEnterActivatesCancel` — Enter on focused=1 triggers cancel
  - [ ] T5d: `TestConfirmDialogEscCancels` — Esc clears the dialog
  - [ ] T5e: `TestJKeyNoLongerScrollsTable` — j key does not move table cursor
  - [ ] T5f: `TestKKeyNoLongerScrollsTable` — k key does not move table cursor
  - [ ] T5g: `TestArrowDownStillScrollsTable` — down arrow still works

---

## Dev Notes

### Architecture

- `confirmDialog` is a value-type struct stored as `m.confirmDialog *confirmDialog` (nil = no dialog active)
- Opening a confirm dialog: set `m.confirmDialog = &confirmDialog{...}` — do NOT change `m.activeModal`
- `confirmDialog` replaces `ModalConfirmDelete` and `ModalConfirmQuit` — those `ModalType` values can remain but routing goes through confirmDialog
- In `View()`, if `m.confirmDialog != nil`, call `m.renderConfirmDialog(baseView)` as the final overlay (after all other modals)
- Actually, simpler: keep `ModalConfirmDelete` and `ModalConfirmQuit`, but their render functions now use the generic `confirmDialog` render logic, passing the base view in.

### Button styling

- Focused button: `m.styles.BtnSaveFocused` (confirm) or `m.styles.BtnCancelFocused` (cancel)
- Unfocused button: `m.styles.BtnSave` / `m.styles.BtnCancel`
- Buttons side-by-side with spacing: `confirmBtn + "   " + cancelBtn`
- Default focus = confirm button (index 0)

### Overlay requirement

All modals that currently call `lipgloss.Place(width, height, Center, Center, modal)` as their top-level return need to instead return the modal box only, and the calling code in `View()` wraps with `overlayCenter(baseView, box)`. This matches the pattern already used for `renderConfirmDeleteModal`.

### Key changes in update.go

Current combined cases like `case "j", "down", "ctrl+d":` become `case "down", "ctrl+d":`. Standalone `case "j":` and `case "k":` sections are removed entirely.

### Confirm dialog Tab + Enter flow

```
Tab pressed → toggle m.confirmDialog.focusedBtn (0↔1)
Enter pressed → if focusedBtn==0 → run confirmAction; if focusedBtn==1 → cancel
confirmKey pressed (e.g. "ctrl+d") → run confirmAction regardless
Esc pressed → cancel (nil out m.confirmDialog, reset pendingDeleteID etc)
```

---

## Dev Agent Record

### Implementation Plan
_To be filled during implementation_

### Debug Log
_To be filled as needed_

### Completion Notes
_To be filled on completion_

---

## File List

_Updated as tasks complete_

---

## Change Log

- 2026-02-23: Story created
