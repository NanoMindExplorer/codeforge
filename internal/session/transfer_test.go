package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeforge/tui/internal/provider"
)

func TestExportImport(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "sess"))

	s := New("gemini", "flash", "/tmp")
	s.Messages = []provider.Message{{Role: provider.RoleUser, Content: "hello export test"}}
	s.PermissionMode = "always_approve"
	s.SessionMode = "YOLO"
	s.Model = "gemini-2.5-flash"
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(home, "out.json")
	if err := Export(s.ID, dest); err != nil {
		t.Fatal(err)
	}
	// verify export JSON carries modes + model (Q4.5)
	raw, _ := os.ReadFile(dest)
	body := string(raw)
	if !strings.Contains(body, `"permission_mode": "always_approve"`) && !strings.Contains(body, `"permission_mode":"always_approve"`) {
		t.Fatalf("export missing permission_mode:\n%s", body)
	}
	if !strings.Contains(body, "YOLO") {
		t.Fatalf("export missing session_mode:\n%s", body)
	}
	if !strings.Contains(body, "gemini-2.5-flash") {
		t.Fatalf("export missing model:\n%s", body)
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
	if imp.PermissionMode != "always_approve" {
		t.Fatalf("perm=%q", imp.PermissionMode)
	}
	if imp.SessionMode != "YOLO" {
		t.Fatalf("mode=%q", imp.SessionMode)
	}
	if imp.Model != "gemini-2.5-flash" {
		t.Fatalf("model=%q", imp.Model)
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
