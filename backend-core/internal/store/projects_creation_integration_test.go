package store_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

// #349 codex P2 hotfix — standalone(repository_id NULL) project key 중복 방지
// (migration 000039 partial unique index) + repo 동반 생성 atomicity (단일 tx).
// CI backend-unit 은 DEVHUB_TEST_DB_URL 미설정으로 setupApplicationsTest 가 t.Skip.

func standaloneProject(key, name string) domain.Project {
	return domain.Project{
		Key:         key,
		Name:        name,
		Status:      domain.ApplicationStatusActive,
		Visibility:  domain.ApplicationVisibilityInternal,
		OwnerUserID: "u1",
		// RepositoryID 0 → NULL (standalone)
	}
}

// migration 000039 — repository_id NULL row 들의 key 중복을 partial unique index 가 차단.
func TestIntegration_StandaloneProjectKey_PartialUnique(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	key := fmt.Sprintf("standalone-uniq-%d", time.Now().UnixNano())
	defer func() { _, _ = pool.Exec(ctx, `DELETE FROM projects WHERE key = $1`, key) }()

	if _, err := pgStore.CreateProjectWithRepositoryPayload(ctx, standaloneProject(key, "A"), nil, nil); err != nil {
		t.Fatalf("first standalone create: %v", err)
	}
	_, err := pgStore.CreateProjectWithRepositoryPayload(ctx, standaloneProject(key, "B"), nil, nil)
	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("duplicate standalone key should conflict (000039 partial unique), got %v", err)
	}
}

// atomicity — repoPayload 동반 생성 중 project insert 가 실패하면 (여기선 존재하지 않는
// application_id FK 위반) repository 도 rollback 되어 고아 row 가 남지 않아야 한다.
func TestIntegration_CreateProjectWithRepositoryPayload_RollbackOnProjectFailure(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	key := fmt.Sprintf("atomic-%d", time.Now().UnixNano())
	slug := fmt.Sprintf("org/atomic-repo-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM projects WHERE key = $1`, key)
		_, _ = pool.Exec(ctx, `DELETE FROM repositories WHERE full_name = $1`, slug)
	}()

	// 존재하지 않는 application_id → project insert FK 위반 → 전체 tx rollback.
	badAppID := "00000000-0000-0000-0000-000000000000"
	_, err := pgStore.CreateProjectWithRepositoryPayload(ctx, domain.Project{
		ApplicationID: badAppID,
		Key:           key,
		Name:          "Atomic",
		Status:        domain.ApplicationStatusActive,
		Visibility:    domain.ApplicationVisibilityInternal,
		OwnerUserID:   "u1",
	}, nil, &store.RepositoryCreatePayload{Key: "ATOMIC", Slug: slug, SCMProvider: "gitea"})
	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("project FK failure should map to ErrConflict, got %v", err)
	}

	var cnt int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM repositories WHERE full_name = $1`, slug).Scan(&cnt); err != nil {
		t.Fatalf("count repository: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("companion repository must be rolled back on project failure, found %d (codex #349 P2 atomicity)", cnt)
	}
}
