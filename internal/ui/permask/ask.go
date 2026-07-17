// Package permask is the interactive permission ask modal (y/n/always).
package permask

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/codeforge/tui/internal/theme"
)

// Risk level for the permission badge (Q8.2).
type Risk int

const (
	RiskLow Risk = iota
	RiskMedium
	RiskHigh
)

func (r Risk) String() string {
	switch r {
	case RiskHigh:
		return "HIGH"
	case RiskMedium:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// Model is the y/n/always prompt overlay.
type Model struct {
	Active    bool
	Tool      string
	Input     string
	Reason    string
	Dangerous bool
	Risk      Risk
	// Result
	Done   bool
	Allow  bool
	Always bool // remember for session/project
}

func New() Model { return Model{} }

func (m *Model) Open(tool, input, reason string, dangerous bool) {
	m.Active = true
	m.Done = false
	m.Allow = false
	m.Always = false
	m.Tool = tool
	m.Input = input
	m.Reason = reason
	m.Dangerous = dangerous
	m.Risk = ClassifyRisk(tool, input, dangerous)
}

func (m *Model) Close() {
	m.Active = false
	m.Done = false
}

func (m *Model) Yes(always bool) {
	if always && m.Dangerous {
		always = false // never remember dangerous
	}
	m.Done = true
	m.Active = false
	m.Allow = true
	m.Always = always
}

func (m *Model) No(always bool) {
	if always && m.Dangerous {
		always = false
	}
	m.Done = true
	m.Active = false
	m.Allow = false
	m.Always = always
}

// ClassifyRisk maps tool+input to a badge (Q8.2).
func ClassifyRisk(tool, input string, dangerous bool) Risk {
	if dangerous {
		return RiskHigh
	}
	tl := strings.ToLower(tool)
	in := strings.ToLower(input)
	switch {
	case tl == "run_command" || tl == "run_terminal_command":
		// Shell always elevated
		if strings.Contains(in, "sudo") || strings.Contains(in, "curl ") ||
			strings.Contains(in, "wget ") || strings.Contains(in, "dd ") ||
			strings.Contains(in, "mkfs") || strings.Contains(in, "chmod 777") {
			return RiskHigh
		}
		return RiskMedium
	case tl == "write_file" || tl == "search_replace" || tl == "apply_patch":
		return RiskMedium
	case strings.Contains(tl, "delete") || strings.Contains(tl, "rm"):
		return RiskHigh
	default:
		return RiskLow
	}
}

// FormatCommand extracts a human-readable full command/payload for display.
// For JSON tool input with a "command" field, shows the raw command unescaped.
// Does not silently truncate mid-command under 4k runes (Q8.2 full command).
func FormatCommand(tool, input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return "(no input)"
	}
	// Try JSON { "command": "..." }
	var obj map[string]any
	if err := json.Unmarshal([]byte(input), &obj); err == nil {
		if cmd, ok := obj["command"].(string); ok && strings.TrimSpace(cmd) != "" {
			extra := ""
			if cwd, ok := obj["cwd"].(string); ok && cwd != "" {
				extra += "\ncwd: " + cwd
			}
			if bg, ok := obj["background"].(bool); ok && bg {
				extra += "\nbackground: true"
			}
			return "command:\n" + cmd + extra
		}
		if path, ok := obj["path"].(string); ok {
			line := "path: " + path
			if c, ok := obj["content"].(string); ok {
				// show size not full content for writes
				line += fmt.Sprintf("\ncontent: (%d bytes)", len(c))
			}
			return line
		}
	}
	return input
}

// MaxDisplayRunes caps overlay body so huge blobs don't flood the terminal.
const MaxDisplayRunes = 4000

func (m Model) View() string {
	if !m.Active {
		return ""
	}
	t := theme.Current()
	w := 72
	if w > 100 {
		w = 100
	}

	// Risk badge
	badgeColor := t.Success
	switch m.Risk {
	case RiskHigh:
		badgeColor = t.Danger
	case RiskMedium:
		badgeColor = t.Warning
	}
	badge := lipgloss.NewStyle().Bold(true).Foreground(badgeColor).
		Render(fmt.Sprintf("[%s RISK]", m.Risk.String()))
	title := lipgloss.NewStyle().Bold(true).Foreground(t.Warning).Render("⚠ Permission")
	tool := lipgloss.NewStyle().Foreground(t.AccentTool).Bold(true).Render(m.Tool)
	reason := m.Reason
	if reason == "" {
		reason = "Tool requires approval before running."
	}
	reason = lipgloss.NewStyle().Foreground(t.TextMuted).Render(reason)

	// Full command body (Q8.2)
	cmdBody := FormatCommand(m.Tool, m.Input)
	cmdBody = clampRunes(cmdBody, MaxDisplayRunes)
	cmdBody = lipgloss.NewStyle().Foreground(t.TextPrimary).Render(cmdBody)

	keys := "y allow  ·  n deny  ·  a always allow  ·  d always deny  ·  Esc deny"
	if m.Dangerous || m.Risk == RiskHigh {
		keys = "y allow once  ·  n deny  ·  Esc deny  ·  (high risk — cannot remember always)"
	}
	bar := lipgloss.NewStyle().Foreground(t.TextMuted).Render(keys)

	header := fmt.Sprintf("%s  %s\ntool: %s\n%s", title, badge, tool, reason)
	body := header + "\n\n" + cmdBody + "\n\n" + bar
	return theme.OverlayStyle(w).Render(body)
}

func clampRunes(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max]) + "\n… (truncated)"
}
