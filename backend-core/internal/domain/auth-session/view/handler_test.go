package view

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

type fakeAuditStore struct {
	created []domain.AuditLog
	err     error
}

func (f *fakeAuditStore) CreateAuditLog(_ context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	if f.err != nil {
		return domain.AuditLog{}, f.err
	}
	log.AuditID = "audit_test_id"
	f.created = append(f.created, log)
	return log, nil
}

func TestNewAuthHandler_ConfigPropagation(t *testing.T) {
	cfg := AuthConfig{OnboardingGateEnabled: true, AuthDevFallback: true}
	h := NewAuthHandler(cfg)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if !h.cfg.OnboardingGateEnabled || !h.cfg.AuthDevFallback {
		t.Fatal("config not propagated")
	}
}

func TestRecordAuditBestEffort_NilStoreReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(AuthConfig{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	got := h.recordAuditBestEffort(c, "action", "target_type", "id", nil)
	if got.AuditID != "" {
		t.Fatalf("expected zero AuditLog, got %+v", got)
	}
}

func TestRecordAuditBestEffort_PersistAndFillsActorSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAuditStore{}
	h := NewAuthHandler(AuthConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set("devhub_actor_login", "alice")

	got := h.recordAuditBestEffort(c, "test.action", "user", "u-1", nil)
	if got.AuditID != "audit_test_id" {
		t.Fatalf("expected audit id stamped, got %+v", got)
	}
	if len(store.created) != 1 {
		t.Fatalf("expected 1 created, got %d", len(store.created))
	}
	created := store.created[0]
	if created.ActorLogin != "alice" {
		t.Fatalf("actor_login = %q", created.ActorLogin)
	}
	if created.Action != "test.action" || created.TargetType != "user" || created.TargetID != "u-1" {
		t.Fatalf("action mapping = %+v", created)
	}
	src, _ := created.Payload["actor_source"].(string)
	if src != "authenticated_context" {
		t.Fatalf("actor_source = %q", src)
	}
}

func TestRecordAuditBestEffort_PayloadPreservedAndAugmented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeAuditStore{}
	h := NewAuthHandler(AuthConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)

	h.recordAuditBestEffort(c, "a", "t", "id", map[string]any{"existing": "value"})
	created := store.created[0]
	if created.Payload["existing"] != "value" {
		t.Fatalf("existing payload key lost: %+v", created.Payload)
	}
	if _, ok := created.Payload["actor_source"]; !ok {
		t.Fatal("actor_source must be augmented")
	}
}

func TestRecordAuditBestEffort_PersistFailureLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logBuf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(orig) })

	store := &fakeAuditStore{err: errors.New("db_down")}
	h := NewAuthHandler(AuthConfig{AuditStore: store})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/x", nil)
	c.Set(httphelp.CtxKeyRequestID, "req_x")

	got := h.recordAuditBestEffort(c, "a", "t", "id", nil)
	if got.AuditID != "" {
		t.Fatal("expected zero audit on err")
	}
	if !strings.Contains(logBuf.String(), "audit log persistence failed") {
		t.Fatalf("expected log, got %q", logBuf.String())
	}
}

func TestRequireOnboardingFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("enabled returns true and no body", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: true})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		if !h.RequireOnboardingFlag(c) {
			t.Fatal("expected true")
		}
		if rec.Code == 404 {
			t.Fatal("must not write 404")
		}
	})

	t.Run("disabled returns false + 404 body", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: false})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		if h.RequireOnboardingFlag(c) {
			t.Fatal("expected false")
		}
		if rec.Code != 404 {
			t.Fatalf("got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "onboarding_feature_disabled") {
			t.Fatalf("body=%q", rec.Body.String())
		}
	})
}

