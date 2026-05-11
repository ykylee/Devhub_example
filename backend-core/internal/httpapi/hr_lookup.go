package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h Handler) hrLookup(c *gin.Context) {
	systemID := strings.TrimSpace(c.Query("system_id"))
	employeeID := strings.TrimSpace(c.Query("employee_id"))
	name := strings.TrimSpace(c.Query("name"))

	if systemID == "" && employeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "system_id or employee_id is required for lookup",
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

	email, userID, dept, err := h.cfg.HRDB.Lookup(c.Request.Context(), systemID, employeeID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "not_found",
			"error":  "person not found in HR DB",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"email":        email,
			"user_id":      userID,
			"department":   dept,
			"display_name": name, // fallback if not in mock yet
		},
	})
}
