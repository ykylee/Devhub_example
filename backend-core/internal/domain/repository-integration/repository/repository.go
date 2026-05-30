package repository

import (
	"context"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

// RepositoryStore defines the store-level operations needed by the
// repository-integration domain. Implemented by *store.PostgresStore.
type RepositoryStore interface {
	UpsertRepository(ctx context.Context, repository domain.Repository) error
	ListRepositoriesByProvider(ctx context.Context, providerID string) ([]domain.Repository, error)
}

// IntegrationRepository wraps *store.PostgresStore and exposes the store
// methods required by the repository-integration domain.
type IntegrationRepository struct {
	*store.PostgresStore
}

// NewIntegrationRepository creates a new IntegrationRepository backed by
// the given PostgresStore.
func NewIntegrationRepository(s *store.PostgresStore) *IntegrationRepository {
	return &IntegrationRepository{PostgresStore: s}
}
