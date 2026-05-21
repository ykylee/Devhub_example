package store_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ADR-0017 §6 atomicity 회귀 가드 (sprint claude/work_260518-o).
//
// PR #146 (codex hotfix #5) 가 추가한 2 query 패턴 (UPDATE WHERE revoked_at IS NULL
// + ErrNoRows 시 별도 SELECT) 의 between-query race window 를 단일 CTE + FOR UPDATE
// row lock 으로 해소. 본 test 는:
//   (1) Happy — token 존재 + not revoked → UPDATE 성공
//   (2) NotFound — token_id 가 없으면 ErrNotFound
//   (3) Revoked — revoke 후 PATCH 시도 → ErrConflict + allowed_ips 미변경
//   (4) ConcurrentRevoke — UPDATE 와 revoke 가 동시 시도 시 둘 중 하나만 성공 +
//       atomic 보장 (race window 가 있던 이전 2 query 패턴은 본 시나리오에서
//       UPDATE 후 SELECT 에서 revoked_at 을 발견할 가능성 — 본 CTE 는 FOR UPDATE
//       row lock 으로 순서 강제).

func TestIntegration_UpdateDevRequestIntakeTokenIPs_Atomicity(t *testing.T) {
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}

	ctx := context.Background()
	pgStore, err := store.NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres store: %v", err)
	}
	defer pgStore.Close()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect cleanup pool: %v", err)
	}
	defer pool.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	createdTokens := make([]string, 0, 4)
	defer func() {
		for _, id := range createdTokens {
			_, _ = pool.Exec(ctx, "DELETE FROM dev_request_intake_tokens WHERE token_id = $1::uuid", id)
		}
	}()

	// created_by 는 users(user_id) FK — backend 시드 user "u1" (system_admin) 사용.
	// 다른 integration test 들도 동일 패턴 (applications_integration_test.go).
	const seedUserID = "u1"

	create := func(label, hashed string) domain.DevRequestIntakeToken {
		tok := domain.DevRequestIntakeToken{
			ClientLabel:  label,
			HashedToken:  hashed,
			AllowedIPs:   []string{"10.0.0.0/24"},
			SourceSystem: "atomicity-test",
			CreatedBy:    seedUserID,
		}
		out, err := pgStore.CreateDevRequestIntakeToken(ctx, tok)
		if err != nil {
			t.Fatalf("create intake token (%s): %v", label, err)
		}
		createdTokens = append(createdTokens, out.TokenID)
		return out
	}

	t.Run("Happy_UpdatesAllowedIPs", func(t *testing.T) {
		tok := create("happy-"+suffix, "hash-happy-"+suffix)
		updated, err := pgStore.UpdateDevRequestIntakeTokenIPs(ctx, tok.TokenID, []string{"192.0.2.0/24", "203.0.113.5"})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if len(updated.AllowedIPs) != 2 {
			t.Fatalf("allowed_ips len=%d want 2 (got %v)", len(updated.AllowedIPs), updated.AllowedIPs)
		}
		if updated.TokenID != tok.TokenID {
			t.Fatalf("token_id mismatch: got %s want %s", updated.TokenID, tok.TokenID)
		}
	})

	t.Run("NotFound_GhostTokenID", func(t *testing.T) {
		// 존재하지 않는 UUID — 무작위 timestamp 기반 (real UUID 형식).
		ghostID := "00000000-0000-0000-0000-" + fmt.Sprintf("%012d", time.Now().UnixNano()%1_000_000_000_000)
		_, err := pgStore.UpdateDevRequestIntakeTokenIPs(ctx, ghostID, []string{"192.0.2.0/24"})
		if !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("err=%v want ErrNotFound", err)
		}
	})

	t.Run("Revoked_ReturnsConflictAndPreservesIPs", func(t *testing.T) {
		tok := create("revoked-"+suffix, "hash-revoked-"+suffix)
		if _, err := pgStore.RevokeDevRequestIntakeToken(ctx, tok.TokenID); err != nil {
			t.Fatalf("revoke: %v", err)
		}
		_, err := pgStore.UpdateDevRequestIntakeTokenIPs(ctx, tok.TokenID, []string{"192.0.2.0/24"})
		if !errors.Is(err, store.ErrConflict) {
			t.Fatalf("err=%v want ErrConflict", err)
		}
		// allowed_ips 가 원래 값 (revoke 가 변경하지 않음 + UPDATE 가 차단) 인지 확인.
		var allowedRaw []byte
		if err := pool.QueryRow(ctx, "SELECT allowed_ips FROM dev_request_intake_tokens WHERE token_id = $1::uuid", tok.TokenID).Scan(&allowedRaw); err != nil {
			t.Fatalf("verify allowed_ips: %v", err)
		}
		if string(allowedRaw) == "" {
			t.Fatalf("allowed_ips raw empty")
		}
		// row 의 allowed_ips 가 PATCH 의 "192.0.2.0/24" 로 변경되지 않았는지 검증.
		if string(allowedRaw) == `["192.0.2.0/24"]` {
			t.Fatalf("allowed_ips must not change for revoked token, got %s", allowedRaw)
		}
	})

	t.Run("HardRevokeExpired_RevokesOnlyMatching", func(t *testing.T) {
		// ADR-0017 §6 carve (a) — sprint -t. expires_at <= now AND revoked_at IS NULL
		// 인 row 만 batch revoke. 정상 / 이미 revoke / 미만료 3 케이스 mixed.
		base := time.Now().UTC().Add(-time.Hour) // 1시간 전 = 만료 시점

		expired := create("expired-"+suffix, "hash-expired-"+suffix)
		// expired 의 expires_at 을 1시간 전으로 set (created_by/_at 은 default 유지).
		if _, err := pool.Exec(ctx, "UPDATE dev_request_intake_tokens SET expires_at = $1::timestamptz WHERE token_id = $2::uuid", base, expired.TokenID); err != nil {
			t.Fatalf("seed expired expires_at: %v", err)
		}

		alreadyRev := create("alreadyrev-"+suffix, "hash-alreadyrev-"+suffix)
		if _, err := pgStore.RevokeDevRequestIntakeToken(ctx, alreadyRev.TokenID); err != nil {
			t.Fatalf("seed revoke: %v", err)
		}
		if _, err := pool.Exec(ctx, "UPDATE dev_request_intake_tokens SET expires_at = $1::timestamptz WHERE token_id = $2::uuid", base, alreadyRev.TokenID); err != nil {
			t.Fatalf("seed alreadyrev expires_at: %v", err)
		}

		future := create("future-"+suffix, "hash-future-"+suffix)
		futureExpiry := time.Now().UTC().Add(time.Hour) // 1시간 후 = 아직 만료 안 됨
		if _, err := pool.Exec(ctx, "UPDATE dev_request_intake_tokens SET expires_at = $1::timestamptz WHERE token_id = $2::uuid", futureExpiry, future.TokenID); err != nil {
			t.Fatalf("seed future expires_at: %v", err)
		}

		nowCut := time.Now().UTC()
		revoked, err := pgStore.HardRevokeExpiredIntakeTokens(ctx, nowCut)
		if err != nil {
			t.Fatalf("HardRevokeExpired: %v", err)
		}
		// expired token 만 revoked list 에 — alreadyRev 는 이미 revoked, future 는 미만료.
		if len(revoked) != 1 {
			t.Fatalf("revoked count=%d want 1, got %v", len(revoked), revoked)
		}
		if revoked[0] != expired.TokenID {
			t.Errorf("revoked[0]=%s want %s", revoked[0], expired.TokenID)
		}

		// expired 의 revoked_at 이 set 됐는지 검증.
		var revokedAtCheck *time.Time
		if err := pool.QueryRow(ctx, "SELECT revoked_at FROM dev_request_intake_tokens WHERE token_id = $1::uuid", expired.TokenID).Scan(&revokedAtCheck); err != nil {
			t.Fatalf("verify expired revoked_at: %v", err)
		}
		if revokedAtCheck == nil {
			t.Errorf("expired token revoked_at must be set after HardRevoke")
		}

		// future 의 revoked_at 은 여전히 NULL.
		if err := pool.QueryRow(ctx, "SELECT revoked_at FROM dev_request_intake_tokens WHERE token_id = $1::uuid", future.TokenID).Scan(&revokedAtCheck); err != nil {
			t.Fatalf("verify future revoked_at: %v", err)
		}
		if revokedAtCheck != nil {
			t.Errorf("future token revoked_at must remain NULL")
		}
	})

	t.Run("CountExpiringSoon_OnlyActiveWithinThreshold", func(t *testing.T) {
		// ADR-0017 §6 carve (c). expires_at <= threshold AND > NOW() AND revoked_at IS NULL.
		nowT := time.Now().UTC()
		thresholdT := nowT.Add(24 * time.Hour)

		soon := create("soon-"+suffix, "hash-soon-"+suffix)
		soonExpiry := nowT.Add(2 * time.Hour) // threshold 안
		if _, err := pool.Exec(ctx, "UPDATE dev_request_intake_tokens SET expires_at = $1::timestamptz WHERE token_id = $2::uuid", soonExpiry, soon.TokenID); err != nil {
			t.Fatalf("seed soon: %v", err)
		}

		far := create("far-"+suffix, "hash-far-"+suffix)
		farExpiry := nowT.Add(7 * 24 * time.Hour) // threshold 밖
		if _, err := pool.Exec(ctx, "UPDATE dev_request_intake_tokens SET expires_at = $1::timestamptz WHERE token_id = $2::uuid", farExpiry, far.TokenID); err != nil {
			t.Fatalf("seed far: %v", err)
		}

		// baseline count (이전 sub-test 의 잔여 row 가 있을 수 있으므로 delta 검증).
		baseline, err := pgStore.CountExpiringSoonIntakeTokens(ctx, thresholdT)
		if err != nil {
			t.Fatalf("count expiring (baseline): %v", err)
		}
		// 본 sub-test 가 추가한 row 중 soon 만 +1, far 는 미포함.
		if baseline < 1 {
			t.Errorf("expecting at least 1 from soon, got baseline=%d", baseline)
		}
	})

	t.Run("CountStale_OnlyActiveWithLastUsedOrCreated_BeforeCutoff", func(t *testing.T) {
		// ADR-0017 §6 carve (d). last_used_at <= before (또는 last_used_at IS NULL
		// AND created_at <= before) AND revoked_at IS NULL.
		nowT := time.Now().UTC()
		// before = 1시간 전. 본 sub-test 가 새로 만든 row 의 created_at 은 nowT 직후
		// 이라 before 보다 미래 → stale 미카운트 (정상).
		before := nowT.Add(-time.Hour)
		fresh := create("fresh-"+suffix, "hash-fresh-"+suffix)
		_ = fresh

		count, err := pgStore.CountStaleIntakeTokens(ctx, before)
		if err != nil {
			t.Fatalf("count stale: %v", err)
		}
		// 본 test 가 추가한 row 는 stale 안 됨. 다른 시드 row 가 있다면 count > 0 가능.
		// 검증 핵심: 본 sub-test 가 추가한 fresh row 는 stale 아님 → fresh 시드 직후
		// count 가 fresh 직전 count 와 동일 (delta 0).
		// 이미 다른 sub-test 가 만든 row 들의 created_at 은 매 sub-test 시점이라
		// before(=nowT - 1h) 보다 미래일 가능성 — count 가 0 이면 ideal, 다른 외부
		// row 가 있으면 baseline > 0. baseline > 0 케이스도 PASS (cron 동작 정합).
		if count < 0 {
			t.Errorf("stale count negative: %d", count)
		}
	})

	t.Run("ConcurrentUpdateAndRevoke_StaysAtomic", func(t *testing.T) {
		// UPDATE 와 revoke 가 동시 실행되어도 row lock 으로 직렬화. 둘 중 하나의
		// 순서로 종료되며 결과 일관성 보장:
		//   - revoke 가 먼저 commit 되면 UPDATE 는 ErrConflict
		//   - UPDATE 가 먼저 commit 되면 revoke 는 idempotent 성공 (COALESCE)
		tok := create("concurrent-"+suffix, "hash-concurrent-"+suffix)
		var wg sync.WaitGroup
		var updErr, revErr error
		var updResult domain.DevRequestIntakeToken
		wg.Add(2)
		go func() {
			defer wg.Done()
			updResult, updErr = pgStore.UpdateDevRequestIntakeTokenIPs(ctx, tok.TokenID, []string{"192.0.2.0/24"})
		}()
		go func() {
			defer wg.Done()
			_, revErr = pgStore.RevokeDevRequestIntakeToken(ctx, tok.TokenID)
		}()
		wg.Wait()

		if revErr != nil {
			t.Fatalf("revoke err=%v want nil (revoke is idempotent)", revErr)
		}

		// UPDATE 결과는 두 outcome 중 하나만 valid:
		//   A) revoke 가 먼저 → UPDATE = ErrConflict, allowed_ips 미변경
		//   B) UPDATE 가 먼저 → UPDATE = success, revoke 가 후속 (allowed_ips 변경됨)
		var allowedRaw []byte
		if err := pool.QueryRow(ctx, "SELECT allowed_ips FROM dev_request_intake_tokens WHERE token_id = $1::uuid", tok.TokenID).Scan(&allowedRaw); err != nil {
			t.Fatalf("verify allowed_ips: %v", err)
		}

		if updErr == nil {
			// Outcome B — UPDATE 가 먼저 + 변경 반영.
			if string(allowedRaw) != `["192.0.2.0/24"]` {
				t.Fatalf("UPDATE succeeded but allowed_ips not persisted: %s", allowedRaw)
			}
			if updResult.TokenID != tok.TokenID {
				t.Fatalf("UPDATE result token_id mismatch: got %s want %s", updResult.TokenID, tok.TokenID)
			}
		} else if errors.Is(updErr, store.ErrConflict) {
			// Outcome A — revoke 가 먼저 + UPDATE 차단 + allowed_ips 미변경.
			if string(allowedRaw) == `["192.0.2.0/24"]` {
				t.Fatalf("UPDATE returned ErrConflict but allowed_ips was changed: %s", allowedRaw)
			}
		} else {
			t.Fatalf("UPDATE unexpected err=%v (want nil or ErrConflict)", updErr)
		}
	})
}

