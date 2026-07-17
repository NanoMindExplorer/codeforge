package session

import (
	"strings"
	"testing"

	"github.com/codeforge/tui/internal/provider"
)

func TestCompactPreservesToolOutcomes(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", home+"/s")

	s := New("x", "y", ".")
	// long thread with tool work
	s.Messages = []provider.Message{
		{Role: provider.RoleUser, Content: "fix the bug in main.go"},
		{Role: provider.RoleAssistant, Content: "I'll inspect the file", ToolCalls: []provider.ToolCall{
			{ID: "1", Name: "read_file", Input: `{"path":"main.go"}`},
		}},
		{Role: provider.RoleTool, ToolName: "read_file", ToolCallID: "1", Content: "package main\nfunc main() {}\n"},
		{Role: provider.RoleAssistant, Content: "applying fix", ToolCalls: []provider.ToolCall{
			{ID: "2", Name: "write_file", Input: `{"path":"main.go","content":"fixed"}`},
		}},
		{Role: provider.RoleTool, ToolName: "write_file", ToolCallID: "2", Content: "wrote main.go"},
		{Role: provider.RoleUser, Content: "run tests"},
		{Role: provider.RoleAssistant, Content: "running", ToolCalls: []provider.ToolCall{
			{ID: "3", Name: "run_command", Input: `{"command":"go test ./..."}`},
		}},
		{Role: provider.RoleTool, ToolName: "run_command", ToolCallID: "3", Content: "PASS ok"},
		{Role: provider.RoleAssistant, Content: "all green"},
		// keep tail
		{Role: provider.RoleUser, Content: "ship it"},
		{Role: provider.RoleAssistant, Content: "ready"},
		{Role: provider.RoleUser, Content: "thanks"},
		{Role: provider.RoleAssistant, Content: "you're welcome"},
	}
	_ = s.Save()

	// Snapshot of summary builder before compact (old portion)
	old := s.Messages[:len(s.Messages)-4]
	summary := buildCompactSummary(old, "keep test results")

	// Golden structure assertions (Q4.4 snapshot)
	if !strings.Contains(summary, "Focus: keep test results") {
		t.Fatal(summary)
	}
	if !strings.Contains(summary, "Tool outcomes") {
		t.Fatal("expected tool outcomes section:\n", summary)
	}
	if !strings.Contains(summary, "write_file") {
		t.Fatal("expected write_file outcome:\n", summary)
	}
	if !strings.Contains(summary, "run_command") || !strings.Contains(summary, "PASS") {
		t.Fatal("expected test result preserved:\n", summary)
	}
	if !strings.Contains(summary, "call read_file") && !strings.Contains(summary, "read_file") {
		t.Fatal("expected tool call/outcome:\n", summary)
	}

	res, err := s.Compact(4, "keep test results")
	if err != nil {
		t.Fatal(err)
	}
	if res.AfterMsgs >= res.BeforeMsgs {
		t.Fatalf("%+v", res)
	}
	// compacted blob in first messages should still mention tools
	joined := ""
	for _, m := range s.Messages {
		joined += m.Content + "\n"
	}
	if !strings.Contains(joined, "Tool outcomes") && !strings.Contains(joined, "write_file") {
		t.Fatal("compacted history lost tool facts:\n", joined)
	}
}
