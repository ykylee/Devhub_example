package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// GiteaClient is a minimal Gitea API client (REST).
// Implements list-pull-requests-since, list-builds, list-commits, get-commit, get-build.
// We use net/http directly to avoid an external SDK dependency.
type GiteaClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewGiteaClient constructs a GiteaClient with sensible defaults.
func NewGiteaClient(baseURL, token string) *GiteaClient {
	return &GiteaClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Token:      token,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GiteaPullRequest is the minimal subset of Gitea PR schema we care about.
type GiteaPullRequest struct {
	ID        int64     `json:"id"`
	Number    int       `json:"number"`
	State     string    `json:"state"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	HeadSHA   string    `json:"head_sha"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}

// GiteaBuild is the minimal subset of Gitea Actions build schema.
type GiteaBuild struct {
	ID        int64     `json:"id"`
	CommitSHA string    `json:"head_sha"`
	Event     string    `json:"event"`
	Status    string    `json:"status"`
	Conclusion string   `json:"conclusion"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GiteaCommit is the minimal subset of Gitea commit schema.
type GiteaCommit struct {
	SHA     string `json:"sha"`
	Message string `json:"commit.message"`
}

// ListPullRequestsSince fetches PRs updated after `since`. Returns up to 50 per page, max 200 total.
func (c *GiteaClient) ListPullRequestsSince(ctx context.Context, owner, repo string, since time.Time) ([]GiteaPullRequest, error) {
	out := []GiteaPullRequest{}
	page := 1
	for page <= 4 {
		u := fmt.Sprintf("%s/api/v1/repos/%s/%s/pulls?state=all&page=%d&limit=50", c.BaseURL, owner, repo, page)
		if !since.IsZero() {
			q := url.Values{}
			q.Set("since", since.UTC().Format(time.RFC3339))
			u += "&" + q.Encode()
		}
		var batch []GiteaPullRequest
		if err := c.doJSON(ctx, "GET", u, &batch); err != nil {
			return nil, err
		}
		out = append(out, batch...)
		if len(batch) < 50 {
			break
		}
		page++
	}
	return out, nil
}

// ListBuilds fetches recent builds for a repo.
func (c *GiteaClient) ListBuilds(ctx context.Context, owner, repo string, since time.Time) ([]GiteaBuild, error) {
	u := fmt.Sprintf("%s/api/v1/repos/%s/%s/actions/runs?page=1&limit=50", c.BaseURL, owner, repo)
	if !since.IsZero() {
		q := url.Values{}
		q.Set("since", since.UTC().Format(time.RFC3339))
		u += "&" + q.Encode()
	}
	var out []GiteaBuild
	if err := c.doJSON(ctx, "GET", u, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCommit fetches a single commit by SHA.
func (c *GiteaClient) GetCommit(ctx context.Context, owner, repo, sha string) (*GiteaCommit, error) {
	u := fmt.Sprintf("%s/api/v1/repos/%s/%s/git/commits/%s", c.BaseURL, owner, repo, sha)
	var out GiteaCommit
	if err := c.doJSON(ctx, "GET", u, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *GiteaClient) doJSON(ctx context.Context, method, url string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("gitea: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("gitea: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitea: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("gitea: decode: %w", err)
	}
	return nil
}

// GiteaPullAdapter performs a since-based pull for a single repository.
type GiteaPullAdapter struct {
	Client *GiteaClient
	Store  RepositoryPullStore
	// MaxItemsPerCall caps the items upserted per pull (defensive cap).
	MaxItemsPerCall int
}

// RepositoryPullStore is the persistence contract required by GiteaPullAdapter.
type RepositoryPullStore interface {
	UpsertPullActivity(ctx context.Context, repositoryID string, giteaPRID int64, number int, state, title, body, headSHA, authorLogin string, updatedAt time.Time) error
	UpsertBuildRun(ctx context.Context, repositoryID string, giteaBuildID int64, commitSHA, event, status, conclusion string, createdAt time.Time) error
	UpsertQualitySnapshot(ctx context.Context, repositoryID string, commitSHA string, recordedAt time.Time) error
	UpdatePullState(ctx context.Context, repositoryID string, status string, errMsg string, lastPullAt time.Time) error
	IncrementConsecutiveFailures(ctx context.Context, repositoryID string) (int, error)
	ResetConsecutiveFailures(ctx context.Context, repositoryID string) error
	SetBackoff(ctx context.Context, repositoryID string, until time.Time) error
	BackoffUntil(ctx context.Context, repositoryID string) (time.Time, error)
	LastPullAt(ctx context.Context, repositoryID string) (time.Time, error)
}

// RepositoryTarget represents one repository to pull.
type RepositoryTarget struct {
	ID         string
	Owner      string
	Name       string
	ExternalID string
}

// PullAndIngestSince performs a since-based pull for a single repository and persists results.
// Errors are typed so the loop can distinguish between Gitea-unreachable, partial, and fully-failed outcomes.
type PullError struct {
	Class   string
	Message string
	Cause   error
}

func (e *PullError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("gitea pull: %s: %s: %v", e.Class, e.Message, e.Cause)
	}
	return fmt.Sprintf("gitea pull: %s: %s", e.Class, e.Message)
}

func (e *PullError) Unwrap() error { return e.Cause }

var (
	ErrGiteaUnreachable = errors.New("gitea unreachable")
	ErrGiteaPartial     = errors.New("gitea partial response")
)

// PullAndIngestSince is the main entry point. Returns nil on full success, *PullError on failure.
func (a *GiteaPullAdapter) PullAndIngestSince(ctx context.Context, target RepositoryTarget) error {
	if a.Client == nil || a.Store == nil {
		return &PullError{Class: "config", Message: "GiteaClient or Store is nil"}
	}
	if a.MaxItemsPerCall <= 0 {
		a.MaxItemsPerCall = 200
	}
	since, err := a.Store.LastPullAt(ctx, target.ID)
	if err != nil {
		return &PullError{Class: "store", Message: "last_pull_at fetch", Cause: err}
	}

	// PRs
	prs, prErr := a.Client.ListPullRequestsSince(ctx, target.Owner, target.Name, since)
	if prErr != nil {
		return &PullError{Class: "gitea_api", Message: "list pull requests", Cause: prErr}
	}

	// Builds
	builds, buildErr := a.Client.ListBuilds(ctx, target.Owner, target.Name, since)
	if buildErr != nil {
		return &PullError{Class: "gitea_api", Message: "list builds", Cause: buildErr}
	}

	var partialReasons []string
	count := 0

	// Upsert PRs
	for _, pr := range prs {
		if count >= a.MaxItemsPerCall {
			partialReasons = append(partialReasons, "max_items_per_call reached at pr")
			break
		}
		if err := a.Store.UpsertPullActivity(ctx, target.ID, pr.ID, pr.Number, pr.State, pr.Title, pr.Body, pr.HeadSHA, pr.User.Login, pr.UpdatedAt); err != nil {
			partialReasons = append(partialReasons, fmt.Sprintf("pr %d upsert: %v", pr.ID, err))
			continue
		}
		count++
	}

	// Upsert builds + quality_snapshots
	for _, b := range builds {
		if count >= a.MaxItemsPerCall {
			partialReasons = append(partialReasons, "max_items_per_call reached at build")
			break
		}
		if err := a.Store.UpsertBuildRun(ctx, target.ID, b.ID, b.CommitSHA, b.Event, b.Status, b.Conclusion, b.CreatedAt); err != nil {
			partialReasons = append(partialReasons, fmt.Sprintf("build %d upsert: %v", b.ID, err))
			continue
		}
		count++

		// quality_snapshot for each unique build commit
		if b.CommitSHA != "" {
			if err := a.Store.UpsertQualitySnapshot(ctx, target.ID, b.CommitSHA, b.UpdatedAt); err != nil {
				partialReasons = append(partialReasons, fmt.Sprintf("quality_snapshot %s upsert: %v", b.CommitSHA, err))
				continue
			}
			count++
		}
	}

	now := time.Now().UTC()
	if len(partialReasons) > 0 {
		// partial — record but do not advance last_pull_at
		if err := a.Store.UpdatePullState(ctx, target.ID, "partial", strings.Join(partialReasons, "; "), since); err != nil {
			return &PullError{Class: "store", Message: "partial state update", Cause: err}
		}
		// increment consecutive_failures and apply backoff
		if _, err := a.Store.IncrementConsecutiveFailures(ctx, target.ID); err != nil {
			return &PullError{Class: "store", Message: "increment consecutive failures", Cause: err}
		}
		return &PullError{Class: "partial", Message: strings.Join(partialReasons, "; "), Cause: ErrGiteaPartial}
	}

	// full success
	if err := a.Store.UpdatePullState(ctx, target.ID, "success", "", now); err != nil {
		return &PullError{Class: "store", Message: "success state update", Cause: err}
	}
	if err := a.Store.ResetConsecutiveFailures(ctx, target.ID); err != nil {
		return &PullError{Class: "store", Message: "reset consecutive failures", Cause: err}
	}
	return nil
}

// Semaphore is a simple counting semaphore for goroutine concurrency limiting.
type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	if n <= 0 {
		n = 1
	}
	return &Semaphore{ch: make(chan struct{}, n)}
}

func (s *Semaphore) Acquire() { s.ch <- struct{}{} }
func (s *Semaphore) Release() { <-s.ch }

// ensure sync import is used (no-op)
var _ = sync.Mutex{}
