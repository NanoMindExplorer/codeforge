package provider

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClassifyAuth(t *testing.T) {
	pe := Classify(nil, 401, `{"error":"invalid api key"}`, "grok")
	if pe.Code != ErrAuth {
		t.Fatal(pe.Code)
	}
	um := pe.UserMessage()
	if !strings.Contains(um, "API key") {
		t.Fatal(um)
	}
	if strings.Contains(um, `{"error"`) {
		t.Fatal("raw json leaked:", um)
	}
}

func TestClassifyRateLimit(t *testing.T) {
	pe := Classify(nil, 429, `{"error":{"message":"rate_limit exceeded","retry_after":20}}`, "openai")
	if pe.Code != ErrRateLimit || !pe.Retry {
		t.Fatal(pe)
	}
	if pe.RetryAfter < 19*time.Second {
		t.Log("retry_after parse:", pe.RetryAfter)
	}
	um := pe.UserMessage()
	if strings.Contains(um, `{"error"`) {
		t.Fatal("json in user msg", um)
	}
}

func TestClassifyRateLimitHeader(t *testing.T) {
	hdr := http.Header{}
	hdr.Set("Retry-After", "15")
	err := HTTPErrorHeaders("grok", 429, []byte("too many requests"), hdr, nil)
	pe, ok := AsProviderError(err)
	if !ok || pe.Code != ErrRateLimit {
		t.Fatal(pe)
	}
	if pe.RetryAfter < 14*time.Second {
		t.Fatalf("retry after: %v", pe.RetryAfter)
	}
}

func TestClassifyReasoningUnsupported(t *testing.T) {
	pe := Classify(nil, 400, `unknown field include_reasoning`, "grok")
	if pe.Code != ErrUnsupported {
		t.Fatalf("%s %s", pe.Code, pe.Message)
	}
	um := pe.UserMessage()
	if !strings.Contains(um, "Reasoning") && !strings.Contains(um, "thinking") {
		t.Fatal(um)
	}
}

func TestClassifyContext(t *testing.T) {
	pe := Classify(nil, 400, `maximum context length exceeded`, "openai")
	if pe.Code != ErrContext {
		t.Fatal(pe.Code)
	}
}

func TestClassifyNetwork(t *testing.T) {
	pe := Classify(fmt.Errorf("dial tcp: connection refused"), 0, "", "ollama")
	if pe.Code != ErrNetwork {
		t.Fatal(pe.Code)
	}
}

func TestFormatUserErrorNoJSON(t *testing.T) {
	body := `{"error":{"message":"rate limit exceeded","type":"rate_limit_error","code":"rate_limit"}}`
	err := HTTPError("gemini", http.StatusTooManyRequests, []byte(body), nil)
	s := FormatUserError(err)
	if strings.Contains(s, `{"error"`) {
		t.Fatal("full json dumped:", s)
	}
	if !strings.Contains(s, "Rate limited") && !strings.Contains(strings.ToLower(s), "rate") {
		t.Fatal(s)
	}
}

func TestFormatUserErrorNoStack(t *testing.T) {
	stack := "something failed\ngoroutine 1 [running]:\nmain.foo()\n\t/tmp/x.go:10\n"
	s := FormatUserError(errors.New(stack))
	if strings.Contains(s, "goroutine") {
		t.Fatal(s)
	}
}

func TestAsProviderError(t *testing.T) {
	err := &ProviderError{Code: ErrModel, Message: "nope", Hint: "pick another"}
	pe, ok := AsProviderError(err)
	if !ok || pe.Code != ErrModel {
		t.Fatal(ok, pe)
	}
	pe2, ok := AsProviderError(errors.New("invalid api key"))
	if !ok || pe2.Code != ErrAuth {
		t.Fatal(pe2)
	}
}

func TestAuthErrorValidate(t *testing.T) {
	err := AuthError("gemini", "GEMINI_API_KEY not set")
	s := FormatUserError(err)
	if !strings.Contains(s, "GEMINI") && !strings.Contains(s, "API key") {
		t.Fatal(s)
	}
}

func TestShortToast(t *testing.T) {
	pe := Classify(nil, 429, `rate limit`, "grok")
	sh := pe.Short()
	if strings.Contains(sh, "\n") {
		t.Fatal(sh)
	}
}

func TestRedactInRaw(t *testing.T) {
	pe := Classify(nil, 401, `{"error":"invalid api key sk-abcdefghijklmnopqrstuvwxyz1234"}`, "openai")
	if strings.Contains(pe.Raw, "sk-abcdefghijklmnop") {
		t.Fatal("secret in raw:", pe.Raw)
	}
}

func TestProviderErrorLog(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEFORGE_PROVIDER_ERROR_LOG", "1")
	p := filepath.Join(home, ".codeforge", "logs", "provider-error.jsonl")
	logMu.Lock()
	logPath = p
	logMu.Unlock()
	pe := &ProviderError{Code: ErrAuth, Message: "test log", Provider: "grok", Status: 401}
	logProviderError(pe)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "auth") {
		t.Fatal(string(b))
	}
}

func TestExtractAPIMessage(t *testing.T) {
	m := extractAPIMessage(`{"error":{"message":"hello world","type":"x"}}`)
	if m != "hello world" {
		t.Fatal(m)
	}
}
