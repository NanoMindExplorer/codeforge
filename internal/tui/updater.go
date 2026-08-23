package tui

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type UpdateFinishedMsg struct {
	Err error
}

func doUpdate() tea.Msg {
	// Securely update CodeForge by cloning to a temp directory and building.
	// This avoids running 'git pull' in the user's current workspace.
	exe, err := os.Executable()
	if err != nil {
		exe = "codeforge" // fallback
	}

	script := fmt.Sprintf(`
set -e
DIR=$(mktemp -d)
git clone https://github.com/NanoMindExplorer/codeforge.git "$DIR"
cd "$DIR"
go build -o codeforge ./cmd/codeforge
# Try to replace the executable
mv codeforge "%s" || sudo mv codeforge "%s" || echo "Failed to move binary"
rm -rf "$DIR"
`, exe, exe)

	cmd := exec.Command("sh", "-c", script)
	err = cmd.Run()
	return UpdateFinishedMsg{Err: err}
}
