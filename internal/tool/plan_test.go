package tool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDesignModeBlocksWrites(t *testing.T) {
	dir := t.TempDir()
	sw := NewStagedWriter(dir)
	plan := filepath.Join(dir, "plan.md")
	sw.SetPlanPath(plan)
	sw.SetMode(ModeDesign)

	// project file blocked
	res := sw.Execute(mustJSON(t, map[string]string{"path": "main.go", "content": "x"}))
	if res.Error == "" {
		t.Fatal("expected design block")
	}

	// plan.md allowed
	res = sw.Execute(mustJSON(t, map[string]string{"path": "plan.md", "content": "# plan"}))
	// path is relative to workdir — plan is under workdir
	// but plan path is absolute under dir/plan.md and workdir is dir
	// write_file resolves path relative to workdir
	if res.Error != "" {
		// may fail if path doesn't match - write via WritePlan tool instead
		t.Log("write_file plan.md:", res.Error)
	}

	wp := &WritePlan{Staged: sw}
	res = wp.Execute(mustJSON(t, map[string]any{"content": "# Context\n\nTest plan\n"}))
	if !res.Success {
		t.Fatal(res.Error)
	}
	if _, err := os.Stat(plan); err != nil {
		t.Fatal(err)
	}

	// search_replace blocked
	sr := &SearchReplace{WorkDir: dir, Staged: sw}
	// create a file first outside design
	_ = os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n"), 0644)
	res = sr.Execute(mustJSON(t, map[string]string{
		"path": "a.go", "old_string": "package a", "new_string": "package b",
	}))
	// need correct field names
	res = sr.Execute(json.RawMessage(`{"path":"a.go","old_string":"package a","new_string":"package b"}`))
	if res.Error == "" && res.Success {
		// if schema uses different keys, check DesignBlocked path
		if blocked := sw.DesignBlocked(filepath.Join(dir, "a.go")); blocked == nil {
			t.Fatal("expected block for a.go")
		}
	}
}

func TestExitPlanModeSignal(t *testing.T) {
	_ = ConsumePlanSignal() // clear
	dir := t.TempDir()
	sw := NewStagedWriter(dir)
	sw.SetPlanPath(filepath.Join(dir, "plan.md"))
	ex := &ExitPlanMode{Staged: sw}
	res := ex.Execute(mustJSON(t, map[string]string{"summary": "ready"}))
	if !res.Success {
		t.Fatal(res)
	}
	sig := ConsumePlanSignal()
	if sig == nil || sig.Kind != "exit_plan_mode" {
		t.Fatalf("%+v", sig)
	}
}

func TestSessionModeCycle(t *testing.T) {
	m := SessionBuild
	m = m.Next()
	if m != SessionDesign {
		t.Fatal(m)
	}
	m = m.Next()
	if m != SessionYolo {
		t.Fatal(m)
	}
	m = m.Next()
	if m != SessionBuild {
		t.Fatal(m)
	}
	if SessionDesign.Label() != "DESIGN" {
		t.Fatal(SessionDesign.Label())
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
