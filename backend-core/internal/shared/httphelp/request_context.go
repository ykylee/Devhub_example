package httphelp

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

// Gin context keys.
const (
	CtxKeyRequestID  = "devhub_request_id"
	CtxKeySourceType = "devhub_source_type"
)

type requestIDCtxKey struct{}

var callerRequestIDPattern = regexp.MustCompile(`^[A-Za-z0-9_\-]{1,128}$`)

func GenerateRequestID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "req_unknown"
	}
	return "req_" + hex.EncodeToString(buf[:])
}

func ValidateCallerRequestID(raw string) string {
	id := strings.TrimSpace(raw)
	if id == "" {
		return ""
	}
	if !callerRequestIDPattern.MatchString(id) {
		return ""
	}
	return id
}

// RequireRequestID is a middleware that stamps every request with a request id.
func RequireRequestID(c *gin.Context) {
	id := ValidateCallerRequestID(c.GetHeader("X-Request-ID"))
	if id == "" {
		id = GenerateRequestID()
	}
	c.Set(CtxKeyRequestID, id)
	if c.Request != nil {
		ctx := context.WithValue(c.Request.Context(), requestIDCtxKey{}, id)
		c.Request = c.Request.WithContext(ctx)
	}
	c.Header("X-Request-ID", id)
	c.Next()
}

func RequestIDFrom(c *gin.Context) string {
	if value, ok := c.Get(CtxKeyRequestID); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(requestIDCtxKey{}).(string); ok {
		return v
	}
	return ""
}

func SourceTypeFrom(c *gin.Context) domain.AuditSourceType {
	if value, ok := c.Get(CtxKeySourceType); ok {
		if s, ok := value.(domain.AuditSourceType); ok && s != "" {
			return s
		}
		if s, ok := value.(string); ok && s != "" {
			return domain.AuditSourceType(s)
		}
	}
	return ""
}

func ClientIPFrom(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return c.ClientIP()
}

func LogRequest(c *gin.Context, format string, args ...any) {
	rid := RequestIDFrom(c)
	if rid == "" {
		log.Printf(format, args...)
		return
	}
	log.Printf("request_id=%s "+format, append([]any{rid}, args...)...)
}

func LogRequestCtx(ctx context.Context, format string, args ...any) {
	rid := RequestIDFromContext(ctx)
	if rid == "" {
		log.Printf(format, args...)
		return
	}
	log.Printf("request_id=%s "+format, append([]any{rid}, args...)...)
}

type RequestActorInfo struct {
	Login  string
	Source string
}

func RequestActor(c *gin.Context) RequestActorInfo {
	if value, ok := c.Get("devhub_actor_login"); ok {
		if actor, ok := value.(string); ok {
			actor = strings.TrimSpace(actor)
			if actor != "" {
				return RequestActorInfo{Login: actor, Source: "authenticated_context"}
			}
		}
	}
	return RequestActorInfo{Login: "system", Source: "system_fallback"}
}

func DevFallbackEnabled(c *gin.Context) bool {
	value, ok := c.Get("devhub_auth_dev_fallback")
	if !ok {
		return false
	}
	enabled, ok := value.(bool)
	return ok && enabled
}

