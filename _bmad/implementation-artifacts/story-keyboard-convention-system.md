# Story: Keyboard Convention System

## Summary

Codify, document, and enforce a systematic keyboard convention across all views. Ensure the help screen groups keys by convention category and that the specification includes a formal key allocation table.

## Motivation

Great TUIs like k9s and lazygit succeed because their keyboard conventions are predictable. o8n already follows sensible patterns (Ctrl+ for destructive, single key for safe actions), but this is implicit — not documented as a design principle and not enforced in the help screen.

## Acceptance Criteria

- [ ] **AC-1:** The help screen (`?`) groups key bindings by category: Global, Navigation, View Actions, Resource Actions (context-specific)
- [ ] **AC-2:** Each category in the help screen has a visible header/separator
- [ ] **AC-3:** Resource-specific actions (from `o8n-cfg.yaml`) are shown dynamically based on the current resource type
- [ ] **AC-4:** The help screen shows action key bindings for the current context only (not all 35 resource types at once)
- [ ] **AC-5:** Key binding conflicts are detectable: `Ctrl+D` is documented as context-dependent (half-page scroll in navigation, terminate/delete on resource views where the action is configured)

## Tasks

### Task 1: Refactor help screen content generation

Currently the help screen renders a static list. Refactor to:
1. Group bindings into categories (Global, Navigation, View, Context-Specific)
2. Add section headers with visual separators
3. Dynamically inject resource-specific action keys from the current table's `actions` config

### Task 2: Context-aware action display

When the help screen opens:
1. Read the current table's `actions` array from `o8n-cfg.yaml`
2. Append a "Resource Actions" section showing each action's key + label
3. If no resource-specific actions exist, omit the section

### Task 3: Update tests

1. Test that help content changes based on current resource type
2. Test that help content includes category headers
3. Test that destructive action keys are in the correct category

## Technical Notes

- The help screen rendering is in `internal/app/` (view logic)
- Action definitions are already loaded in the model via `appConfig.Tables[].Actions`
- No new dependencies required
