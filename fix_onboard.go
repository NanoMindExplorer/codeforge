package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("internal/tui/onboard.go")
	str := string(content)

	oldViewStart := strings.Index(str, "func (m *OnboardModel) View() string {")
	
	newView := `func (m *OnboardModel) View() string {
	t := theme.Current()
	var body strings.Builder

	asciiArt := " ____          _      _____\n" +
		"    / ___|___   __| | ___|  ___|__  _ __ __ _  ___\n" +
		"   | |   / _ \\ / _` + "`" + ` |/ _ \\ |_ / _ \\| '__/ _` + "`" + ` |/ _ \\\n" +
		"   | |__| (_) | (_| |  __/  _| (_) | | | (_| |  __/\n" +
		"    \\____\\___/ \\__,_|\\___|_|  \\___/|_|  \\__, |\\___|\n" +
		"                                        |___/"

	asciiStyle := lipgloss.NewStyle().Foreground(t.AccentUser).Bold(true).Align(lipgloss.Center).Width(m.width)
	body.WriteString(asciiStyle.Render(asciiArt))
	body.WriteString("\n\n")

	subtitleStyle := lipgloss.NewStyle().Foreground(t.TextMuted).Align(lipgloss.Center).Width(m.width)
	body.WriteString(subtitleStyle.Render("By NanoMindExplorer\n\nFirst-run setup · multi-provider"))
	body.WriteString("\n\n")

	descStyle := lipgloss.NewStyle().Foreground(t.TextPrimary).Align(lipgloss.Center).Width(m.width)

	if m.step == StepSelectProvider {
		body.WriteString(descStyle.Render("You need one API key. Several providers can coexist;\nonly ONE is active at a time.\nPriority: grok → gemini → claude → openai\n\n(Use Up/Down arrows to select, Enter to confirm, Mouse clicks supported)"))
		body.WriteString("\n\n")

		menuStyle := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width)
		var menu strings.Builder
		for i, p := range m.providers {
			cursor := "   ○ "
			style := lipgloss.NewStyle().Foreground(t.TextMuted)
			if i == m.cursor {
				cursor = "   ● "
				style = lipgloss.NewStyle().Foreground(t.AccentUser).Bold(true)
			}
			menu.WriteString(style.Render(cursor + p))
			menu.WriteString("\n")
		}
		body.WriteString(menuStyle.Render(menu.String()))

	} else {
		body.WriteString(descStyle.Render("Enter API Key for " + strings.ToUpper(m.selectedProvider)))
		body.WriteString("\n\n")

		var helpText string
		switch m.selectedProvider {
		case "gemini":
			helpText = "Get your free API key at: https://aistudio.google.com/apikey"
		case "claude":
			helpText = "Get your API key at: https://console.anthropic.com/"
		case "grok":
			helpText = "Get your API key at: https://console.x.ai/"
		case "openai":
			helpText = "Get your API key at: https://platform.openai.com/api-keys"
		}

		body.WriteString(subtitleStyle.Render(helpText))
		body.WriteString("\n\n")

		inputStyle := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width)
		body.WriteString(inputStyle.Render(m.input.View()))

		body.WriteString("\n\n")
		body.WriteString(subtitleStyle.Render("(Press Enter to save, Esc to go back)"))
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body.String())
}`
	
	newStr := str[:oldViewStart] + newView
	os.WriteFile("internal/tui/onboard.go", []byte(newStr), 0644)
}
