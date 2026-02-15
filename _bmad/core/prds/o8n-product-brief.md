# Product Brief: o8n - Terminal UI for Operaton

**Version:** 0.1.0  
**Author:** Karsten Thoms  
**Date:** February 15, 2026  
**Status:** In Development

---

## Executive Summary

**o8n** is a modern, terminal-based user interface (TUI) for interacting with the Operaton BPM (Business Process Management) engine. Built in Go using the Charmbracelet ecosystem, o8n provides developers and operators with a fast, keyboard-driven experience for managing process definitions, instances, tasks, and other workflow resources directly from the command line.

The application follows design principles inspired by popular terminal tools like **k9s** (Kubernetes CLI), offering intuitive navigation, real-time monitoring, responsive layouts, and a focus on efficiency without requiring a graphical desktop environment.

---

## Problem Statement

Modern BPM engines like Operaton expose comprehensive REST APIs for process orchestration, but working with these APIs typically requires:

1. **Web-based UIs** that consume significant resources and require browser context-switching
2. **Manual API calls** using curl or Postman, which are verbose and error-prone
3. **Custom scripts** that must be maintained for each environment and use case

No efficient, keyboard-driven terminal interface allows operators to:
- Browse and inspect process definitions and instances
- Monitor running workflows in real-time
- Drill down into process variables and task states
- Execute administrative operations (terminate, retry, etc.)
- Switch between multiple environments seamlessly

---

## Target Users

### Primary Personas

1. **Backend Developers**
   - Need to test and debug process definitions during development
   - Want quick access to instance state and variables
   - Prefer terminal-based workflows integrated with their development environment

2. **DevOps Engineers / SREs**
   - Monitor production BPM instances
   - Investigate incidents and troubleshoot failed processes
   - Require rapid environment switching (dev, staging, production)

3. **Automation Engineers**
   - Integrate process monitoring into CI/CD pipelines
   - Need scriptable, headless access to BPM state
   - Value consistent, reproducible interfaces

### Secondary Personas

4. **Technical Process Analysts**
   - Review process execution patterns
   - Validate business logic implementation
   - Generate reports from process data

---

## Core Value Propositions

| Value | Description |
|-------|-------------|
| **Speed** | Terminal-native performance with instant startup and keyboard-driven navigation |
| **Efficiency** | No context switching between browser tabs; work alongside code and logs |
| **Flexibility** | Multi-environment support with secure credential management |
| **Visibility** | Real-time auto-refresh, drill-down navigation, and comprehensive resource views |
| **Customization** | Theme support, configurable tables, and extensible command structure |
| **Simplicity** | Zero configuration to get started; sensible defaults with optional customization |

---

## Key Features

### 1. **Multi-Environment Management**
- Configure multiple Operaton environments (local, dev, staging, production)
- Switch environments on-the-fly with a single keypress (`e`)
- Secure credential storage in `o8n-env.yaml` (excluded from version control)
- Active environment persisted across sessions

### 2. **Responsive Terminal Layout**
- **Header Section** (8 rows): Environment context, key bindings, ASCII logo
- **Context Selection** (dynamic): Inline command palette with fuzzy completion
- **Main Content** (responsive table): Adapts to terminal size, column visibility rules
- **Footer** (1 row): Breadcrumb navigation, status messages, remote call indicator

### 3. **Resource Navigation & Drill-Down**
- Navigate process definitions → instances → variables
- Arrow keys for selection, Enter for drill-down, Esc to go back
- Support for all major Operaton resources (tasks, jobs, historic data, etc.)
- Configurable table definitions for each resource type

### 4. **Real-Time Monitoring**
- Auto-refresh mode (`r` to toggle) with configurable intervals
- Visual indicator (⚡) when REST API calls are in progress
- Graceful error handling with footer status messages

### 5. **Theming & Customization**
- 25+ built-in color schemes (Dracula, Nord, Gruvbox, Solarized, etc.)
- Per-environment theme selection
- Configurable table layouts: column visibility, width, alignment

### 6. **Robust API Integration**
- Generated Go client from Operaton OpenAPI specification
- Comprehensive coverage of REST API endpoints
- Automatic error recovery and user-friendly error messages

---

## User Journeys

### Journey 1: Developer Debugging a Process Instance

**Actor:** Backend Developer  
**Goal:** Inspect variables of a failed process instance

