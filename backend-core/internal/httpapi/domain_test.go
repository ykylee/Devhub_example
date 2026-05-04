package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

type memoryDomainStore struct {
	repositories []domain.Repository
	issues       []domain.Issue
	pullRequests []domain.PullRequest
	ciRuns       []domain.CIRun
	risks        []domain.Risk
}

func (s *memoryDomainStore) ListRepositories(_ context.Context, _ domain.ListOptions) ([]domain.Repository, error) {
	return s.repositories, nil
}

func (s *memoryDomainStore) ListIssues(_ context.Context, opts domain.ListOptions) ([]domain.Issue, error) {
	if opts.State == "" && opts.RepositoryName == "" {
		return s.issues, nil
	}
	filtered := make([]domain.Issue, 0, len(s.issues))
	for _, issue := range s.issues {
		if opts.State != "" && issue.State != opts.State {
			continue
		}
		if opts.RepositoryName != "" && issue.RepositoryName != opts.RepositoryName {
			continue
		}
		filtered = append(filtered, issue)
	}
	return filtered, nil
}

func (s *memoryDomainStore) ListPullRequests(_ context.Context, opts domain.ListOptions) ([]domain.PullRequest, error) {
	if opts.State == "" && opts.RepositoryName == "" {
		return s.pullRequests, nil
	}
	filtered := make([]domain.PullRequest, 0, len(s.pullRequests))
	for _, pullRequest := range s.pullRequests {
		if opts.State != "" && pullRequest.State != opts.State {
			continue
		}
		if opts.RepositoryName != "" && pullRequest.RepositoryName != opts.RepositoryName {
			continue
		}
		filtered = append(filtered, pullRequest)
	}
	return filtered, nil
}

func (s *memoryDomainStore) ListCIRuns(_ context.Context, _ domain.ListOptions) ([]domain.CIRun, error) {
	return s.ciRuns, nil
}

func (s *memoryDomainStore) ListRisks(_ context.Context, opts domain.ListOptions) ([]domain.Risk, error) {
	if opts.Status == "" && opts.Impact == "" {
		return s.risks, nil
	}
	filtered := make([]domain.Risk, 0, len(s.risks))
	for _, risk := range s.risks {
		if opts.Status != "" && risk.Status != opts.Status {
			continue
		}
		if opts.Impact != "" && risk.Impact != opts.Impact {
			continue
		}
		filtered = append(filtered, risk)
	}
	return filtered, nil
}

func TestRepositoriesReturnsDBBackedEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	router := NewRouter(RouterConfig{DomainStore: &memoryDomainStore{
		repositories: []domain.Repository{
			{ID: 1, GiteaID: 42, FullName: "acme/api", OwnerLogin: "acme", Name: "api", DefaultBranch: "main", UpdatedAt: now},
		},
	}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/repositories?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var response struct {
		Status string `json:"status"`
		Data   []struct {
			FullName      string `json:"full_name"`
			DefaultBranch string `json:"default_branch"`
		} `json:"data"`
		Meta struct {
			Limit int `json:"limit"`
			Count int `json:"count"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Status != "ok" || response.Meta.Limit != 10 || response.Meta.Count != 1 {
		t.Fatalf("unexpected response meta: %+v", response)
	}
	if response.Data[0].FullName != "acme/api" || response.Data[0].DefaultBranch != "main" {
		t.Fatalf("unexpected repository: %+v", response.Data[0])
	}
}

func TestIssuesSupportsStateAndRepositoryFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	router := NewRouter(RouterConfig{DomainStore: &memoryDomainStore{
		issues: []domain.Issue{
			{ID: 1, RepositoryName: "acme/api", Number: 17, Title: "Open issue", State: "open", UpdatedAt: now},
			{ID: 2, RepositoryName: "acme/web", Number: 18, Title: "Closed issue", State: "closed", UpdatedAt: now},
		},
	}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?repository_name=acme/api&state=open", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var response struct {
		Data []struct {
			RepositoryName string `json:"repository_name"`
			State          string `json:"state"`
		} `json:"data"`
		Meta struct {
			RepositoryName string `json:"repository_name"`
			State          string `json:"state"`
			Count          int    `json:"count"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Meta.RepositoryName != "acme/api" || response.Meta.State != "open" || response.Meta.Count != 1 {
		t.Fatalf("unexpected meta: %+v", response.Meta)
	}
	if response.Data[0].RepositoryName != "acme/api" || response.Data[0].State != "open" {
		t.Fatalf("unexpected issue: %+v", response.Data[0])
	}
}

func TestPullRequestsReturnsBranchFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	router := NewRouter(RouterConfig{DomainStore: &memoryDomainStore{
		pullRequests: []domain.PullRequest{
			{ID: 1, RepositoryName: "acme/api", Number: 23, Title: "PR", State: "open", HeadBranch: "feature/api", BaseBranch: "main", HeadSHA: "abc1234", UpdatedAt: now},
		},
	}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/pull-requests", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var response struct {
		Data []struct {
			HeadBranch string `json:"head_branch"`
			BaseBranch string `json:"base_branch"`
			HeadSHA    string `json:"head_sha"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Data) != 1 || response.Data[0].HeadBranch != "feature/api" || response.Data[0].BaseBranch != "main" {
		t.Fatalf("unexpected pull request response: %+v", response.Data)
	}
	if response.Data[0].HeadSHA != "abc1234" {
		t.Fatalf("expected head sha, got %+v", response.Data[0])
	}
}

func TestCIRunsUsesDBWhenAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	startedAt := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	duration := 87
	router := NewRouter(RouterConfig{DomainStore: &memoryDomainStore{
		ciRuns: []domain.CIRun{
			{ID: 1, ExternalID: "run-501", RepositoryName: "acme/api", Branch: "main", CommitSHA: "abc1234", Status: "success", StartedAt: &startedAt, DurationSeconds: &duration},
		},
	}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ci-runs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var response struct {
		Data []struct {
			ID       string `json:"id"`
			Duration int    `json:"duration_seconds"`
		} `json:"data"`
		Meta struct {
			Source string `json:"source"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Meta.Source != "db" || len(response.Data) != 1 || response.Data[0].ID != "run-501" || response.Data[0].Duration != 87 {
		t.Fatalf("unexpected ci response: %+v", response)
	}
}

func TestRisksReturnsDBBackedEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	router := NewRouter(RouterConfig{DomainStore: &memoryDomainStore{
		risks: []domain.Risk{
			{
				RiskKey:          "ci_failure:502",
				Title:            "CI run failed for acme/api",
				Reason:           "CI run 502 failed on branch main",
				Impact:           "high",
				Status:           "action_required",
				SourceType:       "ci_run",
				SourceID:         "502",
				SuggestedActions: []string{"inspect_logs", "rerun_ci"},
				CreatedAt:        now,
				UpdatedAt:        now,
			},
			{
				RiskKey: "ci_failure:503",
				Title:   "CI run failed for acme/web",
				Impact:  "medium",
				Status:  "investigation",
			},
		},
	}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/risks?status=action_required&impact=high", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var response struct {
		Data []struct {
			ID               string   `json:"id"`
			Impact           string   `json:"impact"`
			Status           string   `json:"status"`
			SuggestedActions []string `json:"suggested_actions"`
		} `json:"data"`
		Meta struct {
			Status string `json:"status"`
			Impact string `json:"impact"`
			Count  int    `json:"count"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Meta.Status != "action_required" || response.Meta.Impact != "high" || response.Meta.Count != 1 {
		t.Fatalf("unexpected meta: %+v", response.Meta)
	}
	if response.Data[0].ID != "ci_failure:502" || response.Data[0].Impact != "high" || response.Data[0].Status != "action_required" {
		t.Fatalf("unexpected risk response: %+v", response.Data)
	}
	if len(response.Data[0].SuggestedActions) != 2 || response.Data[0].SuggestedActions[0] != "inspect_logs" {
		t.Fatalf("unexpected suggested actions: %+v", response.Data[0].SuggestedActions)
	}
}
