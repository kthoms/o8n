# Story 4.4: Color Skins & Environment Identity

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want **to switch between available color skins and see the environment name in a distinct color in the header**,
so that **I can instantly distinguish production from staging and customize my visual experience**.

## Acceptance Criteria

1. **Given** skin files in `skins/` (e.g., `dracula.yaml`, `nord.yaml`)
   **When** the operator switches skins using the skin selection key (from `o6n-cfg.yaml`)
   **Then** the new skin is applied immediately without restarting the application.
   **And** all UI elements use semantic color roles from the new skin (no hardcoded colors).
   **And** the active skin is persisted to `o6n-stat.yaml` and restored on startup.

2. **Given** any view is rendered
   **When** the header is drawn
   **Then** the environment name (from `o6n-env.yaml`) is displayed in the fixed top-right position.
   **And** it uses the `env_name` semantic color role from the active skin as the primary identity signal.
   **And** the label remains visible even at the 120-column minimum width.

3. **Given** a `ui_color` (hex) is set for the active environment in `o6n-env.yaml`
   **When** the skin is applied
   **Then** this `ui_color` overrides the **border accent color** only.
   **And** it does NOT affect the `env_name` label color or text content colors.

4. **Given** the application is running
   **When** the operator switches environments
   **Then** the environment name and its associated identity colors update immediately in the header.

## Tasks / Subtasks

- [ ] **Implement Skin Persistence & Restore (AC: 1)**
  - [ ] Update `internal/config/config.go` and `loader.go` to handle `active_skin` in `StatConfig`.
  - [ ] Ensure `m.skin` is initialized from `StatConfig` at startup.
- [ ] **Implement Semantic `env_name` Role (AC: 2)**
  - [ ] Add `EnvName` (Lipgloss style) to the `Skin` struct in `internal/app/skin.go`.
  - [ ] Update `renderHeader` in `internal/app/view.go` to use the `env_name` role for the top-right environment label.
- [ ] **Implement Border Accent Override (AC: 3)**
  - [ ] Update `internal/app/skin.go` to apply `m.cfg.Env.UiColor` to the border style if present.
  - [ ] Ensure this override is secondary to the `env_name` role for identity.
- [ ] **Immediate Visual Refresh (AC: 1, 4)**
  - [ ] Verify that switching skins or environments triggers a full re-render with new styles.
  - [ ] Ensure all `lipgloss.Style` objects are rebuilt using the new skin roles.
- [ ] **Testing & Validation (AC: all)**
  - [ ] Create `internal/app/skin_test.go` to verify semantic role resolution and overrides.
  - [ ] Manually validate color identity across multiple environments in VSCode/iTerm2.

## Dev Notes

### Previous Story Intelligence
- **Story 4.1-4.3 Learnings:** Always respect `m.termWidth` and `m.termHeight`.
- **Chrome Allocation:** The `env_name` label is part of the Row 1 header. It must not overlap with the resource title or pagination info.

### Architecture Compliance
- **Semantic Colors:** Never use `lipgloss.Color("#RRGGBB")` in view logic. Always use `m.skin.Primary`, `m.skin.Accent`, `m.skin.EnvName`, etc.
- **Config Ownership:** `env_name` color is set per skin; `ui_color` (border accent) is set per environment in `o6n-env.yaml`.

### Anti-Patterns to Avoid (LLM Guardrails)
- **❌ DO NOT** hardcode environment colors in Go.
- **❌ DO NOT** allow the environment label to be hidden or truncated at 120 columns; truncate the resource title instead.
- **❌ DO NOT** use the `ui_color` for text content; it is for borders only.

### UI/UX Standards
- **Environment Identity:** The `env_name` color is the primary signal. It should be high-contrast and distinctive.
- **Visual Polish:** Skin transitions must be flicker-free and immediate.

### Project Structure Notes
- **Skin Logic:** `internal/app/skin.go` and `internal/app/styles.go`.
- **Header Render:** `internal/app/view.go`.
- **Persistence:** `internal/config/loader.go`.

### References
- [Source: `_bmad/planning-artifacts/epics.md#Story 4.4`]
- [Source: `_bmad/planning-artifacts/prd.md#FR38`]
- [Source: `_bmad/planning-artifacts/architecture.md#Theming`]
- [Source: `_bmad/planning-artifacts/ux-design-specification.md#Environment visibility redesign`]

## Dev Agent Record

### Agent Model Used

Gemini 2.0 Flash

### Debug Log References

### Completion Notes List

### File List
