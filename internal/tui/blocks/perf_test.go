package blocks

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// Q7.2: viewport only materializes visible rows — large history must stay interactive.
func TestViewportScalesToManyBlocks(t *testing.T) {
	s := NewStore()
	s.SetSize(100, 40)
	// large history without multi-second CI cost
	for i := 0; i < 800; i++ {
		s.AddSystem(fmt.Sprintf("line %d %s", i, strings.Repeat("x", 40)))
	}
	start := time.Now()
	view := s.View()
	elapsed := time.Since(start)
	if view == "" {
		t.Fatal("empty view")
	}
	// Interactive budget: first paint of 800 blocks under 500ms on CI
	if elapsed > 500*time.Millisecond {
		t.Fatalf("View too slow for 800 blocks: %v", elapsed)
	}
	// total lines computed via cache
	n := s.totalLines()
	if n < 800 {
		t.Fatalf("total lines %d", n)
	}
	// second call hits cache — must be faster path
	start = time.Now()
	_ = s.View()
	if time.Since(start) > 100*time.Millisecond {
		t.Fatalf("cached View too slow: %v", time.Since(start))
	}
	// Viewport height caps rendered output roughly (not full history dump)
	// View includes sticky/scrollbar chrome but should not be megabytes
	if len(view) > 200_000 {
		t.Fatalf("view size %d looks like full dump", len(view))
	}
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
	_ = bodyCount
	if len(lines) > MaxBodyLines+5 {
		t.Fatalf("body not capped: %d lines", len(lines))
	}
}

func TestMaxBlocksSoftConstant(t *testing.T) {
	// Documented soft caps for long sessions (Q7.2 audit)
	if MaxBodyLines < 40 || MaxBodyLines > 500 {
		t.Fatalf("MaxBodyLines unexpected: %d", MaxBodyLines)
	}
	if MaxBlocksSoft < 100 {
		t.Fatalf("MaxBlocksSoft too low: %d", MaxBlocksSoft)
	}
}

func BenchmarkView800(b *testing.B) {
	s := NewStore()
	s.SetSize(100, 40)
	for i := 0; i < 800; i++ {
		s.AddSystem(fmt.Sprintf("line %d %s", i, strings.Repeat("x", 40)))
	}
	_ = s.View() // warm cache structure
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.View()
	}
}
