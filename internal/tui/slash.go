package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codeforge/tui/internal/bgtask"
	"github.com/codeforge/tui/internal/onboarding"
	"github.com/codeforge/tui/internal/tool"
	"github.com/codeforge/tui/internal/ui/components"
)

func isImmediateSlash(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	switch cmd {
	case "/act", "/plan", "/review", "/git", "/gh", "/session", "/agent", "/settings", "/clear", "/help":
		return true
	default:
		return false
	}
}

func (m *Model) executeSlashCommand(input string) tea.Cmd {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	raw := input
	if strings.HasPrefix(raw, "/") {
		raw = raw[1:]
	}
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return nil
	}
	cmd := strings.ToLower(parts[0])
	args := parts[1:]
	argStr := strings.Join(args, " ")

	switch cmd {
	case "act", "a":
		m.chat.AddSystemMessage("Mode beralih ke ACT. Ketik instruksi Anda.")
		m.setSessionMode(tool.SessionYolo)
		m.syncWriteMode()
		if argStr != "" {
			m.toast = components.NewToast("Mengeksekusi perintah...", "info", 3*time.Second)
			return m.chat.SubmitAgent(argStr)
		}

	case "plan":
		m.chat.AddSystemMessage("Mode beralih ke PLAN. Ketik instruksi riset Anda.")
		m.setSessionMode(tool.SessionBuild)
		m.syncWriteMode()
		if argStr != "" {
			m.toast = components.NewToast("Menyusun plan...", "info", 3*time.Second)
			return m.chat.SubmitAgent(argStr)
		}

	case "review", "view-plan", "plan-view":
		m.mode = ModePlanReview

	case "git":
		if len(args) == 0 {
			m.chat.AddSystemMessage("Penggunaan: /git <commit|push|pull|undo>")
			return nil
		}
		subcmd := args[0]
		switch subcmd {
		case "commit":
			msg := "Update"
			if len(args) > 1 {
				msg = strings.Join(args[1:], " ")
			}
			m.toast = components.NewToast("git commit...", "info", 3*time.Second)
			if _, err := bgtask.Global.Start(m.workdir, fmt.Sprintf("git add . && git commit -m %q", msg)); err != nil {
				m.chat.AddSystemMessage("Error: " + err.Error())
			}
		case "push":
			m.toast = components.NewToast("Menjalankan: git push...", "info", 3*time.Second)
			if _, err := bgtask.Global.Start(m.workdir, "git push"); err != nil {
				m.chat.AddSystemMessage("Error: " + err.Error())
			}
		case "pull":
			m.toast = components.NewToast("Menjalankan: git pull...", "info", 3*time.Second)
			if _, err := bgtask.Global.Start(m.workdir, "git pull"); err != nil {
				m.chat.AddSystemMessage("Error: " + err.Error())
			}
		case "undo":
			m.chat.AddSystemMessage("Mengembalikan 1 commit terakhir (soft)...")
			m.toast = components.NewToast("Menjalankan: git reset --soft HEAD~1...", "info", 3*time.Second)
			if _, err := bgtask.Global.Start(m.workdir, "git reset --soft HEAD~1"); err != nil {
				m.chat.AddSystemMessage("Error: " + err.Error())
			}
		default:
			m.chat.AddSystemMessage("Perintah git tidak dikenal: " + subcmd)
		}

	case "gh", "github":
		if len(args) == 0 {
			m.chat.AddSystemMessage("Penggunaan: /gh <pr|issue|status>")
			return nil
		}
		subcmd := args[0]
		switch subcmd {
		case "pr":
			m.toast = components.NewToast("Menjalankan: gh pr create --fill...", "info", 3*time.Second)
			if _, err := bgtask.Global.Start(m.workdir, "gh pr create --fill"); err != nil {
				m.chat.AddSystemMessage("Error: " + err.Error())
			}
		case "issue":
			m.toast = components.NewToast("Menjalankan: gh issue create...", "info", 3*time.Second)
			if _, err := bgtask.Global.Start(m.workdir, "gh issue create"); err != nil {
				m.chat.AddSystemMessage("Error: " + err.Error())
			}
		case "status":
			m.toast = components.NewToast("Menjalankan: gh pr status...", "info", 3*time.Second)
			if _, err := bgtask.Global.Start(m.workdir, "gh pr status"); err != nil {
				m.chat.AddSystemMessage("Error: " + err.Error())
			}
		default:
			m.chat.AddSystemMessage("Perintah GitHub tidak dikenal: " + subcmd)
		}

	case "session":
		if len(args) == 0 {
			m.chat.AddSystemMessage("Penggunaan: /session <new|list|rename|resume|dashboard>")
			return nil
		}
		subcmd := args[0]
		switch subcmd {
		case "new":
			if m.session != nil {
				m.saveSession()
				m.session = nil
			}
			m.chat.Clear()
			m.context = NewContextModel(m.workdir)
			m.diff = NewDiffModel()
			if healthy := onboarding.ProviderHealthy(m.providerReg); healthy {
				m.chat.AddSystemMessage(onboarding.StatusCard(m.cfg, m.providerReg.CurrentName(), "", healthy))
			}
			m.toast = components.NewToast("Sesi baru dimulai", "success", 2*time.Second)
		case "list", "dashboard", "resume":
			m.mode = ModeSessionPick
			m.sessions.Open(m.workdir)
			return nil
		case "rename":
			if len(args) > 1 {
				title := strings.Join(args[1:], " ")
				if m.session != nil {
					m.session.Title = title
					m.toast = components.NewToast("Sesi diubah nama", "success", 2*time.Second)
				}
			}
		}

	case "agent":
		if len(args) == 0 {
			m.chat.AddSystemMessage("Penggunaan: /agent <subagents|tasks|memory|skills>")
			return nil
		}
		subcmd := strings.ToLower(args[0])
		switch subcmd {
		case "subagents", "subs":
			m.chat.AddSystemMessage(tool.SubJobs.Summary())
		case "tasks", "bg":
			m.chat.AddSystemMessage(bgtask.Global.Summary())
		case "memory", "mem":
			m.chat.AddSystemMessage("Memory integration temporarily handled via @memory.")
		case "skills":
			m.chat.AddSystemMessage("Skill system active.")
		}

	case "settings", "config", "setup", "provider", "model":
		m.mode = ModeSettings
		return nil

	case "clear":
		m.chat.Clear()

	case "help", "h", "?":
		m.chat.AddSystemMessage(helpText())

	case "read", "ls", "grep", "run", "explain", "fix":
		m.chat.AddSystemMessage("ℹ️ Tip: Anda tidak perlu menggunakan /" + cmd + " lagi! Cukup sebutkan @nama_file atau gunakan kata-kata natural di chat, dan saya akan mengeksekusi alatnya secara otomatis.")

	case "theme", "vim-mode", "compact-mode":
		m.chat.AddSystemMessage("ℹ️ Tip: Pengaturan UI ini telah dipindahkan! Tekan `Ctrl+K` untuk membuka Command Palette, atau gunakan /settings.")

	default:
		// Unknown commands pass gracefully
		m.chat.AddSystemMessage("⚠️ Perintah tidak dikenal: /" + cmd + ". Ketik /help untuk bantuan.")
	}
	return nil
}

