package provider

import (
    "context"
    "fmt"
    "os"

    "github.com/liushuangls/go-anthropic"
)

type ClaudeProvider struct {
    client *anthropic.Client
    model  string
    apiKey string
}

func NewClaudeProvider(apiKey, defaultModel string) *ClaudeProvider {
    if apiKey == "" {
        apiKey = os.Getenv("ANTHROPIC_API_KEY")
    }
    cp := &ClaudeProvider{
        apiKey: apiKey,
        model:  defaultModel,
    }
    if apiKey != "" {
        cp.client = anthropic.NewClient(apiKey)
    }
    return cp
}

func (p *ClaudeProvider) Name() string { return "claude" }

func (p *ClaudeProvider) Models() []ModelInfo {
    return []ModelInfo{
        {ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", ContextWindow: 200000, InputCost: 3.0, OutputCost: 15.0},
        {ID: "claude-opus-4-0-20250918", Name: "Claude Opus 4", ContextWindow: 200000, InputCost: 15.0, OutputCost: 75.0},
        {ID: "claude-haiku-4-20250414", Name: "Claude Haiku 4", ContextWindow: 200000, InputCost: 0.80, OutputCost: 4.0},
    }
}

func (p *ClaudeProvider) ValidateConfig() error {
    if p.apiKey == "" {
        return fmt.Errorf("ANTHROPIC_API_KEY not set")
    }
    if p.client == nil {
        return fmt.Errorf("anthropic client not initialized")
    }
    return nil
}

func toAnthropicMessages(msgs []Message) []anthropic.Message {
    out := make([]anthropic.Message, 0, len(msgs))
    for _, m := range msgs {
        switch m.Role {
        case "assistant":
            out = append(out, anthropic.NewAssistantTextMessage(m.Content))
        default:
            out = append(out, anthropic.NewUserTextMessage(m.Content))
        }
    }
    return out
}

func (p *ClaudeProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
    if err := p.ValidateConfig(); err != nil {
        return nil, err
    }
    model := req.Model
    if model == "" {
        model = p.model
    }
    maxTokens := req.MaxTokens
    if maxTokens == 0 {
        maxTokens = 4096
    }
    anthropicReq := anthropic.MessagesRequest{
        Model:     model,
        Messages:  toAnthropicMessages(req.Messages),
        MaxTokens: maxTokens,
    }
    if req.System != "" {
        anthropicReq.System = req.System
    }
    if req.Temperature > 0 {
        anthropicReq.SetTemperature(float32(req.Temperature))
    }
    resp, err := p.client.CreateMessages(ctx, anthropicReq)
    if err != nil {
        return nil, fmt.Errorf("anthropic: %w", err)
    }
    result := &CompletionResponse{
        InputTokens:  resp.Usage.InputTokens,
        OutputTokens: resp.Usage.OutputTokens,
        StopReason:   resp.StopReason,
    }
    for _, content := range resp.Content {
        if content.Type == "text" {
            result.Content += content.Text
        }
    }
    return result, nil
}

func (p *ClaudeProvider) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamToken, error) {
    if err := p.ValidateConfig(); err != nil {
        return nil, err
    }
    model := req.Model
    if model == "" {
        model = p.model
    }
    maxTokens := req.MaxTokens
    if maxTokens == 0 {
        maxTokens = 4096
    }
    anthropicReq := anthropic.MessagesRequest{
        Model:     model,
        Messages:  toAnthropicMessages(req.Messages),
        MaxTokens: maxTokens,
    }
    if req.System != "" {
        anthropicReq.System = req.System
    }
    if req.Temperature > 0 {
        anthropicReq.SetTemperature(float32(req.Temperature))
    }
    out := make(chan StreamToken, 100)
    go func() {
        defer close(out)
        var inputTokens, outputTokens int
        streamReq := anthropic.MessagesStreamRequest{
            MessagesRequest: anthropicReq,
            OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
                if data.Delta.Text != "" {
                    out <- StreamToken{Text: data.Delta.Text}
                }
            },
            OnMessageStart: func(data anthropic.MessagesEventMessageStartData) {
                inputTokens = data.Message.Usage.InputTokens
            },
            OnMessageDelta: func(data anthropic.MessagesEventMessageDeltaData) {
                outputTokens = data.Usage.OutputTokens
            },
            OnMessageStop: func(data anthropic.MessagesEventMessageStopData) {
                out <- StreamToken{
                    Done:         true,
                    InputTokens:  inputTokens,
                    OutputTokens: outputTokens,
                }
            },
            OnError: func(err anthropic.ErrorResponse) {
                out <- StreamToken{
                    Done:  true,
                    Error: fmt.Errorf("stream: %s", err.Error.Message),
                }
            },
        }
        _, err := p.client.CreateMessagesStream(ctx, streamReq)
        if err != nil {
            out <- StreamToken{Done: true, Error: fmt.Errorf("anthropic: %w", err)}
        }
    }()
    return out, nil
}

func (p *ClaudeProvider) CountTokens(messages []Message) int {
    total := 0
    for _, m := range messages {
        total += len(m.Content) / 4
        total += 4
    }
    return total
}
