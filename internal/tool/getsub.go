package tool

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GetSubagentOutput polls a tracked subagent job (Grok get_command_or_subagent_output).
type GetSubagentOutput struct{}

func (g *GetSubagentOutput) Name() string { return "get_subagent_output" }
func (g *GetSubagentOutput) Description() string {
	return `Get status/output of a background or recent subagent job by id (e.g. sub-1).
Also works for completed sync subagents (they are recorded with an id).
Alias: get_command_or_subagent_output.
Optional wait_ms blocks until done or timeout (max 120000).`
}
func (g *GetSubagentOutput) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "Subagent job id (sub-N)",
			},
			"task_id": map[string]any{
				"type":        "string",
				"description": "Alias for id",
			},
			"wait_ms": map[string]any{
				"type":        "integer",
				"description": "Optional wait for completion (0 = no wait, max 120000)",
			},
		},
		"required": []string{},
	}
}

type getSubInput struct {
	ID     string `json:"id"`
	TaskID string `json:"task_id"`
	WaitMS int    `json:"wait_ms"`
}

func (g *GetSubagentOutput) Execute(input json.RawMessage) Result {
	var in getSubInput
	if err := json.Unmarshal(input, &in); err != nil {
		return Result{Error: err.Error()}
	}
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = strings.TrimSpace(in.TaskID)
	}
	if id == "" {
		// list if no id
		return Result{Success: true, Output: SubJobs.Summary()}
	}

	wait := in.WaitMS
	if wait < 0 {
		wait = 0
	}
	if wait > 120_000 {
		wait = 120_000
	}

	deadline := time.Now().Add(time.Duration(wait) * time.Millisecond)
	for {
		j, ok := SubJobs.Get(id)
		if !ok {
			return Result{Error: fmt.Sprintf("unknown subagent id %s — use /subagents or get_subagent_output without id to list", id)}
		}
		if j.Status != SubRunning || wait == 0 {
			out := FormatJobOutput(j)
			okRes := j.Status == SubSucceeded || j.Status == SubRunning
			if j.Status == SubFailed {
				return Result{Success: false, Output: out, Error: j.Error}
			}
			if j.Status == SubCancelled {
				return Result{Success: false, Output: out, Error: "cancelled"}
			}
			return Result{Success: okRes || j.Status == SubSucceeded, Output: out}
		}
		if time.Now().After(deadline) {
			return Result{Success: true, Output: FormatJobOutput(j) + "\n\n(wait timed out — still running)"}
		}
		time.Sleep(150 * time.Millisecond)
	}
}
