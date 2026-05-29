package repository

import (
	"github.com/devhub/backend-core/internal/store"
)

type DevRequestRepository struct {
	store *store.PostgresStore
}

func NewDevRequestRepository(s *store.PostgresStore) *DevRequestRepository {
	return &DevRequestRepository{store: s}
}
