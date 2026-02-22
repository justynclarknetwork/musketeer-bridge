#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT=18889
export MUSKETEER_BRIDGE_LISTEN_ADDR="127.0.0.1:${PORT}"
export MUSKETEER_BRIDGE_REGISTRY_DIR="$HOME/.musketeer/registry"
export MUSKETEER_BRIDGE_RUNS_DIR="$HOME/.musketeer/runs"
mkdir -p "$HOME/.musketeer/registry/tools" "$HOME/.musketeer"
cat > "$HOME/.musketeer/bridge.json" <<JSON
{
  "listen_addr": "127.0.0.1:${PORT}",
  "allowlisted_roots": ["/tmp"],
  "env_allowlist": ["PATH","HOME","USER","SHELL","TERM"],
  "max_runtime_ms": 600000,
  "registry_dir": "${HOME}/.musketeer/registry",
  "runs_dir": "${HOME}/.musketeer/runs"
}
JSON
rm -rf "$HOME/.musketeer/registry/tools/loopexec" "$HOME/.musketeer/registry/tools/musketeer"

"$ROOT/target/musketeer-bridge" > /tmp/mbridge.log 2>&1 &
PID=$!
trap 'kill $PID >/dev/null 2>&1 || true' EXIT
sleep 1

curl -s "http://127.0.0.1:${PORT}/v1/health" | jq -e '.ok==true and .exit_code==0' >/dev/null
printf 'PASS health\n'

curl -s "http://127.0.0.1:${PORT}/v1/tools" | jq -e '.tools|type=="array"' >/dev/null
printf 'PASS tools empty\n'

mkdir -p \
  "$HOME/.musketeer/registry/tools/loopexec/0.1.0" \
  "$HOME/.musketeer/registry/tools/loopexec/0.1.1" \
  "$HOME/.musketeer/registry/tools/musketeer/0.1.0" \
  "$HOME/.musketeer/registry/tools/musketeer/0.1.1"
cp "$ROOT/registry-examples/tools/loopexec/0.1.0/tool.json" "$HOME/.musketeer/registry/tools/loopexec/0.1.0/tool.json"
cp "$ROOT/registry-examples/tools/loopexec/0.1.1/tool.json" "$HOME/.musketeer/registry/tools/loopexec/0.1.1/tool.json"
cp "$ROOT/registry-examples/tools/musketeer/0.1.0/tool.json" "$HOME/.musketeer/registry/tools/musketeer/0.1.0/tool.json"
cp "$ROOT/registry-examples/tools/musketeer/0.1.1/tool.json" "$HOME/.musketeer/registry/tools/musketeer/0.1.1/tool.json"

kill $PID || true
sleep 1
"$ROOT/target/musketeer-bridge" > /tmp/mbridge.log 2>&1 &
PID=$!
sleep 1

curl -s "http://127.0.0.1:${PORT}/v1/tools" | jq -e '(.tools|index("loopexec"))!=null and (.tools|index("musketeer"))!=null' >/dev/null
printf 'PASS tools loaded\n'

# loopexec strict json check
resp=$(curl -s -X POST "http://127.0.0.1:${PORT}/v1/tools/loopexec/run" -H 'content-type: application/json' -d '{"version":"0.1.1","args":{},"cwd":"/tmp","env":{},"mode":"json","client":{"name":"smoke"}}')
echo "$resp" | jq -e '.exit_code==0 and (.stdout_json|type=="object") and (.stderr=="")' >/dev/null
printf 'PASS loopexec json_mode assertions\n'

# musketeer strict json check (success if installed, deterministic error if not)
mresp=$(curl -s -X POST "http://127.0.0.1:${PORT}/v1/tools/musketeer/run" -H 'content-type: application/json' -d '{"version":"0.1.1","args":{},"cwd":"/tmp","env":{},"mode":"json","client":{"name":"smoke"}}')
echo "$mresp" | jq -e '.exit_code!=null' >/dev/null
mcode=$(echo "$mresp" | jq -r '.exit_code')
if [ "$mcode" = "0" ]; then
  echo "$mresp" | jq -e '(.stdout_json|type=="object") and (.stderr=="")' >/dev/null
  printf 'PASS musketeer json_mode success assertions\n'
else
  echo "$mresp" | jq -e '.error.code=="ERR_EXEC_FAILED"' >/dev/null
  printf 'PASS musketeer deterministic error envelope assertions\n'
fi

run_dir=$(find "$HOME/.musketeer/runs" -type d | sort | tail -n 1)
test -f "$run_dir/result.json"
printf 'PASS run artifact result.json exists\n'
printf 'PASS smoke\n'
