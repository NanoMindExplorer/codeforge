package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Export writes a session (by ID) to destPath as JSON.
func Export(id, destPath string) error {
	s, err := Load(id)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0644)
}

// Import reads a session JSON file into the sessions dir (new copy).
func Import(srcPath string) (*Session, error) {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.ID == "" {
		s.ID = "import-" + slugify(filepath.Base(srcPath))
	}
	// avoid clobber: re-id if exists
	if _, err := Load(s.ID); err == nil {
		s.ID = s.ID + "-copy"
	}
	if err := s.Save(); err != nil {
		return nil, err
	}
	return &s, nil
}

// ExportAll copies all sessions into destDir.
func ExportAll(destDir string) (int, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return 0, err
	}
	list, err := List(0)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, s := range list {
		name := fmt.Sprintf("%s-%s.json", s.ID, s.Slug)
		name = strings.ReplaceAll(name, string(filepath.Separator), "_")
		data, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(destDir, name), data, 0644); err != nil {
			continue
		}
		n++
	}
	return n, nil
}
