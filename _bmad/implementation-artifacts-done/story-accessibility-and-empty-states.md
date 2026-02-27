# Story: Accessibility and Empty State Improvements

## Summary

Improve accessibility by documenting and verifying color-blind safe behavior, and add contextual messaging for empty states. These are polish items that elevate the UX from good to excellent.

## Motivation

1. **Accessibility:** o8n uses color-coded status indicators (`green circle`=operational, `red X`=unreachable, `yellow circle`=unknown). The shape/symbol differentiation already works for color-blind users, but this is accidental — not verified or documented. High-contrast skins exist but aren't labeled as accessible.

2. **Empty states:** When a table has no rows (e.g., no process instances for a definition), the content box shows blank space. Great TUIs show contextual hints in empty states (k9s: "No resources found", lazygit: contextual tips).

## Acceptance Criteria

### Accessibility
- [x] **AC-1:** All information conveyed by color is also conveyed by shape or text (verify: status indicators use distinct symbols `green circle`/`X`/`dim circle`, sort direction uses `^`/`v`, errors use `X` prefix)
- [x] **AC-1b:** The help screen is fully navigable without color perception — section headers use text/structural separators (e.g., `--- Global ---`, `--- Navigation ---`), not just colored text
- [x] **AC-2:** At least 2 skins are tagged as high-contrast friendly in the skin system (e.g., `stock`, `black-and-wtf`)
- [x] **AC-3:** The specification documents the accessibility approach

### Empty States
- [x] **AC-4:** When a table fetch returns zero rows, the content box displays a centered hint: `No <resource-name> found`
- [x] **AC-5:** When a drilldown fetch returns zero rows, the hint includes context: `No instances for <definition-key>`
- [x] **AC-6:** When an API error occurs and rows are cleared, the hint displays: `Error loading data — press r to retry`
- [x] **AC-7:** Empty state messages use muted color (not error red, not accent — neutral grey)

## Tasks

### Task 1: Implement empty state messaging

1. In the view rendering, detect when `len(tableRows) == 0`
2. Render a centered message in the content box area
3. Choose message based on context:
   - Fresh load with no results: `No <displayName> found`
   - Drilldown with no children: `No <target> for <parent-key>`
   - Error state: `Error loading data — press r to retry`
4. Style with muted/grey color from the skin

### Task 2: Verify accessibility compliance

1. Audit all color-only information channels in the UI
2. Verify that each has a non-color fallback (symbol, text, position)
3. Document findings in the specification (section on accessibility)
4. Tag high-contrast skins in the skin metadata (optional: add a `high_contrast: true` field)

### Task 3: Update tests

1. Test empty state message rendering for zero-row tables
2. Test empty state after error
3. Test empty state message includes resource context

## Technical Notes

- Empty state rendering goes in the view logic where the table is rendered
- The `displayName` for resources is derived from the table `name` (capitalize, replace `-` with space)
- Skin metadata could use a `tags` field for accessibility labeling, but this is optional
