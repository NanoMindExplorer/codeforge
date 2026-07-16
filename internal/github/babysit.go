package github

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CheckStatus is a normalized CI rollup.
type CheckStatus struct {
	AllGreen bool
	Pending  int
	Failed   int
	Passed   int
	Raw      string
	Summary  string
}

// ParseChecksOutput heuristically classifies `gh pr checks` text output.
func ParseChecksOutput(raw string) CheckStatus {
	cs := CheckStatus{Raw: raw}
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		l := strings.ToLower(line)
		if strings.TrimSpace(line) == "" {
			continue
		}
		switch {
		case strings.Contains(l, "pass") || strings.Contains(l, "success"):
			cs.Passed++
		case strings.Contains(l, "fail") || strings.Contains(l, "error") || strings.Contains(l, "cancelled"):
			cs.Failed++
		case strings.Contains(l, "pending") || strings.Contains(l, "queued") || strings.Contains(l, "progress") || strings.Contains(l, "waiting"):
			cs.Pending++
		}
	}
	cs.AllGreen = cs.Failed == 0 && cs.Pending == 0 && (cs.Passed > 0 || strings.TrimSpace(raw) == "" || strings.Contains(strings.ToLower(raw), "no checks"))
	// If no checks reported at all, treat as green-with-warning
	if cs.Passed == 0 && cs.Failed == 0 && cs.Pending == 0 {
		cs.AllGreen = true
		cs.Summary = "No failing checks detected"
	} else {
		cs.Summary = fmt.Sprintf("passed=%d failed=%d pending=%d", cs.Passed, cs.Failed, cs.Pending)
	}
	return cs
}

// BabysitOptions configures PR check polling.
type BabysitOptions struct {
	PRNumber   int
	Interval   time.Duration
	Timeout    time.Duration
	OnProgress func(CheckStatus) // optional callback each poll
}

// Babysit polls CI checks until green, failed (terminal), or timeout.
func (c *Client) Babysit(ctx context.Context, opt BabysitOptions) (CheckStatus, error) {
	if opt.Interval <= 0 {
		opt.Interval = 20 * time.Second
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 15 * time.Minute
	}
	deadline := time.Now().Add(opt.Timeout)
	var last CheckStatus
	for {
		if err := ctx.Err(); err != nil {
			return last, err
		}
		raw, err := c.Checks(ctx, opt.PRNumber)
		if err != nil {
			last = CheckStatus{Summary: "error: " + err.Error(), Raw: err.Error()}
			if opt.OnProgress != nil {
				opt.OnProgress(last)
			}
		} else {
			last = ParseChecksOutput(raw)
			if opt.OnProgress != nil {
				opt.OnProgress(last)
			}
			if last.Failed > 0 && last.Pending == 0 {
				return last, fmt.Errorf("checks failed: %s", last.Summary)
			}
			if last.AllGreen {
				return last, nil
			}
		}
		if time.Now().After(deadline) {
			return last, fmt.Errorf("babysit timeout after %s — last: %s", opt.Timeout, last.Summary)
		}
		select {
		case <-ctx.Done():
			return last, ctx.Err()
		case <-time.After(opt.Interval):
		}
	}
}

// BabysitOnce returns a single snapshot (for agent tool without long poll).
func (c *Client) BabysitOnce(ctx context.Context, prNumber int) (CheckStatus, error) {
	raw, err := c.Checks(ctx, prNumber)
	if err != nil {
		return CheckStatus{}, err
	}
	return ParseChecksOutput(raw), nil
}

// FormatCheckStatus for chat display.
func FormatCheckStatus(cs CheckStatus) string {
	icon := "⏳"
	if cs.AllGreen {
		icon = "✅"
	} else if cs.Failed > 0 {
		icon = "❌"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s CI %s\n", icon, cs.Summary))
	if cs.Raw != "" {
		// cap raw
		raw := cs.Raw
		if len(raw) > 4000 {
			raw = raw[:4000] + "\n… (truncated)"
		}
		b.WriteString("\n")
		b.WriteString(raw)
	}
	return b.String()
}
