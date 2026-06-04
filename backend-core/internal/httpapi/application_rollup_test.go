package httpapi

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

// Platform 롤업 handler tests (API-57, sprint claude/work_260514-c).

// 1) GET /platforms/:id/rollup — happy (equal policy default).
func TestPlatformRollup_DefaultEqual(t *testing.T) {
	store := newMemoryPlatformStore()
	store.platforms["some-id"] = domain.Platform{ID: "some-id", Key: "APP1", Status: domain.PlatformStatusActive}
	router := newPlatformsRouter(store)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms/some-id/rollup", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !bytes.Contains([]byte(body), []byte(`"weight_policy":"equal"`)) {
		t.Errorf("expected default equal policy: %s", body)
	}
}

// 2) GET /platforms/:id/rollup — invalid weight_policy → 400.
func TestPlatformRollup_InvalidPolicy(t *testing.T) {
	store := newMemoryPlatformStore()
	store.platforms["some-id"] = domain.Platform{ID: "some-id", Key: "APP1", Status: domain.PlatformStatusActive}
	router := newPlatformsRouter(store)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms/some-id/rollup?weight_policy=bogus", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 3) GET /platforms/:id/rollup — custom weights summing to ≠ 1.0 → 422.
func TestPlatformRollup_CustomSumMismatch(t *testing.T) {
	store := newMemoryPlatformStore()
	store.platforms["some-id"] = domain.Platform{ID: "some-id", Key: "APP1", Status: domain.PlatformStatusActive}
	router := newPlatformsRouter(store)
	weights := url.QueryEscape(`{"team/a":0.4,"team/b":0.3}`) // sum=0.7
	rec := doJSON(t, router, http.MethodGet,
		"/api/v1/platforms/some-id/rollup?weight_policy=custom&custom_weights="+weights, "")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid_weight_policy"`)) {
		t.Errorf("expected invalid_weight_policy: %s", rec.Body.String())
	}
}

// 4) GET /platforms/:id/rollup — malformed custom_weights JSON → 400.
func TestPlatformRollup_MalformedCustomWeights(t *testing.T) {
	store := newMemoryPlatformStore()
	store.platforms["some-id"] = domain.Platform{ID: "some-id", Key: "APP1", Status: domain.PlatformStatusActive}
	router := newPlatformsRouter(store)
	rec := doJSON(t, router, http.MethodGet,
		"/api/v1/platforms/some-id/rollup?weight_policy=custom&custom_weights=not-json", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 5) PR #107 codex review P1 회귀 guard — custom_weights fallback 후 합이 1.0 으로
// 정규화되어야 함. memoryPlatformStore 의 ComputePlatformRollup stub 은 실제
// 정규화 로직을 흉내내지 않으므로 본 회귀 guard 는 실 store (PostgresStore) 의
// repository_ops.go::ComputePlatformRollup 단위 호출로 검증해야 한다. 본 sprint 는
// store integration test 가 carve out 이므로, normalize 로직의 핵심 분기인 application
// 측 호출 (handler → store) 의 흐름은 happy path 만 검증. 실 정규화 검증은 후속
// integration test 에서 SQL pool 을 갖춘 환경에서 수행.
//
// 본 test 는 handler 의 custom_weights query 파싱이 store stub 으로 정상 전달되는지만
// 확인 — fallback 가 0건인 trivial case.
func TestPlatformRollup_CustomWeightsExact(t *testing.T) {
	store := newMemoryPlatformStore()
	store.platforms["some-id"] = domain.Platform{ID: "some-id", Key: "APP1", Status: domain.PlatformStatusActive}
	router := newPlatformsRouter(store)
	weights := `{"team/a":1.0}` // exact 1.0
	rec := doJSON(t, router, http.MethodGet,
		"/api/v1/platforms/some-id/rollup?weight_policy=custom&custom_weights="+weights, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s (expected 200 — exact sum 1.0)", rec.Code, rec.Body.String())
	}
}
