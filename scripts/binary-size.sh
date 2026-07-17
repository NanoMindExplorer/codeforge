#!/usr/bin/env bash
# binary-size.sh — Q7.3 track default binary footprint.
# Budget: soft warn + hard fail thresholds (bytes).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# 30 MiB hard (matches CI historically); 26 MiB soft warn for trend watching
HARD_MAX="${BINARY_SIZE_HARD:-31457280}"
SOFT_MAX="${BINARY_SIZE_SOFT:-27262976}"

VER="$(tr -d '[:space:]' < VERSION)"
OUT="${BINARY_SIZE_OUT:-/tmp/codeforge-size-$$}"
TAGS="${BINARY_SIZE_TAGS:-}"

echo "Building size probe (tags=${TAGS:-none})…"
# shellcheck disable=SC2086
CGO_ENABLED=0 go build -tags "${TAGS}" -ldflags="-s -w -X main.ProjectVersion=${VER}" -o "$OUT" ./cmd/codeforge/

SIZE=$(stat -c%s "$OUT" 2>/dev/null || stat -f%z "$OUT")
MIB=$(awk -v s="$SIZE" 'BEGIN{printf "%.2f", s/1048576}')

echo "binary_bytes=$SIZE"
echo "binary_mib=$MIB"
echo "hard_max=$HARD_MAX soft_max=$SOFT_MAX tags=${TAGS:-default}"

# Write report for CI artifact consumers
{
  echo "version=$VER"
  echo "bytes=$SIZE"
  echo "mib=$MIB"
  echo "tags=${TAGS:-default}"
  echo "hard_max=$HARD_MAX"
  echo "soft_max=$SOFT_MAX"
} > binary-size.txt

rm -f "$OUT"

if [[ "$SIZE" -ge "$HARD_MAX" ]]; then
  echo "FAIL: binary ${SIZE} >= hard max ${HARD_MAX} (30MiB)"
  exit 1
fi
if [[ "$SIZE" -ge "$SOFT_MAX" ]]; then
  echo "WARN: binary ${SIZE} >= soft max ${SOFT_MAX} (~26MiB) — watch dependency growth"
fi
echo "OK: size within budget"

# Optional plainmd comparison note
if [[ -z "$TAGS" ]]; then
  echo ""
  echo "Tip: lean build for Termux: BINARY_SIZE_TAGS=plainmd bash scripts/binary-size.sh"
fi
