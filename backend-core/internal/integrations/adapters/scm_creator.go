package adapters

import (
	"context"
	"errors"
	"sync"
	"time"
)

// SCMCreateStatus is the 4-state machine for repository SCM create.
type SCMCreateStatus string

const (
	SCMCreatePending  SCMCreateStatus = "pending"
	SCMCreateSuccess  SCMCreateStatus = "success"
	SCMCreateFailed   SCMCreateStatus = "failed"
	SCMCreateRetryScd SCMCreateStatus = "retry_scheduled"
)

// SCMCreateRequest is the input to SCMCreator.CreateRepository.
type SCMCreateRequest struct {
	RepositoryID   string // DevHub DB row ID
	ProjectID      string // optional, for audit
	SCMProvider    string // gitea | github | gitlab | (future)
	GiteaOrg       string // empty = user repo
	RepoName       string // Gitea repo name
	Description    string
	Private        bool
	AutoInit       bool
}

// SCMCreateResult is the post-call state for the response envelope.
type SCMCreateResult struct {
	RepositoryID    string
	Status          SCMCreateStatus
	HTTPStatus      int    // 0 if no API call attempted
	DurationMs      int
	ExternalID      int64  // Gitea repo ID
	CloneURL        string
	HTMLURL         string
	ErrorClass      string
	ErrorMessage    string
	CompensationAction string // none | retry_scheduled
	NextRetryAt        *time.Time
}

// SCMCreateStore is the persistence contract for SCMCreate state updates.
type SCMCreateStore interface {
	UpdateSCMCreateState(ctx context.Context, repositoryID string, status SCMCreateStatus, errMsg string, externalID int64, cloneURL, htmlURL string, at time.Time) error
}

// SCMCreator performs post-commit Gitea API calls to create a repo, then updates
// the store with the outcome. Best-effort: any error from the API is recorded,
// not propagated to the project creation flow (which already committed).
type SCMCreator struct {
	Client    *GiteaClient
	Store     SCMCreateStore
	Timeout   time.Duration
	OnSuccess func(ctx context.Context, result SCMCreateResult) // metric + audit hook
	OnError   func(ctx context.Context, result SCMCreateResult)
	OnCompensation func(ctx context.Context, result SCMCreateResult)

	once     sync.Once
}

// CreateRepository performs a single SCM create call. Never returns a hard error
// to the caller; instead, fills SCMCreateResult.Status + ErrorClass/Message.
// The caller (project creation handler) returns a 200 response based on the DB state.
func (c *SCMCreator) CreateRepository(ctx context.Context, req SCMCreateRequest) SCMCreateResult {
	start := time.Now().UTC()
	result := SCMCreateResult{
		RepositoryID: req.RepositoryID,
		Status:       SCMCreatePending,
	}

	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Mark pending before the call (in case crash recovery / crash visibility)
	if c.Store != nil {
		_ = c.Store.UpdateSCMCreateState(callCtx, req.RepositoryID, SCMCreatePending, "", 0, "", "", start)
	}

	if c.Client == nil {
		result.Status = SCMCreateFailed
		result.ErrorClass = "config"
		result.ErrorMessage = "GiteaClient is nil"
		c.recordOutcome(ctx, start, &result, false, false)
		return result
	}

	options := GiteaRepoOptions{
		Description: req.Description,
		Private:     req.Private,
		AutoInit:    req.AutoInit,
	}

	var (
		repo *GiteaRepo
		err  error
	)
	if req.GiteaOrg == "" {
		repo, err = c.Client.CreateUserRepo(callCtx, req.RepoName, options)
	} else {
		repo, err = c.Client.CreateOrgRepo(callCtx, req.GiteaOrg, req.RepoName, options)
	}
	result.DurationMs = int(time.Since(start).Milliseconds())

	if err == nil {
		result.Status = SCMCreateSuccess
		result.ExternalID = repo.ID
		result.CloneURL = repo.CloneURL
		result.HTMLURL = repo.HTMLURL
		if c.Store != nil {
			_ = c.Store.UpdateSCMCreateState(callCtx, req.RepositoryID, SCMCreateSuccess, "", repo.ID, repo.CloneURL, repo.HTMLURL, time.Now().UTC())
		}
		c.recordOutcome(ctx, start, &result, true, false)
		return result
	}

	// Failure path
	result.Status = SCMCreateFailed
	var ge *GiteaAPIError
	if errors.As(err, &ge) {
		result.HTTPStatus = ge.HTTPStatus
		result.ErrorClass = ge.Class
		result.ErrorMessage = ge.Message
	} else {
		result.ErrorClass = "unknown"
		result.ErrorMessage = err.Error()
	}

	if c.Store != nil {
		// Use parent ctx (not callCtx) for the failure-state write. callCtx may already
		// be expired due to the Gitea timeout that triggered this failure path, and
		// a stale-context write would silently drop the row. codex review #2 (P2).
		_ = c.Store.UpdateSCMCreateState(ctx, req.RepositoryID, SCMCreateFailed,
			result.ErrorClass+": "+result.ErrorMessage, 0, "", "", time.Now().UTC())
	}

	// Best-effort compensation: schedule retry 24h later (cap), no automatic retry.
	nextRetry := time.Now().UTC().Add(24 * time.Hour)
	result.CompensationAction = "retry_scheduled"
	result.NextRetryAt = &nextRetry
	c.recordOutcome(ctx, start, &result, false, true)
	return result
}

func (c *SCMCreator) recordOutcome(ctx context.Context, start time.Time, result *SCMCreateResult, success bool, compensation bool) {
	// success path: emit metric unconditionally; emit hook only if wired.
	// codex review #3 (P2): callers without OnSuccess were falling through to
	// observeSCMError and double-counting success as failure.
	if success {
		observeSCMSuccess(time.Since(start))
		if c.OnSuccess != nil {
			c.OnSuccess(ctx, *result)
		}
		return
	}
	// failure path: emit hook if wired; emit metric unconditionally.
	if c.OnError != nil {
		c.OnError(ctx, *result)
	}
	observeSCMError(result.ErrorClass, time.Since(start))
	if compensation && c.OnCompensation != nil {
		c.OnCompensation(ctx, *result)
	}
	observeSCMError(result.ErrorClass, time.Since(start))
}
