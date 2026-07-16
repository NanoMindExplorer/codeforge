package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/codeforge/tui/internal/provider"
)

func TestSaveLoadV2Layout(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "sessions"))

	s := New("gemini", "gemini-2.5-flash", "/tmp/proj")
	s.Messages = []provider.Message{
		{Role: provider.RoleUser, Content: "hello world test session"},
		{Role: provider.RoleAssistant, Content: "hi"},
	}
	s.TotalCost = 0.01
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	// v2 layout: sessions/<enc-cwd>/<id>/summary.json
	dir, err := s.DirPath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "summary.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "chat_history.jsonl")); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Messages) != 2 {
		t.Fatalf("messages=%d", len(loaded.Messages))
	}
	if loaded.Preview == "" {
		t.Fatal("expected preview")
	}
	if loaded.Format != "v2" {
		t.Fatal(loaded.Format)
	}
}

func TestListForWorkdir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "s"))
	for i := 0; i < 3; i++ {
		s := New("claude", "sonnet", "/work/a")
		s.Messages = []provider.Message{{Role: provider.RoleUser, Content: "msg"}}
		if err := s.Save(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Millisecond) // unique ids
	}
	// other cwd
	o := New("claude", "sonnet", "/work/b")
	o.Messages = []provider.Message{{Role: provider.RoleUser, Content: "other"}}
	_ = o.Save()

	list, err := ListForWorkdir("/work/a", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 3 {
		t.Fatalf("expected >=3 for /work/a, got %d", len(list))
	}
}

func TestForkAndRewind(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "s"))
	s := New("x", "y", "/tmp/p")
	s.Messages = []provider.Message{
		{Role: provider.RoleUser, Content: "first"},
		{Role: provider.RoleAssistant, Content: "ok1"},
		{Role: provider.RoleUser, Content: "second"},
		{Role: provider.RoleAssistant, Content: "ok2"},
	}
	_ = s.Save()
	_, err := s.RecordRewindPoint("first", "turn-1")
	if err != nil {
		t.Fatal(err)
	}
	// message index after 2 msgs would be recorded mid-way; simulate point at index 2
	pts, _ := s.LoadRewindPoints()
	if len(pts) < 1 {
		t.Fatal("no points")
	}
	// manually set message index for test
	pts[0].MessageIndex = 2
	_ = s.SaveRewindPoints(pts)

	child, err := s.Fork()
	if err != nil {
		t.Fatal(err)
	}
	if child.ParentID != s.ID {
		t.Fatal(child.ParentID)
	}
	if len(child.Messages) != 4 {
		t.Fatal(len(child.Messages))
	}

	restored, err := s.ApplyRewind(pts[0])
	if err != nil {
		t.Fatal(err)
	}
	_ = restored
	if len(s.Messages) != 2 {
		t.Fatalf("after rewind messages=%d", len(s.Messages))
	}
}

func TestCompact(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "s"))
	s := New("x", "y", ".")
	for i := 0; i < 10; i++ {
		s.Messages = append(s.Messages,
			provider.Message{Role: provider.RoleUser, Content: "u"},
			provider.Message{Role: provider.RoleAssistant, Content: "a"},
		)
	}
	_ = s.Save()
	res, err := s.Compact(4, "keep API details")
	if err != nil {
		t.Fatal(err)
	}
	if res.AfterMsgs >= res.BeforeMsgs {
		t.Fatalf("compact did not reduce: %+v", res)
	}
}

func TestEncodeCWD(t *testing.T) {
	a := EncodeCWD("/home/user/proj")
	b := EncodeCWD("/home/user/proj")
	if a != b {
		t.Fatal("stable encode")
	}
	if a == "" {
		t.Fatal("empty")
	}
}

func TestShouldAutoCompact(t *testing.T) {
	if !ShouldAutoCompact(900, 1000, 0.85) {
		t.Fatal("expected true")
	}
	if ShouldAutoCompact(100, 1000, 0.85) {
		t.Fatal("expected false")
	}
}
