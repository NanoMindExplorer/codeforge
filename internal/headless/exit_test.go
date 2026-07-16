package headless

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// O7: agent without any provider key returns structured no_provider.
func TestNoProviderJSON(t *testing.T) {
	// Clear provider envs for this process
	for _, e := range []string{
		"XAI_API_KEY", "GROK_API_KEY", "GEMINI_API_KEY",
		"ANTHROPIC_API_KEY", "OPENAI_API_KEY",
	} {
		t.Setenv(e, "")
	}
	// Isolate config so no saved keys
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", home+"/.config")

	var buf bytes.Buffer
	res, err := Run(Options{
		Task:  "hello",
		JSON:  true,
		Quiet: true,
	}, &buf)
	if err == nil && res.OK {
		t.Fatal("expected failure without provider")
	}
	if res.Code != "no_provider" {
		t.Fatalf("code=%q err=%v out=%s", res.Code, err, buf.String())
	}
	if res.Hint == "" {
		t.Fatal("expected hint")
	}
	var decoded map[string]any
	if jerr := json.Unmarshal(buf.Bytes(), &decoded); jerr != nil {
		t.Fatalf("json: %v body=%s", jerr, buf.String())
	}
	if decoded["ok"] != false {
		t.Fatalf("%v", decoded)
	}
	if !strings.Contains(buf.String(), "no_provider") {
		t.Fatalf("body missing no_provider: %s", buf.String())
	}
	_ = os.Stderr
}
