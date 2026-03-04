# o8n

A terminal UI for Operaton

```
         ____
  ____  ( __ )____
 / __ \/ __  / __ \
/ /_/ / /_/ / / / /
\____/\____/_/ /_/
```

**o8n** is a keyboard-first terminal UI for managing [Operaton](https://operaton.org) workflow engines, inspired by [k9s](https://k9scli.io). Browse 35 resource types, drill into process instances, edit variables, execute actions — all without leaving your terminal.

## Features

- **35 resource types** — Process definitions, instances, tasks, jobs, incidents, history, deployments, and more
- **Config-driven actions** — Suspend, resume, delete, retry, claim, complete — all defined in YAML
- **Drill-down navigation** — Definition -> Instances -> Variables with full state restoration on back
- **Navigation actions** — The actions menu separates HTTP mutations from view-style navigations (`→` suffix) and the help screen now lists these view shortcuts under a dedicated **VIEWS** section
- **Live search & sort** — `/` to filter rows, `s` to sort by any column
- **35 color themes** — Dracula, Nord, Gruvbox, Solarized, and more with live preview via `Ctrl+T`
- **Multi-environment** — Switch between local, staging, production with `Ctrl+E`
- **Auto-refresh** — Toggle with `r` for 5-second polling with visual indicator
- **Overlay modals** — Help, edit, sort, detail view, confirmations — all rendered over live content
- **Responsive layout** — Columns auto-hide on narrow terminals; hints adapt to width
- **Persistent state** — Active environment, skin, and last navigation position restored on startup
- **Two-step confirmations** — Destructive actions require double-press for safety

## Quick Start

### 1. Configure

```bash
cp o8n-env.yaml.example o8n-env.yaml
# Edit with your Operaton environment(s)
```

```yaml
environments:
  local:
    url: http://localhost:8080/engine-rest
    username: demo
    password: demo
    ui_color: "#00A8E1"
    default_timeout: 10s
```

### 2. Build & Run

```bash
go build -o o8n .
./o8n
```

## Keyboard Shortcuts

### Global

| Key | Action |
|---|---|
| `?` | Help screen |
| `:` | Context switcher (jump to any resource type) |
| `/` | Search (live row filtering) |
| `Ctrl+C` | Quit (with confirmation) |
| `Ctrl+E` | Environment picker |
| `Ctrl+T` | Theme picker (live preview) |
| `Ctrl+H` | Home context picker (reopens first-run selection) |
| `r` | Toggle auto-refresh |
| `L` | Toggle API latency display |

### Navigation

| Key | Action |
|---|---|
| `Up` / `Down` | Move selection |
| `Enter` or `Right` | Drill down |
| `Esc` | Go back (restores cursor and data) |
| `PgDn` / `PgUp` | Server-side pagination |
| `Home` / `End` | Jump to first / last row |
| `1`-`4` | Jump to breadcrumb level |

**Vim mode** (`--vim` flag or `vim_mode: true` in config) adds: `j`/`k` navigation, `gg`/`G` first/last, `Ctrl+U`/`Ctrl+D` half-page scroll.

### Actions

| Key | Action |
|---|---|
| `Ctrl+Space` | Open actions menu for selected row |
| `J` | View raw JSON detail |
| `e` | Edit value (on editable columns) |
| `s` | Sort by column |
| `Ctrl+D` | Delete/terminate (with confirmation) |

Actions are resource-specific and defined in `o8n-cfg.yaml`. Press `Space` on any row to see available actions.

Mutation actions (HTTP verbs) are listed first, followed by view-style navigation actions that show a `→` suffix and are separated from the mutations; the help screen also surfaces these navigation actions under a **VIEWS** section for quick reference.

## Configuration

Three files with distinct roles:

| File | Purpose | Committed |
|---|---|---|
| `o8n-env.yaml` | Environment URLs, credentials, accent colors | No (git-ignored) |
| `o8n-cfg.yaml` | Table definitions, columns, actions, drilldowns | Yes |
| `o8n-stat.yml` | Runtime state (active env, skin, last position) | No (auto-generated) |

See [specification.md](specification.md) for the full configuration reference.

## Theming

35 built-in skins. Switch at runtime with `Ctrl+T`:

**Dark:** dracula, nord, gruvbox-dark, one-dark, nightfox, kanagawa, monokai, solarized-dark, everforest-dark, snazzy, vercel, rose-pine, rose-pine-moon, in-the-navy, and more

**Light:** gruvbox-light, everforest-light, solarized-light, rose-pine-dawn, and more

**Special:** o8n-cyber, transparent, stock, kiss

Add custom skins by placing a YAML file in `skins/`. All colors use 25 semantic roles — no hardcoded values.

## Debug Mode

```bash
./o8n --debug
```

Creates `./debug/` with:
- `o8n.log` — Error and debug messages
- `last-screen.txt` — Last rendered TUI frame
- `screen-*.txt` — Screen dumps on panic

Other flags: `--no-splash` (skip startup animation), `--skin <name>` (override theme), `--vim` (enable vim keybindings).

## Development

### Prerequisites

- Go 1.24+
- Docker (for API client generation)

### Commands

```bash
make build      # Build binary to execs/o8n
make test       # Clear cache and run all tests
make cover      # HTML coverage report
go vet ./...    # Static analysis
gofmt -w .      # Format code
```

### API Client Regeneration

```bash
./.devenv/scripts/generate-api-client.sh
```

Generates Go client from `resources/operaton-rest-api.json` into `internal/operaton/`. Do not edit generated files manually.

### Project Structure

```
o8n/
├── main.go                  # Entry point
├── internal/
│   ├── app/                 # TUI logic (model, update, view, nav, commands)
│   ├── client/              # Operaton REST API client wrapper
│   ├── config/              # Config structs and loaders
│   ├── validation/          # Input validation (bool/int/float/json/text)
│   ├── contentassist/       # User suggestion cache
│   ├── dao/                 # Data access interfaces
│   └── operaton/            # Auto-generated OpenAPI client
├── skins/                   # 35 color theme YAML files
├── resources/               # OpenAPI specification
├── o8n-env.yaml             # Environment credentials (git-ignored)
├── o8n-cfg.yaml             # UI table definitions (version-controlled)
└── o8n-stat.yml             # Runtime state (auto-generated)
```

## Documentation

- [specification.md](specification.md) — Complete technical specification (architecture, behavior, configuration reference)

## Security

- `o8n-env.yaml` is git-ignored and should have `chmod 600` permissions
- Never commit credentials to version control
- Use `o8n-env.yaml.example` as a template

## License

See [LICENSE](LICENSE) file.

## Contributing

Contributions welcome. Read [specification.md](specification.md) for architecture details before submitting PRs.
