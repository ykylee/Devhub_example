package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// kratosPasswordChangedPayload is the minimal payload our Kratos web_hook
// jsonnet template sends. We intentionally keep the schema small (3 fields)
// so the handler does not need to track upstream payload churn.
//
// The jsonnet template in infra/idp/kratos_webhooks/settings_password_after.jsonnet
// produces this exact shape; aligning the two is part of PR-M2-AUDIT
// (claude/login_usermanagement_finish).
type kratosPasswordChangedPayload struct {
	IdentityID string `json:"identity_id"`
	Email      string `json:"email"`
	OccurredAt string `json:"occurred_at"`
}

// kratosWebhookSettingsPasswordAfter handles the
// /api/v1/integrations/kratos/hook/settings/password/after webhook fired by
// Kratos after a successful self-service password change. The handler
// authenticates the inbound call via shared secret (Bearer
// DEVHUB_KRATOS_WEBHOOK_TOKEN), parses the minimal jsonnet payload, and
// records an audit_logs row with source_type=kratos so security audit can
// reconstruct who changed their password and when even without DevHub-side
// /api/v1/account/password calls (self-service flow bypasses the proxy).
//
// Misconfiguration policy: empty DEVHUB_KRATOS_WEBHOOK_TOKEN responds 503,
// not 401, so an operator that forgot to set the secret in production sees
// a loud failure instead of silently accepting every payload. Mismatch
// responds 401.
func (h Handler) kratosWebhookSettingsPasswordAfter(c *gin.Context) {
	if strings.TrimSpace(h.cfg.KratosWebhookToken) == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "kratos webhook secret is not configured",
		})
		return
	}

	if err := verifyKratosWebhookSecret(c.GetHeader("Authorization"), h.cfg.KratosWebhookToken); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthenticated",
			"error":  err.Error(),
		})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "failed to read request body"})
		return
	}
	var payload kratosPasswordChangedPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json payload"})
		return
	}
	identityID := strings.TrimSpace(payload.IdentityID)
	if identityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "identity_id is required"})
		return
	}

	if h.cfg.AuditStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "audit store is not configured",
		})
		return
	}

	// Actor resolution: Kratos self-service flow has no DevHub bearer, so we
	// tag the audit row with the Kratos identity_id. Future enhancement can
	// resolve identity_id → users.user_id via OrganizationStore lookup; for
	// now the identity_id is enough to correlate with users.kratos_identity_id.
	actorLogin := "kratos:" + identityID

	auditPayload := map[string]any{
		"identity_id": identityID,
	}
	if email := strings.TrimSpace(payload.Email); email != "" {
		auditPayload["email"] = email
	}
	if occurred := strings.TrimSpace(payload.OccurredAt); occurred != "" {
		auditPayload["occurred_at"] = occurred
	}

	log := domain.AuditLog{
		ActorLogin: actorLogin,
		Action:     "account.password_changed",
		TargetType: "kratos_identity",
		TargetID:   identityID,
		Payload:    auditPayload,
		SourceIP:   c.ClientIP(),
		RequestID:  requestIDFrom(c),
		SourceType: domain.AuditSourceKratos,
	}
	if _, err := h.cfg.AuditStore.CreateAuditLog(c.Request.Context(), log); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "failed to record audit log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// verifyKratosWebhookSecret accepts both `Bearer <token>` and bare `<token>`
// forms. Kratos web_hook auth.config.value in jsonnet/yaml is typically just
// the token string, but operators occasionally include the "Bearer " prefix
// in tooling — accepting both keeps the contract forgiving while still
// using constant-time comparison.
func verifyKratosWebhookSecret(header, expected string) error {
	got := strings.TrimSpace(header)
	if got == "" {
		return errors.New("authorization header required")
	}
	if rest, ok := bearerToken(got); ok {
		got = rest
	}
	if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
		return errors.New("invalid webhook secret")
	}
	return nil
}
