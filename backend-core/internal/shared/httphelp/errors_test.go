package httphelp

import (
	"bytes"
	"errors"
	"log"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseBoundedInt(t *testing.T) {
	cases := []struct {
		name      string
		value     string
		minValue  int
		maxValue  int
		want      int
		expectErr bool
		errIs     error
	}{
		{name: "valid mid range", value: "50", minValue: 0, maxValue: 100, want: 50},
		{name: "valid at min boundary", value: "0", minValue: 0, maxValue: 100, want: 0},
		{name: "valid at max boundary", value: "100", minValue: 0, maxValue: 100, want: 100},
		{name: "below minValue returns ErrSyntax", value: "-1", minValue: 0, maxValue: 100, expectErr: true, errIs: strconv.ErrSyntax},
		{name: "above maxValue returns ErrSyntax", value: "101", minValue: 0, maxValue: 100, expectErr: true, errIs: strconv.ErrSyntax},
		{name: "non-numeric returns Atoi error", value: "abc", minValue: 0, maxValue: 100, expectErr: true},
		{name: "empty string returns Atoi error", value: "", minValue: 0, maxValue: 100, expectErr: true},
		{name: "negative range allowed", value: "-5", minValue: -10, maxValue: 0, want: -5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseBoundedInt(tc.value, tc.minValue, tc.maxValue)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.errIs != nil && !errors.Is(err, tc.errIs) {
					t.Fatalf("expected errors.Is %v, got %v", tc.errIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestWriteServerError_LogsAndReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logBuf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(originalOutput) })

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/x", nil)
	c.Set(CtxKeyRequestID, "req_test_123")

	WriteServerError(c, errors.New("boom"), "test.op")

	if rec.Code != 500 {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"internal error"`) {
		t.Fatalf("expected generic body, got %q", rec.Body.String())
	}
	logged := logBuf.String()
	if !strings.Contains(logged, "op=test.op") {
		t.Fatalf("expected op in log, got %q", logged)
	}
	if !strings.Contains(logged, "request_id=req_test_123") {
		t.Fatalf("expected request_id in log, got %q", logged)
	}
	if !strings.Contains(logged, "err=boom") {
		t.Fatalf("expected err in log, got %q", logged)
	}
}

func TestWriteServerError_DefaultsRequestIDToDash(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logBuf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(originalOutput) })

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/x", nil)

	WriteServerError(c, errors.New("boom"), "op")

	if !strings.Contains(logBuf.String(), "request_id=-") {
		t.Fatalf("expected request_id=- default, got %q", logBuf.String())
	}
}

func TestErrIdentityNotFound_Sentinel(t *testing.T) {
	if ErrIdentityNotFound == nil {
		t.Fatal("ErrIdentityNotFound should not be nil")
	}
	wrapped := errors.New("wrapped: " + ErrIdentityNotFound.Error())
	if errors.Is(wrapped, ErrIdentityNotFound) {
		t.Fatal("non-wrapped error must not match")
	}
	wrappedProper := errors.Join(ErrIdentityNotFound, errors.New("ctx"))
	if !errors.Is(wrappedProper, ErrIdentityNotFound) {
		t.Fatal("errors.Is with errors.Join must match")
	}
}
