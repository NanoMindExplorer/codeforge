#!/usr/bin/env bash
# release-gate.sh — automated v1.9.0 readiness checks (W4).
# Exit 0 only when machine-checkable gates pass.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VER="$(tr -d '[:space:]' < VERSION)"
fail=0
pass() { echo "✓ $*"; }
warn() { echo "⚠ $*"; }
bad()  { echo "✗ $*"; fail=1; }

echo "═══════════════════════════════════════"
echo " CodeForge release gate  v${VER}"
echo "═══════════════════════════════════════"
echo

# 1. Version consistency
if bash scripts/check-version.sh; then
  pass "check-version"
else
  bad "check-version"
fi

# 1b. gofmt
if bash scripts/gofmt-check.sh; then
  pass "gofmt"
else
  bad "gofmt (run: make fmt)"
fi

# 2. Tests + coverage floor (Q0.2)
if bash scripts/coverage-check.sh >/tmp/cf-gate-cover.log 2>&1; then
  pass "go test + coverage floor"
  tail -3 /tmp/cf-gate-cover.log || true
else
  bad "coverage (see /tmp/cf-gate-cover.log)"
  tail -20 /tmp/cf-gate-cover.log || true
fi

if GOSUMDB=off go vet ./... >/tmp/cf-gate-vet.log 2>&1; then
  pass "go vet"
else
  bad "go vet"
fi

# 2b. Race on critical packages (Q0.1) — soft-fail only if CGO/race unavailable
if bash scripts/test-race.sh >/tmp/cf-gate-race.log 2>&1; then
  pass "go test -race (critical packages)"
else
  bad "race tests (see /tmp/cf-gate-race.log)"
  tail -30 /tmp/cf-gate-race.log || true
fi

CGO_ENABLED=0 go build -ldflags="-s -w -X main.ProjectVersion=${VER}" -o /tmp/codeforge-gate ./cmd/codeforge/
out=$(/tmp/codeforge-gate version 2>&1 || true)
if echo "$out" | grep -q "$VER"; then
  pass "binary version: $out"
else
  bad "binary version mismatch: $out (want $VER)"
fi

SIZE=$(stat -c%s /tmp/codeforge-gate 2>/dev/null || stat -f%z /tmp/codeforge-gate)
if [[ "$SIZE" -lt 31457280 ]]; then
  pass "binary size ${SIZE} < 30MiB"
else
  bad "binary too large: ${SIZE}"
fi

# 3. Packaging artifacts
[[ -f install.sh ]] && pass "install.sh" || bad "install.sh missing"
[[ -f Formula/codeforge.rb ]] && pass "Formula/codeforge.rb" || bad "Formula missing"
[[ -f contrib/termux/build.sh ]] && pass "contrib/termux/build.sh" || bad "termux build.sh"
[[ -f .github/workflows/release.yml ]] && pass "release.yml" || bad "release.yml"
[[ -f CHANGELOG.md ]] && grep -q "\[${VER}\]" CHANGELOG.md && pass "CHANGELOG has [${VER}]" || bad "CHANGELOG missing [${VER}]"

# 4. Dogfood / gate docs present
for f in docs/DOGFOOD.md docs/dogfood/SCORECARD.md docs/dogfood/BATCH_F.md docs/RELEASE_GATE.md docs/AUDIT_AND_ROADMAP.md; do
  if [[ -f "$f" ]]; then
    pass "$f"
  else
    bad "missing $f"
  fi
done

# 4b. Offline dogfood suite (Q0.3)
if DOGFOOD_LIVE=0 bash scripts/dogfood-run.sh >/tmp/cf-gate-dogfood.log 2>&1; then
  pass "dogfood offline (DOGFOOD_LIVE=0)"
else
  bad "dogfood offline (see /tmp/cf-gate-dogfood.log)"
  tail -30 /tmp/cf-gate-dogfood.log || true
fi

# 5. Headless no_provider contract (no keys)
(
  unset XAI_API_KEY GROK_API_KEY GEMINI_API_KEY ANTHROPIC_API_KEY OPENAI_API_KEY || true
  export HOME="${TMPDIR:-/tmp}/cf-gate-home-$$"
  mkdir -p "$HOME"
  set +e
  out=$(/tmp/codeforge-gate agent --json "hi" 2>&1)
  ec=$?
  set -e
  if [[ "$ec" -eq 2 ]] && echo "$out" | grep -q 'no_provider'; then
    pass "headless no_provider exit=2"
  else
    # may still pass if local config has keys
    if echo "$out" | grep -q '"ok"'; then
      warn "headless returned structured JSON (exit=$ec) — keys may be in env/config"
    else
      bad "headless no_provider (exit=$ec): $(echo "$out" | head -c 200)"
    fi
  fi
)

# 6. Terminal matrix smoke (no TUI)
if bash scripts/smoke-matrix.sh; then
  pass "smoke-matrix"
else
  bad "smoke-matrix"
fi

echo
echo "───────────────────────────────────────"
if [[ "$fail" -ne 0 ]]; then
  echo "GATE FAILED for v${VER}"
  exit 1
fi
echo "GATE PASSED (automated) for v${VER}"
echo
echo "Human gates still required (see docs/RELEASE_GATE.md):"
echo "  · Dogfood A–C 100% / D–E ≥80% / F matrix on real terminals"
echo "  · Tag v${VER} → GitHub Release (when ready to publish)"
echo "  · First-run UX spot-check"
exit 0
