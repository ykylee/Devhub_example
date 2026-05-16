package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
	"time"
)

// DREQ intake token admin endpoints (sprint claude/work_260515-o, ADR-0014).
// system_admin only — `dev_request_intake_tokens` RBAC resource via routePermissionTable.
// 발급 응답에는 plain token 이 정확히 1회 노출되며 store 는 SHA-256(plain) 만 보관.
// accounts_admin password issuance (plain 1회 노출) 패턴과 정합.

// intakeTokenAdminCreateRequest is the body for POST /api/v1/dev-request-tokens.
type intakeTokenAdminCreateRequest struct {
	ClientLabel  string   `json:"client_label"`
	SourceSystem string   `json:"source_system"`
	AllowedIPs   []string `json:"allowed_ips"`
	ExpiresAt    *string  `json:"expires_at"` // RFC3339 string
}

// intakeTokenAdminUpdateIPsRequest is the body for PATCH /api/v1/dev-request-tokens/:token_id.
type intakeTokenAdminUpdateIPsRequest struct {
	AllowedIPs []string `json:"allowed_ips"`
}

// generatePlainIntakeToken returns a 32-byte base64url-encoded random token
// (43 chars without padding). 256 bits of entropy — adequate for a long-lived
// server-to-server credential and conservative against brute-force.
func generatePlainIntakeToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// validateAllowedIPs accepts a CIDR or a single IP per entry. Empty list is
// rejected — deny-by-default for the same reason clientIPAllowed denies an
// empty allowlist at auth time (ADR-0012 §4.1.2). Returns the canonicalized
// list (trimmed, deduped order-preserved).
func validateAllowedIPs(raw []string) ([]string, string) {
	out := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, entry := range raw {
		e := strings.TrimSpace(entry)
		if e == "" {
			continue
		}
		if !strings.Contains(e, "/") {
			if net.ParseIP(e) == nil {
				return nil, "allowed_ips[" + e + "] is not a valid IP"
			}
		} else {
			if _, _, err := net.ParseCIDR(e); err != nil {
				return nil, "allowed_ips[" + e + "] is not a valid CIDR"
			}
		}
		if seen[e] {
			continue
		}
		seen[e] = true
		out = append(out, e)
	}
	if len(out) == 0 {
		return nil, "allowed_ips must contain at least one CIDR or IP (deny-by-default)"
	}
	return out, ""
}

// intakeTokenResponse is the wire shape for admin endpoints. Excludes hashed_token —
// operators never need to see the hash (and it cannot be reversed anyway).
// plain_token is set ONLY on creation responses (1회 노출).
func intakeTokenResponse(tok domain.DevRequestIntakeToken, plain string) gin.H {
	resp := gin.H{
		"token_id":      tok.TokenID,
		"client_label":  tok.ClientLabel,
		"source_system": tok.SourceSystem,
		"allowed_ips":   tok.AllowedIPs,
		"created_at":    tok.CreatedAt,
		"created_by":    tok.CreatedBy,
		"last_used_at":  tok.LastUsedAt,
		"revoked_at":    tok.RevokedAt,
		"expires_at":    tok.ExpiresAt,
	}
	if plain != "" {
		// 1회 노출 — 클라이언트가 안전한 저장소에 옮긴 뒤 즉시 폐기해야 한다.
		resp["plain_token"] = plain
	}
	return resp
}

