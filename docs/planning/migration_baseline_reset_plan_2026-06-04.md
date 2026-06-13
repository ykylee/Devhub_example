# DB Migration Baseline Reset 계획 (2026-06-04)

- 문서 목적: 누적된 `backend-core/migrations` 체인을 최종 스키마 기준 baseline 으로 재구성하기 위한 실행 계획을 정의한다.
- 범위: `backend-core/migrations` 재편, 관련 seed 정리, CI/스크립트 영향 점검, 문서/추적성 동기화, 검증 게이트 수립.
- 대상 독자: 구현 담당자, 리뷰어, QA, 후속 에이전트.
- 상태: draft
- 최종 수정일: 2026-06-04
- 관련 문서: [v0.1.0 릴리즈 로드맵](/Users/yklee/repos/Devhub_example_codex/docs/planning/release_v0-1_roadmap.md), [문서 작성·관리 표준](/Users/yklee/repos/Devhub_example_codex/docs/governance/document-standards.md), [추적성 매트릭스](/Users/yklee/repos/Devhub_example_codex/docs/traceability/report.md), [마이그레이션 README](/Users/yklee/repos/Devhub_example_codex/backend-core/migrations/README.md)

## 1. 배경

- 현재 백엔드 마이그레이션은 `000001` 부터 `000047` 까지 누적되어 있다.
- 운영 데이터는 아직 없으므로, 과거 upgrade 경로 보존보다 현재 최종 스키마를 기준으로 baseline 을 재정의하는 편이 비용 대비 이점이 크다.
- 로컬 검증 결과, 현재 체인 `000001..000047` 는 빈 PostgreSQL DB 에서 끝까지 정상 replay 된다.
- 따라서 이번 작업의 핵심 리스크는 스키마 생성 실패보다도 seed 누락, CI 스크립트 정합, 문서의 migration 번호 참조 drift 에 있다.

## 2. 목표

1. `backend-core/migrations` 를 최종 상태 기준의 짧은 baseline 체인으로 재구성한다.
2. 개발/CI/staging 환경에서 빈 DB bootstrap 이 baseline 체인만으로 정상 동작하도록 만든다.
3. 기존 migration 번호를 직접 참조하는 source-of-truth 문서와 추적성 문서를 정리한다.
4. 이번 변경 이후 새 migration 추가 규칙을 단순화한다.

## 3. 비목표

- 기존 DB 를 새 baseline 으로 in-place upgrade 하는 경로 제공.
- historical migration numbering (`000001..000047`) 을 영구적으로 runtime source-of-truth 로 유지.
- 이번 작업에서 신규 기능 추가 또는 schema 의미 변경.
- 운영 데이터 마이그레이션 도구 작성.

## 4. 전제와 결정

### 4.1 확정 전제

- 운영 데이터 없음.
- 개발/CI/staging DB 는 재생성 가능.
- `golang-migrate` 기반 적용 방식은 유지.
- 현재 최종 스키마와 필수 seed 가 baseline 의 source-of-truth 다.

### 4.2 사전 결정

1. baseline 은 schema 와 seed 를 분리한다.
   - `000001_initial_schema`
   - `000002_seed_system_rbac`
   - `000003_seed_bootstrap_catalog`
2. 기존 `000001..000047` 는 즉시 삭제하지 않고, 1차 구현 시점에는 별도 legacy 보관 디렉터리로 이동하는 방식을 우선 검토한다.
3. 이번 작업의 성공 기준은 "빈 DB 에서 baseline 만으로 테스트/CI 통과" 이다.

## 5. 대상 범위

### 5.1 코드/스크립트

- `backend-core/migrations/`
- `Makefile`
- `scripts/check-migration-uniqueness.sh`
- `scripts/setup-test-db.sh`
- `.github/workflows/ci.yml`

### 5.2 문서

- `backend-core/migrations/README.md`
- `docs/traceability/report.md`
- `docs/governance/code-taxonomy.md`
- migration 번호를 직접 서술 근거로 사용하는 ADR / 도메인 문서 / 분석 문서

## 6. 구현 전략

### 6.1 Phase A — 최종 상태 고정

1. 현재 `000001..000047` 의 최종 DB 상태를 baseline 입력으로 확정한다.
2. schema 와 seed 를 구분한다.
3. 다음 항목을 별도 체크리스트로 추출한다.
   - table
   - index
   - constraint
   - materialized view
   - default value / CHECK
   - 필수 seed

### 6.2 Phase B — baseline SQL 재구성

1. `000001_initial_schema.up.sql`
   - 최종 상태에 존재하는 schema object 만 포함
   - 중간에 생겼다가 삭제된 컬럼/인덱스는 제외
2. `000001_initial_schema.down.sql`
   - 생성 역순 DROP
3. `000002_seed_system_rbac.up.sql`
   - `users.role -> rbac_policies.role_id` FK 를 만족시키기 위해 bootstrap seed 보다 먼저 적용
   - 최종 RBAC 역할/권한 seed 통합
   - 기존 `000005`, `000018`, `000021`, `000024`, `000026`, `000047` 의 결과를 최종 상태 기준으로 흡수
