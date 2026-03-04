# Story 1.1: Modal Factory Foundation

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **developer contributing to o8n**,
I want a `ModalConfig` struct and `renderModal()` factory function in `internal/app/modal.go`,
so that all modal types are rendered from a single, consistent code path with no per-type layout logic in the view render path.

## Acceptance Criteria

1. **Given** `ModalConfig` is defined with: `sizeHint` (OverlayCenter / OverlayLarge / FullScreen), `title string`, `bodyRenderer func(m Model) string`, button label fields, and `hintLine []Hint`
   **When** `renderModal(m Model, cfg ModalConfig)` is called for any registered modal type
   **Then** the rendered modal has identical border style, padding, and button placement regardless of which type it is
   **And** the view render path contains no `switch modalType` statement (or `if m.activeModal == ...` chain) for modal body content
   **And** new modal types are added by defining a new `ModalConfig` — no changes to `renderModal()` required
   **And** all existing modal types (confirm delete, confirm quit, help, edit, sort, environment, task complete) are migrated to use the factory

2. **Given** `sizeHint` is `OverlayCenter`
   **When** the modal renders
   **Then** it occupies approximately 50% of terminal width, auto-height, centered over the base view

3. **Given** `sizeHint` is `OverlayLarge`
   **When** the modal renders
   **Then** it occupies approximately 80% of terminal width × 80% of terminal height, centered, with background content visible behind it
   **And** the `hintLine []Hint` field is populated and rendered at the modal bottom using the same `Hint{Key, Label, MinWidth, Priority}` system as the main footer
   **And** the hint line includes at minimum `Esc Close` and the modal's primary action

4. **Given** `sizeHint` is `FullScreen`
   **When** the modal renders
   **Then** it occupies the full terminal viewport
   **And** the `hintLine []Hint` field is populated and rendered at the modal bottom

5. **Given** the test file `internal/app/main_modal_test.go` is created
   **When** `make cover` is run
   **Then** line coverage on `modal.go` is ≥ 80%

## Tasks / Subtasks

- [x] Task 1: Define `Hint` struct and `SizeHint` type (AC: 1, 3, 4)
  - [x] Create `internal/app/hints.go` with just the `Hint{Key, Label, MinWidth, Priority int}` struct (stub only — the full push model is Story 2.1)
  - [x] Define `SizeHint` type + constants: `OverlayCenter`, `OverlayLarge`, `FullScreen` in `modal.go`

- [x] Task 2: Define `ModalConfig` struct and modal registry (AC: 1)
  - [x] Define `ModalConfig` struct in `modal.go` with fields: `SizeHint SizeHint`, `Title string`, `BodyRenderer func(m model) string`, `ConfirmLabel string`, `CancelLabel string`, `HintLine []Hint`
  - [x] Define `modalRegistry map[ModalType]ModalConfig` at package level in `modal.go`
  - [x] Add `registerModal(t ModalType, cfg ModalConfig)` helper

- [x] Task 3: Implement `renderModal()` factory and overlay helpers (AC: 1, 2, 3, 4)
  - [x] Implement `renderModal(m model, cfg ModalConfig) string` in `modal.go` — OverlayCenter returns body as-is, OverlayLarge applies RoundedBorder at ~80% width + HintLine, FullScreen returns body as-is
  - [x] Implement `overlayLarge(bg, fg string) string` — delegates to overlayCenter (same centering logic; size distinction enforced by OverlayLarge body width in renderModal)
  - [x] Implement `overlayFullscreen(bg, fg string, termW, termH int) string` — full viewport via lipgloss.Place
  - [x] Wire `renderModal()` in `View()`: replaced the if-else chain (view.go:862–886) with registry lookup + size-class dispatch

