package sessionpicker

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeforge/tui/internal/provider"
	"github.com/codeforge/tui/internal/session"
	"github.com/codeforge/tui/internal/theme"
)

func TestResumePicker(t *testing.T) {
	theme.SetMotion(false)
	home := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "s"))
	s := session.New("g", "m", "/proj")
	s.Messages = []provider.Message{{Role: provider.RoleUser, Content: "hello resume picker"}}
	_ = s.Save()

	var m Model
	m.Width = 60
	m.Open("/proj")
	if !m.Active {
		t.Fatal("active")
	}
	if len(m.Items) < 1 {
		t.Fatal("items")
	}
	view := m.View()
	if !strings.Contains(view, "Resume") {
		t.Fatal(view)
	}
	m.Confirm()
	if m.Selected == nil {
		t.Fatal("selected")
	}
}

func TestRewindPicker(t *testing.T) {
	var m RewindModel
	m.Width = 50
	m.Open([]session.RewindPoint{
		{ID: "1", Preview: "first", MessageIndex: 1},
		{ID: "2", Preview: "second", MessageIndex: 3},
	})
	if len(m.Items) != 2 {
		t.Fatal(len(m.Items))
	}
	m.Move(1)
	m.Confirm()
	if m.Selected == nil {
		t.Fatal("nil")
	}
}
