package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// doJSON is a shared helper that issues a JSON HTTP request against a test
// router and returns the recorded response. Origin: accounts_admin_test.go
// (removed in sprint -i, issue #209, ADR-0020 sub-carve B). The helper is
// still referenced by application / project / repository / integration tests
// so it is preserved here as a top-level test helper.
func doJSON(t *testing.T, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
