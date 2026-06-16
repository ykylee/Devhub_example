# Work Backlog — claude/work_260514-c

- 상태 관리: planned / in_progress / blocked / done

## A. API-51..54 Repository 운영 지표

- [planned] store: `ListRepositoryActivity(ctx, repoID, opts) ([]Activity, error)` — pr_activities aggregation (commit count, active contributor count) — 또는 raw pr_activities 조회 + handler 측 집계
- [planned] store: `ListRepositoryPullRequests(ctx, repoID, opts) ([]PRActivity, int, error)` — pr_activities 의 event 단위 list + pagination
- [planned] store: `ListRepositoryBuildRuns(ctx, repoID, opts) ([]BuildRun, int, error)` — build_runs list
- [planned] store: `ListRepositoryQualitySnapshots(ctx, repoID, opts) ([]QualitySnapshot, error)` — quality_snapshots list
- [planned] domain types: PRActivity / BuildRun / QualitySnapshot
- [planned] handler: 4개 endpoint (read-only)
- [planned] route + permission table

## B. API-55..56 Project CRUD

- [planned] handler: listProjects / createProject / getProject / updateProject / archiveProject
- [planned] body binding + visibility/status enum 검증 + UNIQUE(repository_id, key) 충돌 매핑
- [planned] audit emit: project.created / project.updated / project.archived
- [planned] route + permission table

## C. API-57 Application 롤업 + active→closed 가드

- [planned] store: `ComputeApplicationRollup(ctx, applicationID, opts) (Rollup, error)` — link 목록 + 각 repo 의 pr_activities/build_runs/quality_snapshots 집계 + weight normalize
- [planned] domain types: ApplicationRollup + RollupMeta (period / filters / weight_policy / applied_weights / fallbacks / data_gaps)
- [planned] weight_policy normalize 룰 — equal / repo_role (primary=0.6/sub=0.3/shared=0.1 + 다중 정규화) / custom (합 1.0 ±0.001)
- [planned] handler: getApplicationRollup
- [planned] active→closed 가드 흡수 — store: `CountCriticalRollupWarnings(ctx, applicationID) (int, error)` + handler updateApplication 분기 추가
- [planned] backend_api_contract §13.6 응답 schema 와 정합

## D. API-58 Integration CRUD

- [planned] store: `ListIntegrations(ctx, scope, parentID) / Create / Update / Delete`
- [planned] domain types: ProjectIntegration (이미 PR #105 에서 정의)
- [planned] handler: listIntegrations / createIntegration / updateIntegration / deleteIntegration
- [planned] scope polymorphism 검증 — scope=application 시 application_id 필수, scope=project 시 project_id 필수
- [planned] audit emit: integration.created / integration.updated / integration.deleted
- [planned] route + permission table — RBAC resource 결정 (applications + projects 의 cross-cut)

## E. 매트릭스 + 거버넌스

- [planned] trace.md §2.2 — API-51..58 planned → activated
- [planned] trace.md §2.4 — IMPL-application-{rollup,integration,repo_ops}-01 + IMPL-project-handler-01
- [planned] trace.md §3 — Application/Project row 의 IMPL/UT 컬럼 갱신
- [planned] trace.md §6 변경 이력
- [planned] backend_api_contract.md §13.0 — 상태 컬럼 activated 갱신
- [planned] backend_api_contract.md §13.2 PATCH — active→closed 가드 carve out close (보강 완료)
- [planned] concept.md — 13.5 오픈 이슈 갱신 (롤업 가드 carve out close)

## F. 위험 / 검토

- [planned] 롤업 store 의 N+1 query — 다중 repo 의 각 메트릭 fetch. 1차는 단순 구현, 후속 정리.
- [planned] integration 의 RBAC resource — applications 인가 projects 인가? scope 별 다르게 적용할지. 1차 결정.
