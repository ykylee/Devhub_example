package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

type memoryAuditStore struct {
	logs []domain.AuditLog
}

func (s *memoryAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if log.AuditID == "" {
		log.AuditID = "audit-test"
		if len(s.logs) > 0 {
			log.AuditID = log.AuditID + "-" + string(rune('0'+len(s.logs)))
		}
	}
	if log.ActorLogin == "" {
		log.ActorLogin = "system"
	}
	if log.Payload == nil {
		log.Payload = map[string]any{}
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Date(2026, 5, 7, 12, len(s.logs), 0, 0, time.UTC)
	}
	s.logs = append(s.logs, log)
	return log, nil
}

func (s *memoryAuditStore) ListAuditLogs(_ context.Context, opts store.ListAuditLogsOptions) ([]domain.AuditLog, error) {
	out := make([]domain.AuditLog, 0, len(s.logs))
	for _, log := range s.logs {
		if opts.ActorLogin != "" && log.ActorLogin != opts.ActorLogin {
			continue
		}
		if opts.Action != "" && log.Action != opts.Action {
			continue
		}
		if opts.TargetType != "" && log.TargetType != opts.TargetType {
			continue
		}
		if opts.TargetID != "" && log.TargetID != opts.TargetID {
			continue
		}
		if opts.CommandID != "" && log.CommandID != opts.CommandID {
			continue
		}
		out = append(out, log)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func TestListAuditLogsFiltersByTarget(t *testing.T) {
	audits := &memoryAuditStore{}
	_, _ = audits.CreateAuditLog(context.Background(), domain.AuditLog{
		ActorLogin: "admin",
		Action:     "user.created",
		TargetType: "user",
		TargetID:   "u1",
	})
	_, _ = audits.CreateAuditLog(context.Background(), domain.AuditLog{
		ActorLogin: "admin",
		Action:     "org_unit.created",
		TargetType: "org_unit",
		TargetID:   "team-a",
	})
	router := NewRouter(RouterConfig{AuditStore: audits})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs?target_type=user", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data []struct {
			Action     string `json:"action"`
			TargetType string `json:"target_type"`
			TargetID   string `json:"target_id"`
		} `json:"data"`
		Meta struct {
			Count int `json:"count"`
		} `json:"meta"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)
	if resp.Meta.Count != 1 || len(resp.Data) != 1 {
		t.Fatalf("expected one audit log, got count=%d len=%d", resp.Meta.Count, len(resp.Data))
	}
	if resp.Data[0].Action != "user.created" || resp.Data[0].TargetType != "user" || resp.Data[0].TargetID != "u1" {
		t.Fatalf("unexpected audit log: %+v", resp.Data[0])
	}
}

func TestCreateUserWritesAuditLogWithActorWarning(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	audits := &memoryAuditStore{}
	router := NewRouter(RouterConfig{OrganizationStore: orgs, AuditStore: audits})

	body := []byte(`{
		"user_id": "u-audit",
		"email": "audit@example.com",
		"display_name": "Audit User",
		"role": "developer",
		"status": "active"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("X-Devhub-Actor", "admin")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Devhub-Actor-Deprecated") != "true" {
		t.Fatalf("expected X-Devhub-Actor deprecation header")
	}
	if len(audits.logs) != 1 {
		t.Fatalf("expected one audit log, got %d", len(audits.logs))
	}
	log := audits.logs[0]
	if log.ActorLogin != "admin" || log.Action != "user.created" || log.TargetType != "user" || log.TargetID != "u-audit" {
		t.Fatalf("unexpected audit log: %+v", log)
	}
	if log.Payload["actor_source"] != "x-devhub-actor" {
		t.Fatalf("expected actor_source payload, got %+v", log.Payload)
	}
}
