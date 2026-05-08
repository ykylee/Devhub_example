package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
)

func TestGetMeUsesTokenSubjectAndUserRole(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	orgs.users["u1"] = domain.AppUser{
		UserID:      "u1",
		Email:       "admin@example.com",
		DisplayName: "Admin",
		Role:        domain.AppRoleSystemAdmin,
		Status:      domain.UserStatusActive,
		JoinedAt:    time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC),
	}
	verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{
		Login:   "yklee",
		Subject: "u1",
		Role:    "developer",
	}}
	router := NewRouter(RouterConfig{
		OrganizationStore:   orgs,
		BearerTokenVerifier: verifier,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			User struct {
				UserID string `json:"user_id"`
				Role   string `json:"role"`
			} `json:"user"`
			Actor struct {
				Login   string `json:"login"`
				Subject string `json:"subject"`
				Source  string `json:"source"`
			} `json:"actor"`
			AllowedRoles         []string          `json:"allowed_roles"`
			EffectivePermissions map[string]string `json:"effective_permissions"`
		} `json:"data"`
	}
	decodeJSON(t, rec.Body.Bytes(), &resp)

	if resp.Data.User.UserID != "u1" || resp.Data.User.Role != "system_admin" {
		t.Fatalf("expected mapped system_admin user, got %+v", resp.Data.User)
	}
	if resp.Data.Actor.Login != "yklee" || resp.Data.Actor.Subject != "u1" || resp.Data.Actor.Source != "authenticated_context" {
		t.Fatalf("unexpected actor: %+v", resp.Data.Actor)
	}
	if got := resp.Data.EffectivePermissions["system_config"]; got != "admin" {
		t.Fatalf("expected system_config admin, got %q", got)
	}
	if len(resp.Data.AllowedRoles) != 3 || resp.Data.AllowedRoles[2] != "system_admin" {
		t.Fatalf("unexpected allowed roles: %+v", resp.Data.AllowedRoles)
	}
}

func TestGetMeRejectsInactiveUser(t *testing.T) {
	orgs := newMemoryOrganizationStore()
	orgs.users["u2"] = domain.AppUser{
		UserID: "u2",
		Role:   domain.AppRoleManager,
		Status: domain.UserStatusDeactivated,
	}
	router := NewRouter(RouterConfig{OrganizationStore: orgs})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("X-Devhub-Actor", "u2")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetMeRequiresActor(t *testing.T) {
	router := NewRouter(RouterConfig{OrganizationStore: newMemoryOrganizationStore()})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}
