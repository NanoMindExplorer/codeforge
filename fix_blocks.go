package main

import (
	"os"
	"strings"
)

func main() {
	bBytes, _ := os.ReadFile("internal/tui/blocks/block.go")
	bStr := string(bBytes)
	bStr = strings.ReplaceAll(bStr, "CachedBody  string", "CachedBody  string\n\tCachedVisual      []string\n\tCachedVisualWidth int\n\tCachedVisualSel   bool")
	os.WriteFile("internal/tui/blocks/block.go", []byte(bStr), 0644)

	rBytes, _ := os.ReadFile("internal/tui/blocks/render.go")
	rStr := string(rBytes)
	
	start := strings.Index(rStr, "func (s *Store) renderBlockLines(i int) []string {")
	end := strings.Index(rStr, "func (s *Store) bodyLines(i int, width int) []string {")
	
	newRender := `func (s *Store) renderBlockLines(i int) []string {
	if i < 0 || i >= len(s.blocks) {
		return nil
	}
	b := &s.blocks[i]
	t := theme.Current()
	selected := s.showSelection && i == s.selected
	innerW := s.width - 4 // accent + scrollbar + pad
	if innerW < 12 {
		innerW = 12
	}

	if b.CachedVisual != nil && b.CachedVisualWidth == s.width && b.CachedVisualSel == selected && !b.Streaming && b.CachedBody == b.Body {
		return b.CachedVisual
	}

	accent := t.AccentSystem
	switch b.Kind {
	case KindUser:
		accent = t.AccentUser
	case KindAssistant, KindThinking:
		accent = t.AccentAssistant
	case KindToolCall, KindToolResult, KindDiff:
		accent = t.AccentTool
	}
	pfx := theme.BlockPrefix(accent)
	if selected {
		pfx = lipgloss.NewStyle().Foreground(t.AccentFocus).Bold(true).Render("┃ ")
	}

	var lines []string
	header := s.blockHeader(*b)
	if selected {
		header = lipgloss.NewStyle().Background(t.BgElevated).Foreground(t.AccentFocus).Render(header)
	} else {
		header = lipgloss.NewStyle().Foreground(headerColor(*b, t)).Render(header)
	}
	lines = append(lines, pfx+header)

	if b.Collapsed && b.Foldable {
		// collapsed: one summary line only (already header)
		if b.Kind == KindToolResult || b.Kind == KindDiff {
			sum := strings.TrimSpace(strings.ReplaceAll(b.Body, "\n", " "))
			if len(sum) > 50 {
				sum = sum[:50] + "…"
			}
			if sum != "" {
				lines = append(lines, pfx+lipgloss.NewStyle().Foreground(t.TextMuted).Render("  "+sum))
			}
		}
		lines = append(lines, "")
		
		if !b.Streaming {
			b.CachedVisual = lines
			b.CachedVisualWidth = s.width
			b.CachedVisualSel = selected
		}
		return lines
	}

	// expanded body
	bodyLines := s.bodyLines(i, innerW)
	for _, ln := range bodyLines {
		if b.Kind == KindUser {
			ln = lipgloss.NewStyle().Background(t.BgLight).Foreground(t.TextPrimary).Width(innerW).Render(ln)
		}
		lines = append(lines, pfx+ln)
	}
	lines = append(lines, "") // gap after block
	
	if !b.Streaming {
		b.CachedVisual = lines
		b.CachedVisualWidth = s.width
		b.CachedVisualSel = selected
	}
	
	return lines
}

`
	rStr = rStr[:start] + newRender + rStr[end:]
	os.WriteFile("internal/tui/blocks/render.go", []byte(rStr), 0644)
}
