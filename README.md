# musketeer-bridge

Local daemon for Musketeer governed execution. Exposes a static tool registry and deterministic CLI execution with allowlisted roots, strict JSON mode, and on-disk run logs.

Musketeer-bridge sits at the runtime layer. It does not govern the work cycle - [Musketeer](https://musketeer.dev) does that. The bridge executes bounded tool invocations on behalf of the harness, with full audit trails.

## Quick start (from clean clone)

### 1. Install

Build from source (requires Go 1.21+):

```sh
git clone https://github.com/justyn-clark/musketeer-bridge
cd musketeer-bridge
go build -o target/musketeer-bridge ./cmd/musketeer-bridge
```

### 2. Configure

Copy the example config:

```sh
cp bridge.json.example ~/.musketeer/bridge.json
```

Edit `~/.musketeer/bridge.json` and set `allowlisted_roots` to directories the bridge is allowed to execute tools in:

```json
{
  "listen_addr": "127.0.0.1:18789",
  "allowlisted_roots": ["/Users/yourname/Projects"],
  "env_allowlist": ["PATH", "HOME", "USER", "SHELL", "TERM"],
  "max_runtime_ms": 600000,
  "registry_dir": "~/.musketeer/registry",
  "runs_dir": "~/.musketeer/runs"
}
```

If `allowlisted_roots` is empty, all run requests are rejected with `ERR_CWD_NOT_ALLOWLISTED`.

### 3. Add tool specs to the registry

```sh
mkdir -p ~/.musketeer/registry
cp -R registry-examples/tools ~/.musketeer/registry/
```

The registry layout is:

```
~/.musketeer/registry/tools/<name>/<version>/tool.json
```

### 4. Start the server

```sh
./target/musketeer-bridge serve
```

Expected output:
```
listening on 127.0.0.1:18789
```

The bridge binds on `127.0.0.1:18789` by default. Override with `MUSKETEER_BRIDGE_LISTEN_ADDR`.

### 5. Verify health

```sh
curl -s http://127.0.0.1:18789/v1/health
```

Expected response:
```json
{"exit_code":0,"ok":true}
```

### 6. List tools

```sh
curl -s http://127.0.0.1:18789/v1/tools
```

Expected response (with example registry):
```json
{"exit_code":0,"tools":["loopexec","musketeer"]}
```

### 7. Sample run request

```sh
curl -s -X POST http://127.0.0.1:18789/v1/tools/loopexec/run \
  -H 'content-type: application/json' \
  -d '{
    "version": "0.1.0",
    "args": {},
    "cwd": "/Users/yourname/Projects/myproject",
    "env": {},
    "mode": "json",
    "client": {"name": "manual"}
  }'
```

The `cwd` must be inside an `allowlisted_roots` directory. A successful response includes `exit_code: 0` and `stdout_json` when the tool outputs valid JSON.

Error response shape:
```json
{
  "exit_code": 40,
  "ok": false,
  "error": {
    "code": "ERR_CWD_NOT_ALLOWLISTED",
    "message": "cwd is not within an allowlisted root"
  }
}
```

## Config

Config file: `~/.musketeer/bridge.json`

All fields have defaults. The file is optional - the bridge starts with safe defaults if it is absent. If the file exists but contains invalid JSON, the bridge exits with a structured error (`ERR_CONFIG_INVALID`).

| Field | Default | Description |
|---|---|---|
| `listen_addr` | `127.0.0.1:18789` | TCP address to bind |
| `allowlisted_roots` | `[]` | Directories tools are allowed to run in. Empty = reject all. |
| `env_allowlist` | `["PATH","HOME","USER","SHELL","TERM"]` | Env vars passed to tool processes |
| `max_runtime_ms` | `600000` | Execution timeout in milliseconds (10 min) |
| `registry_dir` | `~/.musketeer/registry` | Tool spec directory |
| `runs_dir` | `~/.musketeer/runs` | Run log storage directory |

Environment overrides:
- `MUSKETEER_BRIDGE_LISTEN_ADDR`
- `MUSKETEER_BRIDGE_REGISTRY_DIR`
- `MUSKETEER_BRIDGE_RUNS_DIR`

## Operational boundaries

- **Timeout**: Every tool execution is bounded by `max_runtime_ms` using a context deadline. Exceeded → `ERR_TIMEOUT`, exit code 124.
- **Allowlist**: `cwd` in the run request must be under an `allowlisted_roots` entry. Symlinks are resolved before comparison. Rejected → `ERR_CWD_NOT_ALLOWLISTED`, exit code 40.
- **Env filtering**: Only keys in `env_allowlist` are passed to tool processes. Request env keys not in the allowlist are silently dropped.
- **No shell**: Tools are executed directly via argv. No shell interpolation.
- **Stdout size**: No hard limit. Stdout is captured in memory; keep tool output bounded.
- **Strict JSON mode**: When `json_mode: true` and request `mode: "json"`, stdout must be exactly one JSON object (not array, not multiple values). Violations → `ERR_STDOUT_NOT_JSON`, exit code 40.

## Endpoints

- `GET /v1/health` - Liveness check. Returns `{"ok": true, "exit_code": 0}`.
- `GET /v1/tools` - List registered tools.
- `GET /v1/tools/{name}` - Get tool spec.
- `POST /v1/tools/{name}/run` - Execute tool.

All responses are JSON and include `exit_code`.

## Structured error codes

| Code | Meaning | HTTP status |
|---|---|---|
| `ERR_INVALID_INPUT` | Request JSON invalid or missing required fields | 400 |
| `ERR_TOOL_NOT_FOUND` | Tool name not in registry | 404 |
| `ERR_CWD_NOT_ALLOWLISTED` | cwd outside allowlisted roots | 400 |
| `ERR_TIMEOUT` | Tool exceeded max_runtime_ms | 400 |
| `ERR_STDOUT_NOT_JSON` | Tool stdout not a single JSON object (json_mode only) | 400 |
| `ERR_EXEC_FAILED` | Tool process failed to start | 500 |
| `ERR_CONFIG_INVALID` | bridge.json exists but is not valid JSON | (startup fatal) |
| `ERR_REGISTRY_INVALID` | Registry tool.json missing required fields | (startup fatal) |

## Registry layout

```
~/.musketeer/registry/tools/<name>/<version>/tool.json
```

`tool.json` required fields:
- `name`
- `version`
- `description`
- `json_mode` (bool)
- `exec.argv` (non-empty)

The latest version is selected by lexicographic sort of version directory names.

## Run logs

Every `POST /run` writes a run directory, including validation rejections:

```
~/.musketeer/runs/YYYY/MM/DD/<run_id>/
  request.json    - original request
  resolved.json   - tool spec used
  stdout.json     - parsed JSON stdout (only when json_mode && stdout is valid JSON)
  stderr.txt      - raw stderr
  result.json     - final result including exit_code and error if any
```

## Security model

- No shell execution; argv only
- `cwd` must be inside allowlisted roots (symlink-safe comparison)
- Environment passed through allowlist only
- stdout and stderr captured separately
- When `json_mode: true` and `mode: "json"`, stdout must be exactly one JSON object

## Contract tests

Contract tests lock three invariants:
- Strict JSON parsing behavior for json_mode tools
- `exit_code` always present in all responses
- Run logs always written for `POST /run` (including failures and rejections)

Run:
```sh
go test ./...
```

## Registry examples

Versioned examples:
- `loopexec 0.1.1` - strict JSON mode
- `musketeer 0.1.1` - strict JSON mode, `musketeer init --json`

## TODO (not implemented)

- Optional streaming stderr endpoint or SSE
- MCP adapter layer (discovery and call forwarding)
- Artifact detection and SHA256 hashing
- Semver parsing for version selection (currently lexicographic)
- Optional auth token even on localhost
