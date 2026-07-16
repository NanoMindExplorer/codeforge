package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MigrateResult reports how many legacy sessions were converted.
type MigrateResult struct {
	Migrated int
	Skipped  int
	Errors   []string
}

// MigrateLegacy converts flat ~/.codeforge/sessions/*.json files into
// Phase 4 v2 directories under sessions/<encoded-cwd>/<id>/.
// Original files are renamed to *.json.legacy after successful migrate.
func MigrateLegacy() (MigrateResult, error) {
	var res MigrateResult
	root, err := Dir()
	if err != nil {
		return res, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return res, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".legacy") {
			continue
		}
		if name == "summary.json" {
			continue
		}
		path := filepath.Join(root, name)
		s, err := loadLegacyFile(path)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		if s.ID == "" {
			res.Skipped++
			continue
		}
		// already have v2?
		if group, err := CwdGroupDir(s.Workdir); err == nil {
			v2 := filepath.Join(group, s.ID)
			if isSessionDir(v2) {
				res.Skipped++
				_ = os.Rename(path, path+".legacy")
				continue
			}
		}
		s.Format = "v2"
		s.dir = ""
		if s.Workdir == "" {
			s.Workdir, _ = os.Getwd()
		}
		if err := s.Save(); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("%s save: %v", name, err))
			continue
		}
		// keep a copy of original as .legacy
		if err := os.Rename(path, path+".legacy"); err != nil {
			// non-fatal
			res.Errors = append(res.Errors, fmt.Sprintf("%s rename: %v", name, err))
		}
		res.Migrated++
	}
	return res, nil
}

// IsLegacyFlat reports whether path is a pre-v2 session JSON file.
func IsLegacyFlat(path string) bool {
	if !strings.HasSuffix(path, ".json") {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var probe struct {
		ID       string `json:"id"`
		Messages []any  `json:"messages"`
		Format   string `json:"format"`
	}
	if json.Unmarshal(data, &probe) != nil {
		return false
	}
	return probe.ID != "" && probe.Format != "v2" && len(probe.Messages) >= 0
}
