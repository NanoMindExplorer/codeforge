package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
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
