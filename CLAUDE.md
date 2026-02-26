# o8n — Claude Code Context

A k9s-inspired terminal UI for managing Operaton workflow engines, built in Go using the Charmbracelet TUI ecosystem (Bubble Tea, Bubbles, Lipgloss).

---

## ⚠️ CRITICAL - READ FIRST

**NEVER delete or empty these files:**
- **o8n-cfg.yaml** (760+ lines) - Table definitions, column specs, drilldown rules
- **o8n-env.yaml** (~10 lines) - Local environment config with credentials

**BEFORE EACH COMMIT:**
```bash
wc -l o8n-*.yaml  # Verify: ~760 cfg, ~10 env
git diff o8n-*.yaml  # Should show NO deletions
```

**IF ACCIDENTALLY DELETED:**
```bash
git show HEAD~2:o8n-cfg.yaml > o8n-cfg.yaml
# Then restore o8n-env.yaml with local settings
```

See the critical section above for details.

---

## Commands

```bash
make test       # Clear test cache and run all tests
make cover      # Generate HTML coverage report (cov.out)
make build      # Build binary to execs/o8n
go test ./...   # Standard test run
go vet ./...    # Static analysis
gofmt -w .      # Format code

./o8n --debug       # Run with debug logging → ./debug/access.log + ./debug/last-screen.txt
./o8n --no-splash   # Skip splash screen and go directly to main view
```

## Architecture

### Key Files

| File | Purpose |
|------|---------|
| `main.go` | Thin entry point, calls `internal/app.Run()` |
| `internal/app/` | All TUI logic: model, update, view, nav, commands, skin, styles, table, edit |
| `internal/client/` | Modern API client with reflection-based parameter binding |
| `internal/config/` | Domain models and config structs |
| `internal/dao/` | DAO interfaces (`DAO`, `HierarchicalDAO`, `ReadOnlyDAO`) |
| `internal/validation/` | Input validation for edit dialogs (bool/int/float/json/text) |
| `internal/contentassist/` | Thread-safe user suggestion cache |
| `internal/operaton/` | **Auto-generated** OpenAPI client — do not edit manually |

### Bubble Tea Pattern

The app follows the Elm-inspired model/update/view cycle:

- **`Init()`** — initial commands (splash screen, config load)
- **`Update(msg tea.Msg)`** — pure message handlers, return new model + commands
- **`View()`** — render to string

All async operations (API calls) return `tea.Cmd` and communicate results back via typed messages (`dataLoadedMsg`, `instancesLoadedMsg`, `errMsg`, etc.).

### Configuration (Split Files)

- **`o8n-env.yaml`** — credentials, URLs, ui_color per environment — git-ignored, 0600 perms
- **`o8n-cfg.yaml`** — table definitions, column specs, drilldown rules — committable
- `LoadSplitConfig()` merges both into a unified `Config` struct at runtime

### Navigation Model

```
Process Definitions → Process Instances → Variables
```

State is captured in `navigationStack []viewState` with Esc popping back. Breadcrumb shown in footer. Context switching (`:` key) allows jumping to any resource type.

### Skins

36 built-in color themes in `skins/` as YAML files. Switched at runtime. `ui_color` in env config sets environment-specific accent color.

## Testing Patterns

- API tests: `httptest.NewServer()` with Basic Auth verification
- Config tests: temp files with `defer os.Remove()`
- UI tests: message dispatch and model state assertions
- Validation tests: type-specific input cases with error message checks

Test files live next to the code they test. New feature tests use the pattern `main_<feature>_test.go`.

## Important Conventions

- **Do not edit `internal/operaton/`** — regenerate with `.devenv/scripts/generate-api-client.sh`
- Nullable types from the generated client use helper wrappers: `NullableString`, `NullableInt32`, `NullableBool`
- Modal types: `ModalNone`, `ModalConfirmDelete`, `ModalConfirmQuit`, `ModalHelp`, `ModalEdit`
- Destructive actions require double-press confirmation before executing
- Errors display in the footer with auto-clear after 5 seconds
- Responsive layout: columns auto-hide below certain terminal widths; hints have visibility priority 1–9
- Build uses `CGO_ENABLED=0` and `GO_TAGS=netgo` for static binaries
