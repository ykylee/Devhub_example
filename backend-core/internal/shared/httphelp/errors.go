package httphelp

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var ErrIdentityNotFound = errors.New("identity not found")

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

