package session

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Summary is the index entry (summary.json) for a session directory.
type Summary struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	Workdir      string    `json:"workdir"`
	Preview      string    `json:"preview"`
	TotalCost    float64   `json:"total_cost"`
	Tokens       int       `json:"tokens"`
	MessageCount int       `json:"message_count"`
	ParentID     string    `json:"parent_id,omitempty"` // fork source
	Compacted    bool      `json:"compacted,omitempty"`
	Format       string    `json:"format"` // "v2" directory layout
}

// EncodeCWD URL-encodes a working directory for the sessions tree.
// Long paths become slug+hash with a .cwd sidecar.
func EncodeCWD(workdir string) string {
	abs := workdir
	if a, err := filepath.Abs(workdir); err == nil {
		abs = a
	}
	abs = filepath.Clean(abs)
	enc := url.PathEscape(abs)
	// PathEscape leaves / as %2F on some platforms — ensure separators encoded
	enc = strings.ReplaceAll(enc, "/", "%2F")
	if len(enc) <= 200 {
		return enc
	}
	// Truncate + hash
	sum := sha1.Sum([]byte(abs))
	h := hex.EncodeToString(sum[:8])
	base := filepath.Base(abs)
	base = slugify(base)
	if base == "" {
		base = "cwd"
	}
	return base + "-" + h
}

// CwdGroupDir returns ~/.codeforge/sessions/<encoded-cwd>/
func CwdGroupDir(workdir string) (string, error) {
	root, err := Dir()
	if err != nil {
		return "", err
	}
	enc := EncodeCWD(workdir)
	dir := filepath.Join(root, enc)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	// Write .cwd for long-path groups (always useful)
	cwdFile := filepath.Join(dir, ".cwd")
	if _, err := os.Stat(cwdFile); os.IsNotExist(err) {
		abs := workdir
		if a, e := filepath.Abs(workdir); e == nil {
			abs = a
		}
		_ = os.WriteFile(cwdFile, []byte(abs+"\n"), 0644)
	}
	return dir, nil
}

// SessionDir returns the directory for a session under its workdir group.
func SessionDir(workdir, id string) (string, error) {
	group, err := CwdGroupDir(workdir)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(group, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// NewID returns a unique session id (time + random hex).
func NewID() string {
	now := time.Now()
	// 4 random bytes from time+pid entropy
	r := fmt.Sprintf("%04x", now.UnixNano()&0xffff)
	return now.Format("20060102-150405") + "-" + r
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// isSessionDir reports whether path looks like a v2 session directory.
func isSessionDir(path string) bool {
	st, err := os.Stat(filepath.Join(path, "summary.json"))
	return err == nil && !st.IsDir()
}
