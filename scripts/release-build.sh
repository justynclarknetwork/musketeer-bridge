#!/usr/bin/env bash
set -euo pipefail

TAG="$(git describe --tags --exact-match 2>/dev/null || true)"
if [ -z "$TAG" ]; then
  echo "ERROR: must run at an exact tag" >&2
  exit 1
fi
if [ "$TAG" != "v0.1.1" ]; then
  echo "ERROR: expected tag v0.1.1, got $TAG" >&2
  exit 1
fi
VERSION="${TAG#v}"
DIST="dist"
rm -rf "$DIST"
mkdir -p "$DIST"

build_one() {
  local goos="$1" goarch="$2" ext="$3"
  local bin="musketeer-bridge${ext}"
  local outdir
  outdir="$(mktemp -d)"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -o "$outdir/$bin" ./cmd/musketeer-bridge
  cp README.md "$outdir/README.md"
  local name="musketeer-bridge_v${VERSION}_${goos}_${goarch}"
  if [ "$goos" = "windows" ]; then
    (cd "$outdir" && zip -q "$OLDPWD/$DIST/${name}.zip" "$bin" README.md)
  else
    (cd "$outdir" && tar -czf "$OLDPWD/$DIST/${name}.tar.gz" "$bin" README.md)
  fi
  rm -rf "$outdir"
}

build_one darwin arm64 ""
build_one darwin amd64 ""
build_one linux amd64 ""
build_one windows amd64 ".exe"

(
  cd "$DIST"
  shasum -a 256 "musketeer-bridge_v${VERSION}_darwin_arm64.tar.gz" \
    "musketeer-bridge_v${VERSION}_darwin_amd64.tar.gz" \
    "musketeer-bridge_v${VERSION}_linux_amd64.tar.gz" \
    "musketeer-bridge_v${VERSION}_windows_amd64.zip" \
    > "checksums_v${VERSION}.txt"
)

echo "Built artifacts in $DIST"
ls -1 "$DIST"
