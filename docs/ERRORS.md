# Provider error UX

CodeForge never shows stack traces or raw API JSON to the user for provider failures.  
Errors are classified into stable **codes** with a short message + actionable hint.

## Codes

| Code | Typical cause | User action |
|------|---------------|-------------|
| `auth` | 401/403, missing/invalid key | `/setup` or set env key · `/provider` |
| `rate_limit` | 429, Retry-After | Wait or `/model` cheaper |
| `quota` | billing / insufficient_quota | Provider console |
| `model` | unknown model id | `/model` list |
| `context` | context window / payload too large | `/compact` or `/new` |
| `network` | DNS, TLS, connection refused, 502/503 | Network / base URL / ollama serve |
| `timeout` | deadline / cancel | Retry or faster model |
| `unsupported` | reasoning/thinking rejected | Auto-retry without thinking; or `CODEFORGE_REASONING=off` |

## Surfaces

| Surface | Behavior |
|---------|----------|
| **TUI stream / agent** | System block via `FormatUserError` (icon + message + hint + code) |
| **Toast** | One-line `Short()` / `FormatUserErrorShort` |
| **Headless `--json`** | `{"ok":false,"code":"…","error":"…","hint":"…"}` |
| **ACP** | Friendly text chunk + `codeforge/error` update |
| **Log (optional)** | `~/.codeforge/logs/provider-error.jsonl` (redacted raw) |

Disable log: `CODEFORGE_PROVIDER_ERROR_LOG=0`.

## Reasoning unsupported (auto)

If a request enables thinking and the API rejects it:

1. One automatic retry with `Reasoning=off`
2. System notice: continued without thinking
3. No stack / raw body in chat

## Examples (user-visible)

**Auth**
```text
🔑 API key rejected or missing  [grok]
  → Set XAI_API_KEY or GROK_API_KEY, then /provider grok · /setup
  code: auth
```

**Rate limit**
```text
⏳ Rate limited by the provider  [openai]
  → Wait ~20s, or switch model (/model)
  ↻ retry after ~20s
  code: rate_limit
```

**Reasoning**
```text
🧠 This model rejected reasoning/thinking parameters  [grok]
  → CodeForge retries without thinking automatically, or set CODEFORGE_REASONING=off
  ↻ safe to retry
  code: unsupported
```

## Developer API

```go
pe := provider.Classify(err, status, body, "grok")
provider.FormatUserError(err)       // multi-line TUI
provider.FormatUserErrorShort(err)  // toast
provider.HTTPErrorHeaders(name, status, body, resp.Header, transportErr)
provider.AuthError("gemini", "GEMINI_API_KEY not set")
```

See also: [ONBOARDING.md](./ONBOARDING.md) · [REASONING.md](./REASONING.md)