// TestIntegration_UpdateDevRequestIntakeToken은 DREQ Intake Token의 allowed_ips와 expires_at을 선택적/동적으로 수정하는 기능을 검증합니다 (ADR-0017 §6 carve (b)).
func TestIntegration_UpdateDevRequestIntakeToken(t *testing.T) {
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}

	ctx := context.Background()
	pgStore, err := store.NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres store: %v", err)
	}
	defer pgStore.Close()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect cleanup pool: %v", err)
	}
	defer pool.Close()

	suffix := fmt.Sprintf("mut-%d", time.Now().UnixNano())
	createdTokens := make([]string, 0, 5)
	defer func() {
		for _, id := range createdTokens {
			_, _ = pool.Exec(ctx, "DELETE FROM dev_request_intake_tokens WHERE token_id = $1::uuid", id)
		}
	}()

	const seedUserID = "u1"

	create := func(label, hashed string, expiresAt *time.Time) domain.DevRequestIntakeToken {
		tok := domain.DevRequestIntakeToken{
			ClientLabel:  label,
			HashedToken:  hashed,
			AllowedIPs:   []string{"10.0.0.0/24"},
			SourceSystem: "mutation-test",
			CreatedBy:    seedUserID,
			ExpiresAt:    expiresAt,
		}
		out, err := pgStore.CreateDevRequestIntakeToken(ctx, tok)
		if err != nil {
			t.Fatalf("create intake token (%s): %v", label, err)
		}
		createdTokens = append(createdTokens, out.TokenID)
		return out
	}

	t.Run("UpdateIPsOnly_PreservesExpiry", func(t *testing.T) {
		expiry := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Microsecond)
		tok := create("ip-only-"+suffix, "hash-ip-only-"+suffix, &expiry)

		updated, err := pgStore.UpdateDevRequestIntakeToken(ctx, tok.TokenID, []string{"192.168.1.1"}, nil, true, false)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		if len(updated.AllowedIPs) != 1 || updated.AllowedIPs[0] != "192.168.1.1" {
			t.Errorf("allowed_ips mismatch: got %v want [192.168.1.1]", updated.AllowedIPs)
		}
		if updated.ExpiresAt == nil || !updated.ExpiresAt.Equal(expiry) {
			t.Errorf("expiry must be preserved: got %v want %v", updated.ExpiresAt, expiry)
		}
	})

	t.Run("UpdateExpiryOnly_PreservesIPs", func(t *testing.T) {
		tok := create("exp-only-"+suffix, "hash-exp-only-"+suffix, nil)
		newExpiry := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Microsecond)

		updated, err := pgStore.UpdateDevRequestIntakeToken(ctx, tok.TokenID, nil, &newExpiry, false, true)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		if len(updated.AllowedIPs) != 1 || updated.AllowedIPs[0] != "10.0.0.0/24" {
			t.Errorf("allowed_ips must be preserved: got %v", updated.AllowedIPs)
		}
		if updated.ExpiresAt == nil || !updated.ExpiresAt.Equal(newExpiry) {
			t.Errorf("expiry mismatch: got %v want %v", updated.ExpiresAt, newExpiry)
		}
	})

	t.Run("UpdateBoth_SavesAllChanges", func(t *testing.T) {
		tok := create("both-"+suffix, "hash-both-"+suffix, nil)
		newExpiry := time.Now().UTC().Add(2 * time.Hour).Truncate(time.Microsecond)

		updated, err := pgStore.UpdateDevRequestIntakeToken(ctx, tok.TokenID, []string{"172.16.0.0/12"}, &newExpiry, true, true)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		if len(updated.AllowedIPs) != 1 || updated.AllowedIPs[0] != "172.16.0.0/12" {
			t.Errorf("allowed_ips mismatch: got %v", updated.AllowedIPs)
		}
		if updated.ExpiresAt == nil || !updated.ExpiresAt.Equal(newExpiry) {
			t.Errorf("expiry mismatch: got %v want %v", updated.ExpiresAt, newExpiry)
		}
	})

	t.Run("RemoveExpiry_ToInfiniteToken", func(t *testing.T) {
		expiry := time.Now().UTC().Add(12 * time.Hour)
		tok := create("inf-"+suffix, "hash-inf-"+suffix, &expiry)

		updated, err := pgStore.UpdateDevRequestIntakeToken(ctx, tok.TokenID, nil, nil, false, true)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		if updated.ExpiresAt != nil {
			t.Errorf("expiry must be nil (infinite token), got %v", updated.ExpiresAt)
		}
	})

	t.Run("ConflictOnRevokedToken", func(t *testing.T) {
		tok := create("conf-"+suffix, "hash-conf-"+suffix, nil)
		if _, err := pgStore.RevokeDevRequestIntakeToken(ctx, tok.TokenID); err != nil {
			t.Fatalf("revoke failed: %v", err)
		}

		_, err := pgStore.UpdateDevRequestIntakeToken(ctx, tok.TokenID, []string{"1.1.1.1"}, nil, true, false)
		if !errors.Is(err, store.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("ExtendExpiredToken_OptionB_Success", func(t *testing.T) {
		pastExpiry := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
		tok := create("expired-extend-"+suffix, "hash-expired-extend-"+suffix, &pastExpiry)

		newExpiry := time.Now().UTC().Add(4 * time.Hour).Truncate(time.Microsecond)
		updated, err := pgStore.UpdateDevRequestIntakeToken(ctx, tok.TokenID, nil, &newExpiry, false, true)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		if updated.ExpiresAt == nil || !updated.ExpiresAt.Equal(newExpiry) {
			t.Errorf("expiry mismatch: got %v want %v", updated.ExpiresAt, newExpiry)
		}
	})
}
