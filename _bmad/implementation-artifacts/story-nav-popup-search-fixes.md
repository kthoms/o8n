# Story: Nav Fixes, Popup Scroll, Context Switch, Search Popup

## Summary
6 interconnected UX issues with navigation, the context-switch popup, and search.

## Issues

### T1 — Esc back navigation: title shows stale (or no) row count
`m.currentRoot` is not restored when popping navigation state. Title rendering
uses `m.pageTotals[m.currentRoot]` so the count appears blank/wrong. Fix: restore
`m.currentRoot = prevState.viewMode` and trigger a re-fetch so count is fresh.

### T2 — Process-variables drilldown broken
`api_path: /process-instance/{parentId}/variables` has a path-param placeholder
`{parentId}`. `fetchGenericCmd` appends all `genericParams` as query params, so
the actual URL becomes `…/variables?processInstanceId=xxx` instead of
`…/xxx/variables`. Fix: substitute `{key}` placeholders in the URL path before
appending remaining params as query string.

### T3 — Arrow-right should also trigger drilldown (addl to Enter)
Minor: add `"right"` key as a drilldown trigger alongside `"enter"`, guarded so
it does not interact with popup selection.

### T4 — Context popup: "…23 more" truncation should be replaced by scrolling
Popup rendering truncates to `maxShow=8` items and appends "…N more". Replace
with a scroll offset so all items are reachable via ↑/↓ and PgUp/PgDown.

### T5 — Context switch to "engines" shows process-definitions
`fetchForRoot` falls back to `fetchGenericCmd("process-definition")` when no
table def exists for the selected context. Fix: remove the fallback — just call
`fetchForRoot(rc)` regardless. `fetchGenericCmd` already handles the no-table-def
case by using `/{root}` as path.

### T6 — Search "/" should use the context-switch popup dialog
Instead of a separate footer search bar, pressing "/" opens the command-palette
popup in a new `popupModeSearch` mode. Typing filters the table rows live. The
popup list shows matching row (first-column) values. Enter locks the filter and
closes the popup; Esc restores rows and closes the popup.

## Files Changed
- `internal/app/model.go` — add `offset int` to `popupState`; add `popupModeSearch`
- `internal/app/update.go` — T1 (restore currentRoot), T3 (right key), T4 (scroll cursor logic), T5 (remove fallback), T6 (search popup mode)
- `internal/app/commands.go` — T2 (path param substitution); T5 (remove fetchForRoot fallback)
- `internal/app/view.go` — T4 (render scroll window); T6 (render search popup)
- `README.md` — document search popup, scrollable context list, right-arrow drilldown

## Acceptance Criteria
- AC1: After Esc back navigation, breadcrumb title shows correct row count
- AC2: Drilling into process-variables loads the variable list for the selected instance
- AC3: Right arrow on a row with drilldown navigates into it
- AC4: Context popup scrolls through all contexts without "…N more" text
- AC5: Switching to "engines" (or any root without a table def) no longer shows process-definitions
- AC6: "/" opens the popup dialog in search mode; typing filters the table; Enter/Esc close popup
