package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeforge/tui/internal/agent"
	"github.com/codeforge/tui/internal/provider"
)

// fakeRunner emits canned agent events without a real model.
type fakeRunner struct{}

func (f *fakeRunner) Run(ctx context.Context, workdir, system string, msgs []provider.Message, auth agent.Authorizer, maxIter int, onEvent func(agent.Event)) {
	onEvent(agent.Event{Kind: agent.EventText, Text: "Hello from ACP fake agent. "})
	onEvent(agent.Event{Kind: agent.EventToolCall, ToolName: "list_dir", ToolInput: `{"path":"."}`})
	onEvent(agent.Event{Kind: agent.EventToolResult, ToolName: "list_dir", ToolOutput: "file.go\n", ToolSuccess: true})
	onEvent(agent.Event{Kind: agent.EventText, Text: "Done."})
	onEvent(agent.Event{Kind: agent.EventDone})
}

type bufTransport struct {
	mu   sync.Mutex
	msgs []map[string]any
}

func (b *bufTransport) Write(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	b.mu.Lock()
	b.msgs = append(b.msgs, m)
	b.mu.Unlock()
	return nil
}

func (b *bufTransport) snapshot() []map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]map[string]any, len(b.msgs))
	copy(out, b.msgs)
	return out
}

func TestACPinitializeSessionNewPrompt(t *testing.T) {
	tx := &bufTransport{}
	srv := NewServer(Options{
		Version:       "test",
		WorkDir:       t.TempDir(),
		AlwaysApprove: true,
		Quiet:         true,
		Runner:        &fakeRunner{},
	})
	srv.SetTransport(tx)

	// initialize
	srv.Handle(mustLine(t, map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]any{"protocolVersion": 1},
	}))
	msgs := waitMsgs(t, tx, 1, time.Second)
	initRes := msgs[0]
	if initRes["error"] != nil {
		t.Fatal(initRes)
	}
	result, _ := initRes["result"].(map[string]any)
	if result["protocolVersion"].(float64) != 1 {
		t.Fatal(result)
	}
	caps, _ := result["agentCapabilities"].(map[string]any)
	if caps["loadSession"] != true {
		t.Fatal("loadSession cap")
	}

	// session/new
	srv.Handle(mustLine(t, map[string]any{
		"jsonrpc": "2.0", "id": 2, "method": "session/new",
		"params": map[string]any{"cwd": t.TempDir()},
	}))
	msgs = waitMsgs(t, tx, 2, 10*time.Second)
	newRes := msgs[1]
	if newRes["error"] != nil {
		t.Fatalf("%v", newRes["error"])
	}
	sid := newRes["result"].(map[string]any)["sessionId"].(string)
	if sid == "" {
		t.Fatal("empty session id")
	}

	// session/prompt
	before := len(tx.snapshot())
	srv.Handle(mustLine(t, map[string]any{
		"jsonrpc": "2.0", "id": 3, "method": "session/prompt",
		"params": map[string]any{
			"sessionId": sid,
			"prompt":    []map[string]any{{"type": "text", "text": "hi"}},
		},
	}))

	deadline := time.Now().Add(5 * time.Second)
	var promptDone map[string]any
	for time.Now().Before(deadline) {
		for _, m := range tx.snapshot()[before:] {
			if idNum, ok := m["id"].(float64); ok && idNum == 3 {
				promptDone = m
			}
		}
		if promptDone != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if promptDone == nil {
		t.Fatalf("prompt did not complete; msgs=%v", tx.snapshot())
	}
	if promptDone["error"] != nil {
		t.Fatal(promptDone["error"])
	}
	pr := promptDone["result"].(map[string]any)
	if pr["stopReason"] != StopEndTurn {
		t.Fatal(pr)
	}

	var sawText, sawTool bool
	for _, m := range tx.snapshot() {
		if m["method"] == "session/update" {
			params, _ := m["params"].(map[string]any)
			upd, _ := params["update"].(map[string]any)
			switch upd["sessionUpdate"] {
			case "agent_message_chunk":
				sawText = true
			case "tool_call":
				sawTool = true
			}
		}
	}
	if !sawText || !sawTool {
		t.Fatalf("missing streams text=%v tool=%v", sawText, sawTool)
	}
}

func TestServeStdioScripted(t *testing.T) {
	// Dual pipes model real stdio. Do NOT poll a shared bytes.Buffer from the
	// test while ServeStdio writes — bytes.Buffer is not concurrent-safe and
	// triggers the race detector (Q0 CI).
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	srv := NewServer(Options{
		Version: "test", WorkDir: t.TempDir(), AlwaysApprove: true,
		Quiet: true, Runner: &fakeRunner{},
	})
	done := make(chan error, 1)
	go func() {
		err := ServeStdio(srv, inR, outW)
		// Unblock any reader waiting on outR after the server exits.
		_ = outW.Close()
		done <- err
	}()
	t.Cleanup(func() {
		_ = inW.Close()
		_ = outR.Close()
	})

	b, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]any{"protocolVersion": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := inW.Write(append(b, '\n')); err != nil {
		t.Fatal(err)
	}

	// Read one NDJSON response line with a hard deadline (no shared buffer).
	type lineResult struct {
		line string
		err  error
	}
	got := make(chan lineResult, 1)
	go func() {
		line, err := bufio.NewReader(outR).ReadString('\n')
		got <- lineResult{line: line, err: err}
	}()

	var line string
	select {
	case r := <-got:
		if r.err != nil {
			t.Fatalf("read initialize response: %v", r.err)
		}
		line = r.line
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for initialize response")
	}
	if !strings.Contains(line, `"result"`) {
		t.Fatalf("got %q", line)
	}

	_ = inW.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("ServeStdio: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stdio hang")
	}
}

func TestExtractPromptText(t *testing.T) {
	s := extractPromptText([]ContentBlock{
		{Type: "text", Text: "hello"},
		{Type: "resource", Resource: map[string]any{"uri": "file:///a.go", "text": "package a"}},
	})
	if !strings.Contains(s, "hello") || !strings.Contains(s, "package a") {
		t.Fatal(s)
	}
}

func mustLine(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func waitMsgs(t *testing.T, tx *bufTransport, n int, timeout time.Duration) []map[string]any {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		s := tx.snapshot()
		if len(s) >= n {
			return s
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %d msgs, have %d", n, len(tx.snapshot()))
	return nil
}
