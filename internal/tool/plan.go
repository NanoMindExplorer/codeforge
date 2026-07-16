package tool

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// PlanSignal is set by plan tools so the TUI can open approval / switch modes.
type PlanSignal struct {
	Kind    string // exit_plan_mode | enter_plan_mode | plan_written
	Message string
}

var (
	planSigMu sync.Mutex
	planSig   *PlanSignal
)

// ConsumePlanSignal returns and clears the last plan tool signal.
func ConsumePlanSignal() *PlanSignal {
	planSigMu.Lock()
	defer planSigMu.Unlock()
	s := planSig
	planSig = nil
	return s
}

// PeekPlanSignal returns the signal without clearing.
func PeekPlanSignal() *PlanSignal {
	planSigMu.Lock()
	defer planSigMu.Unlock()
	return planSig
}

func setPlanSignal(kind, msg string) {
	planSigMu.Lock()
	planSig = &PlanSignal{Kind: kind, Message: msg}
	planSigMu.Unlock()
}

// WritePlan writes the design plan to the session plan.md file.
// Always allowed (even outside Design mode) so the agent can draft plans.
type WritePlan struct {
	Staged *StagedWriter
}

func (w *WritePlan) Name() string { return "write_plan" }
func (w *WritePlan) Description() string {
	return `Write or replace the design plan file (plan.md in the session directory).
Use during DESIGN mode after exploring the codebase. Structure:
## Context
## Approach
## Critical files
## Verification`
}
func (w *WritePlan) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"content": map[string]any{
				"type":        "string",
				"description": "Full markdown content of the plan",
			},
			"append": map[string]any{
				"type":        "boolean",
				"description": "If true, append to existing plan instead of replacing",
			},
		},
		"required": []string{"content"},
	}
}

type writePlanInput struct {
	Content string `json:"content"`
	Append  bool   `json:"append"`
}

func (w *WritePlan) Execute(input json.RawMessage) Result {
	var in writePlanInput
	if err := json.Unmarshal(input, &in); err != nil {
		return Result{Error: fmt.Sprintf("invalid: %v", err)}
	}
	if strings.TrimSpace(in.Content) == "" {
		return Result{Error: "content required"}
	}
	if w.Staged == nil || w.Staged.PlanPath() == "" {
		return Result{Error: "plan path not configured — session may not be ready"}
	}
	path := w.Staged.PlanPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return Result{Error: err.Error()}
	}
	content := in.Content
	if in.Append {
		if old, err := os.ReadFile(path); err == nil && len(old) > 0 {
			content = string(old) + "\n" + content
		}
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return Result{Error: err.Error()}
	}
	setPlanSignal("plan_written", path)
	n := len(strings.Split(strings.TrimSpace(content), "\n"))
	return Result{
		Success: true,
		Output:  fmt.Sprintf("Plan written (%d lines) → %s\nWhen ready, call exit_plan_mode to present for approval.", n, path),
	}
}

// ExitPlanMode signals the TUI to open the plan approval surface.
type ExitPlanMode struct {
	Staged *StagedWriter
}

func (e *ExitPlanMode) Name() string { return "exit_plan_mode" }
func (e *ExitPlanMode) Description() string {
	return `Present the design plan for user approval. Call this when the plan is complete.
The TUI opens a review UI: user presses a=approve, s=request changes, q=quit plan.
Do not implement code until the plan is approved.`
}
func (e *ExitPlanMode) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary": map[string]any{
				"type":        "string",
				"description": "One-line summary of the plan for the user",
			},
		},
	}
}

type exitPlanInput struct {
	Summary string `json:"summary"`
}

func (e *ExitPlanMode) Execute(input json.RawMessage) Result {
	var in exitPlanInput
	_ = json.Unmarshal(input, &in)
	path := ""
	if e.Staged != nil {
		path = e.Staged.PlanPath()
	}
	// Ensure plan file exists (empty plan still opens approval UI)
	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			_ = os.MkdirAll(filepath.Dir(path), 0755)
			_ = os.WriteFile(path, []byte("# Plan\n\n_(no plan content written yet)_\n"), 0644)
		}
	}
	msg := in.Summary
	if msg == "" {
		msg = "Plan ready for review"
	}
	setPlanSignal("exit_plan_mode", msg)
	return Result{
		Success: true,
		Output:  "Plan presented for approval. Waiting for user (a=approve · s=changes · q=quit).",
	}
}

// EnterPlanMode requests the session switch into DESIGN mode (requires user may already be there).
type EnterPlanMode struct {
	Staged *StagedWriter
}

func (e *EnterPlanMode) Name() string { return "enter_plan_mode" }
func (e *EnterPlanMode) Description() string {
	return `Request DESIGN (plan) mode for an ambiguous task. Prefer when architecture choices need user buy-in before coding.
User can also enter with /plan or Shift+Tab.`
}
func (e *EnterPlanMode) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"reason": map[string]any{
				"type":        "string",
				"description": "Why plan mode is appropriate for this task",
			},
		},
	}
}

type enterPlanInput struct {
	Reason string `json:"reason"`
}

func (e *EnterPlanMode) Execute(input json.RawMessage) Result {
	var in enterPlanInput
	_ = json.Unmarshal(input, &in)
	msg := in.Reason
	if msg == "" {
		msg = "Agent requested design plan mode"
	}
	setPlanSignal("enter_plan_mode", msg)
	return Result{
		Success: true,
		Output:  "Requested DESIGN mode. Explore the codebase, write_plan, then exit_plan_mode.",
	}
}

// ReadPlanFile loads plan.md content if present.
func ReadPlanFile(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("no plan path")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
