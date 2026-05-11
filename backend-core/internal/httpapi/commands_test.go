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

func (s *memoryCommandStore) CreateServiceActionCommand(_ context.Context, req domain.ServiceActionCommandRequest) (domain.Command, domain.AuditLog, bool, error) {
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
		CommandID:        "cmd_service_test",
		CommandType:      "service_action",
		TargetType:       "service",
		TargetID:         req.ServiceID,
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
		AuditID:    "audit_service_test",
		ActorLogin: req.ActorLogin,
		Action:     "service_action.requested",
		TargetType: "service",
		TargetID:   req.ServiceID,
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

func (s *memoryCommandStore) ApproveCommand(_ context.Context, req domain.CommandApprovalRequest) (domain.Command, domain.AuditLog, error) {
	for i, command := range s.commands {
		if command.CommandID != req.CommandID {
			continue
		}
		if command.CommandType != "service_action" || command.Status != "pending" || !command.RequiresApproval {
			return domain.Command{}, domain.AuditLog{}, store.ErrConflict
		}
		command.RequiresApproval = false
		command.ResultPayload = map[string]any{
			"approval": map[string]any{
				"status":      "approved",
				"actor_login": req.ActorLogin,
				"reason":      req.Reason,
			},
		}
		command.UpdatedAt = time.Date(2026, 5, 4, 10, 5, 0, 0, time.UTC)
		s.commands[i] = command
		auditLog := domain.AuditLog{
			AuditID:    "audit_approved",
			ActorLogin: req.ActorLogin,
			Action:     "service_action.approved",
			TargetType: command.TargetType,
			TargetID:   command.TargetID,
			CommandID:  command.CommandID,
			CreatedAt:  command.UpdatedAt,
		}
		s.auditLog = auditLog
		return command, auditLog, nil
	}
	return domain.Command{}, domain.AuditLog{}, store.ErrNotFound
}

func (s *memoryCommandStore) RejectCommand(_ context.Context, req domain.CommandApprovalRequest) (domain.Command, domain.AuditLog, error) {
	for i, command := range s.commands {
		if command.CommandID != req.CommandID {
			continue
		}
		if command.CommandType != "service_action" || command.Status != "pending" || !command.RequiresApproval {
			return domain.Command{}, domain.AuditLog{}, store.ErrConflict
		}
		command.Status = "rejected"
		command.RequiresApproval = false
		command.ResultPayload = map[string]any{
			"approval": map[string]any{
				"status":      "rejected",
				"actor_login": req.ActorLogin,
				"reason":      req.Reason,
			},
		}
		command.UpdatedAt = time.Date(2026, 5, 4, 10, 5, 0, 0, time.UTC)
		s.commands[i] = command
		auditLog := domain.AuditLog{
			AuditID:    "audit_rejected",
			ActorLogin: req.ActorLogin,
			Action:     "service_action.rejected",
			TargetType: command.TargetType,
			TargetID:   command.TargetID,
			CommandID:  command.CommandID,
			CreatedAt:  command.UpdatedAt,
		}
		s.auditLog = auditLog
		return command, auditLog, nil
	}
	return domain.Command{}, domain.AuditLog{}, store.ErrNotFound
}

func TestCreateServiceActionReturnsCommandLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	commandStore := &memoryCommandStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{Login: "admin", Subject: "user-admin", Role: "system_admin"}}
	router := NewRouter(RouterConfig{CommandStore: commandStore, BearerTokenVerifier: verifier})

	body := []byte(`{"service_id":"runner-asia-01","action_type":"restart","reason":"Runner queue is blocked","dry_run":true,"force":false,"idempotency_key":"service-restart-1","metadata":{"queue_depth":12}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer t")
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
	if response.Status != "accepted" || response.Data.CommandID != "cmd_service_test" || response.Data.CommandStatus != "pending" {
		t.Fatalf("unexpected command response: %+v", response)
	}
	if response.Data.AuditLogID != "audit_service_test" || response.Data.RequiresApproval || response.Data.IdempotentReplay {
		t.Fatalf("unexpected audit/approval response: %+v", response.Data)
	}
	if len(commandStore.commands) != 1 {
		t.Fatalf("expected command write, got %+v", commandStore.commands)
	}
	command := commandStore.commands[0]
	if command.CommandType != "service_action" || command.TargetID != "runner-asia-01" || command.ActorLogin != "admin" {
		t.Fatalf("unexpected stored command: %+v", command)
	}
	if command.RequestPayload["metadata"] == nil {
		t.Fatalf("expected metadata in request payload: %+v", command.RequestPayload)
	}
}

func TestCreateServiceActionDryRunAllowsManagerPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	commandStore := &memoryCommandStore{}
	router := NewRouter(RouterConfig{
		CommandStore:    commandStore,
		AuthDevFallback: true,
	})

	body := []byte(`{"service_id":"runner-asia-01","action_type":"restart","reason":"Runner queue is blocked","dry_run":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader(body))
	req.Header.Set("X-Devhub-Role", "manager")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateServiceActionLiveRequiresAdminPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	commandStore := &memoryCommandStore{}
	router := NewRouter(RouterConfig{
		CommandStore: commandStore,
		BearerTokenVerifier: &fakeBearerTokenVerifier{actor: AuthenticatedActor{
			Login: "manager",
			Role:  "manager",
		}},
	})

	body := []byte(`{"service_id":"runner-asia-01","action_type":"restart","reason":"Runner queue is blocked","dry_run":false}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer manager-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateServiceActionRequiresApprovalForLiveAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	commandStore := &memoryCommandStore{}
	router := testRouter(RouterConfig{CommandStore: commandStore})

	body := []byte(`{"service_id":"runner-asia-01","action_type":"restart","reason":"Apply live restart","dry_run":false}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected %d, got %d: %s", http.StatusAccepted, rec.Code, rec.Body.String())
	}
	if len(commandStore.commands) != 1 || !commandStore.commands[0].RequiresApproval {
		t.Fatalf("expected live service action to require approval, got %+v", commandStore.commands)
	}
}

