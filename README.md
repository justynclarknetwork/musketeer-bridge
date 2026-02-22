# musketeer-bridge

Local localhost daemon that exposes a static tool registry and deterministic CLI execution with allowlisted cwd, strict JSON mode, and on-disk run logs.

## Config

Config file: `~/.musketeer/bridge.json`

Example:

```json
{
  "listen_addr": "127.0.0.1:18789",
  "allowlisted_roots": ["/Users/justin/Projects"],
  "env_allowlist": ["PATH", "HOME", "USER", "SHELL", "TERM"],
  "max_runtime_ms": 600000,
  "registry_dir": "~/.musketeer/registry",
  "runs_dir": "~/.musketeer/runs"
}
```

If `allowlisted_roots` is empty, all run requests are rejected (`ERR_CWD_NOT_ALLOWLISTED`).

Env overrides:
- `MUSKETEER_BRIDGE_LISTEN_ADDR`
- `MUSKETEER_BRIDGE_REGISTRY_DIR`
- `MUSKETEER_BRIDGE_RUNS_DIR`

## Registry layout

`~/.musketeer/registry/tools/<name>/<version>/tool.json`

`tool.json` required fields:
- `name`
- `version`
- `description`
- `json_mode`
- `exec.argv`

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
- `stdout.json` (only when parsed JSON exists)
- `stderr.txt`
- `result.json`

Logs are written for every `POST /run`, including validation rejections.

## Security model

- no shell execution; argv only
- cwd must be inside allowlisted roots
- env passed through allowlist only
- stdout and stderr captured separately
- when `json_mode=true` and `mode=json`, stdout must be exactly one JSON object

## Acceptance commands

1. `cd ./musketeer-bridge && go build -o target/musketeer-bridge ./cmd/musketeer-bridge`
2. `mkdir -p ~/.musketeer/registry && cp -R registry-examples/tools ~/.musketeer/registry/`
3. `MUSKETEER_BRIDGE_LISTEN_ADDR=127.0.0.1:18789 ./target/musketeer-bridge`
4. `curl -s http://127.0.0.1:18789/v1/health`
5. `curl -s http://127.0.0.1:18789/v1/tools`
6. `curl -s http://127.0.0.1:18789/v1/tools/loopexec`
7. `curl -s -X POST http://127.0.0.1:18789/v1/tools/loopexec/run -H 'content-type: application/json' -d '{"version":"0.1.0","args":{},"cwd":".","env":{},"mode":"json","client":{"name":"manual"}}'`

## TODO (not executed)

- add optional streaming stderr endpoint or SSE
- add MCP adapter layer (discovery and call forwarding)
- add artifact detection and sha256 hashing
- add semver parsing for version selection
- add optional auth token even on localhost

## Contract tests

Contract tests lock three invariants:
- strict JSON parsing behavior for json_mode tools
- exit_code always present in responses
- run logs always written for POST /run (including failures)

Run:
- `go test ./...`
- `go test ./... -run Contract -count=1`


## Registry examples

Versioned examples include:
- loopexec 0.1.1 strict JSON mode example
- musketeer 0.1.1 strict JSON mode example

Placeholder versions (0.1.0) are retained for backward reference.
