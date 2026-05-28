package repository

import (
	"github.com/devhub/backend-core/internal/store"
)

type RealtimeRepository struct {
	store *store.PostgresStore
}

func NewRealtimeRepository(s *store.PostgresStore) *RealtimeRepository {
	return &RealtimeRepository{store: s}
}
