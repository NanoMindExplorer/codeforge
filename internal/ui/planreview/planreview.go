// Package planreview is the Grok-style design plan approval surface.
// Keys: a=approve, s=request changes, q=quit plan; j/k scroll.
package planreview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/codeforge/tui/internal/theme"
	"github.com/muesli/reflow/wordwrap"
)

// Action is the result of the approval UI.
type Action string

const (
	ActionNone     Action = ""
	ActionApprove  Action = "approve"
	ActionChanges  Action = "changes" // request changes
	ActionQuit     Action = "quit"
)

// Focus within the approval surface.
type Focus int

const (
	FocusPreview Focus = iota
	FocusPrompt
)

// Model is the full-screen plan approval overlay.
type Model struct {
	Active  bool
	Content string
	Summary string
	Offset  int // scroll line offset
	Width   int
	Height  int
	Focus   Focus
	// When FocusPrompt: freeform revision notes
	Feedback string
	// Result
	Done   bool
	Action Action
	Notes  string // feedback sent with changes or approve-with-comments
}

func New() Model { return Model{} }

// Open starts the approval UI with plan markdown content.
func (m *Model) Open(content, summary string) {
	m.Active = true
	m.Done = false
	m.Action = ActionNone
	m.Content = content
	if strings.TrimSpace(m.Content) == "" {
		m.Content = "# No plan written yet\n\nThe agent exited design mode without a plan file.\nYou can still **approve** to start implementing, **request changes**, or **quit**."
	}
	m.Summary = summary
	m.Offset = 0
	m.Focus = FocusPreview
	m.Feedback = ""
	m.Notes = ""
}

func (m *Model) Close() {
	m.Active = false
	m.Done = false
}

func (m *Model) lines() []string {
	w := m.Width - 8
	if w < 20 {
		w = 20
	}
	wrapped := wordwrap.String(m.Content, w)
	return strings.Split(wrapped, "\n")
}

func (m *Model) viewportH() int {
	// header + footer bars
	h := m.Height - 6
	if h < 5 {
		h = 5
	}
	return h
}

func (m *Model) Scroll(d int) {
	lines := m.lines()
	vh := m.viewportH()
	m.Offset += d
	maxOff := len(lines) - vh
	if maxOff < 0 {
		maxOff = 0
	}
	if m.Offset < 0 {
		m.Offset = 0
	}
	if m.Offset > maxOff {
		m.Offset = maxOff
	}
}

func (m *Model) Page(d int) {
	m.Scroll(d * m.viewportH())
}

func (m *Model) ToggleFocus() {
	if m.Focus == FocusPreview {
		m.Focus = FocusPrompt
	} else {
		m.Focus = FocusPreview
	}
}

func (m *Model) TypeFeedback(s string) {
	m.Feedback += s
}

func (m *Model) BackspaceFeedback() {
	r := []rune(m.Feedback)
	if len(r) == 0 {
		return
	}
	m.Feedback = string(r[:len(r)-1])
}

func (m *Model) Approve() {
	m.Done = true
	m.Active = false
	m.Action = ActionApprove
	m.Notes = strings.TrimSpace(m.Feedback)
}

func (m *Model) RequestChanges() {
	// If already typing feedback and non-empty, submit; else focus prompt
	if m.Focus == FocusPrompt && strings.TrimSpace(m.Feedback) != "" {
		m.Done = true
		m.Active = false
		m.Action = ActionChanges
		m.Notes = strings.TrimSpace(m.Feedback)
		return
	}
	m.Focus = FocusPrompt
}

func (m *Model) SubmitChanges() {
	m.Done = true
	m.Active = false
	m.Action = ActionChanges
	m.Notes = strings.TrimSpace(m.Feedback)
	if m.Notes == "" {
		m.Notes = "Please revise the plan."
	}
}

func (m *Model) Quit() {
	m.Done = true
	m.Active = false
	m.Action = ActionQuit
}

// View renders the full-screen plan review.
func (m Model) View() string {
	if !m.Active {
		return ""
	}
	t := theme.Current()
	w := m.Width
	if w < 40 {
		w = 40
	}
	h := m.Height
	if h < 12 {
		h = 12
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(t.AccentPlan).Render("◈ Design plan review")
	sum := m.Summary
	if sum == "" {
		sum = "Review the plan, then approve or request changes"
	}
	sub := lipgloss.NewStyle().Foreground(t.TextMuted).Render(sum)

	lines := m.lines()
	vh := m.viewportH()
	var body []string
	end := m.Offset + vh
	if end > len(lines) {
		end = len(lines)
	}
	for i := m.Offset; i < end; i++ {
		body = append(body, lines[i])
	}
	if len(body) == 0 {
		body = append(body, lipgloss.NewStyle().Foreground(t.TextMuted).Render("(empty)"))
	}
	scrollInfo := ""
	if len(lines) > vh {
		scrollInfo = fmt.Sprintf("  %d–%d / %d", m.Offset+1, end, len(lines))
	}
	preview := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderActive).
		Width(w-4).
		Height(vh).
		Padding(0, 1).
		Render(strings.Join(body, "\n"))

	// Action bar
	bar := lipgloss.NewStyle().Foreground(t.TextSecondary).Render(
		"a approve  ·  s request changes  ·  q quit plan  ·  j/k scroll  ·  Tab feedback")
	if m.Focus == FocusPrompt {
		bar = lipgloss.NewStyle().Foreground(t.AccentUser).Render(
			"Feedback: type notes · Enter send  ·  Esc back to plan  ·  a approve w/ notes")
	}
	feedback := ""
	if m.Focus == FocusPrompt || m.Feedback != "" {
		fbStyle := lipgloss.NewStyle().Foreground(t.TextPrimary)
		if m.Focus == FocusPrompt {
			fbStyle = fbStyle.Background(t.BgElevated)
		}
		feedback = fbStyle.Width(w-4).Padding(0, 1).Render("› " + m.Feedback + "▌")
	}

	parts := []string{
		title + scrollInfo,
		sub,
		preview,
		bar,
	}
	if feedback != "" {
		parts = append(parts, feedback)
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
