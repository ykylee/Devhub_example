# Migration 000021 prefix 중복 충돌 — 정정 SOP

- 문서 목적: PR #167 KC-PR-E (2026-05-18, [ADR-0019](../adr/0019-keycloak-only-idp.md)) 가 추가한 `000021_rename_kratos_identity_to_idp_subject.{up,down}.sql` 과 기존 `000021_rbac_team_manager.{up,down}.sql` (PR #119, 2026-05-15) 의 prefix 중복 충돌 정정. sprint `claude/work_260519-l` 의 file rename + 사내 운영 DB 의 `schema_migrations` table 정정 SOP.
- 범위: backend-core/migrations/ 파일 rename + golang-migrate runner 동작 + 사내 운영 DB 정정. backend code 변경 없음 — schema 자체는 동일.
- 대상 독자: 운영자 (SRE / DBA), Backend 담당자, 사내 deploy 책임자.
- 상태: accepted
- 최종 수정일: 2026-05-20
- 결정 근거 sprint: `claude/work_260519-l`
- 관련 문서: [ADR-0019 Keycloak 단일화](../adr/0019-keycloak-only-idp.md), [keycloak_operations.md](./keycloak_operations.md), [backend-core/migrations/README.md](../../backend-core/migrations/README.md).

## 1. 충돌 발견

### 1.1 파일 사실

```text
backend-core/migrations/
├── 000020_application_leader_department.{up,down}.sql  ← PR #100 (M3 sprint -n)
├── 000021_rbac_team_manager.{up,down}.sql               ← PR #119 sprint -d (2026-05-15)
└── 000021_rename_kratos_identity_to_idp_subject.{up,down}.sql  ← PR #167 KC-PR-E (2026-05-18)
                                                          ↑ 동일 prefix 000021 — 충돌
```

### 1.2 golang-migrate 동작

[`backend-core/migrations/README.md:3`](../../backend-core/migrations/README.md) 에 명시 — backend 가 `golang-migrate/migrate` 사용.

`golang-migrate` 의 동일 version prefix 처리:
- filename 패턴 = `{version}_{name}.{direction}.sql`
- version = filename 의 첫 underscore 까지의 숫자 (예: `000021`)
- **동일 version 의 multiple migration 은 ambiguous** — runner 가 directory scan 시 둘 다 발견 → 실행 결과 예측 불가:
  - 경우 A: 즉시 error throw — `make migrate-up` 실패
  - 경우 B: 첫 match 만 적용 + 나머지 silent drop — silent drift

### 1.3 운영 영향

- **clean DB 환경** (신규 deploy):
  - 경우 A 시 deploy 차단
  - 경우 B 시 silent drift → `users.kratos_identity_id` 컬럼 그대로 (rename 미적용) → backend `keycloak_verifier.go` + `accounts_admin.go:123 SetIdPSubject` 의 `idp_subject` 참조가 **SQL error** (`column users.idp_subject does not exist`) → 로그인/계정 admin 동작 실패
- **이미 적용된 DB**:
  - `schema_migrations` table 의 last applied version 이 `000021` 인 경우, 실 적용된 migration 식별 불가 → manual investigation 필요

## 2. 정정 결정 — file rename

본 sprint `claude/work_260519-l` 가 file rename 으로 정정:

```text
000021_rename_kratos_identity_to_idp_subject.up.sql
  → 000030_rename_kratos_identity_to_idp_subject.up.sql

000021_rename_kratos_identity_to_idp_subject.down.sql
  → 000030_rename_kratos_identity_to_idp_subject.down.sql
```

