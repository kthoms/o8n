# Story: Configuration-Driven Application Flow

**Status**: Draft  
**Priority**: High  

---

## Problem Statement

Large portions of o8n's navigation, data fetching, and edit/save logic are hardcoded in Go source files rather than driven by `o8n-cfg.yaml`. This limits extensibility and creates bugs. The goal is to make the full application flow — fetch, drilldown, save, display — derive from configuration.

---

## Analysis Findings

### Critical Bug
- `fetchGenericCmd` (commands.go:382) hardcodes the count URL to `/process-instance/count` for **all** generic tables. Every non-process-instance table shows the wrong total count.
- `fetchInstancesCmd` (commands.go:175) also hardcodes `/process-instance/count` — correct for instances but not extensible.

### Hardcoded Resource Logic (20+ locations)
| Location | Hardcoded Assumption |
|---|---|
| `commands.go:fetchDefinitionsCmd` | Only fetches `/process-definition` via typed `FetchProcessDefinitions()` |
| `commands.go:fetchInstancesCmd` | Hardcodes `/process-instance` URL, specific struct mapping |
| `commands.go:fetchVariablesCmd` | Hardcodes variables fetch via `FetchVariables()` |
| `commands.go:fetchForRoot` | switch/case on 3 named resources |
| `update.go:750–870` | Drilldown dispatch: 2 hardcoded special cases for `process-instance` and `process-variables` before the generic `default:` |
| `update.go:871–930` | Fallback drilldown: hardcoded definitions→instances→variables chain |
| `edit.go:isVariableTable` | Edit/save only works for variable table names |
| `model.go/nav.go/table.go` | viewMode strings "definitions"/"instances"/"variables" in ~20 places |

### Missing Config Fields
| Field | Impact |
|---|---|
| `TableDef.ApiPath` | History tables use `/history/xxx` but name is `history-xxx`; mismatch causes 404 |
| `TableDef.CountPath` | Required to fix count bug generically |
| `TableDef.EditAction` | Can only save variables; all other tables read-only despite having writable REST endpoints |
| `DrillDownDef.Label` | Breadcrumb shows raw target name instead of a readable label |

### REST API Enrichment Opportunities (from OpenAPI spec)
- Every list endpoint has a corresponding `/count` endpoint — should be used
- All history endpoints use `/history/xxx` path (not matching table name)
- Write operations currently missing or incomplete in o8n-cfg.yaml:
  - `DELETE /deployment/{id}` (undeploy)
  - `POST /deployment/{id}/redeploy`
  - `DELETE /process-definition/{id}`
  - `POST /process-definition/{id}/restart`
  - `DELETE /incident/{id}` (resolve)
  - `PUT /incident/{id}/annotation` (annotate)
  - `PUT /job-definition/{id}/retries` (set bulk retries)
  - `PUT /history/user-operation/{id}/set-annotation`
  - `DELETE /history/process-instance/{id}`
  - `PUT /decision-definition/{id}/history-time-to-live`

---

## Proposed Solution

### 1. Config Struct Changes (`internal/config/config.go`)

Add to `TableDef`:
```go
ApiPath    string         `yaml:"api_path,omitempty"`    // REST path; defaults to /{name}
CountPath  string         `yaml:"count_path,omitempty"`  // count endpoint; defaults to {api_path}/count
EditAction *EditActionDef `yaml:"edit_action,omitempty"` // generic save config
```

New struct `EditActionDef`:
```go
type EditActionDef struct {
    Method       string `yaml:"method"`                  // PUT, POST, PATCH
    Path         string `yaml:"path"`                    // e.g. /process-instance/{parentId}/variables/{name}
    BodyTemplate string `yaml:"body_template"`           // JSON with {value}, {type} placeholders
    IDColumn     string `yaml:"id_column,omitempty"`     // column for {id} (default: "id")
    NameColumn   string `yaml:"name_column,omitempty"`   // column for {name} (default: "name")
}
```

Add to `DrillDownDef`:
```go
Label string `yaml:"label,omitempty"` // breadcrumb label (defaults to target)
```

### 2. `fetchGenericCmd` Generalization (`commands.go`)

- Use `def.ApiPath` when set (fall back to `/{name}`)
- Count URL: use `def.CountPath` when set, else `apiPath + "/count"`
- Remove special-case `fetchDefinitionsCmd`, `fetchInstancesCmd`, `fetchVariablesCmd` — replace with `fetchGenericCmd` for all tables
- `fetchForRoot` becomes: always call `fetchGenericCmd(root)` (no switch/case)

### 3. Drilldown Unification (`update.go`)

- Remove the two hardcoded special cases (`process-instance` and `process-variables`) in the drilldown switch — the generic `default:` case already handles them correctly
- Remove the entire "fallback hard-coded drill behaviour" block (lines 871–930) — replace with: if no drilldown is configured, show an info message "no drilldown configured"
- The `viewMode` is simplified: always equals `currentRoot` (table name). Remove semantic strings "definitions"/"instances"/"variables".

### 4. Generic Edit/Save (`edit.go`, `update.go`, `commands.go`)

- Remove `isVariableTable()` hardcoding
- When `TableDef.EditAction` is set, save using a new `executeEditActionCmd`:
  - Resolve `{id}` from row's ID column
  - Resolve `{name}` from row's name column  
  - Resolve `{parentId}` from `m.selectedInstanceID`
  - Resolve `{value}` and `{type}` from the edited cell
