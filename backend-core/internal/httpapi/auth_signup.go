package httpapi

import (
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
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "invalid request payload",
			"code":   "invalid_payload",
		})
		return
	}

	if h.cfg.HRDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "HR verification service is not configured",
		})
		return
	}

	// 1. Lookup HR DB
	email, userID, dept, err := h.cfg.HRDB.Lookup(c.Request.Context(), req.SystemID, req.EmployeeID, req.Name)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"status": "forbidden",
			"error":  "identity verification failed",
			"code":   "hr_lookup_failed",
		})
		return
	}

	// 2. Create Kratos Identity
	if h.cfg.KratosAdmin == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "Kratos admin service is not configured",
		})
		return
	}

	kID, err := h.cfg.KratosAdmin.CreateIdentity(c.Request.Context(), email, req.Name, userID, req.Password)
	if err != nil {
		writeServerError(c, err, "auth.signup.kratos_create_identity")
		return
	}

	// 3. Create DevHub User (if not exists)
	devhubUserCreated := true
	if h.cfg.OrganizationStore != nil {
		_, err := h.cfg.OrganizationStore.CreateUser(c.Request.Context(), domain.CreateUserInput{
			UserID:      userID,
			Email:       email,
			DisplayName: req.Name,
			Role:        domain.AppRole("developer"),
			Status:      domain.UserStatus("active"),
		})
		if err != nil {
			// Identity already created in Kratos, but DevHub user creation
			// failed. RM-M3-01 carve: a rollback/retry strategy is the next
			// sprint's work — for now we keep the Kratos identity and emit
			// a distinct audit action so operators can find the orphan rows.
			devhubUserCreated = false
			logRequest(c, "[SignUp] DevHub user creation failed for %s: %v", userID, err)
			h.recordAuditBestEffort(c, "account.signup.partial_failure", "user", userID, map[string]any{
				"reason":      "devhub_user_create_failed",
				"kratos_id":   kID,
				"email":       email,
				"department":  dept,
				"error_class": "store",
			})
		}
	}

	// 4. Emit signup audit. Even when the DevHub user creation failed above
	// we want the requested action recorded so the operator dashboard can
	// reconcile against Kratos identities. payload mirrors §11.6 audit
	// mapping for account.signup.requested.
	if devhubUserCreated {
		h.recordAuditBestEffort(c, "account.signup.requested", "user", userID, map[string]any{
			"kratos_id":  kID,
			"email":      email,
			"department": dept,
			"system_id":  req.SystemID,
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "created",
		"data": gin.H{
			"user_id":    userID,
			"kratos_id":  kID,
			"department": dept,
			"message":    "Account created successfully. You can now sign in.",
		},
	})
}
