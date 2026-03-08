# Story 5.4: Specification & Documentation Accuracy

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **contributor** (Marco persona),
I want **`specification.md` to accurately reflect the post-sprint implementation**,
so that **I can understand the full system from documentation alone without reading source code**.

## Acceptance Criteria

1. **Given** the modal factory pattern has been implemented
   **When** `specification.md` is reviewed
   **Then** it documents: the `ModalConfig` struct fields (including `sizeHint` with all three size classes and `hintLine`), the `renderModal()` factory signature, and the rule that new modal types are added via config registration — not via switch statements.

2. **Given** the footer hint push model has been implemented
   **When** `specification.md` is reviewed
   **Then** it documents: the `Hint` struct fields, the `currentViewHints(m)` dispatch pattern, and `MinWidth`/`Priority` semantics.

3. **Given** the state transition contract has been implemented
   **When** `specification.md` is reviewed
   **Then** it documents: the `TransitionType` enum values and the rule that `prepareStateTransition` is mandatory for all navigation changes.

4. **Given** the `Ctrl+Space` action dialog and JSON viewer have been implemented
   **When** `specification.md` is reviewed
   **Then** it documents: `ModalActionMenu` (`Ctrl+Space` trigger, action list ordering, `[J]`/`[Ctrl+J]` as last item), and `ModalJSONView` (`J` opens, `Ctrl+J` copies, `OverlayLarge` size class).

5. **Given** the `env_name` semantic color role and `FirstRunModal` have been implemented
   **When** `specification.md` is reviewed
   **Then** it documents: `env_name` role in skin contract (fixed top-right header, primary env signal), `ui_color` as border accent only, and `FirstRunModal` flow (`Ctrl+H` to revisit).

## Tasks / Subtasks

- [x] **Audit Existing Specification (AC: all)**
  - [x] `specification.md` audited against current codebase; stale/missing sections identified
- [x] **Document TUI Architecture (AC: 1, 2, 4)**
  - [x] **Modal Factory**: Added section documenting `ModalConfig` struct, `SizeHint` size classes (`OverlayCenter`, `OverlayLarge`, `FullScreen`), `renderModal()` signature, and registration pattern (review fix)
  - [x] **Hint Push Model**: Added section documenting `Hint` struct fields, `filterHints` signature, `MinWidth`/`Priority` semantics, and `currentViewHints` dispatch pattern (review fix)
  - [x] **Action Discoverability**: `ModalActionMenu` already documented in section 6 (Actions Menu)
- [x] **Document Navigation & State (AC: 3, 5)**
  - [x] **State Transition Contract**: Section 4 documents `TransitionType` enum values and mandatory `prepareStateTransition` rule
  - [x] **FirstRunModal**: Section 13 documents trigger, key bindings, Esc-swallowed behavior, `Ctrl+H` to revisit
- [x] **Document Security & Identity (AC: 5)**
  - [x] **`env_name` semantic color role**: Added to section 8 with primary/secondary identity signal distinction and `ui_color` border-accent-only rule (review fix)
  - [x] Credential isolation rules documented in section 3 (`o6n-env.yaml` description)
- [x] **Cross-Reference Requirements**
  - [x] Specification accurately reflects sprint implementation; no stale `y`/YAML references found

## Dev Notes

### Previous Story Intelligence
- **Sprint Learnings:** This story is the "Definition of Done" for the entire quality sprint. It ensures that the speed of development doesn't leave the documentation behind.
- **Story 5.3 Learnings:** The extensibility of the system relies on contributors knowing how to use the generic dispatcher and table mapper without reading the Go source.

### Architecture Compliance
- **Specification Authority:** `specification.md` is the authoritative technical reference. All implementation changes made during Epics 1-5 MUST be reflected here.

### Anti-Patterns to Avoid (LLM Guardrails)
- **❌ DO NOT** leave stale documentation that refers to old patterns (like `y` for YAML copy or per-view render switches for modals).
- **❌ DO NOT** omit the architectural "why" — explain the rationale for the modal factory and state contract.
- **❌ DO NOT** assume the user knows the key bindings; document the keyboard grammar and case conventions.

### Project Structure Notes
- **Primary Document:** `specification.md`.
- **Reference Code:** `internal/app/modal.go`, `internal/app/hints.go`, `internal/app/nav.go`, `internal/app/view.go`.

### References
- [Source: `_bmad/planning-artifacts/epics.md#Story 5.4`]
- [Source: `_bmad/planning-artifacts/prd.md#Success Criteria`]
- [Source: `_bmad/planning-artifacts/architecture.md#TUI Architecture`]

## Dev Agent Record

### Agent Model Used

Claude Haiku 4.5

### Debug Log References

### Senior Developer Review (AI)

- **Outcome:** Fixed — HIGH documentation gaps resolved
- **Date:** 2026-03-08
- **Reviewer:** Claude Sonnet 4.6 (adversarial code review)
- **Action items taken:**
  - HIGH: `ModalConfig` struct and `renderModal()` factory not documented — FIXED: added "Modal Factory Pattern" section with struct definition, size class table, and registration instructions
  - HIGH: `Hint` struct not documented — FIXED: added "Hint Push Model" section with struct definition, `filterHints` signature, and semantics
  - HIGH: `env_name` semantic color role not documented — FIXED: added table of all key roles with `env_name` primary identity signal contract; documented `ui_color` border-accent-only rule
  - HIGH: All tasks marked `[ ]` — FIXED: tasks updated to `[x]`
  - PASS: State Transition Contract already in section 4
  - PASS: FirstRunModal already in section 13
  - PASS: No stale `y`/YAML references in keyboard section

### Completion Notes List

- Documentation audit performed; 3 major gaps found and fixed.
- Modal Factory Pattern section added to specification.md.
- Hint Push Model section added with Hint struct definition.
- env_name semantic role documented with primary/secondary identity contract.

### File List

- `specification.md` — FIX: Added "Modal Factory Pattern" section (ModalConfig struct, SizeHint classes, renderModal signature)
- `specification.md` — FIX: Added "Hint Push Model" section (Hint struct, filterHints, MinWidth/Priority semantics)
- `specification.md` — FIX: Added env_name to Semantic Color Roles table with primary identity signal documentation

