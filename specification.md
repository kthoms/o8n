# o8n Specification

The authoritative technical specification for o8n — a k9s-inspired terminal UI for managing Operaton workflow engines. This document fully describes the architecture, behavior, and design of the application. A developer can reimplement the same structure and behavior from this specification alone.

---

## 1. Overview

- **Language:** Go (>= 1.24)
- **UI framework:** Charmbracelet ecosystem (Bubble Tea, Bubbles, Lipgloss)
- **Function:** Keyboard-first terminal UI to browse and operate on resources exposed by an Operaton engine-rest API
- **Design model:** k9s — same layout patterns, keyboard conventions, and information density philosophy
- **Binary:** Static build with `CGO_ENABLED=0` and `GO_TAGS=netgo`

### Design Principles

| Principle | Expression |
|---|---|
| Keyboard first | Every action has a key binding; mouse is irrelevant |
| Information hierarchy | Screen space is scarce; only what matters is shown |
| Consistency | Same patterns across all views reduce cognitive load |
| Progressive disclosure | Show the right data at the right drill depth |
| Immediate feedback | Every action confirms itself within the same frame |
| Safe destructive actions | Two-step confirmation before irreversible operations |
| Config-driven | Tables, columns, actions, drilldowns all defined in YAML |

---

## 2. Architecture

### Bubble Tea Pattern

The app follows the Elm-inspired model/update/view cycle:

- **`Init()`** — initial commands (splash screen, config load)
- **`Update(msg tea.Msg)`** — pure message handlers, return new model + commands
- **`View()`** — render to string

All async operations (API calls) return `tea.Cmd` and communicate results back via typed messages.

### Repository Layout

```
o8n/
├── main.go                    # Thin entry point, calls internal/app.Run()
├── internal/
│   ├── app/                   # All TUI logic: model, update, view, nav, commands, skin, styles, table, util, edit
│   ├── client/                # HTTP client wrapper using the generated OpenAPI client
│   ├── config/                # Config structs and loaders for env, app, and state files
│   ├── dao/                   # DAO interfaces (DAO, HierarchicalDAO, ReadOnlyDAO)
│   ├── validation/            # Input validation for edit dialogs (bool/int/float/json/text)
│   ├── contentassist/         # Thread-safe user suggestion cache (SuggestUsers)
│   └── operaton/              # Auto-generated OpenAPI client — do not edit manually
├── skins/                     # 35 color theme YAML files
├── resources/                 # OpenAPI spec (operaton-rest-api.json)
├── o8n-env.yaml               # Environment credentials (git-ignored, 0600 perms)
├── o8n-cfg.yaml               # Table definitions, column specs, actions (version-controlled)
├── o8n-stat.yml               # Runtime state (auto-generated, git-ignored)
└── o8n-env.yaml.example       # Template for environment configuration
```

### Messages and Commands

**Messages (Go types):** `refreshMsg`, `definitionsLoadedMsg`, `instancesLoadedMsg`, `variablesLoadedMsg`, `dataLoadedMsg`, `genericLoadedMsg`, `instancesWithCountMsg`, `errMsg`, `flashOnMsg`, `flashOffMsg`, `clearErrorMsg`, `terminatedMsg`, `splashDoneMsg`, `splashFrameMsg`

**Commands:** `fetchDefinitionsCmd`, `fetchInstancesCmd`, `fetchVariablesCmd`, `fetchDataCmd`, `fetchForRoot`, `terminateInstanceCmd`, `flashOnCmd`, `setVariableCmd`

### API Client

- Generated from `resources/operaton-rest-api.json` using OpenAPI Generator (via Docker)
- Regenerate with `.devenv/scripts/generate-api-client.sh` (runs `go mod tidy` automatically)
- Generated files live in `internal/operaton/` — never edit manually
- The `internal/client/` package wraps the generated client for application use
- Authentication: HTTP Basic Auth via `operaton.ContextBasicAuth` context
- Nullable types: `NullableString`, `NullableInt32`, `NullableBool` with safe extraction helpers

