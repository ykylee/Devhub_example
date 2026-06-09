package httphelp

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var ErrIdentityNotFound = errors.New("identity not found")

// EnvelopeErrorResponse builds the canonical backend envelope error body.
// Per docs/backend_api_contract.md §1, error responses are `{ status: "error", error: { code, message, details? } }`.
func EnvelopeErrorResponse(code, message string) gin.H {
	return gin.H{
		"status": "error",
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
}

// EnvelopeErrorWithDetails builds an error envelope with structured details.
func EnvelopeErrorWithDetails(code, message string, details map[string]any) gin.H {
	body := gin.H{
		"status": "error",
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
	if details != nil {
		body["error"].(gin.H)["details"] = details
	}
	return body
}

func ParseBoundedInt(value string, minValue, maxValue int) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed < minValue || parsed > maxValue {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}

// WriteServerError logs the underlying error with operation context and returns
// a generic 500 response so internal details (DB schema, SQL fragments) do not
// leak to clients. Use only for unexpected server-side failures; client-visible
// errors (400/4xx) should keep their specific messages.
func WriteServerError(c *gin.Context, err error, op string) {
	requestID := RequestIDFrom(c)
	if requestID == "" {
		requestID = "-"
	}
	log.Printf("server error: op=%s request_id=%s err=%v", op, requestID, err)
	c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "internal error"})
}

