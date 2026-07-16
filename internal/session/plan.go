package session

import (
	"os"
	"path/filepath"
	"strings"
)

// PlanPath returns the absolute path to plan.md for this session.
func (s *Session) PlanPath() (string, error) {
	if s == nil {
		return "", os.ErrInvalid
	}
	dir, err := s.DirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "plan.md"), nil
}

// ReadPlan returns plan.md contents (empty string if missing).
func (s *Session) ReadPlan() (string, error) {
	path, err := s.PlanPath()
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}

// WritePlan writes plan.md content.
func (s *Session) WritePlan(content string) error {
	path, err := s.PlanPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// HasPlan reports whether a non-empty plan.md exists.
func (s *Session) HasPlan() bool {
	body, err := s.ReadPlan()
	return err == nil && strings.TrimSpace(body) != ""
}
