# Dogfood Batch D–E (W3)

Use with [`TEMPLATE.md`](./TEMPLATE.md) and [`../DOGFOOD.md`](../DOGFOOD.md).

## Batch D — Grok surface (days 7–8)

| Task | Pass? | Notes |
|------|-------|-------|
| Skills load (`/skills`, SKILL.md inject) | | |
| Personas for spawn (`/personas`) | | |
| `spawn_subagent` background + `/subagents` | | |
| `/pager` or pager.toml layout tweak | | |
| Native reasoning / thinking blocks | | |
| Reasoning-unsupported → auto-retry notice | | |

**Exit:** ≥1 real task per feature (not just `/help` dump).

## Batch E — IDE / ACP (day 9)

| Task | Pass? | Notes |
|------|-------|-------|
| `codeforge agent stdio` initialize + session/new | | |
| Multi-turn tool call over ACP | | |
| Friendly provider error chunk on bad key | | |
| `codeforge agent --json` happy path | | |
| `codeforge agent --json` no key → exit 2 + `no_provider` | | |

Minimal stdio smoke:

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":1}}' \
  | codeforge agent stdio
```

## Packaging smoke (W3)

| Platform | Command | `codeforge version` |
|----------|---------|---------------------|
| Linux/macOS curl | `curl -fsSL …/install.sh \| sh` | matches `VERSION` / latest tag |
| From source | `make build && ./codeforge version` | = `VERSION` file |
| Termux | `bash contrib/termux/build.sh` | = `VERSION` |
| Homebrew formula | `brew install --build-from-source Formula/codeforge.rb` | after release + sha256 |
