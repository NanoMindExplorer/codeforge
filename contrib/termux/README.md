# Termux packaging (R4)

## One command (device)

```bash
pkg install -y golang git curl
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | sh
codeforge version
```

## Build from this repo

```bash
pkg install -y golang git
git clone https://github.com/NanoMindExplorer/codeforge.git
cd codeforge
bash contrib/termux/build.sh          # installs to $PREFIX/bin when set
# or: bash contrib/termux/build.sh "$PREFIX/bin"
codeforge version                     # must match VERSION file
```

Lean binary (no glamour):

```bash
CGO_ENABLED=0 go build -tags plainmd -ldflags="-s -w -X main.ProjectVersion=$(tr -d '[:space:]' < VERSION)" \
  -o "$PREFIX/bin/codeforge" ./cmd/codeforge/
```

## termux-packages skeleton

```bash
bash contrib/termux/package.sh > /tmp/codeforge-termux.txt
# paste into packages/codeforge/build.sh in a termux-packages fork
```

`TERMUX_PKG_VERSION` is always derived from the repo `VERSION` file (same as `scripts/check-version.sh`).

## Runtime tips

```bash
export XAI_API_KEY=…          # or GEMINI_API_KEY
export CODEFORGE_NO_MOTION=1
export CODEFORGE_COMPACT=1
codeforge --skip-wizard --no-motion
```
