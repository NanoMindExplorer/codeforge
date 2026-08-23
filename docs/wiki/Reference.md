# Reference (Keybindings & Config)

## Keybindings reference

### Global

| Key | Action |
|-----|--------|
| `Ctrl+C` | Quit (session is saved) |
| `Ctrl+L` | Clear terminal screen |
| `q` | Quit from NORMAL (session is saved) |
| `?` | Show help text in chat |

### NORMAL mode

| Key | Action |
|-----|--------|
| `i` | INSERT (empty chat input) |
| `I` | INSERT with `/act ` prefilled |
| `/` | INSERT with `/` prefilled |
| `:` | COMMAND line |
| `Ctrl+K` | Command palette |
| `Shift+Tab` | Cycle BUILD → DESIGN → YOLO |
| `1` / `2` / `3` | Focus Chat / Diff / Files |
| `Tab` | Prompt ↔ scrollback |
| `j` `k` / arrows | Scroll chat (or Diff navigation) |
| `g` / `G` | Top / bottom of chat |
| `PgUp` / `PgDn` · `Ctrl+U` / `Ctrl+D` | Page scroll |
| `n` / `p` | Next / previous file tab in Diff pane |

### INSERT mode

| Key | Action |
|-----|--------|
| Type | Edit multi-line prompt |
| `Enter` | Send chat **or** run slash command if line starts with `/` |
| `Esc` | Back to NORMAL |
| `@` | Open file mention picker |
| `Ctrl+K` | Open palette |
| `↑` / `↓` | Input history (previous prompts) |

### Review mode

