package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// counterVecValue — prometheus CounterVec 의 특정 label 값을 dto.Metric 으로 읽는다.
// 다른 stale-while-error test 가 동일 label 을 건드릴 수 있어 절대값이 아닌
// delta 검증에만 사용. audit/metrics_test.go 의 동명 헬퍼와 동일 패턴.
func counterVecValue(t *testing.T, c *prometheus.CounterVec, labels []string) float64 {
	t.Helper()
	if c == nil {
		return 0
	}
	m, err := c.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("GetMetricWithLabelValues(%v): %v", labels, err)
	}
	pb := &dto.Metric{}
	if err := m.Write(pb); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if pb.Counter == nil {
		return 0
	}
	return pb.Counter.GetValue()
}

// histogramSampleCount — prometheus Histogram 의 누적 sample 개수를 dto.Metric
// 으로 읽는다. observe 횟수 delta 검증용 (bucket 분포는 별도 검증). audit 패턴에는
// 아직 histogram 헬퍼가 없어 본 패키지가 처음 도입.
func histogramSampleCount(t *testing.T, h prometheus.Histogram) uint64 {
	t.Helper()
	if h == nil {
		return 0
	}
	pb := &dto.Metric{}
	if err := h.(prometheus.Metric).Write(pb); err != nil {
		t.Fatalf("Write histogram: %v", err)
	}
	if pb.Histogram == nil {
		return 0
	}
	return pb.Histogram.GetSampleCount()
}

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

// TestKeycloakJWKSVerifier_RetriesJWKSOnKidMismatch — sprint claude/work_260519-r
// (ADR-0019 §5.3 sprint -j codex review #9 #3 backend 확장 carve). Keycloak key
// rotation 직후 새 kid 의 token 이 backend cache TTL 동안 401 되던 동작을
// stale-while-error fallback 으로 보강. cache invalidate + 1회 forced refetch +
// retry. 첫 fetch 는 old kid 응답, retry fetch 는 new kid 응답 시 token 정합.
func TestKeycloakJWKSVerifier_RetriesJWKSOnKidMismatch(t *testing.T) {
	oldKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate old rsa key: %v", err)
	}
	newKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate new rsa key: %v", err)
	}
	oldKid := "kid-old"
	newKid := "kid-new"
	issuer := "https://issuer.example.com/realms/devhub"
	aud := "devhub-web"

	var fetchCount int32
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&fetchCount, 1)
		var pubKey *rsa.PublicKey
		var kid string
		if n == 1 {
			pubKey = &oldKey.PublicKey
			kid = oldKid
		} else {
			pubKey = &newKey.PublicKey
			kid = newKid
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": kid,
				"n":   base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes()),
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

	token := mustSignToken(t, newKey, newKid, jwt.MapClaims{
		"iss":                issuer,
		"aud":                aud,
		"sub":                "user-rotation",
		"preferred_username": "alice",
		"exp":                time.Now().Add(5 * time.Minute).Unix(),
	})

	actor, err := v.VerifyBearerToken(context.Background(), token)
	if err != nil {
		t.Fatalf("verify (retry expected to succeed): %v", err)
	}
	if actor.Subject != "user-rotation" {
		t.Fatalf("actor.Subject = %q; want %q", actor.Subject, "user-rotation")
	}
	if got := atomic.LoadInt32(&fetchCount); got != 2 {
		t.Fatalf("jwks endpoint hits = %d; want 2 (1 initial + 1 retry on kid mismatch)", got)
	}
}

