// Package plugin loads third-party tools from YAML manifests without forking
// CodeForge. Plugins are pure command bridges (no CGo plugins).
package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeforge/tui/internal/tool"
	"gopkg.in/yaml.v3"
)

// Manifest is plugin.yaml / *.plugin.yaml shape.
type Manifest struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Version     string            `yaml:"version"`
	Command     string            `yaml:"command"` // executable
	Args        []string          `yaml:"args"`   // static args before JSON stdin
	Env         map[string]string `yaml:"env"`
	TimeoutSec  int               `yaml:"timeout_sec"`
	// InputSchema optional JSON-schema-like map for the agent
	InputSchema map[string]any `yaml:"input_schema"`
	// WorkdirRelative when true, runs with cwd = CodeForge workdir
	WorkdirRelative bool `yaml:"workdir_relative"`
}

// ExecTool bridges a plugin command into the tool registry.
type ExecTool struct {
	Manifest Manifest
	WorkDir  string
}

func (e *ExecTool) Name() string {
	n := e.Manifest.Name
	if n == "" {
		n = "plugin"
	}
	if !strings.HasPrefix(n, "plugin_") {
		n = "plugin_" + sanitize(n)
	}
	return n
}

func (e *ExecTool) Description() string {
	d := e.Manifest.Description
	if d == "" {
		d = "Plugin tool " + e.Manifest.Name
	}
	return "[plugin] " + d
}

func (e *ExecTool) Schema() map[string]any {
	if e.Manifest.InputSchema != nil {
		return e.Manifest.InputSchema
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"input": map[string]any{"type": "string", "description": "Free-form input for the plugin"},
		},
	}
}

func (e *ExecTool) Execute(input json.RawMessage) tool.Result {
	if e.Manifest.Command == "" {
		return tool.Result{Error: "plugin command empty"}
	}
	timeout := time.Duration(e.Manifest.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := append([]string{}, e.Manifest.Args...)
	cmd := exec.CommandContext(ctx, e.Manifest.Command, args...)
	if e.Manifest.WorkdirRelative || e.WorkDir != "" {
		cmd.Dir = e.WorkDir
	}
	env := os.Environ()
	for k, v := range e.Manifest.Env {
		env = append(env, k+"="+v)
	}
	cmd.Env = env
	cmd.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := stdout.String()
	if errOut := stderr.String(); errOut != "" {
		if out != "" {
			out += "\n"
		}
		out += errOut
	}
	if len(out) > 32_000 {
		out = out[:32_000] + "\n… (truncated)"
	}
	if err != nil {
		return tool.Result{Success: false, Error: err.Error(), Output: out}
	}
	return tool.Result{Success: true, Output: out}
}

// LoadAll discovers plugins from default + configured directories.
// Returns tool count and status lines.
func LoadAll(reg *tool.Registry, workdir string, extraDirs []string) (int, []string) {
	var lines []string
	dirs := defaultPluginDirs()
	dirs = append(dirs, extraDirs...)
	// project-local
	dirs = append(dirs, filepath.Join(workdir, ".codeforge", "plugins"))

	seen := map[string]bool{}
	count := 0
	for _, dir := range dirs {
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			path := filepath.Join(dir, e.Name())
			if e.IsDir() {
				// look for plugin.yaml inside
				for _, name := range []string{"plugin.yaml", "plugin.yml", e.Name() + ".yaml"} {
					p := filepath.Join(path, name)
					if n, line := loadOne(reg, workdir, p); n > 0 {
						count += n
						lines = append(lines, line)
					}
				}
				continue
			}
			if strings.HasSuffix(e.Name(), ".plugin.yaml") || strings.HasSuffix(e.Name(), ".plugin.yml") ||
				e.Name() == "plugin.yaml" || e.Name() == "plugin.yml" {
				if n, line := loadOne(reg, workdir, path); n > 0 {
					count += n
					lines = append(lines, line)
				}
			}
		}
	}
	return count, lines
}

func loadOne(reg *tool.Registry, workdir, path string) (int, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, ""
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return 0, fmt.Sprintf("⚠ plugin %s: %v", path, err)
	}
	if m.Name == "" || m.Command == "" {
		return 0, fmt.Sprintf("⚠ plugin %s: name/command required", path)
	}
	reg.Register(&ExecTool{Manifest: m, WorkDir: workdir})
	return 1, fmt.Sprintf("✓ plugin: %s (%s)", m.Name, path)
}

func defaultPluginDirs() []string {
	var dirs []string
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".codeforge", "plugins"))
	}
	if x := os.Getenv("CODEFORGE_PLUGIN_DIR"); x != "" {
		dirs = append(dirs, x)
	}
	return dirs
}

func sanitize(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}
