# Native reasoning / thinking streams

CodeForge surfaces **provider-native** chain-of-thought tokens in the scrollback **thinking** block (and ACP `agent_thought_chunk`), instead of only a synthetic `planning…` placeholder.

## Providers

| Provider | Mechanism |
|----------|-----------|
| **Grok / OpenAI-compatible** | `delta.reasoning_content` / `delta.reasoning` · `message.reasoning_content` · optional `include_reasoning` + `reasoning_effort` |
| **Gemini 2.5+** | `thinkingConfig.includeThoughts` · parts with `"thought": true` |
| **Claude** | Extended thinking (`thinking: { type: enabled, budget_tokens }`) · content type `thinking` / stream `thinking_delta` |
| **Ollama** | No special request; text-only stream (no native split) |

## Auto-enable

Reasoning is requested when the model id looks like a thinking model (`grok`, `o1`/`o3`/`o4`, `gemini-2.5`, `claude`, …), unless disabled.

| Control | Values |
|---------|--------|
| Env `CODEFORGE_REASONING` | `on` / `off` / `low` / `medium` / `high` |
| Request field `reasoning` | same (used internally by tools/tests) |

Examples:

```bash
export CODEFORGE_REASONING=off     # never request thinking
export CODEFORGE_REASONING=high    # prefer high effort where supported
```

## UI

- Chat stream: reasoning deltas → **thinking** block; answer deltas → **assistant** block
- Agent loop: `Complete` responses with `reasoning` → `EventThinking` (then tools / text)
- Placeholder `planning…` is **replaced** by the first native thinking chunk
- Token footer may show `· N reasoning` when the API reports reasoning token counts

## ACP

```json
{
  "method": "session/update",
  "params": {
    "sessionId": "…",
    "update": {
      "sessionUpdate": "agent_thought_chunk",
      "content": { "type": "text", "text": "…" }
    }
  }
}
```

## Notes

- Thinking is **not** stored as assistant content in the message history (avoids leaking CoT into later turns unless the provider requires it).
- Some APIs reject unknown fields (`include_reasoning`); if a call fails, retry with reasoning off or pick a model that supports it.
- Claude: temperature is omitted when extended thinking is enabled (API constraint).
