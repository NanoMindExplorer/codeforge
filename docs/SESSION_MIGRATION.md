# Session migration (v0.8 flat → Phase 4 v2)

CodeForge **v0.8** stored each conversation as a single file:

```text
~/.codeforge/sessions/<id>-<slug>.json
```

**v0.9.3+ (Phase 4)** uses Grok-style directories:

```text
~/.codeforge/sessions/<encoded-cwd>/<session-id>/
  summary.json
  chat_history.jsonl
  updates.jsonl
  rewind_points.jsonl
  plan.md
```

## Automatic compatibility

- **Load** still finds legacy flat files by id prefix.
- **List** merges v2 groups and legacy JSON.
- New sessions always write **v2**.

## One-shot migrate

```bash
codeforge session migrate
```

What it does:

1. Scans the sessions root (`CODEFORGE_SESSIONS_DIR` or `~/.codeforge/sessions`).
2. For each flat `*.json` that looks like a session, rewrites it as a v2 directory under the encoded workdir group.
3. Renames the original to `*.json.legacy` (safe backup).
4. Skips files that already have a matching v2 directory.

Example output:

```text
migrated 12 · skipped 3
```

## Shared / SSH storage

```bash
export CODEFORGE_SESSIONS_DIR=/path/to/sync/sessions
codeforge session migrate
```

## Verify

```bash
codeforge session list
codeforge   # then /resume
```

## Rollback

If something looks wrong, restore from `*.json.legacy`:

```bash
cd ~/.codeforge/sessions
mv 20260101-120000-hello.json.legacy 20260101-120000-hello.json
```

(You may also delete the corresponding v2 directory under the encoded-cwd folder.)
