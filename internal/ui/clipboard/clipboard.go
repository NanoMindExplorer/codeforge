// Package clipboard copies text to the system clipboard (best-effort).
package clipboard

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DefaultTimeout is the max wait for external clipboard tools.
// Without this, wl-copy/xclip can hang indefinitely when no display is available
// (common on SSH/Termux/headless) — freezes the TUI when pressing y/Y.
const DefaultTimeout = 400 * time.Millisecond

// Write tries platform clipboards with a hard timeout; falls back to a temp file.
func Write(text string) error {
	if text == "" {
		return fmt.Errorf("empty")
	}
	if p := os.Getenv("CODEFORGE_CLIPBOARD_FILE"); p != "" {
		return os.WriteFile(p, []byte(text), 0o644)
	}
	// Never block the TUI event loop longer than DefaultTimeout.
	switch runtime.GOOS {
	case "darwin":
		if err := pipeArgsTimeout([]string{"pbcopy"}, text, DefaultTimeout); err == nil {
			return nil
		}
	case "windows":
		if err := pipeArgsTimeout([]string{"clip"}, text, DefaultTimeout); err == nil {
			return nil
		}
	default:
		// Prefer tools that fail fast; each attempt is time-bounded.
		for _, args := range [][]string{
			{"wl-copy"},
			{"xclip", "-selection", "clipboard"},
			{"xsel", "--clipboard", "--input"},
			{"termux-clipboard-set"},
		} {
			if err := pipeArgsTimeout(args, text, DefaultTimeout); err == nil {
				return nil
			}
		}
	}
	path := filepath.Join(os.TempDir(), "codeforge-clipboard.txt")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return err
	}
	return fmt.Errorf("no clipboard tool; wrote %s", path)
}

func pipeArgsTimeout(args []string, text string, timeout time.Duration) error {
	if len(args) == 0 {
		return fmt.Errorf("no cmd")
	}
	// Skip missing binaries quickly
	if _, err := exec.LookPath(args[0]); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdin = strings.NewReader(text)
	// Detach from TTY so a hung clipboard helper cannot steal the terminal.
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("clipboard timeout after %s", timeout)
	}
	return err
}