// TestKeycloakJWKSVerifier_DoesNotRetryOnSignatureError — non-kid error 시 retry
// 안 함 (security 위협 회피). signature error 는 attacker token 의 endpoint 폭격
// 위험 — retry 회피.
func TestKeycloakJWKSVerifier_DoesNotRetryOnSignatureError(t *testing.T) {
	correctKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate correct rsa key: %v", err)
	}
	attackerKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate attacker rsa key: %v", err)
	}
	kid := "kid-1"
	issuer := "https://issuer.example.com/realms/devhub"
	aud := "devhub-web"

	var fetchCount int32
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&fetchCount, 1)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": kid,
				"n":   base64.RawURLEncoding.EncodeToString(correctKey.PublicKey.N.Bytes()),
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

	token := mustSignToken(t, attackerKey, kid, jwt.MapClaims{
		"iss":                issuer,
		"aud":                aud,
		"sub":                "evil",
		"preferred_username": "evil",
		"exp":                time.Now().Add(5 * time.Minute).Unix(),
	})

	_, err = v.VerifyBearerToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected signature verification error")
	}
	if got := atomic.LoadInt32(&fetchCount); got != 1 {
		t.Fatalf("jwks endpoint hits = %d; want 1 (no retry on non-kid error)", got)
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

// TestKeycloakJWKSVerifier_StaleWhileError_KeycloakUnreachable — ADR-0020
// sub-carve D (sprint -l, issue #213). cache TTL 만료 후 Keycloak unreachable
// 시 stale cache 로 검증 통과 + log mark.
func TestKeycloakJWKSVerifier_StaleWhileError_KeycloakUnreachable(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	kid := "kid-stale-1"
	issuer := "https://issuer.example.com/realms/devhub"
	aud := "devhub-web"

	var jwksCallCount atomic.Int32
	jwksHandler := func(w http.ResponseWriter, _ *http.Request) {
		jwksCallCount.Add(1)
		// 첫 호출 (cache 채우기) 만 응답 — 두 번째부터는 500 (Keycloak unreachable 시뮬레이션)
		if jwksCallCount.Load() == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"keys": []map[string]any{{
					"kty": "RSA",
					"kid": kid,
					"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}),
				}},
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("simulated keycloak outage"))
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", jwksHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	token := mustSignToken(t, key, kid, jwt.MapClaims{
		"iss":                issuer,
		"aud":                aud,
		"sub":                "user-stale",
		"preferred_username": "alice",
		"exp":                time.Now().Add(5 * time.Minute).Unix(),
	})

	v := &KeycloakJWKSVerifier{
		IssuerURL:        issuer,
		JWKSURL:          srv.URL + "/jwks",
		ClientID:         aud,
		CacheTTL:         10 * time.Millisecond, // 빠르게 만료
		MaxStaleDuration: 1 * time.Hour,         // stale 1h 안전
	}

	// 1차 호출 — cache 채움
	if _, err := v.VerifyBearerToken(context.Background(), token); err != nil {
		t.Fatalf("1차 호출 err: %v", err)
	}

	// cache TTL 만료 대기
	time.Sleep(20 * time.Millisecond)

	// metric emission 기준선 — 다른 stale test 가 동일 label 을 증가시킬 수 있어
	// 절대값이 아닌 delta 검증. initJWKSMetrics 는 fetchJWKS stale 분기 진입 시
	// 호출되므로 이미 초기화돼 있다고 가정 (1차 호출이 fetch 성공 path → metric
	// 미진입). 안전을 위해 한 번 강제 init.
	initJWKSMetrics()
	beforeOK := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"ok"})
	beforeAge := histogramSampleCount(t, jwksStaleAgeSeconds)

	// 2차 호출 — JWKS 500 응답 (Keycloak unreachable) → stale fallback 통과해야
	actor, err := v.VerifyBearerToken(context.Background(), token)
	if err != nil {
		t.Fatalf("stale-while-error fallback 실패: %v", err)
	}
	if actor.Login != "alice" {
		t.Errorf("actor.Login = %q; want alice", actor.Login)
	}
	// JWKS endpoint 가 2회 호출됐어야 (1차 fetch + 2차 fetch attempt 후 fail → stale)
	if jwksCallCount.Load() != 2 {
		t.Errorf("JWKS endpoint call count = %d; want 2 (1차 fetch + 2차 fetch attempt → fail → stale fallback)", jwksCallCount.Load())
	}

	// stale fallback 진입 시 `devhub_jwks_stale_while_error_total{result="ok"}`
	// 1 증가 + `devhub_jwks_stale_age_seconds` Histogram 1 sample 추가
	// (keycloak_verifier.go 의 observeJWKSStaleWhileError + observeJWKSStaleAge
	// 콜 검증).
	if delta := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"ok"}) - beforeOK; delta != 1 {
		t.Errorf("devhub_jwks_stale_while_error_total{result=ok} delta = %v; want 1", delta)
	}
	if delta := histogramSampleCount(t, jwksStaleAgeSeconds) - beforeAge; delta != 1 {
		t.Errorf("devhub_jwks_stale_age_seconds sample delta = %v; want 1", delta)
	}
}

