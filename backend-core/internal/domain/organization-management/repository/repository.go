package repository

import (
	"github.com/devhub/backend-core/internal/store"
)

type OrganizationRepository struct {
	store *store.PostgresStore
}

func NewOrganizationRepository(s *store.PostgresStore) *OrganizationRepository {
	return &OrganizationRepository{store: s}
}
