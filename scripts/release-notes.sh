#!/usr/bin/env bash
# release-notes.sh — print CHANGELOG section + recent commits for a version (R6).
# Usage:
#   bash scripts/release-notes.sh           # current VERSION
#   bash scripts/release-notes.sh 1.8.4
#   bash scripts/release-notes.sh 1.8.4 v1.8.3   # range since previous tag
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VER="${1:-}"
if [[ -z "$VER" ]]; then
  VER="$(tr -d '[:space:]' < VERSION)"
fi
VER="${VER#v}"
PREV="${2:-}"

echo "# CodeForge v${VER}"
echo
if [[ -f CHANGELOG.md ]]; then
  # Extract ## [VER] section until next ## 
  awk -v ver="$VER" '
    $0 ~ "^## \\[" ver "\\]" {p=1; print; next}
    p && /^## \[/ {exit}
    p {print}
  ' CHANGELOG.md
  echo
fi

echo "## Commits"
if [[ -n "$PREV" ]]; then
  git log --oneline "${PREV}..HEAD" 2>/dev/null || git log --oneline -20
else
  # try previous tag
  if prev_tag=$(git describe --tags --abbrev=0 2>/dev/null); then
    git log --oneline "${prev_tag}..HEAD" 2>/dev/null || git log --oneline -20
  else
    git log --oneline -20
  fi
fi
echo
echo "## Verify install"
echo '```bash'
echo "codeforge version   # → codeforge ${VER}"
echo '```'
