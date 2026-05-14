# Session Handoff — claude/work_260514-a

- 브랜치: `claude/work_260514-a`
- Base: `main` @ `63a7ea2` (PR #104 머지 직후)
- 날짜: 2026-05-14
- 상태: in_progress
- 직전 sprint: `claude/work_260513-q` (PR #102 흡수, in_progress) + `codex/project_concept_design` (PR #104 머지 완료).

## Sprint scope

**Application Design 1차 + ADR-0011 평가** (mixed):

- (A) Application 도메인 backend design 1차
- (B) ADR-0011 RBAC row-scoping 옵션 평가 + 단일 결정

PR #104 가 만든 추적성 라인 (REQ-FR-APP-* → UC-APP-* → API-41~58 (planned)) 을 backend IMPL 까지 끌어내려 첫 사이클을 닫는다.

## 진입 컨텍스트 (이전 세션에서 가져옴)

### PR #104 흡수 결과
- REQ-FR-APP-001..012, REQ-FR-PROJ-000..010, REQ-NFR-PROJ-001..006 발급
- UC-APP-01..10, UC-PROJ-01..10 발급
- ERD §2.5 (Application/Project 확장) 정합 완료 — composite PK 표기 + 신규 컬럼
- `backend_api_contract.md` §13 + §13.0 placeholder API ID 인덱스 (API-41~58 (planned))
- ADR-0011 (RBAC row-scoping) placeholder (proposed) 발급
- concept.md §13.2.1 상태 전이 가드 + §13.4 weight_policy normalize 룰 (±0.001)
- `sync_error_code` 표준 사전 8종 + link 단위 최신 1건 캐시 운영 룰

### 직전 sprint 의 미해결 (handoff 인수)
- planned API → 정식 API ID 활성화 (본 sprint 1차 처리)
- ERD 기반 마이그레이션 초안 (본 sprint 1차 처리)
- 권한 정책 백엔드 인가 매핑 (system_admin/pmo_manager) — 일부 본 sprint
- ADR-0011 옵션 평가 (본 sprint 처리)

## 이번 세션 작업 계획

### Carve in
1. **마이그레이션 000012~000018** — applications / application_repositories (composite PK) / scm_providers / projects / project_members / project_integrations / pr_activities + build_runs + quality_snapshots. up/down + 인덱스.
2. **Store 인터페이스** — `internal/store/applications.go` (CRUD + status transition guard + Application-Repository link CRUD + Adapter Registry lookup).
3. **Handler stub** — `internal/httpapi/applications.go` (API-43~47 + API-48~50 의 RBAC gate + envelope + handler skeleton). 본 sprint 는 stub 만, body 는 store 호출 placeholder.
4. **ADR-0011 결정** — 옵션 A (Casbin), B (RBAC 매트릭스 확장), C (코드 검증), D (PG RLS) 의 비용/리스크/마이그레이션 영향 표 + 단일 채택. status: accepted.
5. **매트릭스 동기** — §2.2 API-41~50 의 (planned) 제거 / §2.4 신규 IMPL-application-* / §4 ADR-0011 accepted / §3 Application/Project row + §6 변경 이력.

### Carve out (다음 sprint)
- Project / rollup / integration handler 본체
- 실 SCM 어댑터 구현
- e2e TC 작성
- frontend `/admin/settings/applications` UI

## 다음 세션 우선 작업

1. **`postgres_applications.go` 본체** — store interface stub 의 `ErrNotImplemented` 메서드 16개를 PostgreSQL 쿼리로 채움.
2. **handler body + 요청 validation** — API-43..50 의 stub 을 실 응답으로 교체. PATCH 의 상태 전이 가드 (concept §13.2.1) + 422 분기 + immutable key 검증.
3. **단위테스트** — UT-application-store-XX + UT-application-handler-XX. RBAC denial 케이스 (developer/manager → 403) + system_admin 정상 응답.
4. **API-51..54 (Repository 운영 지표) 활성화** — handler stub + store stub + route 등록. body 는 ingest pipeline 의존이라 part 후속.
5. **API-55..56 (Project) 활성화** — Project store + handler. project_members + project_integrations 관리.
6. **API-57 (Application 롤업)** — concept §13.4 의 weight_policy normalize 룰을 실 구현.
7. **frontend** — `/admin/settings/applications` IA + 화면 흐름 (별도 frontend sprint).

## 본 sprint 결과 요약

- 마이그레이션 7개 (000012..000018) 작성 — applications / application_repositories / projects / project_members / project_integrations / scm_providers / pr_activities + build_runs + quality_snapshots / RBAC seed 확장.
- 도메인 타입 + store stub + handler stub + RBAC 통합 + route 등록 모두 backend `go test ./...` PASS.
- ADR-0011 proposed → accepted (옵션 C 1차 채택, 옵션 B 단계적 확장 옵션 보존).
- trace.md §2.2 Application/Repository API 서브 표 + §2.4 IMPL-application-XX 서브 표 + §3 row 갱신 + §4 ADR row 갱신 + §6 변경 이력.
- backend_api_contract.md §13.0 placeholder 표 → scaffolded vs planned 구분 표.
