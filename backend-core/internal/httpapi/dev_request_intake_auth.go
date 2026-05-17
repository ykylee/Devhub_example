package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// IntakeTokenStore는 DREQ 외부 수신 토큰의 lookup/touch + admin CRUD 인터페이스.
// ADR-0012 §4.1.1 / §4.1.2 의 인증 검증 흐름 (Lookup / MarkUsed) + ADR-0014 의 admin
// CRUD (Create / List / Revoke) 를 store layer 에 위임. plain token 자체는 절대
// 저장하지 않으며 caller (handler) 가 SHA-256 hex 로 hash 한 뒤 Create 호출한다.
type IntakeTokenStore interface {
	LookupDevRequestIntakeToken(ctx context.Context, hashedToken string) (domain.DevRequestIntakeToken, error)
	MarkDevRequestIntakeTokenUsed(ctx context.Context, tokenID string) error
	CreateDevRequestIntakeToken(ctx context.Context, tok domain.DevRequestIntakeToken) (domain.DevRequestIntakeToken, error)
	ListDevRequestIntakeTokens(ctx context.Context) ([]domain.DevRequestIntakeToken, error)
	RevokeDevRequestIntakeToken(ctx context.Context, tokenID string) (domain.DevRequestIntakeToken, error)
	UpdateDevRequestIntakeTokenIPs(ctx context.Context, tokenID string, allowedIPs []string) (domain.DevRequestIntakeToken, error)
}

// 컨텍스트 키 — DREQ intake 인증 통과 시 토큰의 source_system 매핑값을 핸들러에 전달.
// body 의 self-claim 무시 (ADR-0012 §4.1.2 spoofing 방지).
const (
	ctxKeyDREQSourceSystem = "devhub_dreq_source_system"
	ctxKeyDREQClientLabel  = "devhub_dreq_client_label"
	ctxKeyDREQTokenID      = "devhub_dreq_token_id"
)

// requireIntakeToken은 DREQ 외부 수신 endpoint (POST /api/v1/dev-requests) 의
// 인증 middleware. ADR-0012 옵션 A (API 토큰 + IP allowlist) 채택.
//
// 검증 흐름 (ARCH-DREQ-03):
//  1. Authorization: Bearer <plain-token> 헤더 추출
//  2. SHA-256(plain-token) 계산 후 dev_request_intake_tokens.hashed_token lookup
//  3. revoked_at IS NOT NULL → 401
//  4. caller IP 가 row.allowed_ips CIDR 안에 있는지 검증 (allowed_ips 가 비어 있으면 거절)
//  5. 통과 시 source_system / client_label / token_id 를 컨텍스트에 set
//  6. best-effort 로 last_used_at 갱신
//  7. audit 은 caller(핸들러) 가 dev_request.intake_auth_succeeded/failed emit
//
// 본 middleware 는 v1 group 의 enforceRoutePermission 이전 단계에서 작동하며,
// routePermissionTable 은 본 endpoint 를 Bypass: true 로 매핑한다.
func (h Handler) requireIntakeToken(c *gin.Context) {
	tokenStore := h.cfg.DevRequestIntakeTokenStore
	if tokenStore == nil {
		// 운영 환경에서 store 가 nil 이면 인증 불가 — 503 으로 응답.
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "dev_request intake auth is not configured",
			"code":   "intake_auth_unavailable",
		})
		return
	}

	rawHeader := c.GetHeader("Authorization")
	token := extractBearerToken(rawHeader)
	if token == "" {
		h.recordAuditBestEffort(c, "dev_request.intake_auth_failed", "route", c.FullPath(), map[string]any{
			"reason":         "missing_bearer_header",
			"source_ip":      c.ClientIP(),
			"header_present": rawHeader != "",
		})
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthorized",
			"error":  "Authorization Bearer token is required",
			"code":   "auth_intake_token_missing",
		})
		return
	}

	hashed := hashIntakeToken(token)
	row, err := tokenStore.LookupDevRequestIntakeToken(c.Request.Context(), hashed)
	if errors.Is(err, store.ErrNotFound) {
		h.recordAuditBestEffort(c, "dev_request.intake_auth_failed", "route", c.FullPath(), map[string]any{
			"reason":         "token_not_found",
			"source_ip":      c.ClientIP(),
			"header_present": true,
			"token_prefix4":  tokenPrefix4(token),
		})
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthorized",
			"error":  "intake token is invalid",
			"code":   "auth_intake_token_invalid",
		})
		return
	}
	if err != nil {
		writeServerError(c, err, "dev_request.intake_auth.lookup")
		return
	}

	if !row.IsActive() {
		reason := "token_revoked"
		errCode := "auth_intake_token_revoked"
		errText := "intake token is revoked"

		if row.ExpiresAt != nil && time.Now().After(*row.ExpiresAt) {
			reason = "token_expired"
			errCode = "auth_intake_token_expired"
			errText = "intake token is expired"
		}

		h.recordAuditBestEffort(c, "dev_request.intake_auth_failed", "dev_request_intake_token", row.TokenID, map[string]any{
			"reason":       reason,
			"source_ip":    c.ClientIP(),
			"client_label": row.ClientLabel,
		})
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthorized",
			"error":  errText,
			"code":   errCode,
		})
		return
	}

	if !clientIPAllowed(c.ClientIP(), row.AllowedIPs) {
		h.recordAuditBestEffort(c, "dev_request.intake_auth_failed", "dev_request_intake_token", row.TokenID, map[string]any{
			"reason":       "ip_not_allowed",
			"source_ip":    c.ClientIP(),
			"client_label": row.ClientLabel,
		})
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "unauthorized",
			"error":  "client IP is not in allowlist",
			"code":   "auth_intake_ip_denied",
		})
		return
	}

	// 컨텍스트에 토큰 정보 set. handler 가 dev_request.received audit 시 client_label/token_id 활용.
	c.Set(ctxKeyDREQSourceSystem, row.SourceSystem)
	c.Set(ctxKeyDREQClientLabel, row.ClientLabel)
	c.Set(ctxKeyDREQTokenID, row.TokenID)

	// best-effort touch.
	_ = tokenStore.MarkDevRequestIntakeTokenUsed(c.Request.Context(), row.TokenID)

	h.recordAuditBestEffort(c, "dev_request.intake_auth_succeeded", "dev_request_intake_token", row.TokenID, map[string]any{
		"client_label": row.ClientLabel,
		"source_ip":    c.ClientIP(),
	})

	c.Next()
}

func extractBearerToken(authHeader string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return ""
	}
	return strings.TrimSpace(authHeader[len(prefix):])
}

func hashIntakeToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func tokenPrefix4(plain string) string {
	if len(plain) <= 4 {
		return plain
	}
	return plain[:4]
}

// clientIPAllowed는 callerIP 가 cidrs 의 어느 한 CIDR 에라도 속하는지 검증.
// cidrs 가 비어 있으면 모든 IP 거절 (안전한 deny by default).
func clientIPAllowed(callerIP string, cidrs []string) bool {
	if len(cidrs) == 0 {
		return false
	}
	parsed := net.ParseIP(callerIP)
	if parsed == nil {
		return false
	}
	for _, c := range cidrs {
		// CIDR 표기가 아닌 단일 IP 도 허용 (`192.0.2.1` 또는 `192.0.2.0/24`).
		if !strings.Contains(c, "/") {
			if parsed.Equal(net.ParseIP(c)) {
				return true
			}
			continue
		}
		_, network, err := net.ParseCIDR(c)
		if err != nil {
			continue
		}
		if network.Contains(parsed) {
			return true
		}
	}
	return false
}
