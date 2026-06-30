package tui

import (
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type DiffModel struct {
    width   int
    height  int
    content string
}

func NewDiffModel() DiffModel {
    return DiffModel{
        content: "No changes yet.\n\nWhen the AI edits files,\ndiffs will appear here.",
    }
}

func (d DiffModel) Init() tea.Cmd { return nil }

func (d *DiffModel) SetSize(w, h int) {
    d.width = w
    d.height = h
}

func (d *DiffModel) SetContent(content string) {
    d.content = content
}

func (d DiffModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case DiffUpdateMsg:
        d.content = msg.Content
    }
    return d, nil
}

func (d DiffModel) View() string {
    if d.width < 10 {
        d.width = 10
    }
    if d.height < 5 {
        d.height = 5
    }
    var sb strings.Builder
    sb.WriteString("Diff\n")
    sb.WriteString(strings.Repeat("-", max(0, d.width-4)) + "\n\n")

    for _, line := range strings.Split(d.content, "\n") {
        var styled string
        switch {
        case strings.HasPrefix(line, "+"):
            styled = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render(line)
        case strings.HasPrefix(line, "-"):
            styled = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(line)
        case strings.HasPrefix(line, "@@"):
            styled = lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4")).Render(line)
        default:
            styled = line
        }
        sb.WriteString(styled + "\n")
    }

    style := lipgloss.NewStyle().
        Width(d.width).
        Height(d.height).
        Foreground(lipgloss.Color("#94A3B8"))

    return style.Render(sb.String())
}
