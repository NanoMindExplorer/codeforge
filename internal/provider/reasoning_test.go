package provider

import (
	"encoding/json"
	"testing"
)

func TestWantsReasoning(t *testing.T) {
	cases := []struct {
		model string
		flag  string
		want  bool
	}{
		{"grok-4.5", "", true},
		{"gpt-4o-mini", "", false},
		{"gpt-4o", "on", true},
		{"o3-mini", "", true},
		{"gemini-2.5-flash", "", true},
		{"claude-sonnet-4-20250514", "", true},
		{"grok-4.5", "off", false},
	}
	for _, c := range cases {
		r := CompletionRequest{Reasoning: c.flag}
		if got := r.WantsReasoning(c.model); got != c.want {
			t.Fatalf("%s flag=%q got %v want %v", c.model, c.flag, got, c.want)
		}
	}
}

func TestParseOpenAIReasoningDelta(t *testing.T) {
	// ensure JSON shapes we parse are valid
	raw := `{"choices":[{"delta":{"reasoning_content":"step 1","content":""}}]}`
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content          string `json:"content"`
				ReasoningContent string `json:"reasoning_content"`
				Reasoning        string `json:"reasoning"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(raw), &chunk); err != nil {
		t.Fatal(err)
	}
	if chunk.Choices[0].Delta.ReasoningContent != "step 1" {
		t.Fatal(chunk)
	}
	if firstNonEmpty(chunk.Choices[0].Delta.ReasoningContent, chunk.Choices[0].Delta.Reasoning) != "step 1" {
		t.Fatal("firstNonEmpty")
	}
}

func TestGeminiThoughtPart(t *testing.T) {
	raw := `{"candidates":[{"content":{"parts":[{"text":"thinking hard","thought":true},{"text":"answer"}]}}]}`
	var g geminiResponse
	if err := json.Unmarshal([]byte(raw), &g); err != nil {
		t.Fatal(err)
	}
	var reason, content string
	for _, p := range g.Candidates[0].Content.Parts {
		if p.Thought {
			reason += p.Text
		} else {
			content += p.Text
		}
	}
	if reason != "thinking hard" || content != "answer" {
		t.Fatalf("%q %q", reason, content)
	}
}
