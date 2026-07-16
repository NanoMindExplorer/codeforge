package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeforge/tui/internal/provider"
)

func TestMigrateLegacy(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEFORGE_SESSIONS_DIR", filepath.Join(home, "sessions"))
	root := filepath.Join(home, "sessions")
	_ = os.MkdirAll(root, 0755)

	// write legacy flat file manually
	legacy := map[string]any{
		"id":        "20260101-120000",
		"slug":      "hello",
		"provider":  "gemini",
		"model":     "flash",
		"workdir":   "/tmp/proj",
		"messages":  []provider.Message{{Role: provider.RoleUser, Content: "hello migrate"}},
		"preview":   "hello migrate",
		"total_cost": 0.01,
	}
	data, _ := json.MarshalIndent(legacy, "", "  ")
	path := filepath.Join(root, "20260101-120000-hello.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	res, err := MigrateLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if res.Migrated < 1 {
		t.Fatalf("migrated=%d errors=%v", res.Migrated, res.Errors)
	}
	// v2 dir exists
	loaded, err := Load("20260101-120000")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Format != "v2" || len(loaded.Messages) != 1 {
		t.Fatalf("%+v", loaded)
	}
	// original renamed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("legacy file should be renamed")
	}
	if _, err := os.Stat(path + ".legacy"); err != nil {
		t.Fatal("expected .legacy backup")
	}
}
