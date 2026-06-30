package tool

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

type Result struct {
    Success bool   `json:"success"`
    Output  string `json:"output"`
    Error   string `json:"error,omitempty"`
}

type Tool interface {
    Name() string
    Description() string
    Execute(input json.RawMessage) Result
}

type FileReader struct{ WorkDir string }

func (f *FileReader) Name() string        { return "read_file" }
func (f *FileReader) Description() string { return "Read file contents" }

type readInput struct {
    Path string `json:"path"`
}

func (f *FileReader) Execute(input json.RawMessage) Result {
    var in readInput
    if err := json.Unmarshal(input, &in); err != nil {
        return Result{Error: fmt.Sprintf("invalid: %v", err)}
    }
    if in.Path == "" {
        return Result{Error: "path required"}
    }
    path := in.Path
    if !filepath.IsAbs(path) {
        path = filepath.Join(f.WorkDir, path)
    }
    data, err := os.ReadFile(path)
    if err != nil {
        return Result{Error: fmt.Sprintf("read: %v", err)}
    }
    return Result{Success: true, Output: string(data)}
}

type FileWriter struct{ WorkDir string }

func (f *FileWriter) Name() string        { return "write_file" }
func (f *FileWriter) Description() string { return "Write content to file" }

type writeInput struct {
    Path    string `json:"path"`
    Content string `json:"content"`
}

func (f *FileWriter) Execute(input json.RawMessage) Result {
    var in writeInput
    if err := json.Unmarshal(input, &in); err != nil {
        return Result{Error: fmt.Sprintf("invalid: %v", err)}
    }
    if in.Path == "" {
        return Result{Error: "path required"}
    }
    path := in.Path
    if !filepath.IsAbs(path) {
        path = filepath.Join(f.WorkDir, path)
    }
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return Result{Error: fmt.Sprintf("mkdir: %v", err)}
    }
    if err := os.WriteFile(path, []byte(in.Content), 0644); err != nil {
        return Result{Error: fmt.Sprintf("write: %v", err)}
    }
    return Result{Success: true, Output: fmt.Sprintf("Wrote %d bytes to %s", len(in.Content), path)}
}

type DirLister struct{ WorkDir string }

func (d *DirLister) Name() string        { return "list_dir" }
func (d *DirLister) Description() string { return "List directory contents" }

type listInput struct {
    Path string `json:"path"`
}

func (d *DirLister) Execute(input json.RawMessage) Result {
    var in listInput
    if err := json.Unmarshal(input, &in); err != nil {
        return Result{Error: fmt.Sprintf("invalid: %v", err)}
    }
    path := in.Path
    if path == "" {
        path = d.WorkDir
    }
    if !filepath.IsAbs(path) {
        path = filepath.Join(d.WorkDir, path)
    }
    entries, err := os.ReadDir(path)
    if err != nil {
        return Result{Error: fmt.Sprintf("readdir: %v", err)}
    }
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("Listing: %s\n\n", path))
    for _, e := range entries {
        typ := "DIR "
        if !e.IsDir() {
            typ = "FILE"
        }
        sb.WriteString(fmt.Sprintf("[%s]  %s\n", typ, e.Name()))
    }
    return Result{Success: true, Output: sb.String()}
}

type Registry struct {
    tools map[string]Tool
}

func NewRegistry(workDir string) *Registry {
    r := &Registry{tools: make(map[string]Tool)}
    r.Register(&FileReader{WorkDir: workDir})
    r.Register(&FileWriter{WorkDir: workDir})
    r.Register(&DirLister{WorkDir: workDir})
    return r
}

func (r *Registry) Register(t Tool) {
    r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
    t, ok := r.tools[name]
    return t, ok
}

func (r *Registry) List() []Tool {
    tools := make([]Tool, 0, len(r.tools))
    for _, t := range r.tools {
        tools = append(tools, t)
    }
    return tools
}
