package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWriteServerError_MasksUnderlyingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/boom", func(c *gin.Context) {
		writeServerError(c, errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`), "test.boom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	body := rec.Body.String()
	if strings.Contains(body, "duplicate key") || strings.Contains(body, "users_email_key") || strings.Contains(body, "constraint") {
		t.Fatalf("response body leaks underlying error: %s", body)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response not JSON: %v (%s)", err, body)
	}
	if got := payload["status"]; got != "failed" {
		t.Fatalf("status field: got %v, want \"failed\"", got)
	}
	if got := payload["error"]; got != "internal error" {
		t.Fatalf("error field: got %v, want \"internal error\"", got)
	}
}
