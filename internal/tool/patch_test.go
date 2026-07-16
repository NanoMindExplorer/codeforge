package tool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSearchReplaceUnique(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.go")
	if err := os.WriteFile(path, []byte("package a\n\nfunc Hello() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	sw := NewStagedWriter(dir)
	sw.SetMode(ModeAct)
	sr := &SearchReplace{WorkDir: dir, Staged: sw}
	in, _ := json.Marshal(map[string]any{
		"path": "a.go", "old_string": "Hello", "new_string": "Hi",
	})
	res := sr.Execute(in)
	if !res.Success {
		t.Fatal(res.Error)
	}
	b, _ := os.ReadFile(path)
	if string(b) != "package a\n\nfunc Hi() {}\n" {
		t.Fatalf("got %q", b)
	}
	if res.Diff == "" {
		t.Fatal("expected diff")
	}
}

func TestSearchReplacePlanStages(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "b.txt"), []byte("one two one"), 0644)
	sw := NewStagedWriter(dir)
	sw.SetMode(ModePlan)
	sr := &SearchReplace{WorkDir: dir, Staged: sw}
	in, _ := json.Marshal(map[string]any{
		"path": "b.txt", "old_string": "two", "new_string": "2",
	})
	res := sr.Execute(in)
	if !res.Success {
		t.Fatal(res.Error)
	}
	if !sw.HasPending() {
		t.Fatal("expected staged")
	}
	// file unchanged
	b, _ := os.ReadFile(filepath.Join(dir, "b.txt"))
	if string(b) != "one two one" {
		t.Fatal("plan should not write")
	}
}

func TestApplyPatchUpdate(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "c.txt"), []byte("alpha\nbeta\ngamma\n"), 0644)
	sw := NewStagedWriter(dir)
	sw.SetMode(ModeAct)
	ap := &ApplyPatch{WorkDir: dir, Staged: sw}
	patch := `*** Begin Patch
*** Update File: c.txt
@@
 alpha
-beta
+BETA
 gamma
*** End Patch
`
	in, _ := json.Marshal(map[string]string{"patch": patch})
	res := ap.Execute(in)
	if !res.Success {
		t.Fatal(res.Error)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "c.txt"))
	if string(b) != "alpha\nBETA\ngamma\n" {
		t.Fatalf("got %q", b)
	}
}

func TestApplyHunkLines(t *testing.T) {
	old := "a\nb\nc\n"
	neu, err := applyHunkLines(old, " a\n-b\n+B\n c")
	if err != nil {
		t.Fatal(err)
	}
	if neu != "a\nB\nc\n" {
		t.Fatalf("%q", neu)
	}
}
