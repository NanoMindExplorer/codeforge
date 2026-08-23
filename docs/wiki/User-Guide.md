# CodeForge User Guide

## User guide

### Interface (Grok 4.5–style)

```
┃ you
┃ fix the race in worker.go

┃ ⠋ working
┃ ◆ read_file  worker.go
┃   ✓ 84 lines
┃ Here's the race: …

╭─ ❯ ask anything, /command, or @file ─────────────────╮
│ ▌                                                     │
╰───────────────────────────────────────────────────────╯
 PROMPT  PLAN  gemini · flash  gh:@you · main  $0.01  groknight  14:02
 tab focus  @ file  / commands  ctrl+k  shift+tab plan/act  ctrl+b panels
```

| Region | Role |
|--------|------|
| **Scrollback** | Full-width blocks with left accent bars (you / assistant / tools) |
| **Prompt** | Bottom `❯` composer — focused by default |
| **Footer** | PROMPT/SCROLL · PLAN/ACT · model · git/gh · cost · theme |
| **Panels** | Optional Diff + Files (`Ctrl+B`) |

### Focus & keys

| Key | Action |
|-----|--------|
| *(type)* | Auto-focus prompt |
| `Tab` | Prompt ↔ scrollback |
| `Esc` / **2× Esc** | Scrollback / clear prompt |
| `i` / `Space` | From scrollback → prompt |
| `@` | File picker (cached list; Esc closes) |
| `/` | Slash commands (+ hint strip) |
| `Ctrl+K` | Palette (Esc closes → prompt) |
| `Ctrl+C` | Clear draft → cancel stream → quit |
| `Shift+Tab` | BUILD → DESIGN → YOLO |
| `Ctrl+B` | Toggle side panels |
| `/theme` | Live-preview picker · or `/theme tokyonight` / `auto` |
| `/resume` | Session picker · `/new` `/fork` `/rewind` `/compact` |
| `Shift+Tab` | **BUILD → DESIGN → YOLO** session mode |
| `/plan` | Enter DESIGN plan mode · `/view-plan` approval |
| `/permissions` | allow/deny/ask rules · modes · remember |
| `/hooks` | List PreToolUse / PostToolUse hooks |
| `/todos` | Task list · footer ☑ 2/5 |
| `/memory` | Cross-session notes (`list` / `add` / `search`) · Grok memory tools |
| `/skills` | Grok-compatible SKILL.md packages · invoke `/name` |
| `/personas` | Subagent personas (researcher, concise, reviewer, custom) |
| `/subagents` | Background/recent subagent jobs (show/cancel) |
| `/tasks` | Background shell jobs |
| `/settings` | Settings panel |
| Enter / y | Fullscreen block · copy body |
| `/compact-mode` | Tighter padding (outer_vpad=0) |

### Chat vs agent

| Path | How | Tools? | Best for |
|------|-----|--------|----------|
| **Streaming chat** | Type natural language → Enter | No | Q&A, explanations |
| **Agent** | `/act <task>` or `/read`, `/fix`, … | Yes | Edit code, search, builds |

Agent system behavior (summary):

- Prefer **read before write**
- Uses filesystem tools under the **project workdir** only (path sandbox)
- In **BUILD** mode, `write_file` is staged until you approve; **DESIGN** blocks project writes

### Session modes (BUILD / DESIGN / YOLO)

| | **BUILD** (default) | **DESIGN** | **YOLO** |
|---|---------------------|------------|----------|
| Reads / search | Free | Free | Free |
| `run_command` | Free | Free | Free |
| Project file writes | **Staged** → review UI | **Blocked** | **Immediate** |
| `plan.md` / `write_plan` | Free | Free (auto) | Free |
| Toggle | `Shift+Tab` cycle · `/mode build\|design\|yolo` · `/plan` |

**Recommendation:** **DESIGN** for ambiguous architecture; **BUILD** for normal work; **YOLO** only for tight trusted loops.

### Review overlay

When the agent finishes a turn and there are pending writes:

| Key | Action |
|-----|--------|
| `j` / `k` | Move between changed files |
| `Space` | Toggle accept / reject for current file |
| `a` | Accept all |
| `r` | Reject all |
| `Enter` | Apply accepted files to disk (+ checkpoints) |
| `Esc` | Cancel review (leave pending / discard flow as implemented) |

Accepted files are written to disk and previous contents are checkpointed for `/undo`.

### File mentions (`@file`)

1. Enter **INSERT** (`i`).
2. Press **`@`** → fuzzy file picker opens.
3. Type to filter · `↑`/`↓` · **Enter** to select.
4. The prompt gains `@path` and the file body is **attached** as context for the next send.

Useful for: “explain this file”, “refactor this module”, without a separate `/read` first.

### Command palette

**Ctrl+K** opens a fuzzy overlay fed by three sources:

1. Slash commands (`/act`, `/fix`, …)
2. Project files
3. Saved sessions

Navigate with `↑`/`↓` (or `j`/`k`), confirm with **Enter**, close with **Esc**.

### Sessions (Phase 4)

- Layout: `~/.codeforge/sessions/<encoded-cwd>/<session-id>/` with `summary.json`, `chat_history.jsonl`, `updates.jsonl`, `rewind_points.jsonl`.
- **`/resume`** — full-screen picker (filter, preview, Enter). **`/sessions <id>`** still works.
- **`/new`** — new session id · **`/clear`** — wipe chat only (same id).
- **`/fork`** · **`/rewind`** (also **2× Esc** idle) · **`/compact`** · **`/context`** · **`/session-info`**.
- Headless agent writes the same layout and returns `session_id` in JSON.

### Undo / checkpoints

