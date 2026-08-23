package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/codeforge/tui/internal/theme"
)

func (m Model) ViewWelcome() string {
	t := theme.Current()

	// Redesigned bold, sleek ASCII Logo for CodeForge
	logo := `
   ██████╗ ██████╗ ██████╗ ███████╗███████╗██████╗ ██████╗  ██████╗ ███████╗
  ██╔════╝██╔═══██╗██╔══██╗██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝
  ██║     ██║   ██║██║  ██║█████╗  █████╗  ██║   ██║██████╔╝██║  ███╗█████╗  
  ██║     ██║   ██║██║  ██║██╔══╝  ██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝  
  ╚██████╗╚██████╔╝██████╔╝███████╗██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗
   ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝
	`

	logoStyle := lipgloss.NewStyle().
		Foreground(t.AccentUser).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width).
		PaddingTop(m.height/2 - 10)

	options := []string{
		"Start New Session",
		"Resume Last Session",
		"Select Session History",
		"Open Settings",
	}

	var menu strings.Builder
	menu.WriteString("\n\n")

	for i, opt := range options {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(t.TextMuted)
		if i == m.welcomeCursor {
			cursor = "❯ "
			style = lipgloss.NewStyle().Foreground(t.AccentUser).Bold(true)
		}
		menu.WriteString(style.Render(cursor + opt))
		menu.WriteString("\n")
	}

	menuStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		PaddingTop(2)

	return lipgloss.JoinVertical(lipgloss.Center,
		logoStyle.Render(logo),
		menuStyle.Render(menu.String()),
	)
}
