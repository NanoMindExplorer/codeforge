package github

import (
	"context"
	"io"
	"net/http"
)

func newDiffRequest(ctx context.Context, url, token string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3.diff")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return req, nil
}

func readAllLimit(r io.Reader, max int) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, int64(max)))
}
