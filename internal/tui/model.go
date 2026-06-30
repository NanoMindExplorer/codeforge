package tui

import (
    "fmt"
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    "github.com/codeforge/tui/internal/config"
    "github.com/codeforge/tui/internal/git"
    "github.com/codeforge/tui/internal/provider"
    "github.com/codeforge/tui/internal/tool"
)

type Model struct {
    cfg         *config.Config
    providerReg *provider.Registry
    toolReg     *tool.Registry
    gitRepo     *git.Repo
    workdir     string

    width      int
    height     int
    activePane Pane
    mode       Mode

    chat    ChatModel
    diff    DiffModel
    context ContextModel
    status  StatusBarModel
    command CommandModel

    streamCh     <-chan provider.StreamToken
    streamInTok  int
    streamOutTok int

    quitting    bool
    err         error
    startTime   time.Time
    totalCost   float64
    totalTokens int
}

type Pane int

const (
    PaneChat Pane = iota
    PaneDiff
    PaneContext
)

type Mode int

const (
    ModeNormal Mode = iota
    ModeInsert
    ModeCommand
)

func New(cfg *config.Config, provReg *provider.Registry, toolReg *tool.Registry, repo *git.Repo, workdir string) Model {
    m := Model{
        cfg:         cfg,
        providerReg: provReg,
        toolReg:     toolReg,
        gitRepo:     repo,
        workdir:     workdir,
        activePane:  PaneChat,
        mode:        ModeNormal,
        startTime:   time.Now(),
        chat:        NewChatModel(provReg, toolReg, repo, workdir),
        diff:        NewDiffModel(),
        context:     NewContextModel(workdir),
        status:      NewStatusBarModel(),
        command:     NewCommandModel(),
    }
    return m
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(m.chat.Init(), m.context.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        mainHeight := msg.Height - 2
        if mainHeight < 5 {
            mainHeight = 5
        }
        chatW := msg.Width * 50 / 100
        diffW := msg.Width * 30 / 100
        ctxW := msg.Width - chatW - diffW - 4
        if chatW < 20 {
            chatW = 20
        }
        if diffW < 10 {
            diffW = 10
        }
        if ctxW < 10 {
            ctxW = 10
        }
        m.chat.SetSize(chatW, mainHeight)
        m.diff.SetSize(diffW, mainHeight)
        m.context.SetSize(ctxW, mainHeight)
        m.status.SetSize(msg.Width)
        m.command.SetSize(msg.Width, mainHeight)

    case tea.KeyMsg:
        // Global keys
        switch msg.String() {
        case "ctrl+c":
            m.quitting = true
            return m, tea.Quit
        case "ctrl+l":
            return m, nil
        }

        // Command mode
        if m.mode == ModeCommand {
            newCmd, cmd := m.command.Update(msg)
            m.command = newCmd.(CommandModel)
            if cmd != nil {
                cmds = append(cmds, cmd)
            }
            if m.command.Done {
                action := m.command.FinalValue
                m.command = NewCommandModel()
                m.mode = ModeNormal
                if action == "quit" || action == "q" {
                    m.quitting = true
                    return m, tea.Quit
                }
                cmd := m.executeSlashCommand(action)
                if cmd != nil {
                    cmds = append(cmds, cmd)
                }
            }
            return m, tea.Batch(cmds...)
        }

        // Mode-specific handling
        switch m.mode {
        case ModeNormal:
            switch msg.String() {
            case "i":
                m.mode = ModeInsert
                return m, nil
            case ":":
                m.mode = ModeCommand
                m.command.Activate()
                return m, nil
            case "/":
                m.mode = ModeCommand
                m.command.ActivateWithPrefix("/")
                return m, nil
            case "1":
                m.activePane = PaneChat
            case "2":
                m.activePane = PaneDiff
            case "3":
                m.activePane = PaneContext
            case "tab":
                m.activePane = (m.activePane + 1) % 3
            case "shift+tab":
                m.activePane = (m.activePane + 2) % 3
            case "q":
                m.quitting = true
                return m, tea.Quit
            case "?":
                m.chat.AddSystemMessage(helpText())
                return m, nil
            case "enter":
                if !m.chat.streaming {
                    cmd := m.chat.Submit()
                    cmds = append(cmds, cmd)
                }
                return m, tea.Batch(cmds...)
            case "esc":
                m.mode = ModeNormal
                return m, nil
            }

        case ModeInsert:
            keyStr := msg.String()
            switch keyStr {
            case "esc":
                m.mode = ModeNormal
                return m, nil
            case "enter":
                if m.chat.streaming {
                    return m, nil
                }
                cmd := m.chat.Submit()
                cmds = append(cmds, cmd)
                return m, tea.Batch(cmds...)
            case "backspace":
                m.chat.Backspace()
                return m, nil
            case "tab":
                m.chat.TypeText("\t")
                return m, nil
            }

            // Handle all KeyRunes (termasuk space, letters, numbers, symbols)
            if msg.Type == tea.KeyRunes {
                m.chat.TypeText(string(msg.Runes))
                return m, nil
            }

            // Handle single-char key strings (some terminals send space this way)
            if len(keyStr) == 1 {
                m.chat.TypeText(keyStr)
                return m, nil
            }

            // Ignore other keys (ctrl+shift combos, etc)
            return m, nil
        }

        // Forward to active pane
        switch m.activePane {
        case PaneChat:
            newChat, cmd := m.chat.Update(msg)
            m.chat = newChat.(ChatModel)
            if cmd != nil {
                cmds = append(cmds, cmd)
            }
        case PaneDiff:
            newDiff, cmd := m.diff.Update(msg)
            m.diff = newDiff.(DiffModel)
            if cmd != nil {
                cmds = append(cmds, cmd)
            }
        case PaneContext:
            newCtx, cmd := m.context.Update(msg)
            m.context = newCtx.(ContextModel)
            if cmd != nil {
                cmds = append(cmds, cmd)
            }
        }

    case StreamTickMsg:
        newChat, cmd := m.chat.Update(msg)
        m.chat = newChat.(ChatModel)
        if cmd != nil {
            cmds = append(cmds, cmd)
        }
        if msg.InputTokens > 0 || msg.OutputTokens > 0 {
            m.totalTokens += msg.InputTokens + msg.OutputTokens
            m.totalCost += calculateCost(msg.InputTokens, msg.OutputTokens, "claude")
        }
        if !msg.Done && m.streamCh != nil {
            cmds = append(cmds, pumpStream(m.streamCh))
        }
        if msg.Done {
            m.streamCh = nil
        }

    case StreamOpenedMsg:
        m.streamCh = msg.Ch
        if msg.FirstToken.Text != "" || msg.FirstToken.Done || msg.FirstToken.Error != nil {
            newChat, cmd := m.chat.Update(StreamTickMsg{
                Text:         msg.FirstToken.Text,
                Done:         msg.FirstToken.Done,
                InputTokens:  msg.FirstToken.InputTokens,
                OutputTokens: msg.FirstToken.OutputTokens,
                Error:        msg.FirstToken.Error,
            })
            m.chat = newChat.(ChatModel)
            if cmd != nil {
                cmds = append(cmds, cmd)
            }
            if msg.FirstToken.InputTokens > 0 || msg.FirstToken.OutputTokens > 0 {
                m.totalTokens += msg.FirstToken.InputTokens + msg.FirstToken.OutputTokens
                m.totalCost += calculateCost(msg.FirstToken.InputTokens, msg.FirstToken.OutputTokens, "claude")
            }
        }
        if msg.Ch != nil && !msg.FirstToken.Done {
            cmds = append(cmds, pumpStream(msg.Ch))
        }

    case errMsg:
        m.err = msg.err
        m.chat.AddSystemMessage(fmt.Sprintf("Error: %v", msg.err))
    }

    m.status.Provider = m.providerReg.CurrentName()
    m.status.Mode = modeString(m.mode)
    m.status.Cost = m.totalCost
    m.status.Tokens = m.totalTokens
    if m.gitRepo != nil {
        if branch, err := m.gitRepo.Branch(); err == nil {
            m.status.Branch = branch
        }
    }
    m.status.Workdir = m.workdir
    m.chat.mode = m.mode

    return m, tea.Batch(cmds...)
}