### Content Assist

Package `internal/contentassist` provides a thread-safe suggestion cache used for input completion:

- **`SetUserCache(items []string)`** — replaces the internal cache (populated from API responses or tests)
- **`SuggestUsers(prefix string) []string`** — returns up to 5 suggestions matching the given prefix (case-insensitive); returns top 5 if prefix is empty
- Protected by `sync.RWMutex` for concurrent access
- Currently used for user name suggestions in the edit modal (`input_type: user`)
- Designed to scale: the same pattern can support variable name completion, process key completion, or other domain-specific suggestions

### Error Handling

- All API errors propagate as `errMsg` messages
- `Update()` stores error text in `m.footerError` and schedules auto-clear after **5 seconds**
- Rendering functions use `defer/recover` to catch panics — the TUI never crashes from malformed API responses
- Panic recovery in `Update()` logs to `debug/o8n.log` and shows a user-friendly error

---

## 3. Configuration Model

Three files with distinct responsibilities:

### o8n-env.yaml (Environment Credentials)

Git-ignored. File permissions 0600. Contains only credentials and environment definitions — no mutable runtime state.

```yaml
environments:
  local:
    url: http://localhost:8080/engine-rest
    username: demo
    password: demo
    ui_color: "#00A8E1"
    default_timeout: 10s
  production:
    url: https://operaton.example.com/engine-rest
    username: admin
    password: secret
```

### o8n-cfg.yaml (Application Configuration)

Version-controlled. Static — never written at runtime. Defines all tables, columns, drilldowns, and actions.

**Structure:**
```yaml
tables:
  - name: <table-identifier>
    api_path: <optional REST path override, supports {parentId}>
    count_path: <optional count endpoint override, "" to disable>
    columns:
      - name: <column-name>
        visible: true|false
        width: "<N>%"
        align: left|center
        editable: true|false
        input_type: bool|int|float|json|text|auto|user
        type: <optional type hint>
        hide_order: <int, lower = hide first on narrow terminals>
    drilldown:
      - target: <child table name>
        param: <query parameter name>
        column: <source column for param value>
        label: <optional breadcrumb display text>
        title_attribute: <optional title attribute>
    actions:
      - key: <key binding, e.g. "s", "ctrl+d">
        label: <display name>
        method: GET|POST|PUT|DELETE
        path: <URL template with {id}, {name}, {value}, {type}>
        body: <optional JSON body>
        confirm: true|false
        id_column: <optional, defaults to "id">
    edit_action:
      method: PUT
      path: <URL template with {id}, {name}, {parentId}, {value}, {type}>
      body_template: <JSON body template>
      id_column: <optional, defaults to "id">
      name_column: <optional, defaults to "name">
```

**Column width:** Parsed as percent string (e.g., `"25%"`). Normalized if percentages don't sum to 100. Unspecified columns share remaining space equally.

**Column types:**
- Built-in: `id` (string), `date-time`
- Implicit: columns named `id` or ending with `Id` are treated as type `id`
- `DisplayName` defaults to `name` capitalized with `-` replaced by space

### o8n-stat.yml (Runtime State)

Auto-generated. Git-ignored. Updated on every navigation transition and on clean exit.

```yaml
active_env: local
skin: dracula
show_latency: false
navigation:
  root: process-instance
  breadcrumb: [process-definitions, process-instances]
  selected_definition_key: my-process
  selected_instance_id: abc-123
  generic_params: {}
```

### Config Loading

- `LoadSplitConfig()` merges `o8n-env.yaml` and `o8n-cfg.yaml` into a unified `Config` struct at runtime
- Backward-compatible: `LoadConfig(path)` supports legacy single-file format used by tests
- `LoadEnvConfig`, `LoadAppConfig`, and `Save` helpers available

---

## 4. Navigation Model

### Drill-Down Hierarchy

