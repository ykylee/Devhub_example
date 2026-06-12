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
