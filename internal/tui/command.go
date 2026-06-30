package tui

import (
    "fmt"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type CommandModel struct {
    width      int
    height     int
    active     bool
    input      string
    cursor     int
    Done       bool
    FinalValue string
}

func NewCommandModel() CommandModel {
    return CommandModel{}
}

func (c *CommandModel) SetSize(w, h int) {
    c.width = w
    c.height = h
}

func (c CommandModel) Init() tea.Cmd { return nil }

func (c *CommandModel) Activate() {
    c.active = true
    c.input = ""
    c.Done = false
}

func (c *CommandModel) ActivateWithPrefix(prefix string) {
    c.active = true
    c.input = prefix
    c.Done = false
}

func (c CommandModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if !c.active {
        return c, nil
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "enter":
            c.Done = true
            c.FinalValue = c.input
            c.active = false
            return c, nil
        case "esc":
            c.Done = true
            c.FinalValue = ""
            c.active = false
            return c, nil
        case "backspace":
            if len(c.input) > 0 {
                c.input = c.input[:len(c.input)-1]
            }
            return c, nil
        default:
            if msg.Type == tea.KeyRunes {
                c.input += string(msg.Runes)
            }
            return c, nil
        }
    }
    return c, nil
}

func (c CommandModel) View() string {
    if !c.active {
        return ""
    }

    style := lipgloss.NewStyle().
        Background(lipgloss.Color("#0F172A")).
        Foreground(lipgloss.Color("#06B6D4")).
        Width(c.width).
        Padding(0, 1)

    prompt := ":"
    if strings.HasPrefix(c.input, "/") {
        prompt = ""
    }

    content := prompt + c.input + "_"

    hint := ""
    if strings.HasPrefix(c.input, "/") && len(c.input) > 1 {
        partial := c.input[1:]
        hints := []struct{ cmd, desc string }{
            {"help", "Show help"},
            {"about", "Show about (NanoMind)"},
            {"version", "Show version"},
            {"provider", "List/switch provider"},
            {"cost", "Show cost"},
            {"status", "Git status"},
            {"commit", "Create commit"},
            {"clear", "Clear chat"},
            {"quit", "Exit"},
        }
        var matches []string
        for _, h := range hints {
            if strings.HasPrefix(h.cmd, partial) {
                matches = append(matches, fmt.Sprintf("/%s - %s", h.cmd, h.desc))
            }
        }
        if len(matches) > 0 {
            hint = "\n" + strings.Join(matches, "  |  ")
        }
    }

    return style.Render(content + hint)
}
