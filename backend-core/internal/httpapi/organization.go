package httpapi

import (
	"context"
	"errors"
	"fmt"
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
	CreateUser(context.Context, domain.CreateUserInput) (domain.AppUser, error)
	UpdateUser(context.Context, string, domain.UpdateUserInput) (domain.AppUser, error)
	DeleteUser(context.Context, string) error
	GetHierarchy(context.Context) (domain.Hierarchy, error)
	UpdateHierarchy(context.Context, domain.Hierarchy) error
	GetOrgUnit(context.Context, string) (domain.OrgUnit, error)
	CreateOrgUnit(context.Context, domain.CreateOrgUnitInput) (domain.OrgUnit, error)
	UpdateOrgUnit(context.Context, string, domain.UpdateOrgUnitInput) (domain.OrgUnit, error)
	DeleteOrgUnit(context.Context, string) error
	ListUnitMembers(context.Context, string) ([]domain.AppUser, error)
	ReplaceUnitMembers(context.Context, string, []string) error
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
	Type          string                `json:"type"`
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
		Type:          string(user.Type),
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
		writeServerError(c, err, "organization.list_users")
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
		writeServerError(c, err, "organization.get_user")
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
		writeServerError(c, err, "organization.get_hierarchy")
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
		writeServerError(c, err, "organization.list_unit_members")
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

type createUserRequest struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	Type          string `json:"type"`
	Password      string `json:"password,omitempty"`
	PrimaryUnitID string `json:"primary_unit_id"`
	CurrentUnitID string `json:"current_unit_id"`
	IsSeconded    bool   `json:"is_seconded"`
	JoinedAt      string `json:"joined_at"`
}

type updateUserRequest struct {
	Email         *string `json:"email"`
	DisplayName   *string `json:"display_name"`
	Role          *string `json:"role"`
	Status        *string `json:"status"`
	PrimaryUnitID *string `json:"primary_unit_id"`
	CurrentUnitID *string `json:"current_unit_id"`
	IsSeconded    *bool   `json:"is_seconded"`
	JoinedAt      *string `json:"joined_at"`
}

var validAppRoles = map[string]bool{
	"developer":    true,
	"manager":      true,
	"system_admin": true,
}

var validUserStatuses = map[string]bool{
	"active":      true,
	"pending":     true,
	"deactivated": true,
}

var validUserTypes = map[string]bool{
	"human":  true,
	"system": true,
}

var validUnitTypes = map[string]bool{
	"company":  true,
	"division": true,
	"team":     true,
	"group":    true,
	"part":     true,
}

func parseJoinedAt(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("joined_at must be RFC3339 or YYYY-MM-DD")
}

func writeStoreError(c *gin.Context, err error, op string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"status": "not_found", "error": err.Error()})
	case errors.Is(err, store.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"status": "conflict", "error": err.Error()})
	default:
		writeServerError(c, err, op)
	}
}

func (h Handler) createUser(c *gin.Context) {
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	req.UserID = strings.TrimSpace(req.UserID)
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Role = strings.TrimSpace(req.Role)
	req.Status = strings.TrimSpace(req.Status)
	req.Type = strings.TrimSpace(req.Type)
	req.PrimaryUnitID = strings.TrimSpace(req.PrimaryUnitID)
	req.CurrentUnitID = strings.TrimSpace(req.CurrentUnitID)

	if req.UserID == "" || req.Email == "" || req.DisplayName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "user_id, email, and display_name are required"})
		return
	}
	if req.Type == "" {
		req.Type = "human"
	}
	if !validUserTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "type must be human or system"})
		return
	}
	if !validAppRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "role must be developer, manager, or system_admin"})
		return
	}
	if !validUserStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "status must be active, pending, or deactivated"})
		return
	}
	joinedAt, err := parseJoinedAt(req.JoinedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	// Optional: Create Kratos Identity if password is provided
	var kID string
	if req.Password != "" {
		if h.cfg.KratosAdmin == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Kratos admin service is not configured"})
			return
		}
		id, err := h.cfg.KratosAdmin.CreateIdentity(c.Request.Context(), req.Email, req.DisplayName, req.UserID, req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create identity", "details": err.Error()})
			return
		}
		kID = id
	}

	input := domain.CreateUserInput{
		UserID:        req.UserID,
		Email:         req.Email,
		DisplayName:   req.DisplayName,
		Role:          domain.AppRole(req.Role),
		Status:        domain.UserStatus(req.Status),
		Type:          domain.UserType(req.Type),
		PrimaryUnitID: req.PrimaryUnitID,
		CurrentUnitID: req.CurrentUnitID,
		IsSeconded:    req.IsSeconded,
		JoinedAt:      joinedAt,
	}

	user, err := h.cfg.OrganizationStore.CreateUser(c.Request.Context(), input)
	if err != nil {
		writeStoreError(c, err, "organization.create_user")
		return
	}
	auditLog := h.recordAuditBestEffort(c, "user.created", "user", user.UserID, map[string]any{
		"email":           user.Email,
		"role":            string(user.Role),
		"status":          string(user.Status),
		"type":            string(user.Type),
		"kratos_id":       kID,
		"primary_unit_id": user.PrimaryUnitID,
		"current_unit_id": user.CurrentUnitID,
	})

	response := gin.H{
		"status": "created",
		"data":   appUserFromDomain(user),
	}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusCreated, response)
}

