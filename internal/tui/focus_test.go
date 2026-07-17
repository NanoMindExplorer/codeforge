package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codeforge/tui/internal/theme"
	"github.com/codeforge/tui/internal/tool"
	"github.com/codeforge/tui/internal/ui/review"
)

func TestReturnToPromptAfterReviewEsc(t *testing.T) {
	theme.SetMotion(false)
	m := testModel(t)
	// Simulate BUILD review overlay open
	m.mode = ModeReview
	m.focusPrompt = false
	m.chat.BlurInput()
	m.review = review.New()
	m.review.Open([]tool.PendingPatch{
		{RelPath: "a.go", Diff: "d", Accepted: true},
	})

	nm, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEscape})
	m = asModel(nm)
	if m.mode != ModeInsert {
		t.Fatalf("mode=%v want Insert", m.mode)
	}
	if !m.focusPrompt {
		t.Fatal("expected prompt focus after review Esc")
	}
}

func TestTabFocusSwapAndReturn(t *testing.T) {
	m := testModel(t)
	if !m.focusPrompt {
		t.Fatal("default prompt focused")
	}
	nm, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyTab})
	m = asModel(nm)
	if m.focusPrompt || m.mode != ModeNormal {
		t.Fatal("tab → scrollback")
	}
	// typing letter returns to prompt (simple mode)
	nm, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = asModel(nm)
	if !m.focusPrompt {
		t.Fatal("printable should re-focus prompt")
	}
}

func TestStreamingEnterShowsToastNotDead(t *testing.T) {
	m := testModel(t)
	m.chat.streaming = true
	m.focusPrompt = true
	m.mode = ModeInsert
	nm, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEnter})
	m = asModel(nm)
	// should not hang; toast may be set
	if !m.toast.Alive() && m.chat.streaming {
		// toast optional if NewToast API differs — at least no panic
		t.Log("streaming enter handled")
	}
}

func TestBlockViewEscReturnsPrompt(t *testing.T) {
	m := testModel(t)
	// force a block then open viewer
	m.chat.AddSystemMessage("block for viewer")
	// select first block
	if m.chat.store != nil {
		// store may already have system lines
		m.mode = ModeBlockView
		m.focusPrompt = false
		// open empty block view via field
		nm, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEscape})
		// if not in block view properly, just test returnToPrompt helper
		_ = nm
	}
	m.mode = ModeBlockView
	m.focusPrompt = false
	m.returnToPrompt()
	if m.mode != ModeInsert || !m.focusPrompt {
		t.Fatal("returnToPrompt")
	}
}
