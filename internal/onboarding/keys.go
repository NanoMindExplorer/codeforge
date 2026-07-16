package onboarding

import (
	"fmt"
	"os"
	"strings"

	"github.com/codeforge/tui/internal/config"
)

// EnvKeys maps provider name → env var names (first non-empty wins).
var EnvKeys = map[string][]string{
	"grok":   {"XAI_API_KEY", "GROK_API_KEY"},
	"xai":    {"XAI_API_KEY", "GROK_API_KEY"},
	"gemini": {"GEMINI_API_KEY"},
	"claude": {"ANTHROPIC_API_KEY"},
	"openai": {"OPENAI_API_KEY"},
}

// DefaultModels per provider for wizard defaults.
var DefaultModels = map[string]string{
	"grok":   "grok-4.5",
	"gemini": "gemini-2.5-flash",
	"claude": "claude-sonnet-4-20250514",
	"openai": "gpt-4o-mini",
	"ollama": "llama3.2",
}

// DetectProviderFromKey guesses provider from key prefix / shape.
func DetectProviderFromKey(key string) string {
	k := strings.TrimSpace(key)
	switch {
	case strings.HasPrefix(k, "xai-"), strings.HasPrefix(k, "xai_"):
		return "grok"
	case strings.HasPrefix(k, "sk-ant-"):
		return "claude"
	case strings.HasPrefix(k, "AIza"):
		return "gemini"
	case strings.HasPrefix(k, "sk-"):
		return "openai"
	default:
		return ""
	}
}

// HasAnyAPIKey is true if env or config has at least one provider key.
func HasAnyAPIKey() bool {
	for _, envs := range EnvKeys {
		for _, e := range envs {
			if os.Getenv(e) != "" {
				return true
			}
		}
	}
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		return false
	}
	for name, p := range cfg.Providers {
		if name == "ollama" {
			continue
		}
		if strings.TrimSpace(p.APIKey) != "" {
			return true
		}
	}
	return false
}

// KeySource describes where a provider's key comes from.
// source examples: "env:XAI_API_KEY", "config", "missing"
func KeySource(provider string) (source string, present bool) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "xai" {
		provider = "grok"
	}
	if provider == "ollama" {
		return "local", true
	}
	envs := EnvKeys[provider]
	for _, e := range envs {
		if os.Getenv(e) != "" {
			return "env:" + e, true
		}
	}
	cfg, err := config.Load()
	if err == nil && cfg != nil {
		if p, ok := cfg.Providers[provider]; ok && strings.TrimSpace(p.APIKey) != "" {
			return "config", true
		}
		// alias xai under grok
		if provider == "grok" {
			if p, ok := cfg.Providers["xai"]; ok && strings.TrimSpace(p.APIKey) != "" {
				return "config", true
			}
		}
	}
	return "missing", false
}

// FormatKeySources returns a multi-line summary for /provider and /setup.
func FormatKeySources() string {
	order := []string{"grok", "gemini", "claude", "openai", "ollama"}
	var b strings.Builder
	b.WriteString("API key sources:\n")
	for _, name := range order {
		src, ok := KeySource(name)
		mark := "○"
		if ok {
			mark = "✓"
		}
		b.WriteString(fmt.Sprintf("  %s %-8s  %s\n", mark, name, src))
	}
	b.WriteString("\nDefault priority when several keys exist: grok → gemini → config default_provider → claude/openai.\n")
	b.WriteString("Override: /provider <name>  ·  re-run setup: /setup")
	return b.String()
}

// EnvNameForProvider returns the preferred env var name for docs/hints.
func EnvNameForProvider(provider string) string {
	envs := EnvKeys[strings.ToLower(provider)]
	if len(envs) == 0 {
		return ""
	}
	return envs[0]
}