- [x] Task 4: Migrate existing modals to factory configs (AC: 1)
  - [x] `ModalConfirmDelete` → `OverlayCenter` config; bodyRenderer calls `renderConfirmDeleteModal()`; fixed DoubleBorder → RoundedBorder
  - [x] `ModalConfirmQuit` → `OverlayCenter` config; bodyRenderer calls `renderConfirmQuitModal()`
  - [x] `ModalSort` → `OverlayCenter` config; bodyRenderer calls `renderSortPopup()`
  - [x] `ModalEnvironment` → `OverlayCenter` config; bodyRenderer calls `renderEnvPopup()`
  - [x] `ModalEdit` → `OverlayCenter` config; bodyRenderer calls `renderEditModal()`
  - [x] `ModalHelp` → `OverlayLarge` config; added `modalHelpBody()` (content without hint line); HintLine: `[↑↓ scroll, Esc close]`
  - [x] `ModalDetailView` → `OverlayLarge` config; added `modalDetailViewBody()` (content without hint in title); HintLine: `[↑↓ scroll, q/Esc close]`
  - [x] `ModalTaskComplete` → `FullScreen` config; bodyRenderer calls `renderTaskCompleteModal()`; HintLine: `[Tab switch, Enter confirm, Esc cancel]`
  - [x] Replaced if-else chain from `View()` (view.go ~862–886) with factory dispatch
  - [x] Fixed `ModalConfirmDelete` DoubleBorder → RoundedBorder for uniform border style

- [x] Task 5: Register all modal configs at init (AC: 1)
  - [x] `init()` function in `modal.go` registers all 8 modal types

- [x] Task 6: Write tests (AC: 5)
  - [x] Created `internal/app/main_modal_test.go`
  - [x] Tests `renderModal()` with each size class (OverlayCenter, OverlayLarge, FullScreen)
  - [x] Tests registry lookup: unregistered type produces no config (View() handles gracefully)
  - [x] Tests hint line rendering (renderModalHintLine, OverlayLarge HintLine content)
  - [x] Tests HintLine populated for OverlayLarge (ModalHelp, ModalDetailView) and FullScreen (ModalTaskComplete)
  - [x] `make cover` confirms 95.8% average coverage on `modal.go` (≥80% threshold met)

- [x] Task 7: Verify all tests pass (AC: all)
  - [x] `make test` passes with zero regressions
  - [x] `go vet ./...` passes
  - [x] `gofmt -w .` produces no changes

## Dev Notes

### Context: Why This Story Exists

The current `View()` in `view.go` (lines 862–886) contains a large `if m.activeModal == ... else if ...` chain that calls 8 individual per-type `render*Modal()` methods. Each method independently constructs borders, padding, and button layout — and they are inconsistent (e.g., `ModalConfirmDelete` uses `DoubleBorder` at width=54, `ModalConfirmQuit` uses `RoundedBorder` at width=44). This directly violates FR11 (identical border/padding/button placement) and NFR15 (config-driven modal system).

This story extracts the modal rendering concern into a factory pattern. The deliverable is `internal/app/modal.go` — a new file.

### Existing Code to Understand Before Starting

**`internal/app/view.go`** — The main render file:
- `View()` at the bottom builds `baseView`, then the if-else modal dispatch at line 862
- Individual modal render functions: `renderConfirmDeleteModal()` (line 165), `renderConfirmQuitModal()` (line 207), `renderHelpScreen()` (line 258), `renderDetailView()` (line ~374), `renderSortPopup()` (line ~450), `renderEnvPopup()` (line ~500), `renderTaskCompleteModal()` (line ~580)
- `overlayCenter(bg, fg string) string` (line 1468) — the only existing overlay function; places `fg` centered over `bg` by string manipulation. **Reuse this logic** for `overlayLarge()` but constrain `fg` to 80%×80% of terminal dimensions before passing to centering math.

**`internal/app/model.go`** — Model definition:
- `ModalType` enum (line 138): `ModalNone`, `ModalConfirmDelete`, `ModalConfirmQuit`, `ModalHelp`, `ModalEdit`, `ModalSort`, `ModalDetailView`, `ModalEnvironment`, `ModalTaskComplete`
- `KeyHint` struct (line 130): `{Key string, Description string, Priority int}` — **this is NOT the architecture-spec `Hint` struct**. The architecture spec `Hint` has `{Key, Label, MinWidth, Priority int}`. See naming note below.

**`internal/app/edit.go`** — Edit modal render: `renderEditModal()` is the most complex body — it has live `textinput.Model` state that the body renderer must capture from `m`. The `bodyRenderer func(m model) string` closure pattern handles this correctly since `m` is passed at render time.

