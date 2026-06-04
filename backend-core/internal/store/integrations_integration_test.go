package store_test

import (
	"errors"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

// Projects + Integrations integration test (sprint claude/work_260514-e).
// 핵심: P2 회귀 guard — UpdateIntegration 의 external_key 충돌이 ErrConflict 반환.

func TestIntegration_Project_UniqueRepositoryKey(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	app, err := pgStore.CreatePlatform(ctx, domain.Platform{
		Key: testAppKey1, Name: "X", Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	if err != nil {
		t.Fatalf("seed application: %v", err)
	}
	first, err := pgStore.CreateProject(ctx, domain.Project{
		PlatformID: app.ID, RepositoryID: testRepoID1,
		Key: "sprint-q3", Name: "Q3", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if first.ID == "" {
		t.Errorf("project ID should be generated")
	}
	// 동일 repository_id + key → conflict
	_, err = pgStore.CreateProject(ctx, domain.Project{
		PlatformID: app.ID, RepositoryID: testRepoID1,
		Key: "sprint-q3", Name: "duplicate", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	if !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict on duplicate (repo, key), got %v", err)
	}
	// 다른 repository_id 같은 key → OK
	_, err = pgStore.CreateProject(ctx, domain.Project{
		PlatformID: app.ID, RepositoryID: testRepoID2,
		Key: "sprint-q3", Name: "other repo", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	if err != nil {
		t.Errorf("different repository should not conflict: %v", err)
	}
}

func TestIntegration_ArchiveProject(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	p, err := pgStore.CreateProject(ctx, domain.Project{
		RepositoryID: testRepoID1, Key: "k1", Name: "X",
		Status: domain.PlatformStatusActive, Visibility: domain.PlatformVisibilityInternal,
		OwnerUserID: "u1",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	archived, err := pgStore.ArchiveProject(ctx, p.ID, "test reason")
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	if archived.Status != domain.PlatformStatusArchived {
		t.Errorf("status = %q, want archived", archived.Status)
	}
	if archived.ArchivedAt == nil {
		t.Errorf("archived_at should be set")
	}
}

func TestIntegration_CreateIntegration_ApplicationScope(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	app, err := pgStore.CreatePlatform(ctx, domain.Platform{
		Key: testAppKey1, Name: "X", Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	if err != nil {
		t.Fatalf("seed application: %v", err)
	}
	created, err := pgStore.CreateIntegration(ctx, domain.ProjectIntegration{
		Scope: domain.IntegrationScopePlatform, PlatformID: app.ID,
		IntegrationType: domain.IntegrationTypeJira,
		ExternalKey:     "PROJ-A", URL: "https://x", Policy: domain.IntegrationPolicySummaryOnly,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" {
		t.Errorf("ID should be generated")
	}
	if created.PlatformID != app.ID {
		t.Errorf("platform_id mismatch: %q vs %q", created.PlatformID, app.ID)
	}
}

// P2 codex review 회귀 guard — UpdateIntegration 의 external_key 변경이 다른 row 의
// (scope target, type, external_key) 와 충돌하면 ErrConflict 반환. hotfix #108 의
// store/integrations.go::UpdateIntegration 의 isUniqueViolation 매핑이 회귀하면
// 본 test 가 NotFound 또는 generic error 로 실패.
func TestIntegration_UpdateIntegration_UniqueConflict_P2(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	app, err := pgStore.CreatePlatform(ctx, domain.Platform{
		Key: testAppKey1, Name: "X", Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	if err != nil {
		t.Fatalf("seed application: %v", err)
	}
	// 2 integration: same (scope, application, type), 다른 external_key
	first, err := pgStore.CreateIntegration(ctx, domain.ProjectIntegration{
		Scope: domain.IntegrationScopePlatform, PlatformID: app.ID,
		IntegrationType: domain.IntegrationTypeJira,
		ExternalKey:     "PROJ-A", URL: "https://a", Policy: domain.IntegrationPolicySummaryOnly,
	})
	if err != nil {
		t.Fatalf("seed first integration: %v", err)
	}
	second, err := pgStore.CreateIntegration(ctx, domain.ProjectIntegration{
		Scope: domain.IntegrationScopePlatform, PlatformID: app.ID,
		IntegrationType: domain.IntegrationTypeJira,
		ExternalKey:     "PROJ-B", URL: "https://b", Policy: domain.IntegrationPolicySummaryOnly,
	})
	if err != nil {
		t.Fatalf("seed second integration: %v", err)
	}
	// second.external_key 를 first.external_key 로 변경 → unique 위반
	second.ExternalKey = "PROJ-A"
	_, err = pgStore.UpdateIntegration(ctx, second)
	if !errors.Is(err, store.ErrConflict) {
		t.Errorf("PR #108 P2 회귀: expected ErrConflict on duplicate external_key, got %v", err)
	}
	// first 는 변경 없이 그대로 유지
	if first.ExternalKey != "PROJ-A" {
		t.Errorf("first integration should still have external_key=PROJ-A")
	}
}

func TestIntegration_DeleteIntegration(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	app, err := pgStore.CreatePlatform(ctx, domain.Platform{
		Key: testAppKey1, Name: "X", Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	if err != nil {
		t.Fatalf("seed application: %v", err)
	}
	created, err := pgStore.CreateIntegration(ctx, domain.ProjectIntegration{
		Scope: domain.IntegrationScopePlatform, PlatformID: app.ID,
		IntegrationType: domain.IntegrationTypeConfluence,
		ExternalKey:     "WIKI-A", URL: "https://x", Policy: domain.IntegrationPolicySummaryOnly,
	})
	if err != nil {
		t.Fatalf("seed integration: %v", err)
	}
	if err := pgStore.DeleteIntegration(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	// 재 삭제 → ErrNotFound
	if err := pgStore.DeleteIntegration(ctx, created.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("second delete should return ErrNotFound, got %v", err)
	}
}

func TestIntegration_ListIntegrations_ScopeFilter(t *testing.T) {
	pgStore, _, ctx, teardown := setupApplicationsTest(t)
	defer teardown()
	app, err := pgStore.CreatePlatform(ctx, domain.Platform{
		Key: testAppKey1, Name: "X", Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	if err != nil {
		t.Fatalf("seed application: %v", err)
	}
	project, err := pgStore.CreateProject(ctx, domain.Project{
		PlatformID: app.ID, RepositoryID: testRepoID1, Key: "p1", Name: "P",
		Status: domain.PlatformStatusActive, Visibility: domain.PlatformVisibilityInternal,
		OwnerUserID: "u1",
	})
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}
	// application scope 1
	if _, err := pgStore.CreateIntegration(ctx, domain.ProjectIntegration{
		Scope: domain.IntegrationScopePlatform, PlatformID: app.ID,
		IntegrationType: domain.IntegrationTypeJira,
		ExternalKey:     "APP-PROJ", URL: "https://a", Policy: domain.IntegrationPolicySummaryOnly,
	}); err != nil {
		t.Fatalf("seed app integration: %v", err)
	}
	// project scope 1
	if _, err := pgStore.CreateIntegration(ctx, domain.ProjectIntegration{
		Scope: domain.IntegrationScopeProject, ProjectID: project.ID,
		IntegrationType: domain.IntegrationTypeJira,
		ExternalKey:     "PROJ-X", URL: "https://b", Policy: domain.IntegrationPolicyExecutionSystem,
	}); err != nil {
		t.Fatalf("seed project integration: %v", err)
	}
	// scope=platform 필터 → 1건
	_, total, _ := pgStore.ListIntegrations(ctx, store.IntegrationListOptions{
		Scope: domain.IntegrationScopePlatform,
	})
	if total != 1 {
		t.Errorf("application scope total = %d, want 1", total)
	}
	// scope=project 필터 → 1건
	_, total, _ = pgStore.ListIntegrations(ctx, store.IntegrationListOptions{
		Scope: domain.IntegrationScopeProject,
	})
	if total != 1 {
		t.Errorf("project scope total = %d, want 1", total)
	}
}
