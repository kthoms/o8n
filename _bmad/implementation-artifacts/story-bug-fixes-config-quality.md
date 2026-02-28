# Story: Bug Fixes and Config Quality

## Summary

Five targeted fixes identified by intensive quality review: remove an unconditional DEBUG log that leaks to disk on every startup; fix the Delegate Task action whose empty body causes a 400 API error; add `search_param` to key tables so Ctrl+A server-side search actually works; add edit capability to the `variable-instance` table which is currently read-only despite being identical to the already-editable `process-variables`; and clear `searchMode` in `prepareStateTransition` to prevent the inline search bar from leaking into child resource views.

## Motivation

1. **DEBUG log always on** — `run.go:82` calls `log.Printf("DEBUG: skinName resolved to: %s", skinName)` unconditionally. Every startup writes noise to `debug/o8n.log` regardless of `--debug`. Violates the intent of the debug flag.

2. **Delegate Task fails silently** — `o8n-cfg.yaml:303` configures `body: '{}'` for `POST /task/{id}/delegate`. The Operaton REST API requires `{"userId": "..."}` in the body — an empty object returns HTTP 400/422. The claim action correctly uses `{"userId": "{currentUser}"}`. Delegate should do the same (delegate to current user, i.e. self-delegate which is the common reassignment pattern).

3. **Ctrl+A server search non-functional** — `update.go:1408-1417` and `commands.go` fully implement server-side search triggered by Ctrl+A when `def.SearchParam != ""`. But `search_param` is absent from every table in `o8n-cfg.yaml`. The feature shows a "no search param" message for every resource. Key tables need this wired up.

4. **variable-instance is permanently read-only** — The `variable-instance` table shows in task drilldowns and process-instance drilldowns. It has three columns (name, value, processInstanceId) and zero actions or edit_action. The `process-variables` table already has a working `edit_action` pattern. Variable instances need the same treatment via `PUT /variable-instance/{id}`.

5. **searchMode leaks on navigation** — `transition.go:prepareStateTransition()` clears `searchTerm`, `popup.mode`, `originalRows` but never clears `m.searchMode` or blurs `m.searchInput`. If inline search is active (e.g. legacy `searchMode = true` path via `update.go:753`) when the user drills down, the search bar stays visible on the child resource.

## Acceptance Criteria

### Fix 1: Gate DEBUG log

- [ ] **AC-1:** `run.go:82` — the `log.Printf("DEBUG: skinName resolved to: %s", skinName)` line is removed or moved inside a `if *debug {` block. No debug log on standard startup.

### Fix 2: Delegate Task body

- [ ] **AC-2:** In `o8n-cfg.yaml`, the Delegate Task action for the `task` table changes from `body: '{}'` to `body: '{"userId": "{currentUser}"}'`. The `{currentUser}` placeholder is already resolved by the generic action executor (same path as `c`/claim).
- [ ] **AC-3:** A test verifies that sending `d` on a task table row dispatches a generic action with a body containing the current username.

### Fix 3: search_param for key tables

- [ ] **AC-4:** `o8n-cfg.yaml` — the following tables gain a `search_param` entry:
  - `process-definition` → `search_param: nameLike`
  - `process-instance` → `search_param: businessKeyLike`
  - `task` → `search_param: nameLike`
  - `user` → `search_param: idLike`
  - `group` → `search_param: idLike`
  - `deployment` → `search_param: nameLike`
  - `incident` → `search_param: incidentMessage` *(best available)*
- [ ] **AC-5:** Pressing `Ctrl+A` on each of those tables triggers a server-side fetch with the configured param; the existing code path already handles this.
- [ ] **AC-6:** Tables that have no useful text-search API param remain without `search_param` (e.g. `authorization`, `batch`, history tables). No regression.

### Fix 4: variable-instance edit_action

- [ ] **AC-7:** `o8n-cfg.yaml` — `variable-instance` table gains an `edit_action`:
  ```yaml
  edit_action:
    method: PUT
    path: /variable-instance/{id}
    body_template: '{"value": {value}, "type": "{type}"}'
    name_column: id
  ```
- [ ] **AC-8:** The `value` column in `variable-instance` is marked `editable: true` with `input_type: auto`.
- [ ] **AC-9:** Pressing `e` on a variable-instance row opens the edit modal with the current value pre-populated. Saving calls `PUT /variable-instance/{id}`.
- [ ] **AC-10:** A test verifies the edit modal opens for a variable-instance table row.

### Fix 5: searchMode cleared on navigation

- [ ] **AC-11:** `transition.go:prepareStateTransition()` adds `m.searchMode = false` and `m.searchInput.Blur()` to the existing search-state cleanup block (alongside the existing `m.searchTerm = ""` at line 26).
- [ ] **AC-12:** A test verifies that after `transitionDrilldown`, `m.searchMode == false`.

## Tasks

### Task 1: Config fixes (o8n-cfg.yaml)

**Files:** `o8n-cfg.yaml`

1. Fix Delegate Task body: `body: '{}'` → `body: '{"userId": "{currentUser}"}'`
2. Add `search_param` to: `process-definition`, `process-instance`, `task`, `user`, `group`, `deployment`, `incident`
3. Add `edit_action` and mark `value` editable in `variable-instance` table

### Task 2: Code fix — DEBUG log

**Files:** `internal/app/run.go`

1. Wrap `log.Printf("DEBUG: skinName resolved to: %s", skinName)` in `if *debug { ... }` or remove it

### Task 3: Code fix — searchMode in transition

**Files:** `internal/app/transition.go`

1. In the search-state cleanup block (after `m.searchTerm = ""`), add:
   ```go
   m.searchMode = false
   m.searchInput.Blur()
   ```

### Task 4: Tests

**Files:** `internal/app/` (new or existing test file, e.g. `internal/app/config_quality_test.go`)

1. AC-3: Delegate action body contains currentUser — send `d` key on task table, verify cmd dispatched with correct body
2. AC-10: Edit modal opens for variable-instance table — set up model with variable-instance table key, press `e`, verify `activeModal == ModalEdit`
3. AC-12: searchMode cleared on drilldown — set `m.searchMode = true`, call `prepareStateTransition(transitionDrilldown)`, assert `m.searchMode == false`

## Dev Notes

- `{currentUser}` resolution: `resolveActionBody()` in `commands.go` already handles this via `strings.ReplaceAll(body, "{currentUser}", username)` — no code change needed for Fix 2.
- `variable-instance` edit uses `id` as the identifier column (not `name`) because variable instances have a proper `id` field unlike process-variables which use variable name as key.
- `search_param` values are Operaton REST query parameter names. `nameLike` appends `%` wildcards server-side in Operaton. Verify against Operaton REST API docs for each resource.
- `searchMode` is legacy inline-search state; the current UI uses popup search. The fix ensures it can't get stuck `true` through navigation.

## File List

- `o8n-cfg.yaml` — delegate body fix, search_param additions, variable-instance edit_action
- `internal/app/run.go` — gate DEBUG log
- `internal/app/transition.go` — clear searchMode
- `internal/app/config_quality_test.go` (new) — tests for AC-3, AC-10, AC-12

## Change Log

- 2026-02-28: Story created (party-mode quality review — 5 bug/config fixes)

## Status

ready