**`internal/app/transition.go`** — Navigation: The existing `prepareStateTransition` uses `transitionScope` enum (`transitionEnvSwitch`, `transitionContextSwitch`, `transitionDrilldown`, `transitionBack`, `transitionBreadcrumb`). **Do NOT rename or touch** `transitionScope` in this story — that naming alignment to `TransitionType` is Story 1.2's responsibility. Story 1.1 has no dependency on the transition contract refactor.

### Critical Naming Issue: `Hint` vs `KeyHint`

| | Existing (`model.go`) | Architecture Spec (Decision 3) |
|---|---|---|
| Type name | `KeyHint` | `Hint` |
| Fields | `Key, Description, Priority` | `Key, Label, MinWidth, Priority` |
| File | `model.go` | `hints.go` (Story 2.1 deliverable) |

**Resolution for Story 1.1:** Create `internal/app/hints.go` with **only the `Hint` struct definition** — the full push model (per-view hint functions, `currentViewHints()`, footer wiring) is Story 2.1. This avoids a blocking dependency and gives Story 2.1 the correct type to build on.

**Do NOT rename `KeyHint` to `Hint`** in this story — `KeyHint` is used by existing hint display code in `view.go`'s `getKeyHints()`. Story 2.1 will handle that rename/replacement. The `Hint` type in `hints.go` is a separate, new type used only by `ModalConfig.hintLine` for now.

### Architecture Spec: Size Class Assignments

From `architecture.md` Decision 1:
```
OverlayCenter  — Edit, Sort, ConfirmDelete, ConfirmQuit, ModalActionMenu, FirstRunModal
OverlayLarge   — ModalHelp, ModalDetailView, ModalJSONView  (~96×16 at 120×20 minimum)
FullScreen     — TaskComplete dialog
```

ModalActionMenu, FirstRunModal, ModalJSONView are future stories — do NOT add their `ModalType` constants yet. This story only migrates the 8 **existing** modal types.

### `renderModal()` Factory Design

```go
// modal.go — suggested structure

type SizeHint int
const (
    OverlayCenter  SizeHint = iota
    OverlayLarge
    FullScreen
)

type ModalConfig struct {
    SizeHint     SizeHint
    Title        string
    BodyRenderer func(m model) string
    ConfirmLabel string // e.g. "Confirm", "Delete", "Save"
    CancelLabel  string // e.g. "Cancel", "Close"
    HintLine     []Hint // required for OverlayLarge and FullScreen
}

var modalRegistry = map[ModalType]ModalConfig{}

func registerModal(t ModalType, cfg ModalConfig) {
    modalRegistry[t] = cfg
}

// renderModal is the sole entry point for all modal rendering.
// Called from View() instead of the if-else chain.
func renderModal(m model, cfg ModalConfig) string {
    body := cfg.BodyRenderer(m)
    // apply uniform border, padding, button layout
    // append hintLine if OverlayLarge or FullScreen
    ...
}
```

**Uniform border spec** (to fix existing inconsistency):
- All modals: `lipgloss.RoundedBorder()`, `BorderForeground(col(m.skin, "borderFocus"))`, `Padding(1, 2)`
- Width: determined by `SizeHint` (OverlayCenter: ~50% of `m.lastWidth`; OverlayLarge: ~80% of `m.lastWidth`)

**View() after migration** (simplified):
```go
// Replace lines 862-886 with:
if m.activeModal != ModalNone {
    if cfg, ok := modalRegistry[m.activeModal]; ok {
        overlay := renderModal(m, cfg)
        switch cfg.SizeHint {
        case OverlayLarge:
            return overlayLarge(baseView, overlay, m.lastWidth, m.lastHeight)
        case FullScreen:
            return overlayFullscreen(baseView, overlay, m.lastWidth, m.lastHeight)
        default: // OverlayCenter
            return overlayCenter(baseView, overlay)
        }
    }
}
```

### `overlayLarge()` Implementation Notes

