package view

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type AuditStore interface {
	CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
}

type OrganizationStore interface {
	GetUser(ctx context.Context, userID string) (domain.AppUser, error)
	SetIdPSubject(ctx context.Context, userID string, identityID string) error
	UpdateUser(ctx context.Context, login string, input domain.UpdateUserInput) (domain.AppUser, error)
}

type IdentityAdmin interface {
	FindIdentityByUserID(ctx context.Context, userID string) (string, error)
	LogoutUserSession(ctx context.Context, identityID string) error
}

type RealtimeTicketStore interface {
	Consume(ctx context.Context, ticket string) (store.RealtimeTicket, bool, error)
}

type AuthConfig struct {
	AuthDevFallback       bool
	RealtimeTickets       RealtimeTicketStore
	BearerTokenVerifier   BearerTokenVerifier
	OrganizationStore     OrganizationStore
	IdentityAdmin         IdentityAdmin
	OIDCLogoutClient      OIDCLogoutClient
	AuditStore            AuditStore
	OnboardingGateEnabled bool
}

type AuthHandler struct {
	cfg AuthConfig
}

func NewAuthHandler(cfg AuthConfig) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

func (h *AuthHandler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
	if h.cfg.AuditStore == nil {
		return domain.AuditLog{}
	}
	actor := httphelp.RequestActor(c)
	if payload == nil {
		payload = map[string]any{}
	}
	payload["actor_source"] = actor.Source
	logRow, err := h.cfg.AuditStore.CreateAuditLog(c.Request.Context(), domain.AuditLog{
		ActorLogin: actor.Login,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Payload:    payload,
		SourceIP:   httphelp.ClientIPFrom(c),
		RequestID:  httphelp.RequestIDFrom(c),
		SourceType: httphelp.SourceTypeFrom(c),
	})
	if err != nil {
		httphelp.LogRequest(c, "audit log persistence failed: action=%s target=%s/%s err=%v", action, targetType, targetID, err)
	}
	return logRow
}

func (h *AuthHandler) RequireOnboardingFlag(c *gin.Context) bool {
	if h.cfg.OnboardingGateEnabled {
		return true
	}
	c.JSON(http.StatusNotFound, gin.H{
		"status": "not_found",
		"code":   "onboarding_feature_disabled",
		"error":  "onboarding endpoints are disabled (DEVHUB_ONBOARDING_GATE_ENABLED=false)",
	})
	return false
}

type appointmentResponse struct {
	UnitID          string `json:"unit_id"`
	AppointmentRole string `json:"appointment_role"`
}

