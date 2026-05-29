package repository

import (
	"github.com/devhub/backend-core/internal/store"
)

type IntegrationRepository struct {
	store *store.PostgresStore
	*store.PostgresStore
}

func NewIntegrationRepository(s *store.PostgresStore) *IntegrationRepository {
	return &IntegrationRepository{
		store:         s,
		PostgresStore: s,
	}
}
