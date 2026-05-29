package repository

import (
	"github.com/devhub/backend-core/internal/store"
)

// ApplicationRepository — application-lifecycle 도메인의 persistence root.
// issue #421 (sprint claude/work_260529-n) — 기존에 `*intgregrep.IntegrationRepository`
// 를 embed 해 cross-domain 상향 호출을 유발했음 (호출 규칙 1 "상향 호출 금지" 위반).
// 본 sprint 에서 embed 제거. integration-registry 의 메서드가 필요한 호출자는
// `intgregrep.NewIntegrationRepository(pgStore)` 를 별도로 구성해 inject 한다.
type ApplicationRepository struct {
	store *store.PostgresStore
	*store.PostgresStore
}

func NewApplicationRepository(s *store.PostgresStore) *ApplicationRepository {
	return &ApplicationRepository{
		store:         s,
		PostgresStore: s,
	}
}
