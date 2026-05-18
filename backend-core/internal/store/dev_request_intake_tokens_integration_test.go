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

	create := func(label, hashed string) domain.DevRequestIntakeToken {
		tok := domain.DevRequestIntakeToken{
			ClientLabel:  label,
			HashedToken:  hashed,
			AllowedIPs:   []string{"10.0.0.0/24"},
			SourceSystem: "atomicity-test",
			CreatedBy:    "system:" + suffix,
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
