# Story 5.2: Credential Security

Status: ready-for-dev

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

- [ ] **Filesystem Security (AC: 1)**
  - [ ] Verify `o6n-env.yaml` is in `.gitignore`.
  - [ ] Update `LoadEnvConfig` in `internal/config/loader.go` to check file permissions on Unix systems and issue a warning if they are not `0600`.
- [ ] **Log & Debug Scrubbing (AC: 2)**
  - [ ] Audit `internal/app/update.go` and `main.go` logging to ensure no raw environment config objects are logged.
  - [ ] Implement a scrubbing wrapper or ensure `fmt.Printf("%+v")` on config structs is avoided in debug paths.
- [ ] **Clipboard & Viewer Sanitization (AC: 3)**
  - [ ] Update `ModalJSONView` rendering logic in `internal/app/view.go` to ensure only the resource data (not the environment context) is serialized.
  - [ ] Verify `Ctrl+J` command in `internal/app/update.go` uses a sanitized data source.
- [ ] **Error Message Sanitization (AC: 4)**
  - [ ] Update the `errMsg` handler in `internal/app/update.go` to scrub potential credentials from error strings before displaying in the footer.
- [ ] **Verification & Tests**
  - [ ] Add a test in `internal/config/loader_test.go` to verify permission checking.
  - [ ] Add a test in `internal/app/json_test.go` verifying that `J`/`Ctrl+J` output is credential-free.

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

Gemini 2.0 Flash

### Debug Log References

### Completion Notes List

### File List
