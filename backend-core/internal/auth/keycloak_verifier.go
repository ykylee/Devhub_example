package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/devhub/backend-core/internal/httpapi"
	"github.com/golang-jwt/jwt/v5"
)

// KeycloakJWKSVerifier verifies JWT bearer tokens against a Keycloak-compatible
// JWKS endpoint. JWKS can be configured explicitly or discovered from issuer.
type KeycloakJWKSVerifier struct {
	IssuerURL  string
	JWKSURL    string
	ClientID   string
	HTTPClient *http.Client

	// Optional cache TTL. Zero means defaultTTL.
	CacheTTL time.Duration

	mu          sync.RWMutex
	cachedKeys  map[string]*rsa.PublicKey
	cachedUntil time.Time
}

const defaultJWKSTTL = 5 * time.Minute

// errKidMismatch — token 의 kid 가 cached JWKS 에 없을 때 reported (stale-while-error
// retry trigger). sprint -j codex review #9 (#3) 의 backend 확장 carve — keycloak_verifier
// 가 cache 유효 동안 새 kid 의 token 을 401 처리하던 동작을 1회 forced refetch + retry 로 보강.
// sprint claude/work_260519-r 구현.
var errKidMismatch = errors.New("jwks key not found")

func (v *KeycloakJWKSVerifier) VerifyBearerToken(ctx context.Context, token string) (httpapi.AuthenticatedActor, error) {
	if strings.TrimSpace(v.IssuerURL) == "" && strings.TrimSpace(v.JWKSURL) == "" {
		return httpapi.AuthenticatedActor{}, errors.New("KeycloakJWKSVerifier requires DEVHUB_OIDC_ISSUER_URL or DEVHUB_OIDC_JWKS_URL")
	}

	parserOpts := []jwt.ParserOption{jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"})}
	if issuer := strings.TrimSpace(v.IssuerURL); issuer != "" {
		parserOpts = append(parserOpts, jwt.WithIssuer(issuer))
	}
	if aud := strings.TrimSpace(v.ClientID); aud != "" {
		parserOpts = append(parserOpts, jwt.WithAudience(aud))
	}

	// 1차 시도 — cached JWKS 우선
	actor, err := v.parseWithJWKS(ctx, token, parserOpts)
	if err == nil {
		return actor, nil
	}

	// stale-while-error fallback: kid mismatch 시 cache invalidate + JWKS forced refetch + 1회 retry
	// (Keycloak key rotation 직후 새 kid token 이 cache TTL 만료까지 401 되는 문제 해소).
	// 다른 error (signature / expired / issuer / audience 등) 은 retry 안 함 — security 위협 회피.
	if errors.Is(err, errKidMismatch) {
		v.invalidateCache()
		actor, err = v.parseWithJWKS(ctx, token, parserOpts)
		if err != nil {
			return httpapi.AuthenticatedActor{}, err
		}
		return actor, nil
	}

	return httpapi.AuthenticatedActor{}, err
}

// parseWithJWKS — VerifyBearerToken 의 핵심 parse 단계 분리. retry path 의 단일 entry.
func (v *KeycloakJWKSVerifier) parseWithJWKS(ctx context.Context, token string, parserOpts []jwt.ParserOption) (httpapi.AuthenticatedActor, error) {
	keySet, err := v.fetchJWKS(ctx)
	if err != nil {
		return httpapi.AuthenticatedActor{}, err
	}

	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		if strings.TrimSpace(kid) == "" {
			return nil, errors.New("token header missing kid")
		}
		pub, ok := keySet[kid]
		if !ok {
			return nil, fmt.Errorf("%w for kid=%q", errKidMismatch, kid)
		}
		return pub, nil
	}, parserOpts...)
	if err != nil {
		// errKidMismatch 는 wrap 해서 errors.Is 가 동작하도록 — caller 가 retry 분기.
		if errors.Is(err, errKidMismatch) {
			return httpapi.AuthenticatedActor{}, err
		}
		return httpapi.AuthenticatedActor{}, fmt.Errorf("verify keycloak token: %w", err)
	}
	if !parsed.Valid {
		return httpapi.AuthenticatedActor{}, errors.New("invalid token")
	}

	subject := claimString(claims, "sub")
	if subject == "" {
		return httpapi.AuthenticatedActor{}, errors.New("token subject(sub) is required")
	}
	login := claimString(claims, "preferred_username")
	if login == "" {
		login = claimString(claims, "email")
	}
	if login == "" {
		login = subject
	}

	return httpapi.AuthenticatedActor{
		Subject: subject,
		Login:   login,
		Role:    extractKeycloakRole(claims),
	}, nil
}

type openIDConfiguration struct {
	JWKSURI string `json:"jwks_uri"`
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (v *KeycloakJWKSVerifier) fetchJWKS(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	if keys := v.readCachedKeys(); len(keys) > 0 {
		return keys, nil
	}

	jwksURL, err := v.resolveJWKSURL(ctx)
	if err != nil {
		return nil, err
	}
	body, err := v.getJSON(ctx, jwksURL)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	var doc jwksDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}

	out := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, k := range doc.Keys {
		if strings.ToUpper(strings.TrimSpace(k.Kty)) != "RSA" || strings.TrimSpace(k.Kid) == "" {
			continue
		}
		pub, err := rsaPublicKeyFromJWK(k.N, k.E)
		if err != nil {
			continue
		}
		out[k.Kid] = pub
	}
	if len(out) == 0 {
		return nil, errors.New("jwks did not contain usable RSA keys")
	}
	v.writeCachedKeys(out)
	return out, nil
}