// POST /api/v1/dev-request-tokens
func (h Handler) createDevRequestIntakeToken(c *gin.Context) {
	if h.cfg.DevRequestIntakeTokenStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "dev_request intake token admin requires IntakeTokenStore",
		})
		return
	}
	var req intakeTokenAdminCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	clientLabel := strings.TrimSpace(req.ClientLabel)
	sourceSystem := strings.TrimSpace(req.SourceSystem)
	if clientLabel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "client_label is required"})
		return
	}
	if sourceSystem == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "source_system is required"})
		return
	}
	canonIPs, problem := validateAllowedIPs(req.AllowedIPs)
	if problem != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  problem,
			"code":   "invalid_allowed_ips",
		})
		return
	}

	plain, err := generatePlainIntakeToken()
	if err != nil {
		writeServerError(c, err, "dev_request_intake_tokens.generate")
		return
	}
	hashed := hashIntakeToken(plain)

	actorLoginVal, _ := c.Get("devhub_actor_login")
	actorLogin, _ := actorLoginVal.(string)
	if actorLogin == "" {
		actorLogin = "system"
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "expires_at must be RFC3339 string"})
			return
		}
		expiresAt = &parsed
	}

	created, err := h.cfg.DevRequestIntakeTokenStore.CreateDevRequestIntakeToken(c.Request.Context(), domain.DevRequestIntakeToken{
		ClientLabel:  clientLabel,
		HashedToken:  hashed,
		AllowedIPs:   canonIPs,
		SourceSystem: sourceSystem,
		CreatedBy:    actorLogin,
		ExpiresAt:    expiresAt,
	})
	if errors.Is(err, store.ErrConflict) {
		// hashed_token collision — extremely unlikely with 256-bit entropy, but caller can retry.
		c.JSON(http.StatusConflict, gin.H{
			"status": "conflict",
			"error":  "token generation collision; retry",
			"code":   "intake_token_collision",
		})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request_intake_tokens.create")
		return
	}

	h.recordAuditBestEffort(c, "dev_request_intake_token.issued", "dev_request_intake_token", created.TokenID, map[string]any{
		"client_label":  created.ClientLabel,
		"source_system": created.SourceSystem,
		"allowed_ips":   created.AllowedIPs,
		// plain token 은 audit 에 절대 기록 안 함. hashed 도 운영 가치 0 이므로 생략.
	})

	c.JSON(http.StatusCreated, gin.H{
		"status": "ok",
		"data":   intakeTokenResponse(created, plain),
	})
}

// GET /api/v1/dev-request-tokens
func (h Handler) listDevRequestIntakeTokens(c *gin.Context) {
	if h.cfg.DevRequestIntakeTokenStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "dev_request intake token admin requires IntakeTokenStore",
		})
		return
	}
	tokens, err := h.cfg.DevRequestIntakeTokenStore.ListDevRequestIntakeTokens(c.Request.Context())
	if err != nil {
		writeServerError(c, err, "dev_request_intake_tokens.list")
		return
	}
	resp := make([]gin.H, 0, len(tokens))
	for _, tok := range tokens {
		resp = append(resp, intakeTokenResponse(tok, ""))
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   resp,
		"meta": gin.H{
			"total": len(tokens),
		},
	})
}

// DELETE /api/v1/dev-request-tokens/:token_id
func (h Handler) revokeDevRequestIntakeToken(c *gin.Context) {
	if h.cfg.DevRequestIntakeTokenStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "dev_request intake token admin requires IntakeTokenStore",
		})
		return
	}
	tokenID := strings.TrimSpace(c.Param("token_id"))
	if tokenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "token_id is required"})
		return
	}
	revoked, err := h.cfg.DevRequestIntakeTokenStore.RevokeDevRequestIntakeToken(c.Request.Context(), tokenID)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "intake token not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request_intake_tokens.revoke")
		return
	}
	h.recordAuditBestEffort(c, "dev_request_intake_token.revoked", "dev_request_intake_token", revoked.TokenID, map[string]any{
		"client_label":  revoked.ClientLabel,
		"source_system": revoked.SourceSystem,
	})
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   intakeTokenResponse(revoked, ""),
	})
}

// PATCH /api/v1/dev-request-tokens/:token_id
func (h Handler) updateDevRequestIntakeTokenIPs(c *gin.Context) {
	if h.cfg.DevRequestIntakeTokenStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "dev_request intake token admin requires IntakeTokenStore",
		})
		return
	}
	tokenID := strings.TrimSpace(c.Param("token_id"))
	if tokenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "token_id is required"})
		return
	}
	var req intakeTokenAdminUpdateIPsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}
	canonIPs, problem := validateAllowedIPs(req.AllowedIPs)
	if problem != "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": problem, "code": "invalid_allowed_ips"})
		return
	}

	updated, err := h.cfg.DevRequestIntakeTokenStore.UpdateDevRequestIntakeTokenIPs(c.Request.Context(), tokenID, canonIPs)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "intake token not found"})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request_intake_tokens.update_ips")
		return
	}

	h.recordAuditBestEffort(c, "dev_request_intake_token.updated", "dev_request_intake_token", updated.TokenID, map[string]any{
		"client_label":  updated.ClientLabel,
		"source_system": updated.SourceSystem,
		"allowed_ips":   updated.AllowedIPs,
	})

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   intakeTokenResponse(updated, ""),
	})
}
