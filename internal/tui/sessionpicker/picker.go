// Package sessionpicker implements full-screen /resume session browser.
package sessionpicker

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/codeforge/tui/internal/session"
	"github.com/codeforge/tui/internal/theme"
	"github.com/sahilm/fuzzy"
)

// Model is the session resume picker.
type Model struct {
	Active   bool
	Items    []session.Session
	Filtered []session.Session
	Cursor   int
	Query    string
	Width    int
	Height   int
	Workdir  string
	Done     bool
	Selected *session.Session
	// Mode: "resume" or "rewind"
	Mode string
}

// RewindItem is one row in the rewind picker.
type RewindItem struct {
	Point session.RewindPoint
}

// RewindModel is the /rewind point picker.
type RewindModel struct {
	Active   bool
	Items    []session.RewindPoint
	Cursor   int
	Width    int
	Done     bool
	Selected *session.RewindPoint
}

func New() Model {
	return Model{Mode: "resume"}
}

func NewRewind() RewindModel {
	return RewindModel{}
}

// Open loads sessions for workdir (and others) into the resume picker.
func (m *Model) Open(workdir string) {
	m.Active = true
	m.Done = false
	m.Selected = nil
	m.Query = ""
	m.Cursor = 0
	m.Workdir = workdir
	m.Mode = "resume"
	list, err := session.ListForWorkdir(workdir, 40)
	if err != nil {
		list = nil
	}
	// also pull global recent if few local
	if len(list) < 10 {
		all, _ := session.List(30)
		seen := map[string]bool{}
		for _, s := range list {
			seen[s.ID] = true
		}
		for _, s := range all {
			if !seen[s.ID] {
				list = append(list, s)
				seen[s.ID] = true
			}
		}
	}
	m.Items = list
	m.Filtered = list
}

func (m *Model) Close() {
	m.Active = false
	m.Done = false
	m.Selected = nil
}

func (m *Model) SetQuery(q string) {
	m.Query = q
	m.refilter()
}

func (m *Model) Type(s string) {
	m.Query += s
	m.refilter()
}

func (m *Model) Backspace() {
	r := []rune(m.Query)
	if len(r) == 0 {
		return
	}
	m.Query = string(r[:len(r)-1])
	m.refilter()
}

func (m *Model) refilter() {
	q := strings.TrimSpace(m.Query)
	if q == "" {
		m.Filtered = m.Items
		m.Cursor = 0
		return
	}
	labels := make([]string, len(m.Items))
	for i, s := range m.Items {
		labels[i] = s.ID + " " + s.Slug + " " + s.Title + " " + s.Preview + " " + s.Workdir
	}
	matches := fuzzy.Find(q, labels)
	m.Filtered = make([]session.Session, 0, len(matches))
	for _, match := range matches {
		m.Filtered = append(m.Filtered, m.Items[match.Index])
	}
	m.Cursor = 0
}

func (m *Model) Move(d int) {
	if len(m.Filtered) == 0 {
		return
	}
	m.Cursor += d
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	if m.Cursor >= len(m.Filtered) {
		m.Cursor = len(m.Filtered) - 1
	}
}

func (m *Model) Confirm() {
	m.Done = true
	m.Active = false
	if len(m.Filtered) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Filtered) {
		sel := m.Filtered[m.Cursor]
		// Full load
		full, err := session.Load(sel.ID)
		if err == nil {
			m.Selected = full
		} else {
			m.Selected = &sel
		}
	}
}

func (m *Model) Cancel() {
	m.Done = true
	m.Active = false
	m.Selected = nil
}

