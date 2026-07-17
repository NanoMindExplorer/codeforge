package redact

import "testing"

func TestRedactAPIKey(t *testing.T) {
	in := `export GEMINI_API_KEY=AIzaSyDummyKeyValue1234567890abcd`
	out := Redact(in)
	if out == in {
		t.Fatal("expected redaction")
	}
	if !contains(out, "REDACTED") {
		t.Fatalf("%q", out)
	}
}

func TestSensitiveFile(t *testing.T) {
	out, blocked := RedactFile(".env", "SECRET=1")
	if !blocked {
		t.Fatal("expected blocked")
	}
	if !contains(out, "REDACTED") {
		t.Fatal(out)
	}
}

func TestRedactGitHubPAT(t *testing.T) {
	in := "token ghp_abcdefghijklmnopqrstuvwxyz012345"
	out := Redact(in)
	if contains(out, "ghp_abcd") {
		t.Fatalf("pat leaked: %q", out)
	}
}

func TestRedactXAIAndHuggingFace(t *testing.T) {
	// Q8.4 — build token samples at runtime so secret scanners do not flag literals.
	suffix := "abcdefghijklmnopqrstuvwxyz0123456789"
	cases := []string{
		"export XAI_API_KEY=xai-" + suffix,
		"Authorization: Bearer hf_" + suffix[:22],
		"ANTHROPIC_API_KEY=sk-ant-api03-" + suffix,
		"OPENAI_API_KEY=sk-proj-" + suffix,
		"npm_token=npm_" + suffix[:22],
		// stripe: split prefix so push-protection does not match sk_live_…
		"stripe " + "sk" + "_live_" + suffix[:26],
	}
	for _, in := range cases {
		out := Redact(in)
		if out == in {
			t.Fatalf("expected redaction for %q", in)
		}
		if !contains(out, "REDACTED") {
			t.Fatalf("no REDACTED in %q → %q", in, out)
		}
	}
}

func TestSensitiveNetrcAndNpmrc(t *testing.T) {
	for _, name := range []string{".npmrc", ".netrc", "id_ed25519", "secrets.yaml"} {
		_, blocked := RedactFile(name, "secret=1")
		if !blocked {
			t.Fatal(name)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && (stringIndex(s, sub) >= 0)))
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
