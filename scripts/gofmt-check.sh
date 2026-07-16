#!/usr/bin/env bash
# gofmt-check.sh — fail if any .go file needs gofmt (CI / pre-commit).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

out="$(gofmt -l . 2>/dev/null || true)"
if [[ -z "$out" ]]; then
  echo "OK: gofmt clean"
  exit 0
fi

echo "FAIL: the following files need gofmt:"
echo "$out"
echo
echo "Fix with:  make fmt   # or: gofmt -w ."
echo "Hook:      make install-hooks"
exit 1