// TestKeycloakJWKSVerifier_StaleExpired_Fails401 — MaxStaleDuration 초과 후
// fetch 실패 시 stale 사용 안 함, 401.
func TestKeycloakJWKSVerifier_StaleExpired_Fails401(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	kid := "kid-expired-1"
	issuer := "https://issuer.example.com/realms/devhub"
	aud := "devhub-web"

	var jwksCallCount atomic.Int32
	jwksHandler := func(w http.ResponseWriter, _ *http.Request) {
		jwksCallCount.Add(1)
		if jwksCallCount.Load() == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"keys": []map[string]any{{
					"kty": "RSA",
					"kid": kid,
					"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}),
				}},
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", jwksHandler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	token := mustSignToken(t, key, kid, jwt.MapClaims{
		"iss":                issuer,
		"aud":                aud,
		"sub":                "user-expired",
		"preferred_username": "ghost",
		"exp":                time.Now().Add(5 * time.Minute).Unix(),
	})

	v := &KeycloakJWKSVerifier{
		IssuerURL:        issuer,
		JWKSURL:          srv.URL + "/jwks",
		ClientID:         aud,
		CacheTTL:         10 * time.Millisecond,
		MaxStaleDuration: 50 * time.Millisecond, // 매우 짧음 — stale 빠르게 만료
	}

	// 1차 호출 — cache 채움
	if _, err := v.VerifyBearerToken(context.Background(), token); err != nil {
		t.Fatalf("1차 호출 err: %v", err)
	}

	// CacheTTL 10ms + MaxStaleDuration 50ms = 60ms — 그 이상 대기
	time.Sleep(80 * time.Millisecond)

	// metric emission 기준선 — fail label delta 검증.
	initJWKSMetrics()
	beforeFail := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"fail"})
	beforeAge := histogramSampleCount(t, jwksStaleAgeSeconds)

	// 2차 호출 — Keycloak unreachable + stale 도 expired → 401
	_, err = v.VerifyBearerToken(context.Background(), token)
	if err == nil {
		t.Fatal("stale expired 후 fetch fail 은 401 이어야 (revoked key 보호)")
	}

	// stale 도 expired → `devhub_jwks_stale_while_error_total{result="fail"}` 1 증가.
	// age histogram 은 fail 분기에서 stale 자체가 없거나 사용 안 되므로 sample
	// 추가 없음 (metrics.go 의 observeJWKSStaleAge 는 ok 분기에서만 호출).
	if delta := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"fail"}) - beforeFail; delta != 1 {
		t.Errorf("devhub_jwks_stale_while_error_total{result=fail} delta = %v; want 1", delta)
	}
	if delta := histogramSampleCount(t, jwksStaleAgeSeconds) - beforeAge; delta != 0 {
		t.Errorf("devhub_jwks_stale_age_seconds sample delta = %v; want 0 (fail 분기는 age 미관측)", delta)
	}
}

