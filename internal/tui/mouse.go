package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/codeforge/tui/internal/onboarding"
)

func (m Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {

	if m.mode == ModeOnboard {
		cmd, done, prov, key := m.onboard.Update(msg)
		if done {
			if prov != "ollama" && key == "" {
				// wait
			} else {
				m.mode = ModeInsert
				m.focusPrompt = true
				if _, err := onboarding.ApplyKey(m.providerReg, prov, key, ""); err == nil {
					m.chat.AddSystemMessage("✅ API Key configured successfully for " + prov)
				}
			}
		}
		if cmd != nil {
			return m, cmd
		}
		return m, nil
	}

	if m.mode == ModeWelcome {
		if msg.Type == tea.MouseLeft {
			startY := m.height/2 + 1
			if msg.Y >= startY && msg.Y <= startY+3 {
				m.welcomeCursor = msg.Y - startY
				return m.updateWelcome(tea.KeyMsg(tea.Key{Type: tea.KeyEnter, Alt: false}))
			}
		}
		if msg.Type == tea.MouseWheelUp && m.welcomeCursor > 0 {
			m.welcomeCursor--
		} else if msg.Type == tea.MouseWheelDown && m.welcomeCursor < 3 {
			m.welcomeCursor++
		}
		return m, nil
	}

	var cmds []tea.Cmd

	var sideW, chatW int
	if m.showPanels && m.width >= 100 {
		chatW = m.width * 58 / 100
		sideW = (m.width - chatW - 4) / 2
		if sideW < 14 {
			sideW = 14
		}
	} else {
		chatW = m.width
	}

	isContextPane := m.showPanels && m.width >= 100 && msg.X >= m.width-sideW
	// we assume diff pane is in the middle (chatW to m.width-sideW), but we don't have scrolling in diff pane currently anyway

	isChatInput := msg.Y >= m.height-5 // Rough estimate of textarea height area

	switch msg.Type {
	case tea.MouseWheelUp:
		if isContextPane {
			nc, cmd := m.context.Update(msg)
			m.context = nc.(ContextModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else if msg.X < chatW {
			// Scroll chat
			nc, cmd := m.chat.Update(msg)
			m.chat = nc.(ChatModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case tea.MouseWheelDown:
		if isContextPane {
			nc, cmd := m.context.Update(msg)
			m.context = nc.(ContextModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else if msg.X < chatW {
			nc, cmd := m.chat.Update(msg)
			m.chat = nc.(ChatModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case tea.MouseLeft:
		if isChatInput && msg.X < chatW {
			// Focus textarea
			m.chat.FocusInput()
		} else if !isContextPane && msg.X < chatW {
			// Blur textarea, focus chat history
			m.chat.BlurInput()
			m.chat.SetMode(ModeNormal)
		}

		// Pass to context in case it handles click
		if isContextPane {
			nc, cmd := m.context.Update(msg)
			m.context = nc.(ContextModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}
