package gitea

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client handles HTTP API interactions with Gitea.
type Client struct {
	BaseURL string
	Token   string // legacy: Gitea PAT (token scheme). Used when AuthHeader is empty.
	// AuthHeader, when set, is the full Authorization header value applied to
	// every request (e.g. "Basic <b64>", "Bearer <token>"). Takes precedence
	// over Token so the client can carry any outbound auth mode.
	AuthHeader string
	HTTPClient *http.Client
}

// NewClient creates a new Gitea API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GiteaRepository represents the Gitea API repository payload.
type GiteaRepository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
}

// GiteaIssue represents Gitea issue payload.
type GiteaIssue struct {
	ID        int64      `json:"id"`
	Number    int64      `json:"number"`
	Title     string     `json:"title"`
	State     string     `json:"state"`
	HTMLURL   string     `json:"html_url"`
	CreatedAt time.Time  `json:"created_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	User      *GiteaUser `json:"user"`
	Assignee  *GiteaUser `json:"assignee"`
}

// GiteaUser represents a Gitea user payload.
type GiteaUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

// GiteaPullRequest represents Gitea pull request payload.
type GiteaPullRequest struct {
	ID       int64        `json:"id"`
	Number   int64        `json:"number"`
	Title    string       `json:"title"`
	State    string       `json:"state"`
	HTMLURL  string       `json:"html_url"`
	MergedAt *time.Time   `json:"merged_at"`
	ClosedAt *time.Time   `json:"closed_at"`
	User     *GiteaUser   `json:"user"`
	Head     *GiteaBranch `json:"head"`
	Base     *GiteaBranch `json:"base"`
}

// GiteaBranch represents a Gitea branch/ref payload in PRs.
type GiteaBranch struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// ListUserRepos retrieves the authenticated user's repositories using pagination.
func (c *Client) ListUserRepos(ctx context.Context) ([]GiteaRepository, error) {
	var allRepos []GiteaRepository
	page := 1
	limit := 50
	for {
		req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/user/repos", nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("page", fmt.Sprintf("%d", page))
		q.Set("limit", fmt.Sprintf("%d", limit))
		req.URL.RawQuery = q.Encode()

		var repos []GiteaRepository
		if err := c.do(req, &repos); err != nil {
			return nil, err
		}
		if len(repos) == 0 {
			break
		}
		allRepos = append(allRepos, repos...)
		if len(repos) < limit {
			break
		}
		page++
	}
	return allRepos, nil
}

// ListIssues retrieves issues for a specific repository using pagination.
func (c *Client) ListIssues(ctx context.Context, owner, repo string, state string) ([]GiteaIssue, error) {
	path := fmt.Sprintf("/api/v1/repos/%s/%s/issues", url.PathEscape(owner), url.PathEscape(repo))
	var allIssues []GiteaIssue
	page := 1
	limit := 50
	for {
		req, err := c.newRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("state", state)
		q.Set("type", "issues")
		q.Set("page", fmt.Sprintf("%d", page))
		q.Set("limit", fmt.Sprintf("%d", limit))
		req.URL.RawQuery = q.Encode()

		var issues []GiteaIssue
		if err := c.do(req, &issues); err != nil {
			return nil, err
		}
		if len(issues) == 0 {
			break
		}
		allIssues = append(allIssues, issues...)
		if len(issues) < limit {
			break
		}
		page++
	}
	return allIssues, nil
}

// ListPullRequests retrieves pull requests for a specific repository using pagination.
func (c *Client) ListPullRequests(ctx context.Context, owner, repo string, state string) ([]GiteaPullRequest, error) {
	path := fmt.Sprintf("/api/v1/repos/%s/%s/pulls", url.PathEscape(owner), url.PathEscape(repo))
	var allPulls []GiteaPullRequest
	page := 1
	limit := 50
	for {
		req, err := c.newRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("state", state)
		q.Set("page", fmt.Sprintf("%d", page))
		q.Set("limit", fmt.Sprintf("%d", limit))
		req.URL.RawQuery = q.Encode()

		var pulls []GiteaPullRequest
		if err := c.do(req, &pulls); err != nil {
			return nil, err
		}
		if len(pulls) == 0 {
			break
		}
		allPulls = append(allPulls, pulls...)
		if len(pulls) < limit {
			break
		}
		page++
	}
	return allPulls, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	u := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if c.AuthHeader != "" {
		req.Header.Set("Authorization", c.AuthHeader)
	} else if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}
	return req, nil
}

func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gitea api returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}
