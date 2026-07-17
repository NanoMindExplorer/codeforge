package tui

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codeforge/tui/internal/index"
)

// buildIndexAsync starts a background codebase index if none is loaded (Q7.1).
// Bootstrap skips sync index for fast TUI cold start; this fills search tools later.
// Opt-out: CODEFORGE_INDEX=0|false|off
func buildIndexAsync(workdir string) tea.Cmd {
	return func() tea.Msg {
		if workdir == "" {
			return nil
		}
		if index.Global() != nil {
			return nil
		}
		v := strings.ToLower(strings.TrimSpace(os.Getenv("CODEFORGE_INDEX")))
		if v == "0" || v == "false" || v == "off" {
			return nil
		}
		res := <-index.BuildAsync(workdir)
		if res.Err != nil {
			return IndexReadyMsg{Err: res.Err}
		}
		if res.Index != nil {
			index.SetGlobal(res.Index)
			f, s := res.Index.Stats()
			return IndexReadyMsg{Files: f, Symbols: s}
		}
		return nil
	}
}
