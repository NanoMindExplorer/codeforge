package onboarding

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codeforge/tui/internal/provider"
)

// WizardOptions control the CLI first-run flow.
type WizardOptions struct {
	In  io.Reader
	Out io.Writer
	// Registry is optional; when set, successful keys are registered immediately.
	Registry *provider.Registry
	// SkipValidation skips live ValidateConfig (tests).
	SkipValidation bool
}

// RunWizard is the interactive first-run setup (O2).
func RunWizard(opt WizardOptions) error {
	in := opt.In
	if in == nil {
		in = os.Stdin
	}
	out := opt.Out
	if out == nil {
		out = os.Stdout
	}
	r := bufio.NewReader(in)

	fmt.Fprintln(out)
	fmt.Fprintln(out, "╔══════════════════════════════════════════════════════╗")
	fmt.Fprintln(out, "║  CodeForge — First Run Setup (≤ 3 min)               ║")
	fmt.Fprintln(out, "╚══════════════════════════════════════════════════════╝")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "① Detected keys:")
	printDetected(out)

	fmt.Fprintln(out)
	fmt.Fprintln(out, "② Choose provider [1] grok  [2] gemini  [3] claude  [4] openai  [5] ollama  [s] skip")
	fmt.Fprint(out, "   > ")
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "s" || line == "skip" || line == "q" {
		_ = MarkSkipped()
		fmt.Fprintln(out, "   Skipped. Run /setup later or set XAI_API_KEY / GEMINI_API_KEY.")
		return nil
	}

	name := mapChoice(line)
	if name == "" {
		// allow typing provider name
		name = strings.ToLower(line)
		if name != "grok" && name != "gemini" && name != "claude" && name != "openai" && name != "ollama" {
			// try paste-as-key flow
			if det := DetectProviderFromKey(line); det != "" {
				return finishWithKey(opt, out, r, det, line)
			}
			fmt.Fprintln(out, "   Unknown choice — try /setup in the TUI later.")
			_ = MarkSkipped()
			return nil
		}
	}

	if name == "ollama" {
		if opt.Registry != nil {
			p, err := ApplyKey(opt.Registry, "ollama", "", "")
			if err != nil {
				fmt.Fprintf(out, "   ⚠ Ollama: %v\n   Start `ollama serve` and pull a model, then /setup.\n", err)
				return nil
			}
			fmt.Fprintf(out, "   ✓ Ollama · model %s\n", p.Model())
		} else {
			_ = MarkCompleted("ollama", DefaultModels["ollama"])
			fmt.Fprintln(out, "   ✓ Ollama selected (ensure `ollama serve` is running)")
		}
		printDone(out)
		return nil
	}

	fmt.Fprintf(out, "\n③ Paste API key for %s (%s), or Enter to skip:\n", name, EnvNameForProvider(name))
	fmt.Fprint(out, "   key> ")
	key, _ := r.ReadString('\n')
	key = strings.TrimSpace(key)
	if key == "" {
		fmt.Fprintln(out, "   No key entered. You can /setup later.")
		_ = MarkSkipped()
		return nil
	}
	// re-detect if user pasted wrong provider's key
	if det := DetectProviderFromKey(key); det != "" && det != name {
		fmt.Fprintf(out, "   (key looks like %s — using that)\n", det)
		name = det
	}
	return finishWithKey(opt, out, r, name, key)
}

func finishWithKey(opt WizardOptions, out io.Writer, r *bufio.Reader, name, key string) error {
	model := DefaultModels[name]
	fmt.Fprintf(out, "\n④ Default model [%s] (Enter to keep, or type id):\n", model)
	fmt.Fprint(out, "   model> ")
	mline, _ := r.ReadString('\n')
	mline = strings.TrimSpace(mline)
	if mline != "" {
		model = mline
	}

	reg := opt.Registry
	if reg == nil {
		reg = provider.NewRegistry()
	}
	p, err := ApplyKey(reg, name, key, model)
	if err != nil {
		fmt.Fprintf(out, "   ⚠ %v\n   Key not saved. Fix and run /setup.\n", err)
		return nil
	}
	if !opt.SkipValidation {
		if err := p.ValidateConfig(); err != nil {
			fmt.Fprintf(out, "   ⚠ validate: %v\n", err)
			return nil
		}
	}
	fmt.Fprintf(out, "   ✓ %s ready · model %s\n", name, p.Model())
	fmt.Fprintf(out, "   key source: %s\n", mustSource(name))
	printDone(out)
	return nil
}

func printDetected(out io.Writer) {
	for _, name := range []string{"grok", "gemini", "claude", "openai"} {
		src, ok := KeySource(name)
		if ok {
			fmt.Fprintf(out, "   ✓ %-8s  %s\n", name, src)
		} else {
			fmt.Fprintf(out, "   ○ %-8s  missing\n", name)
		}
	}
}

func mapChoice(line string) string {
	switch line {
	case "1", "g", "grok", "xai":
		return "grok"
	case "2", "gemini", "gem":
		return "gemini"
	case "3", "claude", "anthropic":
		return "claude"
	case "4", "openai", "oai":
		return "openai"
	case "5", "ollama", "local":
		return "ollama"
	default:
		return ""
	}
}

func mustSource(name string) string {
	src, _ := KeySource(name)
	return src
}

func printDone(out io.Writer) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "You're set. Type a question or /act <task>.")
	fmt.Fprintln(out, "  Shift+Tab = BUILD → DESIGN → YOLO · /setup reopens this flow · /provider shows key source")
	fmt.Fprintln(out)
}
