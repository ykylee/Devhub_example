package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type memoryCommandStore struct {
	commands []domain.Command
	auditLog domain.AuditLog
	err      error
}

func (s *memoryCommandStore) CreateRiskMitigationCommand(_ context.Context, req domain.RiskMitigationCommandRequest) (domain.Command, domain.AuditLog, bool, error) {
	if s.err != nil {
		return domain.Command{}, domain.AuditLog{}, false, s.err
	}
	for _, command := range s.commands {
		if req.IdempotencyKey != "" && command.IdempotencyKey == req.IdempotencyKey {
			return command, s.auditLog, true, nil
		}
	}
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	command := domain.Command{
		CommandID:        "cmd_test",
		CommandType:      "risk_mitigation",
		TargetType:       "risk",
		TargetID:         req.RiskID,
		ActionType:       req.ActionType,
		Status:           "pending",
		ActorLogin:       req.ActorLogin,
		Reason:           req.Reason,
		DryRun:           req.DryRun,
		RequiresApproval: req.RequiresApproval,
		IdempotencyKey:   req.IdempotencyKey,
		RequestPayload:   req.RequestPayload,
		ResultPayload:    map[string]any{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	auditLog := domain.AuditLog{
		AuditID:    "audit_test",
		ActorLogin: req.ActorLogin,
		Action:     "risk_mitigation.requested",
		TargetType: "risk",
		TargetID:   req.RiskID,
		CommandID:  command.CommandID,
		CreatedAt:  now,
	}
	s.commands = append(s.commands, command)
	s.auditLog = auditLog
	return command, auditLog, false, nil
}

func (s *memoryCommandStore) GetCommand(_ context.Context, commandID string) (domain.Command, error) {
	for _, command := range s.commands {
		if command.CommandID == commandID {
			return command, nil
		}
	}
	return domain.Command{}, store.ErrNotFound
}

func TestCreateRiskMitigationReturnsCommandLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	commandStore := &memoryCommandStore{}
	router := NewRouter(RouterConfig{CommandStore: commandStore})

	body := []byte(`{"action_type":"rerun_ci","reason":"CI failure blocks release","dry_run":true,"idempotency_key":"risk-502-rerun","metadata":{"ci_run_id":"502"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/risks/ci_failure:502/mitigations", bytes.NewReader(body))
	req.Header.Set("X-Devhub-Actor", "yklee")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected %d, got %d: %s", http.StatusAccepted, rec.Code, rec.Body.String())
	}
	var response struct {
		Status string `json:"status"`
		Data   struct {
			CommandID        string `json:"command_id"`
			CommandStatus    string `json:"command_status"`
			AuditLogID       string `json:"audit_log_id"`
			RequiresApproval bool   `json:"requires_approval"`
			IdempotentReplay bool   `json:"idempotent_replay"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Status != "accepted" || response.Data.CommandID != "cmd_test" || response.Data.CommandStatus != "pending" {
		t.Fatalf("unexpected command response: %+v", response)
	}
	if response.Data.AuditLogID != "audit_test" || response.Data.IdempotentReplay {
		t.Fatalf("unexpected audit/idempotency response: %+v", response.Data)
	}
	if len(commandStore.commands) != 1 {
		t.Fatalf("expected command write, got %+v", commandStore.commands)
	}
	command := commandStore.commands[0]
	if command.ActorLogin != "yklee" || command.TargetID != "ci_failure:502" || command.RequestPayload["metadata"] == nil {
		t.Fatalf("unexpected stored command: %+v", command)
	}
}

func TestCreateRiskMitigationRejectsMissingReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(RouterConfig{CommandStore: &memoryCommandStore{}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/risks/ci_failure:502/mitigations", bytes.NewReader([]byte(`{"action_type":"rerun_ci"}`)))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestCreateRiskMitigationReportsIdempotentReplay(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	commandStore := &memoryCommandStore{
		commands: []domain.Command{
			{
				CommandID:      "cmd_existing",
				CommandType:    "risk_mitigation",
				TargetType:     "risk",
				TargetID:       "ci_failure:502",
				ActionType:     "rerun_ci",
				Status:         "pending",
				ActorLogin:     "yklee",
				Reason:         "CI failure blocks release",
				DryRun:         true,
				IdempotencyKey: "risk-502-rerun",
				RequestPayload: map[string]any{},
				ResultPayload:  map[string]any{},
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
		auditLog: domain.AuditLog{AuditID: "audit_existing", CreatedAt: now},
	}
	router := NewRouter(RouterConfig{CommandStore: commandStore})

	body := []byte(`{"action_type":"rerun_ci","reason":"CI failure blocks release","idempotency_key":"risk-502-rerun"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/risks/ci_failure:502/mitigations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected %d, got %d: %s", http.StatusAccepted, rec.Code, rec.Body.String())
	}
	var response struct {
		Data struct {
			CommandID        string `json:"command_id"`
			AuditLogID       string `json:"audit_log_id"`
			IdempotentReplay bool   `json:"idempotent_replay"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.CommandID != "cmd_existing" || response.Data.AuditLogID != "audit_existing" || !response.Data.IdempotentReplay {
		t.Fatalf("unexpected replay response: %+v", response.Data)
	}
	if len(commandStore.commands) != 1 {
		t.Fatalf("expected no duplicate command, got %+v", commandStore.commands)
	}
}

func TestGetCommandReturnsCommandStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	router := NewRouter(RouterConfig{CommandStore: &memoryCommandStore{
		commands: []domain.Command{
			{
				CommandID:      "cmd_test",
				CommandType:    "risk_mitigation",
				TargetType:     "risk",
				TargetID:       "ci_failure:502",
				ActionType:     "rerun_ci",
				Status:         "pending",
				ActorLogin:     "yklee",
				Reason:         "CI failure blocks release",
				DryRun:         true,
				RequestPayload: map[string]any{"action_type": "rerun_ci"},
				ResultPayload:  map[string]any{},
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
	}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/commands/cmd_test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var response struct {
		Data struct {
			CommandID     string `json:"command_id"`
			CommandStatus string `json:"command_status"`
			TargetID      string `json:"target_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.CommandID != "cmd_test" || response.Data.CommandStatus != "pending" || response.Data.TargetID != "ci_failure:502" {
		t.Fatalf("unexpected command response: %+v", response.Data)
	}
}
