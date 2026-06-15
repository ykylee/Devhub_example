package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// RepositoryPullStore 의 ingest 3 method (X-5 production wire follow-up).
//
// 책임: pr_activities / build_runs / quality_snapshots 테이블에 Gitea API 결과를
// idempotent 하게 upsert. 어댑터 (adapters.GiteaPullAdapter) 가 결정한 enum 값
// (event_type, status 등) 을 그대로 받아 column 매핑만 책임진다. 본 follow-up
// (X-5 store wire) 의 IMPL-GITEA-PULL-STORE-01 추적성 ID.
//
// 어댑터 책임 (gitea_pull.go: stateToEventType):
//   - state (open/closed) + merged bool → event_type enum (opened/closed/merged/updated)
//   - b.Event → build_runs.branch (default_branch fallback 은 store 가 default_branch
//     를 모름 → 어댑터가 repositories.default_branch 조회 후 전달)
//
// repositoryID 는 string (RepositoryPullStore interface signature) 으로 들어오나
// repositories.id / pr_activities.repository_id / build_runs.repository_id 는 bigint.
// strconv.ParseInt 변환.

const (
	giteaBuildTool = "gitea-build" // quality_snapshots.tool 의 본 sprint 정합 값
)

// UpsertPullActivity inserts (or updates) a single pr_activities row for the given
// (repository_id, external_pr_id, event_type, occurred_at) tuple. The UNIQUE constraint
// `pr_activities_event_unique` on those four columns drives the ON CONFLICT branch.
//
//   - repositoryID: DevHub repositories.id (bigint) — string form for cross-package signature stability.
//   - giteaPRID:    Gitea pull request id (stored as text `external_pr_id`).
//   - number:       Gitea pull request number (#NNN).
//   - eventType:    Caller-decided `pr_activities.event_type` enum value
//                   (opened/reviewed/commented/closed/merged/reopened/updated). Migration 000001 L411 CHECK.
//   - state/title/body/headSHA/authorLogin: stored as-is (title/body truncated to TEXT cap-free).
//   - updatedAt:    `occurred_at` value. UNIQUE constraint conflict-key column.
func (s *PostgresStore) UpsertPullActivity(
	ctx context.Context,
	repositoryID string,
	giteaPRID int64,
	number int,
	eventType, title, body, headSHA, authorLogin string,
	updatedAt time.Time,
) error {
	repoIDInt, err := strconv.ParseInt(repositoryID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse repository id %q: %w", repositoryID, err)
	}
	if eventType == "" {
		eventType = "updated" // defensive fallback, CHECK constraint 통과
	}
	// pr_activities schema (migration 000001 L402) column list 1:1 매핑.
	// number / title / body / head_sha 는 별도 column 미존재 → payload jsonb 에 묶어 저장
	// (ingest pipeline 별도 sprint 의 정규화 단계 전 임시 정공법 — 후속에서 컬럼 분리 가능).
	const query = `
INSERT INTO pr_activities (
    repository_id, external_pr_id, event_type, actor_login, occurred_at, payload
) VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6::jsonb)
ON CONFLICT (repository_id, external_pr_id, event_type, occurred_at) DO UPDATE SET
    actor_login = EXCLUDED.actor_login,
    payload = EXCLUDED.payload`
	// payload jsonb — number/title/body/head_sha 묶음 (ingest pipeline 별도 sprint
	// 에서 컬럼 분리 가능; 현재는 Sprint A 의 minimum viable storage).
	payload := map[string]any{
		"number":   number,
		"title":    title,
		"body":     body,
		"head_sha": headSHA,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal pr_activities payload: %w", err)
	}
	_, err = s.pool.Exec(ctx, query,
		repoIDInt,
		strconv.FormatInt(giteaPRID, 10),
		eventType,
		authorLogin,
		updatedAt.UTC(),
		string(payloadJSON),
	)
	if err != nil {
		return fmt.Errorf("upsert pr_activities: %w", err)
	}
	return nil
}

