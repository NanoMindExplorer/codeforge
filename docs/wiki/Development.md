# CodeForge Development & Architecture

## Architecture

```text
cmd/codeforge/          CLI entry, wizard, provider registration
internal/
  agent/                Tool-calling agent loop (events → TUI)
  provider/             Gemini · Claude · OpenAI · Ollama · MCP scaffold
  tool/                 Registry + sandboxed tools + StagedWriter + github tool
  github/               gh CLI + REST client (PRs, issues, checks, push)
  git/  diff/  config/  Supporting core
  theme/                Design tokens (single source of color truth)
  keymap/               Central keybindings
  session/              Persist / resume conversations
  checkpoint/           Local undo snapshots
  ui/
    components/         Panel, toast, badges
    markdown/           Glamour wrapper
    diffview/           Rich diff renderer
    palette/            Fuzzy command palette
    filepicker/         @file picker
    review/             Multi-file Plan review UI
  tui/                  Bubble Tea orchestrator (chat, panes, routing)
```

Design principles (Neo-Forge / Terminal Glass):

- **Depth over flatness** — elevation via surface/border tokens  
- **Motion carries meaning** — not decoration; kill-switch available  
- **Color is status language** — cyan AI · violet agent · emerald success · rose danger · amber attention  
- **Trust before write** — Plan mode default  

Strategy document: [`CODEFORGE_STRATEGY.md`](./CODEFORGE_STRATEGY.md).

---

## Development & tests

```bash
git clone https://github.com/NanoMindExplorer/codeforge.git
cd codeforge
go mod tidy
make install-hooks      # gofmt pre-commit (core.hooksPath=scripts/githooks)

# Format / check
make fmt                # gofmt -w .
make fmt-check          # fail if drift

# Unit + smoke tests
go test ./...

# Build
make build

# Run against this repo
export GEMINI_API_KEY=...
./codeforge --skip-wizard .
```

CI (GitHub Actions): on push/PR runs `check-version`, **gofmt**, `go test`, `go vet`, and a CGO-free build that must report the `VERSION` file. Tags matching `v*` run the [release workflow](./.github/workflows/release.yml) (tag must equal `VERSION`, then GoReleaser).

Local gates:

```bash
make ci                 # check-version + fmt-check + vet + coverage floor + build
make test-race          # -race on agent/tool/session/acp/provider/permission
make dogfood-offline    # measurable dogfood without live API
make govulncheck        # vuln scan (warn; GOVULNCHECK_STRICT=1 to fail)
make release-gate       # full automated public-ready gate
make bump V=1.9.4       # bump VERSION + all string locations
bash scripts/update-formula.sh v1.9.3   # after release: fill Formula sha256
```

---

## Distribution

| Artifact | Location |
|----------|----------|
| Install script | [`install.sh`](./install.sh) |
| GoReleaser config | [`.goreleaser.yaml`](./.goreleaser.yaml) |
| CI workflow | [`.github/workflows/ci.yml`](./.github/workflows/ci.yml) |
| Release workflow | [`.github/workflows/release.yml`](./.github/workflows/release.yml) |
| Version SSOT | [`VERSION`](./VERSION) |
| Homebrew formula | [`Formula/codeforge.rb`](./Formula/codeforge.rb) |
| Termux package | [`contrib/termux/`](./contrib/termux/) |
| Release notes helper | `make release-notes` |

Release matrix (intended): `linux/amd64`, `linux/arm64` (Termux), `darwin/arm64`, `windows/amd64`.

---

## Troubleshooting

| Symptom | What to try |
|---------|-------------|
| “Provider config” / no API key | `/setup` or export `XAI_API_KEY` / `GEMINI_API_KEY`. See [docs/ONBOARDING.md](./docs/ONBOARDING.md). |
| Rate limited / 429 | Wait or `/model` cheaper. Friendly message (not raw JSON) — [docs/ERRORS.md](./docs/ERRORS.md). |
| Reasoning / thinking rejected | Auto-retry without thinking, or `CODEFORGE_REASONING=off`. |
| Empty / hanging stream | Check network and key validity. Gemini free tier has rate limits. |
| Agent can’t see files outside project | By design — tools are sandboxed to the workdir. |
| Writes don’t appear on disk | You are in **BUILD** (staged) — finish the **review** overlay. Or `/mode yolo`. **DESIGN** blocks project writes by design. |
| Want to reverse a write | `/undo` for last applied file; or use git. |
| TUI feels laggy (SSH / phone) | `codeforge --no-motion` or `CODEFORGE_NO_MOTION=1`. |
| Icons look broken | Unset Nerd Font env, or install a Nerd Font and set `NERD_FONT=1`. |
| Ollama not listed | Ensure `ollama serve` is up; check `OLLAMA_HOST`. |
| Custom OpenAI proxy fails | Verify `OPENAI_BASE_URL` has no trailing slash issues; must expose `/chat/completions`. |
| Binary large (~21MB) | Expected with Glamour/Chroma; still pure Go / no CGO. |
| `/gh auth` fails | Run `gh auth login` or export `GITHUB_TOKEN` with `repo` scope. |
| `/pr create` fails | Ensure branch is pushed (`/push`), remote is GitHub, and you have permission. |
| Checks empty | Need `gh` CLI for best CI rollup; open PR first. |

---

