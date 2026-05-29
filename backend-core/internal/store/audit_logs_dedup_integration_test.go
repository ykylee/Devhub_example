package store_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	auditrep "github.com/devhub/backend-core/internal/domain/audit-ops/repository"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// audit_logs source_event_id dedup integration test (sprint claude/work_260519-w PR-D).
//
// CI backend-unit job 은 DEVHUB_TEST_DB_URL 미설정으로 t.Skip — postgres_*_test.go 패턴 정합.
// migration 000032 (source_event_id 컬럼 + partial UNIQUE INDEX) 가 적용된 DB 환경 필요.

// TestIntegration_AuditLogs_SourceEventID_DedupesViaUniqueIndex — 동일 (source_type,
// source_event_id) 으로 INSERT 2회 시도 시, 2번째는 ON CONFLICT DO NOTHING 으로 기존
// row 반환 + 신규 INSERT 없음 검증.
func TestIntegration_AuditLogs_SourceEventID_DedupesViaUniqueIndex(t *testing.T) {
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
		t.Fatalf("connect raw pool: %v", err)
	}
	defer pool.Close()
	aud := auditrep.NewAuditRepository(pgStore)

	// 고유 source_event_id — 다른 테스트 row 와의 collision 회피.
	sourceEventID := fmt.Sprintf("test-dedup-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_logs WHERE source_event_id = $1`, sourceEventID)
	}()

	first := domain.AuditLog{
		ActorLogin:    "alice",
		Action:        "auth.login.success",
		TargetType:    "auth",
		TargetID:      "alice",
		SourceType:    domain.AuditSourceKeycloakEvent,
		SourceEventID: sourceEventID,
		Payload:       map[string]any{"keycloak_event_type": "LOGIN", "ip_address": "10.0.0.1"},
	}
	inserted, err := aud.CreateAuditLog(ctx, first)
	if err != nil {
		t.Fatalf("first CreateAuditLog: %v", err)
	}
	if inserted.SourceEventID != sourceEventID {
		t.Fatalf("inserted.SourceEventID = %q; want %q", inserted.SourceEventID, sourceEventID)
	}

	// 동일 source_event_id 로 재시도 — 다른 ActorLogin / Payload 로도 dedup 되는지 검증.
	second := domain.AuditLog{
		ActorLogin:    "alice-dup",
		Action:        "auth.login.success",
		TargetType:    "auth",
		TargetID:      "alice",
		SourceType:    domain.AuditSourceKeycloakEvent,
		SourceEventID: sourceEventID,
		Payload:       map[string]any{"keycloak_event_type": "LOGIN", "duplicate": true},
	}
	deduped, err := aud.CreateAuditLog(ctx, second)
	if err != nil {
		t.Fatalf("second CreateAuditLog (dedup): %v", err)
	}
	// ON CONFLICT DO NOTHING 후 store 가 기존 row 를 SELECT 해 반환 — 첫 INSERT 의 ID/AuditID
	// 와 일치해야 함.
	if deduped.ID != inserted.ID {
		t.Fatalf("deduped.ID = %d; want %d (first row should be returned, not new INSERT)", deduped.ID, inserted.ID)
	}
	if deduped.ActorLogin != "alice" {
		t.Fatalf("deduped.ActorLogin = %q; want 'alice' (first row preserved, second silently dropped)", deduped.ActorLogin)
	}

	// audit_logs 테이블에 정확히 1개 row 만 존재 확인.
	var count int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit_logs WHERE source_event_id = $1`, sourceEventID).Scan(&count); err != nil {
		t.Fatalf("count audit_logs: %v", err)
	}
	if count != 1 {
		t.Fatalf("audit_logs rows = %d; want 1 (UNIQUE INDEX must dedup)", count)
	}
}

// TestIntegration_AuditLogs_EmptySourceType_NotDeduped — Stage 3 보강 정합 검증.
// SourceType 이 빈 문자열 + SourceEventID 가 nonempty 인 row 2건 INSERT 시도 시,
// partial WHERE 의 source_type IS NOT NULL 가드로 unique 제약 적용 안 됨 → 2개 row
// 정상 INSERT. (운영 path 에서는 emitter 가 SourceType 항상 set 이지만, 잘못된
// emitter 회귀 방지 검증.)
func TestIntegration_AuditLogs_EmptySourceType_NotDeduped(t *testing.T) {
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
		t.Fatalf("connect raw pool: %v", err)
	}
	defer pool.Close()
	aud := auditrep.NewAuditRepository(pgStore)

	sourceEventID := fmt.Sprintf("test-empty-type-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_logs WHERE source_event_id = $1`, sourceEventID)
	}()

	// 빈 SourceType + 동일 SourceEventID 로 2회 INSERT — partial WHERE 가
	// source_type IS NOT NULL 가드로 unique 제약 미적용 → 2 row 모두 정상 INSERT.
	row1, err := aud.CreateAuditLog(ctx, domain.AuditLog{
		ActorLogin:    "a",
		Action:        "x.y",
		SourceEventID: sourceEventID,
		// SourceType: "" — intentionally empty
	})
	if err != nil {
		t.Fatalf("first INSERT empty SourceType: %v", err)
	}
	row2, err := aud.CreateAuditLog(ctx, domain.AuditLog{
		ActorLogin:    "b",
		Action:        "x.y",
		SourceEventID: sourceEventID,
	})
	if err != nil {
		t.Fatalf("second INSERT empty SourceType: %v", err)
	}
	if row1.ID == row2.ID {
		t.Fatalf("rows must be distinct (partial WHERE source_type IS NOT NULL must exclude): id1=%d id2=%d", row1.ID, row2.ID)
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE source_event_id = $1`, sourceEventID).Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != 2 {
		t.Fatalf("audit_logs rows = %d; want 2 (empty source_type rows must NOT be unique-constrained)", count)
	}
}

// TestIntegration_AuditLogs_EmptySourceEventID_AllowsMultipleRows — partial UNIQUE
// INDEX 의 WHERE source_event_id IS NOT NULL 가 작동하는지 검증. 빈 source_event_id
// row 는 unique 제약을 받지 않아 여러 row 가 정상 INSERT.
func TestIntegration_AuditLogs_EmptySourceEventID_AllowsMultipleRows(t *testing.T) {
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
		t.Fatalf("connect raw pool: %v", err)
	}
	defer pool.Close()
	aud := auditrep.NewAuditRepository(pgStore)

	tag := fmt.Sprintf("test-empty-source-%d", time.Now().UnixNano())
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_logs WHERE action = $1`, tag)
	}()

	// 빈 SourceEventID 로 같은 action 2번 INSERT — 둘 다 성공해야 함 (partial WHERE 가
	// NULL row 를 unique 제약에서 제외).
	row1, err := aud.CreateAuditLog(ctx, domain.AuditLog{
		ActorLogin: "alice",
		Action:     tag,
		SourceType: domain.AuditSourceSystem,
	})
	if err != nil {
		t.Fatalf("first INSERT empty source_event_id: %v", err)
	}
	row2, err := aud.CreateAuditLog(ctx, domain.AuditLog{
		ActorLogin: "bob",
		Action:     tag,
		SourceType: domain.AuditSourceSystem,
	})
	if err != nil {
		t.Fatalf("second INSERT empty source_event_id: %v", err)
	}
	if row1.ID == row2.ID {
		t.Fatalf("rows must be distinct (empty source_event_id must NOT dedup): id1=%d id2=%d", row1.ID, row2.ID)
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE action = $1`, tag).Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != 2 {
		t.Fatalf("audit_logs rows = %d; want 2 (empty source_event_id rows must not be unique-constrained)", count)
	}
}
