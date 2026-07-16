# Termux package notes

→ Full guide: [`contrib/termux/README.md`](./termux/README.md)

## Quick install (recommended)

```bash
pkg install -y golang git curl
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | sh
codeforge version
```

## Build from source

```bash
pkg install -y golang git
git clone https://github.com/NanoMindExplorer/codeforge.git
cd codeforge
bash contrib/termux/build.sh
```

## Package metadata

```bash
bash contrib/termux/package.sh   # prints TERMUX_PKG_* with VERSION from repo
```

## Runtime tips

```bash
export XAI_API_KEY=…   # preferred; or GEMINI_API_KEY
export CODEFORGE_NO_MOTION=1
codeforge --skip-wizard --no-motion
```

## Headless on device

```bash
codeforge agent --json --workdir ~/project "run go test ./... and summarize"
# no key → exit 2 + {"code":"no_provider",...}
```
