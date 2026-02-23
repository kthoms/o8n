# Story: Configuration-Driven Application Flow

**Story Key**: config-driven-flow  
**Status**: ready-for-dev  
**Priority**: High  

---

## Story

As a developer extending o8n,  
I want the application flow — fetch, drilldown, and save — to be fully driven by `o8n-cfg.yaml`,  
So that adding new Operaton REST resources requires only config changes, not Go code changes.

---

## Acceptance Criteria

- **AC1**: `TableDef` in `internal/config/config.go` has `ApiPath`, `CountPath`, and `EditAction` fields; `DrillDownDef` has `Label`.
- **AC2**: `fetchGenericCmd` uses `def.ApiPath` (falling back to `/{name}`) and the correct count URL (`def.CountPath` or `{api_path}/count`), NOT the hardcoded `/process-instance/count`.
- **AC3**: `fetchForRoot` always delegates to `fetchGenericCmd` — no special-case switch for named resources.
- **AC4**: Drilldown dispatch in `update.go` has no hardcoded `case "process-instance"` or `case "process-variables"` — all drilldown targets handled by the generic `default:` path.
- **AC5**: The fallback hard-coded `definitions→instances→variables` drill chain is removed.
- **AC6**: Edit/save uses `TableDef.EditAction` when present; `isVariableTable()` hardcoding is removed from the save path.
- **AC7**: `o8n-cfg.yaml` has `api_path` and `count_path` on all tables where the REST path differs from the table name (all history tables + process-variables).
- **AC8**: `process-variables` table has `edit_action` in config; variable value editing works as before.
- **AC9**: At least 5 new write actions added to `o8n-cfg.yaml` (deployment delete/redeploy, process-definition delete, incident delete/annotate).
- **AC10**: All existing tests pass; new tests cover the count URL fix and generic save.

---

## Tasks/Subtasks

### Task 1: Config struct additions
- [x] 1.1 Add `ApiPath string` and `CountPath string` to `TableDef` in `internal/config/config.go`
- [x] 1.2 Add `EditActionDef` struct (`Method`, `Path`, `BodyTemplate`, `IDColumn`, `NameColumn`)
- [x] 1.3 Add `EditAction *EditActionDef` to `TableDef`
- [x] 1.4 Add `Label string` to `DrillDownDef`
- [x] 1.5 Write/update unit tests for new config fields (marshal/unmarshal round-trip)

### Task 2: Fix count URL bug + generalize api_path
- [x] 2.1 In `fetchGenericCmd`: resolve `apiPath` from `def.ApiPath` (fallback `/{name}`); count URL from `def.CountPath` (fallback `apiPath + "/count"`)
- [x] 2.2 In `fetchInstancesCmd`: apply same count URL fix using process-instance table def
- [x] 2.3 Add unit test: `fetchGenericCmd` uses the table's `CountPath` when set; falls back correctly
- [x] 2.4 Add unit test: `fetchGenericCmd` uses `ApiPath` when set

### Task 3: Generalize fetchForRoot
- [x] 3.1 Replace `fetchForRoot` switch/case with single `fetchGenericCmd(root)` call
- [x] 3.2 Remove special handling for `process-definitions`, `process-instances`, `process-variables` from `fetchForRoot`
- [x] 3.3 Update `Init()` restore logic to use `fetchGenericCmd`
- [x] 3.4 Verify existing fetch tests still pass

### Task 4: Unify drilldown dispatch
- [x] 4.1 Remove hardcoded `case "process-instance", "process-instances":` from drilldown switch in `update.go`
- [x] 4.2 Remove hardcoded `case "process-variables", "variables", "variable-instance", "variable-instances":` from drilldown switch
- [x] 4.3 Remove the fallback hard-coded drilldown block (`if m.viewMode == "definitions"` ... `else if m.viewMode == "instances"`)
- [x] 4.4 Update `viewMode` handling: remove semantic strings "definitions"/"instances"/"variables"; use table name as mode (or simplify to just `currentRoot`)
- [x] 4.5 Update all `nav.go`, `table.go`, `model.go` references to old viewMode strings
- [x] 4.6 Verify drilldown tests pass (drilldown_test.go)