func TestAppUserFromDomain_FullMapping(t *testing.T) {
	user := domain.AppUser{
		ID:            42,
		UserID:        "u-1",
		Email:         "alice@example.com",
		DisplayName:   "Alice",
		Role:          domain.AppRoleDeveloper,
		Status:        domain.UserStatusActive,
		PrimaryUnitID: "team-a",
		CurrentUnitID: "team-b",
		IsSeconded:    true,
		Appointments: []domain.UnitAppointment{
			{UnitID: "team-a", AppointmentRole: domain.AppointmentRoleLeader},
			{UnitID: "team-b", AppointmentRole: domain.AppointmentRoleMember},
		},
		ReviewStatus: "approved",
	}

	got := appUserFromDomain(user)
	if got.UserID != "u-1" || got.Email != "alice@example.com" || got.DisplayName != "Alice" {
		t.Fatalf("basic field mapping wrong: %+v", got)
	}
	if got.Role != string(domain.AppRoleDeveloper) {
		t.Fatalf("role mapping wrong: %q", got.Role)
	}
	if got.PrimaryUnitID != "team-a" || got.CurrentUnitID != "team-b" {
		t.Fatalf("unit mapping wrong: %+v", got)
	}
	if !got.IsSeconded {
		t.Fatal("seconded flag")
	}
	if len(got.Appointments) != 2 {
		t.Fatalf("appointments count = %d", len(got.Appointments))
	}
	if got.Appointments[0].UnitID != "team-a" || got.Appointments[0].AppointmentRole != string(domain.AppointmentRoleLeader) {
		t.Fatalf("appt 0: %+v", got.Appointments[0])
	}
	if got.ReviewStatus != "approved" {
		t.Fatalf("review status: %q", got.ReviewStatus)
	}
}

func TestAppUserFromDomain_EmptyAppointments(t *testing.T) {
	user := domain.AppUser{UserID: "u-1"}
	got := appUserFromDomain(user)
	if got.Appointments == nil {
		t.Fatal("Appointments must not be nil (slice allocation)")
	}
	if len(got.Appointments) != 0 {
		t.Fatalf("len = %d", len(got.Appointments))
	}
}

func TestAddAuditMeta(t *testing.T) {
	t.Run("with audit id stamps key", func(t *testing.T) {
		resp := gin.H{}
		addAuditMeta(resp, domain.AuditLog{AuditID: "audit_x"})
		if resp["audit_log_id"] != "audit_x" {
			t.Fatalf("got %v", resp["audit_log_id"])
		}
	})
	t.Run("without audit id no-op", func(t *testing.T) {
		resp := gin.H{}
		addAuditMeta(resp, domain.AuditLog{})
		if _, ok := resp["audit_log_id"]; ok {
			t.Fatal("must not stamp empty id")
		}
	})
}

// --- Fakes for AuthenticateActor Middleware Tests --------------------

type fakeBearerTokenVerifier struct {
	actor AuthenticatedActor
	err   error
}

func (f *fakeBearerTokenVerifier) VerifyBearerToken(_ context.Context, _ string) (AuthenticatedActor, error) {
	return f.actor, f.err
}

type fakeOrganizationStore struct {
	user          domain.AppUser
	getUserErr    error
	updateUserErr error
	subjects      map[string]string
	setSubErr     error
}

func (f *fakeOrganizationStore) GetUser(_ context.Context, _ string) (domain.AppUser, error) {
	if f.getUserErr != nil {
		return domain.AppUser{}, f.getUserErr
	}
	return f.user, nil
}

func (f *fakeOrganizationStore) SetIdPSubject(_ context.Context, userID string, identityID string) error {
	if f.setSubErr != nil {
		return f.setSubErr
	}
	if f.subjects == nil {
		f.subjects = make(map[string]string)
	}
	f.subjects[userID] = identityID
	return nil
}

func (f *fakeOrganizationStore) UpdateUser(_ context.Context, _ string, _ domain.UpdateUserInput) (domain.AppUser, error) {
	if f.updateUserErr != nil {
		return domain.AppUser{}, f.updateUserErr
	}
	return f.user, nil
}

type fakeRealtimeTicketStore struct {
	ticket store.RealtimeTicket
	ok     bool
	err    error
}

func (f *fakeRealtimeTicketStore) Consume(_ context.Context, _ string) (store.RealtimeTicket, bool, error) {
	return f.ticket, f.ok, f.err
}

// --- AuthenticateActor Middleware Tests ------------------------------