- When a write is **applied** (review accept, or Act mode), a snapshot of the previous content is stored under `~/.codeforge/checkpoints/<session-id>/`.
- **`/undo`** restores the **last** written file · **`/rewind`** restores all files after a turn.

This complements—not replaces—git. Prefer git commits for permanent history.

### Git helpers

If the workdir is a git repository (CodeForge may init one if missing):

| Command | Effect |
|---------|--------|
| `/status` | Show branch + working tree status; refresh file glyphs |
| `/commit [msg]` | `git add -A` + commit (optional message) |
| `/push` | `git push -u origin HEAD` |
| `/pull` | `git pull` (ff-only, then plain pull fallback) |

---

## GitHub integration

### What you can do

| Capability | Slash command | Agent tool action |
|------------|---------------|-------------------|
| Auth / identity | `/gh auth` | `auth_status` |
| Repo metadata | `/gh repo` | `repo_view` |
| List / view PRs | `/pr list` · `/pr view [n]` | `pr_list` · `pr_view` |
| Create PR | `/pr create <title> [| body]` | `pr_create` |
| Merge PR | `/pr merge <n> [squash\|merge\|rebase]` | `pr_merge` |
| CI checks | `/pr checks [n]` | `checks` |
| Issues | `/issue list` · `/issue view` · `/issue create` | `issue_*` |
| Push / pull | `/push` · `/pull` | `push` · `pull` |
| Branch | `/gh branch [name]` | `branch_create` |
| Log | `/gh log` | `log` |

### End-to-end: ship a feature like an AI agent

```text
1. /mode plan                    # safe writes
2. /act implement feature X using search_replace/apply_patch
3. Review overlay → Enter        # apply patches
4. /commit feat: implement X
5. /push
6. /pr create feat: implement X | ## Summary …
7. /pr babysit --fix           # poll CI; on failure auto-agent-fix
```

Or in one agent turn:

```text
/act implement the change with search_replace, run tests, commit, push,
     open a PR, then babysit checks until green (fix and push if red)
```

### Surgical edits

Prefer agent tools:

| Tool | Use when |
|------|----------|
| `search_replace` | Exact old→new text (unique match or `replace_all`) |
| `apply_patch` | Multi-hunk / multi-file CodeForge patch format |
| `write_file` | New files or full rewrites only |

### Multi-root monorepo

In `~/.config/codeforge/config.yaml`:

```yaml
workspace:
  extra_roots:
    - ../shared-lib
    - /abs/path/to/package
  # optional override:
  # ignore_dirs: [node_modules, vendor, dist]
```

Paths resolve against primary workdir first, then extra roots. Grep skips secrets (`.env`, `*.pem`) and heavy dirs by default.

### PR babysit

```text
/pr babysit              # current branch PR
/pr babysit 42           # PR #42
/pr babysit 42 --fix    # on failure → agent fix loop
```

Also via agent: `github` action `babysit` / `babysit_once`.

### Project rules (AGENTS.md)

Place any of these in the project root (merged if several exist):

- `AGENTS.md` · `CLAUDE.md` · `CODEFORGE.md`
- `.codeforge/rules.md` · `.cursorrules` · `.github/copilot-instructions.md`

```text
/rules          # show loaded rules in chat
```

Rules are injected into every chat + agent system prompt.

### Codebase intelligence

```text
/index                              # stats
/act where is authentication handled?
# agent uses codebase_search → read_file → …
```

### MCP servers

```yaml
# ~/.config/codeforge/config.yaml
mcp:
  servers:
    - name: filesystem
      command: npx
      args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
```

Tools appear as `mcp_<server>_<tool>` for the agent.

### Cost budget

```yaml
budget:
  max_cost_usd: 2.0
  warn_at_usd: 1.0
```

```text
/budget
```

When the cap is hit, chat/agent submits are blocked until config is raised.

### Headless / CI mode (Tier-3)

```bash
# Human-readable
codeforge agent "run go test ./... and fix failures"

# Machine-readable (CI)
codeforge agent --json --workdir . "run go test ./internal/... "
echo $?   # 0 ok, 1 agent/tool failure

# Plan mode (stage writes — not applied)
codeforge agent --plan --json "propose a patch for README typos"
```

GitHub Actions example:

```yaml
- name: CodeForge agent
  env:
    GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
  run: |
    codeforge agent --json --act "run go test ./... and summarize failures" | tee agent-out.json
```

### Plugins

Drop a YAML file into `~/.codeforge/plugins/` (see `examples/plugins/echo.plugin.yaml`):

```yaml
name: mytool
description: Does a thing
command: /path/to/binary
args: []
workdir_relative: true
```

Appears to the agent as `plugin_mytool`. Extra dirs: `plugins.dirs` in config or `CODEFORGE_PLUGIN_DIR`.

### Session sync (laptop ↔ VPS)

```bash
export CODEFORGE_SESSIONS_DIR="$HOME/Sync/codeforge-sessions"
codeforge session list
codeforge session export 20260716-101500 ./backup.json
codeforge session import ./backup.json
codeforge session export-all ./all-sessions/
```

### Telemetry (opt-in)

Default **off**. Enable local JSONL only:

```yaml
telemetry:
  enabled: true
  local_only: true
```

```bash
export CODEFORGE_TELEMETRY=1
# events → ~/.codeforge/telemetry/events.jsonl
# never includes source code or prompt text
```

### Architecture note

```text
internal/app/        shared bootstrap (TUI + headless)
internal/headless/   CI agent runner (--json)
internal/plugin/     YAML command plugins
internal/telemetry/  opt-in local analytics
internal/rules/      AGENTS.md loader
internal/index/      offline codebase index
internal/redact/     secret redaction
internal/github/     gh + babysit
internal/workspace/  multi-root sandbox
internal/tool/       agent tools
internal/agent/      tool loop + progress
```

---

