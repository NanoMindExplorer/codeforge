package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codeforge/tui/internal/provider"
)

func TestExportImport(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "sess"))

	s := New("gemini", "flash", "/tmp")
	s.Messages = []provider.Message{{Role: provider.RoleUser, Content: "hello export test"}}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(home, "out.json")
	if err := Export(s.ID, dest); err != nil {
		t.Fatal(err)
	}
	// wipe sessions dir content by using new dir
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "sess2"))
	imp, err := Import(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(imp.Messages) != 1 {
		t.Fatal(imp.Messages)
	}
	if imp.Format != "v2" {
		t.Fatal(imp.Format)
	}
}

func TestExportAll(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "s"))
	s := New("x", "y", ".")
	s.Messages = []provider.Message{{Role: provider.RoleUser, Content: "z"}}
	_ = s.Save()
	n, err := ExportAll(filepath.Join(home, "all"))
	if err != nil || n < 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	entries, _ := os.ReadDir(filepath.Join(home, "all"))
	if len(entries) < 1 {
		t.Fatal("no files")
	}
}
