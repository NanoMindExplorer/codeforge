package session

import (
	"fmt"
	"strings"
	"time"
)

// LastForWorkdir returns the most recently updated session for workdir, or nil.
// Full messages are loaded so callers can apply immediately (Q4.2).
func LastForWorkdir(workdir string) (*Session, error) {
	list, err := ListForWorkdir(workdir, 1)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	// List is light (messages stripped) — reload full
	return Load(list[0].ID)
}

// FormatResumeList builds a human-readable /sessions · /resume preview table (Q4.2).
// Each row: id, relative time, model, session mode, preview snippet.
func FormatResumeList(list []Session, max int) string {
	if max <= 0 {
		max = 15
	}
	if len(list) == 0 {
		return "No saved sessions for this project.\n\nChat to create one, then /resume or /sessions."
	}
	if len(list) > max {
		list = list[:max]
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Sessions (%d) — /resume · /resume last · /resume <id>\n", len(list)))
	for i, s := range list {
		title := s.Title
		if title == "" {
			title = s.Slug
		}
		if title == "" {
			title = "untitled"
		}
		prev := s.Preview
		if prev == "" {
			prev = "(no preview)"
		}
		prev = truncate(prev, 72)
		mode := s.SessionMode
		if mode == "" {
			mode = "—"
		}
		model := s.Model
		if model == "" {
			model = s.Provider
		}
		if model == "" {
			model = "?"
		}
		marker := " "
		if i == 0 {
			marker = "→" // newest = default for /resume last
		}
		fmt.Fprintf(&b, "%s %s  %s\n", marker, s.ID, title)
		fmt.Fprintf(&b, "    %s · %s · mode=%s · %s\n",
			relTime(s.UpdatedAt), model, mode, prev)
	}
	b.WriteString("\n/resume last  → newest for this cwd\n/resume       → interactive picker")
	return b.String()
}

func relTime(t time.Time) string {
	if t.IsZero() {
		return "?"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}