```
Context Switch (:)
      |
      v
Root Resource  --Enter/-->  Child Resource  --Enter/-->  Grandchild
      <----------- Esc ----------------- Esc -----------
```

Examples of config-driven drilldown chains:
- Process Definition -> Process Instances -> Variables
- Job Definition -> Jobs
- Deployment -> Process Definitions or Decision Definitions
- 10+ additional chains defined in `o8n-cfg.yaml`

### Navigation Stack

- **`viewState`** struct: captures a complete snapshot — viewMode, breadcrumb, contentHeader, selectedDefinitionKey, selectedInstanceID, tableRows, tableCursor, cachedDefinitions, tableColumns
- **`navigationStack []viewState`** — LIFO stack. Enter pushes, Esc pops
- **Full state restoration** on Esc: rows, cursor position, columns, filters all restored
- **State persisted to `o8n-stat.yml`** — the app reopens at the last resource/drilldown level

### Breadcrumb

- **`breadcrumb []string`** — ordered list of context labels shown in the footer
- Up to 3 depth levels (e.g., `[process-definitions, my-process, variables]`)
- Keys `1`–`4` navigate directly to breadcrumb level N (1 = root)

### Context Switching

- **`:`** opens a popup with single-line text input + match list (up to 8 visible rows, scrollable)
- Ghost completion in `#666666` after first character typed
- `Tab` completes to first match
- `Enter` switches context (requires exact match or selected popup item)
- `Esc` cancels and clears input
- `Up/Down` moves popup cursor
- On switch: resets navigation stack, clears errors, fetches new root resource

---

## 5. Table System

### Rendering

- Boxed table using `table.Model` fills remaining height between header and footer
- Table header text is uppercase, uses header color (white)
- Column widths computed from configured percentages
- All rows normalized to configured column count (pad/truncate) to prevent renderer panics

### Content Box Title

Adapts to state:
```
process-definitions — 42 items             (normal)
process-definitions [/proc/ — 3 of 42]    (search active)
process-instances(my-key)                  (drilled down)
```

### Selection

- Selected row: inverted colors (font <-> background)
- Drillable rows: prefixed with `>` focus indicator (stripped when resolving IDs)
- Editable columns: cell value shown with `[E]` suffix

### Responsive Column Visibility

Terminal width determines which columns are visible:

| Terminal Width | Behavior |
|---|---|
| < 80 | Minimum viable; most hints hidden |
| 80-99 | Core navigation hints visible |
| 100-119 | Actions and refresh hints appear |
| 120-139 | Environment and quit hints visible |
| 140+ | All hints visible |

Columns with `hide_order` auto-hide on narrow terminals. Low-priority columns shrink to header width + 1 space before hiding entirely.

### Sorting

- `s` opens sort popup: list of sortable columns with `>` cursor
- Sort direction toggles on re-select: `^` ascending -> `v` descending
- "Clear sort" option at top when sort is active
- Client-side sort — type-aware (integers, dates, strings)
- Sort state resets on context switch or data refresh

### Search & Filter

- `/` opens a command-palette-style popup
- Typing filters rows client-side (case-insensitive, searches all visible columns)
- `Enter` locks filter (close popup, keep filtered rows); `Esc` cancels and restores all rows
- `Up/Down` navigate matching entries in popup
- Pagination warning: footer message shown if search is limited to current page
- Title updates to show match count: `[/term/ — 3 of 42]`

### Pagination

- Server-side via `firstResult` (offset) and `maxResults` (page size) query parameters
- `pageOffsets map[string]int` tracks current offset per root context
- `pageTotals map[string]int` tracks total count (from `/count` endpoint)
- Page size computed from visible table height (`getPageSize()`)
- `PgDn`/`Ctrl+F` advances; `PgUp`/`Ctrl+B` retreats; clamped to bounds
- `pendingCursorAfterPage` restores cursor position after page fetch
- Flash indicator shown during fetch

---

## 6. Action System

