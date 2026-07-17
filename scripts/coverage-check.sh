#!/usr/bin/env bash
# coverage-check.sh — run tests with coverprofile and enforce minimum total %.
# Floor is scripts/coverage-floor.txt (default 33). Fail if coverage drops below floor.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FLOOR_FILE="${COVERAGE_FLOOR_FILE:-scripts/coverage-floor.txt}"
FLOOR="${COVERAGE_FLOOR:-}"
if [[ -z "$FLOOR" && -f "$FLOOR_FILE" ]]; then
  FLOOR="$(tr -d '[:space:]%' < "$FLOOR_FILE")"
fi
if [[ -z "$FLOOR" ]]; then
  FLOOR=33
fi

OUT="${COVERPROFILE:-coverage.out}"
MODE="${COVERMODE:-atomic}"

echo "Running go test with -coverprofile=$OUT (floor=${FLOOR}%)"
GOSUMDB=off go test ./... -count=1 -covermode="$MODE" -coverprofile="$OUT"

if [[ ! -f "$OUT" ]]; then
  echo "ERROR: coverprofile not created: $OUT"
  exit 1
fi

# total: (statements) XX.X%
total_line="$(go tool cover -func="$OUT" | grep '^total:' || true)"
if [[ -z "$total_line" ]]; then
  echo "ERROR: could not parse total coverage from go tool cover"
  go tool cover -func="$OUT" | tail -5
  exit 1
fi

# Extract last field like 33.7%
pct="$(echo "$total_line" | awk '{print $NF}' | tr -d '%')"
echo "Coverage total: ${pct}%  (floor ${FLOOR}%)"
echo "$total_line"

# Compare with awk (float)
awk -v pct="$pct" -v floor="$FLOOR" 'BEGIN {
  if (pct+0 < floor+0) {
    printf "FAIL: coverage %.1f%% is below floor %s%%\n", pct, floor > "/dev/stderr"
    exit 1
  }
  printf "OK: coverage %.1f%% >= floor %s%%\n", pct, floor
}'

# Optional: write summary for CI
SUMMARY="${COVER_SUMMARY:-coverage-summary.txt}"
{
  echo "total_percent=$pct"
  echo "floor_percent=$FLOOR"
  echo "profile=$OUT"
  echo "line=$total_line"
} > "$SUMMARY"
echo "Wrote $SUMMARY"
exit 0
