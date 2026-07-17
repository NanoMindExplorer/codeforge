# Token & cost accounting (Q7.5)

CodeForge estimates **USD session cost** from provider-reported token counts and
static **list prices** (USD per 1M tokens). This is a **budget UX estimate**, not
a billing invoice.

## Formula

```text
cost_usd = (input_tokens  * input_cost_per_1m  / 1_000_000)
         + (output_tokens * output_cost_per_1m / 1_000_000)
```

Implemented by `provider.CostForModel(p, modelID, in, out)`.

When the API reports separate **reasoning** tokens, the TUI may show them in the
footer; they are **not** always billed as a third column (provider-dependent).
Prefer treating reasoning as part of output when the API does not split cost.

## Estimate when counts are missing

```text
estimate_tokens ≈ sum(len(content)/4 + 4) + tool_call overhead
```

`session.EstimateTokens` and some `CountTokens` implementations use character/4
heuristics when the provider has no local tokenizer.

## Model list prices (ship defaults)

Prices are embedded in each provider’s `Models()` table (`InputCost` /
`OutputCost` = USD per **1M** tokens). Examples:

| Model | In $/1M | Out $/1M | Context |
|-------|---------|----------|---------|
| grok-4.5 | 2.00 | 6.00 | 500k |
| gemini-2.5-flash | 0.15 | 0.60 | 1M |
| gemini-2.5-pro | 1.25 | 10.00 | 1M |
| claude-sonnet-4 | 3.00 | 15.00 | 200k |
| gpt-4o-mini | 0.15 | 0.60 | 128k |
| ollama/* | 0 | 0 | local |

**Update policy:** bump when provider public pricing changes; tests lock a few
golden pairs (`internal/provider/cost_test.go`).

## Budget hard-stop

Config:

```yaml
budget:
  max_cost_usd: 5.0    # 0 = unlimited
  warn_at_usd: 2.5     # 0 = 50% of max when max set
```

When `totalCost >= max_cost_usd`, new agent/chat submits are blocked (`/budget`).

## Accuracy limits

- Cached / batch pricing, prompt caching discounts, and free-tier quotas are **not** modeled  
- Multi-model sessions sum each turn’s estimate independently  
- Always trust the provider dashboard for real spend  

## Related

- Footer cost sparkline · `/cost` · `/budget`  
- `provider.CostForModel` · `CostBreakdown`  