// invalidateCache — kid mismatch 시 stale-while-error fallback 의 forced refetch trigger.
// sprint -j codex review #9 (#3) backend 확장 carve.
func (v *KeycloakJWKSVerifier) invalidateCache() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cachedKeys = nil
	v.cachedUntil = time.Time{}
}

func (v *KeycloakJWKSVerifier) readCachedKeys() map[string]*rsa.PublicKey {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if len(v.cachedKeys) == 0 || time.Now().After(v.cachedUntil) {
		return nil
	}
	out := make(map[string]*rsa.PublicKey, len(v.cachedKeys))
	for k, pk := range v.cachedKeys {
		out[k] = pk
	}
	return out
}

func (v *KeycloakJWKSVerifier) writeCachedKeys(keys map[string]*rsa.PublicKey) {
	ttl := v.CacheTTL
	if ttl <= 0 {
		ttl = defaultJWKSTTL
	}
	copyMap := make(map[string]*rsa.PublicKey, len(keys))
	for k, pk := range keys {
		copyMap[k] = pk
	}
	v.mu.Lock()
	v.cachedKeys = copyMap
	v.cachedUntil = time.Now().Add(ttl)
	v.mu.Unlock()
}

func (v *KeycloakJWKSVerifier) resolveJWKSURL(ctx context.Context) (string, error) {
	if explicit := strings.TrimSpace(v.JWKSURL); explicit != "" {
		return explicit, nil
	}
	issuer := strings.TrimSpace(v.IssuerURL)
	if issuer == "" {
		return "", errors.New("missing OIDC issuer and jwks url")
	}
	discoveryURL := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"
	body, err := v.getJSON(ctx, discoveryURL)
	if err != nil {
		return "", fmt.Errorf("fetch openid configuration: %w", err)
	}
	var cfg openIDConfiguration
	if err := json.Unmarshal(body, &cfg); err != nil {
		return "", fmt.Errorf("decode openid configuration: %w", err)
	}
	jwks := strings.TrimSpace(cfg.JWKSURI)
	if jwks == "" {
		return "", errors.New("openid configuration missing jwks_uri")
	}
	return jwks, nil
}

func (v *KeycloakJWKSVerifier) getJSON(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	client := v.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("endpoint status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func rsaPublicKeyFromJWK(nRaw, eRaw string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(nRaw))
	if err != nil {
		return nil, fmt.Errorf("decode jwk n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(eRaw))
	if err != nil {
		return nil, fmt.Errorf("decode jwk e: %w", err)
	}
	if len(nBytes) == 0 || len(eBytes) == 0 {
		return nil, errors.New("invalid jwk rsa parameters")
	}
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes).Int64()
	if e <= 1 {
		return nil, errors.New("invalid rsa exponent")
	}
	return &rsa.PublicKey{N: n, E: int(e)}, nil
}

func claimString(claims jwt.MapClaims, key string) string {
	raw, ok := claims[key]
	if !ok {
		return ""
	}
	v, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(v)
}

// devhubRolePriority — DevHub RBAC role 위계 (높을수록 우선). Keycloak token 의 multi-role
// (group composite role 가입 시 자연 발생) 시 highest priority role 선택. ADR-0019 §5.3
// codex review #9 의 'Default Group 권장 철회 + multi-role order-dependency 위험' 해소 —
// sprint claude/work_260519-q 의 backend 확장 carve 구현.
//
// 정합: rbac_policies seed (migration 000004) 의 4 role + ADR-0011 row-scoping. 알 수 없는
// role (예: Keycloak 측 임의 role) 은 priority 0 으로 처리 → 다른 known role 이 있으면 우선.
var devhubRolePriority = map[string]int{
	"system_admin": 4,
	"pmo_manager":  3,
	"manager":      2,
	"developer":    1,
}

// selectHighestPriorityRole 는 roles list 에서 DevHub priority 가 가장 높은 role 을
// 반환한다. 같은 priority 면 list 첫 번째 (Keycloak 의 token role include order 유지).
// 모든 role 이 priority 0 (unknown) 이면 list[0] 반환 (legacy fallback).
func selectHighestPriorityRole(roles []string) string {
	if len(roles) == 0 {
		return ""
	}
	best := roles[0]
	bestPriority := devhubRolePriority[best]
	for _, r := range roles[1:] {
		p := devhubRolePriority[r]
		if p > bestPriority {
			best = r
			bestPriority = p
		}
	}
	return best
}

func extractKeycloakRole(claims jwt.MapClaims) string {
	if role := claimString(claims, "role"); role != "" {
		return role
	}
	if raw, ok := claims["roles"]; ok {
		if roles := anyToStrings(raw); len(roles) > 0 {
			return selectHighestPriorityRole(roles)
		}
	}
	if raw, ok := claims["realm_access"]; ok {
		if m, ok := raw.(map[string]any); ok {
			if roles := anyToStrings(m["roles"]); len(roles) > 0 {
				return selectHighestPriorityRole(roles)
			}
		}
	}
	if raw, ok := claims["resource_access"]; ok {
		if byClient, ok := raw.(map[string]any); ok {
			for _, clientAccess := range byClient {
				m, ok := clientAccess.(map[string]any)
				if !ok {
					continue
				}
				if roles := anyToStrings(m["roles"]); len(roles) > 0 {
					return selectHighestPriorityRole(roles)
				}
			}
		}
	}
	return ""
}

func anyToStrings(v any) []string {
	list, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		s, ok := item.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
