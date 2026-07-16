#!/usr/bin/env bash
# Emit termux-packages metadata with PKG_VERSION from VERSION file (R4).
# Usage: bash contrib/termux/package.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
VER="$(tr -d '[:space:]' < "$ROOT/VERSION")"

cat <<EOF
# Auto-generated from CodeForge VERSION=${VER}
# Copy into termux-packages/packages/codeforge/

TERMUX_PKG_HOMEPAGE=https://github.com/NanoMindExplorer/codeforge
TERMUX_PKG_DESCRIPTION="Terminal AI coding companion (Grok-compatible TUI agent)"
TERMUX_PKG_LICENSE="Apache-2.0"
TERMUX_PKG_MAINTAINER="NanoMind"
TERMUX_PKG_VERSION=${VER}
TERMUX_PKG_SRCURL=https://github.com/NanoMindExplorer/codeforge/archive/refs/tags/v\${TERMUX_PKG_VERSION}.tar.gz
TERMUX_PKG_SHA256=SKIP_ME_UNTIL_RELEASE
TERMUX_PKG_BUILD_IN_SRC=true
TERMUX_PKG_DEPENDS="golang"
TERMUX_PKG_BUILD_DEPENDS="golang"

termux_step_make() {
  CGO_ENABLED=0 go build -trimpath \\
    -ldflags="-s -w -X main.ProjectVersion=\${TERMUX_PKG_VERSION}" \\
    -o codeforge ./cmd/codeforge/
}

termux_step_make_install() {
  install -Dm700 codeforge "\${TERMUX_PREFIX}/bin/codeforge"
}
EOF