func TestAuthenticateActor_XDevhubActorRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(AuthConfig{})
	rec := httptest.NewRecorder()
	
	r := gin.New()
	r.Use(h.AuthenticateActor)
	r.GET("/api/v1/apps", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/api/v1/apps", nil)
	req.Header.Set("X-Devhub-Actor", "legacy-actor")
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "X-Devhub-Actor header is removed") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestAuthenticateActor_PublicBypassPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(AuthConfig{})

	t.Run("gitea webhook sets webhook source", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r := gin.New()
		var capturedSourceType any
		r.Use(h.AuthenticateActor)
		r.POST("/api/v1/integrations/gitea/webhooks", func(c *gin.Context) {
			capturedSourceType, _ = c.Get(httphelp.CtxKeySourceType)
			c.Status(200)
		})

		req := httptest.NewRequest("POST", "/api/v1/integrations/gitea/webhooks", nil)
		r.ServeHTTP(rec, req)

		if rec.Code == 401 {
			t.Fatal("must bypass auth")
		}
		if capturedSourceType != domain.AuditSourceWebhook {
			t.Fatalf("source type = %v", capturedSourceType)
		}
	})

	t.Run("provider webhook sets webhook source", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r := gin.New()
		var capturedSourceType any
		r.Use(h.AuthenticateActor)
		r.POST("/api/v1/integration/providers/:provider_id/webhook", func(c *gin.Context) {
			capturedSourceType, _ = c.Get(httphelp.CtxKeySourceType)
			c.Status(200)
		})

		req := httptest.NewRequest("POST", "/api/v1/integration/providers/123/webhook", nil)
		r.ServeHTTP(rec, req)

		if capturedSourceType != domain.AuditSourceWebhook {
			t.Fatalf("source type = %v", capturedSourceType)
		}
	})
}

func TestAuthenticateActor_WebSocketTicketBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid ticket sets context", func(t *testing.T) {
		storeVal := &fakeRealtimeTicketStore{
			ticket: store.RealtimeTicket{
				ActorLogin: "bob",
				ActorRole:  "Admin",
				SourceType: string(domain.AuditSourceOIDC),
			},
			ok: true,
		}
		h := NewAuthHandler(AuthConfig{RealtimeTickets: storeVal})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		var capturedLogin, capturedRole, capturedSourceType any
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/realtime/ws", func(c *gin.Context) {
			capturedLogin, _ = c.Get("devhub_actor_login")
			capturedRole, _ = c.Get("devhub_actor_role")
			capturedSourceType, _ = c.Get(httphelp.CtxKeySourceType)
			c.Status(200)
		})

		req := httptest.NewRequest("GET", "/api/v1/realtime/ws?ticket=t-123", nil)
		r.ServeHTTP(rec, req)

		if rec.Code == 401 {
			t.Fatal("must accept valid ticket")
		}
		if capturedLogin != "bob" || capturedRole != "Admin" || capturedSourceType != string(domain.AuditSourceOIDC) {
			t.Fatalf("ctx mismatch: login=%v role=%v st=%v", capturedLogin, capturedRole, capturedSourceType)
		}
	})

	t.Run("invalid ticket aborts 401", func(t *testing.T) {
		storeVal := &fakeRealtimeTicketStore{ok: false}
		h := NewAuthHandler(AuthConfig{RealtimeTickets: storeVal})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/realtime/ws", func(c *gin.Context) { c.Status(200) })

		req := httptest.NewRequest("GET", "/api/v1/realtime/ws?ticket=bad", nil)
		r.ServeHTTP(rec, req)

		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("store failure aborts 503", func(t *testing.T) {
		storeVal := &fakeRealtimeTicketStore{err: errors.New("db outage")}
		h := NewAuthHandler(AuthConfig{RealtimeTickets: storeVal})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/realtime/ws", func(c *gin.Context) { c.Status(200) })

		req := httptest.NewRequest("GET", "/api/v1/realtime/ws?ticket=t-99", nil)
		r.ServeHTTP(rec, req)

		if rec.Code != 503 {
			t.Fatalf("expected 503, got %d", rec.Code)
		}
	})
}

func TestAuthenticateActor_NoHeaderBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("no header auth dev fallback", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{AuthDevFallback: true})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/apps", func(c *gin.Context) { c.Status(200) })

		req := httptest.NewRequest("GET", "/api/v1/apps", nil)
		r.ServeHTTP(rec, req)

		if rec.Code == 401 {
			t.Fatal("fallback must bypass")
		}
		if rec.Header().Get("X-Devhub-Auth") != "dev_fallback_no_header" {
			t.Fatalf("header = %q", rec.Header().Get("X-Devhub-Auth"))
		}
	})

	t.Run("no header aborts 401", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{AuthDevFallback: false})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/apps", func(c *gin.Context) { c.Status(200) })

		req := httptest.NewRequest("GET", "/api/v1/apps", nil)
		r.ServeHTTP(rec, req)

		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("header invalid scheme aborts 401", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/apps", func(c *gin.Context) { c.Status(200) })

		req := httptest.NewRequest("GET", "/api/v1/apps", nil)
		req.Header.Set("Authorization", "Basic credentials")
		r.ServeHTTP(rec, req)

		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}

