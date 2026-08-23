package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codeforge/tui/internal/theme"
)

type OnboardStep int

const (
	StepSelectProvider OnboardStep = iota
	StepInputKey
)

type OnboardModel struct {
	step      OnboardStep
	cursor    int
	providers []string
	input     textinput.Model
	width     int
	height    int

	selectedProvider string
}

func NewOnboardModel() OnboardModel {
	ti := textinput.New()
	ti.Placeholder = "Paste your API key here (Ctrl+V / Shift+Insert)..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60
	ti.EchoCharacter = '•'
	ti.EchoMode = textinput.EchoPassword

	return OnboardModel{
		step: StepSelectProvider,
		providers: []string{
			"Gemini (Google) - Free Tier Available",
			"Claude (Anthropic)",
			"Grok (xAI)",
			"OpenAI (ChatGPT)",
			"Ollama (Local / Offline)",
		},
		input: ti,
	}
}

func (m *OnboardModel) providerID() string {
	switch m.cursor {
	case 0:
		return "gemini"
	case 1:
		return "claude"
	case 2:
		return "grok"
	case 3:
		return "openai"
	case 4:
		return "ollama"
	}
	return "gemini"
}

func (m *OnboardModel) Update(msg tea.Msg) (tea.Cmd, bool, string, string) {
	var cmd tea.Cmd
	done := false
	prov := ""
	key := ""

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.step == StepInputKey {
				m.step = StepSelectProvider
				m.input.SetValue("")
				return nil, false, "", ""
			}
			// Let it pass through if they want to exit
		}

		if m.step == StepSelectProvider {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.providers)-1 {
					m.cursor++
				}
			case "enter", "l":
				m.selectedProvider = m.providerID()
				if m.selectedProvider == "ollama" {
					done = true
					prov = "ollama"
					key = ""
				} else {
					m.step = StepInputKey
				}
			}
		} else if m.step == StepInputKey {
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.input.Value())
				if val != "" {
					done = true
					prov = m.selectedProvider
					key = val
				}
			default:
				m.input, cmd = m.input.Update(msg)
			}
		}

	case tea.MouseMsg:
		if m.step == StepSelectProvider {
			if msg.Type == tea.MouseLeft {
				startY := m.height/2 - 2
				if msg.Y >= startY && msg.Y < startY+len(m.providers) {
					m.cursor = msg.Y - startY
					m.selectedProvider = m.providerID()
					if m.selectedProvider == "ollama" {
						done = true
						prov = "ollama"
						key = ""
					} else {
						m.step = StepInputKey
					}
				}
			}
			if msg.Type == tea.MouseWheelUp && m.cursor > 0 {
				m.cursor--
			} else if msg.Type == tea.MouseWheelDown && m.cursor < len(m.providers)-1 {
				m.cursor++
			}
		}
	}
	return cmd, done, prov, key
}

func (m *OnboardModel) View() string {
	t := theme.Current()
	var body strings.Builder

	asciiArt := " ____          _      _____\n" +
		"    / ___|___   __| | ___|  ___|__  _ __ __ _  ___\n" +
		"   | |   / _ \\ / _` |/ _ \\ |_ / _ \\| '__/ _` |/ _ \\\n" +
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
}
