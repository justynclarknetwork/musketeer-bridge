# PR Checklist - musketeer-bridge v0.1

## Scope
- local localhost daemon
- static tool registry
- allowlisted cwd execution only
- strict JSON handling
- run logs on disk

## Acceptance checklist
- [x] `go test ./...`
- [x] `go build -o target/musketeer-bridge ./cmd/musketeer-bridge`
- [x] `bash scripts/smoke.sh`
- [x] JSON-only responses
- [x] `exit_code` present in every response
- [x] allowlist rejection returns `ERR_CWD_NOT_ALLOWLISTED`
- [x] run logs written under `~/.musketeer/runs/...`

## Notes
- `json_mode` in example tool specs is set to `false` until real single-object JSON flags are wired.
- Empty allowlist safely rejects all runs.
