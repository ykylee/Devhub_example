package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type authTokenRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
}

func (h Handler) authToken(c *gin.Context) {
	if h.cfg.HydraToken == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "auth token exchange is not configured (HydraToken)",
		})
		return
	}

	var req authTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "invalid json body"})
		return
	}

	req.Code = strings.TrimSpace(req.Code)
	req.CodeVerifier = strings.TrimSpace(req.CodeVerifier)
	req.RedirectURI = strings.TrimSpace(req.RedirectURI)
	req.ClientID = strings.TrimSpace(req.ClientID)
	if req.Code == "" || req.CodeVerifier == "" || req.RedirectURI == "" || req.ClientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "code, code_verifier, redirect_uri, client_id are required",
		})
		return
	}

	token, err := h.cfg.HydraToken.ExchangeAuthorizationCode(c.Request.Context(), HydraTokenExchangeRequest{
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		RedirectURI:  req.RedirectURI,
		ClientID:     req.ClientID,
	})
	if errors.Is(err, ErrHydraTokenInvalidGrant) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "unauthorized",
			"error":  "authorization code is invalid or expired",
			"code":   "authorization_code_invalid",
		})
		return
	}
	if err != nil {
		writeServerError(c, err, "auth.token.exchange")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": token})
}
