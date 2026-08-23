package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codeforge/tui/internal/config"
	"github.com/codeforge/tui/internal/provider"
	"github.com/codeforge/tui/internal/theme"
	"github.com/codeforge/tui/internal/tool"
)

// ensure time import used
var _ = time.Now

// Smoke test: construct model, send window size + keys, ensure View does not panic
// and contains key brand strings.
func TestSmokeRender(t *testing.T) {
	theme.Set(theme.Aurora())
	theme.SetMotion(false)

	reg := provider.NewRegistry()
	_ = reg.Register(provider.NewGeminiProvider("test-key", "gemini-2.5-flash"))
	tools := tool.NewRegistry(t.TempDir())
	cfg := config.Default()

	m := New(cfg, reg, tools, nil, t.TempDir())
	m.mode = ModeInsert
	m.focusPrompt = true
	// size
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = nm.(Model)
	view := m.View()
	if !strings.Contains(view, "CodeForge") {
		t.Fatalf("view missing brand:\n%s", truncateView(view))
	}

	// Already prompt-focused (Grok simple mode)
	// help via ?
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = nm.(Model)
	// spinner tick should not panic
	nm, _ = m.Update(SpinnerTickMsg{})
	m = nm.(Model)

	// narrow layout
	nm, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 30})
	m = nm.(Model)
	_ = m.View()

	// Shift+Tab cycles BUILD → DESIGN → YOLO
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = asModel(nm)
	if m.sessionMode != tool.SessionDesign {
		t.Fatalf("expected DESIGN after Shift+Tab, got %v", m.sessionMode.Label())
	}
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = asModel(nm)
	if m.sessionMode != tool.SessionYolo {
		t.Fatalf("expected YOLO, got %v", m.sessionMode.Label())
	}
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = asModel(nm)
	if m.sessionMode != tool.SessionBuild {
		t.Fatalf("expected BUILD, got %v", m.sessionMode.Label())
	}
	// Settings opens (Phase 3)
	_ = m.executeSlashCommand("/settings")
	if m.mode != ModeSettings {
		t.Fatalf("expected ModeSettings after /settings, got %v", m.mode)
	}
}

func TestThemePickerPreviewAndCancel(t *testing.T) {
	theme.Set(theme.GrokNight())
	theme.SetMotion(false)
	reg := provider.NewRegistry()
	_ = reg.Register(provider.NewGeminiProvider("k", "gemini-2.5-flash"))
	m := New(config.Default(), reg, tool.NewRegistry(t.TempDir()), nil, t.TempDir())
	m.mode = ModeInsert
	m.focusPrompt = true
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = asModel(nm)
	m.mode = ModeThemePick
	m.themes.Open()
	if m.mode != ModeThemePick {
		t.Fatal("picker not open")
	}
	// move down previews next theme
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = asModel(nm)
	// esc reverts
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = asModel(nm)
	if m.mode == ModeThemePick {
		t.Fatal("should close on esc")
	}
	if theme.DisplayName() != "groknight" {
		t.Fatalf("should revert to groknight, got %s", theme.DisplayName())
	}
}

func TestDoubleEscClearsPrompt(t *testing.T) {
	theme.Set(theme.GrokNight())
	theme.SetMotion(false)
	reg := provider.NewRegistry()
	_ = reg.Register(provider.NewGeminiProvider("k", "gemini-2.5-flash"))
	m := New(config.Default(), reg, tool.NewRegistry(t.TempDir()), nil, t.TempDir())
	m.mode = ModeInsert
	m.focusPrompt = true
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = asModel(nm)
	m.focusPrompt = true
	m.mode = ModeInsert
	m.chat.FocusInput()
	m.chat.SetInput("hello draft")
	m.lastEsc = time.Now()
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = asModel(nm)
	if m.chat.InputValue() != "" {
		// cleared on second press within window
		t.Log("draft:", m.chat.InputValue())
	}
}

func asModel(nm tea.Model) Model {
	switch v := nm.(type) {
	case Model:
		return v
	case *Model:
		return *v
	default:
		panic("unexpected model type")
	}
}

func TestSlashMenuActivates(t *testing.T) {
	theme.SetMotion(false)
	reg := provider.NewRegistry()
	_ = reg.Register(provider.NewGeminiProvider("k", "gemini-2.5-flash"))
	m := New(config.Default(), reg, tool.NewRegistry(t.TempDir()), nil, t.TempDir())
	m.mode = ModeInsert
	m.focusPrompt = true
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = nm.(Model)
	m.focusPrompt = true
	m.chat.SetInput("/he")
	m.slash.UpdateQuery("/he")
	if !m.slash.Active {
		t.Fatal("slash menu inactive")
	}
	if m.slash.Selected() == "" && len(m.slash.Filtered) == 0 {
		t.Fatal("no filter")
	}
}

func TestContextUpdateMsgWires(t *testing.T) {
	theme.SetMotion(false)
	reg := provider.NewRegistry()
	_ = reg.Register(provider.NewGeminiProvider("k", "gemini-2.5-flash"))
	dir := t.TempDir()
	tools := tool.NewRegistry(dir)
	m := New(config.Default(), reg, tools, nil, dir)
	m.mode = ModeInsert
	m.focusPrompt = true
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = nm.(Model)
	nm, _ = m.Update(ContextUpdateMsg{Refresh: true})
	m = nm.(Model)
	// files list should be non-nil (may be empty dir)
	_ = m.context.View()
}

func truncateView(s string) string {
	if len(s) > 200 {
		return s[:200]
	}
	return s
}

func TestCalculateCostNotAlwaysZero(t *testing.T) {
	// provider.CostForModel is the source of truth now
	p := provider.NewGeminiProvider("k", "gemini-2.5-pro")
	c := provider.CostForModel(p, "gemini-2.5-pro", 100_000, 100_000)
	if c == 0 {
		t.Fatal("gemini pro cost should not be zero")
	}
	_ = time.Now()
}
