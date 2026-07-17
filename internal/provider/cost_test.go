package provider

import (
	"math"
	"testing"
)

func TestCostForModelGeminiPro(t *testing.T) {
	p := NewGeminiProvider("fake", "gemini-2.5-pro")
	cost := CostForModel(p, "gemini-2.5-pro", 1_000_000, 1_000_000)
	if cost < 10 {
		t.Fatalf("expected paid pro cost, got %v", cost)
	}
}

func TestCostForModelClaude(t *testing.T) {
	p := NewClaudeProvider("fake", "claude-sonnet-4-20250514")
	cost := CostForModel(p, "claude-sonnet-4-20250514", 1_000_000, 1_000_000)
	// 3 + 15 = 18
	if cost < 17 || cost > 19 {
		t.Fatalf("cost=%v", cost)
	}
}

func TestCostDetailFormula(t *testing.T) {
	// Q7.5 golden formula: cost = in*$in/1e6 + out*$out/1e6
	p := NewGrokProvider("fake", "grok-4.5")
	d := CostDetail(p, "grok-4.5", 2_000_000, 500_000)
	// 2*2.0 + 0.5*6.0 = 4 + 3 = 7
	want := 7.0
	if math.Abs(d.TotalUSD-want) > 0.001 {
		t.Fatalf("got %v want %v (detail=%+v)", d.TotalUSD, want, d)
	}
	if d.InputPer1M != 2.0 || d.OutputPer1M != 6.0 {
		t.Fatalf("%+v", d)
	}
}

func TestCostZeroTokens(t *testing.T) {
	p := NewOpenAIProvider("fake", "gpt-4o-mini")
	if CostForModel(p, "gpt-4o-mini", 0, 0) != 0 {
		t.Fatal("zero tokens")
	}
}

func TestCostOllamaFree(t *testing.T) {
	p := NewOllamaProvider("")
	// even with tokens, local is $0
	if CostForModel(p, "llama3.2", 1_000_000, 1_000_000) != 0 {
		t.Fatal("ollama should be free")
	}
}

func TestSetModel(t *testing.T) {
	p := NewGeminiProvider("x", "gemini-2.5-flash")
	if err := p.SetModel("gemini-2.5-pro"); err != nil {
		t.Fatal(err)
	}
	if p.Model() != "gemini-2.5-pro" {
		t.Fatal(p.Model())
	}
}