func (m Model) View() string {
    if m.quitting {
        return "Goodbye!\n"
    }
    topBar := m.status.ViewTop()
    chatView := m.chat.View()
    diffView := m.diff.View()
    ctxView := m.context.View()

    chatStyle := paneStyle(m.activePane == PaneChat)
    diffStyle := paneStyle(m.activePane == PaneDiff)
    ctxStyle := paneStyle(m.activePane == PaneContext)

    mainRow := lipgloss.JoinHorizontal(lipgloss.Top,
        chatStyle.Render(chatView),
        diffStyle.Render(diffView),
        ctxStyle.Render(ctxView),
    )

    bottomBar := m.status.ViewBottom()

    if m.mode == ModeCommand {
        cmdOverlay := m.command.View()
        return lipgloss.JoinVertical(lipgloss.Left,
            topBar,
            mainRow,
            cmdOverlay,
            bottomBar,
        )
    }

    return lipgloss.JoinVertical(lipgloss.Left,
        topBar,
        mainRow,
        bottomBar,
    )
}

func paneStyle(active bool) lipgloss.Style {
    borderColor := lipgloss.Color("#334155")
    if active {
        borderColor = lipgloss.Color("#06B6D4")
    }
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(borderColor).
        Padding(0, 1)
}

func modeString(m Mode) string {
    switch m {
    case ModeNormal:
        return "NORMAL"
    case ModeInsert:
        return "INSERT"
    case ModeCommand:
        return "COMMAND"
    }
    return "?"
}

