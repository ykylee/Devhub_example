package httpapi

import (
	"bytes"
	"log"
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
