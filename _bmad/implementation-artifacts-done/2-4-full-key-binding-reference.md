# Story 2.4: Full Key Binding Reference

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want to open the full key binding reference via the `?` key,
So that I can discover any binding I've forgotten without leaving the application.

## Acceptance Criteria

1. **Given** the operator is in any view
   **When** the operator presses `?`
   **Then** the help modal opens as an `OverlayLarge` overlay displaying all key bindings organized by category
   **And** the background content remains visible behind the modal
   **And** a hint line is rendered at the bottom showing at minimum: `Esc Close`

2. **Given** the help modal is open
   **When** the operator presses Esc or `?` again
   **Then** the help modal closes and the operator returns to the previous view unchanged
   **And** the help modal is rendered via the modal factory (`ModalConfig` with `OverlayLarge` size hint)

## Tasks

- [x] Add `case "?"` to the `ModalHelp` key handler in `update.go` so pressing `?` again toggles the modal closed (AC2 gap)
- [x] Add test `TestHelpModalToggleWithQuestionMark` to `keybindings_test.go` — presses `?` to open, then `?` again to close, expects `ModalNone`
- [x] Update `specification.md` modal behavior table — `ModalHelp` row: add `?` as an additional close key
- [x] `make test` green; `go vet` clean; `gofmt` no diff

## Dev Notes

### Current implementation state (nearly done — one line gap)

`ModalHelp` is fully implemented and factory-registered **before this story**:

- **`internal/app/modal.go`** — `ModalHelp` registered with `OverlayLarge`, `HintLine: [{↑↓,scroll,1},{q/Esc,close,1}]`
- **`internal/app/view.go`** — `modalHelpBody()` provides scrollable content: NAVIGATION/ACTIONS/GLOBAL/SEARCH/CONTEXT/STATUS INDICATORS/RESOURCE ACTIONS (dynamic by config)/VIEWS (navigate actions)
- **`internal/app/update.go` line ~397** — `if m.activeModal == ModalHelp {` handler handles: `down/ctrl+d` scroll, `up/ctrl+u` scroll, `esc/q` close, **all other keys swallowed** (`return m, nil`)
- **`internal/app/update.go` line ~727** — `case "?":` sets `m.activeModal = ModalHelp` and resets `m.helpScroll = 0`

### The single AC gap

AC2 requires `?` again to close the modal. Currently, `?` while ModalHelp is open falls through to the **final `return m, nil`** at the end of the ModalHelp `if` block, swallowing the key.

**Fix:** In the `switch s` block inside `if m.activeModal == ModalHelp`, add:
```go
case "?":
    m.activeModal = ModalNone
    m.helpScroll = 0
    return m, nil
```

This mirrors the existing `esc`/`q` case exactly.

### Existing tests to be aware of

- `internal/app/keybindings_test.go:141` — `TestHelpModalOpenAndDismiss`: tests `?` opens, `Esc` closes. After the fix, add a companion test for `?`→`?` toggle.
- `internal/app/keyboard_convention_test.go` — tests help screen content categories (headers, resource actions, separators) — **do not modify** these tests
- `internal/app/main_modal_test.go:100` — `TestModalSizeHint_OverlayLargeHelp` — confirms ModalHelp uses OverlayLarge — **do not modify**

### Specification update

`specification.md` modal behavior table row for `ModalHelp`:

**Current:**
```
| `ModalHelp` | Close + reset scroll | Swallowed (no-op) | `q` | Only `q`/`Esc` close; all other keys swallowed |
```

**New:**
```
| `ModalHelp` | Close + reset scroll | Swallowed (no-op) | `q`, `?` | `q`/`Esc`/`?` close; all other keys swallowed |
```

### Key files to touch

| File | Change |
|---|---|
| `internal/app/update.go` | Add `case "?":` to ModalHelp switch — closes + resets scroll |
| `internal/app/keybindings_test.go` | Add `TestHelpModalToggleWithQuestionMark` |
| `specification.md` | Update ModalHelp row in modal behavior table |

### Files NOT to touch

- `internal/app/view.go` — help content rendering is complete
- `internal/app/modal.go` — ModalHelp registration is complete
- `internal/app/keyboard_convention_test.go` — content tests unrelated to this story
- `o6n-cfg.yaml`, `o6n-env.yaml` — configuration files

### Project Structure Notes

- All changes are in `internal/app/` following existing patterns
- Test files live next to the code they test; new test in `keybindings_test.go` alongside `TestHelpModalOpenAndDismiss`
- Modal handler pattern: `if m.activeModal == X { switch s { case ...: return m, nil } return m, nil }` — match this exactly
- Test helper: use `sendKeyString(m, "?")` from `keybindings_test.go`

### References

- Epics Story 2.4 ACs: `_bmad/planning-artifacts/epics.md` — Story 2.4 section
- Current ModalHelp handler: `internal/app/update.go:397` (`if m.activeModal == ModalHelp`)
- Current `?` trigger: `internal/app/update.go:727` (`case "?":`)
- Existing test: `internal/app/keybindings_test.go:141` (`TestHelpModalOpenAndDismiss`)
- Spec modal table: `specification.md` — Section 7 Modal Keyboard Behaviour table

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.6 (claude-sonnet-4.6)

### Debug Log References

### Completion Notes List

- Added `case "?":` to `ModalHelp` switch in `update.go` alongside existing `esc`/`q` — closes modal and resets scroll
- `ModalHelp` was 97% pre-implemented; this was the single remaining AC2 gap
- All 4 ACs verified: `?` opens (AC1), `?` again closes (AC2), `Esc` closes (AC2), OverlayLarge factory (AC2), hint line present (AC1)
- New test `TestHelpModalToggleWithQuestionMark` covers the `?`→`?` toggle path
- `make test` clean, `go vet` clean, `gofmt` no diff

### File List

- `internal/app/update.go` — added `"?"` to `esc/q` case in ModalHelp handler
- `internal/app/keybindings_test.go` — added `TestHelpModalToggleWithQuestionMark`
- `specification.md` — updated ModalHelp modal behavior table row (added `?` as close key)
- `internal/app/modal.go` — updated ModalHelp hint to `q/?/Esc close` for UI consistency
- `internal/app/main_hints_test.go` — updated ModalHelp hint expectation to `q/?/Esc`
- `README.md` — documented `?` toggles help closed when already open
- `_bmad/implementation-artifacts/2-4-full-key-binding-reference.md` — story file (status → done)
- `_bmad/implementation-artifacts/sprint-status.yaml` — 2-4 status → done

## Senior Developer Review (AI)

### Reviewer

Karsten (AI-assisted) on 2026-03-05

### Findings Summary

- Fixed 2 medium issues and 1 low issue from review:
  - Help modal hint line now reflects actual close keys (`q/?/Esc`)
  - README global shortcut row now documents `?` toggles close
  - Stale test comment corrected; hint test updated to match current behavior
- ACs verified as implemented and tests/vet green.

### Outcome

Approve — story complete.

## Change Log

- 2026-03-05: Applied review fixes for help hint consistency and docs accuracy; set story status to done.