type appUserResponse struct {
	ID                    int64                 `json:"id"`
	UserID                string                `json:"user_id"`
	Email                 string                `json:"email"`
	DisplayName           string                `json:"display_name"`
	Role                  string                `json:"role"`
	Status                string                `json:"status"`
	Type                  string                `json:"type"`
	PrimaryUnitID         string                `json:"primary_unit_id,omitempty"`
	CurrentUnitID         string                `json:"current_unit_id,omitempty"`
	IsSeconded            bool                  `json:"is_seconded"`
	JoinedAt              time.Time             `json:"joined_at"`
	Appointments          []appointmentResponse `json:"appointments"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
	OnboardingCompletedAt *time.Time            `json:"onboarding_completed_at,omitempty"`
	ReviewStatus          string                `json:"review_status,omitempty"`
}

func appUserFromDomain(user domain.AppUser) appUserResponse {
	appointments := make([]appointmentResponse, 0, len(user.Appointments))
	for _, appointment := range user.Appointments {
		appointments = append(appointments, appointmentResponse{
			UnitID:          appointment.UnitID,
			AppointmentRole: string(appointment.AppointmentRole),
		})
	}
	return appUserResponse{
		ID:                    user.ID,
		UserID:                user.UserID,
		Email:                 user.Email,
		DisplayName:           user.DisplayName,
		Role:                  string(user.Role),
		Status:                string(user.Status),
		Type:                  string(user.Type),
		PrimaryUnitID:         user.PrimaryUnitID,
		CurrentUnitID:         user.CurrentUnitID,
		IsSeconded:            user.IsSeconded,
		JoinedAt:              user.JoinedAt,
		Appointments:          appointments,
		CreatedAt:             user.CreatedAt,
		UpdatedAt:             user.UpdatedAt,
		OnboardingCompletedAt: user.OnboardingCompletedAt,
		ReviewStatus:          user.ReviewStatus,
	}
}

func addAuditMeta(resp gin.H, log domain.AuditLog) {
	if log.AuditID != "" {
		resp["audit_log_id"] = log.AuditID
	}
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
}

// OIDCLogoutClient — sprint mvs/work_260608-i-488-sign-out (N-8). Keycloak
// OIDC /protocol/openid-connect/logout endpoint wrapper. IdentityAdmin 과
// 분리: admin endpoint 가 아닌 user-facing OIDC endpoint. 인터페이스는
// 최소 단일 메서드 (test fixture 주입용).
type OIDCLogoutClient interface {
	OIDCLogout(ctx context.Context, refreshToken string) error
}

// Logout — issue #488 spec 정합. POST /api/v1/auth/logout 의 v1.0 contract:
//
//   - Authorization: Bearer header 의 access token 인증 (route middleware 가
//     사전 검증, 미인증 시 401). 본 handler 는 인증된 요청만 도달.
//   - request body: { refresh_token?, id_token? } (둘 다 optional — OIDC
//     logout 은 refresh_token 이 primary identifier).
//   - 동작: server-managed cookie 4종 clear + Keycloak OIDC logout endpoint
//     (refresh_token 폐기) + audit emit. keycloak_revoke_status (ok /
//     unreachable / invalid) 를 audit payload 에 동봉.
//   - 응답: 204 No Content (성공, idempotent). Keycloak unreachable 시 502.
//     본 handler 의 401 경로는 route middleware 가 이미 enforce.
//   - idempotency: 같은 refresh_token 으로 두번 호출해도 두번 다 204
//     (OIDCLogout 가 4xx 를 nil 로 정규화).
func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if c.Request.Body != nil {
		_ = json.NewDecoder(c.Request.Body).Decode(&req)
	}

	actor := httphelp.RequestActor(c)

	// server-managed cookies 4종 clear (frontend 가 cookie-based 일 때 보조).
	// spec 에 명시 없으나 기존 동작 유지 (frontend 영향 0).
	for _, name := range []string{
		"devhub_session",
		"devhub_access_token",
		"devhub_refresh_token",
		"devhub_id_token",
	} {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	// OIDC logout — refresh_token 이 있을 때만 Keycloak endpoint 호출.
	revokeStatus := "ok"
	if h.cfg.OIDCLogoutClient != nil && strings.TrimSpace(req.RefreshToken) != "" {
		if err := h.cfg.OIDCLogoutClient.OIDCLogout(c.Request.Context(), req.RefreshToken); err != nil {
			revokeStatus = "unreachable"
			// 502 즉시 반환 (issue #488 spec 권장: 정합 우선).
			h.recordAuditBestEffort(c, "auth.logout", "auth", "current_session", map[string]any{
				"actor":                 actor.Login,
				"refresh_token_present": req.RefreshToken != "",
				"id_token_present":      req.IdToken != "",
				"revoke_status":         revokeStatus,
			})
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
				"status": "failed",
				"error":  "keycloak unreachable",
				"code":   "auth_logout.keycloak_unreachable",
			})
			return
		}
	} else if strings.TrimSpace(req.RefreshToken) == "" {
		revokeStatus = "skipped_no_refresh_token"
	}

	h.recordAuditBestEffort(c, "auth.logout", "auth", "current_session", map[string]any{
		"actor":                 actor.Login,
		"refresh_token_present": req.RefreshToken != "",
		"id_token_present":      req.IdToken != "",
		"revoke_status":         revokeStatus,
	})

	// 204 No Content (spec).
	c.Status(http.StatusNoContent)
}
