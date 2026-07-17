#!/usr/bin/env bash
# govulncheck.sh — dependency vulnerability scan (Q0.5).
# Default WARN mode: print findings but exit 0 (harden later with STRICT=1).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

STRICT="${GOVULNCHECK_STRICT:-0}"

if ! command -v govulncheck >/dev/null 2>&1; then
  echo "Installing govulncheck..."
  GOSUMDB=off go install golang.org/x/vuln/cmd/govulncheck@latest
  export PATH="$(go env GOPATH)/bin:$PATH"
fi

echo "Running govulncheck ./..."
set +e
out="$(GOSUMDB=off govulncheck ./... 2>&1)"
ec=$?
set -e
echo "$out"

if [[ "$ec" -eq 0 ]]; then
  echo "OK: govulncheck clean"
  exit 0
fi

if [[ "$STRICT" == "1" ]]; then
  echo "FAIL: govulncheck found issues (STRICT=1)"
  exit "$ec"
fi

echo "WARN: govulncheck reported issues (non-blocking; set GOVULNCHECK_STRICT=1 to fail)"
exit 0