The existing `overlayCenter()` (view.go:1468) places `fg` centered over `bg` using string manipulation. For `overlayLarge()`:
1. First constrain `fg` to 80%×80% of terminal: target width = `int(float64(termW) * 0.8)`, target height = `int(float64(termH) * 0.8)`
2. Then use the same centering math as `overlayCenter()` to position it
3. At 120×20 minimum: OverlayLarge is approximately 96 columns × 16 rows

### Testing Requirements

Test file: `internal/app/main_modal_test.go` (co-located with `modal.go`, not in a separate `tests/` dir).

Coverage target: ≥80% line coverage on `modal.go` per architecture convention (`make cover`).

Key test scenarios:
- `renderModal()` with `OverlayCenter` config: verify output contains the uniform border style
- `renderModal()` with `OverlayLarge` config: verify `HintLine` content appears in output
- `renderModal()` with `FullScreen` config: verify `HintLine` appears
- Registry miss: calling from `View()` with unregistered `ModalType` produces no overlay (not a panic)
- All 8 migrated modal body renderers: can be called with a minimal `model{}` without panicking (smoke test)

Use the existing test pattern from `internal/app/overlay_modal_test.go` and `internal/app/confirm_dialog_test.go` as reference for model construction and assertion style.

### Project Structure Notes

- **New files:** `internal/app/modal.go`, `internal/app/hints.go` (stub), `internal/app/main_modal_test.go`
- **Modified files:** `internal/app/view.go` (replace if-else chain, keep individual body render helpers as private functions), `internal/app/model.go` (no enum changes — do not add new ModalType constants in this story)
- **Off-limits:** `internal/operaton/` — never modify
- **`o8n-cfg.yaml` and `o8n-env.yaml`** — no changes required by this story
- The actions menu (`showActionsMenu` bool flag + `renderActionsMenu()`) is **not part of the modal factory** in this story — it uses a separate `bool` flag. `ModalActionMenu` integration is a future story.

### Async / Bubble Tea Constraints

- `renderModal()` is called from `View()` — it is a **pure function**. No side effects, no model mutations.
- Body renderers are closures over `model` (value, not pointer) — they read state, never write it.
- No `tea.Cmd` in modal rendering. Any async actions triggered by modal key presses live in `Update()`, unchanged.

### Specification Update Obligation

Per the architecture definition of done: after implementation, `specification.md` MUST be updated to document:
- The modal factory pattern (`ModalConfig`, `renderModal()`, `SizeHint`)
- The three size classes and their usage assignments
- The `HintLine []Hint` contract for OverlayLarge/FullScreen modals

### References

- [Source: `_bmad/planning-artifacts/architecture.md#Decision 1: Modal Factory Pattern`] — ModalConfig struct, SizeHint enum, size class assignments, factory function signature, rationale
- [Source: `_bmad/planning-artifacts/epics.md#Story 1.1: Modal Factory Foundation`] — Acceptance criteria (BDD format)
- [Source: `internal/app/view.go:862–886`] — Existing if-else modal dispatch to eliminate
- [Source: `internal/app/view.go:165–203`] — `renderConfirmDeleteModal()`: DoubleBorder, width=54, inline hint text
- [Source: `internal/app/view.go:207–224`] — `renderConfirmQuitModal()`: RoundedBorder, width=44
- [Source: `internal/app/view.go:258–~370`] — `renderHelpScreen()`: scrollable help content, FullScreen candidate → architecture says OverlayLarge
- [Source: `internal/app/view.go:1468`] — `overlayCenter()`: existing overlay placement implementation (reuse logic for overlayLarge/overlayFullscreen)
- [Source: `internal/app/model.go:130–135`] — `KeyHint` struct (NOT the same as architecture spec `Hint`)
- [Source: `internal/app/model.go:138–151`] — `ModalType` enum and existing constants
- [Source: `_bmad/planning-artifacts/architecture.md#Decision 3: Footer Hint Rendering`] — `Hint{Key, Label, MinWidth, Priority}` struct spec
- [Source: `internal/app/overlay_modal_test.go`] — Existing overlay test pattern to follow
- [Source: `internal/app/confirm_dialog_test.go`] — Existing confirm dialog test pattern

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

None — no panics or errors encountered during implementation.

### Completion Notes List