1. Launch `o8n` → defaults to local environment, shows process definitions
2. Press `r` to enable auto-refresh
3. Arrow down to select "ReviewInvoice" definition
4. Press Enter → drills into instances view
5. Arrow to select specific failed instance (sorted by start time)
6. Press Enter → views all process variables (name, value, type)
7. Identifies incorrect variable value causing failure
8. Press Esc twice → returns to definitions view

**Outcome:** Issue identified in <30 seconds without leaving terminal

---

### Journey 2: DevOps Engineer Investigating Production Incident

**Actor:** DevOps Engineer  
**Goal:** Terminate stuck process instances in production

1. Launch `o8n` → starts in last-used environment (production)
2. Press `:` → context selection dialog appears
3. Type "task" → autocomplete suggests "task"
4. Press Tab → completes to "task", press Enter → switches to tasks view
5. Sorts by create date, identifies stuck tasks
6. Press `:`, type "proc", Tab/Enter → switches to process-instances
7. Filters to running instances (future feature: inline filtering capability)
8. Arrow to stuck instance, press `x` → confirmation prompt
9. Press `y` → instance terminated, API call indicator flashes
10. Auto-refresh shows updated state

**Outcome:** Incident resolved with rapid context switching and safe confirmations

---

### Journey 3: Engineer Switching Environments

**Actor:** Backend Developer  
**Goal:** Compare process behavior between local and staging

1. Start o8n → local environment active
2. Navigate to instance, note variable values
3. Press `e` → cycles to next configured environment (staging)
4. Header updates to show staging URL and credentials
5. Same instance ID exists, variables differ
6. Press `e` again → cycle back to local
7. Active environment persisted to `o8n-env.yaml`

**Outcome:** Cross-environment comparison completed in seconds

---

## Technical Architecture

### Technology Stack

- **Language:** Go 1.24+
- **UI Framework:** Bubble Tea (Elm-inspired TUI framework)
- **Components:** Bubbles (reusable Bubble Tea components)
- **Styling:** Lipgloss (layout and color management)
- **API Client:** OpenAPI-generated client from `operaton-rest-api.json`

### Key Design Principles

1. **Elm Architecture Pattern**
   - Model (application state)
   - Update (message handling and state transitions)
   - View (rendering logic)

2. **Configuration Separation**
   - **o8n-env.yaml:** Environment credentials and secrets (not committed)
   - **o8n-cfg.yaml:** Application settings and table definitions (committed)

3. **Graceful Degradation**
   - Provide sensible defaults when config files are missing
   - Recovers from API errors without crashing
   - Column hiding when terminal width insufficient

4. **Responsive Design**
   - Tables adapt to terminal size changes
   - Column width percentages with minimum viable sizes
   - Fixed and flexible layout sections

---

## Configuration Model

### Environment Configuration (`o8n-env.yaml`)

```yaml
environments:
  local:
    url: "http://localhost:8080/engine-rest"
    username: "demo"
    password: "demo"
    ui_color: "#00FF00"
  staging:
    url: "https://staging.operaton.example.com/engine-rest"
    username: "staging-user"
    password: "***"
    ui_color: "#FFA500"
active: local
```

### Application Configuration (`o8n-cfg.yaml`)

```yaml
tables:
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
```

---

## Scope

### In Scope (v0.1.0)

✅ Process definitions, instances, and variables navigation  
✅ Multi-environment configuration and switching  
✅ Auto-refresh and manual navigation modes  
✅ Theme/skin support with 25+ built-in options  
✅ Responsive layout with configurable table definitions  
✅ Instance termination with confirmation  
✅ Context switching via command palette (`:`)  
✅ Graceful error handling and status messages

### In Scope (Future Releases)

🔮 **v0.2.0:** Task management (claim, complete, assign)  
🔮 **v0.3.0:** Historic data views and filtering  
🔮 **v0.4.0:** Inline filtering and search  
🔮 **v0.5.0:** Job retry and external task handling  
🔮 **v1.0.0:** Full API coverage, plugin system, export/reporting

### Out of Scope

❌ Process definition modeling (use Camunda Modeler)  
❌ Form rendering (use Operaton Tasklist)  
❌ User/group administration (use Operaton Admin)  
❌ Deployment of process definitions (use REST API or CI/CD)

---

## Non-Functional Requirements

### Performance

