# Story 5.1: Multi-Environment Configuration

Status: done

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

- [x] **Config Loader Enhancements (AC: 1, 3)**
  - [x] `EnvConfig.Environments` is a `map[string]Environment` in `config.go`
  - [x] `LoadEnvConfig` parses multiple environment entries from `o6n-env.yaml`
  - [x] Default environment logic: persisted `AppState.ActiveEnv` > `"local"` > first sorted key
- [x] **Environment Switching Logic (AC: 2)**
  - [x] `switchToEnvironment(name)` in `nav.go` updates `m.currentEnv` and `m.client`
  - [x] `prepareStateTransition(TransitionFull)` called on every env switch (`update.go:506`)
  - [x] `client.NewClient(env, debug)` re-initializes client with new environment credentials
- [x] **UI Integration (AC: 1, 2)**
  - [x] Header displays active env name with `env_name` semantic role (Story 4.4 fix applied)
  - [x] `applyStyle()` applies `ui_color` from active env to border accent
  - [x] `ModalEnvironment` registered in `modal.go` via `registerModal()`
- [x] **Testing & Validation**
  - [x] Multi-environment config parsing tested via existing config tests
  - [x] Environment switch behavior covered in `skin_integration_test.go`
  - [x] `specification.md` section 9 documents multi-environment support

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

Claude Haiku 4.5

### Debug Log References

### Senior Developer Review (AI)

- **Outcome:** Approved with task-checkbox fixes
- **Date:** 2026-03-08
- **Reviewer:** Claude Sonnet 4.6 (adversarial code review)
- **Action items taken:**
  - HIGH: All tasks marked `[ ]` despite full implementation — FIXED: tasks updated to `[x]`
  - Verified: `EnvConfig.Environments` map supports unlimited environments
  - Verified: `switchToEnvironment` calls `prepareStateTransition(TransitionFull)` and rebuilds client
  - Verified: `ModalEnvironment` registered in modal factory
  - LOW: No dedicated `env_switch_test.go` — covered by `skin_integration_test.go` which tests env switching

### Completion Notes List

- Story implementation verified and tested.
- All acceptance criteria addressed.
- Multi-environment switching with TransitionFull confirmed.

### File List

- `internal/config/config.go` — EXISTING: `EnvConfig.Environments` map, `LoadEnvConfig`
- `internal/app/nav.go` — EXISTING: `switchToEnvironment`
- `internal/app/update.go` — EXISTING: env switch handler with `prepareStateTransition(TransitionFull)`
- `internal/app/modal.go` — EXISTING: `ModalEnvironment` registered
- `internal/app/skin_integration_test.go` — EXISTING: environment switch tests

