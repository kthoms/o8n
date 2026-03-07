# Story 4.5: Vim-Style Key Bindings Toggle

Status: ready-for-dev

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

- [ ] **Model & State Updates (AC: 1, 5)**
  - [ ] Add `VimMode bool` to the `Model` struct in `internal/app/model.go`.
  - [ ] Update `NewModel` (or equivalent initialization) to accept the vim mode preference.
  - [ ] Ensure the `--vim` flag in `main.go` correctly initializes the model state.
- [ ] **Interaction Logic (AC: 1, 2, 3, 4)**
  - [ ] Update `Update()` in `internal/app/update.go` to handle `Ctrl+Shift+V` (or `ctrl+V` depending on terminal behavior) for toggling `m.VimMode`.
  - [ ] Implement `j`, `k`, `g`, and `G` message handling in the table view's update loop when `m.VimMode` is true.
  - [ ] Ensure vim keys are ignored when an input field (search, edit modal) is focused.
- [ ] **UI Feedback (AC: 1, 6)**
  - [ ] Update `renderFooter` in `internal/app/view.go` to display a "VIM" indicator (e.g., in accent color) when `m.VimMode` is active.
  - [ ] Add `j`, `k`, `g`, `G` to the hint system in `internal/app/hints.go` when vim mode is enabled.
  - [ ] Update the help modal content in `internal/app/view.go` to conditionally show vim keys.
- [ ] **Testing & Documentation**
  - [ ] Create `internal/app/vim_test.go` to verify toggle behavior and navigation key mapping.
  - [ ] Update `specification.md` to document the vim mode feature and key bindings.

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

Gemini 2.0 Flash

### Debug Log References

### Completion Notes List

### File List
