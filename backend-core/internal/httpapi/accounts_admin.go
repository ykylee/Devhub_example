package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// adminCreateAccountRequest is the body for POST /api/v1/accounts. The
// system admin issues a credential for a (possibly new) DevHub user; the
// handler creates both the DevHub users row and the Kratos identity, then
// returns the temporary password exactly once.
type adminCreateAccountRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	DisplayName  string `json:"display_name" binding:"required"`
	Role         string `json:"role"`         // optional, defaults to developer
	TempPassword string `json:"temp_password"` // optional, generated when empty
}

type adminPasswordResetRequest struct {
	TempPassword string `json:"temp_password,omitempty"`
}

type adminUpdateAccountRequest struct {
	Status string `json:"status,omitempty"` // "active" | "disabled"
}

const minTempPasswordLength = 12

// generateTempPassword returns a base64url-encoded 18-byte token (24 chars).
// Sufficient entropy for one-shot temporary credentials and meets Kratos's
// default min_password_length=12.
func generateTempPassword() (string, error) {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (h Handler) createAccount(c *gin.Context) {
	if h.cfg.OrganizationStore == nil || h.cfg.KratosAdmin == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "account admin requires OrganizationStore + KratosAdmin",
		})
		return
	}

	var req adminCreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	req.UserID = strings.TrimSpace(req.UserID)
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "developer"
	}
	tempPassword := strings.TrimSpace(req.TempPassword)
	if tempPassword == "" {
		generated, err := generateTempPassword()
		if err != nil {
			writeServerError(c, err, "account.create.temp_password")
			return
		}
		tempPassword = generated
	}
	if len(tempPassword) < minTempPasswordLength {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "temp_password must be at least 12 characters",
		})
		return
	}

	ctx := c.Request.Context()

	// 1) DevHub users row (master). Failures here surface to the caller
	// before we touch Kratos so we never strand a Kratos identity without
	// a matching DevHub record.
	user, err := h.cfg.OrganizationStore.CreateUser(ctx, domain.CreateUserInput{
		UserID:      req.UserID,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Role:        domain.AppRole(role),
		Status:      domain.UserStatus("active"),
		Type:        domain.UserType("human"),
		JoinedAt:    time.Now().UTC(),
	})
	if err != nil {
		writeServerError(c, err, "account.create.devhub_user")
		return
	}

	// 2) Kratos identity. If this fails we leave the DevHub row in place
	// and return 500 — operator can either delete it or retry the call,
	// which CreateUser will reject as duplicate.
	identityID, err := h.cfg.KratosAdmin.CreateIdentity(ctx, req.Email, req.DisplayName, req.UserID, tempPassword)
	if err != nil {
		h.recordAuditBestEffort(c, "account.issue.kratos_failed", "user", req.UserID, map[string]any{
			"reason": err.Error(),
		})
		writeServerError(c, err, "account.create.kratos_identity")
		return
	}

	// 3) Eagerly cache the identity_id on the DevHub users row so subsequent
	// admin/self-service flows skip the /admin/identities page scan (L4-A).
	// Failure here is non-fatal: the lazy backfill path will catch it on the
	// next lookup, and surfacing 500 would leave the DevHub+Kratos pair in
	// the correct state but make the caller think creation failed.
	if cacheErr := h.cfg.OrganizationStore.SetKratosIdentityID(ctx, req.UserID, identityID); cacheErr != nil {
		log.Printf("[kratos-cache] eager backfill on account.create for %s skipped: %v", req.UserID, cacheErr)
	}

	h.recordAuditBestEffort(c, "account.issued", "user", req.UserID, map[string]any{
		"email":       req.Email,
		"role":        role,
		"identity_id": identityID,
	})

	c.JSON(http.StatusCreated, gin.H{
		"status": "ok",
		"data": gin.H{
			"user_id":       user.UserID,
			"email":         user.Email,
			"role":          string(user.Role),
			"identity_id":   identityID,
			"temp_password": tempPassword,
		},
	})
}