### Keyboard Convention

| Category | Pattern | Examples |
|---|---|---|
| Destructive | `Ctrl+` prefix | `Ctrl+D` delete/terminate, `Ctrl+C` quit |
| Navigation | Arrow keys, Enter, Esc, PgUp/PgDn | Drill down, go back, paginate |
| Global UI | Single key | `?` help, `:` context switch, `/` search, `s` sort |
| View toggles | Single key | `r` auto-refresh, `L` latency, `y` detail view |
| Edit | `e` | Opens edit modal for selected row |
| Actions menu | `Space` | Opens context-specific actions for selected row |

### Config-Driven Actions

Defined per table in `o8n-cfg.yaml`. Each action specifies:
- Key binding, label, HTTP method, URL path template, optional JSON body
- `confirm: true` triggers the two-step confirmation pattern
- `{id}` placeholder resolved from the row's ID column (configurable via `id_column`)

Resource-specific action examples:
- **Process Instance**: Ctrl+D=Delete, s=Suspend, a=Activate, r=Resume
- **Task**: c=Complete, k=Claim, u=Unclaim, d=Delegate, Ctrl+D=Delete
- **Job**: r=Retry, x=Execute, s=Suspend, a=Activate, Ctrl+D=Delete
- **Incident**: a=Set Annotation, Ctrl+D=Resolve

### Actions Menu

`Space` opens a context-sensitive overlay menu for the selected row:
```
+------------------------------------+
|  Actions: process-instance         |
|  my-process-id                     |
|                                    |
|  > [s] Suspend                     |
|    [r] Resume                      |
|    [y] View as JSON                |
|                                    |
|  Esc: Close                        |
+------------------------------------+
```

"View as JSON" (`y`) is always appended as the last item.

### Two-Step Confirmation Pattern

For destructive actions (`confirm: true`):
1. First keypress opens `ModalConfirmDelete` dialog with item details
2. Same keypress again confirms and executes
3. Any other key or `Esc` cancels with "Cancelled" feedback (2 seconds)
4. Default focus is Cancel button (safe default); Tab toggles focus

---

## 7. Modal System

All modals are **overlays** on top of the main UI — the background table remains visible. Never blank-canvas replacements.

### Modal Types

| Type | Trigger | Rendering |
|---|---|---|
| `ModalNone` | — | No modal |
| `ModalConfirmDelete` | Destructive action key | `overlayCenter(baseView, modal)` |
| `ModalConfirmQuit` | `Ctrl+C` | `overlayCenter(baseView, modal)` |
| `ModalHelp` | `?` | Full-screen `lipgloss.Place` (scrollable) |
| `ModalEdit` | `e` | `overlayCenter(baseView, modal)` |
| `ModalSort` | `s` | Full-screen `lipgloss.Place` (centered) |
| `ModalDetailView` | `y` | Full-screen `lipgloss.Place` (centered, scrollable) |
| `ModalEnvironment` | `Ctrl+E` | Full-screen `lipgloss.Place` (centered) |

### Edit Modal

- Opens on `e` when the current table has editable columns; shows "No editable columns" error otherwise
- `Tab`/`Shift+Tab` cycle through editable columns
- Type-aware input validation:
  - **Bool**: Space toggles `true`/`false`
  - **Int**: Numeric input only
  - **Float**: Decimal number validation
  - **JSON**: Parses and validates JSON syntax
  - **Text**: Free-form text
  - **User**: Content-assist suggestions from user cache
  - **Auto**: Infers type from variable metadata
- `Enter` validates and saves via API; closes on success
- `Esc` closes without saving
- Save button disabled (grey) when validation fails
- Errors displayed inline in the modal

### Help Screen

Full-screen three-column layout. Scrollable with `j`/`k`/`Up`/`Down`/`Ctrl+D`/`Ctrl+U`. Scroll indicators: `^ more above` / `v more below`. Any other key closes.

### Detail View