// UpsertBuildRun inserts (or updates) a single build_runs row. UNIQUE constraint
// `build_runs_run_external_id_key` on `run_external_id` drives the ON CONFLICT branch.
//
//   - repositoryID: DevHub repositories.id (bigint, string form).
//   - giteaBuildID: Gitea Actions run id (stored as text `run_external_id`).
//   - commitSHA:    Gitea head SHA (also propagated to quality_snapshots by the adapter).
//   - event:        Gitea Actions trigger event (push/pull_request/schedule). Stored
//                   in the `branch` column when non-empty (the schema has no `event` column).
//                   If empty, the caller must pass the repo's default branch via event.
//   - status:       Gitea status (queued/running/success/failed/cancelled/skipped/unknown).
//                   CHECK constraint `build_runs_status_check` 정합 (migration 000001 L107).
//   - conclusion:   Accepted but ignored — `status` already encodes the conclusion
//                   (success/failed/cancelled/skipped). Stored in metric_payload 의 여지는 후속.
//   - createdAt:    Gitea run creation time → `started_at` (and `finished_at` estimate).
func (s *PostgresStore) UpsertBuildRun(
	ctx context.Context,
	repositoryID string,
	giteaBuildID int64,
	commitSHA, event, status, conclusion string,
	createdAt time.Time,
) error {
	repoIDInt, err := strconv.ParseInt(repositoryID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse repository id %q: %w", repositoryID, err)
	}
	branch := event
	if branch == "" {
		// Caller did not resolve repositories.default_branch. We fall back to "main"
		// rather than failing the upsert (NOT NULL constraint; X-5 §1.4 결정 — operationally
		// the Gitea event field is always present, so this branch is defensive).
		branch = "main"
	}
	// normalize Gitea Actions completed runs: status="completed" + conclusion=success/failure
	// → store.status = conclusion (CHECK constraint `build_runs_status_check` 정합).
	// 미 normalize 시 Postgres reject; 정상 빌드도 ingest 안 됨 (codex P1).
	normalized := status
	if status == "completed" && conclusion != "" {
		normalized = conclusion
	}
	const query = `
INSERT INTO build_runs (
    repository_id, run_external_id, branch, commit_sha, status,
    duration_seconds, started_at, finished_at
) VALUES ($1, $2, $3, $4, $5, 0, $6, $6)
ON CONFLICT (run_external_id) DO UPDATE SET
    repository_id = EXCLUDED.repository_id,
    branch = EXCLUDED.branch,
    commit_sha = EXCLUDED.commit_sha,
    status = EXCLUDED.status`
	_, err = s.pool.Exec(ctx, query,
		repoIDInt,
		strconv.FormatInt(giteaBuildID, 10),
		branch,
		commitSHA,
		normalized,
		createdAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("upsert build_runs: %w", err)
	}
	return nil
}

// UpsertQualitySnapshot inserts (or updates) a single quality_snapshots row.
// UNIQUE constraint is the partial index `quality_snapshots_repo_ref_unique` on
// (repository_id, ref_name) WHERE tool = 'gitea-build' (migration 000045, 본 follow-up
// 신규). Other tools' snapshots with the same ref_name coexist (partial WHERE clause).
//
//   - repositoryID: DevHub repositories.id (bigint, string form).
//   - commitSHA:    Gitea head SHA. ref_name 으로도 사용 (git 40자 SHA 그대로; truncation 불요).
//   - recordedAt:   Adapter-supplied time (typically Gitea run updated_at) → `measured_at`.
func (s *PostgresStore) UpsertQualitySnapshot(
	ctx context.Context,
	repositoryID, commitSHA string,
	recordedAt time.Time,
) error {
	repoIDInt, err := strconv.ParseInt(repositoryID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse repository id %q: %w", repositoryID, err)
	}
	if commitSHA == "" {
		return fmt.Errorf("commit_sha is required for quality_snapshots")
	}
	const query = `
INSERT INTO quality_snapshots (
    repository_id, tool, ref_name, commit_sha, metric_payload, measured_at
) VALUES ($1, $2, $3, $4, '{}'::jsonb, $5)
ON CONFLICT (repository_id, ref_name) WHERE tool = 'gitea-build' DO UPDATE SET
    commit_sha = EXCLUDED.commit_sha,
    measured_at = EXCLUDED.measured_at`
	_, err = s.pool.Exec(ctx, query,
		repoIDInt,
		giteaBuildTool,
		commitSHA,
		commitSHA,
		recordedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("upsert quality_snapshots: %w", err)
	}
	return nil
}
