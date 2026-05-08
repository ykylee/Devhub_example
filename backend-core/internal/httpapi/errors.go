package httpapi

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// writeServerError logs the underlying error with operation context and returns
// a generic 500 response so internal details (DB schema, SQL fragments) do not
// leak to clients. Use only for unexpected server-side failures; client-visible
// errors (400/4xx) should keep their specific messages.
func writeServerError(c *gin.Context, err error, op string) {
	log.Printf("server error: op=%s err=%v", op, err)
	c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "internal error"})
}
