package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeforge/tui/internal/provider"
)

// CompactThreshold is the default fraction of max context that triggers auto-compact.
const DefaultAutoCompactPct = 0.85

// CompactResult describes what compaction did.
type CompactResult struct {
	BeforeMsgs int
	AfterMsgs  int
	Summary    string
}

// Compact compresses early conversation into a single system summary message.
// Keeps the last keepLast user/assistant turns (default 6 messages).
// Optional hint guides what to preserve in the summary text.
func (s *Session) Compact(keepLast int, hint string) (CompactResult, error) {
	if s == nil {
		return CompactResult{}, fmt.Errorf("nil session")
	}
	if keepLast <= 0 {
		keepLast = 6
	}
	n := len(s.Messages)
	if n <= keepLast+1 {
		return CompactResult{BeforeMsgs: n, AfterMsgs: n, Summary: "nothing to compact"}, nil
	}
	// Save pre-compact checkpoint of full history
	if dir, err := s.DirPath(); err == nil {
		cpDir := filepath.Join(dir, "compaction_checkpoints")
		_ = os.MkdirAll(cpDir, 0755)
		_ = writeJSON(filepath.Join(cpDir, time.Now().Format("20060102-150405")+".json"), s.Messages)
	}

	old := s.Messages[:n-keepLast]
	keep := s.Messages[n-keepLast:]
	summary := buildCompactSummary(old, hint)
	sys := provider.Message{
		Role:    provider.RoleUser,
		Content: "[compacted history]\n" + summary + "\n[/compacted history]\nContinue from the recent messages below.",
	}
	// Use assistant acknowledgment so roles alternate cleanly for most providers
	ack := provider.Message{
		Role:    provider.RoleAssistant,
		Content: "Understood — I have the compacted earlier context and will continue from the recent messages.",
	}
	s.Messages = append([]provider.Message{sys, ack}, keep...)
	s.Preview = truncate("compacted: "+summary, 80)
	if err := s.Save(); err != nil {
		return CompactResult{}, err
	}
	_ = s.AppendEvent("compact", map[string]any{
		"before": n, "after": len(s.Messages), "hint": hint,
	})
	return CompactResult{
		BeforeMsgs: n,
		AfterMsgs:  len(s.Messages),
		Summary:    summary,
	}, nil
}

func buildCompactSummary(msgs []provider.Message, hint string) string {
	var b strings.Builder
	if hint != "" {
		b.WriteString("Focus: ")
		b.WriteString(hint)
		b.WriteByte('\n')
	}
	b.WriteString(fmt.Sprintf("Earlier conversation (%d messages):\n", len(msgs)))
	userN, asstN, toolN, callN := 0, 0, 0, 0
	// Q4.4: preserve tool outcomes so the model retains edit/run results after compact.
	var toolOutcomes []string
	for _, m := range msgs {
		switch m.Role {
		case provider.RoleUser:
			userN++
			if userN <= 8 {
				b.WriteString("- User: ")
				b.WriteString(truncate(m.Content, 160))
				b.WriteByte('\n')
			}
		case provider.RoleAssistant:
			asstN++
			if asstN <= 6 {
				if strings.TrimSpace(m.Content) != "" {
					b.WriteString("- Assistant: ")
					b.WriteString(truncate(m.Content, 120))
					b.WriteByte('\n')
				}
			}
			for _, tc := range m.ToolCalls {
				callN++
				if callN <= 16 {
					toolOutcomes = append(toolOutcomes, fmt.Sprintf(
						"call %s(%s)", tc.Name, truncate(tc.Input, 80)))
				}
			}
		case provider.RoleTool:
			toolN++
			if toolN <= 16 {
				status := "ok"
				if m.IsError {
					status = "ERR"
				}
				name := m.ToolName
				if name == "" {
					name = "tool"
				}
				toolOutcomes = append(toolOutcomes, fmt.Sprintf(
					"%s [%s]: %s", name, status, truncate(m.Content, 100)))
			}
		}
	}
	if len(toolOutcomes) > 0 {
		b.WriteString("Tool outcomes (preserve facts):\n")
		for _, line := range toolOutcomes {
			b.WriteString("  · ")
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	if userN > 8 || asstN > 6 || toolN > 16 {
		b.WriteString(fmt.Sprintf("… (%d user, %d assistant, %d tool results total)\n", userN, asstN, toolN))
	}
	return strings.TrimSpace(b.String())
}

// ShouldAutoCompact reports whether tokens exceeded pct of maxContext.
func ShouldAutoCompact(tokens, maxContext int, pct float64) bool {
	if maxContext <= 0 || tokens <= 0 {
		return false
	}
	if pct <= 0 {
		pct = DefaultAutoCompactPct
	}
	return float64(tokens) >= float64(maxContext)*pct
}

// EstimateTokens is a rough char/4 estimate when provider CountTokens unavailable.
func EstimateTokens(msgs []provider.Message) int {
	n := 0
	for _, m := range msgs {
		n += len(m.Content)/4 + 4
		for _, tc := range m.ToolCalls {
			n += len(tc.Input)/4 + 8
		}
	}
	return n
}
