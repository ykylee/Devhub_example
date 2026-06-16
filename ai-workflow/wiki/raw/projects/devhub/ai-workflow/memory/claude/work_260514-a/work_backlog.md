# Work Backlog — claude/work_260514-a

- 상태 관리: planned / in_progress / blocked / done

## A. Application 도메인 backend design 1차

- [in_progress] sprint scope 정리 + 추적성 영향 사전 진단 (sync-checklist §1)
- [planned] 마이그레이션 000012 `create_applications` (PK id UUID + UNIQUE key + status/visibility + audit cols + indexes)
- [planned] 마이그레이션 000013 `create_application_repositories` (composite PK + sync_status + sync_error_* + role + indexes)
- [planned] 마이그레이션 000014 `create_scm_providers` (provider_key PK + enabled + adapter_version + seed bitbucket/gitea/forgejo/github)
- [planned] 마이그레이션 000015 `create_projects` (UUID PK + UNIQUE (repository_id, key) + audit cols)
- [planned] 마이그레이션 000016 `create_project_members` (composite PK + project_role + joined_at)
- [planned] 마이그레이션 000017 `create_project_integrations` (id PK + scope + project_id/application_id nullable + policy)
- [planned] 마이그레이션 000018 `create_pr_activities_build_runs_quality_snapshots` (3 entity 묶음)
- [planned] `internal/store/applications.go` interface 초안 (CRUD + status transition + link CRUD)
- [planned] `internal/store/postgres_applications.go` placeholder (실 구현 carve, error stub)
- [planned] `internal/httpapi/applications.go` handler stub (API-43~50 + RBAC gate + envelope + 404/422 분기 placeholder)
- [planned] RBAC matrix seed 확장 — `applications` / `application_repositories` / `projects` / `scm_providers` 리소스 추가 (system_admin only)
- [planned] 단위테스트 — 마이그레이션 up/down + status transition guard

## B. ADR-0011 평가/결정

- [planned] 옵션 A (Casbin / ABAC), B (RBAC 매트릭스 확장 row_predicate), C (handler/service 코드 검증), D (PG RLS) 의 비용/리스크/마이그레이션 영향 표 작성
- [planned] 단일 옵션 채택 + `docs/adr/0011-rbac-row-scoping.md` 의 §3~§5 채우기 + status proposed → accepted
- [planned] REQ-FR-PROJ-009 / -010 활성화 조건 binding 명시
- [planned] 매트릭스 §2.2 RBAC 인덱스 + §3 Application/Project row IMPL 영향 반영

## C. 매트릭스 + 거버넌스 동기

- [planned] trace.md §2.2 API contract — API-41~50 의 (planned) 제거 후 정식 인덱스 갱신 (API-51~58 placeholder 유지)
- [planned] trace.md §2.4 IMPL — `IMPL-application-store-XX`, `IMPL-application-handler-XX` 정의
- [planned] trace.md §3 Application/Project row 갱신 (ARCH/API/IMPL 컬럼 채움)
- [planned] trace.md §4 ADR-0011 row 상태 갱신 (proposed → accepted)
- [planned] trace.md §6 변경 이력 row 추가 (sprint claude/work_260514-a)
- [planned] PR body — sync-checklist §1 추적성 영향 섹션 채우기
- [planned] backend_api_contract.md §13.0 placeholder 표 — 활성화된 ID 의 (planned) 접미사 제거

## D. 위험 / 검토 항목

- [planned] PR 크기 — 마이그레이션 7개 + store + handler + ADR 결정이 한 PR 에 들어가면 비대. 1차 평가 후 분할 결정.
- [planned] `pr_activities`/`build_runs`/`quality_snapshots` 가 기존 마이그레이션 (000001~000011) 의 webhook_events / ci_runs / repositories 와 정합한지 점검 — 중복/충돌 검출.
- [planned] composite PK 마이그레이션 시 기존 인덱스 전략 (org_units 의 BFS, hrdb 의 indexes) 과 패턴 정합 점검.

## 완료
- (없음)
