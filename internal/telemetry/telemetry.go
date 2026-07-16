// Package telemetry is privacy-first, opt-in analytics.
// Default: OFF. When enabled, events go to a local JSONL file and optionally
// a user-configured HTTPS endpoint. No content of source files or prompts
// is ever sent — only coarse event names and counters.
package telemetry

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config from application config / env.
type Config struct {
	Enabled  bool   `mapstructure:"enabled" json:"enabled"`
	Endpoint string `mapstructure:"endpoint" json:"endpoint"` // optional HTTPS
	// LocalOnly forces file-only even if endpoint set
	LocalOnly bool `mapstructure:"local_only" json:"local_only"`
}

// Client buffers events.
type Client struct {
	cfg    Config
	mu     sync.Mutex
	events []map[string]any
	path   string
}

// New creates a client. Disabled clients are no-ops.
func New(cfg Config) *Client {
	// Env override
	if os.Getenv("CODEFORGE_TELEMETRY") == "1" || os.Getenv("CODEFORGE_TELEMETRY") == "true" {
		cfg.Enabled = true
	}
	if os.Getenv("CODEFORGE_TELEMETRY") == "0" {
		cfg.Enabled = false
	}
	if ep := os.Getenv("CODEFORGE_TELEMETRY_URL"); ep != "" {
		cfg.Endpoint = ep
	}
	c := &Client{cfg: cfg}
	if !cfg.Enabled {
		return c
	}
	if home, err := os.UserHomeDir(); err == nil {
		dir := filepath.Join(home, ".codeforge", "telemetry")
		_ = os.MkdirAll(dir, 0755)
		c.path = filepath.Join(dir, "events.jsonl")
	}
	return c
}

// Event records a coarse event (no PII / no file contents).
func (c *Client) Event(name string, props map[string]any) {
	if c == nil || !c.cfg.Enabled {
		return
	}
	ev := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339),
		"event": name,
		"v":     1,
	}
	// only allow safe prop types
	for k, v := range props {
		switch v.(type) {
		case string, bool, int, int64, float64:
			// truncate long strings
			if s, ok := v.(string); ok && len(s) > 80 {
				v = s[:80]
			}
			ev[k] = v
		}
	}
	c.mu.Lock()
	c.events = append(c.events, ev)
	// flush every 20 events
	n := len(c.events)
	c.mu.Unlock()
	if n >= 20 {
		c.Flush()
	}
}

// Flush writes buffered events to disk and optional endpoint.
func (c *Client) Flush() {
	if c == nil || !c.cfg.Enabled {
		return
	}
	c.mu.Lock()
	batch := c.events
	c.events = nil
	c.mu.Unlock()
	if len(batch) == 0 {
		return
	}
	if c.path != "" {
		f, err := os.OpenFile(c.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			enc := json.NewEncoder(f)
			for _, ev := range batch {
				_ = enc.Encode(ev)
			}
			_ = f.Close()
		}
	}
	if c.cfg.Endpoint != "" && !c.cfg.LocalOnly {
		go post(c.cfg.Endpoint, batch)
	}
}

func post(endpoint string, batch []map[string]any) {
	body, _ := json.Marshal(map[string]any{"events": batch})
	ctxClient := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CodeForge-Telemetry/1")
	resp, err := ctxClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}