Scrollable JSON viewer with syntax highlighting:
- Keys: `#00BCD4` (cyan)
- Strings: `#50C878` (green)
- Numbers: `#FFD700` (gold)
- Booleans/null: `#FF69B4` (pink)

### Environment Picker

Lists all configured environments with:
- Connection status: `green circle` (operational), `yellow circle` (unknown), `red X` (unreachable)
- Environment accent color applied to name
- `check` marker on currently active environment
- URL shown inline

---

## 8. Theming

### Skin System

35 built-in skins in `skins/*.yaml`:

**Dark:** dracula, nord, gruvbox-dark (3 variants), one-dark, nightfox, kanagawa, monokai, solarized-dark, everforest-dark, snazzy, vercel, rose-pine, rose-pine-moon, in-the-navy, narsingh, red, black-and-wtf

**Light:** gruvbox-light (3 variants), everforest-light, solarized-light, gruvbox-material-light (3 variants), rose-pine-dawn

**Material:** gruvbox-material-dark (3 variants)

**Special:** o8n-cyber, transparent, stock, kiss, axual, solarized-16

### Semantic Color Roles

All colors driven by 25 semantic roles — no hardcoded values:
- `border`, `borderFocus`, `fg`, `bg`
- `accent` (environment-specific via `ui_color`)
- `success`, `warning`, `danger`, `muted`
- Body fg/bg, logo color, header colors, etc.

Environment `ui_color` always overrides the border accent color.

### Theme Picker

- `Ctrl+T` opens the picker
- `Up/Down` to preview live (instant application)
- `Enter` applies and persists to `o8n-stat.yml`
- `Esc` reverts to the previous skin
- Scrollable list — no truncation
- `--skin <name>` CLI flag overrides startup skin

Falls back to `stock.yaml` if the configured skin is not found.

### Accessibility

Color is never the sole information channel. Every color-coded element has a non-color fallback:

| Element | Color Signal | Non-Color Fallback |
|---|---|---|
| Environment status | Green/Yellow/Red | Distinct symbols: `●` / `○` / `✗` |
| Sort direction | — | Text indicators: `▲` / `▼` |
| Error messages | Red | `✗` prefix + text content |
| Success messages | Green | `✓` prefix + text content |
| Selected row | Inverted colors | Cursor prefix `▶` + positional context |
| Editable columns | — | `[E]` text suffix |

High-contrast friendly skins: `stock`, `black-and-wtf`, `solarized-16`.

---

## 9. Environment Management

### Multi-Environment Support

- Configure unlimited environments in `o8n-env.yaml`
- Each environment has: URL, username, password, ui_color, default_timeout
- Active environment persisted in `o8n-stat.yml`
- Credentials isolated per environment

### Environment Switching

- `Ctrl+E` opens the environment picker modal
- `Up/Down` to select, `Enter` to switch
- Auto-selects "local" or first configured environment on fresh start

### Health Monitoring

- All environments polled on startup and every 60 seconds
- Status per environment: `operational` / `unreachable` / `unknown`
- Header shows current environment name + color-coded status indicator:
  - Green (`green circle`) = API responding
  - Red (`X`) = API unreachable
  - Yellow (`dim circle`) = Status unknown / not checked yet

---

## 10. UI Layout

### Screen Anatomy

```
+------------------------------------------------------------------+
| o8n v0.1.0 | local ●            ? help  : switch  ↑↓ nav  / find | <- Header row 1 (status + hints)
| Ctrl+E env  Ctrl+T skin  s sort  Space actions  r refresh        | <- Header row 2 (more hints)
|                                                                  | <- Header row 3 (spacer)
+------------------------------------------------------------------+
| proc|ess-definitions                              (ghost suffix) | <- Context popup (:) - shown/hidden
| ↑↓:select  Tab/Enter:switch  Esc:cancel                         |
| > process-definitions                                            |
|   process-instances                                              |
+------------------------------------------------------------------+
| /search-term                                                     | <- Search bar (/) - shown/hidden
+------------------------------------------------------------------+
+========= process-definitions — 42 items =========================+
| KEY            NAME                      VERSION   RESOURCE       | <- Table header (uppercase)
| > my-proc      My Process                3         my-proc.bpmn   | <- Selected row (inverted colors)
|   other-proc   Other Process             1         other.bpmn     |
|   ...                                                            |
+==================================================================+
<process-definitions> | Ready                           | ⚡ 34ms    <- Footer (breadcrumb | status | remote)
```