// handleGitHubCommand runs GitHub ops from slash commands.
// Forms:
//
//	/gh auth|repo|push|pull|log|branch <name>
//	/gh pr list|view [n]|create <title> [| body]|merge <n>|checks [n]
//	/gh issue list|view <n>|create <title> [| body]
//	/pr … and /issue … are aliases without the leading "pr"/"issue" word duplicated.

func modeString(m Mode) string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeCommand:
		return "COMMAND"
	case ModePalette:
		return "COMMAND"
	case ModeReview:
		return "REVIEW"
	case ModeFilePick:
		return "INSERT"
	case ModeAskUser:
		return "ASK"
	case ModePermAsk:
		return "PERM"
	}
	return "?"
}

var slashCommands = []string{
	"/act", "/plan", "/review", "/git", "/gh", "/session", "/agent", "/settings", "/clear", "/help",
}

func autocomplete(input string) string {
	for _, cmd := range slashCommands {
		if strings.HasPrefix(cmd, input) {
			return cmd + " "
		}
	}
	return ""
}

func helpText() string {
	return `✨ **CodeForge Core Commands**

    /act [prompt]   Langsung eksekusi perubahan kode
    /plan [prompt]  Riset mendalam & buat draf rencana
    /review         Lihat draf rencana saat ini
    
    /git <args>     Manajemen Repositori (commit, push, pull, undo)
    /gh <args>      Integrasi GitHub (pr, issue, status)
    
    /session <args> Manajemen Sesi (new, list, rename)
    /agent <args>   Manajemen Subagen & Background Tasks (tasks, subs)
    
    /settings       Buka Menu Pengaturan Lengkap (API, Provider, Model)
    /clear          Bersihkan layar chat saat ini

💡 **TIPS:**
- Gunakan **@** untuk menyebut file, folder, atau tautan web (cth: @src/main.go)
- Tekan **Ctrl+K** untuk membuka Command Palette (Ubah Tema, Mode Vim, dll)
`
}
func aboutText() string {
	return `CodeForge TUI v1.9.3
Created by NanoMind — 2026 — Apache 2.0

Grok Build TUI–compatible (Phases 1–9 + G1–G10 + W1–W4):
  blocks · input · themes · sessions · design plan
  permissions/hooks · todos/tasks · ACP + x.ai/* extensions
  Grok 4.5 · native thinking · Landlock · skills · personas
  pager.toml · /setup · /doctor · release gate
See docs/PAGER.md · docs/REASONING.md · docs/RELEASE_GATE.md
`
}

// tuiAboutVersion extracts X.Y.Z from aboutText first line.
func tuiAboutVersion() string {
	lines := strings.Split(aboutText(), "\n")
	if len(lines) == 0 {
		return ""
	}
	for _, w := range strings.Fields(lines[0]) {
		if strings.HasPrefix(w, "v") && strings.Count(w, ".") >= 2 {
			return strings.TrimPrefix(w, "v")
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
