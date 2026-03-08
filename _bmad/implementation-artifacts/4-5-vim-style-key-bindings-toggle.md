# Story 4.5: Vim-Style Key Bindings Toggle

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator** (Marco persona),
I want to **toggle vim-style key bindings in-session without restarting**,
so that **keyboard-native operators can use familiar `j`/`k`/`g`/`G` navigation alongside the default bindings.**

## Acceptance Criteria

1. **Given** the application is running
   **When** the operator presses `Ctrl+Shift+V` to toggle vim mode
   **Then** vim-style bindings (`j`/`k` for up/down, `g`/`G` for top/bottom) become active immediately.
   **And** the footer or status area indicates that vim mode is on.
   **And** pressing `Ctrl+Shift+V` again restores the default bindings and clears the indicator.

2. **Given** vim mode is active
   **When** the operator presses `j` or `k`
   **Then** the cursor moves down or up respectively — identical to arrow key behavior.

3. **Given** vim mode is active
   **When** the operator presses `g`
   **Then** the cursor jumps to the first row of the table.

4. **Given** vim mode is active
   **When** the operator presses `G` (Shift+G)
   **Then** the cursor jumps to the last row of the table.

5. **Given** the application is started with the `--vim` flag
   **When** the first frame renders
   **Then** vim mode is active by default.

6. **Given** the help modal is open (`?`)
   **When** vim mode is active
   **Then** the help screen includes the vim-style navigation keys in the "Navigation" category.

## Tasks / Subtasks

- [x] **Model & State Updates (AC: 1, 5)**
  - [x] `vimMode bool` added to `model` struct in `internal/app/model.go`
  - [x] `newModelEnvApp` / `run.go` initializes `m.vimMode = *vimFlag || appCfg.VimMode`
  - [x] `--vim` flag in `run.go` correctly initializes model state
- [x] **Interaction Logic (AC: 1, 2, 3, 4)**
  - [x] `Ctrl+Shift+V` toggle added to `Update()` in `update.go` (review fix): toggles `m.vimMode`, shows "VIM mode ON/OFF" footer message; guarded by `!m.searchMode && m.activeModal == ModalNone`
  - [x] `j`, `k`, `g`, `G`, `Ctrl+U`, `Ctrl+D` handlers in `Update()` gated on `m.vimMode`
  - [x] Vim keys disabled in search mode and when modal is active
- [x] **UI Feedback (AC: 1, 6)**
  - [x] `VIM` badge added to footer right area in `view.go` (review fix): rendered in accent color when `m.vimMode` is true
  - [x] `j/k` and `gg/G` hints added to `tableViewHints` in `hints.go` when `m.vimMode` is true (review fix)
  - [x] Help modal in `view.go` conditionally shows "VIM NAVIGATION" section when vim mode active
- [x] **Testing & Documentation**
  - [x] `internal/app/vim_test.go` covers j/k, gg, G, Ctrl+U, disabled-in-modal, disabled-in-search, pendingG timeout
  - [x] `TestCtrlShiftV_TogglesVimMode`, `TestCtrlShiftV_IgnoredInSearchMode`, `TestCtrlShiftV_IgnoredInModal` added (review fix)
  - [x] `specification.md` section 11 documents vim mode keys

## Dev Notes

### Previous Story Intelligence
- **Story 4.1-4.4 Learnings:** Layout decisions must use `m.termWidth` and `m.termHeight`.
- **Keyboard Grammar:** Use the established convention: uppercase in help = no Shift required. `G` in help means press Shift+G.

### Architecture Compliance
- **Key Conflict Prevention:** `j`, `k`, `g`, `G` must not conflict with any resource-specific action keys. If a resource has an action on `j`, vim navigation takes precedence when `m.VimMode` is true.
- **Input Focus:** Vim navigation MUST be disabled when `m.searchActive` or any `m.activeModal` with text input is active.

### Anti-Patterns to Avoid (LLM Guardrails)
- **❌ DO NOT** hardcode `j`/`k` as global keys; they are only active when `m.VimMode` is true.
- **❌ DO NOT** toggle vim mode with `V` alone; use `Ctrl+Shift+V` as per the latest UX design specification to free up `V` for future use.
- **❌ DO NOT** forget to clear the vim indicator when the mode is toggled off.

### UI/UX Standards
- **VIM Indicator:** Should be small and placed in the footer status area, near the API and refresh indicators.
- **Hint Integration:** Only show `j/k` hints in the footer when vim mode is active to avoid cluttering for default users.

### Project Structure Notes
- **Update Logic:** `internal/app/update.go`.
- **View Logic:** `internal/app/view.go`.
- **Model Logic:** `internal/app/model.go`.
- **Hints Logic:** `internal/app/hints.go`.

### References
- [Source: `_bmad/planning-artifacts/epics.md#Story 4.5`]
- [Source: `_bmad/planning-artifacts/prd.md#FR39`]
- [Source: `_bmad/planning-artifacts/ux-design-specification.md#Key binding changes`]

## Dev Agent Record

### Agent Model Used

Claude Haiku 4.5

### Debug Log References

### Senior Developer Review (AI)

- **Outcome:** Fixed — CRITICAL and HIGH issues resolved
- **Date:** 2026-03-08
- **Reviewer:** Claude Sonnet 4.6 (adversarial code review)
- **Action items taken:**
  - CRITICAL: `Ctrl+Shift+V` toggle was completely absent from `update.go` — FIXED: added handler that toggles `m.vimMode`, shows status message, is guarded by search/modal state
  - HIGH: VIM indicator not in footer — FIXED: added `VIM` badge in accent color to `view.go` footer right area
  - HIGH: Vim hints not added to hint system — FIXED: `tableViewHints` now appends `j/k` and `gg/G` hints when `m.vimMode` is true
  - HIGH: All tasks marked `[ ]` — FIXED: tasks updated to `[x]`
  - 3 new tests added for toggle behavior: `TestCtrlShiftV_TogglesVimMode`, `TestCtrlShiftV_IgnoredInSearchMode`, `TestCtrlShiftV_IgnoredInModal`

### Completion Notes List

- Story implementation verified and tested.
- CRITICAL fix: `Ctrl+Shift+V` runtime vim toggle added to `update.go`.
- HIGH fix: VIM indicator added to footer in `view.go`.
- HIGH fix: Vim hints added to `hints.go` (j/k, gg/G shown when vim mode active).
- All tests pass.

### File List

- `internal/app/update.go` — FIX: `Ctrl+Shift+V` toggle handler added
- `internal/app/view.go` — FIX: VIM badge in footer right area
- `internal/app/hints.go` — FIX: j/k and gg/G hints appended when `m.vimMode` is true
- `internal/app/vim_test.go` — NEW: vim navigation tests + 3 new Ctrl+Shift+V toggle tests
- `internal/app/vim_removal_test.go` — NEW: verifies j/k are no-ops in default mode
- `internal/app/model.go` — EXISTING: `vimMode bool` field
- `internal/app/run.go` — EXISTING: `--vim` flag initialization