- **Startup Time:** < 100ms on modern hardware
- **Refresh Latency:** < 500ms for typical API calls
- **Memory Footprint:** < 50 MB during normal operation
- **Responsiveness:** UI updates within 16ms (60 FPS)

### Security

- Credentials stored in plain YAML (user's responsibility to protect)
- Recommended file permissions: `chmod 600 o8n-env.yaml`
- No credentials logged or transmitted insecurely
- HTTPS support for remote environments

### Usability

- Zero-config startup (defaults to local if present)
- Discoverable keyboard shortcuts (always visible in header)
- Consistent navigation patterns across all resource types
- Clear error messages with suggested remediation

### Compatibility

- Cross-platform: macOS, Linux, Windows (via WSL or native Go builds)
- Terminal requirements: ANSI color support, UTF-8, minimum 80x24 size
- Operaton API compatibility: Operaton 7.x and Camunda Platform 7.x

---

## Success Metrics

### Adoption Metrics

- **Downloads:** 300+ in the first 3 months after public release
- **GitHub Stars:** 100+ within 6 months
- **Active Users:** 50+ weekly active users by the sixth month

### Engagement Metrics

- **Session Duration:** Average 5+ minutes per session
- **Environment Switches:** 3+ switches per session (power user indicator)
- **Drill-Down Depth:** 80%+ of sessions navigate at least two levels deep

### Quality Metrics

- **Crash Rate:** < 0.1% of sessions
- **Error Recovery:** 95%+ of API errors handled gracefully
- **Test Coverage:** > 80% unit test coverage
- **User-Reported Issues:** < 10 critical bugs in first 3 months

---

## Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **API changes in Operaton** | High | Medium | Use generated client, version compatibility checks |
| **Terminal compatibility issues** | Medium | Low | Test across major terminal emulators and document requirements |
| **Credential security concerns** | High | Medium | Clear documentation, consider integration with OS keychain (future) |
| **Performance with large datasets** | Medium | Medium | Implement pagination, lazy loading, configurable limits |
| **Adoption by non-terminal users** | Low | High | This is acceptable because the target audience is CLI-native users |

---

## Open Questions

1. **Should we support Windows Command Prompt natively?**
   - Current decision: Support WSL and PowerShell only (revisit if demand exists)

2. **How should filtering be implemented?**
   - Options: Inline fuzzy search, SQL-like query language, or predefined filters
   - Lean toward inline fuzzy search for simplicity

3. **Should we add mouse support?**
   - Pro: Lower barrier to entry
   - Con: Dilutes keyboard-first philosophy
   - Decision: Optional mouse support without compromising keyboard UX

4. **Plugin architecture priority?**
   - Custom commands, integrations, and extensions
   - Target for v1.0.0

---

## Next Steps

### Immediate Actions (Sprint 1)

1. ✅ Core UI layout and navigation implemented
2. ✅ Process definitions/instances/variables working
3. 🔄 Comprehensive test suite (in progress)
4. 🔄 Documentation: README, user guide, configuration reference

### Short-Term (Next 3 Sprints)

1. Add task management operations (claim, complete)
2. Implement filtering and search
3. Historic data views (activity, incidents)
4. Packaging and distribution (Homebrew, apt, releases)

### Medium-Term (Next 6 Months)

1. Full API coverage (jobs, external tasks, deployments)
2. Plugin system design and implementation
3. Export/reporting capabilities
4. Integration with OS keychains for secure credentials

---

## Appendix: Design Inspirations

- **k9s:** Kubernetes terminal UI (layout, navigation patterns)
- **htop/btop:** System monitoring (real-time updates, color coding)
- **lazygit:** Git terminal UI (keyboard shortcuts, drill-down)
- **Charmbracelet demos:** Modern TUI design patterns

---

## Appendix: Glossary

- **Operaton:** Open-source BPM platform (fork of Camunda 7)
- **BPM:** Business Process Management
- **TUI:** Terminal User Interface
- **Bubble Tea:** Go framework for building terminal applications (Elm architecture)
- **Drill-Down:** Navigation pattern from summary to detail views
- **Auto-Refresh:** Automatic periodic data updates
- **Context Switching:** Changing between different resource types (definitions, tasks, etc.)

---

**Document Control:**  
This product brief is a living document maintained in `_bmad/core/prds/o8n-product-brief.md`.  
Review and update as product requirements evolve.

