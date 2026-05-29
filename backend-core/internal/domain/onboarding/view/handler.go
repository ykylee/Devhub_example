package view

import (
	"context"
	"net/http"
	"time"
	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/gin-gonic/gin"
)


type AuditStore interface {
	CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
}


type OrganizationStore interface {
	GetUser(ctx context.Context, login string) (domain.AppUser, error)
	UpdateUser(ctx context.Context, login string, u domain.UpdateUserInput) (domain.AppUser, error)
	SubmitOnboarding(ctx context.Context, info domain.OnboardingSubmitInput) (domain.AppUser, error)
}

type OnboardingConfig struct {
	OrganizationStore     OrganizationStore
	OnboardingGateEnabled bool
	AuditStore            AuditStore
}

type OnboardingHandler struct {
	cfg OnboardingConfig
}

func NewOnboardingHandler(cfg OnboardingConfig) *OnboardingHandler {
	return &OnboardingHandler{cfg: cfg}
}

func (h *OnboardingHandler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
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

func (h *OnboardingHandler) RequireOnboardingFlag(c *gin.Context) bool {
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

