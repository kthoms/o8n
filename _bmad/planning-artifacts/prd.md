---
stepsCompleted: ["step-01-init", "step-02-discovery", "step-02b-vision", "step-02c-executive-summary", "step-03-success", "step-04-journeys", "step-05-domain", "step-06-innovation", "step-07-project-type", "step-08-scoping", "step-09-functional", "step-10-nonfunctional", "step-11-polish"]
inputDocuments:
  - "CLAUDE.md"
  - "README.md"
  - "specification.md"
  - "story-accessibility-and-empty-states.md"
  - "story-claim-complete-user-tasks.md"
  - "story-keyboard-convention-system.md"
  - "story-layout-optimization.md"
  - "story-search-pagination-awareness.md"
  - "story-state-transition-contract.md"
  - "story-task-complete-dialog-ux.md"
  - "story-task-complete-dialog-polish.md"
  - "story-vim-mode-toggle.md"
workflowType: 'prd'
projectType: 'brownfield'
classification:
  projectType: cli_tool
  domain: general
  complexity: medium
  projectContext: brownfield
documentCounts:
  briefCount: 0
  researchCount: 0
  brainstormingCount: 0
  projectDocsCount: 12
---

# Product Requirements Document — o8n

**Author:** Karsten
**Date:** 2026-03-02
**Project Type:** Brownfield — Established project with significant completed work

---

## Executive Summary