**결정 근거**:
- 000021 prefix 는 `000021_rbac_team_manager` 가 먼저 (PR #119, 2026-05-15) → 그대로 유지
- 새 migration 은 sequence 의 last position (000028 / 000029 다음) 으로 — 000030 채택
- file 내용은 변경 없음 — schema 자체는 동일

## 3. clean DB 환경 SOP

신규 deploy 환경 (사내 staging 신설 / dev 재구축 등):

```bash
# 1. 신규 DB 생성 후 migrations 적용
make migrate-up

# 결과:
#   000001~000020 적용
#   000021 (rbac_team_manager) 적용 — team_manager realm role seed
#   000022~000029 적용 — DREQ + integration_registry + infra_service_snapshots
#   000030 (rename_kratos_identity_to_idp_subject) 적용 — users.kratos_identity_id → idp_subject
```

`schema_migrations` table 의 last version = `000030`. backend code 와 정합.

## 4. 이미 적용된 운영 DB 정정 SOP

**전제**: 사내 운영 DB 가 PR #167 머지 시점 (2026-05-18) 또는 이후 migration 적용 시도 → `schema_migrations` table 의 version 동일 prefix 처리 결과 확인 필요.

### 4.1 현황 조사 SQL

```sql
-- 1. schema_migrations 의 current version
SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 5;

-- 2. users table 의 컬럼 확인 — rename 적용 여부
SELECT column_name FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'users'
  AND column_name IN ('kratos_identity_id', 'idp_subject');

-- 3. rbac_policies 의 team_manager role 존재 여부 — 000021_rbac_team_manager 적용 여부
SELECT role_id, role_name FROM rbac_roles WHERE role_name = 'team_manager';
```

### 4.2 case A — 두 migration 모두 적용된 경우 (예상 가장 흔함)

조건:
- `schema_migrations.version` = `000021` (또는 마지막 entry)
- `users.idp_subject` 컬럼 존재 + `users.kratos_identity_id` 미존재
- `rbac_roles` 에 `team_manager` row 존재

처리:
```sql
-- schema_migrations 의 version 000021 → 000030 정정 (rename_kratos_identity 측이 실제 적용된 것을 명시)
-- 단, 000021_rbac_team_manager 측은 이미 적용됐으므로 별도 entry 필요 없음 (golang-migrate 는 single version-row 운영)
-- 본 case 의 schema_migrations 정정 = 000030 으로 마킹 + 000021 은 그대로 유지

-- 옵션 1: 단순 — 000030 entry 추가 (000021 도 그대로 둠)
INSERT INTO schema_migrations (version, dirty) VALUES (30, false)
ON CONFLICT (version) DO NOTHING;

-- 옵션 2: 명확 — 000021 → 000030 으로 update
-- (단, golang-migrate 는 version 1개만 track — UPDATE 가 자연)
UPDATE schema_migrations SET version = 30 WHERE version = 21;
```

권장 = **옵션 2** (`UPDATE`). 단일 version 추적이 golang-migrate 의 운영 모델.

### 4.3 case B — silent drift (rename 미적용)

조건:
- `users.kratos_identity_id` 컬럼 존재 + `users.idp_subject` 미존재
- backend 가 SQL error 로 동작 안 함 (이미 발견된 운영 장애)

처리:
```bash
# rename 적용 강제
psql $MIGRATE_DB_URL -f backend-core/migrations/000030_rename_kratos_identity_to_idp_subject.up.sql

# schema_migrations 에 000030 entry 추가
psql $MIGRATE_DB_URL -c "INSERT INTO schema_migrations (version, dirty) VALUES (30, false);"
```

또는 본 sprint -l merge 후 사내 운영 DB 에 `make migrate-up` 재실행 (file rename 후 sequence 정상).

### 4.4 case C — backend code 와 schema 정합 안 되는 어떤 상태

조사 후 사내 DBA + Backend 팀 합의 → manual schema 정정 + `schema_migrations` 정합.

## 5. 검증

### 5.1 clean DB 검증

```bash
# Test DB 신규 생성
createdb devhub_migration_test
MIGRATE_DB_URL=postgres://...@localhost:5432/devhub_migration_test?sslmode=disable make migrate-up

# 결과 확인
psql -d devhub_migration_test -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 3;"
# 기대: 30, 29, 28

psql -d devhub_migration_test -c "SELECT column_name FROM information_schema.columns WHERE table_name = 'users' AND column_name IN ('idp_subject', 'kratos_identity_id');"
# 기대: idp_subject (kratos_identity_id 없음)
```

### 5.2 backend 동작 검증

```bash
# 환경 변수 설정 후 backend-core 시작
cd backend-core && go run .

# /api/v1/auth/me 요청 — SQL error 없이 정상 응답 확인
curl -H "Authorization: Bearer $TEST_TOKEN" http://localhost:8080/api/v1/me
```

## 6. 잔여 carve

- **(carve)** Migration sequence numbering 정책 정정 — `golang-migrate` 의 동일 prefix 충돌 회피를 위한 CI lint (예: `find migrations -name "[0-9]*_*.up.sql" | awk -F'_' '{print $1}' | sort | uniq -d` 가 비어있어야 함). 별도 CI workflow carve.
- **(carve)** Migration squash 정책 — 005년 추가 시 squash 결정.

## 7. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-19 | 1차 작성. PR #167 KC-PR-E 의 000021 prefix 충돌 발견 + git mv 000021 → 000030 rename + clean DB + 이미 적용된 DB 정정 SOP (case A/B/C) + 검증 절차. | `claude/work_260519-l` |
