package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePrimary(t *testing.T) {
	dir := t.TempDir()
	w := New(dir)
	_ = os.WriteFile(filepath.Join(dir, "x.go"), []byte("x"), 0644)
	abs, root, err := w.ResolvePath("x.go")
	if err != nil {
		t.Fatal(err)
	}
	if root.Path != dir {
		t.Fatal(root)
	}
	if abs != filepath.Join(dir, "x.go") {
		t.Fatal(abs)
	}
}

func TestMultiRoot(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	_ = os.WriteFile(filepath.Join(b, "pkg.go"), []byte("p"), 0644)
	w := New(a)
	if err := w.AddRoot(b, "lib"); err != nil {
		t.Fatal(err)
	}
	abs, root, err := w.ResolvePath("pkg.go")
	if err != nil {
		t.Fatal(err)
	}
	if root.Path != b {
		t.Fatalf("root=%s", root.Path)
	}
	if abs != filepath.Join(b, "pkg.go") {
		t.Fatal(abs)
	}
}

func TestSandboxEscape(t *testing.T) {
	dir := t.TempDir()
	w := New(dir)
	_, _, err := w.ResolvePath("../outside")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSkipSecrets(t *testing.T) {
	w := New(t.TempDir())
	if !w.ShouldSkipFile(".env") {
		t.Fatal("should skip .env")
	}
	if !w.ShouldSkipDir("node_modules") {
		t.Fatal("should skip node_modules")
	}
}
