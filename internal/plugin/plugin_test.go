package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeforge/tui/internal/tool"
)

func TestLoadAndExec(t *testing.T) {
	dir := t.TempDir()
	yaml := `
name: demo
description: demo
command: /bin/echo
args: ["hello-plugin"]
timeout_sec: 5
`
	path := filepath.Join(dir, "demo.plugin.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	reg := tool.NewRegistry(dir)
	n, lines := LoadAll(reg, dir, []string{dir})
	if n < 1 {
		t.Fatalf("n=%d lines=%v", n, lines)
	}
	ttool, ok := reg.Get("plugin_demo")
	if !ok {
		// name might be plugin_demo
		for _, x := range reg.List() {
			if x.Name() == "plugin_demo" || x.Name() == "demo" {
				ttool = x
				ok = true
				break
			}
		}
	}
	if !ok {
		t.Fatalf("tool missing, have %v", names(reg))
	}
	res := ttool.Execute(json.RawMessage(`{}`))
	if !res.Success {
		t.Fatal(res.Error)
	}
	if res.Output == "" {
		t.Fatal("empty output")
	}
}

func names(r *tool.Registry) []string {
	var o []string
	for _, t := range r.List() {
		o = append(o, t.Name())
	}
	return o
}
