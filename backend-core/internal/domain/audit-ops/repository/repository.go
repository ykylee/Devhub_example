package repository

import (
	"github.com/devhub/backend-core/internal/store"
)

type AuditRepository struct {
	store *store.PostgresStore
}

func NewAuditRepository(s *store.PostgresStore) *AuditRepository {
	return &AuditRepository{store: s}
}
