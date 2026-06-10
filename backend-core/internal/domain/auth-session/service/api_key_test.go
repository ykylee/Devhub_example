package service

import (
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/shared/authkey"
)

// TestGenerateAPIKey_BasicShape — 32 byte random + base64url + dhk_ prefix +
// sha256 hash 일관성. ADR-0029 §6 (f) P3 + sprint plan §3.2.
func TestGenerateAPIKey_BasicShape(t *testing.T) {
	res, err := GenerateAPIKey("ci-runner", "alice", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(res.RawKey, authkey.APIKeyPrefix) {
		t.Fatalf("raw key must start with %q, got %q", authkey.APIKeyPrefix, res.RawKey)
	}
	if len(res.KeyHash) != 32 {
		t.Fatalf("key hash must be 32 bytes (sha256), got %d", len(res.KeyHash))
	}
	if len(res.APIKey.KeyPrefix) != authkey.APIKeyDisplayPrefixLength {
		t.Fatalf("key prefix must be %d chars, got %q (len=%d)", authkey.APIKeyDisplayPrefixLength, res.APIKey.KeyPrefix, len(res.APIKey.KeyPrefix))
	}
	if res.APIKey.KeyPrefix != res.RawKey[:authkey.APIKeyDisplayPrefixLength] {
		t.Fatalf("key prefix must be first %d chars of raw key", authkey.APIKeyDisplayPrefixLength)
	}
	if res.APIKey.Status != APIKeyStatusActive {
		t.Fatalf("newly generated key must be active, got %q", res.APIKey.Status)
	}
	if res.APIKey.Name != "ci-runner" {
		t.Fatalf("name not propagated: %q", res.APIKey.Name)
	}
	if res.APIKey.CreatedBy != "alice" {
		t.Fatalf("createdBy not propagated: %q", res.APIKey.CreatedBy)
	}
}

// TestGenerateAPIKey_HashMatchesAuthKey — view/auth.go 의 auth middleware 가
// 동일 hash 알고리즘 (authkey.HashAPIKey) 사용 — drift 시 silent invalid.
// 두 helper 의 hash 가 같은 raw key 에 대해 동일한지 확인.
func TestGenerateAPIKey_HashMatchesAuthKey(t *testing.T) {
	res, err := GenerateAPIKey("k", "u", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := authkey.HashAPIKey(res.RawKey)
	if len(want) != len(res.KeyHash) {
		t.Fatalf("hash length mismatch: want %d, got %d", len(want), len(res.KeyHash))
	}
	for i := range want {
		if want[i] != res.KeyHash[i] {
			t.Fatalf("hash byte %d differs: want %02x, got %02x", i, want[i], res.KeyHash[i])
		}
	}
}

// TestGenerateAPIKey_RejectsEmpty — name / createdBy empty 면 거부. 보안 +
// audit 추적 invariant.
func TestGenerateAPIKey_RejectsEmpty(t *testing.T) {
	if _, err := GenerateAPIKey("", "alice", nil, nil); err == nil {
		t.Fatal("expected error on empty name")
	}
	if _, err := GenerateAPIKey("name", "", nil, nil); err == nil {
		t.Fatal("expected error on empty createdBy")
	}
	if _, err := GenerateAPIKey("   ", "alice", nil, nil); err == nil {
		t.Fatal("expected error on whitespace name")
	}
	if _, err := GenerateAPIKey("name", "   ", nil, nil); err == nil {
		t.Fatal("expected error on whitespace createdBy")
	}
}

// TestGenerateAPIKey_EntropyUniqueness — 1000회 생성 시 모두 unique 한 raw
// key. crypto/rand 256bit 으로 collision 0 보장 — regression 시 즉시 감지.
func TestGenerateAPIKey_EntropyUniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		res, err := GenerateAPIKey("k", "u", nil, nil)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if _, dup := seen[res.RawKey]; dup {
			t.Fatalf("duplicate raw key at iteration %d: %q", i, res.RawKey)
		}
		seen[res.RawKey] = struct{}{}
	}
}

// TestGenerateAPIKey_PassesThroughOptionalFields — expires_at + allowed_cidrs
// 가 그대로 APIKey 에 보존되는지.
func TestGenerateAPIKey_PassesThroughOptionalFields(t *testing.T) {
	exp := "2027-01-01T00:00:00Z"
	cidrs := []string{"10.0.0.0/8", "192.168.0.0/16"}
	res, err := GenerateAPIKey("k", "u", &exp, cidrs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.APIKey.ExpiresAt == nil || *res.APIKey.ExpiresAt != exp {
		t.Fatalf("expires_at not propagated: %v", res.APIKey.ExpiresAt)
	}
	if len(res.APIKey.AllowedCIDRs) != 2 {
		t.Fatalf("allowed_cidrs not propagated: %v", res.APIKey.AllowedCIDRs)
	}
}
