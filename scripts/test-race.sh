#!/usr/bin/env bash
# test-race.sh — race detector on critical packages (Q0.1).
# Full ./... race is slow; focus on concurrency-sensitive paths first.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PKGS=(
  ./internal/agent/...
  ./internal/tool/...
  ./internal/session/...
  ./internal/acp/...
  ./internal/provider/...
  ./internal/permission/...
  ./internal/bgtask/...
)

echo "go test -race on critical packages:"
printf '  %s\n' "${PKGS[@]}"

# Race detector needs cgo on some platforms; enable when available.
export CGO_ENABLED="${CGO_ENABLED:-1}"

# Probe: some GOOS/GOARCH (e.g. android/arm64) do not support -race.
probe_log="$(mktemp)"
if ! GOSUMDB=off go test -race -c -o /tmp/cf-race-probe ./internal/bgtask >"$probe_log" 2>&1; then
  if grep -qiE 'not supported|race is not supported' "$probe_log"; then
    echo "SKIP: -race not supported on $(go env GOOS)/$(go env GOARCH) (CI runs this on ubuntu-latest)"
    cat "$probe_log" | head -5 || true
    rm -f "$probe_log" /tmp/cf-race-probe
    exit 0
  fi
  echo "WARN: race probe compile issue; continuing full race run"
  cat "$probe_log" | tail -20 || true
fi
rm -f "$probe_log" /tmp/cf-race-probe

GOSUMDB=off go test -race -count=1 "${PKGS[@]}"
echo "OK: race tests passed"
