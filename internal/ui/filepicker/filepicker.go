// Package filepicker provides Grok-style @file fuzzy mention UI.
// Supports gitignore, hidden files (!), and path:line-range attachments.
package filepicker

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/codeforge/tui/internal/theme"
	"github.com/sahilm/fuzzy"
)

// MaxListedFiles caps directory walks so @ never freezes the TUI on huge trees.
const MaxListedFiles = 400

// MaxWalkDepth limits recursion depth under workdir.
const MaxWalkDepth = 8

// Model is a compact file picker popup.
type Model struct {
	Active     bool
	Workdir    string
	Query      string
	Files      []string // cached listing (refreshed on Open only)
	Filtered   []string
	Cursor     int
	Width      int
	Done       bool
	Selected   string // path or path:start-end
	ShowHidden bool   // ! prefix in query
	// Line range parsed from query trailing :N-M
	RangeStart int
	RangeEnd   int

	// listOnce avoids re-walking the disk on every keystroke (freeze fix).
	listed     bool
	listHidden bool
}

func New(workdir string) Model {
	return Model{Workdir: workdir}
}

// Open loads project files (gitignore-aware).
func (m *Model) Open() {
	m.OpenWithQuery("")
}

// OpenWithQuery opens and seeds query (after @).
func (m *Model) OpenWithQuery(q string) {
	m.Active = true
	m.Done = false
	m.Selected = ""
	m.Cursor = 0
	m.RangeStart, m.RangeEnd = 0, 0
	// Force one disk scan for this open session
	m.listed = false
	m.SetQuery(q)
}

func (m *Model) Close() {
	m.Active = false
	m.Done = false
}

func (m *Model) Type(s string) {
	m.Query += s
	m.refilter()
}

func (m *Model) SetQuery(q string) {
	m.Query = q
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
	q := m.Query
	showHidden := false
	m.RangeStart, m.RangeEnd = 0, 0

	// hidden: leading !
	if strings.HasPrefix(q, "!") {
		showHidden = true
		q = q[1:]
	}
	// line range: path:10-50 or path:10
	pathQ := q
	if i := strings.LastIndex(q, ":"); i >= 0 {
		tail := q[i+1:]
		if isLineRange(tail) {
			pathQ = q[:i]
			m.RangeStart, m.RangeEnd = parseLineRange(tail)
		}
	}

	// Disk walk only when first open or when hidden toggle flips — not every key.
	if !m.listed || m.listHidden != showHidden {
		m.Files = listFiles(m.Workdir, showHidden)
		m.listed = true
		m.listHidden = showHidden
	}
	m.ShowHidden = showHidden

	if strings.TrimSpace(pathQ) == "" {
		m.Filtered = m.Files
		if m.Cursor >= len(m.Filtered) {
			m.Cursor = 0
		}
		return
	}
	matches := fuzzy.Find(pathQ, m.Files)
	m.Filtered = make([]string, 0, len(matches))
	for _, match := range matches {
		m.Filtered = append(m.Filtered, m.Files[match.Index])
	}
	m.Cursor = 0
}

func isLineRange(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	return true
}

func parseLineRange(s string) (start, end int) {
	parts := strings.SplitN(s, "-", 2)
	start, _ = strconv.Atoi(parts[0])
	if len(parts) == 2 {
		end, _ = strconv.Atoi(parts[1])
	} else {
		end = start
	}
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	return start, end
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
	if len(m.Filtered) == 0 || m.Cursor >= len(m.Filtered) {
		return
	}
	sel := m.Filtered[m.Cursor]
	if m.RangeStart > 0 {
		if m.RangeEnd > 0 && m.RangeEnd != m.RangeStart {
			sel = sel + ":" + strconv.Itoa(m.RangeStart) + "-" + strconv.Itoa(m.RangeEnd)
		} else {
			sel = sel + ":" + strconv.Itoa(m.RangeStart)
		}
	}
	m.Selected = sel
}

func (m *Model) Cancel() {
	m.Done = true
	m.Active = false
	m.Selected = ""
}