func helpText() string {
    return `CodeForge TUI v0.1.0-alpha - Quick Reference
Created by NanoMind · 2026 · Apache 2.0

MODES:
  i         Enter INSERT mode (type messages)
  Esc       Back to NORMAL mode
  :         Open command palette
  /         Open slash command palette

NAVIGATION:
  1 / 2 / 3   Focus Chat / Diff / Context pane
  Tab         Cycle focus between panes
  j / k       Scroll down / up
  g / G       Top / bottom

CHAT:
  Enter       Send message (in INSERT mode)
  Ctrl+C      Interrupt AI operation

COMMANDS:
  /help       Show this help
  /about      Show about (NanoMind)
  /version    Show version
  /provider   List/switch provider
  /model      List models
  /cost       Show cost summary
  /status     Show git status
  /commit     Create git commit
  /clear      Clear chat history
  /quit       Exit

"Building the future of terminal AI coding, one keystroke at a time."
                    — NanoMind, 2026`
}

func aboutText() string {
    return `CodeForge TUI v0.1.0-alpha
Created by NanoMind - 2026
License: Apache 2.0
Language: Go 1.25 + Bubble Tea

"Building the future of terminal AI coding, one keystroke at a time."
                    — NanoMind, 2026`
}

func versionText() string {
    return `CodeForge TUI v0.1.0-alpha
Author:  NanoMind
Year:    2026
License: Apache 2.0

Project by NanoMind. Permanently etched in history.`
}

func calculateCost(inputTokens, outputTokens int, providerName string) float64 {
    switch providerName {
    case "claude":
        return float64(inputTokens)*3.0/1_000_000 + float64(outputTokens)*15.0/1_000_000
    default:
        return 0 // Gemini free tier
    }
}

