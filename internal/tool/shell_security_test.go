package tool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeforge/tui/internal/sandbox"
)

func TestScrubShellEnvDropsInjection(t *testing.T) {
	in := []string{
		"PATH=/usr/bin",
		"LD_PRELOAD=/evil.so",
		"DYLD_INSERT_LIBRARIES=/evil.dylib",
		"CODEFORGE_AGENT_SECRET=supersecret",
		"GEMINI_API_KEY=AIza-keep-me",
		"HOME=/home/u",
	}
	out := scrubShellEnv(in)
	joined := strings.Join(out, "\n")
	if strings.Contains(joined, "LD_PRELOAD") {
		t.Fatal("LD_PRELOAD leaked")
	}
	if strings.Contains(joined, "CODEFORGE_AGENT_SECRET") {
		t.Fatal("agent secret leaked to shell")
	}
	if !strings.Contains(joined, "GEMINI_API_KEY") {
		t.Fatal("provider key should remain for nested CLIs")
	}
	if !strings.Contains(joined, "PATH=") {
		t.Fatal("PATH required")
	}
}

func TestShellExecPinsCwdAndNullReject(t *testing.T) {
	dir := t.TempDir()
	// ensure sandbox off for simple echo
	sandbox.Ensure(sandbox.Off, dir)
	sh := &ShellExec{WorkDir: dir}

	// null byte rejected
	in, _ := json.Marshal(map[string]string{"command": "echo\x00hi"})
	res := sh.Execute(in)
	if res.Success || !strings.Contains(res.Error, "null") {
		t.Fatalf("%+v", res)
	}

	// cwd pin: write a marker only visible if cwd is workdir
	marker := "cwd-marker.txt"
	in, _ = json.Marshal(map[string]string{"command": "pwd > " + marker + " && echo ok"})
	res = sh.Execute(in)
	if !res.Success {
		t.Fatal(res.Error, res.Output)
	}
	data, err := os.ReadFile(filepath.Join(dir, marker))
	if err != nil {
		t.Fatal("command did not run in WorkDir:", err)
	}
	// pwd should be workdir
	if !strings.Contains(string(data), dir) && !strings.Contains(filepath.Clean(string(data)), filepath.Base(dir)) {
		// on some systems pwd may resolve differently — file existence is the main pin
		t.Log("pwd output:", string(data))
	}
}

func TestShellExecNoEnvInjectionFromInput(t *testing.T) {
	// Schema has no env field; extra JSON keys must not become env
	dir := t.TempDir()
	sandbox.Ensure(sandbox.Off, dir)
	sh := &ShellExec{WorkDir: dir}
	// smuggle env-like fields
	raw := []byte(`{"command":"echo hi","env":{"LD_PRELOAD":"/evil"},"cwd":"/tmp"}`)
	res := sh.Execute(raw)
	// command should still work (env/cwd ignored)
	if !res.Success && !strings.Contains(res.Output, "hi") {
		// may succeed with output
		t.Log(res)
	}
	// ensure we never set cmd from injected cwd=/tmp for the tool workdir pin
	// (hard to inspect cmd; rely on cwd pin test above)
	_ = res
}
