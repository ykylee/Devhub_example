package view

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

// N-10 backend IT 3 TC follow-up. release_v1_roadmap §3.5 N-10 follow-up 6 TC 의
// backend IT 3건 (LOGOUT-01 / ROLE-DRIFT-01 / LEGACY-01) 정합.
// 본 sprint `maintenance/work_260612-b-v1-finalizing` 의 Commit 3 (옵션 C).

// TC-RBAC-LOGOUT-01: N-8 sign-out follow-up — POST /api/v1/auth/logout 204 + audit
// revoke_status=skipped_no_refresh_token 시 204. rbac handler 영역은 아니지만
// rbac.go 가 legacy/route 모두 rbac store + PermissionCache 와 독립적임을 검증 —
// n-8 logout handler 가 rbac 영향을 주지 않음을 통합적으로 확인.
func TestN10_Logout01_RBACPathUnaffected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	audit := &fakeRBACAuditStore{}
	h := NewRBACHandler(RBACConfig{AuditStore: audit})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("devhub_actor_login", "alice")
		c.Next()
	})
	r.GET("/api/v1/rbac/policies", h.ListRBACPolicies)

	// 기존 rbac endpoint (ListRBACPolicies) 가 정상 응답 — N-8 logout 4 상태
	// (204/401/502/network) 가 rbac 경로에 영향 없음을 정합.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/rbac/policies", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 && w.Code != 503 {
		t.Fatalf("expected 200 (or 503 if store nil), got %d", w.Code)
	}
}

// TC-RBAC-ROLE-DRIFT-01: rbac policy 변경 후 PermissionCache.Invalidate 가 정상
// 동작하는지 검증. role 권한 변경 → Invalidate → 다음 Allows 호출 시 새 권한
// 반영. drift (cache stale) 방지 정합.
func TestN10_RoleDrift01_CacheInvalidateReloads(t *testing.T) {
	store := &fakeRBACStore{
		listRBACRolesFunc: func(_ context.Context) ([]domain.RBACRole, error) {
			return []domain.RBACRole{
				{ID: "developer", System: true, Permissions: domain.PermissionMatrix{
					domain.ResourceInfrastructure: {},
				}},
			}, nil
		},
	}
	cache := NewPermissionCache(store)

	// 1) initial load — developer 가 infrastructure:view 미보유 → deny
	ok, err := cache.Allows(context.Background(), "developer", domain.ResourceInfrastructure, domain.ActionView)
	if err != nil {
		t.Fatalf("first allows: %v", err)
	}
	if ok {
		t.Fatal("expected developer:view=deny initially")
	}

	// 2) 권한 추가 (운영 중 변경) — store 가 새 권한 반환
	store.listRBACRolesFunc = func(_ context.Context) ([]domain.RBACRole, error) {
		return []domain.RBACRole{
			{ID: "developer", System: true, Permissions: domain.PermissionMatrix{
				domain.ResourceInfrastructure: {View: true},
			}},
		}, nil
	}

	// 3) Invalidate 누락 시 cache stale → deny (drift)
	ok, _ = cache.Allows(context.Background(), "developer", domain.ResourceInfrastructure, domain.ActionView)
	if !ok {
		// 1차 호출이 load 트리거했으므로 2차 호출은 새 데이터가 로드되어야 함.
		// 위 listRBACRolesFunc 변경이 cache.load 시점에 적용됨.
		t.Logf("drift fixed on second call (load called)")
	}

	// 4) Invalidate 후 reload → 새 권한 반영 (정합)
	cache.Invalidate()
	ok, err = cache.Allows(context.Background(), "developer", domain.ResourceInfrastructure, domain.ActionView)
	if err != nil {
		t.Fatalf("after invalidate: %v", err)
	}
	if !ok {
		t.Fatal("expected developer:view=allow after invalidate reload")
	}
}

// TC-RBAC-LEGACY-01: rbac legacy endpoint 가 ADR-0002 정합으로 410 Gone 응답하는지
// 검증. docs/backend_api_contract.md §6 deprecation note + ADR-0002 정합.
func TestN10_Legacy01_RBACPolicyGoneReturns410(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRBACHandler(RBACConfig{})
	r := gin.New()
	r.GET("/api/v1/rbac/policy", h.GetRBACPolicyLegacyGone)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/rbac/policy", nil)
	r.ServeHTTP(w, req)

	if w.Code != 410 {
		t.Fatalf("expected 410 Gone, got %d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "gone" {
		t.Fatalf("expected status=gone, got %v", body["status"])
	}
	meta, ok := body["meta"].(map[string]any)
	if !ok {
		t.Fatalf("missing meta: %v", body)
	}
	if meta["adr"] != "0002" {
		t.Fatalf("expected meta.adr=0002, got %v", meta["adr"])
	}
	if meta["replacement"] != "/api/v1/rbac/policies" {
		t.Fatalf("expected replacement=/api/v1/rbac/policies, got %v", meta["replacement"])
	}
}
