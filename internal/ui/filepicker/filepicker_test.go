package filepicker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLineRangeParse(t *testing.T) {
	if !isLineRange("10-50") || !isLineRange("3") {
		t.Fatal("range")
	}
	s, e := parseLineRange("10-50")
	if s != 10 || e != 50 {
		t.Fatal(s, e)
	}
}

func TestReadFileContentRange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.go")
	body := "line1\nline2\nline3\nline4\n"
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := ReadFileContent(dir, "a.go:2-3", 1000)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "line2") || !strings.Contains(out, "line3") {
		t.Fatal(out)
	}
	if strings.Contains(out, "line1") && strings.Count(out, "line1") > 0 {
		// header may not include line1 as content
		if strings.Contains(out, "\nline1\n") {
			t.Fatal("should not include line1 body")
		}
	}
}

func TestListRespectsGitignore(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("secret.txt\n"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "ok.go"), []byte("x"), 0644)
	files := listFiles(dir, false)
	for _, f := range files {
		if f == "secret.txt" {
			t.Fatal("gitignore leak")
		}
	}
}
