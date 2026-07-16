// Package session persists and resumes CodeForge conversations.
// Phase 4: Grok-style layout under ~/.codeforge/sessions/<encoded-cwd>/<id>/.
package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/codeforge/tui/internal/provider"
)

// Session is a persisted conversation (in-memory + on disk).
type Session struct {
	ID        string             `json:"id"`
	Slug      string             `json:"slug"`
	Title     string             `json:"title,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Provider  string             `json:"provider"`
	Model     string             `json:"model"`
	Workdir   string             `json:"workdir"`
	Messages  []provider.Message `json:"messages"`
	TotalCost float64            `json:"total_cost"`
	Tokens    int                `json:"tokens"`
	Preview   string             `json:"preview"`
	ParentID  string             `json:"parent_id,omitempty"`
	// Format is "v2" for directory layout; empty/legacy for flat JSON.
	Format string `json:"format,omitempty"`
	// dir is the absolute session directory (v2); empty for legacy.
	dir string
}

// Dir returns the sessions root directory.
// Override with CODEFORGE_SESSIONS_DIR for shared/SSH-synced storage.
func Dir() (string, error) {
	if d := os.Getenv("CODEFORGE_SESSIONS_DIR"); d != "" {
		if err := os.MkdirAll(d, 0755); err != nil {
			return "", err
		}
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".codeforge", "sessions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// New creates a new in-memory session (v2 layout on first Save).
func New(providerName, model, workdir string) *Session {
	now := time.Now()
	return &Session{
		ID:        NewID(),
		Slug:      "new",
		CreatedAt: now,
		UpdatedAt: now,
		Provider:  providerName,
		Model:     model,
		Workdir:   workdir,
		Format:    "v2",
	}
}

// DirPath returns the on-disk session directory (creates for v2).
func (s *Session) DirPath() (string, error) {
	if s.dir != "" {
		return s.dir, nil
	}
	if s.Format == "" || s.Format == "v2" {
		d, err := SessionDir(s.Workdir, s.ID)
		if err != nil {
			return "", err
		}
		s.dir = d
		s.Format = "v2"
		return d, nil
	}
	// legacy: no dir
	return "", fmt.Errorf("legacy session has no directory")
}

// Path returns a path for export/compat (summary.json for v2, flat file for legacy).
func (s *Session) Path() (string, error) {
	if s.Format == "legacy" {
		dir, err := Dir()
		if err != nil {
			return "", err
		}
		slug := s.Slug
		if slug == "" || slug == "new" {
			slug = "session"
		}
		return filepath.Join(dir, fmt.Sprintf("%s-%s.json", s.ID, slug)), nil
	}
	d, err := s.DirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "summary.json"), nil
}

// Save writes the session (v2: summary + chat_history.jsonl + updates line).
func (s *Session) Save() error {
	s.UpdatedAt = time.Now()
	if s.Preview == "" && len(s.Messages) > 0 {
		for _, m := range s.Messages {
			if m.Role == provider.RoleUser {
				s.Preview = truncate(m.Content, 80)
				if s.Slug == "" || s.Slug == "new" {
					s.Slug = slugify(m.Content)
				}
				if s.Title == "" {
					s.Title = truncate(m.Content, 60)
				}
				break
			}
		}
	}
	if s.Format == "legacy" {
		return s.saveLegacy()
	}
	// Default v2
	s.Format = "v2"
	dir, err := s.DirPath()
	if err != nil {
		return err
	}
	sum := Summary{
		ID:           s.ID,
		Slug:         s.Slug,
		Title:        s.Title,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
		Provider:     s.Provider,
		Model:        s.Model,
		Workdir:      s.Workdir,
		Preview:      s.Preview,
		TotalCost:    s.TotalCost,
		Tokens:       s.Tokens,
		MessageCount: len(s.Messages),
		ParentID:     s.ParentID,
		Format:       "v2",
	}
	if err := writeJSON(filepath.Join(dir, "summary.json"), sum); err != nil {
		return err
	}
	if err := s.writeChatHistory(dir); err != nil {
		return err
	}
	// Append a save event to updates.jsonl
	_ = s.appendUpdate(map[string]any{
		"type":     "save",
		"ts":       s.UpdatedAt.Format(time.RFC3339),
		"messages": len(s.Messages),
		"tokens":   s.Tokens,
	})
	return nil
}

func (s *Session) saveLegacy() error {
	path, err := s.Path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Session) writeChatHistory(dir string) error {
	path := filepath.Join(dir, "chat_history.jsonl")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, m := range s.Messages {
		if err := enc.Encode(m); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) appendUpdate(ev map[string]any) error {
	dir, err := s.DirPath()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "updates.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(ev)
}

// AppendEvent writes a structured event to updates.jsonl (tools, turns, etc.).
func (s *Session) AppendEvent(kind string, fields map[string]any) error {
	if s == nil || s.Format == "legacy" {
		return nil
	}
	ev := map[string]any{"type": kind, "ts": time.Now().Format(time.RFC3339)}
	for k, v := range fields {
		ev[k] = v
	}
	return s.appendUpdate(ev)
}

// Load reads a session by ID prefix, full id, or path.
func Load(idOrPath string) (*Session, error) {
	// Direct file path?
	if strings.HasSuffix(idOrPath, ".json") || strings.HasSuffix(idOrPath, "summary.json") {
		return loadAny(idOrPath)
	}
	if strings.Contains(idOrPath, string(filepath.Separator)) {
		// directory or path fragment
		if isSessionDir(idOrPath) {
			return loadV2Dir(idOrPath)
		}
		if st, err := os.Stat(idOrPath); err == nil && st.IsDir() {
			if isSessionDir(idOrPath) {
				return loadV2Dir(idOrPath)
			}
		}
	}

	// Deep search by walking the tree
	root, err := Dir()
	if err != nil {
		return nil, err
	}
	var found *Session
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || found != nil {
			return err
		}
		if info == nil {
			return nil
		}
		// v2: directory named with id
		if info.IsDir() && (info.Name() == idOrPath || strings.HasPrefix(info.Name(), idOrPath)) {
			if isSessionDir(path) {
				s, e := loadV2Dir(path)
				if e == nil {
					found = s
					return filepath.SkipAll
				}
			}
		}
		// legacy flat json
		if !info.IsDir() && strings.HasPrefix(info.Name(), idOrPath) && strings.HasSuffix(info.Name(), ".json") && info.Name() != "summary.json" {
			s, e := loadLegacyFile(path)
			if e == nil {
				found = s
				return filepath.SkipAll
			}
		}
		// summary.json whose parent dir matches
		if info.Name() == "summary.json" {
			parent := filepath.Base(filepath.Dir(path))
			if parent == idOrPath || strings.HasPrefix(parent, idOrPath) {
				s, e := loadV2Dir(filepath.Dir(path))
				if e == nil {
					found = s
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	if found != nil {
		return found, nil
	}
	return nil, fmt.Errorf("session %q not found", idOrPath)
}

func loadAny(path string) (*Session, error) {
	if strings.HasSuffix(path, "summary.json") {
		return loadV2Dir(filepath.Dir(path))
	}
	// maybe session dir
	if isSessionDir(path) {
		return loadV2Dir(path)
	}
	return loadLegacyFile(path)
}

func loadV2Dir(dir string) (*Session, error) {
	var sum Summary
	if err := readJSON(filepath.Join(dir, "summary.json"), &sum); err != nil {
		return nil, err
	}
	msgs, err := readChatHistory(filepath.Join(dir, "chat_history.jsonl"))
	if err != nil {
		// allow empty history
		msgs = nil
	}
	s := &Session{
		ID:        sum.ID,
		Slug:      sum.Slug,
		Title:     sum.Title,
		CreatedAt: sum.CreatedAt,
		UpdatedAt: sum.UpdatedAt,
		Provider:  sum.Provider,
		Model:     sum.Model,
		Workdir:   sum.Workdir,
		Messages:  msgs,
		TotalCost: sum.TotalCost,
		Tokens:    sum.Tokens,
		Preview:   sum.Preview,
		ParentID:  sum.ParentID,
		Format:    "v2",
		dir:       dir,
	}
	if s.ID == "" {
		s.ID = filepath.Base(dir)
	}
	return s, nil
}

func readChatHistory(path string) ([]provider.Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var msgs []provider.Message
	sc := bufio.NewScanner(f)
	// allow long lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var m provider.Message
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		msgs = append(msgs, m)
	}
	return msgs, sc.Err()
}

func loadLegacyFile(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	s.Format = "legacy"
	return &s, nil
}

// List returns sessions newest-first (max n, 0 = all).
// Includes v2 directories and legacy flat JSON files.
func List(max int) ([]Session, error) {
	return ListFilter("", max)
}

// ListForWorkdir returns sessions for a working directory (encoded group).
func ListForWorkdir(workdir string, max int) ([]Session, error) {
	return ListFilter(workdir, max)
}

// ListFilter lists sessions; if workdir != "", prefers that cwd group first then others.
func ListFilter(workdir string, max int) ([]Session, error) {
	root, err := Dir()
	if err != nil {
		return nil, err
	}
	var out []Session
	seen := map[string]bool{}

	// Prefer cwd group
	if workdir != "" {
		group, err := CwdGroupDir(workdir)
		if err == nil {
			out = append(out, listGroup(group, seen)...)
		}
	}

	// Walk entire tree for remaining
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		path := filepath.Join(root, e.Name())
		if e.IsDir() {
			// cwd group
			if workdir != "" && EncodeCWD(workdir) == e.Name() {
				continue // already listed
			}
			// could be cwd group OR (unlikely) bare session
			if isSessionDir(path) {
				if s, err := loadV2Dir(path); err == nil && !seen[s.ID] {
					// light list: drop full messages for index
					s.Messages = nil
					out = append(out, *s)
					seen[s.ID] = true
				}
				continue
			}
			for _, s := range listGroup(path, seen) {
				out = append(out, s)
			}
			continue
		}
		// legacy flat json
		if strings.HasSuffix(e.Name(), ".json") {
			s, err := loadLegacyFile(path)
			if err != nil || seen[s.ID] {
				continue
			}
			s.Messages = nil
			out = append(out, *s)
			seen[s.ID] = true
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	if max > 0 && len(out) > max {
		out = out[:max]
	}
	return out, nil
}

func listGroup(group string, seen map[string]bool) []Session {
	entries, err := os.ReadDir(group)
	if err != nil {
		return nil
	}
	var out []Session
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(group, e.Name())
		if !isSessionDir(path) {
			continue
		}
		s, err := loadV2Dir(path)
		if err != nil || seen[s.ID] {
			continue
		}
		// light: keep preview only
		msgCount := len(s.Messages)
		s.Messages = nil
		if s.Preview == "" && msgCount > 0 {
			s.Preview = fmt.Sprintf("%d messages", msgCount)
		}
		out = append(out, *s)
		seen[s.ID] = true
	}
	return out
}

// Fork creates a new session branching from this one (copy messages).
func (s *Session) Fork() (*Session, error) {
	if s == nil {
		return nil, fmt.Errorf("nil session")
	}
	// Ensure latest state saved
	_ = s.Save()
	child := New(s.Provider, s.Model, s.Workdir)
	child.ParentID = s.ID
	child.Messages = append([]provider.Message(nil), s.Messages...)
	child.TotalCost = 0 // fork starts cost counter fresh
	child.Tokens = s.Tokens
	child.Preview = s.Preview
	child.Slug = s.Slug + "-fork"
	if s.Title != "" {
		child.Title = s.Title + " (fork)"
	}
	if err := child.Save(); err != nil {
		return nil, err
	}
	_ = child.AppendEvent("fork", map[string]any{"parent": s.ID})
	// Copy rewind points
	if pts, err := s.LoadRewindPoints(); err == nil && len(pts) > 0 {
		_ = child.SaveRewindPoints(pts)
	}
	return child, nil
}

// TruncateMessages keeps the first n messages (for rewind).
func (s *Session) TruncateMessages(n int) {
	if n < 0 {
		n = 0
	}
	if n > len(s.Messages) {
		return
	}
	s.Messages = s.Messages[:n]
}

// InfoText returns a human-readable /session-info block.
func (s *Session) InfoText(maxContext int) string {
	if s == nil {
		return "No active session"
	}
	title := s.Title
	if title == "" {
		title = s.Slug
	}
	pct := ""
	if maxContext > 0 && s.Tokens > 0 {
		p := float64(s.Tokens) * 100 / float64(maxContext)
		pct = fmt.Sprintf(" (%.0f%% of %d)", p, maxContext)
	}
	parent := ""
	if s.ParentID != "" {
		parent = "\n  Parent   : " + s.ParentID
	}
	return fmt.Sprintf(`Session Info
  Title    : %s
  ID       : %s
  Workdir  : %s
  Provider : %s
  Model    : %s
  Messages : %d
  Tokens   : %d%s
  Cost     : $%.4f
  Updated  : %s
  Format   : %s%s`,
		title, s.ID, s.Workdir, s.Provider, s.Model,
		len(s.Messages), s.Tokens, pct, s.TotalCost,
		s.UpdatedAt.Format(time.RFC3339), s.Format, parent,
	)
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
		if b.Len() >= 32 {
			break
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "session"
	}
	return out
}
