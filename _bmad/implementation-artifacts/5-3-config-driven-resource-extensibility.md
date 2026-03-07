# Story 5.3: Config-Driven Resource Extensibility

Status: ready-for-dev

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

- [ ] **Generic API Dispatcher (AC: 1, 3)**
  - [ ] Update `internal/client/client.go` to include a `ExecuteGenericAction(method, path string, body interface{})` method.
  - [ ] Implement placeholder interpolation logic (e.g., replacing `{id}` with the actual resource ID).
- [ ] **Dynamic Table Mapping (AC: 1)**
  - [ ] Verify `internal/app/table.go` correctly iterates over `TableDef.Columns` from the config to render rows.
  - [ ] Ensure the mapping from API JSON response to table columns uses the `key` field from the column config.
- [ ] **Dynamic Action Handling (AC: 1, 3)**
  - [ ] Update `internal/app/commands.go` to support a generic `executeActionCmd` that reads action details from the resource configuration.
  - [ ] Ensure that actions with `type: mutation` trigger the generic dispatcher.
- [ ] **OpenAPI Client Generation (AC: 2)**
  - [ ] Verify that `generate-api-client.sh` is functional and correctly places files in `internal/operaton/`.
- [ ] **Verification & Validation**
  - [ ] Add a "dummy" resource type to `o6n-cfg.yaml` (e.g., `test-deployments`) and verify it appears and functions in the UI.
  - [ ] Assert that no new `case` statements for specific resource types were added to `Update()` or `View()`.

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

Gemini 2.0 Flash

### Debug Log References

### Completion Notes List

### File List
