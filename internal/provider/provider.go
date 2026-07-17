package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Role constants for provider.Message.Role.
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

type Message struct {
	Role    string `json:"role"` // "user" | "assistant" | "tool"
	Content string `json:"content"`

	// ToolCalls is set on assistant messages that invoke one or more tools.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// The following three fields are set on role="tool" messages, which
	// carry the result of executing a single tool call back to the model.
	ToolCallID string `json:"tool_call_id,omitempty"` // references ToolCall.ID
	ToolName   string `json:"tool_name,omitempty"`
	IsError    bool   `json:"is_error,omitempty"`
}

type ToolCall struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Input string `json:"input"` // raw JSON arguments object
}

type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"input_schema"`
}

type CompletionRequest struct {
	Messages    []Message        `json:"messages"`
	Model       string           `json:"model"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	System      string           `json:"system,omitempty"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	// Reasoning enables native thinking/reasoning streams when the provider supports it.
	// Empty = auto (on for known reasoning models). "off" | "low" | "medium" | "high" | "on".
	Reasoning string `json:"reasoning,omitempty"`
}

type CompletionResponse struct {
	Content string `json:"content"`
	// Reasoning is native chain-of-thought / thinking text (not shown as assistant content).
	Reasoning    string     `json:"reasoning,omitempty"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	InputTokens  int        `json:"input_tokens"`
	OutputTokens int        `json:"output_tokens"`
	// ReasoningTokens when the API reports a separate count.
	ReasoningTokens int    `json:"reasoning_tokens,omitempty"`
	StopReason      string `json:"stop_reason"`
}

type StreamToken struct {
	Text string `json:"text,omitempty"`
	// Reasoning is a native thinking/reasoning delta (Grok/OpenAI/Gemini/Claude).
	Reasoning       string    `json:"reasoning,omitempty"`
	ToolCall        *ToolCall `json:"tool_call,omitempty"`
	Done            bool      `json:"done"`
	Error           error     `json:"-"`
	InputTokens     int       `json:"input_tokens,omitempty"`
	OutputTokens    int       `json:"output_tokens,omitempty"`
	ReasoningTokens int       `json:"reasoning_tokens,omitempty"`
}

// WantsReasoning reports whether the request should ask the model for thinking tokens.
func (r CompletionRequest) WantsReasoning(modelID string) bool {
	switch strings.ToLower(strings.TrimSpace(r.Reasoning)) {
	case "off", "none", "false", "0":
		return false
	case "on", "true", "1", "low", "medium", "high", "max":
		return true
	}
	// auto: enable for known reasoning-capable model families
	m := strings.ToLower(modelID)
	for _, p := range []string{"grok", "o1", "o3", "o4", "reason", "think", "r1", "gemini-2.5", "gemini-3", "claude"} {
		if strings.Contains(m, p) {
			return true
		}
	}
	// env override
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("CODEFORGE_REASONING"))); v != "" {
		return v != "0" && v != "off" && v != "false"
	}
	return false
}

type Provider interface {
	Name() string
	Models() []ModelInfo
	// Model returns the currently selected model ID.
	Model() string
	// SetModel switches the active model. Returns error if ID is unknown.
	SetModel(id string) error
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Stream(ctx context.Context, req CompletionRequest) (<-chan StreamToken, error)
	CountTokens(messages []Message) int
	ValidateConfig() error
}

// CostBreakdown documents token→USD math (Q7.5). See docs/COST.md.
type CostBreakdown struct {
	ModelID      string
	InputTokens  int
	OutputTokens int
	// InputPer1M / OutputPer1M are USD per 1M tokens from ModelInfo.
	InputPer1M  float64
	OutputPer1M float64
	// TotalUSD = in*InputPer1M/1e6 + out*OutputPer1M/1e6
	TotalUSD float64
}

// CostForModel returns USD cost for token counts using ModelInfo pricing.
// Formula: (in * input_$/1M + out * output_$/1M) / 1_000_000
func CostForModel(p Provider, modelID string, in, out int) float64 {
	return CostDetail(p, modelID, in, out).TotalUSD
}

// CostDetail returns the full breakdown for audits and tests (Q7.5).
func CostDetail(p Provider, modelID string, in, out int) CostBreakdown {
	b := CostBreakdown{InputTokens: in, OutputTokens: out}
	if p == nil {
		return b
	}
	if modelID == "" {
		modelID = p.Model()
	}
	b.ModelID = modelID
	var chosen *ModelInfo
	for i := range p.Models() {
		m := p.Models()[i]
		if m.ID == modelID {
			chosen = &m
			break
		}
	}
	if chosen == nil {
		models := p.Models()
		if len(models) == 0 {
			return b
		}
		chosen = &models[0]
		b.ModelID = chosen.ID
	}
	b.InputPer1M = chosen.InputCost
	b.OutputPer1M = chosen.OutputCost
	b.TotalUSD = float64(in)*chosen.InputCost/1_000_000 + float64(out)*chosen.OutputCost/1_000_000
	return b
}

type ModelInfo struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ContextWindow int     `json:"context_window"`
	InputCost     float64 `json:"input_cost_per_1m"`
	OutputCost    float64 `json:"output_cost_per_1m"`
}

type Registry struct {
	providers map[string]Provider
	current   string
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

func (r *Registry) Register(p Provider) error {
	if p == nil {
		return errors.New("provider is nil")
	}
	r.providers[p.Name()] = p
	if r.current == "" {
		r.current = p.Name()
	}
	return nil
}

func (r *Registry) Get(name string) (Provider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", name)
	}
	return p, nil
}

func (r *Registry) Current() (Provider, error) {
	if r.current == "" {
		return nil, errors.New("no provider registered")
	}
	return r.Get(r.current)
}

func (r *Registry) Switch(name string) error {
	if _, ok := r.providers[name]; !ok {
		return fmt.Errorf("provider %q not registered", name)
	}
	r.current = name
	return nil
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

func (r *Registry) CurrentName() string {
	return r.current
}
