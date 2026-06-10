package httpapi

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/shared/httphelp"
)

// Cache hit: when the DevHub users row already carries a idp_subject,
// resolveIdPSubject must return that value without calling the slow
// /admin/identities scan.
func TestResolveIdPSubject_CacheHit(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	idp := &MockIdentityAdmin{
		FindError: errors.New("FindIdentityByUserID should not be called when cache is warm"),
	}
	if _, err := orgStore.CreateUser(context.Background(), domain.CreateUserInput{
		UserID:      "alice",
		Email:       "alice@example.com",
		DisplayName: "Alice",
		Role:        domain.AppRoleDeveloper,
		Status:      domain.UserStatusActive,
		Type:        domain.UserTypeHuman,
		JoinedAt:    time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := orgStore.SetIdPSubject(context.Background(), "alice", "cached-identity-id"); err != nil {
		t.Fatalf("set idp_subject: %v", err)
	}

	h := Handler{cfg: RouterConfig{OrganizationStore: orgStore, IdentityAdmin: idp}}
	id, err := h.resolveIdPSubject(context.Background(), "alice")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if id != "cached-identity-id" {
		t.Errorf("identity_id = %q, want cached-identity-id", id)
	}
	if idp.FindCalls != 0 {
		t.Errorf("FindIdentityByUserID was called %d times; cache hit should skip the scan", idp.FindCalls)
	}
}

// Lazy backfill: when the DevHub users row exists but the column is empty,
// the slow path fires once and writes the result back so the next call hits
// the cache.
func TestResolveIdPSubject_LazyBackfill(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	idp := &MockIdentityAdmin{FindIDOverride: map[string]string{"bob": "scanned-identity-id"}}
	if _, err := orgStore.CreateUser(context.Background(), domain.CreateUserInput{
		UserID:      "bob",
		Email:       "bob@example.com",
		DisplayName: "Bob",
		Role:        domain.AppRoleDeveloper,
		Status:      domain.UserStatusActive,
		Type:        domain.UserTypeHuman,
		JoinedAt:    time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	h := Handler{cfg: RouterConfig{OrganizationStore: orgStore, IdentityAdmin: idp}}

	id, err := h.resolveIdPSubject(context.Background(), "bob")
	if err != nil {
		t.Fatalf("first resolve: %v", err)
	}
	if id != "scanned-identity-id" {
		t.Errorf("first resolve identity_id = %q, want scanned-identity-id", id)
	}
	if idp.FindCalls != 1 {
		t.Errorf("first resolve should have triggered exactly one scan; FindCalls=%d", idp.FindCalls)
	}

	// After the first call the cache must be populated.
	user, err := orgStore.GetUser(context.Background(), "bob")
	if err != nil {
		t.Fatalf("re-fetch user: %v", err)
	}
	if user.IdPSubject != "scanned-identity-id" {
		t.Errorf("cache not populated; IdPSubject=%q", user.IdPSubject)
	}

	// Second call must not re-scan.
	id, err = h.resolveIdPSubject(context.Background(), "bob")
	if err != nil {
		t.Fatalf("second resolve: %v", err)
	}
	if id != "scanned-identity-id" {
		t.Errorf("second resolve identity_id = %q, want scanned-identity-id", id)
	}
	if idp.FindCalls != 1 {
		t.Errorf("second resolve should hit cache; FindCalls=%d", idp.FindCalls)
	}
}

// Missing user row: when the DevHub users row does not exist at all (e.g.
// tests that use bare newMemoryOrganizationStore() without CreateUser), the
// resolver falls through to the slow scan and tolerates the SetIdPSubject
// best-effort failure.
func TestResolveIdPSubject_NoUserRowFallsBackToScan(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	idp := &MockIdentityAdmin{}
	h := Handler{cfg: RouterConfig{OrganizationStore: orgStore, IdentityAdmin: idp}}

	id, err := h.resolveIdPSubject(context.Background(), "carol")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if id != "mock-k-id-carol" {
		t.Errorf("identity_id = %q, want mock-k-id-carol", id)
	}
	if idp.FindCalls != 1 {
		t.Errorf("FindIdentityByUserID should run once; FindCalls=%d", idp.FindCalls)
	}
}

// Propagates ErrIdentityNotFound so callers can return 404 instead of
// 500.
func TestResolveIdPSubject_NotFoundPropagates(t *testing.T) {
	orgStore := newMemoryOrganizationStore()
	idp := &MockIdentityAdmin{FindError: httphelp.ErrIdentityNotFound}
	h := Handler{cfg: RouterConfig{OrganizationStore: orgStore, IdentityAdmin: idp}}

	_, err := h.resolveIdPSubject(context.Background(), "ghost")
	if !errors.Is(err, httphelp.ErrIdentityNotFound) {
		t.Errorf("err = %v, want ErrIdentityNotFound", err)
	}
}

// ADR-0020 sub-carve B (sprint -i, issue #209): TestCreateAccount_EagerBackfillsIdPSubject
// 제거 — POST /api/v1/accounts 폐기. eager backfill 흐름은 lazy auto-create
// (authenticateActor, §5.2) 로 대체. 신규 lazy 시나리오 test 는 별도 commit 에서
// 추가 (CreateUser 호출 시 IdPSubject set 확인).
