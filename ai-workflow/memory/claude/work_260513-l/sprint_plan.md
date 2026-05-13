# Sprint Plan — claude/work_260513-l (M3 진입 — RM-M3-01..03)

- 문서 목적: M3 마일스톤 1차 진입 sprint. drift 정합 직후 (PR #97 머지) 의 첫 M3 sprint.
- 진입 base: main HEAD `3d7d5a2` (PR #97 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. 현황 분석

PoC 구현 이미 존재:
- `internal/hrdb/mock.go`: HR DB Client interface + `MockClient` (3 시드 인원: yklee/akim/sjones).
- `internal/httpapi/auth_signup.go`: `POST /api/v1/auth/signup` handler — hrdb.Lookup → Kratos identity 생성 → DevHub user 생성. publicAPIPaths 등록됨.

잔여 (M3 closing 후보):
- **단위테스트 부재**: `auth_signup_test.go` 미존재.
- **audit emit 부재**: 새 user 생성에 대한 audit log 미작성 (`account.signup.requested` action 필요).
- **§11.5.2 detail spec 부재**: `backend_api_contract.md` §11.5 의 signup 행만 1줄 — request/response 본격 spec 없음.
- **production HRDB 어댑터**: PoC `MockClient` 만. 실 인사 DB 어댑터 결정 부재.
- **조직 polish**: §10.4 의 endpoint 별 자세한 schema 부재 (현재 endpoint 표 + 권한 + envelope 1차만).

## 2. 작업 항목

### A. RM-M3-01 Sign Up 정합화

1. `backend_api_contract.md` §11.5.2 신설 — `POST /api/v1/auth/signup` 본격 spec (요청 schema + 응답 + 에러 매트릭스).
2. `auth_signup.go` 에 audit emit (`account.signup.requested` action) 추가.
3. `auth_signup_test.go` 신규 — 단위테스트 4 case:
   - HRDB 조회 성공 → Kratos identity 생성 + DevHub user 생성 + audit log
   - HRDB miss → 403 + `code=hr_lookup_failed`
   - Kratos identity 생성 실패 → 500 + audit log 미작성
   - DevHub user 충돌 (이미 존재) → 응답은 created 이지만 fmt.Printf 잔재 정리

### B. RM-M3-02 + ADR-0008 (production HRDB 어댑터 결정)

`docs/adr/0008-hrdb-production-adapter.md` 신규 — 표준 §1~§7 양식.

검토 옵션:
- **A. (선택)** PostgreSQL `hrdb` schema (별도 schema 또는 별도 table) + migration + `PostgresClient` 구현
- B. REST API 어댑터 (사내 HR REST API 연동)
- C. LDAP / AD 어댑터
- D. MockClient 유지 (PoC 영구화)

결정만 발급, **실 PostgresClient 구현은 carve out** (M3 후속 sprint 또는 별도 인프라 sprint).

### C. RM-M3-03 조직 polish — §10.4 보강 (1차)

`backend_api_contract.md` §10.4 의 8 endpoint 에 각 request body / response example / 에러 매트릭스 1차 추가. parent_id 검증 (cycle 방지) / primary_dept 자동 판정 코드 구현은 carve out.

### D. 매트릭스 갱신

- §2.4 IMPL-hrdb-01 책임 정의 (`internal/hrdb/mock.go::MockClient` + `Client` interface).
- §3 회원가입 행: `hrdb-01 (placeholder)` → 책임 정의 명시.
- §3 조직 계층 행: §10.4 spec 보강 표기.
- §4 ADR 인덱스에 ADR-0008.
- §5.2 의 \"Backend AI placeholder\" 같은 형식으로 \"HRDB production adapter\" open 명시.
- §6 변경 이력.

### E. main flat sync (PR #97 흡수) + PR + 2-pass + merge

## 3. carve out (M3 후속 sprint)

- ADR-0008 PostgreSQL `PostgresClient` 실 구현 + migration
- parent_id cycle 검증 코드 + primary_dept 자동 판정 코드
- 1:N 파견/겸임 테이블 + total_count Materialized View

## 4. 검증

- `go test ./internal/httpapi/...` PASS (신규 signup test 4 case 포함)
- frontend 영향 0 (docs only — frontend signup form 은 별도)
- CI 4 잡 SUCCESS
