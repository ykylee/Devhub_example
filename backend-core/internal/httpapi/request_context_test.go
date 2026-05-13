package httpapi

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestLogRequest_PercentInIDIsNotInterpretedAsVerb pins the format-safety
// fix for the self-review pass (PR-D follow-up, work_260512-i): the request
// id is rendered through a %s verb, so a caller-supplied X-Request-ID like
// "req_%v" cannot turn into a printf verb that consumes args.
func TestLogRequest_PercentInIDIsNotInterpretedAsVerb(t *testing.T) {
	var buf bytes.Buffer
	prevOut := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})

	gin.SetMode(gin.TestMode)
	c := &gin.Context{}
	c.Set(ctxKeyRequestID, "req_%v_evil")

	logRequest(c, "value=%s", "ok")

	got := buf.String()
	if !strings.Contains(got, "request_id=req_%v_evil") {
		t.Errorf("missing literal request_id prefix: %q", got)
	}
	if !strings.Contains(got, "value=ok") {
		t.Errorf("user-supplied value got eaten by printf misalignment: %q", got)
	}
	if strings.Contains(got, "%!") {
		t.Errorf("printf reported MISSING/BAD verb — format-safety regression: %q", got)
	}
}

// TestValidateCallerRequestID_AcceptsWellFormed pins the happy path: a
// caller-supplied id matching [A-Za-z0-9_-]{1..128} is preserved as-is
// (trimmed) so downstream audit_logs.request_id stays caller-correlatable.
func TestValidateCallerRequestID_AcceptsWellFormed(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"server-style hex", "req_0123456789abcdef01234567", "req_0123456789abcdef01234567"},
		{"alnum mixed", "Trace-2026-05-13_001", "Trace-2026-05-13_001"},
		{"trimmed leading/trailing ws", "  req_abc123  ", "req_abc123"},
		{"single char", "x", "x"},
		{"max length 128", strings.Repeat("a", 128), strings.Repeat("a", 128)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validateCallerRequestID(tc.in); got != tc.want {
				t.Errorf("validateCallerRequestID(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestValidateCallerRequestID_RejectsMalformed covers the reject branches:
// control chars (CR/LF/tab), shell/percent metacharacters, unicode, empty,
// and over-length inputs all collapse to "" so requireRequestID falls back
// to a server-generated value.
func TestValidateCallerRequestID_RejectsMalformed(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"only whitespace", "   "},
		{"contains tab", "req_abc\tdef"},
		{"contains LF", "req_abc\ndef"},
		{"contains CR", "req_abc\rdef"},
		{"contains slash", "req_abc/def"},
		{"contains percent (printf verb)", "req_%v_evil"},
		{"contains space inside", "req abc"},
		{"contains equals", "req=abc"},
		{"contains dollar", "req_$abc"},
		{"unicode letter", "req_αβγ"},
		{"length 129", strings.Repeat("a", 129)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validateCallerRequestID(tc.in); got != "" {
				t.Errorf("validateCallerRequestID(%q) = %q, want \"\"", tc.in, got)
			}
		})
	}
}

// TestRequireRequestID_HonoursWellFormed pins the middleware happy path:
// a well-formed caller-supplied X-Request-ID is stashed on gin.Context and
// echoed on the response header verbatim.
func TestRequireRequestID_HonoursWellFormed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Handler{}

	r := gin.New()
	r.Use(h.requireRequestID)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"request_id": requestIDFrom(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Request-ID", "req_caller_supplied_42")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got != "req_caller_supplied_42" {
		t.Errorf("response X-Request-ID = %q, want %q", got, "req_caller_supplied_42")
	}
	if !strings.Contains(w.Body.String(), "req_caller_supplied_42") {
		t.Errorf("body did not echo request_id: %q", w.Body.String())
	}
}

// TestRequireRequestID_FallsBackOnMalformed pins the reject branch: a
// caller-supplied X-Request-ID containing control chars / invalid
// characters is dropped, the middleware generates a fresh req_<hex> id,
// and the response header carries that generated id (not the bad input).
func TestRequireRequestID_FallsBackOnMalformed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Handler{}

	r := gin.New()
	r.Use(h.requireRequestID)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"request_id": requestIDFrom(c)})
	})

	cases := []string{
		"req_%v_evil",
		"req abc",
		"req/abc",
		strings.Repeat("z", 200),
	}
	for _, badInput := range cases {
		t.Run(badInput, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			// net/http strips CR/LF on inbound, so we cannot exercise those
			// via the test transport — validateCallerRequestID-level tests
			// above cover them. Here we only exercise the chars net/http
			// passes through.
			req.Header.Set("X-Request-ID", badInput)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			got := w.Header().Get("X-Request-ID")
			if got == badInput {
				t.Fatalf("malformed caller id %q was honoured; expected fallback", badInput)
			}
			if !strings.HasPrefix(got, "req_") {
				t.Errorf("fallback id %q missing req_ prefix", got)
			}
		})
	}
}

