# Story 5.2: Credential Security

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want **credentials to be isolated in a git-ignored, permission-restricted file and never appear in logs, debug output, or clipboard operations**,
so that **sensitive API credentials cannot leak into version control or observability tooling**.

## Acceptance Criteria

1. **Given** `o6n-env.yaml` contains credentials
   **When** the file is checked in the workspace
   **Then** it is explicitly listed in `.gitignore`.
   **And** the application enforces `chmod 600` (read/write by owner only) permissions on the file at load time, warning the user if permissions are too open.

2. **Given** the application is running with the `--debug` flag
   **When** debug output is written to `./debug/o6n.log` and `./debug/last-screen.txt`
   **Then** no credentials (username, password, or API URLs with embedded basic auth) are present in any log or screen dump.

3. **Given** the operator uses the JSON viewer (`J`) or direct copy (`Ctrl+J`)
   **When** the JSON content is displayed or copied to the clipboard
   **Then** all credential-related fields from the environment configuration (username, password) are absent from the exported data.
   **And** only the resource data itself is present.

4. **Given** an API error occurs (e.g., 401 Unauthorized)
   **When** the error message is rendered in the footer
   **Then** the message contains the HTTP status and API response error, but no credential values or auth headers.

## Tasks / Subtasks

- [x] **Filesystem Security (AC: 1)**
  - [x] `o6n-env.yaml` is in `.gitignore` (line 34 confirmed)
  - [x] `LoadEnvConfig` now checks file permissions and sets `config.PermissionWarning` on Unix (review fix)
  - [x] `run.go` logs the permission warning and sets it as a footer message at startup (review fix)
- [x] **Log & Debug Scrubbing (AC: 2)**
  - [x] Audited `update.go` and `run.go` — no raw config structs are logged via `%+v`
  - [x] `friendlyError` in `util.go` maps network errors to safe messages; raw error only shown for unknown errors (acceptable — HTTP client uses BasicAuth context, not URL-embedded credentials)
- [x] **Clipboard & Viewer Sanitization (AC: 3)**
  - [x] `ModalJSONView` uses `m.rowData[cursor]` (resource data only) — verified in `view.go`
  - [x] `Ctrl+J` copy uses same `m.rowData` source — credentials never in row data
- [x] **Error Message Sanitization (AC: 4)**
  - [x] `friendlyError()` in `util.go` translates network errors to env-name-only messages (no credentials)
  - [x] HTTP client uses `operaton.ContextBasicAuth` — credentials not in URL or error strings
- [x] **Verification & Tests**
  - [x] `TestLoadEnvConfig_WarnOnTooOpenPermissions` and `TestLoadEnvConfig_NoWarnOnRestrictedPermissions` added to `config_test.go` (review fix)

## Dev Notes

### Architecture Compliance
- **NFR8:** Credentials (username, password) must never appear in log files, debug output, or clipboard operations.
- **NFR9:** `o6n-env.yaml` must be git-ignored and maintained at `chmod 600` file permissions at all times.

### Project Structure Notes
- **Config Loader:** `internal/config/loader.go` — handle permission checks here.
- **JSON Export:** `internal/app/view.go` (viewer) and `internal/app/update.go` (copy cmd).
- **Error Handling:** `internal/app/update.go` (handling `errMsg`).

### References
- [Source: `_bmad/planning-artifacts/epics.md#Story 5.2`]
- [Source: `_bmad/planning-artifacts/prd.md#NFR8`, `NFR9`]
- [Source: `_bmad/planning-artifacts/architecture.md#Authentication & Security`]

## Dev Agent Record

### Agent Model Used

Claude Haiku 4.5

### Debug Log References

### Senior Developer Review (AI)

- **Outcome:** Fixed — HIGH issues resolved
- **Date:** 2026-03-08
- **Reviewer:** Claude Sonnet 4.6 (adversarial code review)
- **Action items taken:**
  - HIGH: `LoadEnvConfig` did NOT check file permissions (AC 1 not implemented) — FIXED: added `os.Stat` + mode check, sets `config.PermissionWarning` global; `run.go` now logs warning and shows in footer
  - HIGH: All tasks marked `[ ]` — FIXED: tasks updated to `[x]`
  - HIGH: No tests for permission checking — FIXED: 2 new tests in `config_test.go`
  - MEDIUM: `friendlyError` default case returns raw error string — reviewed as acceptable (HTTP client does not embed credentials in error messages)
  - PASS: `.gitignore` confirmed to have `o6n-env.yaml`
  - PASS: JSON viewer / Ctrl+J use `m.rowData` which contains only resource data, not env config

### Completion Notes List

- Story implementation partially existed; permission check was missing.
- Review fix: `LoadEnvConfig` now checks permissions and sets `PermissionWarning`.
- Review fix: `run.go` surfaces permission warning at startup.
- 2 new permission tests added to `config_test.go`.

### File List

- `internal/config/config.go` — FIX: `LoadEnvConfig` checks permissions; `PermissionWarning` global added
- `internal/app/run.go` — FIX: surfaces `config.PermissionWarning` as footer message at startup
- `internal/config/config_test.go` — FIX: 2 new permission check tests
- `.gitignore` — EXISTING: `o6n-env.yaml` listed (line 34)
- `internal/app/util.go` — EXISTING: `friendlyError` credential-safe error mapping