// TestKeycloakJWKSVerifier_FreshCache_NoStaleFallback — fresh cache hit 시
// stale path 진입 안 함 (network 호출 0회).
func TestKeycloakJWKSVerifier_FreshCache_NoStaleFallback(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	kid := "kid-fresh-1"
	issuer := "https://issuer.example.com/realms/devhub"
	aud := "devhub-web"

	var jwksCallCount atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		jwksCallCount.Add(1)
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
		"sub":                "user-fresh",
		"preferred_username": "bob",
		"exp":                time.Now().Add(5 * time.Minute).Unix(),
	})

	v := &KeycloakJWKSVerifier{
		IssuerURL:        issuer,
		JWKSURL:          srv.URL + "/jwks",
		ClientID:         aud,
		CacheTTL:         10 * time.Second, // 길게 — TTL 안 만료
		MaxStaleDuration: 1 * time.Hour,
	}

	// metric emission 기준선 — fresh cache hit 분기는 stale fallback 진입하지
	// 않으므로 ok / fail 둘 다 delta 0 이어야 한다 (회귀 가드).
	initJWKSMetrics()
	beforeOK := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"ok"})
	beforeFail := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"fail"})
	beforeAge := histogramSampleCount(t, jwksStaleAgeSeconds)

	// 3회 호출 — 1차 fetch + 2/3차 cache hit. JWKS endpoint call count = 1
	for i := 0; i < 3; i++ {
		if _, err := v.VerifyBearerToken(context.Background(), token); err != nil {
			t.Fatalf("call %d err: %v", i, err)
		}
	}
	if jwksCallCount.Load() != 1 {
		t.Errorf("fresh cache hit 시 JWKS network call = 1 expected, got %d", jwksCallCount.Load())
	}

	// fresh path 회귀 가드 — stale metric 어느 쪽도 증가하지 않아야 한다.
	if delta := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"ok"}) - beforeOK; delta != 0 {
		t.Errorf("devhub_jwks_stale_while_error_total{result=ok} delta = %v; want 0 (fresh path 회귀)", delta)
	}
	if delta := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"fail"}) - beforeFail; delta != 0 {
		t.Errorf("devhub_jwks_stale_while_error_total{result=fail} delta = %v; want 0 (fresh path 회귀)", delta)
	}
	if delta := histogramSampleCount(t, jwksStaleAgeSeconds) - beforeAge; delta != 0 {
		t.Errorf("devhub_jwks_stale_age_seconds sample delta = %v; want 0 (fresh path 회귀)", delta)
	}
}

// TestKeycloakJWKSVerifier_StaleCutoff_BasedOnTTLExpiry — PR #242 codex P1
// hotfix 회귀. MaxStaleDuration 의 의미는 "TTL 만료 후 grace period". 즉
// stale 사용 가능 window = cachedUntil + MaxStaleDuration. 본 회귀 test 는
// CacheTTL=200ms + MaxStaleDuration=200ms 로 (이전 잘못된 cutoff = cachedAt +
// 200ms 면 250ms 대기 후 401. 정공법 = cachedUntil + 200ms = cachedAt + 400ms
// 까지 stale 사용 → 250ms 대기 후 stale 통과).
func TestKeycloakJWKSVerifier_StaleCutoff_BasedOnTTLExpiry(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	kid := "kid-cutoff-1"
	issuer := "https://issuer.example.com/realms/devhub"
	aud := "devhub-web"

	var jwksCallCount atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		jwksCallCount.Add(1)
		if jwksCallCount.Load() == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"keys": []map[string]any{{
					"kty": "RSA",
					"kid": kid,
					"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}),
				}},
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	token := mustSignToken(t, key, kid, jwt.MapClaims{
		"iss":                issuer,
		"aud":                aud,
		"sub":                "user-cutoff",
		"preferred_username": "dave",
		"exp":                time.Now().Add(5 * time.Minute).Unix(),
	})

	v := &KeycloakJWKSVerifier{
		IssuerURL:        issuer,
		JWKSURL:          srv.URL + "/jwks",
		ClientID:         aud,
		CacheTTL:         200 * time.Millisecond,
		MaxStaleDuration: 200 * time.Millisecond,
	}

	// 1차 호출 — cache 채움 (T=0). cachedAt ≈ 0, cachedUntil ≈ 200ms.
	if _, err := v.VerifyBearerToken(context.Background(), token); err != nil {
		t.Fatalf("1차 호출: %v", err)
	}

	// T=250ms 대기. TTL 만료 (200ms) 후 50ms 경과.
	// 정공법: stale window = cachedUntil(200ms) + maxStale(200ms) = 400ms → 250ms 안전 (stale 통과).
	// 이전 잘못된 cutoff (cachedAt + maxStale = 200ms 만 사용 가능) → 250ms 초과로 401.
	time.Sleep(250 * time.Millisecond)

	// metric emission 기준선 — stale window 안 진입 → ok label 1 + age sample 1.
	initJWKSMetrics()
	beforeOK := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"ok"})
	beforeAge := histogramSampleCount(t, jwksStaleAgeSeconds)

	actor, err := v.VerifyBearerToken(context.Background(), token)
	if err != nil {
		t.Fatalf("stale window 안인데 401 (PR #242 codex P1 회귀): %v", err)
	}
	if actor.Login != "dave" {
		t.Errorf("actor.Login = %q; want dave", actor.Login)
	}

	// PR #242 회귀 가드의 metric 면 — stale window 안에서 fetch fail → ok 1 증가.
	if delta := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"ok"}) - beforeOK; delta != 1 {
		t.Errorf("devhub_jwks_stale_while_error_total{result=ok} delta = %v; want 1", delta)
	}
	if delta := histogramSampleCount(t, jwksStaleAgeSeconds) - beforeAge; delta != 1 {
		t.Errorf("devhub_jwks_stale_age_seconds sample delta = %v; want 1", delta)
	}
}

