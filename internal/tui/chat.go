package tui

import (
    "context"
    "fmt"
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    "github.com/codeforge/tui/internal/git"
    "github.com/codeforge/tui/internal/provider"
    "github.com/codeforge/tui/internal/tool"
)

type ChatModel struct {
    providerReg *provider.Registry
    toolReg     *tool.Registry
    gitRepo     *git.Repo
    workdir     string

    width     int
    height    int
    input     string
    lines     []string
    cursor    int
    streaming bool
    mode      Mode

    messages []provider.Message
}

func NewChatModel(provReg *provider.Registry, toolReg *tool.Registry, repo *git.Repo, workdir string) ChatModel {
    return ChatModel{
        providerReg: provReg,
        toolReg:     toolReg,
        gitRepo:     repo,
        workdir:     workdir,
        lines: []string{
            "CodeForge TUI v0.1.0-alpha",
            "Created by NanoMind - 2026",
            "Type /help for commands, or 'i' to start chatting.",
            "",
        },
    }
}

func (c ChatModel) Init() tea.Cmd { return nil }

func (c *ChatModel) SetSize(w, h int) {
    c.width = w
    c.height = h
}

func (c *ChatModel) TypeText(s string) {
    c.input += s
}

func (c *ChatModel) Backspace() {
    if len(c.input) > 0 {
        c.input = c.input[:len(c.input)-1]
    }
}

func (c *ChatModel) SetInput(s string) {
    c.input = s
}

func (c *ChatModel) Submit() tea.Cmd {
    if c.input == "" || c.streaming {
        return nil
    }
    userMsg := c.input
    c.input = ""
    c.AddUserMessage(userMsg)
    c.streaming = true

    msgs := make([]provider.Message, len(c.messages))
    copy(msgs, c.messages)

    prov := c.providerReg
    return func() tea.Msg {
        p, err := prov.Current()
        if err != nil {
            return errMsg{err: fmt.Errorf("no provider: %w", err)}
        }

        ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
        defer cancel()

        req := provider.CompletionRequest{
            Messages:  msgs,
            MaxTokens: 4096,
            System:    "You are CodeForge TUI, an AI pair programming assistant created by NanoMind. Be concise and helpful.",
        }

        ch, err := p.Stream(ctx, req)
        if err != nil {
            return errMsg{err: err}
        }

        var firstToken provider.StreamToken
        var firstOk bool
        select {
        case t, ok := <-ch:
            if !ok {
                return StreamOpenedMsg{Ch: nil}
            }
            firstToken = t
            firstOk = true
        case <-time.After(30 * time.Second):
            return errMsg{err: fmt.Errorf("stream timeout")}
        }

        if firstOk {
            return StreamOpenedMsg{
                Ch:         ch,
                FirstToken: firstToken,
            }
        }
        return StreamOpenedMsg{Ch: nil}
    }
}

func (c *ChatModel) AddUserMessage(text string) {
    c.messages = append(c.messages, provider.Message{
        Role:    "user",
        Content: text,
    })
    c.lines = append(c.lines, fmt.Sprintf("> %s", text), "")
}

func (c *ChatModel) AddSystemMessage(text string) {
    for _, line := range strings.Split(text, "\n") {
        c.lines = append(c.lines, "  "+line)
    }
    c.lines = append(c.lines, "")
}

func (c *ChatModel) AddAssistantMessage(text string) {
    c.messages = append(c.messages, provider.Message{
        Role:    "assistant",
        Content: text,
    })
}

func (c *ChatModel) AppendStreamingText(text string) {
    if len(c.lines) == 0 || strings.HasPrefix(c.lines[len(c.lines)-1], "> ") {
        c.lines = append(c.lines, "")
    }
    lastIdx := len(c.lines) - 1
    if c.lines[lastIdx] == "" {
        c.lines[lastIdx] = text
    } else {
        c.lines[lastIdx] += text
    }
}

func (c *ChatModel) FinalizeStreaming(fullText string) {
    c.streaming = false
    c.AddAssistantMessage(fullText)
    c.lines = append(c.lines, "")
}

func (c *ChatModel) Clear() {
    c.lines = []string{"Chat cleared.", ""}
    c.messages = nil
}

func (c ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case StreamTickMsg:
        if msg.Text != "" {
            c.AppendStreamingText(msg.Text)
        }
        if msg.Done {
            var fullText string
            for i := len(c.lines) - 1; i >= 0; i-- {
                if c.lines[i] != "" && !strings.HasPrefix(c.lines[i], "> ") && !strings.HasPrefix(c.lines[i], "  ") {
                    fullText = c.lines[i]
                    break
                }
            }
            c.FinalizeStreaming(fullText)
            if msg.InputTokens > 0 || msg.OutputTokens > 0 {
                c.lines = append(c.lines, fmt.Sprintf("  [tokens: %d in / %d out]", msg.InputTokens, msg.OutputTokens), "")
            }
        }
        if msg.Error != nil {
            c.lines = append(c.lines, fmt.Sprintf("  ERROR: %v", msg.Error), "")
            c.streaming = false
        }
        return c, nil

    case tea.KeyMsg:
        switch msg.String() {
        case "j", "down":
            c.cursor++
        case "k", "up":
            if c.cursor > 0 {
                c.cursor--
            }
        case "g":
            c.cursor = 0
        case "G":
            c.cursor = len(c.lines)
        }
    }
    return c, nil
}

func (c ChatModel) View() string {
    var content strings.Builder
    content.WriteString("Chat\n")
    content.WriteString(strings.Repeat("-", max(0, c.width-4)) + "\n")

    visibleHeight := c.height - 5
    if visibleHeight < 3 {
        visibleHeight = 3
    }

    start := 0
    if len(c.lines) > visibleHeight {
        start = len(c.lines) - visibleHeight
    }
    for i := start; i < len(c.lines); i++ {
        content.WriteString(c.lines[i] + "\n")
    }

    content.WriteString("\n" + strings.Repeat("-", max(0, c.width-4)) + "\n")
    inputLine := c.input
    if c.streaming {
        inputLine = "[streaming...]"
    } else if c.mode == ModeInsert {
        inputLine += "_"
    } else if inputLine == "" {
        inputLine = "(press 'i' to type, '/' for commands)"
    }
    content.WriteString("> " + inputLine + "\n")

    style := lipgloss.NewStyle().
        Width(c.width).
        Height(c.height).
        Foreground(lipgloss.Color("#E2E8F0"))

    return style.Render(content.String())
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
