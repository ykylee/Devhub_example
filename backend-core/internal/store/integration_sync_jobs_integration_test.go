package store_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/store"
)

// integration_sync_jobs scm-gate integration test (codex review PR #341 P1).
//
// CI backend-unit job 은 DEVHUB_TEST_DB_URL 미설정으로 t.Skip.
// AcquireNextQueuedSyncJob 이 provider_type='scm' 작업만 claim 하고 비-SCM
// (alm/Jira 등) 작업은 건드리지 않음을 검증 — Gitea 워커의 false-completion 회귀 가드.

func TestIntegration_AcquireNextQueuedSyncJob_OnlyClaimsSCM(t *testing.T) {
	pgStore, pool, ctx, teardown := setupApplicationsTest(t)
	defer teardown()

	suffix := time.Now().UnixNano()
	scmKey := fmt.Sprintf("gitea-gate-%d", suffix)
	almKey := fmt.Sprintf("jira-gate-%d", suffix)

	var scmProviderID, almProviderID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO integration_providers (provider_key, provider_type, display_name, auth_mode, credentials_ref)
		 VALUES ($1, 'scm', 'Gitea Gate Test', 'token', 'ref') RETURNING provider_id::text`,
		scmKey).Scan(&scmProviderID); err != nil {
		t.Fatalf("seed scm provider: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO integration_providers (provider_key, provider_type, display_name, auth_mode, credentials_ref)
		 VALUES ($1, 'alm', 'Jira Gate Test', 'token', 'ref') RETURNING provider_id::text`,
		almKey).Scan(&almProviderID); err != nil {
		t.Fatalf("seed alm provider: %v", err)
	}
	// FK ON DELETE CASCADE — provider 삭제가 sync job 까지 정리.
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM integration_providers WHERE provider_id = ANY($1)`,
			[]string{scmProviderID, almProviderID})
	}()

	// ALM job 을 더 오래된 created_at 으로 큐잉 — 게이트가 없으면 ORDER BY created_at ASC
	// 가 이 ALM job 을 먼저 집어 false-complete 한다.
	var almJobID, scmJobID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO integration_sync_jobs (provider_id, status, created_at)
		 VALUES ($1, 'queued', NOW() - interval '10 seconds') RETURNING job_id::text`,
		almProviderID).Scan(&almJobID); err != nil {
		t.Fatalf("queue alm job: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO integration_sync_jobs (provider_id, status) VALUES ($1, 'queued') RETURNING job_id::text`,
		scmProviderID).Scan(&scmJobID); err != nil {
		t.Fatalf("queue scm job: %v", err)
	}

	gotJobID, gotProviderID, err := pgStore.AcquireNextQueuedSyncJob(ctx)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if gotJobID != scmJobID {
		t.Fatalf("expected scm job %s claimed, got %s (older alm job must be skipped)", scmJobID, gotJobID)
	}
	if gotProviderID != scmProviderID {
		t.Fatalf("expected scm provider %s, got %s", scmProviderID, gotProviderID)
	}

	// ALM job 은 여전히 queued (claim 되지 않음).
	var almStatus string
	if err := pool.QueryRow(ctx, `SELECT status FROM integration_sync_jobs WHERE job_id = $1`, almJobID).Scan(&almStatus); err != nil {
		t.Fatalf("read alm job status: %v", err)
	}
	if almStatus != "queued" {
		t.Fatalf("alm job must remain queued (not stolen by SCM worker), got %q", almStatus)
	}

	// SCM job 이 모두 소진되면 더 이상 claim 할 SCM 작업 없음 (ALM 은 여전히 무시).
	if _, _, err := pgStore.AcquireNextQueuedSyncJob(ctx); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when no queued SCM jobs remain, got %v", err)
	}
}
