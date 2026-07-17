# Streaming + tools (Q7.4)

## Current agent loop

The multi-step tool loop in `internal/agent` uses **`Complete` (non-stream)** for every
tool-bearing turn:

```text
user → Complete(tools) → tool_calls? → execute tools → Complete again → …
```

Plain chat in the TUI uses **`Stream`** (with reasoning fallback) for the first token
path when no agent tools are needed.

## Why Complete for tools?

| Provider | Stream + tools | Notes |
|----------|----------------|-------|
| **xAI Grok** | Supported (OpenAI-compatible) | Could stream tool deltas |
| **OpenAI** | Supported | `tool_calls` on stream deltas |
| **Gemini** | Partial / evolving | Prefer Complete for stability |
| **Claude** | **No** tools on Stream path | Documented: Stream is text/thinking only |
| **Ollama** | Varies by model | Prefer Complete |

Using **Complete for the tool loop** keeps one code path that works on all providers
and matches permission/redact hooks after each full response.

## Optional experimental path

Env flag (future / experimental):

```bash
# Not required for production. When enabled, agent may attempt Stream for
# tool-capable providers (Grok/OpenAI) and fall back to Complete on error.
export CODEFORGE_AGENT_STREAM=1
```

Status in this release: **design only** — `agent.Run` still uses Complete.
A stream+tools path must:

1. Detect provider capability (Grok/OpenAI yes; Claude no)
2. Accumulate partial tool_call deltas into complete `ToolCall`s
3. Preserve cancel via context
4. Fall back to Complete on protocol errors

## UX implication

- **Chat (no tools):** low time-to-first-token via Stream  
- **Agent (tools):** slightly higher TTFB; tool progress still streams to the UI via
  `EventToolCall` / `EventToolProgress` as tools execute locally  

## Related

- `internal/agent/agent.go` — `completeWithRetries`  
- `internal/provider/fallback.go` — reasoning retry  
- `docs/REASONING.md`  
