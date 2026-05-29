package httphelp

import (
	"bytes"
	"context"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

func TestGenerateRequestID_PrefixAndLength(t *testing.T) {
	id := GenerateRequestID()
	if !strings.HasPrefix(id, "req_") {
		t.Fatalf("expected req_ prefix, got %q", id)
	}
	if len(id) != len("req_")+24 {
		t.Fatalf("expected 4+24 length, got %d (%q)", len(id), id)
	}
	id2 := GenerateRequestID()
	if id == id2 {
		t.Fatal("two consecutive ids must differ")
	}
}

func TestValidateCallerRequestID(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"abc-123_XYZ", "abc-123_XYZ"},
		{"   trimmed   ", "trimmed"},
		{"", ""},
		{"   ", ""},
		{"has space", ""},
		{"has@symbol", ""},
		{strings.Repeat("a", 128), strings.Repeat("a", 128)},
		{strings.Repeat("a", 129), ""},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := ValidateCallerRequestID(tc.in)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRequireRequestID_UsesCallerHeaderWhenValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireRequestID)
	r.GET("/x", func(c *gin.Context) {
		c.String(200, RequestIDFrom(c))
	})

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("X-Request-ID", "caller-id-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Body.String() != "caller-id-123" {
		t.Fatalf("body=%q", rec.Body.String())
	}
	if rec.Header().Get("X-Request-ID") != "caller-id-123" {
		t.Fatalf("response header=%q", rec.Header().Get("X-Request-ID"))
	}
}

func TestRequireRequestID_GeneratesWhenHeaderInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireRequestID)
	r.GET("/x", func(c *gin.Context) {
		c.String(200, RequestIDFrom(c))
	})

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("X-Request-ID", "has space invalid")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.HasPrefix(body, "req_") {
		t.Fatalf("expected generated req_ prefix, got %q", body)
	}
	if rec.Header().Get("X-Request-ID") != body {
		t.Fatal("response X-Request-ID header must match body")
	}
}

func TestRequireRequestID_PropagatesToContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireRequestID)
	r.GET("/x", func(c *gin.Context) {
		if got := RequestIDFromContext(c.Request.Context()); got == "" {
			t.Error("request context must carry request_id")
		}
		c.Status(200)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
}

func TestRequestIDFrom_MissingReturnsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if got := RequestIDFrom(c); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	c.Set(CtxKeyRequestID, 12345) // wrong type — should ignore
	if got := RequestIDFrom(c); got != "" {
		t.Fatalf("expected empty on non-string, got %q", got)
	}
}

func TestRequestIDFromContext_NilOrMissing(t *testing.T) {
	if got := RequestIDFromContext(nil); got != "" {
		t.Fatalf("nil ctx: got %q", got)
	}
	if got := RequestIDFromContext(context.Background()); got != "" {
		t.Fatalf("empty ctx: got %q", got)
	}
	ctx := context.WithValue(context.Background(), requestIDCtxKey{}, "stamped")
	if got := RequestIDFromContext(ctx); got != "stamped" {
		t.Fatalf("got %q, want stamped", got)
	}
	ctxWrong := context.WithValue(context.Background(), requestIDCtxKey{}, 42)
	if got := RequestIDFromContext(ctxWrong); got != "" {
		t.Fatalf("non-string: got %q", got)
	}
}