o8n is a keyboard-first terminal UI for managing [Operaton](https://operaton.org) BPMN workflow engines, designed as the definitive operator interface for the Operaton community. Built in Go using the Charmbracelet TUI stack (Bubble Tea, Bubbles, Lipgloss), it provides a k9s-inspired experience across 35 resource types — process definitions, instances, tasks, jobs, incidents, deployments, history, and more — without requiring a browser or REST client.

The primary users are DevOps engineers and BPMN-familiar technical operators who manage Operaton engines as part of platform or operational workflows. o8n eliminates context switching between terminal and browser by bringing the full operational surface of Operaton's REST API into a single, keyboard-driven, configuration-driven TUI. The application is scoped to the Operaton REST API — multi-engine support is explicitly out of scope.

**State-of-the-art TUI quality.** o8n aspires to the interaction quality of tools like k9s, lazygit, and modern AI terminals. Every action provides immediate feedback. Every view surfaces the right information at the right depth. The UI is reactive, colorful, and polished — designed to signal craft.

**Config-driven power.** Tables, columns, actions, drilldowns, key bindings, and resource types are defined in `o8n-cfg.yaml`. Community extensibility without code changes; Go implementation focused on rendering, state management, and API orchestration.

**Smart where it matters.** Content-assist completions, tailored views for specific resource types (task completion dialog with form variables), and context-aware action menus go beyond generic CRUD.

**Community-first architecture.** Clean module separation, comprehensive configuration, and a stable specification (`specification.md`) make it straightforward for contributors to add resource types, actions, and views without deep Go expertise.

---

## Success Criteria

### User Success

- A DevOps engineer or BPMN operator can open o8n, navigate to any resource type, and complete actions (drill-down, edit, execute, delete) without inconsistent behavior, silent failures, or UI corruption
- Available actions are visible without opening the help menu — discoverable via footer hints and the Space action dialog
- All modal dialogs behave predictably: `Esc` cancels, `Enter` confirms, styling is consistent, errors surface inline
- Navigation transitions (context switch, environment switch, drill-down, breadcrumb jump) leave no stale state
- Application renders correctly at 120×20 in VSCode integrated terminal and IntelliJ IDEA terminal with no overflow or truncation of critical information

### Technical Success

- All 35 resource types defined in `o8n-cfg.yaml` load, paginate, and navigate correctly
- All config-defined actions issue the correct REST call and surface errors when they fail
- Key bindings respond predictably in all views — no silent no-ops due to modal or context state mismatches
- State transitions use the centralized contract (`prepareStateTransition`) — no state leakage between environments, contexts, or navigation levels
- Modal system is config-driven — new modal types added via YAML schema without hardcoded Go logic per modal type
- Application does not crash on any malformed API response or unexpected state
- Test suite passes with no regressions after quality sprint changes

### Business Success

- o8n is stable and consistent enough to demonstrate to core Operaton users and gather informal feedback
- `specification.md` and `README.md` accurately reflect the implemented behavior post-sprint
- A Go-familiar contributor can add a new resource type or action by editing `o8n-cfg.yaml` alone — no code changes required for standard resources

### Measurable Outcomes

| Outcome | Measure |
|---|---|
| Action discoverability | Footer hints visible for primary actions in every view |
| Modal consistency | All modal types render from a single config-driven factory |
| State correctness | Zero known state leakage bugs after regression tests |
| API correctness | All actions in `o8n-cfg.yaml` verified to issue correct REST calls |
| Terminal compatibility | Renders correctly at 120×20 in VSCode and IntelliJ terminals |
| Documentation accuracy | `specification.md` updated to match post-sprint implementation |

---

## Product Scope & Roadmap

### MVP — Make It Work Correctly

The current sprint goal. o8n has 35 resource types and a rich feature set — the priority is making what exists **work correctly and consistently**, not adding new capabilities. Single developer executing; 1-week sprint.

**Core User Journeys Supported:**
- Alex: Navigate any resource type, drill-down into incidents, execute actions with feedback
- Priya: Claim and complete tasks via the task completion dialog
- Both: Consistent modal behavior, discoverable actions via footer hints and Space dialog

**Must-Have Capabilities:**
- All 35 resource types: load, paginate, drill-down, and execute configured actions correctly
- Config-driven modal system: consistent styling and Esc/Enter behavior across all modal types
- Footer action hints: primary actions discoverable without opening help menu
- Space action dialog: context-aware action menu surfacing configured actions
- State transition contract: all navigation transitions use `prepareStateTransition` — zero known leakage bugs
- Correct rendering at 120×20 in VSCode and IntelliJ IDEA terminals
- `specification.md` updated to reflect post-sprint behavior

### Growth Features (Phase 2 — Post-Sprint)

- Mouse support for table rows and modal buttons (keyboard-first remains primary)
- Smart completions beyond user suggestions: variable names, process keys, deployment names
- Per-resource tailored views beyond the generic table
- Additional terminal size profiles (80×24, 160×40)
- Configurable auto-refresh intervals per resource type
- Export actions: copy row as JSON, export to file

### Vision — Community Platform (Phase 3)

- Community-maintained resource type library with contributed configurations
- Advanced BPMN process navigation (visual drill-down into BPMN diagrams from terminal)
- Integration with Operaton's event/notification system for live process monitoring
- Plugin architecture allowing community modules to extend the tool without forking
- Operaton community recognition as official tooling

### Risk Mitigation

**Technical:** Modal extraction is the highest-risk change — hardcoded logic moving to config-driven factory. Mitigation: extract one modal type at a time, validate against test suite before proceeding.

**Resource:** If time runs short, footer hints (Day 3) and Space dialog (Day 4) are highest-value — modal extraction (Day 1–2) can be scoped to most-used modal types.

---

## User Journeys

### Journey 1: Alex — DevOps Engineer (Operations + Incident Response)

**Persona:** Platform engineer at a mid-sized company running Operaton as their BPMN workflow backbone. Manages three environments: local dev, staging, and production.

**Normal Operations:**
Alex opens VSCode, splits the terminal, launches `./o8n`. The app restores to last position: `process-instances` in production. The footer shows the environment name in accent color — green dot, API reachable. Alex scans the table: 47 active instances, no anomalies. `r` toggles auto-refresh, footer badge starts pulsing. Morning check done in 30 seconds.

**Incident Investigation:**
Slack pings — a process instance for `invoice-approval` is stuck. Alex presses `:`, types `inc`, selects `incidents`. The incident table loads: one open incident on instance `abc-1234`, message `NullPointerException in service task`. Alex presses `Enter` to drill into the incident's process instance, then `Enter` again to inspect variables. The variable `approver` is empty. Alex presses `e` to edit the variable inline, sets a value, saves. Back in the incident view, presses `a` to set an annotation documenting the fix, then `Space` → selects `Retry` to resume the job.

**Outcome:** Process instance transitions from INCIDENT to ACTIVE. Footer flashes `✓ Job retried`. Incident resolved without opening a browser or writing a curl command. Investigations that used to take 10+ minutes now take under 2.

**Requirements revealed:** Incident drill-down, variable editing, action execution with feedback, environment switching, auto-refresh, breadcrumb navigation.

---

### Journey 2: Priya — Process Manager / BPMN Operator (Task Work)

**Persona:** Business process analyst at a financial services firm. BPMN-fluent, uses Operaton Cockpit daily. Skeptical of terminal UIs but intrigued by speed.

**Task Processing:**
Priya is assigned a batch of loan review tasks. She launches o8n, opens help (`?`), scans key categories. Two minutes later she's navigating like a pro. She switches to `task` view with `:` + `task` + `Enter`. The table shows 12 open tasks. She filters with `/loan-review` — 5 matching tasks appear.

She selects the first, presses `Space` — the action menu appears. She presses `c` to claim the task; the assignee column updates instantly. She presses `Enter` — the task completion dialog opens, showing input variables (loan application data) read-only and output form variables (her review decision) editable. She fills in `decision: approved`, `notes: all criteria met`, presses `Enter` on `[Complete]`. The task disappears. The footer confirms: `✓ Completed: Loan Review #4521`.

**Outcome:** All 5 tasks completed in 8 minutes. Cockpit would have taken 25+ minutes across browser tabs. Priya adopts o8n for daily task processing and starts teaching colleagues.

**Requirements revealed:** Task claim/unclaim/complete dialog, form variable editing, type-aware input (bool, string, integer), footer action hints, search/filter.

---

### Journey 3: Marco — Go Developer / Community Contributor

**Persona:** Backend developer at a company using Operaton. Has been using o8n for a month; his team needs visibility into `external-task` worker registrations — a resource type not yet in `o8n-cfg.yaml`.

**Contribution Path:**
Marco forks the repo, opens `specification.md` first — it's comprehensive. The configuration model section explains exactly how tables, columns, actions, and drilldowns work. He reads one existing table definition in `o8n-cfg.yaml` to understand the pattern.

Marco adds a new table entry for `external-task-topic` — columns, a drilldown to active external tasks, two actions: `unlock` and `set-retries`. He builds (`make build`), runs `./o8n --no-splash`, presses `:`, types his resource name — it appears. The table loads. He tests both actions. They work.

For a tailored view showing worker registration counts, Marco opens `internal/app/`, follows the Bubble Tea model/update/view cycle, adds his feature, writes tests using the established `httptest` pattern, runs `make test`. All pass. He opens a PR; the reviewer understands immediately. The PR merges in two days.

**Outcome:** The pattern is clear and respected: `o8n-cfg.yaml` for standard resources, Go for tailored behavior. Marco becomes a recurring contributor.

**Requirements revealed:** Comprehensive `specification.md`, clear config schema, modular Go architecture, test patterns.

---

### Journey Requirements Summary

| Capability | Journey |
|---|---|
| Environment switching with status indicators | Alex |
| Incident drill-down and job retry | Alex |
| Variable inspection and inline editing | Alex, Priya |
| Action discoverability (footer hints + Space dialog) | All |
| Task claim / complete with form variable dialog | Priya |
| Search and filter in table views | Priya |
| Context switcher (`:`) for resource type navigation | All |
| Config-driven table/column/action definitions | Marco |
| Clear contribution path: config-first, Go for tailored behavior | Marco |
| Auto-refresh with visual indicator | Alex |
| Breadcrumb navigation and drill-down | Alex |

---

## Domain-Specific Requirements

### Credential Security

- `o8n-env.yaml` must remain git-ignored at all times and carry `chmod 600` file permissions
- Credentials (username, password, API URL) are stored per environment and never written to version-controlled files
- No credential caching, logging, or exposure in debug output

### Operaton API Integration

- The Go client in `internal/operaton/` is auto-generated from `resources/operaton-rest-api.json` via OpenAPI Generator — it must not be edited manually
- API client must be regenerated when the Operaton API spec evolves (`.devenv/scripts/generate-api-client.sh`)
- All API responses must be handled defensively: malformed or unexpected responses surface as user-visible errors, never panics
- Basic Auth is the only supported authentication mechanism (matching Operaton's engine-rest API)

---

## Application Interface Requirements

### Command Structure

Single binary; no subcommands.

| Flag | Behavior |
|---|---|
| `--debug` | Enables verbose logging to `./debug/access.log` and `./debug/last-screen.txt` |
| `--no-splash` | Skips the animated ASCII splash screen and goes directly to the main view |
| `--skin <name>` | Overrides the active skin at startup (must match a file in `skins/`) |
| `--vim` | Enables vim-style keybindings at startup (toggleable in-session with `V`) |

### Output Formats

Terminal rendering only — ANSI escape codes, Lipgloss-styled output, box-drawing characters. No JSON stdout, no pipe-friendly output mode. The `y` key copies the selected row as YAML to the clipboard — convenience action, not a machine-readable output mode.

### Configuration Schema

| File | Contents | Committed |
|---|---|---|
| `o8n-env.yaml` | Per-environment credentials, API URLs, `ui_color` | No — git-ignored, `chmod 600` |
| `o8n-cfg.yaml` | Table definitions, columns, actions, drilldowns, key bindings | Yes |
| `o8n-stat.yaml` | Runtime state (last context, last environment, active skin) | No — auto-managed |

---

## Functional Requirements

### Resource Navigation & Browse

- **FR1**: Operator can navigate to any of the 35 configured resource types using the context switcher (`:` key)
- **FR2**: Operator can browse a paginated table of resources in the current context
- **FR3**: Operator can drill down from a parent resource to related child resources as configured in `o8n-cfg.yaml`
- **FR4**: Operator can navigate back through the drill-down history level by level using Escape
- **FR5**: Operator can jump directly to a specific level in the breadcrumb trail

### Action Execution

- **FR6**: Operator can execute any action configured for the current resource type on the selected row
- **FR7**: Operator is prompted to confirm destructive actions before they are executed
- **FR8**: Operator receives visible success or error feedback in the footer after an action completes
- **FR9**: Operator can retry a failed job associated with an incident
- **FR10**: Operator can set an annotation on an incident

### Modal & Dialog System

- **FR11**: All modal dialog types render with consistent visual styling and layout
- **FR12**: Operator can dismiss any modal by pressing Escape
- **FR13**: Operator can confirm any modal by pressing Enter on the confirm action
- **FR14**: Operator can interact with edit dialogs that validate input by type (string, integer, boolean, JSON)

### Action Discoverability

- **FR15**: Operator can see the primary available actions for the current view in the footer without opening a help screen
- **FR16**: Operator can open a context-sensitive action menu via the Space key showing all available actions for the selected row
- **FR17**: Operator can view the full key binding reference via the `?` key

### State Management & Navigation

- **FR18**: Operator can switch between configured environments at any time
- **FR19**: Operator can switch to any resource context using the `:` context switcher without leaving stale state
- **FR20**: All navigation transitions (environment switch, context switch, drill-down, breadcrumb jump) clear prior view state completely
- **FR21**: Application restores the last active context and environment on startup

### Task Operations

- **FR22**: Operator can claim an unassigned task
- **FR23**: Operator can unclaim a task
- **FR24**: Operator can complete a claimed task via a dialog that displays input variables read-only and allows editing output (form) variables
- **FR25**: Task completion dialog supports variable types: string, integer, boolean

### Environment & Configuration

- **FR26**: Operator can configure multiple named environments with distinct API URLs, credentials, and accent colors
- **FR27**: Application reads resource types, columns, actions, and drilldown rules from `o8n-cfg.yaml` at startup
- **FR28**: Contributor can add a new standard resource type by editing `o8n-cfg.yaml` without modifying Go source code

### Data & Variable Management

- **FR29**: Operator can inspect process variables associated with a process instance
- **FR30**: Operator can edit a process variable value inline with type validation
- **FR31**: Operator can copy the selected resource row as YAML to the system clipboard

### Search & Filter

- **FR32**: Operator can filter the current resource table by entering a search term
- **FR33**: Operator can clear the active search filter and return to the full table
- **FR34**: Operator can toggle auto-refresh to continuously update the current table view

### Visual Presentation

- **FR35**: Application renders without overflow or truncation of critical information at 120×20 terminal size
- **FR36**: Application adapts column visibility and hint display when the terminal is narrower than optimal
- **FR37**: Application handles terminal resize events without corrupting the layout
- **FR38**: Operator can switch between available color skins
- **FR39**: Operator can toggle vim-style key bindings in-session

---

## Non-Functional Requirements

### Performance

- **NFR1**: The UI responds to any key press within 100ms — no perceptible input lag
- **NFR2**: API calls that take longer than 500ms surface a visible loading indicator; the UI does not appear frozen
- **NFR3**: All API calls are asynchronous — network operations never block the Bubble Tea event loop or UI rendering

### Reliability

- **NFR4**: The application must not panic or crash on any malformed, partial, or unexpected API response
- **NFR5**: All API errors surface as user-visible footer messages — no silent failures
- **NFR6**: The application recovers gracefully from network timeouts and connection failures without requiring a restart
- **NFR7**: Footer error and success messages auto-clear after 5 seconds and do not block further interaction

### Security

- **NFR8**: Credentials (username, password) must never appear in log files, debug output, or clipboard operations
- **NFR9**: `o8n-env.yaml` must be git-ignored and maintained at `chmod 600` file permissions at all times

### Terminal Compatibility

- **NFR10**: Application renders correctly at 120×20 in VSCode integrated terminal and IntelliJ IDEA terminal without overflow or truncation of critical information
- **NFR11**: Application functions correctly in standard POSIX terminals (xterm, iTerm2, Alacritty) without modification
- **NFR12**: Application handles terminal resize events without corrupting the rendered layout

### Maintainability

- **NFR13**: Adding a standard resource type (table + columns + actions + drilldowns) requires only edits to `o8n-cfg.yaml` — no Go source changes
- **NFR14**: The OpenAPI client in `internal/operaton/` remains auto-generated — no manual edits permitted; regenerated via `.devenv/scripts/generate-api-client.sh`
- **NFR15**: The modal system is config-driven — new modal types are supported through the modal factory without hardcoded per-type Go logic
