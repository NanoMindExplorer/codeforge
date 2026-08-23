package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"os/exec"
)

type UpdateFinishedMsg struct {
	Err error
}

func doUpdate() tea.Msg {
	script := "curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | CODEFORGE_VERSION=source sh"
	cmd := exec.Command("sh", "-c", script)
	err := cmd.Run()
	return UpdateFinishedMsg{Err: err}
}