func (h Handler) updateUser(c *gin.Context) {
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

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	input := domain.UpdateUserInput{
		IsSeconded: req.IsSeconded,
	}
	if req.Email != nil {
		trimmed := strings.TrimSpace(*req.Email)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "email cannot be empty"})
			return
		}
		input.Email = &trimmed
	}
	if req.DisplayName != nil {
		trimmed := strings.TrimSpace(*req.DisplayName)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "display_name cannot be empty"})
			return
		}
		input.DisplayName = &trimmed
	}
	if req.Role != nil {
		trimmed := strings.TrimSpace(*req.Role)
		if !validAppRoles[trimmed] {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "role must be developer, manager, or system_admin"})
			return
		}
		role := domain.AppRole(trimmed)
		input.Role = &role
	}
	if req.Status != nil {
		trimmed := strings.TrimSpace(*req.Status)
		if !validUserStatuses[trimmed] {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "status must be active, pending, or deactivated"})
			return
		}
		status := domain.UserStatus(trimmed)
		input.Status = &status
	}
	if req.PrimaryUnitID != nil {
		trimmed := strings.TrimSpace(*req.PrimaryUnitID)
		input.PrimaryUnitID = &trimmed
	}
	if req.CurrentUnitID != nil {
		trimmed := strings.TrimSpace(*req.CurrentUnitID)
		input.CurrentUnitID = &trimmed
	}
	if req.JoinedAt != nil {
		joinedAt, err := parseJoinedAt(*req.JoinedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
			return
		}
		input.JoinedAt = &joinedAt
	}

	user, err := h.cfg.OrganizationStore.UpdateUser(c.Request.Context(), userID, input)
	if err != nil {
		writeStoreError(c, err, "organization.update_user")
		return
	}
	auditLog := h.recordAuditBestEffort(c, "user.updated", "user", user.UserID, userUpdateAuditPayload(input))
	response := gin.H{
		"status": "ok",
		"data":   appUserFromDomain(user),
	}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusOK, response)
}

func (h Handler) deleteUser(c *gin.Context) {
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

	if err := h.cfg.OrganizationStore.DeleteUser(c.Request.Context(), userID); err != nil {
		writeStoreError(c, err, "organization.delete_user")
		return
	}
	auditLog := h.recordAuditBestEffort(c, "user.deleted", "user", userID, nil)
	response := gin.H{"status": "deleted", "data": gin.H{"user_id": userID}}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusOK, response)
}

type createOrgUnitRequest struct {
	UnitID       string `json:"unit_id"`
	ParentUnitID string `json:"parent_unit_id"`
	UnitType     string `json:"unit_type"`
	Label        string `json:"label"`
	LeaderUserID string `json:"leader_user_id"`
	PositionX    int    `json:"position_x"`
	PositionY    int    `json:"position_y"`
}

type updateOrgUnitRequest struct {
	ParentUnitID *string `json:"parent_unit_id"`
	UnitType     *string `json:"unit_type"`
	Label        *string `json:"label"`
	LeaderUserID *string `json:"leader_user_id"`
	PositionX    *int    `json:"position_x"`
	PositionY    *int    `json:"position_y"`
}

