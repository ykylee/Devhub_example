package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type OrganizationStore interface {
	ListUsers(context.Context, domain.UserListOptions) ([]domain.AppUser, int, error)
	GetUser(context.Context, string) (domain.AppUser, error)
	GetHierarchy(context.Context) (domain.Hierarchy, error)
	ListUnitMembers(context.Context, string) ([]domain.AppUser, error)
}

type appointmentResponse struct {
	UnitID          string `json:"unit_id"`
	AppointmentRole string `json:"appointment_role"`
}

type appUserResponse struct {
	ID            int64                 `json:"id"`
	UserID        string                `json:"user_id"`
	Email         string                `json:"email"`
	DisplayName   string                `json:"display_name"`
	Role          string                `json:"role"`
	Status        string                `json:"status"`
	PrimaryUnitID string                `json:"primary_unit_id,omitempty"`
	CurrentUnitID string                `json:"current_unit_id,omitempty"`
	IsSeconded    bool                  `json:"is_seconded"`
	JoinedAt      time.Time             `json:"joined_at"`
	Appointments  []appointmentResponse `json:"appointments"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

type orgUnitResponse struct {
	ID           int64     `json:"id"`
	UnitID       string    `json:"unit_id"`
	ParentUnitID string    `json:"parent_unit_id,omitempty"`
	UnitType     string    `json:"unit_type"`
	Label        string    `json:"label"`
	LeaderUserID string    `json:"leader_user_id,omitempty"`
	PositionX    int       `json:"position_x"`
	PositionY    int       `json:"position_y"`
	DirectCount  int       `json:"direct_count"`
	TotalCount   int       `json:"total_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type orgEdgeResponse struct {
	SourceUnitID string `json:"source_unit_id"`
	TargetUnitID string `json:"target_unit_id"`
}

type hierarchyResponse struct {
	Units []orgUnitResponse `json:"units"`
	Edges []orgEdgeResponse `json:"edges"`
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
		ID:            user.ID,
		UserID:        user.UserID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Role:          string(user.Role),
		Status:        string(user.Status),
		PrimaryUnitID: user.PrimaryUnitID,
		CurrentUnitID: user.CurrentUnitID,
		IsSeconded:    user.IsSeconded,
		JoinedAt:      user.JoinedAt,
		Appointments:  appointments,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

func orgUnitFromDomain(unit domain.OrgUnit) orgUnitResponse {
	return orgUnitResponse{
		ID:           unit.ID,
		UnitID:       unit.UnitID,
		ParentUnitID: unit.ParentUnitID,
		UnitType:     string(unit.UnitType),
		Label:        unit.Label,
		LeaderUserID: unit.LeaderUserID,
		PositionX:    unit.PositionX,
		PositionY:    unit.PositionY,
		DirectCount:  unit.DirectCount,
		TotalCount:   unit.TotalCount,
		CreatedAt:    unit.CreatedAt,
		UpdatedAt:    unit.UpdatedAt,
	}
}

func (h Handler) listUsers(c *gin.Context) {
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}

	limit, err := parseBoundedInt(c.DefaultQuery("limit", "50"), 1, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be an integer between 1 and 100"})
		return
	}
	offset, err := parseBoundedInt(c.DefaultQuery("offset", "0"), 0, 100000)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "offset must be a non-negative integer"})
		return
	}
	opts := domain.UserListOptions{
		Limit:         limit,
		Offset:        offset,
		Role:          strings.TrimSpace(c.Query("role")),
		Status:        strings.TrimSpace(c.Query("status")),
		PrimaryUnitID: strings.TrimSpace(c.Query("primary_unit_id")),
	}

	users, total, err := h.cfg.OrganizationStore.ListUsers(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	data := make([]appUserResponse, 0, len(users))
	for _, user := range users {
		data = append(data, appUserFromDomain(user))
	}

	meta := gin.H{
		"limit":  opts.Limit,
		"offset": opts.Offset,
		"count":  len(data),
		"total":  total,
	}
	if opts.Role != "" {
		meta["role"] = opts.Role
	}
	if opts.Status != "" {
		meta["status"] = opts.Status
	}
	if opts.PrimaryUnitID != "" {
		meta["primary_unit_id"] = opts.PrimaryUnitID
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   data,
		"meta":   meta,
	})
}

func (h Handler) getUser(c *gin.Context) {
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}

	userID := strings.TrimSpace(c.Param("user_id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "user_id is required"})
		return
	}

	user, err := h.cfg.OrganizationStore.GetUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   appUserFromDomain(user),
	})
}

func (h Handler) getHierarchy(c *gin.Context) {
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}

	hierarchy, err := h.cfg.OrganizationStore.GetHierarchy(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	units := make([]orgUnitResponse, 0, len(hierarchy.Units))
	for _, unit := range hierarchy.Units {
		units = append(units, orgUnitFromDomain(unit))
	}
	edges := make([]orgEdgeResponse, 0, len(hierarchy.Edges))
	for _, edge := range hierarchy.Edges {
		edges = append(edges, orgEdgeResponse{
			SourceUnitID: edge.SourceUnitID,
			TargetUnitID: edge.TargetUnitID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": hierarchyResponse{
			Units: units,
			Edges: edges,
		},
	})
}

func (h Handler) listUnitMembers(c *gin.Context) {
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}

	unitID := strings.TrimSpace(c.Param("unit_id"))
	if unitID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "unit_id is required"})
		return
	}

	members, err := h.cfg.OrganizationStore.ListUnitMembers(c.Request.Context(), unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	data := make([]appUserResponse, 0, len(members))
	for _, user := range members {
		data = append(data, appUserFromDomain(user))
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   data,
		"meta": gin.H{
			"unit_id": unitID,
			"count":   len(data),
		},
	})
}
