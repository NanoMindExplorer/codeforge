package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codeforge/tui/internal/provider"
)

func TestSessionDirsMode0700(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CODEFORGE_SESSIONS_DIR", root)
	// recreate with our Dir()
	d, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(d)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm()&0o077 != 0 {
		// On some FS (e.g. restrictive umask already 0700) — require no group/other
		t.Fatalf("sessions root mode %o should be 0700", st.Mode().Perm())
	}

	s := New("gemini", "flash", "/proj/q8")
	s.Messages = []provider.Message{{Role: provider.RoleUser, Content: "hi"}}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	sessDir, err := s.DirPath()
	if err != nil {
		t.Fatal(err)
	}
	st, err = os.Stat(sessDir)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm()&0o077 != 0 {
		t.Fatalf("session dir mode %o", st.Mode().Perm())
	}
	// summary.json 0600
	sum := filepath.Join(sessDir, "summary.json")
	st, err = os.Stat(sum)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm()&0o077 != 0 {
		t.Fatalf("summary mode %o", st.Mode().Perm())
	}
}
