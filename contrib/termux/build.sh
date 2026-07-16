#!/usr/bin/env bash
# Termux package build skeleton for termux-packages (R4).
#
# Usage (standalone, on device or CI arm64):
#   bash contrib/termux/build.sh
#   bash contrib/termux/build.sh /path/to/install/prefix
#
# When vendored into termux-packages, copy this file to:
#   packages/codeforge/build.sh
# and set TERMUX_PKG_* vars from the generated package metadata below.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

VER="$(tr -d '[:space:]' < VERSION 2>/dev/null || echo 0.0.0)"
DEST="${1:-}"
if [[ -z "$DEST" ]]; then
  if [[ -n "${PREFIX:-}" && -d "${PREFIX}/bin" ]]; then
    DEST="${PREFIX}/bin"
  elif [[ -n "${TERMUX_PREFIX:-}" ]]; then
    DEST="${TERMUX_PREFIX}/bin"
  else
    DEST="${ROOT}"
  fi
fi

echo "Building codeforge v${VER} for Termux → ${DEST}"

# termux-packages hooks (no-op when run standalone)
if declare -f termux_step_pre_configure >/dev/null 2>&1; then
  :
fi

export CGO_ENABLED=0
go build -trimpath -ldflags="-s -w -X main.ProjectVersion=${VER}" \
  -o "${ROOT}/codeforge" ./cmd/codeforge/

mkdir -p "$DEST"
install -m 755 "${ROOT}/codeforge" "${DEST}/codeforge"
echo "✓ Installed ${DEST}/codeforge"
"${DEST}/codeforge" version || true
