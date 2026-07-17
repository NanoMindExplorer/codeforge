# Threat model & mitigations (Q8)

**Scope:** CodeForge TUI / headless / ACP agent running on a developer machine or CI.  
**Assets:** API keys, source code, session transcripts, host filesystem, network credentials.

## Actors

| Actor | Capability |
|-------|------------|
| **User** | Runs CodeForge, pastes keys, approves tools |
| **Model / agent** | Proposes tool calls (read/write/shell) under policy |
| **Malicious repo** | Crafted files, hooks, skills, MCP configs |
| **Network** | Provider APIs, MCP servers, web_fetch |
| **Local malware** | Same-user process; disk read of config/sessions |

## Threats → mitigations

| ID | Threat | Mitigation | Code / doc |
|----|--------|------------|------------|
| T1 | API key theft from disk | Prefer env; keys/ 0600; optional keyring; config 0600 | `docs/SECRETS.md` Q3 |
| T2 | Key leak to model context | Redact patterns + sensitive filenames | `internal/redact` Q8.4 |
| T3 | Destructive shell (`rm -rf`) | Permission deny rules; high-risk badge; no always-remember for dangerous | `permission` + `permask` Q8.2 |
| T4 | Write outside project | Path sandbox + OS sandbox profiles | `sandbox` G4; **default workspace** Q8.1 |
| T5 | Shell env injection (`LD_PRELOAD`) | scrubShellEnv; no env from tool JSON | `run_command` Q8.3 |
| T6 | Shell cwd escape | Always pin `cmd.Dir` to tool WorkDir | Q8.3 |
| T7 | Session transcript leak | Session dirs **0700**, files **0600** | Q8.5 |
| T8 | ACP unauthorized access | Bearer / `?secret=` on WebSocket | `docs/ACP.md` Q6 |
| T9 | Supply-chain CI compromise | Pin Actions to commit SHAs; Dependabot | Q8.6 |
| T10 | YOLO over-trust | Deny-dangerous still applies; UX labels risk | permissions + review overlay |

## Defaults (interactive)

| Setting | Default after Q8 | Opt-out |
|---------|------------------|---------|
| `sandbox.profile` | **`workspace`** | `off` / `CODEFORGE_SANDBOX=off` / `--sandbox off` |
| Session store | `~/.codeforge/sessions` mode **0700** | `CODEFORGE_SESSIONS_DIR` |
| Config | `~/.config/codeforge` **0600** files | — |
| Shell env | Host env minus injection vars | — |

## Residual risks (accepted)

- Same-user malware can still read `0600` files  
- Model can still run any **allowed** shell within sandbox (treat YOLO carefully)  
- `web_fetch` / MCP network trust is user-configured  
- List prices / free tiers are not security boundaries  

## Related docs

- [SANDBOX.md](./SANDBOX.md) · [SECRETS.md](./SECRETS.md) · [ERRORS.md](./ERRORS.md) · [ACP.md](./ACP.md)
