package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveUndoLastYOLOStyle(t *testing.T) {
	// YOLO: write hits disk immediately; checkpoint holds pre-image for /undo.
	root := t.TempDir()
	t.Setenv("CODEFORGE_CHECKPOINTS_DIR", root)
	workdir := t.TempDir()
	rel := "app.txt"
	abs := filepath.Join(workdir, rel)
	if err := os.WriteFile(abs, []byte("v1-yolo"), 0o644); err != nil {
		t.Fatal(err)
	}
	sid := "yolo-sess"
	// before overwrite
	id, err := Save(sid, abs, rel, "v1-yolo")
	if err != nil || id == "" {
		t.Fatal(err, id)
	}
	// agent writes new content
	if err := os.WriteFile(abs, []byte("v2-yolo"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := UndoLast(sid)
	if err != nil {
		t.Fatal(err)
	}
	if got != rel {
		t.Fatalf("rel=%q", got)
	}
	data, _ := os.ReadFile(abs)
	if string(data) != "v1-yolo" {
		t.Fatalf("restored=%q", data)
	}
}

func TestBUILDApplyThenCheckpointUndo(t *testing.T) {
	// BUILD: stage → accept → ApplyAccepted returns OldContent → checkpoint.Save → undo
	root := t.TempDir()
	t.Setenv("CODEFORGE_CHECKPOINTS_DIR", root)
	workdir := t.TempDir()
	rel := "lib.go"
	abs := filepath.Join(workdir, rel)
	_ = os.WriteFile(abs, []byte("package lib\n// v1\n"), 0o644)

	// simulate staged apply result
	old := "package lib\n// v1\n"
	newContent := "package lib\n// v2 applied\n"
	if err := os.WriteFile(abs, []byte(newContent), 0o644); err != nil {
		t.Fatal(err)
	}
	sid := "build-sess"
	if _, err := Save(sid, abs, rel, old); err != nil {
		t.Fatal(err)
	}
	// second file
	rel2 := "other.txt"
	abs2 := filepath.Join(workdir, rel2)
	_ = os.WriteFile(abs2, []byte("a"), 0o644)
	if _, err := Save(sid, abs2, rel2, "a"); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(abs2, []byte("b"), 0o644)

	// undo last (other.txt)
	got, err := UndoLast(sid)
	if err != nil {
		t.Fatal(err)
	}
	if got != rel2 {
		t.Fatalf("expected %s got %s", rel2, got)
	}
	d2, _ := os.ReadFile(abs2)
	if string(d2) != "a" {
		t.Fatalf("other=%q", d2)
	}
	// undo lib.go
	got, err = UndoLast(sid)
	if err != nil {
		t.Fatal(err)
	}
	if got != rel {
		t.Fatalf("got %s", got)
	}
	d1, _ := os.ReadFile(abs)
	if string(d1) != old {
		t.Fatalf("lib=%q", d1)
	}
}

func TestUndoAfterRewindWindow(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CODEFORGE_CHECKPOINTS_DIR", root)
	workdir := t.TempDir()
	sid := "rewind-sess"
	// Use past mark so second-precision timestamps still count as "after"
	mark := time.Now().Add(-2 * time.Second)

	rel := "f.txt"
	abs := filepath.Join(workdir, rel)
	_ = os.WriteFile(abs, []byte("new"), 0o644)
	if _, err := Save(sid, abs, rel, "old"); err != nil {
		t.Fatal(err)
	}
	// write new after checkpoint
	_ = os.WriteFile(abs, []byte("new"), 0o644)

	restored, err := UndoAfter(sid, mark)
	if err != nil {
		t.Fatal(err)
	}
	if len(restored) < 1 {
		t.Fatal("expected restore")
	}
	data, _ := os.ReadFile(abs)
	if string(data) != "old" {
		t.Fatalf("%q", data)
	}
}

func TestUndoEmpty(t *testing.T) {
	t.Setenv("CODEFORGE_CHECKPOINTS_DIR", t.TempDir())
	_, err := UndoLast("empty-sess")
	if err == nil {
		t.Fatal("expected error")
	}
}
