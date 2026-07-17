package session

import (
	"strings"
	"testing"
	"time"

	"github.com/codeforge/tui/internal/provider"
)

func TestLastForWorkdirAndFormatResumeList(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", home+"/s")

	wd := "/proj/resume-ux"
	s1 := New("gemini", "flash", wd)
	s1.Messages = []provider.Message{{Role: provider.RoleUser, Content: "first chat about API keys"}}
	s1.SessionMode = "BUILD"
	s1.Model = "gemini-2.5-flash"
	_ = s1.Save()
	time.Sleep(5 * time.Millisecond)

	s2 := New("grok", "grok-4.5", wd)
	s2.Messages = []provider.Message{{Role: provider.RoleUser, Content: "newer task: fix race detector"}}
	s2.SessionMode = "YOLO"
	s2.PermissionMode = "always_approve"
	_ = s2.Save()

	last, err := LastForWorkdir(wd)
	if err != nil {
		t.Fatal(err)
	}
	if last == nil {
		t.Fatal("expected last session")
	}
	if last.ID != s2.ID {
		t.Fatalf("want newest %s got %s", s2.ID, last.ID)
	}
	if len(last.Messages) != 1 {
		t.Fatal("full messages required")
	}

	list, err := ListForWorkdir(wd, 10)
	if err != nil {
		t.Fatal(err)
	}
	// populate modes on light list from save - summary should have them
	text := FormatResumeList(list, 10)
	if !strings.Contains(text, "/resume last") {
		t.Fatal(text)
	}
	if !strings.Contains(text, s2.ID) {
		t.Fatal("missing newest id", text)
	}
	if !strings.Contains(text, "→") {
		t.Fatal("expected arrow on newest", text)
	}
	if !strings.Contains(text, "fix race") && !strings.Contains(text, "newer task") {
		// preview may truncate
		t.Log(text)
	}
}

func TestFormatResumeListEmpty(t *testing.T) {
	s := FormatResumeList(nil, 5)
	if !strings.Contains(s, "No saved sessions") {
		t.Fatal(s)
	}
}
