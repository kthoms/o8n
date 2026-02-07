o8n — Specification

Overview
--------
This document specifies the o8n terminal UI application (written in Go). It fully describes the architecture, data models, UI layout and behavior, configuration files, API interactions, error handling, and tests. The intent is that a developer or an AI coder can reimplement the same structure and behavior from this specification.

High-level summary
------------------
- Language: Go (>= 1.24).
- Primary runtime library: Charmbracelet ecosystem (bubbletea, bubbles, lipgloss).
- Function: Provide a terminal UI to browse and operate on resources exposed by an Operaton/engine-rest API (process definitions, instances, variables, tasks, jobs, etc.).
- Important UX rules:
  - Vertical layout: Header (8 rows) -> Context selection (1 row boxed, dynamic) -> Main content (boxed table filling remaining height minus footer) -> Footer (1 row).
  - Default environment: "local" (if present) or first configured environment.
  - Default root context at start: process-definitions.
  - A flash indicator (⚡) appears on the bottom-right for 0.2s whenever a remote REST call is issued.
  - Context selection opened with `:`; offers inline completion (only after first char); Tab completes; Enter switches only on exact match; Esc cancels.
  - Drill-down interaction: definitions -> instances -> variables (name, value, type). Arrow keys move selection; Enter drills in; Esc goes back.

Repository layout (conceptual)
-----------------------------
- main.go — main application, model, update/view, UI wiring.
- api.go — HTTP client code interacting with Operaton engine-rest endpoints.
- config.go — config loading/saving for environment and app config.
- o8n-env.yaml.example — example environment file with ui_color and credentials (sensitive, do not commit real secrets).
- o8n-env.yaml (runtime) — environment configuration (kept out of source control).
- o8n-cfg.yaml — application-level configuration (table definitions, column specs).
- resources/operaton-rest-api.json — OpenAPI / swagger reference used to derive root contexts.
- tests (go test unit tests) — exercises config load and API client behaviours.

Architecture and major components
---------------------------------
1. Program model (main.go)
   - Uses the Bubble Tea model pattern with Init, Update, and View.
   - Model fields include:
     - config *Config — compatibility view combining environment and table definitions.
     - envConfig *EnvConfig and appConfig *AppConfig — split configs;
     - UI components: bubbles list.Model (for a list-based view) and bubbles table.Model for the main table.
     - UI state: currentEnv, envNames, currentRoot (root context like `process-definitions`), viewMode ("definitions","instances","variables"), selectedInstanceID, autoRefresh, flashActive, footerError, showRootPopup, rootInput.
     - Layout sizing: lastWidth/lastHeight, paneWidth/paneHeight.
   - Main event loop reacts to key presses and window size messages. It orchestrates API fetch commands and updates the table/list state.

2. API client (api.go)
   - Client struct holds the environment info and *http.Client with a timeout.
   - Primary methods:
     - FetchProcessDefinitions() -> []ProcessDefinition
     - FetchInstances(processDefinitionKey) -> []ProcessInstance
     - FetchVariables(instanceId) -> []Variable
     - TerminateInstance(instanceId) -> error
   - FetchVariables robustly handles the two formats the REST API may return: a map keyed by variable name (value/type), or an array; returns []Variable with Name, Value (stringified), Type.
   - Errors are returned and propagated back as errMsg messages.

3. Configuration (config.go + YAML files)
   - EnvConfig (o8n-env.yaml):
     - environments: map[string]{ url, username, password, ui_color }
     - active: string (name of active environment)
   - AppConfig (o8n-cfg.yaml):
     - tables: array of TableDef (name, columns)
     - ColumnDef: { name, visible, width (percent string like "25%"), align }
   - For compatibility the application exposes LoadConfig(path) to load legacy single-file config.yaml used by tests; also provides LoadEnvConfig and LoadAppConfig and Save helpers. SaveConfig writes back to o8n-env.yaml and o8n-cfg.yaml (best-effort).

