# Release Notes - v0.1.0

## What this release guarantees

- `musketeer-bridge serve` starts an HTTP daemon on `127.0.0.1:18789` (configurable).
- `GET /v1/health` always returns `{"ok": true, "exit_code": 0}`.
- `GET /v1/tools` lists all tools loaded from the registry.
- `GET /v1/tools/{name}` returns the tool spec or `ERR_TOOL_NOT_FOUND`.
- `POST /v1/tools/{name}/run` executes the tool if `cwd` is within `allowlisted_roots`.
- All responses are JSON and always include `exit_code`.
- Tool processes are launched via argv directly. No shell interpolation.
- Every `POST /run` writes a run log to disk, including validation rejections.
- `cwd` in the run request is validated against `allowlisted_roots` with symlink resolution.
- Env vars passed to tool processes are filtered to `env_allowlist` only.
- When `json_mode: true` and `mode: "json"`, stdout is parsed as a single JSON object. Multiple values, arrays, and non-JSON output are rejected.
- Execution timeout is enforced via `max_runtime_ms` (default 10 minutes). Exceeded processes return `ERR_TIMEOUT` with exit code 124.
- Config is loaded from `~/.musketeer/bridge.json` with safe defaults. Invalid JSON in the config file causes a structured startup failure (`ERR_CONFIG_INVALID`).
- `musketeer-bridge --help` or `help` exits 0 without binding the listen address.

## Structured error codes

| Code | Trigger |
|---|---|
| `ERR_INVALID_INPUT` | Request body is not valid JSON |
| `ERR_TOOL_NOT_FOUND` | Tool name not in registry |
| `ERR_CWD_NOT_ALLOWLISTED` | cwd outside allowlisted roots |
| `ERR_TIMEOUT` | Tool exceeded max_runtime_ms |
| `ERR_STDOUT_NOT_JSON` | json_mode tool produced non-object or multi-value stdout |
| `ERR_EXEC_FAILED` | Tool process failed to start |
| `ERR_CONFIG_INVALID` | bridge.json has invalid JSON (startup fatal) |
| `ERR_REGISTRY_INVALID` | tool.json missing required fields (startup fatal) |

## What is still intentionally limited

- No output size limits on tool stdout or stderr.
- No request body size limits.
- No authentication. The daemon binds to localhost only; no external auth token is enforced.
- Registry version selection is lexicographic, not semver.
- No SSE or streaming stderr endpoint.
- No MCP adapter layer.
- No artifact detection or SHA256 hashing of tool outputs.

## Test coverage

- Contract tests: `internal/contracttest/contract_test.go` (7 tests: good JSON, bad JSON variants, timeout, allowlist rejection, help command)
- Unit tests: `internal/runner/runner_test.go` (allowlist logic, JSON parsing)
- Config tests: `internal/config/config_test.go` (defaults, env overrides)
- HTTP API tests: `internal/httpapi/httpapi_test.go` (health check, invalid request, tool not found, exit_code contract)

Run: `go test ./...`

## Verified behavior

All tests pass:

```
go test ./...
ok  musketeer-bridge/internal/config
ok  musketeer-bridge/internal/contracttest
ok  musketeer-bridge/internal/httpapi
ok  musketeer-bridge/internal/runner
```
