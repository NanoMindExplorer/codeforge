# Dogfood Batch B–C (W2)

Use with [`TEMPLATE.md`](./TEMPLATE.md) and the main checklist in [`../DOGFOOD.md`](../DOGFOOD.md).

## Batch B — Session lifecycle (days 4–5)

| Task | Pass? | Notes |
|------|-------|-------|
| Kill terminal mid-task → `/resume` | | |
| `/fork` branch conversation | | |
| `/rewind` (or 2× Esc) restore files | | |
| `/compact` long thread | | |
| `session migrate` after upgrade from v0.8 | | |

**Exit:** 5/5 pass (or document intentional skip).

## Batch C — Safety (day 6)

| Task | Pass? | Notes |
|------|-------|-------|
| `rm -rf` denied by default rule | | |
| Shell ask modal y/n/a | | |
| Hook PreToolUse deny (exit 2) | | |
| DESIGN blocks project writes | | |
| Sandbox workspace / strict soft deny | | |

**Exit:** 0 false-allow for dangerous shell.

## Also verify (W2 product)

| Task | Pass? | Notes |
|------|-------|-------|
| First-run wizard appears once; second launch no spam | | |
| Footer shows `⚠ no API key · /setup` without keys | | |
| `/setup grok <key>` validates and clears footer | | |
| `/provider` shows key source (`env:…` / `config` / `missing`) | | |
| Bad key → friendly auth error (not raw JSON) | | |
| Reasoning-unsupported model → auto-retry notice | | |
| `codeforge agent --json "…"` without key → exit 2 + `code: no_provider` | | |
