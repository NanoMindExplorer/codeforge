# Termux package notes

## Quick install (recommended)

```bash
pkg install -y golang git
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | sh
# or build:
git clone https://github.com/NanoMindExplorer/codeforge.git
cd codeforge
CGO_ENABLED=0 go build -ldflags="-s -w" -o $PREFIX/bin/codeforge ./cmd/codeforge/
```

Lean binary (no glamour):

```bash
CGO_ENABLED=0 go build -tags plainmd -ldflags="-s -w" -o $PREFIX/bin/codeforge ./cmd/codeforge/
```

## Runtime tips

```bash
export GEMINI_API_KEY=...
export CODEFORGE_NO_MOTION=1
export CODEFORGE_PLAIN_MD=1   # optional
codeforge --skip-wizard --no-motion
```

## Headless on device

```bash
codeforge agent --json --workdir ~/project "run go test ./... and summarize"
```

## Optional: local package skeleton

```text
termux-packages/packages/codeforge/build.sh
TERMUX_PKG_HOMEPAGE=https://github.com/NanoMindExplorer/codeforge
TERMUX_PKG_DESCRIPTION="Terminal AI coding companion"
TERMUX_PKG_LICENSE="Apache-2.0"
TERMUX_PKG_VERSION=0.7.0
TERMUX_PKG_SRCURL=https://github.com/NanoMindExplorer/codeforge/archive/refs/tags/v${TERMUX_PKG_VERSION}.tar.gz
TERMUX_PKG_BUILD_IN_SRC=true
termux_step_make() {
  CGO_ENABLED=0 go build -ldflags="-s -w" -o codeforge ./cmd/codeforge/
}
termux_step_make_install() {
  install -Dm700 codeforge $TERMUX_PREFIX/bin/codeforge
}
```