func (h Handler) getOrgUnit(c *gin.Context) {
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

	unit, err := h.cfg.OrganizationStore.GetOrgUnit(c.Request.Context(), unitID)
	if err != nil {
		writeStoreError(c, err, "organization.get_org_unit")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": orgUnitFromDomain(unit)})
}

func (h Handler) createOrgUnit(c *gin.Context) {
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}
	var req createOrgUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	req.UnitID = strings.TrimSpace(req.UnitID)
	req.ParentUnitID = strings.TrimSpace(req.ParentUnitID)
	req.UnitType = strings.TrimSpace(req.UnitType)
	req.Label = strings.TrimSpace(req.Label)
	req.LeaderUserID = strings.TrimSpace(req.LeaderUserID)

	if req.UnitID == "" || req.Label == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "unit_id and label are required"})
		return
	}
	if !validUnitTypes[req.UnitType] {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "unit_type must be one of company/division/team/group/part"})
		return
	}
	if req.ParentUnitID == req.UnitID {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "parent_unit_id cannot equal unit_id"})
		return
	}

	input := domain.CreateOrgUnitInput{
		UnitID:       req.UnitID,
		ParentUnitID: req.ParentUnitID,
		UnitType:     domain.UnitType(req.UnitType),
		Label:        req.Label,
		LeaderUserID: req.LeaderUserID,
		PositionX:    req.PositionX,
		PositionY:    req.PositionY,
	}

	unit, err := h.cfg.OrganizationStore.CreateOrgUnit(c.Request.Context(), input)
	if err != nil {
		writeStoreError(c, err, "organization.create_org_unit")
		return
	}
	auditLog := h.recordAuditBestEffort(c, "org_unit.created", "org_unit", unit.UnitID, map[string]any{
		"parent_unit_id": unit.ParentUnitID,
		"unit_type":      string(unit.UnitType),
		"label":          unit.Label,
		"leader_user_id": unit.LeaderUserID,
	})

	response := gin.H{
		"status": "created",
		"data":   orgUnitFromDomain(unit),
	}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusCreated, response)
}

func (h Handler) updateOrgUnit(c *gin.Context) {
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

	var req updateOrgUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	input := domain.UpdateOrgUnitInput{
		PositionX: req.PositionX,
		PositionY: req.PositionY,
	}
	if req.ParentUnitID != nil {
		trimmed := strings.TrimSpace(*req.ParentUnitID)
		input.ParentUnitID = &trimmed
	}
	if req.UnitType != nil {
		trimmed := strings.TrimSpace(*req.UnitType)
		if !validUnitTypes[trimmed] {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "unit_type must be one of company/division/team/group/part"})
			return
		}
		ut := domain.UnitType(trimmed)
		input.UnitType = &ut
	}
	if req.Label != nil {
		trimmed := strings.TrimSpace(*req.Label)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "label cannot be empty"})
			return
		}
		input.Label = &trimmed
	}
	if req.LeaderUserID != nil {
		trimmed := strings.TrimSpace(*req.LeaderUserID)
		input.LeaderUserID = &trimmed
	}

	unit, err := h.cfg.OrganizationStore.UpdateOrgUnit(c.Request.Context(), unitID, input)
	if err != nil {
		writeStoreError(c, err, "organization.update_org_unit")
		return
	}
	auditLog := h.recordAuditBestEffort(c, "org_unit.updated", "org_unit", unit.UnitID, orgUnitUpdateAuditPayload(input))
	response := gin.H{"status": "ok", "data": orgUnitFromDomain(unit)}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusOK, response)
}

func (h Handler) deleteOrgUnit(c *gin.Context) {
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

	if err := h.cfg.OrganizationStore.DeleteOrgUnit(c.Request.Context(), unitID); err != nil {
		writeStoreError(c, err, "organization.delete_org_unit")
		return
	}
	auditLog := h.recordAuditBestEffort(c, "org_unit.deleted", "org_unit", unitID, nil)
	response := gin.H{"status": "deleted", "data": gin.H{"unit_id": unitID}}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusOK, response)
}

type replaceUnitMembersRequest struct {
	UserIDs []string `json:"user_ids"`
}

func (h Handler) replaceUnitMembers(c *gin.Context) {
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

	var req replaceUnitMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	cleaned := make([]string, 0, len(req.UserIDs))
	seen := make(map[string]bool, len(req.UserIDs))
	for _, raw := range req.UserIDs {
		userID := strings.TrimSpace(raw)
		if userID == "" || seen[userID] {
			continue
		}
		seen[userID] = true
		cleaned = append(cleaned, userID)
	}

	if err := h.cfg.OrganizationStore.ReplaceUnitMembers(c.Request.Context(), unitID, cleaned); err != nil {
		writeStoreError(c, err, "organization.replace_unit_members")
		return
	}
	auditLog := h.recordAuditBestEffort(c, "org_unit.members_replaced", "org_unit", unitID, map[string]any{
		"user_ids": cleaned,
	})

	members, err := h.cfg.OrganizationStore.ListUnitMembers(c.Request.Context(), unitID)
	if err != nil {
		writeServerError(c, err, "organization.replace_unit_members.list")
		return
	}
	data := make([]appUserResponse, 0, len(members))
	for _, user := range members {
		data = append(data, appUserFromDomain(user))
	}
	response := gin.H{
		"status": "ok",
		"data":   data,
		"meta": gin.H{
			"unit_id": unitID,
			"count":   len(data),
		},
	}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusOK, response)
}

