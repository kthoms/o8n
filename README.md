# o8n
A terminal UI for Operaton

```
         ____      
  ____  ( __ )____ 
 / __ \/ __  / __ \
/ /_/ / /_/ / / / /
\____/\____/_/ /_/ 
```

**o8n** is a powerful terminal UI for managing Operaton workflow engines, inspired by k9s.

## Quick Start

### 1. Configuration

There are two config files:

- `o8n-env.yaml` — Environment credentials and UI colors (keep secret)
- `o8n-cfg.yaml` — UI table definitions and app settings

Create your environment configuration:

```bash
cp o8n-env.yaml.example o8n-env.yaml
# Edit o8n-env.yaml to add your Operaton environments
```

Example `o8n-env.yaml`:
```yaml
environments:
  local:
    url: http://localhost:8080/engine-rest
    username: demo
    password: demo
    ui_color: "#00A8E1"
  production:
    url: https://operaton.example.com/engine-rest
    username: admin
    password: secret
    ui_color: "#FF5733"
active: local
```

### 2. Building

```bash
go build -o o8n .
```

### 3. Running

```bash
./o8n
```

## Usage

### Keyboard Shortcuts

**Global Actions:**
- `?` — Show help screen with all shortcuts
- `:` — Open context switcher (process-definition, process-instance, task, job, etc.)
- `<ctrl>+e` — Switch environment
- `<ctrl>+c` — Quit application

**Navigation:**
- `↑/↓` or `j/k` — Move selection up/down
- `Page Up/Down` — Jump through table
- `Enter` — Drill down (definitions → instances → variables)
- `Esc` — Go back one level

**View Actions:**
- `r` — Toggle auto-refresh (5s interval)
- `<ctrl>-r` — Manual refresh
- `/` — Filter/search (if implemented)

**Instance Actions:**
- `<ctrl>+d` — Delete selected instance (press twice to confirm)
- `<ctrl>+t` — Terminate instance

**Variable Actions:**
- `e` — Edit selected value (when column is editable)

### Features

**🎨 Theming & Skins**
- 20+ built-in color schemes (dracula, nord, gruvbox, solarized, etc.)
- Runtime skin switching
- Custom skin support in `/skins` folder

**📊 Dynamic Tables**
- Responsive column sizing based on terminal width
- Auto-hide low-priority columns on narrow terminals
- Customizable column visibility and widths in `o8n-cfg.yaml`
- Editable columns with type-aware input (opt-in)

**🔍 Context Switching**
- Fast context switching with `:` key
- Inline completion as you type
- Access all Operaton resources (process definitions, instances, tasks, jobs, incidents, etc.)

**⚡ Real-Time Updates**
- Auto-refresh mode with configurable intervals
- Visual indicator for API activity
- Error messages in footer with auto-clear

**🔒 Multi-Environment Support**
- Switch between environments with `<ctrl>+e`
- Environment-specific UI colors
- Secure credential management

**🎯 Drill-Down Navigation**
- Process Definition → Process Instances → Variables
- Breadcrumb navigation in footer
- Intuitive back navigation with `Esc`

### Configuration Files

**o8n-env.yaml** (Environment Configuration):
```yaml
environments:
  <env-name>:
    url: <operaton-rest-api-url>
    username: <user>
    password: <password>
    ui_color: <hex-color>  # e.g., "#00A8E1"
active: <default-env-name>
```

**o8n-cfg.yaml** (UI Configuration):
```yaml
tables:
  - name: process-definition
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
  - name: process-variables
    columns:
      - name: name
        visible: true
        width: 30%
        align: left
      - name: value
        visible: true
        width: 70%
        align: left
        editable: true
        input_type: auto
```

### Security Note

⚠️ **Important**: The environment file contains sensitive credentials.

- Add `o8n-env.yaml` to `.gitignore` (already configured)
- Never commit your actual `o8n-env.yaml` to version control
- Use appropriate file permissions: `chmod 600 o8n-env.yaml`
- Consider using environment variables for production deployments

## Development

### Prerequisites
- Go 1.24 or higher
- Docker (for API client generation)

### API Client Generation

Regenerate the Operaton REST API client:

```bash
./.devenv/scripts/generate-api-client.sh
```

### Testing

```bash
go test ./... -v
```

### Project Structure

```
o8n/
├── main.go              # Bubble Tea model, UI logic
├── api.go               # Operaton REST API client wrapper
├── config.go            # Configuration loading/saving
├── o8n-env.yaml         # Environment config (git-ignored)
├── o8n-cfg.yaml         # UI table definitions
├── internal/operaton/   # Generated OpenAPI client
├── resources/           # OpenAPI spec
├── skins/               # Color schemes
└── _bmad/core/prds/     # Design specifications
```

## Documentation

- [specification.md](specification.md) — Complete technical specification
- [Splash Screen Design](_bmad/core/prds/splash-screen-design.md)
- [Compact Header Design](_bmad/core/prds/compact-header-design.md)
- [Layout Design](_bmad/core/prds/layout-design-optimized.md)
- [Modal Confirmation Design](_bmad/core/prds/modal-confirmation-design.md)
- [Help Screen Design](_bmad/core/prds/help-screen-design.md)

## License

See [LICENSE](LICENSE) file.

## Contributing

Contributions welcome! Please read the specification.md for architecture details before submitting PRs.
