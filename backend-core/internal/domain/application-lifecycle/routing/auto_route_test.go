package routing

import (
	"context"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

type fakePlatformStore struct {
	platforms []domain.Platform
	err       error
}

func (f *fakePlatformStore) ListEnabledInboundSourcePlatforms(ctx context.Context) ([]domain.Platform, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.platforms, nil
}

func TestAutoRoute_ExternalRefPattern_GiteaOK(t *testing.T) {
	store := &fakePlatformStore{
		platforms: []domain.Platform{
			{ID: "plat-gitea-1", InboundSourceType: "gitea", DevelopmentUnitID: "dev-team"},
			{ID: "plat-jira-1", InboundSourceType: "jira", DevelopmentUnitID: "other-team"},
		},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:  "GITEA-123",
		SourceSystem: "gitea",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !decision.Matched {
		t.Fatal("expected matched=true")
	}
	if decision.PlatformID != "plat-gitea-1" {
		t.Fatalf("expected platform plat-gitea-1, got %s", decision.PlatformID)
	}
	if decision.Reason != "external_ref_pattern" {
		t.Fatalf("expected reason external_ref_pattern, got %s", decision.Reason)
	}
}

func TestAutoRoute_RequesterEmail_OK(t *testing.T) {
	store := &fakePlatformStore{
		platforms: []domain.Platform{
			{ID: "plat-eng-1", InboundSourceType: "gitea", DevelopmentUnitID: "alice@example.com"},
		},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:  "REF-001",
		SourceSystem: "manual",
		Requester:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !decision.Matched {
		t.Fatal("expected matched=true")
	}
	if decision.PlatformID != "plat-eng-1" {
		t.Fatalf("expected platform plat-eng-1, got %s", decision.PlatformID)
	}
	if decision.Reason != "requester_email" {
		t.Fatalf("expected reason requester_email, got %s", decision.Reason)
	}
}

func TestAutoRoute_ReqDepartment_OK(t *testing.T) {
	store := &fakePlatformStore{
		platforms: []domain.Platform{
			{ID: "plat-eng-1", InboundSourceType: "gitea", DevelopmentUnitID: "Engineering"},
			{ID: "plat-mkt-1", InboundSourceType: "jira", DevelopmentUnitID: "Marketing"},
		},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:   "REF-002",
		SourceSystem:  "manual",
		ReqDepartment: "Engineering",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !decision.Matched {
		t.Fatal("expected matched=true")
	}
	if decision.PlatformID != "plat-eng-1" {
		t.Fatalf("expected platform plat-eng-1, got %s", decision.PlatformID)
	}
	if decision.Reason != "req_department" {
		t.Fatalf("expected reason req_department, got %s", decision.Reason)
	}
}

func TestAutoRoute_NoMatch(t *testing.T) {
	store := &fakePlatformStore{
		platforms: []domain.Platform{
			{ID: "plat-1", InboundSourceType: "jira", DevelopmentUnitID: "jira-team"},
		},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:   "REF-NO-MATCH",
		SourceSystem:  "manual",
		Requester:     "nobody@example.com",
		ReqDepartment: "Nonexistent",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if decision.Matched {
		t.Fatal("expected matched=false")
	}
	if decision.Reason != "no_match" {
		t.Fatalf("expected reason no_match, got %s", decision.Reason)
	}
}

func TestAutoRoute_MultiplePlatformsFirstMatch(t *testing.T) {
	store := &fakePlatformStore{
		platforms: []domain.Platform{
			{ID: "plat-gitea-1", InboundSourceType: "gitea", DevelopmentUnitID: "dev-team"},
			{ID: "plat-gitea-2", InboundSourceType: "gitea", DevelopmentUnitID: "dev-team"},
		},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:  "GITEA-456",
		SourceSystem: "gitea",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !decision.Matched {
		t.Fatal("expected matched=true")
	}
	if decision.PlatformID != "plat-gitea-1" {
		t.Fatalf("expected first platform plat-gitea-1, got %s", decision.PlatformID)
	}
}

func TestAutoRoute_EmptyPlatforms(t *testing.T) {
	store := &fakePlatformStore{
		platforms: []domain.Platform{},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:  "GITEA-123",
		SourceSystem: "gitea",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if decision.Matched {
		t.Fatal("expected matched=false with empty platforms")
	}
	if decision.Reason != "no_match" {
		t.Fatalf("expected reason no_match, got %s", decision.Reason)
	}
}

// ============================================================================
// X-2 multi-provider depth 정밀화 unit test (release_v0-1_roadmap.md §3.5 X-2,
// sprint `feat/work_260614-x2-multi-provider-webhook`).
//
// auto_route.go 의 multi-provider depth:
//   - giteaExternalRefPattern (GITEA-<digits>)
//   - jiraExternalRefPattern (<PROJECT_KEY>-<digits>)
//   - githubExternalRefPattern (#<digits>)
//   - gitlabExternalRefPattern (!<digits>)
//   - 'other' provider 의 custom pattern (InboundSourceConfig.CustomExternalRefPattern)
// ============================================================================

func TestAutoRoute_ExternalRefPattern_JiraOK(t *testing.T) {
	store := &fakePlatformStore{
		platforms: []domain.Platform{
			{ID: "plat-jira-1", InboundSourceType: "jira"},
		},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:  "DEV-456",
		SourceSystem: "jira",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !decision.Matched {
		t.Fatal("expected matched=true for jira external_ref")
	}
	if decision.PlatformID != "plat-jira-1" {
		t.Fatalf("expected plat-jira-1, got %s", decision.PlatformID)
	}
	if decision.ProviderHint != "jira" {
		t.Fatalf("expected ProviderHint=jira, got %s", decision.ProviderHint)
	}
}

func TestAutoRoute_ExternalRefPattern_GithubOK(t *testing.T) {
	store := &fakePlatformStore{
		platforms: []domain.Platform{
			{ID: "plat-gh-1", InboundSourceType: "gitea"}, // github = gitea enum? — gitea/GitHub 모두 git forge → 'gitea' enum 재활용
		},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:  "#789",
		SourceSystem: "github",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// NOTE: gitea enum 만 지원 → github 의 `#789` 는 no_match (운영자 custom pattern 필요).
	if decision.Matched {
		t.Fatalf("expected matched=false for github (gitea enum only)")
	}
}

func TestAutoRoute_OtherProvider_CustomPatternOK(t *testing.T) {
	store := &fakePlatformStore{
		platforms: []domain.Platform{
			{
				ID:                "plat-other-1",
				InboundSourceType: "other",
				InboundSourceConfig: `{"custom_external_ref_pattern":"^CUSTOM-(?<id>\\d+)$"}`,
			},
		},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:  "CUSTOM-999",
		SourceSystem: "other",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !decision.Matched {
		t.Fatal("expected matched=true for other provider with custom pattern")
	}
	if decision.ProviderHint != "other_custom" {
		t.Fatalf("expected ProviderHint=other_custom, got %s", decision.ProviderHint)
	}
}

func TestAutoRoute_OtherProvider_InvalidCustomPattern(t *testing.T) {
	store := &fakePlatformStore{
		platforms: []domain.Platform{
			{
				ID:                "plat-other-bad",
				InboundSourceType: "other",
				InboundSourceConfig: `{"custom_external_ref_pattern":"[invalid("}`,
			},
		},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:  "ANY-REF",
		SourceSystem: "other",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// invalid pattern = silent skip (no match)
	if decision.Matched {
		t.Fatal("expected matched=false for invalid custom pattern")
	}
}

func TestParseInboundSourceRoutingConfig_EmptyAndValid(t *testing.T) {
	// empty → zero value, no error
	cfg, err := ParseInboundSourceRoutingConfig("")
	if err != nil {
		t.Fatalf("unexpected err for empty: %v", err)
	}
	if cfg.CustomExternalRefPattern != "" {
		t.Errorf("expected empty custom pattern, got %s", cfg.CustomExternalRefPattern)
	}
	// valid JSON
	cfg, err = ParseInboundSourceRoutingConfig(`{"custom_external_ref_pattern":"^X-\\d+$"}`)
	if err != nil {
		t.Fatalf("unexpected err for valid JSON: %v", err)
	}
	if cfg.CustomExternalRefPattern != `^X-\d+$` {
		t.Errorf("expected custom pattern ^X-\\d+$, got %s", cfg.CustomExternalRefPattern)
	}
	// invalid JSON → error
	_, err = ParseInboundSourceRoutingConfig(`{invalid json`)
	if err == nil {
		t.Fatal("expected err for invalid JSON")
	}
}

func TestAutoRoute_CrossProvider_NotMatched(t *testing.T) {
	// source_system = jira 이지만 platform 의 InboundSourceType = gitea → no match
	store := &fakePlatformStore{
		platforms: []domain.Platform{
			{ID: "plat-gitea-1", InboundSourceType: "gitea"},
		},
	}
	router := NewAutoRouter(store)
	decision, err := router.Route(context.Background(), VocRegistration{
		ExternalRef:  "DEV-456",
		SourceSystem: "jira",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if decision.Matched {
		t.Fatal("expected no_match: jira source vs gitea platform")
	}
}