See [Review overlay](#review-overlay).

---

## Slash commands reference

Type in INSERT (prefix `/`) or via `:` / palette. Aliases in parentheses.

### Agent & code

| Command | Description | Example |
|---------|-------------|---------|
| `/act` (`/a`) | Free-form agent task with tools | `/act add retries to HTTP client` |
| `/read` (`/r`) | Read and display a file | `/read internal/agent/agent.go` |
| `/ls` (`/list`) | List a directory | `/ls cmd` |
| `/grep` (`/find`) | Search project with regex | `/grep TODO` |
| `/run` | Run a shell command in project root (via agent) | `/run go test ./...` |
| `/explain` (`/e`) | Deep explanation of a file | `/explain main.go` |
| `/fix` | Find and fix bugs in a file | `/fix handler.go` |

### Provider & session

| Command | Description | Example |
|---------|-------------|---------|
| `/provider` (`/p`) | List or switch provider | `/provider claude` |
| `/model` (`/m`) | List or switch model | `/model gemini-2.5-pro` |
| `/mode` | BUILD / DESIGN / YOLO | `/mode design` · `/mode yolo` |
| `/plan` | Enter DESIGN (+ optional task) | `/plan add auth` |
| `/view-plan` | Plan approval UI (a/s/q) | `/view-plan` |
| `/resume` | Session picker | `/resume` · `/resume <id>` |
| `/new` | New session id | `/new` |
| `/fork` | Branch conversation | `/fork` · `/fork continue with X` |
| `/rewind` | Restore files + truncate chat | `/rewind` · `/rewind last` |
| `/compact` | Compress history | `/compact` · `/compact keep API` |
| `/context` | Token breakdown | `/context` |
| `/session-info` | Session metadata | `/session-info` |
| `/sessions` | List sessions | `/sessions` · `/sessions <id>` |
| `/undo` | Restore last applied write | `/undo` |
| `/cost` (`/c`) | Session tokens, cost, duration | `/cost` |
| `/clear` | Clear chat + start a fresh session id | `/clear` |

### Git & GitHub

| Command | Description |
|---------|-------------|
| `/status` (`/s`) | Local git status |
| `/commit [msg]` | Stage all + commit |
| `/push` | Push current branch to origin |
| `/pull` | Pull from remote |
| `/gh` … | GitHub hub (`/gh help` for full list) |
| `/pr` … | Pull requests (list/view/create/merge/checks) |
| `/issue` … | Issues (list/view/create) |

### Meta

| Command | Description |
|---------|-------------|
| `/help` (`/h` `/?`) | In-app help |
| `/about` | Version / author / stack |
| `/quit` (`/q` `/exit`) | Exit CodeForge |

Unknown `/…` strings are forwarded to the **agent** as a task.

**Tab** in the command line autocompletes known slash commands.

---

## CLI flags

```text
codeforge [workdir] [flags]

  workdir          Optional project directory (default: current directory)

  --no-motion      Disable animations (slow SSH / Termux)
  --minimal        No chrome; terminal-native 16 colors
  --compact        Tighter padding (same as /compact-mode)
  --skip-wizard    Skip first-run setup wizard
  -y, --yes        Same as --skip-wizard
  -h, --help       Print CLI help
  -v, --version    Print version
```

Examples:

```bash
codeforge
codeforge ~/src/myapp
codeforge --skip-wizard --no-motion ~/src/myapp
codeforge --minimal --compact
CODEFORGE_THEME=auto codeforge
```

---

## Environment variables

| Variable | Purpose |
|----------|---------|
| `XAI_API_KEY` / `GROK_API_KEY` | xAI Grok 4.5 (preferred) |
| `GEMINI_API_KEY` | Google Gemini |
| `ANTHROPIC_API_KEY` | Anthropic Claude |
| `OPENAI_API_KEY` | OpenAI or compatible API |
| `OPENAI_BASE_URL` | Override API base (default `https://api.openai.com/v1`) |
| `OLLAMA_HOST` | Ollama base URL (default `http://127.0.0.1:11434`) |
| `OLLAMA_MODEL` | Default Ollama model (default `llama3.2`) |
| `GITHUB_TOKEN` / `GH_TOKEN` | GitHub REST auth (optional if `gh auth login` is done) |
| `CODEFORGE_THEME` | `groknight` (default), `grokday`, `tokyonight`, `rosepine`, `oscura`, `aurora`, `auto` |
| `CODEFORGE_AUTO_DARK` / `CODEFORGE_AUTO_LIGHT` | Themes mapped when `theme=auto` |
| `CODEFORGE_COMPACT` / `CODEFORGE_MINIMAL` | Compact padding / terminal-native 16-color chrome |
| `CODEFORGE_COLOR` | Force quantize: `true` · `256` · `16` · `none` |
| `CODEFORGE_NO_MOTION` | `1` / `true` disables motion |
| `NO_COLOR` | Monochrome + no motion (a11y) |
| `CODEFORGE_SSH_TUNE` | Auto compact + no-motion when SSH_* is set |
| `CODEFORGE_PLAIN_MD` / `CODEFORGE_NO_GLAMOUR` | Skip rich markdown (faster / leaner) |
| `NERD_FONT` / `NERD_FONTS` | Prefer Nerd Font file/git glyphs |

Optional smaller binary (no glamour/chroma at compile time):

```bash
CGO_ENABLED=0 go build -tags plainmd -ldflags="-s -w" -o codeforge ./cmd/codeforge/
```

---

## Configuration files

| Path | Purpose |
|------|---------|
| `~/.config/codeforge/config.yaml` | Default provider, theme, git, permissions (example created on first run) |
| `~/.codeforge/theme.yaml` | Optional color token overrides |
| `~/.codeforge/sessions/<cwd>/<id>/` | Sessions (summary.json, chat_history.jsonl, rewind_points) |
| `~/.codeforge/checkpoints/<session-id>/` | Pre-write file snapshots for `/undo` |

Example `config.yaml` keys (see generated file for full template):

```yaml
default_provider: gemini
theme: groknight   # or auto / tokyonight / rosepine / oscura
ui:
  compact_mode: false
  auto_dark_theme: groknight
  auto_light_theme: grokday
session:
  auto_compact_pct: 0.85   # auto /compact near context limit
permissions:
  mode: default   # default | plan | always_approve | dont_ask
  require_confirm_write: true
  require_confirm_shell: true
  require_confirm_push: true
  # rules:
  #   - { tool: run_command, pattern: "rm -rf *", effect: deny }
  #   - { tool: run_command, pattern: "go test *", effect: allow }
git:
  auto_commit: true
  commit_style: conventional
  branch_prefix: ai/
```

---

