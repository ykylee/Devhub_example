package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"strings"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// gin.Context keys for the request-scoped operator-actor enrichment used by
// recordAudit + writeServerError. Readers should use requestIDFrom /
// sourceTypeFrom helpers below rather than touching the raw Get keys.
const (
	ctxKeyRequestID  = "devhub_request_id"
	ctxKeySourceType = "devhub_source_type"
)

// generateRequestID returns a 24-char hex token prefixed with "req_". DEC-3=B
// (work_26_05_11-c) — the prefix matches the realtime "evt_" convention so
// log greps can pick out request ids visually.
func generateRequestID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// rand.Read on the standard runtime never fails in practice. If it
		// does we still want a non-empty id so audit + logging can proceed;
		// "req_unknown" is preferable to dropping the trace key entirely.
		return "req_unknown"
	}
	return "req_" + hex.EncodeToString(buf[:])
}

// requireRequestID stamps every /api/v1/* request with a request id, exposes
// it on the response (X-Request-ID) for client-side correlation, and stashes
// it on gin.Context so audit + error helpers can pick it up.
//
// If the inbound request already carries an X-Request-ID header (e.g., the
// caller is a downstream service or a load-balancer that injects one), we
// honour it as-is. Otherwise we generate one with the req_ prefix per DEC-3=B.
func (h Handler) requireRequestID(c *gin.Context) {
	id := strings.TrimSpace(c.GetHeader("X-Request-ID"))
	if id == "" {
		id = generateRequestID()
	}
	c.Set(ctxKeyRequestID, id)
	c.Header("X-Request-ID", id)
	c.Next()
}

func requestIDFrom(c *gin.Context) string {
	if value, ok := c.Get(ctxKeyRequestID); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func sourceTypeFrom(c *gin.Context) domain.AuditSourceType {
	if value, ok := c.Get(ctxKeySourceType); ok {
		if s, ok := value.(domain.AuditSourceType); ok && s != "" {
			return s
		}
		if s, ok := value.(string); ok && s != "" {
			return domain.AuditSourceType(s)
		}
	}
	return ""
}

// clientIPFrom mirrors gin.Context.ClientIP but stays nil-safe so audit code
// can use it from background contexts.
func clientIPFrom(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return c.ClientIP()
}

// logRequest writes a log line prefixed with the gin.Context request id so
// grepping a request trace pulls every handler/middleware emission together.
// When the request id is unset (background calls, tests without
// requireRequestID) the line is emitted without a prefix — same shape as a
// plain log.Printf — so callers don't have to special-case.
//
// PR-D follow-up (work_260512-i sub-task #2). errors.go::writeServerError
// already embeds request_id inline ("op=%s request_id=%s err=%v"); the
// helper is for the handler/middleware log lines that are not gated by an
// op tag.
func logRequest(c *gin.Context, format string, args ...any) {
	rid := requestIDFrom(c)
	if rid == "" {
		log.Printf(format, args...)
		return
	}
	log.Printf("request_id="+rid+" "+format, args...)
}