**Key observations:**
- Context popup and search bar are **mutually exclusive overlays** — they shrink the content box
- Content box title is embedded in the top border and adapts to state
- Footer breadcrumb uses environment `ui_color` as background
- All modals render **on top** of this layout — background stays visible

### Vertical Regions (in order)

| Region | Height | Always Visible |
|---|---|---|
| Header | 3 rows + 1 spacer = 4 rows | Yes |
| Context popup | 0 or 4-13 rows | Only when `:` pressed |
| Search bar | 0 or 1 row | Only when `/` active |
| Content box | Fills remainder | Yes |
| Footer | 1 row | Yes |

### Header

- **Row 1 (status line):** App version, environment name with status indicator, auto-refresh badge (`circular arrow` in accent color when enabled)
- **Row 2 (key hints):** Priority-based list of key+description pairs, space-separated
- **Row 3:** Empty spacer
- Full terminal width, 1-character horizontal padding, bold font, no forced background color

### Key Hint Priority System

| Priority | Width Threshold | Examples |
|---|---|---|
| 1 | Always | `? help` |
| 2 | Always | `: switch` |
| 3 | Always | `Up/Down nav`, `PgDn/PgUp page` |
| 4 | 85+ | `Enter drill`, `e edit var` |
| 5 | 88+ | `s sort`, `Esc back` |
| 6 | 100+ | `Space actions`, `Ctrl+R refresh` |
| 7 | 100+ | `Ctrl+D terminate` |
| 8 | 110+ | `Ctrl+C quit` |
| 9 | 105+ | `Ctrl+E env` |

### Footer

Three columns:

| Column | Content | Style |
|---|---|---|
| Left | Breadcrumb context tag (e.g., `<process-instance>`) | Environment ui_color as background, black text |
| Center | Status message (errors, info, success) | Error: red+bold. Success: green. Info: blue. Auto-clears after 5s |
| Right | Remote activity indicator + optional latency | `lightning bolt` flashes 200ms on API calls. `L` toggles latency display |

### Splash Screen

On startup: animated ASCII logo reveal over 15 frames (~1.2s total). Version number appears at frame 8. Centered. Bypassed with `--no-splash`. Any key skips.

```
         ____
  ____  ( __ )____
 / __ \/ __  / __ \
/ /_/ / /_/ / / / /
\____/\____/_/ /_/

         v0.1.0
```

---

## 11. Keyboard Reference

### Global

| Key | Action |
|---|---|
| `?` | Open help (scrollable) |
| `:` | Open context switcher |
| `/` | Open search |
| `Ctrl+C` | Quit (with confirmation) |
| `Ctrl+E` | Open environment picker |
| `Ctrl+T` | Open theme picker |
| `r` / `Ctrl+R` | Toggle auto-refresh (5s interval) |
| `L` | Toggle latency display |

### Navigation

| Key | Action |
|---|---|
| `Up` / `k` | Move selection up |
| `Down` / `j` | Move selection down |
| `Enter` / `Right` | Drill down |
| `Esc` | Go back one level |
| `PgDn` / `Ctrl+F` | Next page |
| `PgUp` / `Ctrl+B` | Previous page |
| `gg` | Jump to first row |
| `G` | Jump to last row |
| `Ctrl+U` | Half-page scroll up |
| `Ctrl+D` | Half-page scroll down |
| `1`-`4` | Jump to breadcrumb level N |