func TestAuthenticateActor_BearerVerifier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("no verifier dev fallback", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{AuthDevFallback: true})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/apps", func(c *gin.Context) { c.Status(200) })

		req := httptest.NewRequest("GET", "/api/v1/apps", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		r.ServeHTTP(rec, req)

		if rec.Code == 401 {
			t.Fatal("fallback must bypass")
		}
		if rec.Header().Get("X-Devhub-Auth") != "bearer_unverified" {
			t.Fatalf("header = %q", rec.Header().Get("X-Devhub-Auth"))
		}
	})

	t.Run("no verifier aborts 401", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/apps", func(c *gin.Context) { c.Status(200) })

		req := httptest.NewRequest("GET", "/api/v1/apps", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		r.ServeHTTP(rec, req)

		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("verifier returns error aborts 401", func(t *testing.T) {
		verifier := &fakeBearerTokenVerifier{err: errors.New("bad signature")}
		h := NewAuthHandler(AuthConfig{BearerTokenVerifier: verifier})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/apps", func(c *gin.Context) { c.Status(200) })

		req := httptest.NewRequest("GET", "/api/v1/apps", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		r.ServeHTTP(rec, req)

		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}

func TestAuthenticateActor_SuccessPathStoreScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("empty login and subject aborts 401", func(t *testing.T) {
		verifier := &fakeBearerTokenVerifier{actor: AuthenticatedActor{}}
		h := NewAuthHandler(AuthConfig{BearerTokenVerifier: verifier})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/apps", func(c *gin.Context) { c.Status(200) })

		req := httptest.NewRequest("GET", "/api/v1/apps", nil)
		req.Header.Set("Authorization", "Bearer empty")
		r.ServeHTTP(rec, req)

		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("user exists idp_subject backfilled and role lookup", func(t *testing.T) {
		verifier := &fakeBearerTokenVerifier{
			actor: AuthenticatedActor{
				Login:   "alice",
				Subject: "keycloak-sub-123",
				Role:    "Developer",
			},
		}
		orgStore := &fakeOrganizationStore{
			user: domain.AppUser{
				UserID:      "alice",
				IdPSubject:  "", // empty to trigger backfill
				Role:        domain.AppRoleSystemAdmin,
				DisplayName: "Alice Admin",
			},
		}
		h := NewAuthHandler(AuthConfig{
			BearerTokenVerifier: verifier,
			OrganizationStore:   orgStore,
		})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		var capturedLogin, capturedRole any
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/apps", func(c *gin.Context) {
			capturedLogin, _ = c.Get("devhub_actor_login")
			capturedRole, _ = c.Get("devhub_actor_role")
			c.Status(200)
		})

		req := httptest.NewRequest("GET", "/api/v1/apps", nil)
		req.Header.Set("Authorization", "Bearer token-123")
		r.ServeHTTP(rec, req)

		if rec.Code == 401 {
			t.Fatal("must succeed")
		}
		if capturedLogin != "alice" || capturedRole != string(domain.AppRoleSystemAdmin) {
			t.Fatalf("actor mismatch: login=%v role=%v", capturedLogin, capturedRole)
		}
		sub, ok := orgStore.subjects["alice"]
		if !ok || sub != "keycloak-sub-123" {
			t.Fatalf("lazy backfill failed: ok=%v sub=%v", ok, sub)
		}
	})

	t.Run("user not found OIDC token-only actor with onboarding flag", func(t *testing.T) {
		verifier := &fakeBearerTokenVerifier{
			actor: AuthenticatedActor{
				Login:       "newuser",
				Email:       "new@example.com",
				DisplayName: "New User",
				Role:        "Manager",
			},
		}
		orgStore := &fakeOrganizationStore{getUserErr: store.ErrNotFound}
		h := NewAuthHandler(AuthConfig{
			BearerTokenVerifier: verifier,
			OrganizationStore:   orgStore,
		})
		rec := httptest.NewRecorder()
		
		r := gin.New()
		var login, role, email, dName, onb any
		r.Use(h.AuthenticateActor)
		r.GET("/api/v1/apps", func(c *gin.Context) {
			login, _ = c.Get("devhub_actor_login")
			role, _ = c.Get("devhub_actor_role")
			email, _ = c.Get("devhub_actor_email")
			dName, _ = c.Get("devhub_actor_display_name")
			onb, _ = c.Get("devhub_onboarding_required")
			c.Status(200)
		})

		req := httptest.NewRequest("GET", "/api/v1/apps", nil)
		req.Header.Set("Authorization", "Bearer new-token")
		r.ServeHTTP(rec, req)

		if login != "newuser" || role != "Manager" || email != "new@example.com" || dName != "New User" || onb != true {
			t.Fatalf("token-only actor mapping mismatch: login=%v role=%v email=%v dName=%v onb=%v", login, role, email, dName, onb)
		}
	})
}

// --- GetMe and PatchMe Handler Tests ---------------------------------

func TestGetMe_SuccessAndErrorPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("empty login aborts 401", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		// RequestActor returns empty actor on empty context
		c.Request = httptest.NewRequest("GET", "/me", nil)

		h.GetMe(c)

		if rec.Code != 401 {
			t.Fatalf("got %d", rec.Code)
		}
	})

	t.Run("system login aborts 401", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/me", nil)
		c.Set("devhub_actor_login", "system")

		h.GetMe(c)

		if rec.Code != 401 {
			t.Fatalf("got %d", rec.Code)
		}
	})

	t.Run("no store success basic data", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/me", nil)
		c.Set("devhub_actor_login", "charlie")
		c.Set("devhub_actor_role", "Developer")
		c.Set("devhub_actor_subject", "sub-charlie")

		h.GetMe(c)

		if rec.Code != 200 {
			t.Fatalf("got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "sub-charlie") {
			t.Fatalf("body = %q", rec.Body.String())
		}
	})

	t.Run("user exists store returns full details", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{
			user: domain.AppUser{
				UserID:        "charlie",
				Email:         "charlie@example.com",
				DisplayName:   "Charlie C",
				PrimaryUnitID: "unit-1",
				CurrentUnitID: "unit-1",
				ReviewStatus:  "pending_review",
			},
		}
		h := NewAuthHandler(AuthConfig{OrganizationStore: orgStore})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/me", nil)
		c.Set("devhub_actor_login", "charlie")

		h.GetMe(c)

		if rec.Code != 200 {
			t.Fatalf("got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Charlie C") || !strings.Contains(body, "pending_review") {
			t.Fatalf("body = %q", body)
		}
	})

	t.Run("store err not found resolves token claims", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{getUserErr: store.ErrNotFound}
		h := NewAuthHandler(AuthConfig{OrganizationStore: orgStore})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/me", nil)
		c.Set("devhub_actor_login", "charlie")
		c.Set("devhub_actor_email", "charlie_token@example.com")
		c.Set("devhub_actor_display_name", "Charlie Token")

		h.GetMe(c)

		if rec.Code != 200 {
			t.Fatalf("got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Charlie Token") || !strings.Contains(body, "charlie_token@example.com") {
			t.Fatalf("body = %q", body)
		}
	})

	t.Run("store general error defaults onboarding required", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{getUserErr: errors.New("db outage")}
		h := NewAuthHandler(AuthConfig{OrganizationStore: orgStore})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", "/me", nil)
		c.Set("devhub_actor_login", "charlie")

		h.GetMe(c)

		if rec.Code != 200 {
			t.Fatalf("got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"onboarding_required":true`) {
			t.Fatalf("body = %q", rec.Body.String())
		}
	})
}

