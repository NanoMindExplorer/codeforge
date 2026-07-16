//go:build !plainmd

package markdown

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/muesli/reflow/wordwrap"
)

type glamRenderer = *glamour.TermRenderer

func renderGlamour(src string, width int) string {
	r, err := getRenderer(width)
	if err != nil || r == nil {
		return wordwrap.String(src, width)
	}
	out, err := r.Render(src)
	if err != nil {
		return wordwrap.String(src, width)
	}
	return strings.TrimRight(out, "\n")
}

func getRenderer(width int) (*glamour.TermRenderer, error) {
	mu.Lock()
	defer mu.Unlock()
	if disabled {
		return nil, nil
	}
	if renderer != nil && lastW == width {
		if gr, ok := renderer.(*glamour.TermRenderer); ok {
			return gr, nil
		}
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(styles.DarkStyle),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}
	renderer = r
	lastW = width
	return r, nil
}
