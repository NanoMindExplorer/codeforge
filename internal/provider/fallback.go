package provider

import (
	"context"
	"strings"
)

// shouldRetryWithoutReasoning is true for unsupported-reasoning failures when
// the request still asked for thinking tokens.
func shouldRetryWithoutReasoning(err error, req CompletionRequest, model string) bool {
	if err == nil {
		return false
	}
	if !req.WantsReasoning(model) {
		return false
	}
	pe, ok := AsProviderError(err)
	if !ok || pe == nil {
		s := strings.ToLower(err.Error())
		return containsAny(s, "include_reasoning", "reasoning_effort", "budget_tokens", "thinking")
	}
	return pe.Code == ErrUnsupported
}

// CompleteRetryingReasoning calls Complete; on unsupported reasoning, retries once with Reasoning=off.
// retried is true when the second attempt was used (caller may surface a one-line notice).
func CompleteRetryingReasoning(ctx context.Context, p Provider, req CompletionRequest) (resp *CompletionResponse, retried bool, err error) {
	if p == nil {
		return nil, false, &ProviderError{Code: ErrAuth, Message: "No AI provider configured", Hint: "Run /setup or set an API key", Retry: false}
	}
	model := req.Model
	if model == "" {
		model = p.Model()
	}
	resp, err = p.Complete(ctx, req)
	if err == nil {
		return resp, false, nil
	}
	if !shouldRetryWithoutReasoning(err, req, model) {
		return nil, false, err
	}
	req2 := req
	req2.Reasoning = "off"
	resp2, err2 := p.Complete(ctx, req2)
	if err2 != nil {
		return nil, true, err2
	}
	return resp2, true, nil
}

// StreamRetryingReasoning opens a stream; if the first token is an unsupported-reasoning
// error, re-opens once with Reasoning=off.
func StreamRetryingReasoning(ctx context.Context, p Provider, req CompletionRequest) (<-chan StreamToken, bool, error) {
	if p == nil {
		return nil, false, &ProviderError{Code: ErrAuth, Message: "No AI provider configured", Hint: "Run /setup", Retry: false}
	}
	model := req.Model
	if model == "" {
		model = p.Model()
	}
	ch, err := p.Stream(ctx, req)
	if err != nil {
		if shouldRetryWithoutReasoning(err, req, model) {
			req2 := req
			req2.Reasoning = "off"
			ch2, err2 := p.Stream(ctx, req2)
			return ch2, true, err2
		}
		return nil, false, err
	}
	// Peek first token for classification; re-stream if needed.
	out := make(chan StreamToken, 64)
	go func() {
		defer close(out)
		first, ok := <-ch
		if !ok {
			return
		}
		if first.Error != nil && shouldRetryWithoutReasoning(first.Error, req, model) {
			req2 := req
			req2.Reasoning = "off"
			ch2, err2 := p.Stream(ctx, req2)
			if err2 != nil {
				out <- StreamToken{Done: true, Error: err2}
				return
			}
			// signal retry via a zero-text token with no error — consumer sees normal stream;
			// we push a synthetic empty reasoning note as first done=false text is awkward.
			// Emit one info-bearing error-free token with Reasoning field used as notice.
			out <- StreamToken{Reasoning: "↻ reasoning not supported — continued without thinking"}
			for tok := range ch2 {
				out <- tok
			}
			return
		}
		out <- first
		for tok := range ch {
			out <- tok
		}
	}()
	return out, false, nil
}