### Task 5: Generic edit/save
- [x] 5.1 Add `executeEditActionCmd` to `commands.go`: resolves `{id}`, `{name}`, `{parentId}`, `{value}`, `{type}` in path and body template
- [x] 5.2 In `update.go` edit save handler: if `TableDef.EditAction != nil` → use `executeEditActionCmd`; keep `setVariableCmd` as fallback only
- [x] 5.3 Remove `isVariableTable()` check from main save dispatch path in `update.go`
- [x] 5.4 Test: editing a variable value still saves correctly with the generic path
- [x] 5.5 Test: generic edit action resolves path template placeholders correctly

### Task 6: Enrich o8n-cfg.yaml
- [x] 6.1 Add `api_path` and `count_path` to all history tables (12 tables)
- [x] 6.2 Add `api_path: /process-instance/{parentId}/variables` and `edit_action` to `process-variables` table
- [x] 6.3 Add drilldown `label` fields to all drilldown entries
- [x] 6.4 Add missing write actions: `job-definition` (set-retries), `incident` (set-annotation), `decision-definition` (delete)
- [x] 6.5 Verify o8n-cfg.yaml loads without error after changes (`go test ./internal/config/...`)

### Task 7: Tests and docs
- [x] 7.1 Run full test suite; fix any regressions from viewMode string removal
- [x] 7.2 Update `specification.md`: document `api_path`, `count_path`, `edit_action`, `DrillDownDef.label`
- [x] 7.3 Commit with message `feat: config-driven application flow`

---

## Dev Notes

### Critical Bug Context
`fetchGenericCmd` (commands.go:382) has:
```go
countURL := base + "/process-instance/count"
```
This is hardcoded for ALL generic tables — must be changed to use `def.CountPath || (apiPath + "/count")`.

### ViewMode Simplification
Current semantic strings: `"definitions"`, `"instances"`, `"variables"` — appear in 20+ places across `model.go`, `update.go`, `nav.go`, `table.go`. Strategy: `viewMode` becomes `currentRoot` (the table name). All conditionals that check `viewMode == "definitions"` become `currentRoot == "process-definition"` or use the config to determine behavior. The `applyDefinitions`/`applyInstances`/`applyVariables` functions remain but are dispatched generically.

### EditActionDef Template Variables
- `{id}` → from row's `IDColumn` (default "id")
- `{name}` → from row's `NameColumn` (default "name")  
- `{parentId}` → `m.selectedInstanceID` (navigation context)
- `{value}` → the edited cell value
- `{type}` → the Operaton type name (for variables: "String", "Integer", etc.)

### process-variables edit_action in config:
```yaml
- name: process-variables
  api_path: /process-instance/{parentId}/variables
  edit_action:
    method: PUT
    path: /process-instance/{parentId}/variables/{name}
    body_template: '{"value": {value}, "type": "{type}"}'
    name_column: name
```

### History Table API Paths
All history tables need `api_path: /history/<name-without-history-prefix>` — e.g., `history-process-instance` → `api_path: /history/process-instance`.

### Drilldown Special Cases to Remove
In `update.go` around line 758:
- Remove `case "process-instance", "process-instances":` block (~30 lines)  
- Remove `case "process-variables", "variables", ...:` block (~35 lines)
- These work fine through the generic `default:` path that calls `fetchGenericCmd(chosen.Target)`

### Fallback Drilldown to Remove
Around `update.go:871` — the block:
```go
if m.viewMode == "definitions" { ... fetchInstancesCmd ...
} else if m.viewMode == "instances" { ... fetchVariablesCmd ...
```
This bypasses config-driven drilldown for `process-definition` tables. Remove it — `process-definition` config already has drilldown entries.

### Fetch Function Deprecation Plan
- `fetchDefinitionsCmd` → routes through `fetchGenericCmd("process-definition")`
- `fetchInstancesCmd` → routes through `fetchGenericCmd("process-instance")` with genericParams
- `fetchVariablesCmd` → routes through `fetchGenericCmd("process-variables")` with parentId in path
- Keep the functions as thin wrappers to avoid breaking test references, then eventually remove

### Pre-existing Test Failure
`TestGenericFetchUsesTableDefAndCount` — pre-existing failure unrelated to this story (count endpoint response handling bug — this story actually fixes the root cause).

---

## Dev Agent Record

### Implementation Plan
*(to be filled during implementation)*

### Completion Notes
*(to be filled on completion)*

---

## File List

*(to be updated as files are changed)*

---

## Change Log

*(to be updated on completion)*
