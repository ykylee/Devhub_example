package httpapi

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// authConsent handles the consent flow. For our first-party client, we always
// skip the consent UI and grant all requested scopes.
func (h Handler) authConsent(c *gin.Context) {
	challenge := c.Query("consent_challenge")
	if challenge == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "consent_challenge is required"})
		return
	}

	ctx := c.Request.Context()

	// 1. Fetch consent request details
	consentReq, err := h.cfg.HydraAdmin.GetConsentRequest(ctx, challenge)
	if err != nil {
		log.Printf("[authConsent] Failed to get consent request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "failed to fetch consent request"})
		return
	}

	// 2. Accept the consent request immediately (silent consent)
	redirectTo, err := h.cfg.HydraAdmin.AcceptConsentRequest(ctx, challenge, consentReq.RequestedScope, true, 3600)
	if err != nil {
		log.Printf("[authConsent] Failed to accept consent request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "failed to accept consent request"})
		return
	}

	// 3. Redirect the browser back to Hydra
	log.Printf("[authConsent] Consent accepted for subject %s, redirecting to: %s", consentReq.Subject, redirectTo)
	c.Redirect(http.StatusFound, redirectTo)
}
