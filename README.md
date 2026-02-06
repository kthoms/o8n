# o8n
A terminal UI for Operaton

## Configuration

Create a `config.yaml` file based on `config.yaml.example`:

```bash
cp config.yaml.example config.yaml
```

Edit the file to add your Operaton environment(s):

```yaml
environments:
  local:
    url: "http://localhost:8080/engine-rest"
    username: "your-username"
    password: "your-password"
    ui_color: "#00A8E1"
```

### Security Note

⚠️ **Important**: The configuration file contains sensitive credentials. 

- Add `config.yaml` to `.gitignore` (already configured)
- Never commit your actual `config.yaml` to version control
- Consider using environment variables for sensitive data in production
- Use appropriate file permissions (e.g., `chmod 600 config.yaml`)

## Building

```bash
go build -o o8n .
```

## Running

```bash
./o8n
```
