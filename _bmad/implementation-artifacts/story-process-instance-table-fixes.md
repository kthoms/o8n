# Story: Process-Instance Table Fixes

**Status:** in-progress  
**Source:** Party mode session 2026-02-23

## Problem

When drilling into process-instances from a process-definition, five issues are present:

1. Title shows "8 items" but only 2 rows rendered (count not filtered by params)
2. ID column truncates UUID ("551ca863-0e50-11…") despite available width
3. Title context shows `processDefinitionId=ReviewInvoice:1:54171f85-…` — verbose, unreadable
4. Row coloring differs from root table (hardcoded colors bypass skin)
5. Last breadcrumb entry has a hotkey `[3]` — already there, no navigation needed

## Acceptance Criteria

- **AC1** Count in title matches actual filtered row count
- **AC2** Full UUID visible in ID column when width allows
- **AC3** Drilldown title shows target name + configured attribute value (no attribute name); fallback to parent ID
- **AC4** Row status colors match skin roles (success/warning/danger/fgMuted)
- **AC5** Last breadcrumb entry has no `[n]` hotkey indicator
- **AC6** All existing tests pass; new tests cover each fix

## Tasks

- [x] **T1** Fix count URL to include filter params (`commands.go`)
- [x] **T2** Fix ID column width to account for drilldown "▶ " prefix (`table.go`)
- [x] **T3** Configurable `title_attribute` on DrillDownDef; update `o8n-cfg.yaml`
- [x] **T4** Move row status colors into StyleSet; remove hardcoded colors from `table.go`
- [x] **T5** Remove `[n]` hotkey from last breadcrumb entry (`view.go`)

## Dev Agent Record

### Decisions
- T3 fallback: when `title_attribute` not set, use value from first visible column of parent row (the ID column equivalent) — keeps it dynamic without needing the raw row data at drilldown time
- T4: add `RowRunning/Suspended/Failed/Ended` to StyleSet; `colorizeRows` receives `RowStyles` struct
- T2: in `buildColumnsFor`, if table has drilldown, add 2 to first column's desired width for "▶ " prefix

## File List

- `internal/config/config.go` — TitleAttribute field on DrillDownDef
- `internal/app/commands.go` — filter params on count URL
- `internal/app/table.go` — drilldown prefix width, skin-based row status colors
- `internal/app/styles.go` — RowRunning/Suspended/Failed/Ended in StyleSet
- `internal/app/update.go` — contentHeader uses title_attribute
- `internal/app/view.go` — last breadcrumb no hotkey
- `o8n-cfg.yaml` — title_attribute: processDefinitionKey on process-instance drilldown
- `internal/app/process_instance_fixes_test.go` — new test file
