package view

import (
	"context"
	"encoding/json"
	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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

// Logout clears any server-managed cookies and records an audit row.
// Token revocation is currently best-effort metadata only; the active session
// is still terminated by the IdP end-session redirect that the frontend drives.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if c.Request.Body != nil {
		_ = json.NewDecoder(c.Request.Body).Decode(&req)
	}

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

	auditLog := h.recordAuditBestEffort(c, "auth.logout", "auth", "current_session", map[string]any{
		"refresh_token_present": req.RefreshToken != "",
		"id_token_present":      req.IdToken != "",
		"revoke_attempted":      false,
	})

	resp := gin.H{
		"status": "ok",
		"data": gin.H{
			"revoked": false,
		},
	}
	addAuditMeta(resp, auditLog)
	c.JSON(http.StatusOK, resp)
}