// TestLogRequest_NoIDFallsBackToPlain pins the no-context behaviour: the
// helper degrades to plain log.Printf when requireRequestID has not run.
func TestLogRequest_NoIDFallsBackToPlain(t *testing.T) {
	var buf bytes.Buffer
	prevOut := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})

	gin.SetMode(gin.TestMode)
	c := &gin.Context{}

	logRequest(c, "plain=%s", "yes")

	got := buf.String()
	if strings.Contains(got, "request_id=") {
		t.Errorf("expected no request_id prefix when ctx is empty: %q", got)
	}
	if !strings.Contains(got, "plain=yes") {
		t.Errorf("plain fallback dropped args: %q", got)
	}
}

// TestRequestIDFromContext pins the ctx-only reader: when requireRequestID
// stashes the id on the request context, downstream code that only sees a
// context.Context can pull it back out without a gin.Context.
func TestRequestIDFromContext(t *testing.T) {
	t.Run("nil ctx returns empty", func(t *testing.T) {
		if got := requestIDFromContext(nil); got != "" {
			t.Errorf("nil ctx → %q, want \"\"", got)
		}
	})
	t.Run("ctx without value returns empty", func(t *testing.T) {
		if got := requestIDFromContext(context.Background()); got != "" {
			t.Errorf("empty ctx → %q, want \"\"", got)
		}
	})
	t.Run("ctx with value returns id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), requestIDCtxKey{}, "req_abc123")
		if got := requestIDFromContext(ctx); got != "req_abc123" {
			t.Errorf("stashed ctx → %q, want req_abc123", got)
		}
	})
}

// TestLogRequestCtx pins the ctx-only logger's percent-safety + prefix +
// no-id fallback. Matches the guarantees of logRequest but driven from a
// context.Context (no gin.Context required).
func TestLogRequestCtx(t *testing.T) {
	cases := []struct {
		name        string
		ctx         context.Context
		format      string
		args        []any
		wantPrefix  bool
		wantContent string
		denyPrefix  bool
	}{
		{
			name:        "ctx with id prefixes line",
			ctx:         context.WithValue(context.Background(), requestIDCtxKey{}, "req_zzz1"),
			format:      "value=%s",
			args:        []any{"ok"},
			wantPrefix:  true,
			wantContent: "request_id=req_zzz1 value=ok",
		},
		{
			name:        "ctx with percent-bearing id is rendered through %s",
			ctx:         context.WithValue(context.Background(), requestIDCtxKey{}, "req_%v_evil"),
			format:      "value=%s",
			args:        []any{"ok"},
			wantPrefix:  true,
			wantContent: "request_id=req_%v_evil value=ok",
		},
		{
			name:        "ctx without id falls back to plain",
			ctx:         context.Background(),
			format:      "plain=%s",
			args:        []any{"yes"},
			wantPrefix:  false,
			wantContent: "plain=yes",
			denyPrefix:  true,
		},
		{
			name:        "nil ctx falls back to plain",
			ctx:         nil,
			format:      "nilctx=%s",
			args:        []any{"sure"},
			wantPrefix:  false,
			wantContent: "nilctx=sure",
			denyPrefix:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			prevOut := log.Writer()
			prevFlags := log.Flags()
			log.SetOutput(&buf)
			log.SetFlags(0)
			t.Cleanup(func() {
				log.SetOutput(prevOut)
				log.SetFlags(prevFlags)
			})

			logRequestCtx(tc.ctx, tc.format, tc.args...)

			got := buf.String()
			if !strings.Contains(got, tc.wantContent) {
				t.Errorf("missing expected content %q in %q", tc.wantContent, got)
			}
			if tc.denyPrefix && strings.Contains(got, "request_id=") {
				t.Errorf("expected no request_id prefix but got: %q", got)
			}
			if strings.Contains(got, "%!") {
				t.Errorf("printf reported MISSING/BAD verb — format-safety regression: %q", got)
			}
		})
	}
}

// TestRequireRequestID_PropagatesToRequestContext pins the A2 wiring:
// requireRequestID stashes the id on BOTH the gin store and the underlying
// HTTP request context, so downstream client/helper code that only sees
// c.Request.Context() can still correlate its log lines.
func TestRequireRequestID_PropagatesToRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Handler{}

	var observedCtxID string
	var observedGinID string

	r := gin.New()
	r.Use(h.requireRequestID)
	r.GET("/ping", func(c *gin.Context) {
		observedCtxID = requestIDFromContext(c.Request.Context())
		observedGinID = requestIDFrom(c)
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Request-ID", "req_propagation_42")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if observedGinID != "req_propagation_42" {
		t.Errorf("gin store id = %q, want req_propagation_42", observedGinID)
	}
	if observedCtxID != "req_propagation_42" {
		t.Errorf("ctx id = %q, want req_propagation_42 (A2 propagation broken)", observedCtxID)
	}
}
