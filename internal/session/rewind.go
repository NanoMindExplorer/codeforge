package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codeforge/tui/internal/checkpoint"
)

// RewindPoint marks a conversation position that can be restored.
// MessageIndex is the number of messages to keep (truncate to this length).
type RewindPoint struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	MessageIndex int       `json:"message_index"`
	Preview      string    `json:"preview"`
	TurnID       string    `json:"turn_id,omitempty"`
}

// RecordRewindPoint appends a rewind point at the current message count
// (call AFTER appending the user message so index includes it).
func (s *Session) RecordRewindPoint(preview, turnID string) (*RewindPoint, error) {
	if s == nil {
		return nil, fmt.Errorf("nil session")
	}
	if s.Format == "legacy" {
		// still track in memory path if possible
		s.Format = "v2"
	}
	pt := RewindPoint{
		ID:           time.Now().Format("20060102-150405.000"),
		CreatedAt:    time.Now(),
		MessageIndex: len(s.Messages),
		Preview:      truncate(preview, 80),
		TurnID:       turnID,
	}
	pts, _ := s.LoadRewindPoints()
	pts = append(pts, pt)
	if err := s.SaveRewindPoints(pts); err != nil {
		return nil, err
	}
	_ = s.AppendEvent("rewind_point", map[string]any{
		"id": pt.ID, "msg_index": pt.MessageIndex, "preview": pt.Preview,
	})
	return &pt, nil
}

// LoadRewindPoints reads rewind_points.jsonl.
func (s *Session) LoadRewindPoints() ([]RewindPoint, error) {
	if s == nil {
		return nil, nil
	}
	dir, err := s.DirPath()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "rewind_points.jsonl")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var pts []RewindPoint
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var p RewindPoint
		if json.Unmarshal(sc.Bytes(), &p) == nil && p.ID != "" {
			pts = append(pts, p)
		}
	}
	return pts, sc.Err()
}

// SaveRewindPoints rewrites rewind_points.jsonl.
func (s *Session) SaveRewindPoints(pts []RewindPoint) error {
	dir, err := s.DirPath()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "rewind_points.jsonl")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, p := range pts {
		if err := enc.Encode(p); err != nil {
			return err
		}
	}
	return nil
}

// ApplyRewind truncates conversation to the point and restores files written
// after that point via checkpoints. Returns restored file paths.
func (s *Session) ApplyRewind(pt RewindPoint) (restored []string, err error) {
	if s == nil {
		return nil, fmt.Errorf("nil session")
	}
	// Restore files changed after this point
	restored, err = checkpoint.UndoAfter(s.ID, pt.CreatedAt)
	if err != nil {
		// non-fatal if no checkpoints
		restored = nil
	}
	s.TruncateMessages(pt.MessageIndex)
	// Drop later rewind points
	pts, _ := s.LoadRewindPoints()
	var keep []RewindPoint
	for _, p := range pts {
		if !p.CreatedAt.After(pt.CreatedAt) {
			keep = append(keep, p)
		}
	}
	_ = s.SaveRewindPoints(keep)
	if err := s.Save(); err != nil {
		return restored, err
	}
	_ = s.AppendEvent("rewind", map[string]any{
		"to": pt.ID, "msg_index": pt.MessageIndex, "files": len(restored),
	})
	return restored, nil
}
