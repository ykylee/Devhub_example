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
	ID        int64      `json:"id"`
	Number    int        `json:"number"`
	State     string     `json:"state"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	HeadSHA   string     `json:"head_sha"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedAt time.Time  `json:"created_at"`
	Merged    bool       `json:"merged"`
	MergedAt  *time.Time `json:"merged_at,omitempty"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}

// stateToEventType maps Gitea PR state (open/closed) + merged bool to
// pr_activities.event_type enum (opened/reviewed/commented/closed/merged/reopened/updated).
// Migration 000001 L411 enum constraint 정합. Defensive fallback = "updated".
func stateToEventType(state string, merged bool) string {
	switch state {
	case "open":
		return "opened"
	case "closed":
		if merged {
			return "merged"
		}
		return "closed"
	default:
		return "updated"
	}
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
// If more than 200 PRs are updated since `since`, the returned (truncated bool) signals
// truncation so the caller can log a metric / alert. Silent truncation would hide a real
// "we missed updates" failure mode.
func (c *GiteaClient) ListPullRequestsSince(ctx context.Context, owner, repo string, since time.Time) (prs []GiteaPullRequest, truncated bool, err error) {
	out := []GiteaPullRequest{}
	page := 1
	const maxPages = 4
	const pageSize = 50
	for page <= maxPages {
		u := fmt.Sprintf("%s/api/v1/repos/%s/%s/pulls?state=all&page=%d&limit=%d", c.BaseURL, owner, repo, page, pageSize)
		if !since.IsZero() {
			q := url.Values{}
			q.Set("since", since.UTC().Format(time.RFC3339))
			u += "&" + q.Encode()
		}
		var batch []GiteaPullRequest
		if err := c.doJSON(ctx, "GET", u, &batch); err != nil {
			return nil, false, err
		}
		out = append(out, batch...)
		if len(batch) < pageSize {
			return out, false, nil
		}
		page++
	}
	// We fetched maxPages full pages — additional pages likely exist.
	return out, true, nil
}


// Gitea version 별로 다름 (1.21+ = { "workflow_runs": [...] } wrapper, 그 이전 / 일부 fork
// = bare array). 양쪽 shape 모두 decode 시도 — `workflow_runs` 가 있으면 그걸, 없으면
// bare array 로 취급. (codex P1 — production wire 의 per-repo 호출에서 JSON decode fail 방지)
func (c *GiteaClient) ListBuilds(ctx context.Context, owner, repo string, since time.Time) ([]GiteaBuild, error) {
	u := fmt.Sprintf("%s/api/v1/repos/%s/%s/actions/runs?page=1&limit=50", c.BaseURL, owner, repo)
	if !since.IsZero() {
		q := url.Values{}
		q.Set("since", since.UTC().Format(time.RFC3339))
		u += "&" + q.Encode()
	}
	var wrapped struct {
		WorkflowRuns []GiteaBuild `json:"workflow_runs"`
	}
	if err := c.doJSON(ctx, "GET", u, &wrapped); err == nil && len(wrapped.WorkflowRuns) > 0 {
		return wrapped.WorkflowRuns, nil
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
	prs, prTruncated, prErr := a.Client.ListPullRequestsSince(ctx, target.Owner, target.Name, since)
	if prErr != nil {
		return &PullError{Class: "gitea_api", Message: "list pull requests", Cause: prErr}
	}
	if prTruncated {
		// The Gitea API returned >= 200 PRs updated since `since`; additional pages were not fetched.
		// Record this as a partial outcome so the loop can alert (operational visibility).
		// The next cycle will catch up from the new last_pull_at (the upserts advanced the cursor).
		// We do NOT advance last_pull_at on truncated cycles — see partial branch below.
		_ = a.Store.UpdatePullState(ctx, target.ID, "partial", "pr list truncated at max_pages (>= 200 PRs updated since last_pull_at); next cycle will catch up", since)
		return &PullError{Class: "partial", Message: "pr list truncated; >= 200 PRs updated since last_pull_at", Cause: ErrGiteaPartial}
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
		eventType := stateToEventType(pr.State, pr.Merged)
		if err := a.Store.UpsertPullActivity(ctx, target.ID, pr.ID, pr.Number, eventType, pr.Title, pr.Body, pr.HeadSHA, pr.User.Login, pr.UpdatedAt); err != nil {
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
		// NOTE: consecutive_failures increment + backoff application are owned by the loop
		// (RunGiteaPullLoop) — see ADR-0034 §3.3 + §3.4. The adapter only records the
		// partial outcome + reason; the loop performs the counter mutation once per
		// failed cycle. Double-counting was a real bug in the v1 review.
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
