package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeforge/tui/internal/provider"
)

// Q4.1 — simulate crash mid-save: partial/corrupt jsonl must not lose valid turns.
func TestCrashMidSaveRecoversValidLines(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "s"))

	s := New("gemini", "flash", "/proj/crash")
	s.Messages = []provider.Message{
		{Role: provider.RoleUser, Content: "step-1"},
		{Role: provider.RoleAssistant, Content: "ok-1"},
		{Role: provider.RoleUser, Content: "step-2"},
		{Role: provider.RoleAssistant, Content: "ok-2"},
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	dir, err := s.DirPath()
	if err != nil {
		t.Fatal(err)
	}
	hist := filepath.Join(dir, "chat_history.jsonl")

	// Append a truncated JSON line (crash while writing next message)
	f, err := os.OpenFile(hist, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`{"role":"user","content":"partial-cut`)
	_ = f.Close()

	loaded, err := Load(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Messages) != 4 {
		t.Fatalf("expected 4 recovered messages, got %d", len(loaded.Messages))
	}
	if loaded.Messages[0].Content != "step-1" || loaded.Messages[3].Content != "ok-2" {
		t.Fatalf("content corrupted: %+v", loaded.Messages)
	}
}

// Atomic save updates history completely; orphan temps do not break Load.
func TestAtomicSaveDoesNotClobberOnFailedWrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "s"))

	s := New("x", "y", "/proj/atomic")
	s.Messages = []provider.Message{{Role: provider.RoleUser, Content: "stable"}}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	dir, _ := s.DirPath()
	hist := filepath.Join(dir, "chat_history.jsonl")
	before, _ := os.ReadFile(hist)

	// Simulate orphan temp from interrupted write
	tmp := filepath.Join(dir, ".cf-orphan.tmp")
	_ = os.WriteFile(tmp, []byte("incomplete"), 0o644)

	// Save again successfully
	s.Messages = append(s.Messages, provider.Message{Role: provider.RoleAssistant, Content: "next"})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(hist)
	if string(after) == string(before) {
		t.Fatal("expected history to update")
	}
	if !strings.Contains(string(after), "next") {
		t.Fatalf("missing next: %s", after)
	}
	loaded, err := Load(s.ID)
	if err != nil || len(loaded.Messages) != 2 {
		t.Fatalf("load: %v n=%d", err, len(loaded.Messages))
	}
}

// Kill after history write but before summary refresh: Load still gets messages.
func TestCrashAfterHistoryBeforeSummary(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "s"))

	s := New("x", "y", "/proj/order")
	s.Messages = []provider.Message{{Role: provider.RoleUser, Content: "v1"}}
	_ = s.Save()
	dir, _ := s.DirPath()

	// Manually write newer history without updating summary message_count
	msgs := []provider.Message{
		{Role: provider.RoleUser, Content: "v1"},
		{Role: provider.RoleAssistant, Content: "v2-assistant"},
	}
	var buf strings.Builder
	enc := json.NewEncoder(&buf)
	for _, m := range msgs {
		_ = enc.Encode(m)
	}
	_ = os.WriteFile(filepath.Join(dir, "chat_history.jsonl"), []byte(buf.String()), 0o644)

	loaded, err := Load(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Messages) != 2 {
		t.Fatalf("messages=%d", len(loaded.Messages))
	}
	if loaded.Messages[1].Content != "v2-assistant" {
		t.Fatal(loaded.Messages[1].Content)
	}
}
