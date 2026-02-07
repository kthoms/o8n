# o8n
A terminal UI for Operaton

## Configuration

There are two config files now:

- `o8n-env.yaml` (environment credentials and ui colors) — keep this file secret
- `o8n-cfg.yaml` (UI table definitions and app-level settings)

Create the example env file and edit it:

```bash
cp o8n-env.yaml.example o8n-env.yaml
# edit o8n-env.yaml to add your environments
```

Create or edit `o8n-cfg.yaml` for table definitions (an example is included).

### Security Note

⚠️ **Important**: The environment file contains sensitive credentials.

- Add `o8n-env.yaml` to `.gitignore` (already configured)
- Never commit your actual `o8n-env.yaml` to version control
- Consider using environment variables for sensitive data in production
- Use appropriate file permissions (e.g., `chmod 600 o8n-env.yaml`)

## Building

```bash
go build -o o8n .
```

## Running

```bash
./o8n
```
