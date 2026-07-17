package clipboard

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteFileFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clip.txt")
	t.Setenv("CODEFORGE_CLIPBOARD_FILE", path)
	if err := Write("hello focus keys"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "hello focus keys" {
		t.Fatalf("%v %q", err, b)
	}
}

func TestWriteEmpty(t *testing.T) {
	if err := Write(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestPipeTimeoutDoesNotHang(t *testing.T) {
	// A command that sleeps longer than our timeout must return quickly.
	start := time.Now()
	err := pipeArgsTimeout([]string{"sleep", "5"}, "x", 100*time.Millisecond)
	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Fatalf("timeout path hung: %v", elapsed)
	}
	if err == nil {
		t.Fatal("expected timeout/error from sleep 5")
	}
}
