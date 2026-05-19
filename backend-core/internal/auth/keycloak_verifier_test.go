package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestKeycloakJWKSVerifier_VerifyBearerTokenWithJWKSURL(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	kid := "kid-1"
	issuer := "https://issuer.example.com/realms/devhub"
	aud := "devhub-web"

	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": kid,
				"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}),
			}},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	token := mustSignToken(t, key, kid, jwt.MapClaims{
		"iss":                issuer,
		"aud":                aud,
		"sub":                "user-1",
		"preferred_username": "alice",
		"realm_access": map[string]any{
			"roles": []string{"manager"},
		},
		"exp": time.Now().Add(5 * time.Minute).Unix(),
	})

	v := &KeycloakJWKSVerifier{
		IssuerURL: issuer,
		JWKSURL:   srv.URL + "/jwks",
		ClientID:  aud,
	}
	actor, err := v.VerifyBearerToken(context.Background(), token)
	if err != nil {
		t.Fatalf("VerifyBearerToken err: %v", err)
	}
	if actor.Subject != "user-1" {
		t.Fatalf("subject = %q; want %q", actor.Subject, "user-1")
	}
	if actor.Login != "alice" {
		t.Fatalf("login = %q; want %q", actor.Login, "alice")
	}
	if actor.Role != "manager" {
		t.Fatalf("role = %q; want %q", actor.Role, "manager")
	}
}

func TestKeycloakJWKSVerifier_VerifyBearerTokenWithDiscovery(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	kid := "kid-2"
	aud := "devhub-api"

	mux := http.NewServeMux()
	var issuer string
	mux.HandleFunc("/realms/devhub/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":   issuer,
			"jwks_uri": issuer + "/jwks",
		})
	})
	mux.HandleFunc("/realms/devhub/jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": kid,
				"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}),
			}},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	issuer = srv.URL + "/realms/devhub"

	token := mustSignToken(t, key, kid, jwt.MapClaims{
		"iss":   issuer,
		"aud":   aud,
		"sub":   "user-2",
		"email": "bob@example.com",
		"exp":   time.Now().Add(5 * time.Minute).Unix(),
	})

	v := &KeycloakJWKSVerifier{
		IssuerURL: issuer,
		ClientID:  aud,
	}
	actor, err := v.VerifyBearerToken(context.Background(), token)
	if err != nil {
		t.Fatalf("VerifyBearerToken err: %v", err)
	}
	if actor.Login != "bob@example.com" {
		t.Fatalf("login = %q; want %q", actor.Login, "bob@example.com")
	}
}

func TestKeycloakJWKSVerifier_CachesJWKSBetweenVerifications(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	kid := "kid-cache"
	issuer := "https://issuer.example.com/realms/devhub"
	aud := "devhub-web"

	var hits int32
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": kid,
				"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}),
			}},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	v := &KeycloakJWKSVerifier{
		IssuerURL: issuer,
		JWKSURL:   srv.URL + "/jwks",
		ClientID:  aud,
		CacheTTL:  time.Hour,
	}

	token1 := mustSignToken(t, key, kid, jwt.MapClaims{
		"iss":                issuer,
		"aud":                aud,
		"sub":                "user-cache-1",
		"preferred_username": "alice",
		"exp":                time.Now().Add(5 * time.Minute).Unix(),
	})
	token2 := mustSignToken(t, key, kid, jwt.MapClaims{
		"iss":                issuer,
		"aud":                aud,
		"sub":                "user-cache-2",
		"preferred_username": "bob",
		"exp":                time.Now().Add(5 * time.Minute).Unix(),
	})

	if _, err := v.VerifyBearerToken(context.Background(), token1); err != nil {
		t.Fatalf("first verify err: %v", err)
	}
	if _, err := v.VerifyBearerToken(context.Background(), token2); err != nil {
		t.Fatalf("second verify err: %v", err)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("jwks endpoint hits = %d; want 1", got)
	}
}

func TestExtractKeycloakRole_UsesResourceAccessFallback(t *testing.T) {
	claims := jwt.MapClaims{
		"resource_access": map[string]any{
			"devhub-frontend": map[string]any{
				"roles": []any{"system_admin"},
			},
		},
	}
	role := extractKeycloakRole(claims)
	if role != "system_admin" {
		t.Fatalf("role = %q; want %q", role, "system_admin")
	}
}

// TestExtractKeycloakRole_MultiRolePriority — sprint claude/work_260519-q
// (ADR-0019 §5.3 codex review #9 backend 확장 carve). Multi-role token (group
// composite role 가입 시 자연 발생) 의 priority filter 검증. selectHighestPriorityRole
// 가 4 known role (system_admin > pmo_manager > manager > developer) 의 highest
// priority 만 반환해야 함. Order-dependency 회피.
func TestExtractKeycloakRole_MultiRolePriority(t *testing.T) {
	cases := []struct {
		name  string
		roles []any
		want  string
	}{
		{"system_admin first wins over manager+developer", []any{"system_admin", "manager", "developer"}, "system_admin"},
		{"system_admin last wins over manager+developer (order-independent)", []any{"developer", "manager", "system_admin"}, "system_admin"},
		{"pmo_manager wins over manager", []any{"manager", "pmo_manager"}, "pmo_manager"},
		{"developer only", []any{"developer"}, "developer"},
		{"unknown role fallback to first", []any{"unknown_role"}, "unknown_role"},
		{"known wins over unknown regardless of order", []any{"unknown_role", "developer"}, "developer"},
		{"all 4 known roles → system_admin", []any{"developer", "manager", "pmo_manager", "system_admin"}, "system_admin"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims := jwt.MapClaims{
				"realm_access": map[string]any{"roles": tc.roles},
			}
			role := extractKeycloakRole(claims)
			if role != tc.want {
				t.Fatalf("role = %q; want %q (roles=%v)", role, tc.want, tc.roles)
			}
		})
	}
}

func TestSelectHighestPriorityRole_EmptyList(t *testing.T) {
	if got := selectHighestPriorityRole(nil); got != "" {
		t.Fatalf("empty list should return empty string, got %q", got)
	}
	if got := selectHighestPriorityRole([]string{}); got != "" {
		t.Fatalf("empty slice should return empty string, got %q", got)
	}
}

func TestKeycloakJWKSVerifier_VerifyBearerTokenRejectsInvalidAudience(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	kid := "kid-3"
	issuer := "https://issuer.example.com/realms/devhub"

	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": kid,
				"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}),
			}},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	token := mustSignToken(t, key, kid, jwt.MapClaims{
		"iss": issuer,
		"aud": "other-client",
		"sub": "user-3",
		"exp": time.Now().Add(5 * time.Minute).Unix(),
	})

	v := &KeycloakJWKSVerifier{
		IssuerURL: issuer,
		JWKSURL:   srv.URL + "/jwks",
		ClientID:  "expected-client",
	}
	if _, err := v.VerifyBearerToken(context.Background(), token); err == nil {
		t.Fatalf("expected audience validation error")
	}
}

func mustSignToken(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}