### View Actions

| Key | Action |
|---|---|
| `s` | Sort popup |
| `Space` | Actions menu for selected row |
| `y` | Detail view (JSON) |
| `e` | Edit value (when editable columns exist) |

### Modal Navigation

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Cycle focus (fields, buttons) |
| `Enter` | Confirm / Save |
| `Esc` | Cancel / Close |
| `Space` | Toggle boolean fields |

---

## 12. Resource Types

35 resource types defined in `o8n-cfg.yaml`:

**Core:** process-definition, process-instance, process-variables, task, job, job-definition, external-task, incident, execution, variable-instance

**Administration:** authorization, user, group, tenant, filter, batch, batch-statistics, deployment, event-subscription, decision-definition, decision-requirements-definition

**History:** history-process-instance, history-activity-instance, history-task, history-job-log, history-incident, history-detail, history-external-task-log, history-identity-link-log, history-user-operation, history-variable-instance, history-batch

**Reports:** history-process-definition-cleanable-process-instance-report, history-decision-definition-cleanable-decision-instance-report, history-batch-cleanable-batch-report

---

## 13. State Persistence

### What Is Persisted (o8n-stat.yml)

- Active environment name
- Active skin name
- Latency display toggle
- Last navigation position (root context, breadcrumb path, selected definition/instance, query params)

### Session Restoration

On startup, the app restores:
- The last active environment (or falls back to "local" / first available)
- The last active skin (or falls back to `stock`)
- The last navigation position (root context and drilldown state)

### Debug Mode

- `--debug` flag enables verbose logging
- Creates `./debug/` directory automatically
- `debug/o8n.log` — error and debug messages (appended continuously)
- `debug/screen-{timestamp}.txt` — screen dump on panic
- `debug/last-screen.txt` — last rendered frame (updated each render cycle)

---

## 14. Testing

### Test Patterns

- **API tests:** `httptest.NewServer()` with Basic Auth verification
- **Config tests:** temp files with `defer os.Remove()`
- **UI tests:** message dispatch and model state assertions
- **Validation tests:** type-specific input cases with error message checks
- **Search/sort/pagination tests:** state transitions and boundary conditions
- **Navigation tests:** stack push/pop, state restoration, breadcrumb updates

Test files live next to the code they test. New feature tests use the pattern `main_<feature>_test.go`.

### Commands

```bash
make test       # Clear cache and run all tests
make cover      # Generate HTML coverage report (cov.out)
go test ./...   # Standard test run
go vet ./...    # Static analysis
```

---

## 15. Build & Run

```bash
# Build
go build -o o8n .

# Run
./o8n                   # Normal startup
./o8n --debug           # Enable debug logging
./o8n --no-splash       # Skip splash screen
./o8n --skin dracula    # Override skin at startup

# Regenerate API client (requires Docker)
./.devenv/scripts/generate-api-client.sh
```

---

## 16. Implementation Notes

- Variables endpoint may return different JSON structures; `FetchVariables` must try `map[string]{value,type}` first, then fall back to array decoding
- When computing column widths, normalize percentages if they don't sum to 100
- Table rows must be normalized to the configured column count to prevent panics
- Always guard UI render code with `recover()` to prevent crashes from malformed API responses
- Context selection: no inline completion until at least one character is typed
- `e` is reserved for edit modal; environment switching uses `Ctrl+E`
- Tab completion in context popup: must check `m.showRootPopup && len(m.rootInput) > 0`

---

## Appendix: API Endpoints

Key Operaton REST API endpoints used:

```
GET    /process-definition
GET    /process-instance?processDefinitionKey=<key>
GET    /process-instance/{id}/variables
DELETE /process-instance/{id}
GET    /<resource>?firstResult=<offset>&maxResults=<pageSize>
GET    /<resource>/count
PUT    /process-instance/{id}/variables/{name}
```

All endpoints support Basic Auth. Full API coverage via the generated OpenAPI client.
