// Package checkpoint stores pre-write file snapshots for /undo.
package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Entry is one saved file version.
type Entry struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Path      string    `json:"path"` // absolute path of the file
	RelPath   string    `json:"rel_path"`
	SavedAt   time.Time `json:"saved_at"`
	// content stored as file on disk
}

// Dir returns ~/.codeforge/checkpoints/<sessionID>
// Override root with CODEFORGE_CHECKPOINTS_DIR (tests / portable installs).
func Dir(sessionID string) (string, error) {
	var root string
	if d := os.Getenv("CODEFORGE_CHECKPOINTS_DIR"); d != "" {
		root = d
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, ".codeforge", "checkpoints")
	}
	dir := filepath.Join(root, sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// Save stores the current content of path before it is overwritten.
// oldContent is the bytes that were on disk (may be empty for new files).
func Save(sessionID, absPath, relPath, oldContent string) (string, error) {
	dir, err := Dir(sessionID)
	if err != nil {
		return "", err
	}
	id := time.Now().Format("20060102-150405.000")
	safe := strings.ReplaceAll(relPath, string(filepath.Separator), "__")
	metaName := fmt.Sprintf("%s_%s.meta", id, safe)
	dataName := fmt.Sprintf("%s_%s.data", id, safe)

	meta := fmt.Sprintf("%s\n%s\n%s\n", absPath, relPath, time.Now().Format(time.RFC3339Nano))
	if err := os.WriteFile(filepath.Join(dir, metaName), []byte(meta), 0644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, dataName), []byte(oldContent), 0644); err != nil {
		return "", err
	}
	return id, nil
}

// UndoLast restores the most recent checkpoint for sessionID and returns
// the relative path restored.
func UndoLast(sessionID string) (relPath string, err error) {
	dir, err := Dir(sessionID)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var metas []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".meta") {
			metas = append(metas, e.Name())
		}
	}
	if len(metas) == 0 {
		return "", fmt.Errorf("no checkpoints to undo")
	}
	sort.Strings(metas)
	last := metas[len(metas)-1]
	return restoreAndRemove(filepath.Join(dir, last))
}

// UndoAfter restores all checkpoints saved strictly after t (newest first)
// and removes them. Used by /rewind.
func UndoAfter(sessionID string, t time.Time) ([]string, error) {
	dir, err := Dir(sessionID)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	type item struct {
		meta string
		at   time.Time
	}
	var items []item
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".meta") {
			continue
		}
		metaPath := filepath.Join(dir, e.Name())
		at, ok := metaTime(metaPath)
		if !ok {
			// fallback: parse id prefix from filename
			at = time.Time{}
		}
		if at.After(t) || (at.IsZero() && strings.Compare(e.Name(), t.Format("20060102-150405")) > 0) {
			items = append(items, item{meta: metaPath, at: at})
		}
	}
	// newest first so we undo in reverse write order
	sort.Slice(items, func(i, j int) bool {
		return items[i].at.After(items[j].at)
	})
	var restored []string
	for _, it := range items {
		rel, err := restoreAndRemove(it.meta)
		if err != nil {
			continue
		}
		restored = append(restored, rel)
	}
	return restored, nil
}

func metaTime(metaPath string) (time.Time, bool) {
	b, err := os.ReadFile(metaPath)
	if err != nil {
		return time.Time{}, false
	}
	lines := strings.SplitN(string(b), "\n", 4)
	if len(lines) >= 3 {
		tsRaw := strings.TrimSpace(lines[2])
		if ts, err := time.Parse(time.RFC3339Nano, tsRaw); err == nil {
			return ts, true
		}
		if ts, err := time.Parse(time.RFC3339, tsRaw); err == nil {
			return ts, true
		}
	}
	return time.Time{}, false
}

func restoreAndRemove(metaPath string) (relPath string, err error) {
	dataPath := strings.TrimSuffix(metaPath, ".meta") + ".data"
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return "", err
	}
	lines := strings.SplitN(string(metaBytes), "\n", 3)
	if len(lines) < 2 {
		return "", fmt.Errorf("corrupt checkpoint meta")
	}
	absPath := lines[0]
	relPath = lines[1]

	data, err := os.ReadFile(dataPath)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		_ = os.Remove(absPath)
	} else {
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(absPath, data, 0644); err != nil {
			return "", err
		}
	}
	_ = os.Remove(metaPath)
	_ = os.Remove(dataPath)
	return relPath, nil
}
