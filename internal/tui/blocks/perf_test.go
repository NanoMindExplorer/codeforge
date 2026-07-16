package blocks

import (
	"fmt"
	"strings"
	"testing"
)

func TestViewportScalesToManyBlocks(t *testing.T) {
	s := NewStore()
	s.SetSize(100, 40)
	// large history without multi-second CI cost
	for i := 0; i < 800; i++ {
		s.AddSystem(fmt.Sprintf("line %d %s", i, strings.Repeat("x", 40)))
	}
	// View must not panic and should be fast enough for interactive use
	view := s.View()
	if view == "" {
		t.Fatal("empty view")
	}
	// total lines computed via cache
	n := s.totalLines()
	if n < 800 {
		t.Fatalf("total lines %d", n)
	}
	// second call hits cache
	_ = s.View()
}

func TestBodyLineCap(t *testing.T) {
	s := NewStore()
	s.SetSize(80, 30)
	big := strings.Repeat("word ", 5000)
	s.AddUser(big)
	// expand is default
	lines := s.renderBlockLines(0)
	// header + body + gap
	bodyCount := 0
	for _, ln := range lines {
		if strings.Contains(ln, "…") || strings.Contains(ln, "lines") {
			bodyCount++
		}
	}
	if len(lines) > MaxBodyLines+5 {
		t.Fatalf("body not capped: %d lines", len(lines))
	}
}