- Keep `setVariableCmd` as the legacy path for now, triggered via `edit_action` in `process-variables` config
- Process variables table gets an explicit `edit_action` in config:
  ```yaml
  edit_action:
    method: PUT
    path: /process-instance/{parentId}/variables/{name}
    body_template: '{"value": {value}, "type": "{type}"}'
    name_column: name
  ```

### 5. o8n-cfg.yaml Enrichment

For every table:
- Add `api_path` where the REST path differs from the table name
- Add `count_path` explicitly
- Add missing actions from REST API
- Add `edit_action` for variable table
- Enrich drilldown `label` fields

Key table corrections:
```yaml
- name: history-process-instance
  api_path: /history/process-instance
  count_path: /history/process-instance/count

- name: history-activity-instance
  api_path: /history/activity-instance
  count_path: /history/activity-instance/count

- name: process-variables
  api_path: /process-instance/{parentId}/variables  # special: uses parentId
  edit_action:
    method: PUT
    path: /process-instance/{parentId}/variables/{name}
    body_template: '{"value": {value}, "type": "{type}"}'
```

---

## Tasks

### Task 1: Config struct additions
- [ ] Add `ApiPath`, `CountPath` to `TableDef`
- [ ] Add `EditActionDef` struct
- [ ] Add `EditAction *EditActionDef` to `TableDef`
- [ ] Add `Label` to `DrillDownDef`
- [ ] Update `internal/config/config_test.go` for new fields

### Task 2: Fix count URL bug
- [ ] In `fetchGenericCmd`: derive count URL from `def.CountPath` or `apiPath + "/count"` (not hardcoded)
- [ ] In `fetchInstancesCmd`: same fix
- [ ] Add unit test confirming count endpoint URL is correct per table

### Task 3: Generalize all fetches
- [ ] Use `def.ApiPath` (falling back to `/{name}`) in `fetchGenericCmd`
- [ ] Make `fetchForRoot` always use `fetchGenericCmd` (remove switch/case)
- [ ] Deprecate `fetchDefinitionsCmd`, `fetchInstancesCmd`, `fetchVariablesCmd` — route through `fetchGenericCmd`
- [ ] Update `Init()` restore logic accordingly

### Task 4: Unify drilldown dispatch
- [ ] Remove hardcoded `case "process-instance"` and `case "process-variables"` from drilldown switch
- [ ] Remove fallback drilldown block (lines 871–930 in update.go)
- [ ] Simplify `viewMode` — use `currentRoot` as the mode everywhere, removing the semantic strings
- [ ] Ensure `applyDefinitions`/`applyInstances`/`applyVariables` are replaced by `applyGeneric` or routed through it
- [ ] Fix all `nav.go` / `table.go` references to the old viewMode strings

### Task 5: Generic edit/save
- [ ] Add `executeEditActionCmd` to `commands.go` — resolves path/body templates
- [ ] Remove `isVariableTable()` check from `update.go` edit save handler
- [ ] When `TableDef.EditAction != nil`, use `executeEditActionCmd`
- [ ] Add `edit_action` to `process-variables` in `o8n-cfg.yaml`
- [ ] Test: editing a variable value still saves correctly

### Task 6: o8n-cfg.yaml enrichment
- [ ] Add `api_path` and `count_path` to all history tables
- [ ] Add `api_path` to any other table where name ≠ REST path
- [ ] Add `label` to all `drilldown` entries
- [ ] Add missing write actions:
  - deployment: delete, redeploy
  - process-definition: delete, restart, set-history-ttl
  - incident: delete (resolve), set-annotation
  - job-definition: set-retries
  - history/user-operation: set-annotation
  - history/process-instance: delete
  - decision-definition: evaluate, set-history-ttl

### Task 7: Tests + docs
- [ ] Update all tests that reference old viewMode strings
- [ ] Add generic fetch test confirming correct api_path used
- [ ] Add edit action test for generic table
- [ ] Update specification.md to document new config fields
- [ ] Update README.md

---

## Config Examples (o8n-cfg.yaml changes)

### EditAction for variables:
```yaml
- name: process-variables
  api_path: /process-instance/{parentId}/variables
  edit_action:
    method: PUT
    path: /process-instance/{parentId}/variables/{name}
    body_template: '{"value": {value}, "type": "{type}"}'
    name_column: name
  columns:
    - name: name
      ...
    - name: value
      editable: true
      input_type: auto
```

### History table with explicit paths:
```yaml
- name: history-process-instance
  api_path: /history/process-instance
  count_path: /history/process-instance/count
  columns: [...]
  drilldown:
    - target: history-activity-instance
      param: processInstanceId
      column: id
      label: Activity Instances
  actions:
    - key: ctrl+d
      label: Delete History
      method: DELETE
      path: /history/process-instance/{id}
      confirm: true
```

### Drilldown with label:
```yaml
drilldown:
  - target: process-instance
    param: processDefinitionId
    column: id
    label: Instances
  - target: history-process-instance
    param: processDefinitionId
    column: id
    label: History
```

---

## Breaking Changes / Risks

- Tests asserting `m.viewMode == "definitions"` etc. will need updating
- `applyDefinitions` / `applyInstances` / `applyVariables` public functions (if referenced in tests)
- State file `o8n-stat.yml` stores `viewMode` in `nav.breadcrumb` — compatible as long as breadcrumb entries match table names
- Variable save must be thoroughly regression-tested (most used write feature)

---

## Out of Scope for this Story
- Debug logging category system (separate story)
- Search/filter UI within a table
- BPMN diagram viewer
