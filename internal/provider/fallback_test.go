package provider

import (
	"context"
	"errors"
	"testing"
)

type mockProv struct {
	calls int
	fail  bool
}

func (m *mockProv) Name() string              { return "mock" }
func (m *mockProv) Models() []ModelInfo       { return nil }
func (m *mockProv) Model() string             { return "mock-think" }
func (m *mockProv) SetModel(string) error     { return nil }
func (m *mockProv) CountTokens([]Message) int { return 0 }
func (m *mockProv) ValidateConfig() error     { return nil }
func (m *mockProv) Stream(context.Context, CompletionRequest) (<-chan StreamToken, error) {
	return nil, errors.New("no stream")
}
func (m *mockProv) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	m.calls++
	if m.fail && req.Reasoning != "off" {
		return nil, &ProviderError{Code: ErrUnsupported, Message: "no thinking", Hint: "off"}
	}
	return &CompletionResponse{Content: "ok"}, nil
}

func TestCompleteRetryingReasoning(t *testing.T) {
	m := &mockProv{fail: true}
	resp, retried, err := CompleteRetryingReasoning(context.Background(), m, CompletionRequest{
		Reasoning: "on",
		Messages:  []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !retried || resp.Content != "ok" || m.calls != 2 {
		t.Fatalf("retried=%v calls=%d resp=%+v", retried, m.calls, resp)
	}
}