func (h Handler) resetAccountPassword(c *gin.Context) {
	if h.cfg.KratosAdmin == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "account admin requires KratosAdmin",
		})
		return
	}
	userID := strings.TrimSpace(c.Param("user_id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "user_id is required"})
		return
	}

	// Empty body is intentional: callers can ask the server to generate a
	// fresh temp password by POSTing nothing. ShouldBindJSON returns io.EOF
	// in that case, so we only reject parse errors that came from a
	// non-empty body.
	var req adminPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	tempPassword := strings.TrimSpace(req.TempPassword)
	if tempPassword == "" {
		generated, err := generateTempPassword()
		if err != nil {
			writeServerError(c, err, "account.reset.temp_password")
			return
		}
		tempPassword = generated
	}
	if len(tempPassword) < minTempPasswordLength {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "temp_password must be at least 12 characters",
		})
		return
	}

	ctx := c.Request.Context()
	identityID, err := h.resolveKratosIdentityID(ctx, userID)
	if errors.Is(err, ErrKratosIdentityNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "kratos identity not found for user_id"})
		return
	}
	if err != nil {
		writeServerError(c, err, "account.reset.find_identity")
		return
	}

	if err := h.cfg.KratosAdmin.UpdateIdentityPassword(ctx, identityID, tempPassword); err != nil {
		writeServerError(c, err, "account.reset.update_password")
		return
	}

	h.recordAuditBestEffort(c, "account.password_force_reset", "user", userID, map[string]any{
		"identity_id": identityID,
	})

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"user_id":       userID,
			"identity_id":   identityID,
			"temp_password": tempPassword,
		},
	})
}

func (h Handler) updateAccountStatus(c *gin.Context) {
	if h.cfg.OrganizationStore == nil || h.cfg.KratosAdmin == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "account admin requires OrganizationStore + KratosAdmin",
		})
		return
	}
	userID := strings.TrimSpace(c.Param("user_id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "user_id is required"})
		return
	}

	var req adminUpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	status := strings.TrimSpace(req.Status)
	if status != "active" && status != "disabled" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "status must be 'active' or 'disabled'"})
		return
	}

	ctx := c.Request.Context()
	identityID, err := h.resolveKratosIdentityID(ctx, userID)
	if errors.Is(err, ErrKratosIdentityNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "kratos identity not found for user_id"})
		return
	}
	if err != nil {
		writeServerError(c, err, "account.status.find_identity")
		return
	}

	active := status == "active"
	if err := h.cfg.KratosAdmin.SetIdentityState(ctx, identityID, active); err != nil {
		writeServerError(c, err, "account.status.kratos_state")
		return
	}

	// Map the public-facing "disabled" label to the persisted status
	// constant. domain.UserStatusDeactivated is the value the DB
	// users_status_check constraint actually accepts; using "disabled"
	// would surface as a 500 after the Kratos state flip succeeds and
	// trigger the rollback path on every disable request.
	devhubStatus := domain.UserStatusActive
	if !active {
		devhubStatus = domain.UserStatusDeactivated
	}
	if _, err := h.cfg.OrganizationStore.UpdateUser(ctx, userID, domain.UpdateUserInput{Status: &devhubStatus}); err != nil {
		// Roll back the Kratos state change so the two stores stay in sync.
		_ = h.cfg.KratosAdmin.SetIdentityState(ctx, identityID, !active)
		writeServerError(c, err, "account.status.devhub_user")
		return
	}

	auditAction := "account.disabled"
	if active {
		auditAction = "account.enabled"
	}
	h.recordAuditBestEffort(c, auditAction, "user", userID, map[string]any{
		"identity_id": identityID,
	})

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   gin.H{"user_id": userID, "status": status},
	})
}

func (h Handler) deleteAccount(c *gin.Context) {
	if h.cfg.OrganizationStore == nil || h.cfg.KratosAdmin == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "account admin requires OrganizationStore + KratosAdmin",
		})
		return
	}
	userID := strings.TrimSpace(c.Param("user_id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "user_id is required"})
		return
	}

	ctx := c.Request.Context()
	identityID, err := h.resolveKratosIdentityID(ctx, userID)
	// Missing Kratos identity is non-fatal — proceed with DevHub delete so an
	// orphaned users row can still be cleaned up.
	if err != nil && !errors.Is(err, ErrKratosIdentityNotFound) {
		writeServerError(c, err, "account.delete.find_identity")
		return
	}

	if identityID != "" {
		if err := h.cfg.KratosAdmin.DeleteIdentity(ctx, identityID); err != nil && !errors.Is(err, ErrKratosIdentityNotFound) {
			writeServerError(c, err, "account.delete.kratos")
			return
		}
	}

	if err := h.cfg.OrganizationStore.DeleteUser(ctx, userID); err != nil {
		writeServerError(c, err, "account.delete.devhub_user")
		return
	}

	h.recordAuditBestEffort(c, "account.deleted", "user", userID, map[string]any{
		"identity_id": identityID,
	})

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   gin.H{"user_id": userID},
	})
}
