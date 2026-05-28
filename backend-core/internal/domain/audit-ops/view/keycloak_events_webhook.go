package view

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

type keycloakWebhookRequest struct {
	ID            string            `json:"id"`
	Time          int64             `json:"time"`
	Type          string            `json:"type"`          // Present for User Event
	OperationType string            `json:"operationType"` // Present for Admin Event
	RealmID       string            `json:"realmId"`
	ClientID      string            `json:"clientId"`
	UserID        string            `json:"userId"`
	IPAddress     string            `json:"ipAddress"`
	Details       map[string]string `json:"details"`
	ResourceType  string            `json:"resourceType"`
	ResourcePath  string            `json:"resourcePath"`
	AuthDetails   struct {
		RealmID   string `json:"realmId"`
		ClientID  string `json:"clientId"`
		UserID    string `json:"userId"`
		IPAddress string `json:"ipAddress"`
	} `json:"authDetails"`
	Error string `json:"error"`
}

func (h *AuditHandler) ReceiveKeycloakEventWebhook(c *gin.Context) {
	// Verify webhook secret — fail-closed if not configured. The route is
	// registered outside the v1 auth group, so without this check anyone able to
	// reach /api/v1/internal/keycloak-events could POST arbitrary events and
	// create audit rows. A missing DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET => 503 so
	// misconfigured deployments fail loud.
	if h.cfg.KeycloakWebhookSecret == "" {
		log.Printf("[keycloak-webhook] Rejecting request: DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET not configured")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "keycloak webhook secret not configured"})
		return
	}
	secretHeader := c.GetHeader("X-Webhook-Secret")
	if secretHeader != h.cfg.KeycloakWebhookSecret {
		log.Printf("[keycloak-webhook] Rejecting request: invalid or missing X-Webhook-Secret header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 2. Parse JSON payload
	var req keycloakWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[keycloak-webhook] Bind JSON error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	var action, targetType, targetID, evHash string
	var payload map[string]any

	if req.Type != "" {
		// User Event
		action, targetType, targetID = mapUserEventToAudit(req)
		evHash = hashUserEvent(req)
		payload = userEventPayload(req)
	} else if req.OperationType != "" {
		// Admin Event
		action, targetType, targetID = mapAdminEventToAudit(req)
		evHash = hashAdminEvent(req)
		payload = adminEventPayload(req)
	} else {
		log.Printf("[keycloak-webhook] Received event payload with neither Type nor OperationType")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unrecognized Keycloak event structure"})
		return
	}

	actorLogin := "system:keycloak-event"
	if req.UserID != "" {
		actorLogin = req.UserID
	} else if req.AuthDetails.UserID != "" {
		actorLogin = req.AuthDetails.UserID
	}

	auditLog := domain.AuditLog{
		ActorLogin:    actorLogin,
		Action:        action,
		TargetType:    targetType,
		TargetID:      targetID,
		SourceType:    domain.AuditSourceKeycloakEvent,
		SourceEventID: evHash,
		Payload:       payload,
	}

	// 3. Save to database via AuditStore
	if h.cfg.AuditStore != nil {
		_, err := h.cfg.AuditStore.CreateAuditLog(c.Request.Context(), auditLog)
		if err != nil {
			// If partial unique index blocks due to duplicate, we return 200 OK (idempotent duplicate ignore)
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				log.Printf("[keycloak-webhook] Ignored duplicate event push: %s", evHash)
				c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "duplicate", "event_id": evHash})
				return
			}
			log.Printf("[keycloak-webhook] Failed to save audit log: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store audit log"})
			return
		}
	} else {
		log.Printf("[keycloak-webhook] AuditStore not configured. Skipping save.")
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "event_id": evHash})
}

func hashUserEvent(ev keycloakWebhookRequest) string {
	sessionID := ev.Details["sessionId"]
	h := sha256.Sum256([]byte(fmt.Sprintf("%d|%s|%s|%s|%s|%s|%s",
		ev.Time, ev.Type, ev.UserID, ev.IPAddress, ev.ClientID, ev.RealmID, sessionID)))
	return hex.EncodeToString(h[:])
}

