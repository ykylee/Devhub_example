package store

import (
	"context"
	"fmt"
)

// ListGiteaPullTargets returns the per-cycle list of repositories that the Gitea
// hourly pull loop should sync. Filtering:
//   - integration_providers.provider_type='scm' AND provider_key='gitea'  (Gitea 한정)
//   - repositories.repository_status='active'                             (draft 제외)
//   - repositories.gitea_repository_id IS NOT NULL                        (Gitea 연동된 것만)
//   - repository_pull_state.backoff_until IS NULL OR < now()              (backoff 아닌 것만)
//   - repositories.deleted_at IS NULL                                     (소프트 삭제 제외)
//
// Returns store-owned type `GiteaPullTarget` (NOT adapters.RepositoryTarget) —
// layering rule: store does not import adapters. The caller (main.go's repoLister
// closure) maps to adapters.RepositoryTarget.
//
// Cold start (state row 부재) 시 LEFT JOIN 으로 repo 포함 (backoff_until IS NULL 분기).
//
// ID slot: IMPL-GITEA-PULL-STORE-01 의 targets 부분.

// GiteaPullTarget is a store-package-local row type for the lister.
// Mirrors adapters.RepositoryTarget field-for-field, kept independent to avoid
// store→adapters import cycles. main.go 의 repoLister closure 가 매핑한다.
type GiteaPullTarget struct {
	ID         int64
	Owner      string
	Name       string
	ExternalID int64
}

// ListGiteaPullTargets implements the X-5 hourly pull lister query.
func (s *PostgresStore) ListGiteaPullTargets(
	ctx context.Context,
) ([]GiteaPullTarget, error) {
	const query = `
SELECT
    r.id,
    COALESCE(r.owner_login, ''),
    r.name,
    COALESCE(r.gitea_repository_id, 0)
FROM repositories r
JOIN integration_providers p ON r.provider_id = p.provider_id
LEFT JOIN repository_pull_state s ON r.id = s.repository_id
WHERE p.provider_type = 'scm'
  AND p.provider_key = 'gitea'
  AND r.repository_status = 'active'
  AND r.gitea_repository_id IS NOT NULL
  AND (s.backoff_until IS NULL OR s.backoff_until < now())
  AND r.deleted_at IS NULL
ORDER BY r.id ASC`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list gitea pull targets: %w", err)
	}
	defer rows.Close()

	out := make([]GiteaPullTarget, 0)
	for rows.Next() {
		var t GiteaPullTarget
		if err := rows.Scan(&t.ID, &t.Owner, &t.Name, &t.ExternalID); err != nil {
			return nil, fmt.Errorf("scan gitea pull target: %w", err)
		}
		if t.ExternalID <= 0 {
			// Defensive: gitea_repository_id IS NOT NULL filter already excludes this,
			// but a zero value would propagate as garbage Gitea API id.
			continue
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gitea pull targets: %w", err)
	}
	return out, nil
}

