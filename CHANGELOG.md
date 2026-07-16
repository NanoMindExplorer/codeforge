# Changelog

All notable changes to CodeForge are documented here.

## [1.8.3] — 2026-07-16

### Onboarding (W2 / O1–O5)

- `~/.codeforge/onboarding.json` tracks completed/skipped first-run (no wizard spam).
- Wizard v2: pick provider → paste key (prefix detect) → ValidateConfig → default model → save config.
- Footer strip: `⚠ no API key · /setup` until a provider validates.
- `/setup` slash (re-run anytime): `/setup <provider> <key> [model]`.
- `/provider` lists key source: `env:XAI_API_KEY` / `config` / `missing`.

### Provider error UX (W2 / E4–E5)

- Reasoning unsupported → one automatic retry with `Reasoning=off` + system notice (agent + stream).
- Headless `--json`: structured `code` + `hint`; exit **2** for `no_provider` / `auth`.
- ACP surfaces `FormatUserError` and `codeforge/error` session updates.

### Dogfood

- Batch B–C checklist: `docs/dogfood/BATCH_BC.md`.

## [1.8.2] — 2026-07-16

### Release automation (W1 / R1–R3)

- `VERSION` is the single source of truth for the product version.
- `scripts/check-version.sh` + `make check-version` / `make ci` gate consistency across `main.go`, README, TUI about, MCP, ACP, and Homebrew Formula.
- `scripts/bump-version.sh` updates all version string locations in one shot.
- `scripts/update-formula.sh` fills Formula sha256 from a published release checksums file.
- CI (`.github/workflows/ci.yml`) runs the version gate before tests/build and smoke-checks `codeforge version`.
- Dedicated release workflow (`.github/workflows/release.yml`) on tag `v*`: tag must match `VERSION`, then GoReleaser publishes.

### Provider error UX (W1 / E1–E3)

- New `provider.ProviderError` with codes: `auth`, `rate_limit`, `quota`, `model`, `context`, `network`, `timeout`, `unsupported`.
- `Classify` / `HTTPError` / `FormatUserError` map HTTP and transport failures to short messages + actionable hints.
- Wired through OpenAI-compatible (incl. Grok), Gemini, Claude, and Ollama Complete/Stream paths.
- TUI stream, agent, and `errMsg` paths render `FormatUserError` instead of raw dump strings.

### Dogfood

- Daily log template at `docs/dogfood/TEMPLATE.md`.
- Checklist remains in `docs/DOGFOOD.md`.

## [1.8.1] — previous

- Permissions, subagent auth, provider config keys, ACP skills + toolCallId, clone parent registry for general subagents (see git history).
