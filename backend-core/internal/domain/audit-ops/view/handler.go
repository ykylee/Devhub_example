package view

import (
	"context"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

type AuditStore interface {
	CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
	ListAuditLogs(ctx context.Context, opts store.ListAuditLogsOptions) ([]domain.AuditLog, error)
}

type AuditConfig struct {
	AuditStore            AuditStore
	KeycloakWebhookSecret string
}

type AuditHandler struct {
	cfg AuditConfig
}

func NewAuditHandler(cfg AuditConfig) *AuditHandler {
	return &AuditHandler{cfg: cfg}
}
