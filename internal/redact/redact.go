// Package secrets redacts credentials before they reach the model or logs.
package redact

import (
	"regexp"
	"strings"
)

var patterns = []*regexp.Regexp{
	// AWS
	regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16})`),
	regexp.MustCompile(`(?i)(aws_secret_access_key\s*[=:]\s*)\S+`),
	// Generic API keys / tokens
	regexp.MustCompile(`(?i)((?:api[_-]?key|apikey|secret|token|password|passwd|auth|bearer|credential)["'\s:=]+)([^\s"'\\]{8,})`),
	// GitHub / Slack / Google-ish
	regexp.MustCompile(`\b(ghp_[A-Za-z0-9_]{20,})\b`),
	regexp.MustCompile(`\b(gho_[A-Za-z0-9_]{20,})\b`),
	regexp.MustCompile(`\b(github_pat_[A-Za-z0-9_]{20,})\b`),
	regexp.MustCompile(`\b(sk-[A-Za-z0-9_\-]{20,})\b`),
	regexp.MustCompile(`\b(xox[baprs]-[A-Za-z0-9\-]{10,})\b`),
	regexp.MustCompile(`\b(AIza[0-9A-Za-z_\-]{20,})\b`),
	// Q8.4 — expanded provider / SaaS tokens
	regexp.MustCompile(`\b(xai-[A-Za-z0-9_\-]{20,})\b`),     // xAI
	regexp.MustCompile(`\b(sk-ant-[A-Za-z0-9_\-]{20,})\b`),  // Anthropic
	regexp.MustCompile(`\b(sk-proj-[A-Za-z0-9_\-]{20,})\b`), // OpenAI project keys
	regexp.MustCompile(`\b(hf_[A-Za-z0-9]{20,})\b`),         // Hugging Face
	regexp.MustCompile(`\b(hf_org_[A-Za-z0-9]{20,})\b`),
	regexp.MustCompile(`\b(npm_[A-Za-z0-9]{20,})\b`),      // npm
	regexp.MustCompile(`\b(pypi-[A-Za-z0-9_\-]{20,})\b`),  // PyPI
	regexp.MustCompile(`\b(glpat-[A-Za-z0-9_\-]{20,})\b`), // GitLab PAT
	regexp.MustCompile(`\b(ghu_[A-Za-z0-9_]{20,})\b`),     // GitHub user-to-server
	regexp.MustCompile(`\b(ghs_[A-Za-z0-9_]{20,})\b`),     // GitHub server-to-server
	regexp.MustCompile(`\b(sk_live_[A-Za-z0-9]{20,})\b`),  // Stripe live
	regexp.MustCompile(`\b(sk_test_[A-Za-z0-9]{20,})\b`),  // Stripe test
	regexp.MustCompile(`\b(rk_live_[A-Za-z0-9]{20,})\b`),
	regexp.MustCompile(`\b(xoxe\.[A-Za-z0-9\-]{20,})\b`),  // Slack enterprise
	regexp.MustCompile(`\b(SG\.[A-Za-z0-9_\-\.]{20,})\b`), // SendGrid
	regexp.MustCompile(`\b(whsec_[A-Za-z0-9_\-]{16,})\b`), // Stripe webhook
	regexp.MustCompile(`\b(dop_v1_[a-f0-9]{64})\b`),       // DigitalOcean
	regexp.MustCompile(`\b(oai-?[A-Za-z0-9_\-]{20,})\b`),  // misc OpenAI-ish
	// Private keys
	regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----[\s\S]*?-----END (?:RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----`),
	// JWT-ish
	regexp.MustCompile(`\beyJ[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]{10,}\b`),
}

// SensitiveName reports filenames that should not be fully sent to models.
func SensitiveName(name string) bool {
	n := strings.ToLower(name)
	base := n
	if i := strings.LastIndex(n, "/"); i >= 0 {
		base = n[i+1:]
	}
	switch {
	case base == ".env" || strings.HasPrefix(base, ".env."):
		return true
	case strings.HasSuffix(base, ".pem"), strings.HasSuffix(base, ".key"), strings.HasSuffix(base, ".p12"), strings.HasSuffix(base, ".pfx"):
		return true
	case base == "id_rsa", base == "id_ed25519", base == "id_ecdsa",
		base == "credentials.json", base == "service-account.json",
		base == "secrets.yaml", base == "secrets.yml", base == "secrets.toml":
		return true
	case base == "netrc" || base == ".netrc" || base == ".npmrc" || base == ".pypirc":
		return true
	case strings.Contains(base, "secret") && (strings.HasSuffix(base, ".yml") || strings.HasSuffix(base, ".yaml") || strings.HasSuffix(base, ".json")):
		return true
	}
	return false
}

// Redact replaces known secret patterns with placeholders.
func Redact(s string) string {
	if s == "" {
		return s
	}
	out := s
	for _, re := range patterns {
		out = re.ReplaceAllStringFunc(out, func(m string) string {
			// Keep short prefix key labels when group-captured style
			if strings.Contains(strings.ToLower(m), "key") || strings.Contains(strings.ToLower(m), "token") ||
				strings.Contains(strings.ToLower(m), "secret") || strings.Contains(strings.ToLower(m), "password") ||
				strings.Contains(strings.ToLower(m), "bearer") {
				// leave left side if "key=VALUE"
				if i := strings.IndexAny(m, "=:"); i > 0 && i < len(m)-1 {
					return m[:i+1] + " [REDACTED]"
				}
			}
			return "[REDACTED]"
		})
	}
	return out
}

// RedactFile returns content safe for the model; for sensitive names, returns a stub.
func RedactFile(name, content string) (string, bool) {
	if SensitiveName(name) {
		return "[REDACTED: sensitive file " + name + " — contents not sent to the model]", true
	}
	return Redact(content), false
}
