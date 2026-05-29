package repository

import (
	"github.com/devhub/backend-core/internal/store"
)

type RBACRepository struct {
	store *store.PostgresStore
}

func NewRBACRepository(s *store.PostgresStore) *RBACRepository {
	return &RBACRepository{store: s}
}