// TestKeycloakJWKSVerifier_StaleFallback_DefaultMaxStale — MaxStaleDuration
// 0 (unset) 시 internal default (24h) 적용 검증. 짧은 CacheTTL 후 fetch fail
// 에도 stale 사용 성공.
func TestKeycloakJWKSVerifier_StaleFallback_DefaultMaxStale(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	kid := "kid-default-1"
	issuer := "https://issuer.example.com/realms/devhub"
	aud := "devhub-web"

	var jwksCallCount atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		jwksCallCount.Add(1)
		if jwksCallCount.Load() == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"keys": []map[string]any{{
					"kty": "RSA",
					"kid": kid,
					"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}),
				}},
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	token := mustSignToken(t, key, kid, jwt.MapClaims{
		"iss":                issuer,
		"aud":                aud,
		"sub":                "user-default",
		"preferred_username": "carol",
		"exp":                time.Now().Add(5 * time.Minute).Unix(),
	})

	v := &KeycloakJWKSVerifier{
		IssuerURL: issuer,
		JWKSURL:   srv.URL + "/jwks",
		ClientID:  aud,
		CacheTTL:  10 * time.Millisecond,
		// MaxStaleDuration 미설정 — defaultJWKSMaxStale (24h) 적용
	}

	if _, err := v.VerifyBearerToken(context.Background(), token); err != nil {
		t.Fatalf("1차 호출: %v", err)
	}
	time.Sleep(20 * time.Millisecond)

	// metric emission 기준선 — default MaxStaleDuration(24h) 적용 후 ok 1 증가.
	initJWKSMetrics()
	beforeOK := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"ok"})
	beforeAge := histogramSampleCount(t, jwksStaleAgeSeconds)

	actor, err := v.VerifyBearerToken(context.Background(), token)
	if err != nil {
		t.Fatalf("default MaxStaleDuration 안에 stale fallback 실패: %v", err)
	}
	if actor.Login != "carol" {
		t.Errorf("actor.Login = %q; want carol", actor.Login)
	}

	// MaxStaleDuration 미설정이라도 internal default(24h) 적용 → ok 분기 진입 검증.
	if delta := counterVecValue(t, jwksStaleWhileErrorTotal, []string{"ok"}) - beforeOK; delta != 1 {
		t.Errorf("devhub_jwks_stale_while_error_total{result=ok} delta = %v; want 1", delta)
	}
	if delta := histogramSampleCount(t, jwksStaleAgeSeconds) - beforeAge; delta != 1 {
		t.Errorf("devhub_jwks_stale_age_seconds sample delta = %v; want 1", delta)
	}
}