4. `000002_seed_system_rbac.down.sql`
   - seed rollback 담당
5. `000003_seed_bootstrap_catalog.up.sql`
   - 기존 bootstrap 동작 유지에 필요한 seed 통합
   - 대상: `org_units`, `users`, `unit_appointments`, `scm_providers`
   - integration / e2e / 조직도 관련 테스트가 기대하는 기본 데이터 유지
6. `000003_seed_bootstrap_catalog.down.sql`
   - bootstrap seed rollback 담당

### 6.3 Phase C — legacy migration 정리

1. 기존 `000001..000047` 파일을 legacy 디렉터리로 이동할지, 저장소에서 제거할지 결정한다.
2. 1차 구현에서는 보수적으로 legacy 보관을 우선한다.
3. runtime 이 legacy 디렉터리를 읽지 않도록 스크립트/CI 경로를 명확히 유지한다.

### 6.4 Phase D — 툴링/CI 정합

1. `Makefile` 의 `migrate-create`, `migrate-up`, `migrate-down`, `lint-migrations` 경로 확인
2. `scripts/check-migration-uniqueness.sh` 가 baseline 이후에도 동일 규칙으로 동작하는지 확인
3. `scripts/setup-test-db.sh` 와 GitHub Actions 의 migrate step 이 새 체인 기준으로 그대로 통과하는지 검증
4. 필요 시 "DB 재생성 필요" 가이드를 문서화

### 6.5 Phase E — 문서/추적성 정리

1. migration 번호를 구현 근거로 직접 쓰는 문서를 식별한다.
2. 아래 세 분류로 나눠 정리한다.
   - 최종 번호를 새 baseline 번호로 치환해야 하는 문서
   - "historical context" 로 남겨야 하는 문서
   - 굳이 번호가 필요 없어서 서술을 일반화할 수 있는 문서
3. traceability 문서는 구현 근거가 깨지지 않도록 우선 정리한다.

## 7. 주요 리스크

### 7.1 Seed 누락

- 가장 큰 리스크.
- 특히 RBAC permission matrix, system role description, 제약 조건과 seed 간 정합 누락 시 런타임보다 테스트에서 먼저 깨질 수 있다.

### 7.2 문서 drift

- source-of-truth 문서가 기존 migration 번호를 대량 참조하고 있다.
- 구현은 맞는데 문서만 오래된 상태가 되면 후속 sprint 복원 비용이 커진다.

### 7.3 Down migration 품질

- 실제 운영 rollback 빈도는 낮더라도 baseline down 은 최소 일관성을 가져야 한다.
- schema object drop 순서와 seed rollback 범위는 명시적으로 검증한다.

### 7.4 Hidden dependency

- 테스트 fixture, store 주석, 분석 문서 중 일부가 특정 migration 번호를 전제로 설명을 갖고 있을 수 있다.
- 코드 동작에는 영향이 없어도 리뷰 혼선을 줄 수 있으므로 정리 대상에 포함한다.

## 8. 검증 게이트

### 8.1 필수 검증

1. 빈 임시 DB 에 baseline 만 적용
2. `schema_migrations` 최종 버전 확인
3. `cd backend-core && go test -count=1 -run 'TestIntegration_' ./internal/store/...`
4. `cd backend-core && go test ./...`

### 8.2 권장 검증

1. `cd frontend && npm run test`
2. `cd frontend && npm run build`
3. CI workflow 상 migrate step 재실행

### 8.3 비교 검증

- 기존 체인 기반 임시 DB 와 새 baseline 기반 임시 DB 간 주요 object 수 비교
- 최소 비교 항목:
  - public tables
  - indexes
  - constraints
  - materialized views

## 9. 완료 조건

다음 조건을 모두 만족하면 이번 작업을 완료로 본다.

1. `backend-core/migrations` 의 active 체인이 baseline 형태로 단축되었다.
2. 빈 DB bootstrap 이 로컬에서 성공한다.
3. backend integration + backend unit test 가 통과한다.
4. 문서/추적성의 핵심 참조 위치가 새 baseline 체인과 정합하다.
5. 개발/CI/staging DB 재생성 필요사항이 문서화되었다.

## 10. 실행 순서 체크리스트

1. baseline 입력 추출
2. 새 migration 파일 작성
3. legacy migration 보관/제거
4. 스크립트/CI 경로 검토
5. 빈 DB replay 검증
6. backend tests
7. frontend smoke 검증
8. 문서/추적성 갱신
9. 최종 diff 리뷰

## 11. 후속 메모

- 이번 작업은 schema 의미를 바꾸는 변경이 아니라 "migration history 재편" 이다.
- 따라서 PR 설명과 추적성에는 기능 추가가 아니라 bootstrap/운영성 개선으로 명확히 적어야 한다.
- staging 이라도 기존 DB 를 유지할 이유가 없으면 새 DB 재생성을 기본 경로로 잡는다.
