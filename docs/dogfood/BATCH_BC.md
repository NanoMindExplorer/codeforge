# Dogfood Batch B–C (W2)

Use with [`TEMPLATE.md`](./TEMPLATE.md) and the main checklist in [`../DOGFOOD.md`](../DOGFOOD.md).

## Batch B — Session lifecycle (days 4–5) ✅ **API complete (Q4)**

| Task | Pass? | Notes |
|------|-------|-------|
| Kill terminal mid-task → `/resume` | ✅ API | `session/crash_test.go` (partial jsonl recovery, atomic save). Field: `/resume last` after real kill still recommended. |
| `/fork` branch conversation | ✅ API | `TestForkAndRewind` |
| `/rewind` (or 2× Esc) restore files | ✅ API | rewind points + `checkpoint.UndoAfter` |
| `/compact` long thread | ✅ API | tool outcomes preserved (`compact_test.go` Q4.4) |
| `session migrate` after upgrade from v0.8 | ✅ API | `migrate_test.go` |

**Exit:** 5/5 automated. Optional HUMAN: kill live TUI mid-agent and `/resume last`.

### Q4 UX helpers

| Command | Effect |
|---------|--------|
| `/resume` | Interactive picker |
| `/resume last` | Newest session for this cwd |
| `/resume list` | Text table (id · time · model · mode · preview) |
| `/sessions` | Same table as `/resume list` |

Export/import round-trip includes `permission_mode`, `session_mode`, `model` (Q4.5).

## Batch C — Safety (day 6)

| Task | Pass? | Notes |
|------|-------|-------|
| `rm -rf` denied by default rule | ✅ | permission unit + dogfood |
| Shell ask modal y/n/a | ⬜ | HUMAN interactive |
| Hook PreToolUse deny (exit 2) | ✅ | hooks tests |
| DESIGN blocks project writes | ✅ | staged/design unit |
| Sandbox workspace / strict soft deny | ✅ | sandbox unit |

**Exit:** 0 false-allow for dangerous shell (automated). Shell modal still field.

## Also verify (W2 product)

| Task | Pass? | Notes |
|------|-------|-------|
| First-run wizard appears once; second launch no spam | ⬜ | HUMAN |
| Footer shows `⚠ no API key · /setup` without keys | ⬜ | HUMAN |
| `/setup grok <key>` validates and clears footer | ⬜ | HUMAN / Q3 secrets |
| `/provider` shows key source (`env:…` / `config` / `missing`) | ✅ | onboarding tests |
| Bad key → friendly auth error (not raw JSON) | ✅ | provider errors |
| Reasoning-unsupported model → auto-retry notice | ✅ | agent Q1 |
| `codeforge agent --json "…"` without key → exit 2 + `code: no_provider` | ✅ | headless |
