package index

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestBuildAndSearch(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main\nfunc DoThing() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Hello\nDoThing docs\n"), 0644); err != nil {
		t.Fatal(err)
	}
	idx, err := Build(dir)
	if err != nil {
		t.Fatal(err)
	}
	f, s := idx.Stats()
	if f < 2 {
		t.Fatalf("files=%d", f)
	}
	if s < 1 {
		t.Fatalf("symbols=%d", s)
	}
	hits := idx.Search("DoThing", 5)
	if len(hits) == 0 {
		t.Fatal("no hits")
	}
	if hits[0].Path != "hello.go" && hits[0].Path != "README.md" {
		// either is fine; prefer go usually
		t.Log(hits[0].Path)
	}
}

func TestBuildWithProgressAndAsync(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 60; i++ {
		_ = os.WriteFile(filepath.Join(dir, filepath.Base(dir)+"_"+string(rune('a'+i%26))+string(rune('0'+i%10))+".go"),
			[]byte("package p\nfunc F() {}\n"), 0o644)
	}
	var calls atomic.Int32
	idx, err := BuildWithProgress(dir, func(n int, last string) {
		calls.Add(1)
	})
	if err != nil {
		t.Fatal(err)
	}
	if idx == nil {
		t.Fatal("nil index")
	}
	// progress at least final callback
	if calls.Load() < 1 {
		t.Fatal("expected progress callbacks")
	}

	ch := BuildAsync(dir)
	select {
	case res := <-ch:
		if res.Err != nil || res.Index == nil {
			t.Fatal(res.Err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("BuildAsync timeout")
	}
}