func TestSourceTypeFrom_AllBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing returns empty", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		if got := SourceTypeFrom(c); got != "" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("AuditSourceType typed value returned directly", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(CtxKeySourceType, domain.AuditSourceOIDC)
		if got := SourceTypeFrom(c); got != domain.AuditSourceOIDC {
			t.Fatalf("got %q, want %q", got, domain.AuditSourceOIDC)
		}
	})

	t.Run("empty typed value falls through to empty", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(CtxKeySourceType, domain.AuditSourceType(""))
		if got := SourceTypeFrom(c); got != "" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("plain string cast to AuditSourceType", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(CtxKeySourceType, "system")
		if got := SourceTypeFrom(c); got != domain.AuditSourceType("system") {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("non-string non-AuditSourceType returns empty", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(CtxKeySourceType, 12345)
		if got := SourceTypeFrom(c); got != "" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestClientIPFrom(t *testing.T) {
	t.Run("nil context returns empty", func(t *testing.T) {
		if got := ClientIPFrom(nil); got != "" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("nil request returns empty", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		// Request is nil by default in CreateTestContext
		if got := ClientIPFrom(c); got != "" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("delegates to gin ClientIP", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/x", nil)
		c.Request.RemoteAddr = "10.0.0.5:1234"
		if got := ClientIPFrom(c); got == "" {
			t.Fatal("expected non-empty from gin ClientIP")
		}
	})
}

func TestLogRequest(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(originalOutput) })

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	t.Run("without request id", func(t *testing.T) {
		buf.Reset()
		LogRequest(c, "hello %s", "world")
		if !strings.Contains(buf.String(), "hello world") {
			t.Fatalf("got %q", buf.String())
		}
		if strings.Contains(buf.String(), "request_id=") {
			t.Fatalf("must not include request_id, got %q", buf.String())
		}
	})

	t.Run("with request id prepends", func(t *testing.T) {
		buf.Reset()
		c.Set(CtxKeyRequestID, "req_abc")
		LogRequest(c, "msg %d", 42)
		if !strings.Contains(buf.String(), "request_id=req_abc msg 42") {
			t.Fatalf("got %q", buf.String())
		}
	})
}

func TestLogRequestCtx(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(originalOutput) })

	t.Run("without request id in ctx", func(t *testing.T) {
		buf.Reset()
		LogRequestCtx(context.Background(), "plain %s", "msg")
		if !strings.Contains(buf.String(), "plain msg") {
			t.Fatalf("got %q", buf.String())
		}
		if strings.Contains(buf.String(), "request_id=") {
			t.Fatalf("must not include request_id, got %q", buf.String())
		}
	})

	t.Run("with request id in ctx", func(t *testing.T) {
		buf.Reset()
		ctx := context.WithValue(context.Background(), requestIDCtxKey{}, "req_ctx")
		LogRequestCtx(ctx, "msg=%d", 7)
		if !strings.Contains(buf.String(), "request_id=req_ctx msg=7") {
			t.Fatalf("got %q", buf.String())
		}
	})
}

func TestRequestActor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("no actor falls back to system", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		actor := RequestActor(c)
		if actor.Login != "system" || actor.Source != "system_fallback" {
			t.Fatalf("got %+v", actor)
		}
	})

	t.Run("authenticated actor returned", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("devhub_actor_login", "alice")
		actor := RequestActor(c)
		if actor.Login != "alice" || actor.Source != "authenticated_context" {
			t.Fatalf("got %+v", actor)
		}
	})

	t.Run("whitespace-only actor falls back to system", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("devhub_actor_login", "   ")
		actor := RequestActor(c)
		if actor.Login != "system" {
			t.Fatalf("got %+v", actor)
		}
	})

	t.Run("non-string actor falls back to system", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("devhub_actor_login", 12345)
		actor := RequestActor(c)
		if actor.Login != "system" {
			t.Fatalf("got %+v", actor)
		}
	})
}

func TestDevFallbackEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing returns false", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		if DevFallbackEnabled(c) {
			t.Fatal("expected false")
		}
	})

	t.Run("true value returns true", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("devhub_auth_dev_fallback", true)
		if !DevFallbackEnabled(c) {
			t.Fatal("expected true")
		}
	})

	t.Run("false value returns false", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("devhub_auth_dev_fallback", false)
		if DevFallbackEnabled(c) {
			t.Fatal("expected false")
		}
	})

	t.Run("non-bool value returns false", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("devhub_auth_dev_fallback", "yes")
		if DevFallbackEnabled(c) {
			t.Fatal("expected false")
		}
	})
}
