package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"regexp"
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

// requestIDCtxKey is the typed context.Context key that mirrors the gin
// store. Background work that only sees a context.Context (Kratos / Hydra
// HTTP clients, store helpers, future workers) can pull the request id
// from here, so log lines they emit stay correlated with the originating
// HTTP request.
//
// Typed-key pattern (vexillographer/blank-struct) avoids collision with any
// other package that happens to use a string-typed value with the same name.
type requestIDCtxKey struct{}

// callerRequestIDPattern bounds the shape of caller-supplied X-Request-ID
// values: 1..128 chars of [A-Za-z0-9_-]. Anything outside this set is
// rejected and the middleware falls back to a server-generated id.
//
// Rationale (work_260513-e A1, surfaced by work_260512-j): net/http already
// strips CR/LF from inbound header values so the format-safety regression
// described in PR-D follow-up cannot recur, but audit_logs.request_id is
// indexed/grepped downstream and arbitrary unicode / shell-metacharacters
// poison those flows. Restricting to the same charset as the generated
// `req_<hex>` keeps the column homogeneous and printable.
var callerRequestIDPattern = regexp.MustCompile(`^[A-Za-z0-9_\-]{1,128}$`)

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

// validateCallerRequestID returns the trimmed value when it matches
// callerRequestIDPattern, or the empty string when it does not. Callers
// treat an empty return as "fall back to server-generated id".
func validateCallerRequestID(raw string) string {
	id := strings.TrimSpace(raw)
	if id == "" {
		return ""
	}
	if !callerRequestIDPattern.MatchString(id) {
		return ""
	}
	return id
}

// requireRequestID stamps every /api/v1/* request with a request id, exposes
// it on the response (X-Request-ID) for client-side correlation, and stashes
// it on gin.Context AND the underlying request context so audit + error
// helpers + downstream client/helper code (Kratos / Hydra HTTP clients,
// background workers) can pick it up.
//
// If the inbound request already carries a well-formed X-Request-ID header
// (1..128 chars of [A-Za-z0-9_-]) we honour it as-is. Otherwise — empty or
// malformed — we generate one with the req_ prefix per DEC-3=B.
func (h Handler) requireRequestID(c *gin.Context) {
	id := validateCallerRequestID(c.GetHeader("X-Request-ID"))
	if id == "" {
		id = generateRequestID()
	}
	c.Set(ctxKeyRequestID, id)
	if c.Request != nil {
		ctx := context.WithValue(c.Request.Context(), requestIDCtxKey{}, id)
		c.Request = c.Request.WithContext(ctx)
	}
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

// requestIDFromContext is the ctx-only counterpart of requestIDFrom. It is
// the entry point for code that does not hold a *gin.Context — Kratos /
// Hydra HTTP clients, store helpers, background workers — so their log
// lines stay correlated with the originating request.
func requestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(requestIDCtxKey{}).(string); ok {
		return v
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
//
// The request id is rendered through a %s verb (not string-concatenated
// into the format) so caller-supplied X-Request-ID values containing
// percent signs cannot turn into stray printf verbs.
func logRequest(c *gin.Context, format string, args ...any) {
	rid := requestIDFrom(c)
	if rid == "" {
		log.Printf(format, args...)
		return
	}
	log.Printf("request_id=%s "+format, append([]any{rid}, args...)...)
}

// logRequestCtx is the ctx-only counterpart of logRequest for client and
// helper code that only sees a context.Context. Format-safety guarantees
// are identical: the request id is rendered through a %s verb, never
// concatenated, so a caller-supplied id containing %v cannot turn into a
// printf verb that consumes args.
func logRequestCtx(ctx context.Context, format string, args ...any) {
	rid := requestIDFromContext(ctx)
	if rid == "" {
		log.Printf(format, args...)
		return
	}
	log.Printf("request_id=%s "+format, append([]any{rid}, args...)...)
}

