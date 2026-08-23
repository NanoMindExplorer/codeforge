import re

with open('internal/tui/onboard.go', 'r') as f:
    content = f.read()

# Replace View method in OnboardModel
old_view_start = content.find("func (m *OnboardModel) View() string {")
if old_view_start != -1:
    new_view = """func (m *OnboardModel) View() string {
	t := theme.Current()

	var body strings.Builder

	// ASCII Art
	asciiArt := ` ____          _      _____                    
|  _ \        | |    |  ___|                   
| |_) | _   _ | |__  | |_  ___  _ __  __ _  ___ 
|  _ < | | | || '_ \ |  _|/ _ \| '__|/ _` |/ _ \
| |_) || |_| || |_) || | | (_) | |  | (_| |  __/
|____/  \__,_||_.__/ \_|  \___/|_|   \__, |\___|
                                      __/ |     
                                     |___/      `
	asciiStyle := lipgloss.NewStyle().Foreground(t.AccentUser).Bold(true).Align(lipgloss.Center).Width(m.width)
	body.WriteString(asciiStyle.Render(asciiArt))
	body.WriteString("\\n\\n")

	subtitleStyle := lipgloss.NewStyle().Foreground(t.TextMuted).Align(lipgloss.Center).Width(m.width)
	body.WriteString(subtitleStyle.Render("First-run setup · multi-provider"))
	body.WriteString("\\n\\n")

	descStyle := lipgloss.NewStyle().Foreground(t.TextNormal).Align(lipgloss.Center).Width(m.width)

	if m.step == StepSelectProvider {
		body.WriteString(descStyle.Render("You need one API key. Priority: grok → gemini → claude → openai\\n(Use Up/Down arrows to select, Enter to confirm, Mouse clicks supported)"))
		body.WriteString("\\n\\n")

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
			menu.WriteString("\\n")
		}
		body.WriteString(menuStyle.Render(menu.String()))

	} else {
		body.WriteString(descStyle.Render("Enter API Key for " + strings.ToUpper(m.selectedProvider)))
		body.WriteString("\\n\\n")

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
		body.WriteString("\\n\\n")

		inputStyle := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width)
		body.WriteString(inputStyle.Render(m.input.View()))

		body.WriteString("\\n\\n")
		body.WriteString(subtitleStyle.Render("(Press Enter to save, Esc to go back)"))
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body.String())
}"""
    content = content[:old_view_start] + new_view

with open('internal/tui/onboard.go', 'w') as f:
    f.write(content)