- ✅ Created `internal/app/hints.go` with `Hint{Key, Label, MinWidth, Priority}` struct stub
- ✅ Created `internal/app/modal.go` with `SizeHint` type, `ModalConfig` struct, `modalRegistry`, `registerModal()`, `renderModal()` factory, `renderModalHintLine()`, `overlayLarge()`, `overlayFullscreen()`, and `init()` registering all 8 existing modal types
- ✅ Fixed `ModalConfirmDelete` border: `DoubleBorder` → `RoundedBorder` (uniform border AC)
- ✅ Added `modalHelpBody()` to view.go: scrolled help content without trailing hint line; factory applies HintLine separately
- ✅ Added `modalDetailViewBody()` to view.go: detail content with clean title (no inline hint); factory applies HintLine separately
- ✅ Replaced the 8-branch if-else modal dispatch in `View()` with registry lookup + size-class dispatch (single code path)
- ✅ OverlayLarge modals render at ~80% terminal width with RoundedBorder and HintLine from `[]Hint` structs
- ✅ FullScreen modals (TaskComplete) render full viewport via `overlayFullscreen()` + `lipgloss.Place()`
- ✅ All 8 existing ModalType constants registered; no new constants added (ModalActionMenu, ModalJSONView etc. are future stories)
- ✅ `showActionsMenu` bool path unchanged (not part of modal factory in this story)
- ✅ 18 tests in `main_modal_test.go` covering registry, size classes, hint line, overlay helpers, body renderers
- ✅ Coverage on modal.go: 95.8% (threshold: 80%)
- ✅ Full test suite: 0 regressions, all packages pass
- ✅ `go vet ./...`: no issues
- ✅ `gofmt -w .`: clean

**Implementation notes:**
- OverlayCenter body renderers return complete styled boxes (existing render methods called as-is); renderModal() for OverlayCenter returns body unchanged. This avoids double-border without requiring full body extraction.
- OverlayLarge body renderers return raw text content; renderModal() applies RoundedBorder at `int(termW*0.80)` width and appends HintLine.
- FullScreen body renderer calls existing `renderTaskCompleteModal()` which manages its own custom layout; renderModal() for FullScreen returns body unchanged; View() passes to `overlayFullscreen()`.
- The `Hint` struct is defined in `hints.go` as a stub; Story 2.1 adds the push model (per-view hint functions, `currentViewHints()`, footer wiring).
- `transitionScope` naming in `transition.go` left unchanged — Story 1.2 is responsible for alignment to `TransitionType`.

### File List

- `internal/app/hints.go` (NEW) — `Hint{Key, Label, MinWidth, Priority}` struct
- `internal/app/modal.go` (NEW) — `SizeHint`, `ModalConfig`, `modalRegistry`, `registerModal`, `renderModal`, `renderModalHintLine`, `overlayLarge`, `overlayFullscreen`, `init`
- `internal/app/view.go` (MODIFIED) — Fixed `renderConfirmDeleteModal` DoubleBorder→RoundedBorder; added `modalHelpBody()`, `modalDetailViewBody()`; replaced if-else modal dispatch with factory registry lookup
- `internal/app/main_modal_test.go` (NEW) — 18 tests covering modal factory, registry, size classes, hint line, overlay helpers
- `_bmad/implementation-artifacts/1-1-modal-factory-foundation.md` (MODIFIED) — review findings and status
- `_bmad/implementation-artifacts/sprint-status.yaml` (MODIFIED) — development status sync

### Senior Developer Review (AI)

- Fixed AC compliance gaps in modal factory:
  - OverlayCenter now applies uniform border/padding/width in `renderModal()` for consistent rendering across center-sized modals.
  - OverlayLarge now enforces height (~80% of terminal) in addition to width.
  - `HintLine` rendering now respects `MinWidth` and `Priority`.
- Normalized center modal body renderers in `view.go` to return content-only strings so factory styling is the single path.
- Validation run after fixes:
  - `go test ./internal/app -run 'Modal|Overlay|Hint|SortModalMinWidth30' -count=1`
  - `go test ./... -count=1`
  - `go vet ./...`
  - `go build ./...`

### Change Log

- 2026-03-04: Senior review completed; fixed high/medium findings for modal factory consistency and set story status to `done`.
