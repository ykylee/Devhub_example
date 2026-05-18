package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
