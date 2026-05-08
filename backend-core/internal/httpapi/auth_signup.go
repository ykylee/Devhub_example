package httpapi

import (
	"fmt"
	"net/http"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

type SignUpRequest struct {
	Name       string `json:"name" binding:"required"`
	SystemID   string `json:"system_id" binding:"required"`
	EmployeeID string `json:"employee_id" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

func (h Handler) authSignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload", "details": err.Error()})
		return
	}

	if h.cfg.HRDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "HR verification service is not configured"})
		return
	}

	// 1. Lookup HR DB
	email, userID, dept, err := h.cfg.HRDB.Lookup(c.Request.Context(), req.SystemID, req.EmployeeID, req.Name)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "identity verification failed", "details": "The provided information does not match our records."})
		return
	}

	// 2. Create Kratos Identity
	if h.cfg.KratosAdmin == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Kratos admin service is not configured"})
		return
	}

	kID, err := h.cfg.KratosAdmin.CreateIdentity(c.Request.Context(), email, req.Name, userID, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create identity", "details": err.Error()})
		return
	}

	// 3. Create DevHub User (if not exists)
	if h.cfg.OrganizationStore != nil {
		_, err := h.cfg.OrganizationStore.CreateUser(c.Request.Context(), domain.CreateUserInput{
			UserID:      userID,
			Email:       email,
			DisplayName: req.Name,
			Role:        domain.AppRole("developer"),
			Status:      domain.UserStatus("active"),
		})
		if err != nil {
			// Identity already created in Kratos, but DevHub user creation failed.
			// In production, we might want to roll back or retry.
			fmt.Printf("[SignUp] DevHub user creation failed for %s: %v\n", userID, err)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Account created successfully. You can now sign in.",
		"user_id": userID,
		"kratos_id": kID,
		"department": dept,
	})
}