func (m *Model) executeSlashCommand(input string) tea.Cmd {
    input = strings.TrimSpace(input)
    if input == "" {
        return nil
    }
    if strings.HasPrefix(input, "/") {
        parts := strings.Fields(input[1:])
        if len(parts) == 0 {
            return nil
        }
        cmd := parts[0]
        args := parts[1:]

        switch cmd {
        case "help", "h":
            m.chat.AddSystemMessage(helpText())
        case "about", "a":
            m.chat.AddSystemMessage(aboutText())
        case "version", "v":
            m.chat.AddSystemMessage(versionText())
        case "provider", "p":
            if len(args) == 0 {
                names := m.providerReg.List()
                var sb strings.Builder
                sb.WriteString("Available providers:\n")
                current := m.providerReg.CurrentName()
                for _, name := range names {
                    marker := "  "
                    if name == current {
                        marker = "* "
                    }
                    sb.WriteString(fmt.Sprintf("  %s%s\n", marker, name))
                }
                m.chat.AddSystemMessage(sb.String())
            } else {
                name := args[0]
                if err := m.providerReg.Switch(name); err != nil {
                    m.chat.AddSystemMessage(fmt.Sprintf("Error: %v", err))
                } else {
                    m.chat.AddSystemMessage(fmt.Sprintf("Switched to: %s", name))
                }
            }
        case "model", "m":
            if len(args) == 0 {
                if cur, err := m.providerReg.Current(); err == nil {
                    var sb strings.Builder
                    sb.WriteString("Available models:\n")
                    for _, mi := range cur.Models() {
                        sb.WriteString(fmt.Sprintf("  %s - %s (ctx: %d)\n", mi.ID, mi.Name, mi.ContextWindow))
                    }
                    m.chat.AddSystemMessage(sb.String())
                }
            } else {
                modelName := args[0]
                m.chat.AddSystemMessage("Model switch: " + modelName + " (next session will use this model)")
            }
        case "cost", "c":
            m.chat.AddSystemMessage(fmt.Sprintf(
                "Session Cost\n  Tokens: %d\n  Cost: $%.4f\n  Duration: %s",
                m.totalTokens, m.totalCost, time.Since(m.startTime).Round(time.Second),
            ))
        case "status", "s":
            if m.gitRepo != nil {
                status, err := m.gitRepo.Status()
                if err != nil {
                    m.chat.AddSystemMessage(fmt.Sprintf("Git: %v", err))
                } else {
                    branch, _ := m.gitRepo.Branch()
                    m.chat.AddSystemMessage(fmt.Sprintf("Branch: %s\n%s", branch, status))
                }
            }
        case "commit":
            if m.gitRepo == nil {
                m.chat.AddSystemMessage("No git repo")
                return nil
            }
            if err := m.gitRepo.AddAll(); err != nil {
                m.chat.AddSystemMessage(fmt.Sprintf("Add: %v", err))
                return nil
            }
            msg := git.GenerateCommitMessage("feat", "", "AI-assisted changes via CodeForge TUI")
            hash, err := m.gitRepo.Commit(msg)
            if err != nil {
                m.chat.AddSystemMessage(fmt.Sprintf("Commit: %v", err))
                return nil
            }
            m.chat.AddSystemMessage(fmt.Sprintf("Committed: %s\nMessage: %s", hash, msg))
        case "quit", "q":
            m.quitting = true
            return tea.Quit
        case "clear":
            m.chat.Clear()
        default:
            m.chat.AddSystemMessage(fmt.Sprintf("Unknown: /%s", cmd))
        }
        return nil
    }

    m.chat.SetInput(input)
    return m.chat.Submit()
}

type errMsg struct{ err error }

type StreamOpenedMsg struct {
    Ch         <-chan provider.StreamToken
    FirstToken provider.StreamToken
}

func pumpStream(ch <-chan provider.StreamToken) tea.Cmd {
    if ch == nil {
        return nil
    }
    return func() tea.Msg {
        token, ok := <-ch
        if !ok {
            return StreamTickMsg{Done: true}
        }
        return StreamTickMsg{
            Text:         token.Text,
            Done:         token.Done,
            InputTokens:  token.InputTokens,
            OutputTokens: token.OutputTokens,
            Error:        token.Error,
        }
    }
}

type StreamTickMsg struct {
    Text         string
    Done         bool
    InputTokens  int
    OutputTokens int
    Error        error
}

type DiffUpdateMsg struct {
    Content string
}

type ContextUpdateMsg struct {
    Files []string
}
