#!/usr/bin/env bash
# Point this repo at scripts/githooks (gofmt pre-commit).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

chmod +x scripts/githooks/pre-commit scripts/gofmt-check.sh 2>/dev/null || true

git config core.hooksPath scripts/githooks
echo "✓ git core.hooksPath = scripts/githooks"
echo "  pre-commit will gofmt staged *.go files"
echo "  disable: git config --unset core.hooksPath"
