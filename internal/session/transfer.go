package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Export writes a session (by ID) to destPath as a single JSON bundle.
func Export(id, destPath string) error {
	s, err := Load(id)
	if err != nil {
		return err
	}
	// Ensure full messages loaded
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0644)
}

// Import reads a session JSON file into the sessions dir (v2 layout).
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
	// avoid clobber
	if existing, err := Load(s.ID); err == nil && existing != nil {
		s.ID = s.ID + "-copy"
	}
	s.Format = "v2"
	s.dir = ""
	if s.Workdir == "" {
		s.Workdir, _ = os.Getwd()
	}
	if err := s.Save(); err != nil {
		return nil, err
	}
	return &s, nil
}

// ExportAll copies all sessions into destDir as flat JSON files.
func ExportAll(destDir string) (int, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return 0, err
	}
	list, err := List(0)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, light := range list {
		s, err := Load(light.ID)
		if err != nil {
			continue
		}
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
