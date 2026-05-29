package repository

import (
	intgregrep "github.com/devhub/backend-core/internal/domain/integration-registry/repository"
	"github.com/devhub/backend-core/internal/store"
)

type ApplicationRepository struct {
	store *store.PostgresStore
	*store.PostgresStore
	*intgregrep.IntegrationRepository
}

func NewApplicationRepository(s *store.PostgresStore) *ApplicationRepository {
	return &ApplicationRepository{
		store:                 s,
		PostgresStore:         s,
		IntegrationRepository: intgregrep.NewIntegrationRepository(s),
	}
}