func (m Model) View() string {
	if !m.Active {
		return ""
	}
	t := theme.Current()
	w := m.Width
	if w <= 0 {
		w = 48
	}
	if w > 56 {
		w = 56
	}
	title := "@ file"
	if m.ShowHidden {
		title += " (hidden)"
	}
	if m.RangeStart > 0 {
		title += "  lines"
	}
	var rows []string
	rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(t.AccentUser).Render(title+"  "+m.Query+"▌"))
	maxN := 10
	for i, f := range m.Filtered {
		if i >= maxN {
			rows = append(rows, lipgloss.NewStyle().Foreground(t.TextMuted).Render("  …"))
			break
		}
		icon := theme.FileIcon(f)
		line := icon + " " + f
		if i == m.Cursor {
			line = lipgloss.NewStyle().Background(t.BgElevated).Foreground(t.AccentFocus).Render("› " + line)
		} else {
			line = lipgloss.NewStyle().Foreground(t.TextSecondary).Render("  " + line)
		}
		rows = append(rows, line)
	}
	if len(m.Filtered) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(t.TextMuted).Render("  (no files)"))
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(t.TextMuted).Italic(true).Render(
		"  !hidden  path:10-50  ↑↓ Enter Esc"))
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.AccentUser).
		Background(t.BgOverlay).
		Padding(0, 1).
		Width(w).
		Render(strings.Join(rows, "\n"))
}

// listFiles walks project respecting gitignore (basic) unless showHidden.
// Bounded by MaxListedFiles / MaxWalkDepth so the TUI never freezes.
func listFiles(workdir string, showHidden bool) []string {
	ignore := loadGitignore(workdir)
	var out []string
	_ = filepath.WalkDir(workdir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(workdir, path)
		if err != nil {
			return nil
		}
		if rel == "." {
			return nil
		}
		// depth limit
		depth := strings.Count(filepath.ToSlash(rel), "/")
		if d.IsDir() && depth >= MaxWalkDepth {
			return filepath.SkipDir
		}
		name := d.Name()
		if d.IsDir() {
			if name == ".git" || name == "node_modules" || name == "vendor" ||
				name == "dist" || name == "build" || name == "target" ||
				name == ".cache" || name == "__pycache__" || name == ".next" ||
				name == "Pods" || name == ".idea" || name == ".vscode" {
				return filepath.SkipDir
			}
			if !showHidden && strings.HasPrefix(name, ".") && name != ".github" {
				return filepath.SkipDir
			}
			if ignoredBy(ignore, rel, true) {
				return filepath.SkipDir
			}
			return nil
		}
		if !showHidden && strings.HasPrefix(name, ".") {
			return nil
		}
		if ignoredBy(ignore, rel, false) {
			return nil
		}
		out = append(out, rel)
		if len(out) >= MaxListedFiles {
			return filepath.SkipAll
		}
		return nil
	})
	return out
}

func loadGitignore(workdir string) []string {
	data, err := os.ReadFile(filepath.Join(workdir, ".gitignore"))
	if err != nil {
		return nil
	}
	var pats []string
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pats = append(pats, line)
	}
	return pats
}

func ignoredBy(pats []string, rel string, isDir bool) bool {
	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)
	for _, p := range pats {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "!") {
			p = p[1:]
		}
		p = strings.TrimSuffix(p, "/")
		match := false
		if ok, _ := filepath.Match(p, base); ok {
			match = true
		}
		if ok, _ := filepath.Match(p, rel); ok {
			match = true
		}
		if strings.HasSuffix(p, "/**") || strings.HasSuffix(p, "/*") {
			pref := strings.TrimSuffix(strings.TrimSuffix(p, "*"), "/")
			if strings.HasPrefix(rel, pref) {
				match = true
			}
		}
		if match {
			return true
		}
		_ = isDir
	}
	return false
}

// ReadFileContent loads a file (optionally path:start-end) capped at maxBytes.
func ReadFileContent(workdir, sel string, maxBytes int) (string, error) {
	path := sel
	start, end := 0, 0
	if i := strings.LastIndex(sel, ":"); i >= 0 {
		tail := sel[i+1:]
		if isLineRange(tail) {
			path = sel[:i]
			start, end = parseLineRange(tail)
		}
	}
	abs := path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(workdir, path)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}
	if maxBytes > 0 && len(data) > maxBytes {
		data = data[:maxBytes]
	}
	if start <= 0 {
		return string(data), nil
	}
	lines := strings.Split(string(data), "\n")
	if start > len(lines) {
		start = len(lines)
	}
	if end <= 0 || end > len(lines) {
		end = len(lines)
	}
	if start < 1 {
		start = 1
	}
	return strings.Join(lines[start-1:end], "\n"), nil
}
