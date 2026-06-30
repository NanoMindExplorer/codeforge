// CodeForge TUI - Terminal AI Coding Companion
// Author: NanoMind (2026)
// License: Apache 2.0
package main

import (
    "fmt"
    "os"
    "path/filepath"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/codeforge/tui/internal/config"
    "github.com/codeforge/tui/internal/git"
    "github.com/codeforge/tui/internal/provider"
    "github.com/codeforge/tui/internal/tool"
    "github.com/codeforge/tui/internal/tui"
)

const (
    ProjectName    = "CodeForge TUI"
    ProjectVersion = "0.1.0-alpha"
    ProjectAuthor  = "NanoMind"
    ProjectYear    = "2026"
    ProjectLicense = "Apache 2.0"
)

func main() {
    workdir, err := os.Getwd()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    if len(os.Args) > 1 {
        if abs, err := filepath.Abs(os.Args[1]); err == nil {
            if info, err := os.Stat(abs); err == nil && info.IsDir() {
                workdir = abs
            }
        }
    }

    cfg, err := config.Load()
    if err != nil {
        cfg = config.Default()
    }
    _ = config.SaveExample()

    // Setup provider registry - PRIORITAS GEMINI (gratis)
    provReg := provider.NewRegistry()

    geminiKey := os.Getenv("GEMINI_API_KEY")
    if geminiKey != "" {
        geminiProv := provider.NewGeminiProvider(geminiKey, "gemini-2.5-flash")
        _ = provReg.Register(geminiProv)
        fmt.Fprintf(os.Stderr, "✓ Gemini provider registered (free tier)\n")
    }

    claudeKey := os.Getenv("ANTHROPIC_API_KEY")
    if claudeKey != "" {
        claudeProv := provider.NewClaudeProvider(claudeKey, "claude-sonnet-4-20250514")
        _ = provReg.Register(claudeProv)
        fmt.Fprintf(os.Stderr, "✓ Claude provider registered\n")
    }

    // Fallback jika tidak ada API key sama sekali
    if _, err := provReg.Current(); err != nil {
        claudeProv := provider.NewClaudeProvider("", "claude-sonnet-4-20250514")
        _ = provReg.Register(claudeProv)
    }

    // Default: Gemini jika ada, selain itu pakai config
    if geminiKey != "" {
        _ = provReg.Switch("gemini")
    } else if cfg.DefaultProvider != "" {
        _ = provReg.Switch(cfg.DefaultProvider)
    }

    toolReg := tool.NewRegistry(workdir)

    repo, err := git.Open(workdir)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: git: %v\n", err)
    }

    // Validate provider config
    if cur, err := provReg.Current(); err == nil {
        if err := cur.ValidateConfig(); err != nil {
            fmt.Fprintf(os.Stderr, "\n⚠️  Provider config issue: %v\n", err)
            fmt.Fprintf(os.Stderr, "   Get free Gemini key: https://aistudio.google.com/apikey\n")
            fmt.Fprintf(os.Stderr, "   Set: export GEMINI_API_KEY=your-key\n\n")
        }
    }

    model := tui.New(cfg, provReg, toolReg, repo, workdir)
    printBanner()

    p := tea.NewProgram(model, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func printBanner() {
    fmt.Printf(`
╔══════════════════════════════════════════════════════════════╗
║   CodeForge TUI v%s  |  by %s  |  %s              ║
║   Terminal AI Coding Companion  |  Apache 2.0                ║
║   Gemini • Claude • Multi-Provider Plug & Play               ║
╚══════════════════════════════════════════════════════════════╝
`, ProjectVersion, ProjectAuthor, ProjectYear)
}
