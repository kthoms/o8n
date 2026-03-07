# Story 5.1: Multi-Environment Configuration

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want to **configure 2 or more named environments with distinct API URLs, credentials, and accent colors**,
so that **I can operate across local, staging, and production without editing config between sessions**.

## Acceptance Criteria

1. **Given** `o6n-env.yaml` contains 2 or more named environment entries, each with `name`, `api_url`, `username`, `password`, and `ui_color`
   **When** the application starts
   **Then** all configured environments are available in the environment switcher (accessed via the environment switch key).
   **And** the active environment's `ui_color` is applied to the border accent color (secondary signal).

2. **Given** the operator switches to a different environment
   **When** the switch completes
   **Then** all subsequent API calls use the new environment's `api_url` and credentials.
   **And** the environment name updates immediately in the top-right header position using the `env_name` semantic color role (primary signal).

3. **Given** a single environment is configured
   **When** the application starts
   **Then** it loads normally — multi-environment is not required for the app to function.

4. **Given** no `o6n-env.yaml` exists
   **When** the application starts
   **Then** it provides a clear error message or guides the user to create one (via `o6n-env.yaml.example`).

## Tasks / Subtasks

- [ ] **Config Loader Enhancements (AC: 1, 3)**
  - [ ] Update `internal/config/config.go` to support a list of environments in `EnvConfig`.
  - [ ] Update `LoadEnvConfig` in `loader.go` to parse multiple environment entries from `o6n-env.yaml`.
  - [ ] Implement logic to select the default environment (first in list or persisted in `o6n-stat.yaml`).
- [ ] **Environment Switching Logic (AC: 2)**
  - [ ] Implement the `switchEnvironmentCmd` in `internal/app/commands.go`.
  - [ ] Ensure `prepareStateTransition(TransitionFull)` is called during an environment switch to clear all stale data.
  - [ ] Update the `internal/client/client.go` to re-initialize with the new environment's credentials and base URL.
- [ ] **UI Integration (AC: 1, 2)**
  - [ ] Update `renderHeader` in `internal/app/view.go` to display the active environment's name using the `env_name` skin role.
  - [ ] Ensure the `ui_color` from the environment config is applied to the border accent in `internal/app/skin.go`.
  - [ ] Register the environment switcher modal using the Modal Factory (`internal/app/modal.go`).
- [ ] **Testing & Validation**
  - [ ] Create `internal/config/env_test.go` to verify multi-environment parsing.
  - [ ] Create `internal/app/env_switch_test.go` to assert state clearing and client re-initialization.
  - [ ] Verify `specification.md` is updated to reflect multi-environment support.

## Dev Notes

### Previous Story Intelligence
- **Story 1.3 Learnings:** Environment switching MUST clear all prior view state via `prepareStateTransition(TransitionFull)`.
- **Story 4.4 Learnings:** `env_name` semantic role is the primary identity signal in the header. `ui_color` is secondary (border accent).

### Architecture Compliance
- **Async Contract:** All environment switching must be handled via `tea.Cmd` and messages.
- **Credential Isolation:** Never log or print the credentials from `o6n-env.yaml`.
- **Modal Factory:** The environment switcher must use the `renderModal()` factory with `OverlayCenter` size hint.

### Anti-Patterns to Avoid (LLM Guardrails)
- **❌ DO NOT** hardcode environment names or URLs in the Go source.
- **❌ DO NOT** modify `o6n-cfg.yaml` for environment-specific data; use `o6n-env.yaml`.
- **❌ DO NOT** forget to call `prepareStateTransition` when switching environments.

### Project Structure Notes
- **Config Management:** `internal/config/`.
- **API Client:** `internal/client/`.
- **TUI Logic:** `internal/app/`.

### References
- [Source: `_bmad/planning-artifacts/epics.md#Story 5.1`]
- [Source: `_bmad/planning-artifacts/prd.md#FR26`]
- [Source: `_bmad/planning-artifacts/architecture.md#Data Architecture`]
- [Source: `_bmad/planning-artifacts/ux-design-specification.md#Environment visibility redesign`]

## Dev Agent Record

### Agent Model Used

Gemini 2.0 Flash

### Debug Log References

### Completion Notes List

### File List
