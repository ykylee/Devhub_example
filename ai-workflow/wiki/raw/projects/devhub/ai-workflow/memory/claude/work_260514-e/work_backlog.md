# Work Backlog — claude/work_260514-e

- 상태 관리: planned / in_progress / blocked / done

## A. Fixture helper

- [planned] `applications_integration_test.go` 의 helper — DB cleanup (TRUNCATE applications/projects/integrations/application_repositories/scm_providers in cascade) + seed (users.user_id, repositories.id)

## B. Applications store integration

- [planned] CreateApplication happy + UNIQUE key 위반 → ErrConflict + invalid FK (owner_user_id) → ErrConflict
- [planned] GetApplication happy + NotFound + GetApplicationByKey
- [planned] UpdateApplication happy + status='archived' 자동 archived_at + status 변경 시 archived_at NULL 화 + FK 위반
- [planned] ArchiveApplication happy
- [planned] ListApplications status / include_archived / q 필터
- [planned] CountActiveApplicationRepositories — 0 / 1 / 2 link
- [planned] Application-Repository link CRUD (composite PK 위반, sync_error_code CHECK)
- [planned] UpdateApplicationRepositorySync 의 errorCode 비어있을 때 reset 동작
- [planned] SCM provider Update happy (adapter_version 변경 안 됨)

## C. Repository ops + rollup integration

- [planned] pr_activities / build_runs / quality_snapshots fixture seed
- [planned] ListRepositoryActivity (PR event + build run aggregate)
- [planned] ListRepositoryPullRequests filter (event_type) + pagination
- [planned] ListRepositoryBuildRuns filter (status, branch)
- [planned] ListRepositoryQualitySnapshots filter (tool)
- [planned] **ComputeApplicationRollup — P1 회귀 guard**:
  - equal policy: 다중 repo 의 1/N 가중치
  - repo_role policy: primary/sub/shared 정상 + unknown role fallback (B1 회귀)
  - custom policy: 합 1.0 검증 + 누락 repo fallback **+ sum=1.0 정규화 (P1 핵심)**
- [planned] CountApplicationCriticalWarnings — critical=0 / >0 케이스

## D. Integrations + Projects integration

- [planned] Projects CRUD + UNIQUE (repository_id, key) + FK 위반
- [planned] Integration CRUD + scope polymorphism + partial UNIQUE
- [planned] **UpdateIntegration unique violation — P2 회귀 guard** (external_key 변경 → 다른 row 와 충돌 → ErrConflict)

## E. 매트릭스 + 거버넌스

- [planned] trace.md §3 row 의 UT 컬럼에 integration test 발급 명시
- [planned] trace.md §6 변경 이력
- [planned] PR body 추적성 영향 + Test plan
