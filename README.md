# musketeer-bridge

Local localhost daemon that exposes a static tool registry and executes allowlisted CLI tools with JSON-only HTTP responses.

## Config

Config file: `~/.musketeer/bridge.json`

Example:

```json
{
  "listen_addr": "127.0.0.1:18789",
  "allowlisted_roots": ["/Users/justin/Projects"],
  "env_allowlist": ["PATH","HOME","USER","SHELL","TERM"],
  "max_runtime_ms": 600000,
  "registry_dir": "~/.musketeer/registry",
  "runs_dir": "~/.musketeer/runs"
}
```

Env overrides:
- `MUSKETEER_BRIDGE_LISTEN_ADDR`
- `MUSKETEER_BRIDGE_REGISTRY_DIR`
- `MUSKETEER_BRIDGE_RUNS_DIR`

## Registry layout

`~/.musketeer/registry/tools/<name>/<version>/tool.json`

`tool.json` fields:
- `name`
- `version`
- `description`
- `json_mode`
- `exec.argv`
- `exec.args_mapping` (optional)

## Endpoints

- `GET /v1/health`
- `GET /v1/tools`
- `GET /v1/tools/{name}`
- `POST /v1/tools/{name}/run`

All responses are JSON and include `exit_code`.

## Run logs

`~/.musketeer/runs/YYYY/MM/DD/<run_id>/`
- `request.json`
- `resolved.json`
- `stdout.json` (when parsed JSON exists)
- `stderr.txt`
- `result.json`

## Security model

- no shell execution; argv only
- cwd must be inside allowlisted roots
- env passed through allowlist only
- stdout and stderr captured separately
- json-mode stdout must be exactly one JSON object
