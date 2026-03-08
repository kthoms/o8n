# Story 5.3: Config-Driven Resource Extensibility

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **contributor** (Marco persona),
I want to **add a new standard resource type by editing only `o6n-cfg.yaml`**,
so that **the community can extend o6n's resource coverage without Go source code changes**.

## Acceptance Criteria

1. **Given** a contributor adds a new table entry to `o6n-cfg.yaml` with columns, actions, and drilldown rules
   **When** the application is built and started
   **Then** the new resource type appears in the context switcher (`:`)
   **And** the table loads with the defined columns by mapping API response fields to column keys.
   **And** configured actions execute the correct API calls by interpolating placeholders (e.g., `{id}`) in the path.
   **And** configured drilldown rules navigate to the child resource type with the correct filter.
   **And** no Go source file was modified to achieve this.

2. **Given** `internal/operaton/` needs updating for a new API endpoint
   **When** the contributor regenerates the client
   **Then** `.devenv/scripts/generate-api-client.sh` produces the updated client without manual file edits to `internal/operaton/`.

3. **Given** a new resource type is added in `o6n-cfg.yaml`
   **When** an action is executed on it
   **Then** the `internal/client/client.go` generic dispatcher handles the request based on the verb and path specified in the config.

## Tasks / Subtasks

- [x] **Generic API Dispatcher (AC: 1, 3)**
  - [x] `client.ExecuteAction(method, path, body string) error` in `client.go` (line 689)
  - [x] Placeholder interpolation via `resolvePathParams` in `update.go` (line 296); `{id}`, `{name}`, `{parentId}`, `{value}`, `{type}` substituted in action paths
- [x] **Dynamic Table Mapping (AC: 1)**
  - [x] `buildColumnsFor` in `table.go` iterates `TableDef.Columns` from config
  - [x] `genericLoadedMsg` handler in `update.go` maps API JSON response keys to configured column names
- [x] **Dynamic Action Handling (AC: 1, 3)**
  - [x] `executeActionCmd(action config.ActionDef, resolvedPath string)` in `commands.go` (line 41)
  - [x] Actions of any HTTP verb (POST/PUT/DELETE/GET) dispatched via `ExecuteAction`
- [x] **OpenAPI Client Generation (AC: 2)**
  - [x] `.devenv/scripts/generate-api-client.sh` exists; `internal/operaton/` is auto-generated
- [x] **Verification & Validation**
  - [x] No resource-specific `switch` cases in `Update()` or `View()` for standard resources (verified)
  - [x] Context switcher (`:`), table rendering, and actions all driven from `o6n-cfg.yaml` config

## Dev Notes

### Previous Story Intelligence
- **Story 5.1 & 5.2 Learnings:** Environment configuration is fully isolated. Generic actions must respect the `api_url` of the active environment.
- **Architecture Contract:** All standard resources must follow the same lifecycle: load → render → act. Tailored Go logic is only for non-standard UI (like Task Complete dialog).

### Architecture Compliance
- **NFR13:** Standard resource type addition requires only `o6n-cfg.yaml` edits.
- **NFR14:** `internal/operaton/` remains auto-generated, never manually edited.
- **Config Authority:** `o6n-cfg.yaml` is the single source of truth for resources.

### Anti-Patterns to Avoid (LLM Guardrails)
- **❌ DO NOT** add resource-specific `switch` cases in `update.go` or `view.go` for standard resources.
- **❌ DO NOT** manually edit any file under `internal/operaton/`.
- **❌ DO NOT** hardcode API paths that are already defined in the config.

### Project Structure Notes
- **Config Management:** `internal/config/`.
- **API Client:** `internal/client/` and `internal/operaton/`.
- **TUI Logic:** `internal/app/`.

### References
- [Source: `_bmad/planning-artifacts/epics.md#Story 5.3`]
- [Source: `_bmad/planning-artifacts/prd.md#FR27`, `FR28`]
- [Source: `_bmad/planning-artifacts/architecture.md#Maintainability`]

## Dev Agent Record

### Agent Model Used

Claude Haiku 4.5

### Debug Log References

### Senior Developer Review (AI)

- **Outcome:** Approved with task-checkbox fixes
- **Date:** 2026-03-08
- **Reviewer:** Claude Sonnet 4.6 (adversarial code review)
- **Action items taken:**
  - HIGH: All tasks marked `[ ]` despite full implementation — FIXED: tasks updated to `[x]`
  - Verified: `client.ExecuteAction` is the generic dispatcher
  - Verified: Placeholder interpolation (`{id}`, `{name}`, etc.) in `resolvePathParams`
  - Verified: No resource-specific switch statements in `Update()` or `View()` for standard resource handling
  - LOW: No dedicated test for a "dummy" resource type — accepted (the config-driven path is exercised by existing generic tests)

### Completion Notes List

- Story implementation verified and tested.
- All acceptance criteria addressed.
- Generic dispatcher (`ExecuteAction`) and placeholder interpolation confirmed present.

### File List

- `internal/client/client.go` — EXISTING: `ExecuteAction(method, path, body)` generic dispatcher
- `internal/app/commands.go` — EXISTING: `executeActionCmd` wrapping generic dispatcher
- `internal/app/update.go` — EXISTING: `resolvePathParams` placeholder interpolation; `genericLoadedMsg` column mapping
- `internal/app/table.go` — EXISTING: `buildColumnsFor` config-driven column layout
- `.devenv/scripts/generate-api-client.sh` — EXISTING: OpenAPI client generation script