4. UI Rendering (main.go View)
   - Vertical layout in exact order:
     1. Header (8 rows unboxed area, 3 columns)
        - Column 1 (left, width 25%): Context information lines (Environment, API URL, User). This column must be left-aligned and use the header area height of 8 rows.
        - Column 2 (center): Key bindings/help text (list of keys and short descriptions).
        - Column 3 (right, fixed width 25 characters): ASCII art logo aligned right.
     2. Context selection box (1 row boxed) — visible when `:` is pressed; otherwise an empty boxed placeholder. It contains a single-line text input.
        - Behavior: input is empty initially; no completions shown until first character typed; inline completion (gray color) shows the remaining suffix of the first matching root context; Tab completes; Enter tries to switch context only if exact match; Esc cancels and clears input.
     3. Main content (boxed table) — uses table.Model to render rows and headers. The table consumes the remaining height between header/context selection/footer. Table header text is uppercase and uses a header color (white). Column widths are computed using the configured percent widths from AppConfig when present; otherwise columns get equal widths.
     4. Footer (1 row) — three columns separated by " | ":
        - Column 1: Context tag (e.g. "<process-definitions>") — left aligned, largest column (space for long names like historic-external-task-log); uses environment UI color as background and black text.
        - Column 2: Message — used to display errors (footerError) or other status. Truncated as needed.
        - Column 3: Remote access indicator — single character; displays "⚡" for 0.2s after issuing a REST call; otherwise blank.

