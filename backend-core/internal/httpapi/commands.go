package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type riskMitigationRequest struct {
	ActionType     string         `json:"action_type"`
	Reason         string         `json:"reason"`
	DryRun         *bool          `json:"dry_run"`
	IdempotencyKey string         `json:"idempotency_key"`
	Metadata       map[string]any `json:"metadata"`
}

type serviceActionRequest struct {
	ServiceID      string         `json:"service_id"`
	ActionType     string         `json:"action_type"`
	Reason         string         `json:"reason"`
	Force          bool           `json:"force"`
	DryRun         *bool          `json:"dry_run"`
	IdempotencyKey string         `json:"idempotency_key"`
	Metadata       map[string]any `json:"metadata"`
}

type commandResponse struct {
	CommandID        string         `json:"command_id"`
	CommandType      string         `json:"command_type"`
	TargetType       string         `json:"target_type"`
	TargetID         string         `json:"target_id"`
	ActionType       string         `json:"action_type"`
	CommandStatus    string         `json:"command_status"`
	ActorLogin       string         `json:"actor_login"`
	Reason           string         `json:"reason"`
	DryRun           bool           `json:"dry_run"`
	RequiresApproval bool           `json:"requires_approval"`
	IdempotencyKey   string         `json:"idempotency_key,omitempty"`
	RequestPayload   map[string]any `json:"request_payload"`
	ResultPayload    map[string]any `json:"result_payload"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type commandAcceptedResponse struct {
	CommandID        string    `json:"command_id"`
	CommandStatus    string    `json:"command_status"`
	RequiresApproval bool      `json:"requires_approval"`
	AuditLogID       string    `json:"audit_log_id"`
	IdempotentReplay bool      `json:"idempotent_replay"`
	CreatedAt        time.Time `json:"created_at"`
}

func (h Handler) createServiceAction(c *gin.Context) {
	if h.cfg.CommandStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "command store is not configured",
		})
		return
	}

	var request serviceActionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "request body must be valid JSON"})
		return
	}
	request.ServiceID = strings.TrimSpace(request.ServiceID)
	request.ActionType = strings.TrimSpace(request.ActionType)
	request.Reason = strings.TrimSpace(request.Reason)
	request.IdempotencyKey = strings.TrimSpace(request.IdempotencyKey)
	if request.ServiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "service_id is required"})
		return
	}
	if request.ActionType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "action_type is required"})
		return
	}
	if request.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "reason is required"})
		return
	}

	dryRun := true
	if request.DryRun != nil {
		dryRun = *request.DryRun
	}
	requiresApproval := request.Force || !dryRun
	payload := map[string]any{
		"service_id":  request.ServiceID,
		"action_type": request.ActionType,
		"dry_run":     dryRun,
		"force":       request.Force,
		"reason":      request.Reason,
	}
	if request.Metadata != nil {
		payload["metadata"] = request.Metadata
	}

	command, auditLog, replayed, err := h.cfg.CommandStore.CreateServiceActionCommand(c.Request.Context(), domain.ServiceActionCommandRequest{
		ServiceID:        request.ServiceID,
		ActorLogin:       actorLogin(c),
		ActionType:       request.ActionType,
		Reason:           request.Reason,
		Force:            request.Force,
		DryRun:           dryRun,
		IdempotencyKey:   request.IdempotencyKey,
		RequestPayload:   payload,
		RequiresApproval: requiresApproval,
	})
	if err != nil {
		writeServerError(c, err, "commands.create_service_action")
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status": "accepted",
		"data": commandAcceptedResponse{
			CommandID:        command.CommandID,
			CommandStatus:    command.Status,
			RequiresApproval: command.RequiresApproval,
			AuditLogID:       auditLog.AuditID,
			IdempotentReplay: replayed,
			CreatedAt:        command.CreatedAt,
		},
	})
}

func (h Handler) createRiskMitigation(c *gin.Context) {
	if h.cfg.CommandStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "command store is not configured",
		})
		return
	}

	riskID := strings.TrimSpace(c.Param("risk_id"))
	if riskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "risk_id is required"})
		return
	}

	var request riskMitigationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "request body must be valid JSON"})
		return
	}
	request.ActionType = strings.TrimSpace(request.ActionType)
	request.Reason = strings.TrimSpace(request.Reason)
	request.IdempotencyKey = strings.TrimSpace(request.IdempotencyKey)
	if request.ActionType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "action_type is required"})
		return
	}
	if request.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "reason is required"})
		return
	}

	dryRun := true
	if request.DryRun != nil {
		dryRun = *request.DryRun
	}
	payload := map[string]any{
		"action_type": request.ActionType,
		"dry_run":     dryRun,
		"reason":      request.Reason,
	}
	if request.Metadata != nil {
		payload["metadata"] = request.Metadata
	}

	command, auditLog, replayed, err := h.cfg.CommandStore.CreateRiskMitigationCommand(c.Request.Context(), domain.RiskMitigationCommandRequest{
		RiskID:         riskID,
		ActorLogin:     actorLogin(c),
		ActionType:     request.ActionType,
		Reason:         request.Reason,
		DryRun:         dryRun,
		IdempotencyKey: request.IdempotencyKey,
		RequestPayload: payload,
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "risk not found"})
			return
		}
		writeServerError(c, err, "commands.create_risk_mitigation")
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status": "accepted",
		"data": commandAcceptedResponse{
			CommandID:        command.CommandID,
			CommandStatus:    command.Status,
			RequiresApproval: command.RequiresApproval,
			AuditLogID:       auditLog.AuditID,
			IdempotentReplay: replayed,
			CreatedAt:        command.CreatedAt,
		},
	})
}

func (h Handler) getCommand(c *gin.Context) {
	if h.cfg.CommandStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "command store is not configured",
		})
		return
	}

	commandID := strings.TrimSpace(c.Param("command_id"))
	if commandID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "command_id is required"})
		return
	}

	command, err := h.cfg.CommandStore.GetCommand(c.Request.Context(), commandID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "command not found"})
			return
		}
		writeServerError(c, err, "commands.get_command")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   commandFromDomain(command),
	})
}

func commandFromDomain(command domain.Command) commandResponse {
	return commandResponse{
		CommandID:        command.CommandID,
		CommandType:      command.CommandType,
		TargetType:       command.TargetType,
		TargetID:         command.TargetID,
		ActionType:       command.ActionType,
		CommandStatus:    command.Status,
		ActorLogin:       command.ActorLogin,
		Reason:           command.Reason,
		DryRun:           command.DryRun,
		RequiresApproval: command.RequiresApproval,
		IdempotencyKey:   command.IdempotencyKey,
		RequestPayload:   command.RequestPayload,
		ResultPayload:    command.ResultPayload,
		CreatedAt:        command.CreatedAt,
		UpdatedAt:        command.UpdatedAt,
	}
}

func actorLogin(c *gin.Context) string {
	return requestActor(c).Login
}

type requestActorInfo struct {
	Login  string
	Source string
}

func requestActor(c *gin.Context) requestActorInfo {
	if value, ok := c.Get("devhub_actor_login"); ok {
		if actor, ok := value.(string); ok {
			actor = strings.TrimSpace(actor)
			if actor != "" {
				return requestActorInfo{Login: actor, Source: "authenticated_context"}
			}
		}
	}
	return requestActorInfo{Login: "system", Source: "system_fallback"}
}

func devFallbackEnabled(c *gin.Context) bool {
	value, ok := c.Get("devhub_auth_dev_fallback")
	if !ok {
		return false
	}
	enabled, ok := value.(bool)
	return ok && enabled
}
