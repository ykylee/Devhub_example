package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRBACPolicyReturnsDefaultMatrix(t *testing.T) {
	router := NewRouter(RouterConfig{})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/rbac/policy", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Roles []struct {
				Role  string `json:"role"`
				Label string `json:"label"`
			} `json:"roles"`
			Resources []struct {
				Resource string `json:"resource"`
				Label    string `json:"label"`
			} `json:"resources"`
			Permissions []struct {
				Permission string `json:"permission"`
				Rank       int    `json:"rank"`
			} `json:"permissions"`
			Matrix map[string]map[string]string `json:"matrix"`
		} `json:"data"`
		Meta struct {
			PolicyVersion string `json:"policy_version"`
			Source        string `json:"source"`
			Editable      bool   `json:"editable"`
		} `json:"meta"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)

	if resp.Status != "ok" {
		t.Fatalf("expected ok status, got %q", resp.Status)
	}
	if len(resp.Data.Roles) != 3 || len(resp.Data.Resources) != 6 || len(resp.Data.Permissions) != 4 {
		t.Fatalf("unexpected policy dimensions: roles=%d resources=%d permissions=%d", len(resp.Data.Roles), len(resp.Data.Resources), len(resp.Data.Permissions))
	}
	if got := resp.Data.Matrix["developer"]["commands"]; got != "none" {
		t.Fatalf("expected developer commands permission none, got %q", got)
	}
	if got := resp.Data.Matrix["manager"]["risks"]; got != "write" {
		t.Fatalf("expected manager risks permission write, got %q", got)
	}
	if got := resp.Data.Matrix["system_admin"]["system_config"]; got != "admin" {
		t.Fatalf("expected system_admin system_config permission admin, got %q", got)
	}
	if resp.Meta.PolicyVersion == "" || resp.Meta.Source != "static_default_policy" || resp.Meta.Editable {
		t.Fatalf("unexpected meta: %+v", resp.Meta)
	}
}
