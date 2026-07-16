// Package workspace supports multi-root / monorepo path resolution and
// smart ignore rules for CodeForge tools.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DefaultIgnoreDirs are never walked for grep/list recursion.
var DefaultIgnoreDirs = []string{
	".git", "node_modules", "vendor", "dist", "build",
	".next", ".nuxt", "target", "bin", "obj",
	"__pycache__", ".venv", "venv", ".tox",
	".idea", ".vscode", ".cache", "coverage",
	".codeforge", ".terraform",
}

// DefaultIgnoreFiles globs (suffix / name match).
var DefaultIgnoreFiles = []string{
	".env", ".env.local", ".env.production",
	"*.pem", "*.key", "*.p12", "*.pfx",
	"id_rsa", "id_ed25519",
	"*.min.js", "*.min.css",
	"go.sum", "package-lock.json", "yarn.lock", "pnpm-lock.yaml",
}

// Root is one project root (absolute path + optional label).
type Root struct {
	Path  string // absolute
	Label string // short name for display
}

// Workspace is the multi-root sandbox.
type Workspace struct {
	mu          sync.RWMutex
	Primary     string // absolute primary workdir
	Roots       []Root
	IgnoreDirs  map[string]bool
	IgnoreGlobs []string
	// AllowSecrets when true disables secret-ish file skips (default false)
	AllowSecrets bool
}

// New creates a workspace with a single primary root.
func New(primary string) *Workspace {
	abs, err := filepath.Abs(primary)
	if err != nil {
		abs = primary
	}
	w := &Workspace{
		Primary:     abs,
		Roots:       []Root{{Path: abs, Label: filepath.Base(abs)}},
		IgnoreDirs:  map[string]bool{},
		IgnoreGlobs: append([]string{}, DefaultIgnoreFiles...),
	}
	for _, d := range DefaultIgnoreDirs {
		w.IgnoreDirs[d] = true
	}
	return w
}

// AddRoot registers an extra monorepo root (absolute or relative to primary).
func (w *Workspace) AddRoot(path, label string) error {
	if !filepath.IsAbs(path) {
		path = filepath.Join(w.Primary, path)
	}
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	if label == "" {
		label = filepath.Base(path)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, r := range w.Roots {
		if r.Path == path {
			return nil
		}
	}
	w.Roots = append(w.Roots, Root{Path: path, Label: label})
	return nil
}

// SetIgnoreDirs replaces the ignore-dir set.
func (w *Workspace) SetIgnoreDirs(dirs []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.IgnoreDirs = map[string]bool{}
	for _, d := range dirs {
		w.IgnoreDirs[d] = true
	}
}

// ResolvePath maps a relative or absolute path into a sandbox-approved absolute path.
// Relative paths are tried against each root (primary first).
func (w *Workspace) ResolvePath(path string) (abs string, root Root, err error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if path == "" {
		return "", Root{}, fmt.Errorf("path required")
	}
	if filepath.IsAbs(path) {
		clean := filepath.Clean(path)
		for _, r := range w.Roots {
			if isInside(r.Path, clean) {
				return clean, r, nil
			}
		}
		return "", Root{}, fmt.Errorf("path %q is outside workspace roots", path)
	}
	// relative: prefer a root where the path already exists; else first valid sandbox path
	var (
		fallbackAbs  string
		fallbackRoot Root
		haveFallback bool
	)
	for _, r := range w.Roots {
		candidate := filepath.Clean(filepath.Join(r.Path, path))
		if !isInside(r.Path, candidate) {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, r, nil
		}
		if !haveFallback {
			fallbackAbs, fallbackRoot, haveFallback = candidate, r, true
		}
	}
	if haveFallback {
		// Allow writes to new files under the first matching root (usually primary)
		return fallbackAbs, fallbackRoot, nil
	}
	return "", Root{}, fmt.Errorf("path %q outside workspace roots", path)
}

// RelPath returns a display path relative to primary or owning root.
func (w *Workspace) RelPath(abs string) string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if rel, err := filepath.Rel(w.Primary, abs); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	for _, r := range w.Roots {
		if rel, err := filepath.Rel(r.Path, abs); err == nil && !strings.HasPrefix(rel, "..") {
			if r.Path == w.Primary {
				return rel
			}
			return filepath.Join(r.Label, rel)
		}
	}
	return abs
}

// ShouldSkipDir reports if a directory name should be skipped during walks.
func (w *Workspace) ShouldSkipDir(name string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if strings.HasPrefix(name, ".") && name != "." {
		// allow .github
		if name == ".github" {
			return false
		}
		return true
	}
	return w.IgnoreDirs[name]
}

// ShouldSkipFile reports if a file should be ignored for search/list bulk ops.
func (w *Workspace) ShouldSkipFile(name string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if !w.AllowSecrets {
		lower := strings.ToLower(name)
		if lower == ".env" || strings.HasPrefix(lower, ".env.") {
			return true
		}
		if strings.HasSuffix(lower, ".pem") || strings.HasSuffix(lower, ".key") {
			return true
		}
	}
	for _, g := range w.IgnoreGlobs {
		if ok, _ := filepath.Match(g, name); ok {
			return true
		}
	}
	return false
}

// ListRoots returns a copy of roots.
func (w *Workspace) ListRoots() []Root {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]Root, len(w.Roots))
	copy(out, w.Roots)
	return out
}

func isInside(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// Global is the process-wide workspace (set at startup).
var (
	globalMu sync.RWMutex
	global   *Workspace
)

// SetGlobal installs the process workspace.
func SetGlobal(w *Workspace) {
	globalMu.Lock()
	global = w
	globalMu.Unlock()
}

// Global returns the process workspace (may be nil before init).
func Get() *Workspace {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return global
}

// Resolve uses global workspace or falls back to single-root workdir.
func Resolve(workdir, path string) (string, error) {
	if w := Get(); w != nil {
		abs, _, err := w.ResolvePath(path)
		return abs, err
	}
	full := path
	if !filepath.IsAbs(full) {
		full = filepath.Join(workdir, full)
	}
	full = filepath.Clean(full)
	rel, err := filepath.Rel(workdir, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q is outside the project directory", path)
	}
	return full, nil
}
