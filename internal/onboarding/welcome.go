package onboarding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeforge/tui/internal/config"
)

// StatusCard is the single first-run / welcome block (Q5.1).
// Designed for ~80 columns: brand + one status panel, no message flood.
func StatusCard(cfg *config.Config, activeName, activeModel string, healthy bool) string {
	if !healthy {
		return "⚡ CodeForge — No AI provider configured.\nPress Ctrl+K for Command Palette to setup."
	}

	var meta []string
	if activeModel != "" {
		meta = append(meta, activeModel)
	}
	if src, _ := KeySource(activeName); src != "" {
		meta = append(meta, src)
	}

	return fmt.Sprintf("🚀 CodeForge Ready · %s\n%s\n\n💡 Tip: Press Ctrl+K for Palette, @ to attach files, or just type a request.",
		activeName, strings.Join(meta, " · "))
}

// WelcomeMessage is the TUI first-run system message (Q5.1: single status card).
// Alias of StatusCard for backward compatibility.
func WelcomeMessage(cfg *config.Config, activeName, activeModel string, healthy bool) string {
	return StatusCard(cfg, activeName, activeModel, healthy)
}

// EmptyStateNoKey is shown when chat is empty and no provider validates (Q5.2).
func EmptyStateNoKey() string {
	return `Nothing to send yet — add a key first:

  /setup gemini <AIza…>     free Google AI Studio key
  /setup grok xai-…         xAI Grok
  export GEMINI_API_KEY=…   then restart CodeForge

Or /setup for the full multi-provider guide.`
}

// EmptyStateNoProject is shown when workdir looks empty of code (Q5.2).
func EmptyStateNoProject(workdir string) string {
	base := filepath.Base(workdir)
	if base == "" || base == "." {
		base = "this folder"
	}
	return fmt.Sprintf(`Project "%s" has few or no source files yet.

  Tips:
  · Open CodeForge from a repo root (cd your-project && codeforge)
  · Or attach a file with @ and ask about it
  · /act scaffold a hello world   to start coding`, base)
}

// ProjectLooksEmpty reports whether workdir has almost no code files (Q5.2).
func ProjectLooksEmpty(workdir string) bool {
	if workdir == "" {
		return true
	}
	entries, err := os.ReadDir(workdir)
	if err != nil {
		return true
	}
	codeExt := map[string]bool{
		".go": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
		".py": true, ".rs": true, ".java": true, ".c": true, ".h": true,
		".cpp": true, ".rb": true, ".php": true, ".swift": true, ".kt": true,
		".md": true, ".toml": true, ".yaml": true, ".yml": true, ".json": true,
	}
	n := 0
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if e.IsDir() {
			// common project dirs count as non-empty
			switch name {
			case "src", "cmd", "pkg", "lib", "app", "internal", "tests", "test":
				return false
			}
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if codeExt[ext] {
			n++
			if n >= 2 {
				return false
			}
		}
	}
	return n < 2
}