5. UI Styles and Colors
   - Primary border and accent color: environment.UIColor (from env config).
   - Table header foreground color is white.
   - Completion suffix is gray (#666666) by default.
   - Footer error message style: red (#FF0000) and bold.
   - The context selection input uses the environment border color as its font color by default (config allows extension to custom color names).

6. Keybindings and behavior
   - q: quit
   - e: switch environment (cycles through configured envNames). Persists active environment to o8n-env.yaml (best-effort).
   - r: toggle auto-refresh. When enabled, periodically refresh definitions with interval (5s). When disabled, manual selection changes may set a manualRefreshTriggered flag used for UI hints.
   - : (colon): toggle context-selection input box. Focus is inside the input when open.
     - Typing: appends a single rune per key event to rootInput.
     - Completion: only when rootInput length >= 1; inline suffix shown for the first match.
     - Tab: completes rootInput to the first matching root context.
     - Enter: switch context only when rootInput exactly equals a known root context. When switched:
       - close the input box
       - clear footer errors
       - set currentRoot
       - fetch top-level data for the new context (e.g., process-definitions table rows)
       - set focus of table to the first row (implementation note below)
     - Esc: cancel input and close box.
   - Arrow keys / j/k: move selection in list/table. Important: in the process-definition view the arrow keys navigate definitions only; Enter is required to drill down into instances.
   - Enter: drills down from definitions -> instances -> variables. In variables view Enter does nothing.
   - x: mark selected instance for kill; y confirms and calls TerminateInstance.

7. Fetching & Flow
   - On Init: program sends fetchDefinitionsCmd and triggers the flash (flashOnCmd). fetchDefinitionsCmd uses the current environment to call Client.FetchProcessDefinitions.
   - When Enter on a definition selected: fetchInstancesCmd(key) is issued which calls Client.FetchInstances(key) and then applyInstances.
   - When Enter on an instance: fetchVariablesCmd(id) issued which calls Client.FetchVariables and then applyVariables.

8. Error handling
   - All API errors are returned as errMsg messages and the main Update() stores the error text in m.footerError and schedules a clear after 4 seconds (clearErrorMsg).
   - Rendering functions (applyDefinitions, applyInstances, applyVariables) use defer/recover to catch panics and set m.footerError with an explanatory message rather than letting the program crash.
   - The UI will show error messages prominently in the footer message column with the error style.

9. Table definition config: o8n-cfg.yaml
   - Top-level: tables: array of { name: string, columns: [ { name, visible, width, align } ] }
   - Example table def for process-definitions:
     - name: process-definitions
       columns:
         - name: key
           visible: true
           width: 20%
           align: left
         - name: name
           visible: true
           width: 40%
           align: left
         - name: version
           visible: true
           width: 15%
           align: center
         - name: resource
           visible: true
           width: 25%
           align: left
   - Column width interpretation: the UI parses the percent string (e.g., "25%") and computes absolute character widths based on available content width; remaining percentage split equally among unspecified columns.

10. Data model (Go structs used by the API client)
    - Environment { URL, Username, Password, UIColor }
    - ColumnDef { Name, Visible bool, Width string, Align string }
    - TableDef { Name string, Columns []ColumnDef }
    - EnvConfig { Environments map[string]Environment, Active string }
    - AppConfig { Tables []TableDef }
    - ProcessDefinition { ID, Key, Category, Description, Name, Version int, Resource, DeploymentID, Diagram, Suspended bool, TenantID }
    - ProcessInstance { ID, DefinitionID, BusinessKey, CaseInstanceID, Ended bool, Suspended bool, TenantID, StartTime, EndTime }
    - Variable { Name, Value string, Type string }

11. Messages and Commands (Bubble Tea)
    - Messages (Go types): refreshMsg, definitionsLoadedMsg, instancesLoadedMsg, variablesLoadedMsg, dataLoadedMsg, errMsg, flashOnMsg, flashOffMsg, clearErrorMsg, terminatedMsg
    - Commands: fetchDefinitionsCmd, fetchInstancesCmd, fetchVariablesCmd, fetchDataCmd, terminateInstanceCmd, flashOnCmd
    - flashOnCmd returns flashOnMsg that Update() uses to set m.flashActive and schedule flashOffMsg after 200ms.

12. Tests
    - Unit tests cover: API client behaviors (mock HTTP server), config loading (legacy and split), and model functions (applyData populates table rows, selection change behaviour triggers manualRefresh flag, flash on/off). Tests are run with `go test ./...`.

13. Build and run
    - Build:
```bash
go build -o o8n .
```
    - Run:
```bash
./o8n
```
    - Tests:
```bash
go test ./... -v
```

14. Implementation notes and edge cases
    - The variables endpoint may return different JSON structures; FetchVariables must try decoding to map[string]{value,type} first and fall back to array decoding (re-request or reuse body accordingly).
    - When computing column widths from percentages, normalize percentages if they do not sum to 100. If some columns do not specify a width, distribute remaining percentage evenly.
    - Table rows must be normalized to the configured column count to avoid panics in the table renderer; implement a normalizeRows(rows, colsCount) helper to pad/truncate rows.
    - Always guard UI render code that depends on external data with recover() so the TUI doesn't crash because of a malformed API response.
    - When context selection is invoked, do not present inline completion until the user has typed at least one character. Tab completes to the first match; Enter requires exact match.
    - When switching root contexts programmatically, update the footer context tag and trigger fetch of the root context's primary resource. Also ensure the table selection/focus moves to the first row (implementation detail: use table.SetCursor or equivalent if available; otherwise ensure rows are selected and visible on first render).

15. Developer checklist to reimplement
    - Create Go module, add dependencies: bubbletea, bubbles, lipgloss, yaml.
    - Define configuration structs and YAML load/save functions.
    - Implement API client with robust JSON parsing for variables.
    - Implement Bubble Tea model with the fields and helpers described.
    - Implement Update() key handling and command orchestration.
    - Implement View() rendering the header (8 rows), boxed context selection (1 row), main boxed table (remaining height minus footer), and footer (1 row) with three columns and separators.
    - Add tests for config loading, client behavior, and simple UI logic.

16. Example flows
    - Start app -> default env selected -> fetch process definitions -> show table -> user navigates with arrows -> press Enter on definition -> fetch instances -> show instances -> press Enter on instance -> fetch variables -> show variables table (name, value, type)
    - User presses `:` -> types `proc` -> inline completion suggests `ess-definitions` remainder -> Tab completes to `process-definitions` -> Enter switches context (if exact) and reloads table.

Appendix — Important API endpoints (Operaton engine-rest example)
- GET /process-definition
- GET /process-instance?processDefinitionKey=<key>
- GET /process-instance/{id}/variables
- DELETE /process-instance/{id}

Contact
-------
If details are unclear or you want the spec extended (for example to include exact unit-test cases, more detailed layout mockups, or a list of exact table column mappings for all root contexts), tell me which parts to expand and I will update the specification.md accordingly.
