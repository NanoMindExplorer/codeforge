package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	gh "github.com/codeforge/tui/internal/github"
)

// GitHubTool exposes GitHub operations to the agent (gh CLI + REST).
type GitHubTool struct {
	Client *gh.Client
}

func (g *GitHubTool) Name() string { return "github" }
func (g *GitHubTool) Description() string {
	return `Interact with GitHub for the current repository (like advanced AI coding agents).
Actions:
  auth_status, repo_view
  pr_list, pr_view, pr_create, pr_merge, pr_diff, pr_comment, pr_review_request, pr_ready, pr_commits
  issue_list, issue_view, issue_create, issue_comment
  checks, babysit (poll CI until green/fail; interval_sec, timeout_sec)
  babysit_once (single CI snapshot)
  push, pull, branch_create, log
  commit_prs (list PRs for a commit sha — field: sha)`
}

func (g *GitHubTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type": "string",
				"enum": []string{
					"auth_status", "repo_view",
					"pr_list", "pr_view", "pr_create", "pr_merge", "pr_diff",
					"pr_comment", "pr_review_request", "pr_ready", "pr_commits",
					"issue_list", "issue_view", "issue_create", "issue_comment",
					"checks", "babysit", "babysit_once",
					"push", "pull", "branch_create", "log", "commit_prs",
				},
			},
			"title":        map[string]any{"type": "string"},
			"body":         map[string]any{"type": "string"},
			"base":         map[string]any{"type": "string"},
			"head":         map[string]any{"type": "string"},
			"draft":        map[string]any{"type": "boolean"},
			"number":       map[string]any{"type": "integer"},
			"state":        map[string]any{"type": "string"},
			"method":       map[string]any{"type": "string"},
			"name":         map[string]any{"type": "string"},
			"labels":       map[string]any{"type": "string"},
			"limit":        map[string]any{"type": "integer"},
			"reviewers":    map[string]any{"type": "string", "description": "Comma-separated GitHub logins"},
			"sha":          map[string]any{"type": "string"},
			"interval_sec": map[string]any{"type": "integer", "description": "Babysit poll interval (default 20)"},
			"timeout_sec":  map[string]any{"type": "integer", "description": "Babysit timeout (default 600)"},
		},
		"required": []string{"action"},
	}
}

type githubInput struct {
	Action      string `json:"action"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Base        string `json:"base"`
	Head        string `json:"head"`
	Draft       bool   `json:"draft"`
	Number      int    `json:"number"`
	State       string `json:"state"`
	Method      string `json:"method"`
	Name        string `json:"name"`
	Labels      string `json:"labels"`
	Limit       int    `json:"limit"`
	Reviewers   string `json:"reviewers"`
	SHA         string `json:"sha"`
	IntervalSec int    `json:"interval_sec"`
	TimeoutSec  int    `json:"timeout_sec"`
}

func (g *GitHubTool) Execute(input json.RawMessage) Result {
	return g.ExecuteStream(input, nil)
}

// ExecuteStream supports progress for babysit polls.
func (g *GitHubTool) ExecuteStream(input []byte, progress ProgressFunc) Result {
	if g.Client == nil {
		return Result{Error: "GitHub client not configured"}
	}
	var in githubInput
	if err := json.Unmarshal(input, &in); err != nil {
		return Result{Error: fmt.Sprintf("invalid: %v", err)}
	}
	timeout := 90 * time.Second
	if strings.EqualFold(in.Action, "babysit") {
		timeout = 20 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	action := strings.ToLower(strings.TrimSpace(in.Action))
	var (
		out string
		err error
	)
	switch action {
	case "auth_status":
		out, err = g.Client.AuthStatus(ctx)
	case "repo_view":
		out, err = g.Client.RepoView(ctx)
	case "pr_list":
		prs, e := g.Client.ListPRs(ctx, in.State, in.Limit)
		err = e
		if err == nil {
			out = gh.FormatPRList(prs)
		}
	case "pr_view":
		out, err = g.Client.ViewPR(ctx, in.Number)
	case "pr_create":
		out, err = g.Client.CreatePR(ctx, in.Title, in.Body, in.Base, in.Head, in.Draft)
	case "pr_merge":
		if in.Number <= 0 {
			return Result{Error: "number required for pr_merge"}
		}
		out, err = g.Client.MergePR(ctx, in.Number, in.Method)
	case "pr_diff":
		out, err = g.Client.PRDiff(ctx, in.Number)
	case "pr_comment":
		out, err = g.Client.CommentOnPR(ctx, in.Number, in.Body)
	case "pr_review_request":
		var revs []string
		for _, r := range strings.Split(in.Reviewers, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				revs = append(revs, r)
			}
		}
		out, err = g.Client.RequestReviewers(ctx, in.Number, revs)
	case "pr_ready":
		out, err = g.Client.ReadyPR(ctx, in.Number)
	case "pr_commits":
		out, err = g.Client.PRCommits(ctx, in.Number)
	case "issue_list":
		issues, e := g.Client.ListIssues(ctx, in.State, in.Limit)
		err = e
		if err == nil {
			out = gh.FormatIssueList(issues)
		}
	case "issue_view":
		if in.Number <= 0 {
			return Result{Error: "number required"}
		}
		out, err = g.Client.ViewIssue(ctx, in.Number)
	case "issue_create":
		var labels []string
		for _, l := range strings.Split(in.Labels, ",") {
			l = strings.TrimSpace(l)
			if l != "" {
				labels = append(labels, l)
			}
		}
		out, err = g.Client.CreateIssue(ctx, in.Title, in.Body, labels)
	case "issue_comment":
		out, err = g.Client.CommentOnIssue(ctx, in.Number, in.Body)
	case "checks":
		out, err = g.Client.Checks(ctx, in.Number)
	case "babysit_once":
		cs, e := g.Client.BabysitOnce(ctx, in.Number)
		err = e
		if err == nil {
			out = gh.FormatCheckStatus(cs)
		}
	case "babysit":
		interval := time.Duration(in.IntervalSec) * time.Second
		to := time.Duration(in.TimeoutSec) * time.Second
		if progress != nil {
			progress(fmt.Sprintf("babysitting PR checks (pr=%d)…", in.Number))
		}
		cs, e := g.Client.Babysit(ctx, gh.BabysitOptions{
			PRNumber: in.Number,
			Interval: interval,
			Timeout:  to,
			OnProgress: func(st gh.CheckStatus) {
				if progress != nil {
					progress(st.Summary)
				}
			},
		})
		out = gh.FormatCheckStatus(cs)
		err = e
		// still return body when failed so agent can read logs
		if err != nil {
			return Result{Success: false, Error: err.Error(), Output: out}
		}
	case "push":
		if progress != nil {
			progress("git push…")
		}
		out, err = g.Client.Push(ctx, true)
	case "pull":
		out, err = g.Client.Pull(ctx)
	case "branch_create":
		out, err = g.Client.CreateBranch(ctx, in.Name)
	case "log":
		n := in.Limit
		if n == 0 {
			n = 15
		}
		out, err = g.Client.LogRecent(ctx, n)
	case "commit_prs":
		out, err = g.Client.LinkCommitPR(ctx, in.SHA)
	default:
		return Result{Error: fmt.Sprintf("unknown action %q", in.Action)}
	}
	if err != nil {
		return Result{Success: false, Error: err.Error(), Output: out}
	}
	return Result{Success: true, Output: out}
}

// ParsePRNumber is a helper for slash commands.
func ParsePRNumber(s string) (int, error) {
	s = strings.TrimPrefix(s, "#")
	return strconv.Atoi(s)
}
