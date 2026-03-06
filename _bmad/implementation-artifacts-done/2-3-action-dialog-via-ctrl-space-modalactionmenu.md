# Story 2.3: Action Dialog via Ctrl+Space (ModalActionMenu)

Status: done

## Story

As an **operator**,
I want pressing `Ctrl+Space` on any table row to open a context-sensitive action menu showing all configured actions for that resource,
So that I can discover and execute any available action without memorising every key binding, while Space remains available for future row selection.

## Acceptance Criteria

1. **Given** `ModalActionMenu` is registered as a factory modal type using `ModalConfig` with `sizeHint: OverlayCenter`
   **When** the operator presses `Ctrl+Space` on a table row
   **Then** `ModalActionMenu` opens as a centered overlay listing all configured actions for the current resource type from `o6n-cfg.yaml`
   **And** mutation actions (HTTP verbs) are listed first
   **And** a visual separator appears before the first `type: navigate` action
   **And** navigate actions display a `→` suffix
   **And** `[J] View as JSON` and `[ctrl+j] Copy as JSON` are always the last two items

2. **Given** `ModalActionMenu` is open
   **When** the operator presses a single-character shortcut matching an action's configured key
   **Then** the action is dispatched immediately without requiring cursor movement or Enter
   **And** `Up`/`Down` moves the cursor; Enter dispatches the highlighted action
   **And** Esc closes the menu without executing any action

3. **Given** `Space` (without Ctrl) is pressed
   **When** the main table is focused
   **Then** the actions menu does NOT open (Space is reserved for future row selection)

4. **Given** `ModalActionMenu` is registered in `modalRegistry`
   **When** `currentViewHints` is called with `activeModal == ModalActionMenu`
   **Then** the hint line shows `↑↓ nav`, `Enter run`, `Esc close` from the modal's `HintLine`
   **And** table-view hints (e.g., `?` help) are NOT shown

5. **Given** `Ctrl+J` is pressed on a selected row
   **When** dispatched from the `ModalActionMenu` via the `ctrl+j` action item
   **Then** the row's JSON is copied to the system clipboard
   **And** the modal closes

## Tasks

- [x] Add `ModalActionMenu` constant to `ModalType` iota in `internal/app/model.go`
- [x] Remove `showActionsMenu bool` field from model; keep `actionsMenuItems` and `actionsMenuCursor`
- [x] Rename `renderActionsMenu` to `renderActionsMenuBody` in `view.go` — returns inner text only (no border/overlayCenter); adjust body to use value receiver (`model` not `*model`)
- [x] Remove `if m.showActionsMenu { ... }` rendering block from `view.go` (rendering now goes through factory)
- [x] Register `ModalActionMenu` in `modal.go` with `OverlayCenter`, body renderer calling `renderActionsMenuBody()`, and `HintLine: [{↑↓,nav,1},{Enter,run,1},{Esc,close,2}]`
- [x] Migrate `update.go` `if m.showActionsMenu` handler block to `if m.activeModal == ModalActionMenu`; replace `m.showActionsMenu = false` with `m.activeModal = ModalNone`; update `Ctrl+Space` trigger; remove `!m.showActionsMenu` guards
- [x] Add `ctrl+j` Copy as JSON item to `buildActionsForRoot()` in `nav.go` using `github.com/atotto/clipboard`
- [x] Update `actions_test.go` — replace `showActionsMenu` field references with `activeModal == ModalActionMenu`
- [x] Add `TestCurrentViewHints_ModalActionMenuActive` to `main_hints_test.go`
- [x] Add `TestActionsMenuItemsIncludeCtrlJ` to `actions_test.go`
- [x] `make test` green; `go vet` clean; `gofmt` no diff
- [x] Update `specification.md` — `ModalActionMenu` in modal types table; update Actions Menu section
- [x] Update `README.md` — actions key table and description

## Completion Notes

- `showActionsMenu bool` removed from `model.go`; `ModalActionMenu` is now a first-class `ModalType` constant
- `renderActionsMenuBody()` (value receiver) returns just the inner menu text; `renderModal` factory applies the `OverlayCenter` border and sizing
- `Ctrl+Space` sets `m.activeModal = ModalActionMenu`; all key guards simplified to `m.activeModal == ModalNone`
- `ctrl+j` Copy as JSON added as second-to-last/last item pair alongside `J` View as JSON; uses `github.com/atotto/clipboard` (already in go.mod)
- `Space` alone does not trigger actions (guard preserved — no rows/activeModal check)
- `currentViewHints` dispatch for `ModalActionMenu` works automatically through `modalRegistry` (Story 2.2 infrastructure)

## File List

- `internal/app/model.go` — added `ModalActionMenu` to `ModalType` iota; removed `showActionsMenu bool`
- `internal/app/view.go` — renamed/refactored `renderActionsMenu` → `renderActionsMenuBody` (value receiver, no border); removed `if m.showActionsMenu` rendering block
- `internal/app/modal.go` — registered `ModalActionMenu` with `OverlayCenter`, body renderer, `HintLine`
- `internal/app/update.go` — migrated all `showActionsMenu` refs to `activeModal == ModalActionMenu`; updated `Ctrl+Space` trigger; removed `!m.showActionsMenu` guards
- `internal/app/nav.go` — added `ctrl+j` Copy as JSON item to `buildActionsForRoot()`; added `github.com/atotto/clipboard` import
- `internal/app/actions_test.go` — updated tests to use `activeModal == ModalActionMenu`; added `TestActionsMenuItemsIncludeCtrlJ`; updated `TestActionsDefaultViewAsJSON`
- `internal/app/drilldown_nav_test.go` — updated `showActionsMenu = true` → `activeModal = ModalActionMenu`
- `internal/app/main_hints_test.go` — added `TestCurrentViewHints_ModalActionMenuActive`
- `specification.md` — updated modal types table; updated Actions Menu section to describe `ModalActionMenu` and `Ctrl+J`
- `README.md` — updated Actions key table and description

## Change Log

| Date | Change | Reason |
|---|---|---|
| 2026-03-05 | Story file created and implemented | Story 2.3 |
