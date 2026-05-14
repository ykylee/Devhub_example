package store

import (
	"context"
	"errors"

	"github.com/devhub/backend-core/internal/domain"
)

// ErrNotImplemented signals that a method is registered in the store interface but
// its PostgreSQL body has not been written yet. Sprint claude/work_260514-a 의 carve in 은
// store 인터페이스 + handler stub 까지이며, 실 SQL 구현은 후속 sprint 의 carve out.
var ErrNotImplemented = errors.New("not implemented")

// ApplicationListOptions parameterizes ListApplications.
type ApplicationListOptions struct {
	Status          string
	IncludeArchived bool
	Query           string
	Limit           int
	Offset          int
}

// ApplicationRepositoryLinkKey identifies a single link row (composite PK).
type ApplicationRepositoryLinkKey struct {
	ApplicationID string
	RepoProvider  string
	RepoFullName  string
}

// ProjectListOptions parameterizes ListProjects (within a Repository scope).
type ProjectListOptions struct {
	RepositoryID    int64
	Status          string
	IncludeArchived bool
	Limit           int
	Offset          int
}

// --- Applications ---

func (s *PostgresStore) ListApplications(ctx context.Context, opts ApplicationListOptions) ([]domain.Application, int, error) {
	return nil, 0, ErrNotImplemented
}

func (s *PostgresStore) GetApplication(ctx context.Context, applicationID string) (domain.Application, error) {
	return domain.Application{}, ErrNotImplemented
}

func (s *PostgresStore) CreateApplication(ctx context.Context, app domain.Application) (domain.Application, error) {
	return domain.Application{}, ErrNotImplemented
}

// UpdateApplication mutates allowed fields and enforces the ADR-0011 §4.1 1차 정책
// (system_admin 일임) via the handler layer; this store method is unconditional.
// 상태 전이 가드 검증은 concept §13.2.1 에 따라 handler/service 에서 수행하고, store 는
// 단순히 단일 SQL UPDATE 만 책임진다.
func (s *PostgresStore) UpdateApplication(ctx context.Context, app domain.Application) (domain.Application, error) {
	return domain.Application{}, ErrNotImplemented
}

// ArchiveApplication is the soft-delete entry point (api §13.2 DELETE = archive).
func (s *PostgresStore) ArchiveApplication(ctx context.Context, applicationID, archivedReason string) (domain.Application, error) {
	return domain.Application{}, ErrNotImplemented
}

// --- Application-Repository link ---

func (s *PostgresStore) ListApplicationRepositories(ctx context.Context, applicationID string) ([]domain.ApplicationRepository, error) {
	return nil, ErrNotImplemented
}

func (s *PostgresStore) CreateApplicationRepository(ctx context.Context, link domain.ApplicationRepository) (domain.ApplicationRepository, error) {
	return domain.ApplicationRepository{}, ErrNotImplemented
}

func (s *PostgresStore) DeleteApplicationRepository(ctx context.Context, key ApplicationRepositoryLinkKey) error {
	return ErrNotImplemented
}

// UpdateApplicationRepositorySync 는 webhook/pull worker 가 sync_status / sync_error_code 를
// link 단위로 1건 캐시 갱신할 때 호출 (api §13.3 운영 룰).
func (s *PostgresStore) UpdateApplicationRepositorySync(ctx context.Context, key ApplicationRepositoryLinkKey, status domain.ApplicationRepositorySyncStatus, errorCode domain.SyncErrorCode) error {
	return ErrNotImplemented
}

// --- SCM Provider catalog ---

func (s *PostgresStore) ListSCMProviders(ctx context.Context) ([]domain.SCMProvider, error) {
	return nil, ErrNotImplemented
}

func (s *PostgresStore) UpdateSCMProvider(ctx context.Context, provider domain.SCMProvider) (domain.SCMProvider, error) {
	return domain.SCMProvider{}, ErrNotImplemented
}

// --- Projects ---

func (s *PostgresStore) ListProjects(ctx context.Context, opts ProjectListOptions) ([]domain.Project, int, error) {
	return nil, 0, ErrNotImplemented
}

func (s *PostgresStore) GetProject(ctx context.Context, projectID string) (domain.Project, error) {
	return domain.Project{}, ErrNotImplemented
}

func (s *PostgresStore) CreateProject(ctx context.Context, project domain.Project) (domain.Project, error) {
	return domain.Project{}, ErrNotImplemented
}

func (s *PostgresStore) UpdateProject(ctx context.Context, project domain.Project) (domain.Project, error) {
	return domain.Project{}, ErrNotImplemented
}

func (s *PostgresStore) ArchiveProject(ctx context.Context, projectID, archivedReason string) (domain.Project, error) {
	return domain.Project{}, ErrNotImplemented
}