func TestPatchMe_SuccessAndErrorPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("gate disabled aborts 404", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: false})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		h.PatchMe(c)

		if rec.Code != 404 {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("no store aborts 503", func(t *testing.T) {
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: true})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		h.PatchMe(c)

		if rec.Code != 503 {
			t.Fatalf("expected 503, got %d", rec.Code)
		}
	})

	t.Run("empty login aborts 401", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{}
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: true, OrganizationStore: orgStore})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("PATCH", "/me", nil)

		h.PatchMe(c)

		if rec.Code != 401 {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("invalid json payload aborts 422", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{}
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: true, OrganizationStore: orgStore})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("PATCH", "/me", strings.NewReader("{invalid-json"))
		c.Set("devhub_actor_login", "alice")

		h.PatchMe(c)

		if rec.Code != 422 {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("empty patch request aborts 422", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{}
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: true, OrganizationStore: orgStore})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("PATCH", "/me", strings.NewReader("{}"))
		c.Set("devhub_actor_login", "alice")

		h.PatchMe(c)

		if rec.Code != 422 {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("invalid display_name limits aborts 422", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{}
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: true, OrganizationStore: orgStore})

		// empty display_name
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("PATCH", "/me", strings.NewReader(`{"display_name":""}`))
		c.Set("devhub_actor_login", "alice")
		h.PatchMe(c)
		if rec.Code != 422 {
			t.Fatalf("expected 422, got %d", rec.Code)
		}

		// 100+ chars display_name
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest("PATCH", "/me", strings.NewReader(`{"display_name":"`+strings.Repeat("a", 101)+`"}`))
		c2.Set("devhub_actor_login", "alice")
		h.PatchMe(c2)
		if rec2.Code != 422 {
			t.Fatalf("expected 422, got %d", rec2.Code)
		}
	})

	t.Run("empty primary_unit_id aborts 422", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{}
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: true, OrganizationStore: orgStore})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("PATCH", "/me", strings.NewReader(`{"primary_unit_id":""}`))
		c.Set("devhub_actor_login", "alice")

		h.PatchMe(c)

		if rec.Code != 422 {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("store ErrNotFound aborts 404", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{updateUserErr: store.ErrNotFound}
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: true, OrganizationStore: orgStore})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("PATCH", "/me", strings.NewReader(`{"display_name":"New Name"}`))
		c.Set("devhub_actor_login", "alice")

		h.PatchMe(c)

		if rec.Code != 404 {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("store general error aborts 500", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{updateUserErr: errors.New("db connection lost")}
		h := NewAuthHandler(AuthConfig{OnboardingGateEnabled: true, OrganizationStore: orgStore})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("PATCH", "/me", strings.NewReader(`{"display_name":"New Name"}`))
		c.Set("devhub_actor_login", "alice")

		h.PatchMe(c)

		if rec.Code != 500 {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("successful profile update with unit change record audit", func(t *testing.T) {
		orgStore := &fakeOrganizationStore{
			user: domain.AppUser{
				UserID:        "alice",
				Email:         "alice@example.com",
				DisplayName:   "Alice Updated",
				PrimaryUnitID: "new-team",
				Role:          domain.AppRoleDeveloper,
				ReviewStatus:  "pending_review",
			},
		}
		auditStore := &fakeAuditStore{}
		h := NewAuthHandler(AuthConfig{
			OnboardingGateEnabled: true,
			OrganizationStore:     orgStore,
			AuditStore:            auditStore,
		})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("PATCH", "/me", strings.NewReader(`{"display_name":"Alice Updated","primary_unit_id":"new-team"}`))
		c.Set("devhub_actor_login", "alice")

		h.PatchMe(c)

		if rec.Code != 200 {
			t.Fatalf("expected 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
		if len(auditStore.created) != 1 {
			t.Fatalf("expected 1 audit log, got %d", len(auditStore.created))
		}
		created := auditStore.created[0]
		if created.Action != "account.unit_changed" || created.TargetID != "alice" {
			t.Fatalf("wrong audit mapping: %+v", created)
		}
	})
}