func hashAdminEvent(ev keycloakWebhookRequest) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d|%s:%s|%s|%s|%s|%s|%s",
		ev.Time, ev.ResourceType, ev.OperationType, ev.ResourcePath,
		ev.AuthDetails.UserID, ev.AuthDetails.ClientID, ev.AuthDetails.IPAddress, ev.RealmID)))
	return hex.EncodeToString(h[:])
}

func mapUserEventToAudit(ev keycloakWebhookRequest) (action, targetType, targetID string) {
	targetID = ev.UserID
	switch ev.Type {
	case "LOGIN":
		return "auth.login.success", "auth", targetID
	case "LOGIN_ERROR":
		return "auth.login.failed", "auth", targetID
	case "LOGOUT":
		return "auth.logout.success", "auth", targetID
	case "LOGOUT_ERROR":
		return "auth.logout.failed", "auth", targetID
	case "REGISTER":
		return "auth.signup.success", "user", targetID
	case "REGISTER_ERROR":
		return "auth.signup.failed", "user", targetID
	case "UPDATE_PASSWORD":
		return "auth.password.changed", "user", targetID
	case "UPDATE_PASSWORD_ERROR":
		return "auth.password.change_failed", "user", targetID
	case "SEND_RESET_PASSWORD":
		return "auth.password.reset_requested", "user", targetID
	case "RESET_PASSWORD":
		return "auth.password.reset_success", "user", targetID
	case "IDENTITY_PROVIDER_LINK_ACCOUNT":
		return "auth.idp.linked", "user", targetID
	case "IDENTITY_PROVIDER_FIRST_LOGIN":
		return "auth.idp.first_login", "user", targetID
	case "VERIFY_EMAIL":
		return "auth.email.verified", "user", targetID
	case "REMOVE_TOTP":
		return "auth.mfa.totp_removed", "user", targetID
	case "UPDATE_TOTP":
		return "auth.mfa.totp_updated", "user", targetID
	default:
		return "keycloak.event.unknown:" + ev.Type, "auth", targetID
	}
}

func mapAdminEventToAudit(ev keycloakWebhookRequest) (action, targetType, targetID string) {
	targetID = ev.ResourcePath
	key := ev.ResourceType + ":" + ev.OperationType
	switch key {
	case "USER:CREATE":
		return "keycloak.user.created", "user", targetID
	case "USER:UPDATE":
		return "keycloak.user.updated", "user", targetID
	case "USER:DELETE":
		return "keycloak.user.deleted", "user", targetID
	case "USER:ACTION":
		return "keycloak.user.action", "user", targetID
	case "REALM_ROLE_MAPPING:CREATE":
		return "keycloak.user.role.granted", "user", targetID
	case "REALM_ROLE_MAPPING:DELETE":
		return "keycloak.user.role.revoked", "user", targetID
	case "CLIENT:UPDATE":
		return "keycloak.client.updated", "client", targetID
	case "REALM:UPDATE":
		return "keycloak.realm.updated", "realm", targetID
	default:
		return "keycloak.admin." + strings.ToLower(key), "realm", targetID
	}
}

func userEventPayload(ev keycloakWebhookRequest) map[string]any {
	p := map[string]any{
		"keycloak_event_type": ev.Type,
		"client_id":           ev.ClientID,
		"realm_id":            ev.RealmID,
		"user_id":             ev.UserID,
		"ip_address":          ev.IPAddress,
	}
	if ev.Error != "" {
		p["error"] = ev.Error
	}
	if sessionID, ok := ev.Details["sessionId"]; ok {
		p["session_id"] = sessionID
	}
	return p
}

func adminEventPayload(ev keycloakWebhookRequest) map[string]any {
	p := map[string]any{
		"resource_type":  ev.ResourceType,
		"operation_type": ev.OperationType,
		"resource_path":  ev.ResourcePath,
		"realm_id":       ev.RealmID,
		"auth_user_id":   ev.AuthDetails.UserID,
		"auth_client_id": ev.AuthDetails.ClientID,
		"ip_address":     ev.AuthDetails.IPAddress,
	}
	if ev.Error != "" {
		p["error"] = ev.Error
	}
	return p
}