func (m Model) View() string {
	if !m.Active {
		return ""
	}
	t := theme.Current()
	w := m.Width
	if w > 80 {
		w = 80
	}
	if w < 40 {
		w = 40
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(t.BorderGlow).Render("◷ Resume session")
	input := lipgloss.NewStyle().Foreground(t.AccentAI).Render("filter: " + m.Query + "▌")
	sub := lipgloss.NewStyle().Foreground(t.TextMuted).Render(
		fmt.Sprintf("%d sessions  ·  ↑↓  Enter resume  Esc cancel", len(m.Filtered)))

	maxShow := 12
	if m.Height > 0 && m.Height/3 < maxShow {
		maxShow = m.Height / 3
	}
	if maxShow < 5 {
		maxShow = 5
	}

	var rows []string
	for i, s := range m.Filtered {
		if i >= maxShow {
			rows = append(rows, lipgloss.NewStyle().Foreground(t.TextMuted).Render(
				fmt.Sprintf("  … +%d more", len(m.Filtered)-maxShow)))
			break
		}
		when := s.UpdatedAt.Format("01-02 15:04")
		if time.Since(s.UpdatedAt) < 24*time.Hour {
			when = s.UpdatedAt.Format("15:04")
		}
		title := s.Title
		if title == "" {
			title = s.Slug
		}
		if title == "" || title == "new" {
			title = s.Preview
		}
		if title == "" {
			title = s.ID
		}
		cwdMark := ""
		if m.Workdir != "" && sameCWD(s.Workdir, m.Workdir) {
			cwdMark = "·"
		}
		line := fmt.Sprintf("%s %-18s  %s  %s", cwdMark, trunc(s.ID, 18), when, trunc(title, 36))
		if i == m.Cursor {
			// preview under selection
			prev := trunc(s.Preview, w-10)
			line = lipgloss.NewStyle().Background(t.BgElevated).Foreground(t.AccentFocus).Bold(true).
				Width(w - 6).Render("› " + line)
			extra := lipgloss.NewStyle().Foreground(t.TextMuted).Width(w - 6).
				Render(fmt.Sprintf("  %s  %s/%s  $%.4f", trunc(s.Workdir, 40), s.Provider, trunc(s.Model, 16), s.TotalCost))
			if prev != "" {
				extra += "\n" + lipgloss.NewStyle().Foreground(t.TextSecondary).Render("  "+prev)
			}
			rows = append(rows, line+"\n"+extra)
		} else {
			rows = append(rows, lipgloss.NewStyle().Foreground(t.TextSecondary).Width(w-6).Render("  "+line))
		}
	}
	if len(rows) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(t.TextMuted).Render("  (no sessions — start chatting to create one)"))
	}

	body := title + "\n" + input + "\n" + sub + "\n" +
		lipgloss.NewStyle().Foreground(t.BorderDim).Render(strings.Repeat("─", w-6)) + "\n" +
		strings.Join(rows, "\n")
	return theme.OverlayStyle(w).Render(body)
}

// --- Rewind picker ---

func (m *RewindModel) Open(pts []session.RewindPoint) {
	m.Active = true
	m.Done = false
	m.Selected = nil
	m.Cursor = 0
	// newest last in file; show newest first
	rev := make([]session.RewindPoint, len(pts))
	for i := range pts {
		rev[len(pts)-1-i] = pts[i]
	}
	m.Items = rev
}

func (m *RewindModel) Move(d int) {
	if len(m.Items) == 0 {
		return
	}
	m.Cursor += d
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	if m.Cursor >= len(m.Items) {
		m.Cursor = len(m.Items) - 1
	}
}

func (m *RewindModel) Confirm() {
	m.Done = true
	m.Active = false
	if len(m.Items) > 0 && m.Cursor >= 0 && m.Cursor < len(m.Items) {
		sel := m.Items[m.Cursor]
		m.Selected = &sel
	}
}

func (m *RewindModel) Cancel() {
	m.Done = true
	m.Active = false
	m.Selected = nil
}

func (m RewindModel) View() string {
	if !m.Active {
		return ""
	}
	t := theme.Current()
	w := m.Width
	if w > 72 {
		w = 72
	}
	if w < 36 {
		w = 36
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(t.Warning).Render("↶ Rewind")
	sub := lipgloss.NewStyle().Foreground(t.TextMuted).Render(
		"Restores files + truncates chat  ·  ↑↓  Enter  Esc")
	var rows []string
	for i, p := range m.Items {
		line := fmt.Sprintf("%s  msg@%d  %s", p.CreatedAt.Format("15:04:05"), p.MessageIndex, trunc(p.Preview, 40))
		if i == m.Cursor {
			line = lipgloss.NewStyle().Background(t.BgElevated).Foreground(t.AccentFocus).Bold(true).
				Width(w - 6).Render("› " + line)
		} else {
			line = lipgloss.NewStyle().Foreground(t.TextSecondary).Width(w - 6).Render("  " + line)
		}
		rows = append(rows, line)
	}
	if len(rows) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(t.TextMuted).Render("  (no rewind points yet)"))
	}
	body := title + "\n" + sub + "\n" +
		lipgloss.NewStyle().Foreground(t.BorderDim).Render(strings.Repeat("─", w-6)) + "\n" +
		strings.Join(rows, "\n")
	return theme.OverlayStyle(w).Render(body)
}

func trunc(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	if n < 2 {
		return s
	}
	return s[:n-1] + "…"
}

func sameCWD(a, b string) bool {
	return strings.TrimRight(a, "/") == strings.TrimRight(b, "/")
}
