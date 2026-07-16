package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// PRDiff returns the diff for a PR (or current branch PR).
func (c *Client) PRDiff(ctx context.Context, number int) (string, error) {
	if c.HasCLI() {
		if number > 0 {
			return c.runGH(ctx, "pr", "diff", strconv.Itoa(number))
		}
		return c.runGH(ctx, "pr", "diff")
	}
	slug, err := c.RepoSlug(ctx)
	if err != nil {
		return "", err
	}
	if number <= 0 {
		return "", fmt.Errorf("PR number required without gh CLI")
	}
	// Accept header for diff
	path := fmt.Sprintf("/repos/%s/pulls/%d", slug, number)
	// use gh-less: request via REST with Accept diff — custom
	raw, err := c.restDiff(ctx, path)
	if err != nil {
		return "", err
	}
	return raw, nil
}

func (c *Client) restDiff(ctx context.Context, path string) (string, error) {
	if c.HasCLI() {
		return c.runGH(ctx, "api", path, "-H", "Accept: application/vnd.github.v3.diff")
	}
	if c.Token == "" {
		return "", fmt.Errorf("no auth for PR diff")
	}
	req, err := newDiffRequest(ctx, c.Host+path, c.Token)
	if err != nil {
		return "", err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := readAllLimit(resp.Body, 512*1024)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("diff %s: %s", resp.Status, truncate(string(b), 200))
	}
	return string(b), nil
}

// CommentOnPR posts a comment on a pull request (issue comments API).
func (c *Client) CommentOnPR(ctx context.Context, number int, body string) (string, error) {
	if number <= 0 || strings.TrimSpace(body) == "" {
		return "", fmt.Errorf("number and body required")
	}
	if c.HasCLI() {
		out, err := c.runGH(ctx, "pr", "comment", strconv.Itoa(number), "--body", body)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(out), nil
	}
	slug, err := c.RepoSlug(ctx)
	if err != nil {
		return "", err
	}
	raw, err := c.REST(ctx, "POST", fmt.Sprintf("/repos/%s/issues/%d/comments", slug, number), map[string]any{
		"body": body,
	})
	if err != nil {
		return "", err
	}
	var res struct {
		HTMLURL string `json:"html_url"`
	}
	_ = json.Unmarshal(raw, &res)
	return res.HTMLURL, nil
}

// CommentOnIssue posts an issue comment.
func (c *Client) CommentOnIssue(ctx context.Context, number int, body string) (string, error) {
	if number <= 0 || strings.TrimSpace(body) == "" {
		return "", fmt.Errorf("number and body required")
	}
	if c.HasCLI() {
		out, err := c.runGH(ctx, "issue", "comment", strconv.Itoa(number), "--body", body)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(out), nil
	}
	slug, err := c.RepoSlug(ctx)
	if err != nil {
		return "", err
	}
	raw, err := c.REST(ctx, "POST", fmt.Sprintf("/repos/%s/issues/%d/comments", slug, number), map[string]any{
		"body": body,
	})
	if err != nil {
		return "", err
	}
	var res struct {
		HTMLURL string `json:"html_url"`
	}
	_ = json.Unmarshal(raw, &res)
	return res.HTMLURL, nil
}

// RequestReviewers requests PR reviews from users (comma-separated logins).
func (c *Client) RequestReviewers(ctx context.Context, number int, reviewers []string) (string, error) {
	if number <= 0 || len(reviewers) == 0 {
		return "", fmt.Errorf("number and reviewers required")
	}
	if c.HasCLI() {
		args := []string{"pr", "edit", strconv.Itoa(number)}
		for _, r := range reviewers {
			args = append(args, "--add-reviewer", r)
		}
		return c.runGH(ctx, args...)
	}
	slug, err := c.RepoSlug(ctx)
	if err != nil {
		return "", err
	}
	raw, err := c.REST(ctx, "POST", fmt.Sprintf("/repos/%s/pulls/%d/requested_reviewers", slug, number), map[string]any{
		"reviewers": reviewers,
	})
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// ReadyPR marks a draft PR as ready for review.
func (c *Client) ReadyPR(ctx context.Context, number int) (string, error) {
	if c.HasCLI() {
		if number > 0 {
			return c.runGH(ctx, "pr", "ready", strconv.Itoa(number))
		}
		return c.runGH(ctx, "pr", "ready")
	}
	return "", fmt.Errorf("pr ready requires gh CLI")
}

// PRCommits lists commits on a PR.
func (c *Client) PRCommits(ctx context.Context, number int) (string, error) {
	if c.HasCLI() {
		if number > 0 {
			return c.runGH(ctx, "pr", "view", strconv.Itoa(number), "--json", "commits",
				"--jq", ".commits[] | \"\\(.oid[0:7]) \\(.messageHeadline)\"")
		}
		return c.runGH(ctx, "pr", "view", "--json", "commits",
			"--jq", ".commits[] | \"\\(.oid[0:7]) \\(.messageHeadline)\"")
	}
	slug, err := c.RepoSlug(ctx)
	if err != nil {
		return "", err
	}
	if number <= 0 {
		return "", fmt.Errorf("PR number required")
	}
	raw, err := c.REST(ctx, "GET", fmt.Sprintf("/repos/%s/pulls/%d/commits", slug, number), nil)
	if err != nil {
		return "", err
	}
	var items []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return "", err
	}
	var b strings.Builder
	for _, it := range items {
		msg := strings.Split(it.Commit.Message, "\n")[0]
		b.WriteString(fmt.Sprintf("%s %s\n", it.SHA[:7], msg))
	}
	return b.String(), nil
}

// LinkCommitPR finds PRs associated with a commit SHA.
func (c *Client) LinkCommitPR(ctx context.Context, sha string) (string, error) {
	if sha == "" {
		return "", fmt.Errorf("sha required")
	}
	if c.HasCLI() {
		return c.runGH(ctx, "api", fmt.Sprintf("repos/{owner}/{repo}/commits/%s/pulls", sha),
			"--jq", ".[] | \"#\\(.number) \\(.title) \\(.html_url)\"")
	}
	slug, err := c.RepoSlug(ctx)
	if err != nil {
		return "", err
	}
	raw, err := c.REST(ctx, "GET", fmt.Sprintf("/repos/%s/commits/%s/pulls", slug, sha), nil)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
