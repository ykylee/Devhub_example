package view

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"
	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/gin-gonic/gin"
)


type AuditStore interface {
	CreateAuditLog(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
}


type PlatformStore interface {
	ListSCMProviders(ctx context.Context) ([]domain.SCMProvider, error)
	CreateProjectWithRepositoryPayload(ctx context.Context, p domain.Project, repoIDs []int64, payload *store.RepositoryCreatePayload) (domain.Project, error)
	CreatePlatform(ctx context.Context, app domain.Platform) (domain.Platform, error)
}

type DevRequestConfig struct {
	DevRequestStore            DevRequestStore
	DevRequestIntakeTokenStore IntakeTokenStore
	PlatformStore           PlatformStore
	AuditStore                 AuditStore
}

type DevRequestHandler struct {
	cfg DevRequestConfig
}

func NewDevRequestHandler(cfg DevRequestConfig) *DevRequestHandler {
	return &DevRequestHandler{cfg: cfg}
}

func (h *DevRequestHandler) recordAuditBestEffort(c *gin.Context, action, targetType, targetID string, payload map[string]any) domain.AuditLog {
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

func (h *DevRequestHandler) enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool {
	if httphelp.DevFallbackEnabled(c) {
		return true
	}

	loginVal, _ := c.Get("devhub_actor_login")
	roleVal, _ := c.Get("devhub_actor_role")
	actorLogin, _ := loginVal.(string)
	actorRole, _ := roleVal.(string)

	if actorRole == string(domain.AppRoleSystemAdmin) {
		return true
	}
	for _, allowed := range allowedRoles {
		if actorRole == allowed {
			return true
		}
	}
	if ownerUserID != "" && actorLogin == ownerUserID {
		return true
	}

	h.recordAuditBestEffort(c, "auth.row_denied", "route", c.FullPath(), map[string]any{
		"actor_role":    actorRole,
		"owner_user_id": ownerUserID,
		"denied_reason": "owner_mismatch",
	})
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"status": "forbidden",
		"error":  "owner_mismatch — row write requires owner or elevated role",
		"code":   "auth_row_denied",
	})
	return false
}

func (h *DevRequestHandler) PlatformStoreOrUnavailable(c *gin.Context) (PlatformStore, bool) {
	if h.cfg.PlatformStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "platform store is not configured"})
		return nil, false
	}
	return h.cfg.PlatformStore, true
}

// dev-request 에서 공유 필요 상수로 로컬 임베드
var platformKeyPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,10}$`)
var validPlatformVisibilities = map[string]bool{
	"public": true, "internal": true, "restricted": true,
}
var validPlatformStatuses = map[string]bool{
	"planning": true, "active": true, "on_hold": true, "closed": true, "archived": true,
}
var validPlatformRepoRoles = map[string]bool{
	"primary": true, "sub": true, "shared": true,
}
func parseDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func platformResponse(app domain.Platform) gin.H {
	return gin.H{
		"id":                  app.ID,
		"key":                 app.Key,
		"name":                app.Name,
		"description":         app.Description,
		"status":              string(app.Status),
		"visibility":          string(app.Visibility),
		"owner_user_id":       app.OwnerUserID,
		"leader_user_id":      app.LeaderUserID,
		"development_unit_id": app.DevelopmentUnitID,
		"start_date":          formatDatePtr(app.StartDate),
		"due_date":            formatDatePtr(app.DueDate),
		"created_at":          app.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":          app.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func projectResponse(p domain.Project) gin.H {
	repositoryID := any(nil)
	if p.RepositoryID > 0 {
		repositoryID = p.RepositoryID
	}
	return gin.H{
		"id":             p.ID,
		"platform_id": p.PlatformID,
		"repository_id":  repositoryID,
		"key":            p.Key,
		"name":           p.Name,
		"description":    p.Description,
		"status":         string(p.Status),
		"visibility":     string(p.Visibility),
		"owner_user_id":  p.OwnerUserID,
		"start_date":     formatDatePtr(p.StartDate),
		"due_date":       formatDatePtr(p.DueDate),
		"archived_at":    formatTimePtr(p.ArchivedAt),
		"created_at":     p.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":     p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func formatDatePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format("2006-01-02")
}

func formatTimePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}