func TestCreateServiceActionReportsIdempotentReplay(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	commandStore := &memoryCommandStore{
		commands: []domain.Command{
			{
				CommandID:      "cmd_service_existing",
				CommandType:    "service_action",
				TargetType:     "service",
				TargetID:       "runner-asia-01",
				ActionType:     "restart",
				Status:         "pending",
				ActorLogin:     "admin",
				Reason:         "Runner queue is blocked",
				DryRun:         true,
				IdempotencyKey: "service-restart-1",
				RequestPayload: map[string]any{},
				ResultPayload:  map[string]any{},
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
		auditLog: domain.AuditLog{AuditID: "audit_service_existing", CreatedAt: now},
	}
	router := testRouter(RouterConfig{CommandStore: commandStore})

	body := []byte(`{"service_id":"runner-asia-01","action_type":"restart","reason":"Runner queue is blocked","idempotency_key":"service-restart-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader(body))
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
	if response.Data.CommandID != "cmd_service_existing" || response.Data.AuditLogID != "audit_service_existing" || !response.Data.IdempotentReplay {
		t.Fatalf("unexpected replay response: %+v", response.Data)
	}
	if len(commandStore.commands) != 1 {
		t.Fatalf("expected no duplicate command, got %+v", commandStore.commands)
	}
}

func TestCreateServiceActionRejectsMissingServiceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testRouter(RouterConfig{CommandStore: &memoryCommandStore{}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader([]byte(`{"action_type":"restart","reason":"Runner queue is blocked"}`)))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestCreateRiskMitigationReturnsCommandLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	commandStore := &memoryCommandStore{}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{Login: "yklee", Subject: "user-yklee", Role: "manager"}}
	router := NewRouter(RouterConfig{CommandStore: commandStore, BearerTokenVerifier: verifier})

	body := []byte(`{"action_type":"rerun_ci","reason":"CI failure blocks release","dry_run":true,"idempotency_key":"risk-502-rerun","metadata":{"ci_run_id":"502"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/risks/ci_failure:502/mitigations", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer t")
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

func TestCreateRiskMitigationRequiresRiskWrite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	commandStore := &memoryCommandStore{}
	router := NewRouter(RouterConfig{
		CommandStore: commandStore,
		BearerTokenVerifier: &fakeBearerTokenVerifier{actor: AuthenticatedActor{
			Login: "developer",
			Role:  "developer",
		}},
	})

	body := []byte(`{"action_type":"rerun_ci","reason":"Fix blocked release"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/risks/risk-1/mitigations", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer developer-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateRiskMitigationRejectsMissingReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testRouter(RouterConfig{CommandStore: &memoryCommandStore{}})

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
	router := testRouter(RouterConfig{CommandStore: commandStore})

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
	router := testRouter(RouterConfig{CommandStore: &memoryCommandStore{
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

func TestApproveServiceActionCommandMarksApprovalAndAudits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	commandStore := &memoryCommandStore{
		commands: []domain.Command{
			{
				CommandID:        "cmd_live",
				CommandType:      "service_action",
				TargetType:       "service",
				TargetID:         "runner-asia-01",
				ActionType:       "restart",
				Status:           "pending",
				ActorLogin:       "admin",
				Reason:           "Apply live restart",
				DryRun:           false,
				RequiresApproval: true,
				RequestPayload:   map[string]any{},
				ResultPayload:    map[string]any{},
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
	}
	router := NewRouter(RouterConfig{CommandStore: commandStore, AuthDevFallback: true})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/commands/cmd_live/approve", bytes.NewReader([]byte(`{"reason":"Approved maintenance window"}`)))
	req.Header.Set("X-Devhub-Actor", "approver")
	req.Header.Set("X-Devhub-Role", "system_admin")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Status string `json:"status"`
		Data   struct {
			CommandID        string `json:"command_id"`
			CommandStatus    string `json:"command_status"`
			RequiresApproval bool   `json:"requires_approval"`
			ApprovalStatus   string `json:"approval_status"`
			AuditLogID       string `json:"audit_log_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Status != "approved" || response.Data.CommandID != "cmd_live" || response.Data.CommandStatus != "pending" || response.Data.RequiresApproval || response.Data.ApprovalStatus != "approved" {
		t.Fatalf("unexpected approval response: %+v", response)
	}
	if response.Data.AuditLogID != "audit_approved" {
		t.Fatalf("expected approval audit id, got %+v", response.Data)
	}
	if commandStore.commands[0].RequiresApproval || commandStore.commands[0].ResultPayload["approval"] == nil {
		t.Fatalf("expected command approval payload, got %+v", commandStore.commands[0])
	}
}

func TestRejectServiceActionCommandMarksRejectedAndAudits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	commandStore := &memoryCommandStore{
		commands: []domain.Command{
			{
				CommandID:        "cmd_live",
				CommandType:      "service_action",
				TargetType:       "service",
				TargetID:         "runner-asia-01",
				ActionType:       "restart",
				Status:           "pending",
				ActorLogin:       "admin",
				Reason:           "Apply live restart",
				DryRun:           false,
				RequiresApproval: true,
				RequestPayload:   map[string]any{},
				ResultPayload:    map[string]any{},
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
	}
	router := NewRouter(RouterConfig{CommandStore: commandStore, AuthDevFallback: true})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/commands/cmd_live/reject", bytes.NewReader([]byte(`{"reason":"Outside maintenance window"}`)))
	req.Header.Set("X-Devhub-Actor", "approver")
	req.Header.Set("X-Devhub-Role", "system_admin")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Status string `json:"status"`
		Data   struct {
			CommandStatus  string `json:"command_status"`
			ApprovalStatus string `json:"approval_status"`
			AuditLogID     string `json:"audit_log_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Status != "rejected" || response.Data.CommandStatus != "rejected" || response.Data.ApprovalStatus != "rejected" || response.Data.AuditLogID != "audit_rejected" {
		t.Fatalf("unexpected reject response: %+v", response)
	}
	if commandStore.commands[0].Status != "rejected" {
		t.Fatalf("expected rejected command, got %+v", commandStore.commands[0])
	}
}

func TestApproveCommandRequiresCommandAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(RouterConfig{
		CommandStore: &memoryCommandStore{},
		BearerTokenVerifier: &fakeBearerTokenVerifier{actor: AuthenticatedActor{
			Login: "manager",
			Role:  "manager",
		}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/commands/cmd_live/approve", bytes.NewReader([]byte(`{"reason":"approve"}`)))
	req.Header.Set("Authorization", "Bearer manager-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApproveCommandRejectsNonPendingApproval(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	router := NewRouter(RouterConfig{
		CommandStore: &memoryCommandStore{
			commands: []domain.Command{
				{
					CommandID:        "cmd_dry",
					CommandType:      "service_action",
					TargetType:       "service",
					TargetID:         "runner-asia-01",
					ActionType:       "restart",
					Status:           "pending",
					DryRun:           true,
					RequiresApproval: false,
					CreatedAt:        now,
					UpdatedAt:        now,
				},
			},
		},
		AuthDevFallback: true,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/commands/cmd_dry/approve", bytes.NewReader([]byte(`{"reason":"approve"}`)))
	req.Header.Set("X-Devhub-Role", "system_admin")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// TestServiceAction_DryRunVsLiveBoundary locks DEC-2=A' (work_26_05_11-b):
// dry-run and live both persist a command + audit row, but they differ in
// the requires_approval flag and the worker that consumes them. The store
// + audit boundary stays identical so operator intent is auditable for both.
func TestServiceAction_DryRunVsLiveBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name             string
		body             string
		wantDryRun       bool
		wantRequiresAppr bool
	}{
		{
			name:             "dry-run defaults to no approval gate",
			body:             `{"service_id":"runner-asia-01","action_type":"restart","reason":"queue blocked","dry_run":true}`,
			wantDryRun:       true,
			wantRequiresAppr: false,
		},
		{
			name:             "live action requires approval before executor runs",
			body:             `{"service_id":"runner-asia-01","action_type":"restart","reason":"queue blocked","dry_run":false}`,
			wantDryRun:       false,
			wantRequiresAppr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			commandStore := &memoryCommandStore{}
			audits := &memoryAuditStore{}
			router := NewRouter(RouterConfig{
				CommandStore: commandStore,
				AuditStore:   audits,
				BearerTokenVerifier: &fakeBearerTokenVerifier{actor: AuthenticatedActor{
					Login: "admin", Subject: "user-admin", Role: "system_admin",
				}},
			})

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/service-actions", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Authorization", "Bearer t")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusAccepted {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if len(commandStore.commands) != 1 {
				t.Fatalf("dry-run and live both persist exactly one command (got %d)", len(commandStore.commands))
			}
			cmd := commandStore.commands[0]
			if cmd.DryRun != tc.wantDryRun {
				t.Errorf("DryRun=%v want %v", cmd.DryRun, tc.wantDryRun)
			}
			if cmd.RequiresApproval != tc.wantRequiresAppr {
				t.Errorf("RequiresApproval=%v want %v", cmd.RequiresApproval, tc.wantRequiresAppr)
			}
			if cmd.Status != string(domain.CommandStatusPending) {
				t.Errorf("Status=%q want %q (worker pickup is what advances it)", cmd.Status, domain.CommandStatusPending)
			}
		})
	}
}
