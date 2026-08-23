// Package provider — MCP (Model Context Protocol) client scaffold.
// Allows external tools to be registered into the tool registry at runtime.
// Full stdio/SSE MCP transport is intentionally thin: tools are registered
// as generic HTTP/stdio bridges when a config entry exists.
package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// MCPServerConfig describes one MCP server (stdio).
type MCPServerConfig struct {
	Name string `json:"name" yaml:"name"`
	// Transport can be "stdio", "websocket", or "sse"
	Transport string            `json:"transport" yaml:"transport"`
	Command   string            `json:"command" yaml:"command"`
	Args      []string          `json:"args" yaml:"args"`
	Env       map[string]string `json:"env" yaml:"env"`
	// URL for websocket or sse transport
	URL string `json:"url" yaml:"url"`
}

// MCPClient is a minimal JSON-RPC over stdio client for MCP tool listing/calls.
type MCPClient struct {
	cfg    MCPServerConfig
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Reader
	ws     *websocket.Conn
	mu     sync.Mutex
	nextID int
}

// MCPToolDef is a tool exposed by an MCP server.
type MCPToolDef struct {
	Name        string
	Description string
	InputSchema any
	Server      string
}

// Connect starts the MCP server process and initializes the session.
func ConnectMCP(cfg MCPServerConfig) (*MCPClient, error) {
	if cfg.Transport == "" {
		cfg.Transport = "stdio"
	}
	if cfg.Transport == "sse" {
		return nil, fmt.Errorf("mcp: sse transport is experimental and currently unimplemented natively")
	}

	c := &MCPClient{
		cfg:    cfg,
		nextID: 1,
	}

	if cfg.Transport == "websocket" {
		if cfg.URL == "" {
			return nil, fmt.Errorf("mcp: url required for websocket transport")
		}
		u, err := url.Parse(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("mcp invalid ws url: %v", err)
		}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("mcp ws dial failed: %v", err)
		}
		c.ws = conn
	} else {
		// Stdio
		if cfg.Command == "" {
			return nil, fmt.Errorf("mcp: command required for stdio transport")
		}
		cmd := exec.Command(cfg.Command, cfg.Args...)
		cmd.Env = os.Environ()
		for k, v := range cfg.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("mcp start %s: %w", cfg.Name, err)
		}
		c.cmd = cmd
		c.stdin = bufio.NewWriter(stdin)
		c.stdout = bufio.NewReader(stdout)
	}
	// initialize
	_, err := c.request(context.Background(), "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "codeforge", "version": "1.9.4"},
	})
	if err != nil {
		_ = c.Close()
		return nil, err
	}
	// notifications/initialized (best-effort, ignore errors)
	_ = c.notify("notifications/initialized", map[string]any{})
	return c, nil
}

func (c *MCPClient) Close() error {
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

// ListTools returns tools from tools/list.
func (c *MCPClient) ListTools(ctx context.Context) ([]MCPToolDef, error) {
	raw, err := c.request(ctx, "tools/list", map[string]any{})
	if err != nil {
		return nil, err
	}
	var result struct {
		Tools []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			InputSchema any    `json:"inputSchema"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	out := make([]MCPToolDef, 0, len(result.Tools))
	for _, t := range result.Tools {
		out = append(out, MCPToolDef{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
			Server:      c.cfg.Name,
		})
	}
	return out, nil
}

// CallTool invokes tools/call.
func (c *MCPClient) CallTool(ctx context.Context, name string, args map[string]any) (string, error) {
	raw, err := c.request(ctx, "tools/call", map[string]any{
		"name":      name,
		"arguments": args,
	})
	if err != nil {
		return "", err
	}
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return string(raw), nil
	}
	var sb strings.Builder
	for _, c := range result.Content {
		if c.Text != "" {
			sb.WriteString(c.Text)
		}
	}
	if result.IsError {
		return sb.String(), fmt.Errorf("mcp tool error")
	}
	return sb.String(), nil
}

func (c *MCPClient) request(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextID
	c.nextID++
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	if c.cfg.Transport == "websocket" {
		if err := c.ws.WriteJSON(msg); err != nil {
			return nil, err
		}
	} else {
		data, _ := json.Marshal(msg)
		data = append(data, '\n')
		if _, err := c.stdin.Write(data); err != nil {
			return nil, err
		}
		if err := c.stdin.Flush(); err != nil {
			return nil, err
		}
	}

	deadline := time.Now().Add(30 * time.Second)
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	}
	for time.Now().Before(deadline) {
		var resp struct {
			ID     int             `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if c.cfg.Transport == "websocket" {
			if err := c.ws.SetReadDeadline(deadline); err != nil {
				return nil, err
			}
			if err := c.ws.ReadJSON(&resp); err != nil {
				return nil, err
			}
		} else {
			line, err := c.stdout.ReadBytes('\n')
			if err != nil {
				return nil, err
			}
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			if err := json.Unmarshal(line, &resp); err != nil {
				continue
			}
		}
		if resp.ID != id {
			continue
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("mcp: %s", resp.Error.Message)
		}
		return resp.Result, nil
	}
	return nil, fmt.Errorf("mcp: timeout waiting for %s", method)
}

func (c *MCPClient) notify(method string, params any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	data, _ := json.Marshal(msg)
	data = append(data, '\n')
	if _, err := c.stdin.Write(data); err != nil {
		return err
	}
	return c.stdin.Flush()
}