func (h Handler) updateHierarchy(c *gin.Context) {
	if h.cfg.OrganizationStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "organization store is not configured",
		})
		return
	}

	var req struct {
		Nodes []struct {
			ID       string `json:"id"`
			Position struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"position"`
			Data struct {
				Label       string `json:"label"`
				Type        string `json:"type"`
				DirectCount int    `json:"direct_count"`
				TotalCount  int    `json:"total_count"`
			} `json:"data"`
		} `json:"nodes"`
		Edges []struct {
			Source string `json:"source"`
			Target string `json:"target"`
		} `json:"edges"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": err.Error()})
		return
	}

	// Map to domain.Hierarchy
	var hie domain.Hierarchy
	for _, n := range req.Nodes {
		hie.Units = append(hie.Units, domain.OrgUnit{
			UnitID:    n.ID,
			UnitType:  domain.UnitType(n.Data.Type),
			Label:     n.Data.Label,
			PositionX: int(n.Position.X),
			PositionY: int(n.Position.Y),
			// Counts are typically calculated or derived, but we'll pass them for now if the store uses them
			DirectCount: n.Data.DirectCount,
			TotalCount:  n.Data.TotalCount,
		})
	}
	for _, e := range req.Edges {
		hie.Edges = append(hie.Edges, domain.OrgEdge{
			SourceUnitID: e.Source,
			TargetUnitID: e.Target,
		})
	}

	if err := h.cfg.OrganizationStore.UpdateHierarchy(c.Request.Context(), hie); err != nil {
		writeStoreError(c, err, "organization.update_hierarchy")
		return
	}

	auditLog := h.recordAuditBestEffort(c, "org_hierarchy.updated", "org_hierarchy", "global", map[string]any{
		"unit_count": len(hie.Units),
		"edge_count": len(hie.Edges),
	})

	response := gin.H{"status": "ok", "message": "Hierarchy updated successfully"}
	addAuditMeta(response, auditLog)
	c.JSON(http.StatusOK, response)
}

func addAuditMeta(response gin.H, auditLog domain.AuditLog) {
	if auditLog.AuditID == "" {
		return
	}
	meta, ok := response["meta"].(gin.H)
	if !ok {
		meta = gin.H{}
		response["meta"] = meta
	}
	meta["audit_log_id"] = auditLog.AuditID
}

func userUpdateAuditPayload(input domain.UpdateUserInput) map[string]any {
	payload := map[string]any{}
	if input.Email != nil {
		payload["email"] = *input.Email
	}
	if input.DisplayName != nil {
		payload["display_name"] = *input.DisplayName
	}
	if input.Role != nil {
		payload["role"] = string(*input.Role)
	}
	if input.Status != nil {
		payload["status"] = string(*input.Status)
	}
	if input.PrimaryUnitID != nil {
		payload["primary_unit_id"] = *input.PrimaryUnitID
	}
	if input.CurrentUnitID != nil {
		payload["current_unit_id"] = *input.CurrentUnitID
	}
	if input.IsSeconded != nil {
		payload["is_seconded"] = *input.IsSeconded
	}
	if input.JoinedAt != nil {
		payload["joined_at"] = input.JoinedAt.Format(time.RFC3339)
	}
	return payload
}

func orgUnitUpdateAuditPayload(input domain.UpdateOrgUnitInput) map[string]any {
	payload := map[string]any{}
	if input.ParentUnitID != nil {
		payload["parent_unit_id"] = *input.ParentUnitID
	}
	if input.UnitType != nil {
		payload["unit_type"] = string(*input.UnitType)
	}
	if input.Label != nil {
		payload["label"] = *input.Label
	}
	if input.LeaderUserID != nil {
		payload["leader_user_id"] = *input.LeaderUserID
	}
	if input.PositionX != nil {
		payload["position_x"] = *input.PositionX
	}
	if input.PositionY != nil {
		payload["position_y"] = *input.PositionY
	}
	return payload
}
