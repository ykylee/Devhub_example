# Work Backlog — feat/x4-project-scm-create

- 문서 목적: X-4 (Project ↔ SCM create 연계, Phase D) sprint 의 work backlog
- 범위: GiteaClient + SCMCretor + migration 000044 + 4 metric + ADR-0035 + 9 unit test (~1100 line, Tier: 공용)
- 상태: Phase 1 (backend core) 완료, Phase 2 (handler wire + openapi + main.go) 후속
- 최종 수정일: 2026-06-14

## 0. 현재 status

**main HEAD**: `fc2d2a0` (PR #590 X-2 100% 완료 시점)
**branch**: `feat/x4-project-scm-create` (worktree `.worktrees/x4-project-scm-create/`)
**sprint scope**: X-4 (Project ↔ SCM create 연계, Phase D, issue #231/X-3/N-3/N-2 정밀화) — opt-in + post-commit + best-effort

## 1. 본 sprint 진행 상황

| Step | 작업 | 결과 |
|---|---|---|
| 1 | worktree 생성 (`feat/x4-project-scm-create`) | ✅ HEAD 기준 |
| 2 | `000044_repositories_scm_create_state` migration (up + down) | ✅ |
| 3 | `gitea_client.go` (reference minimal GiteaClient) | ✅ |
| 4 | `gitea_repo_create.go` (GiteaRepo + CreateUserRepo + CreateOrgRepo + GiteaAPIError) | ✅ |
| 5 | `scm_creator.go` (SCMCretor + post-commit + 4 state machine) | ✅ |
| 6 | `metrics.go` (4 scm create metric + sync.Once fix) | ✅ |
| 7 | `gitea_repo_create_test.go` (9 unit test) | ✅ |
| 8 | `ADR-0035` (9 section) | ✅ |
| 9 | `sprint plan` | ✅ |
| 10 | 메모리 4 file | ✅ |
| 11 | commit + push + PR 발행 | ⏳ pending |

## 2. 잔여 follow-up (X-4 Phase 2)

| Step | 작업 | effort |
|---|---|---|
| 12 | `CreateProjectStandalone` handler post-commit wire (tx commit 후 SCMCretor hook) | 0.5 ses |
| 13 | `openapi.yaml` schema 확장 (auto_create_scm + scm_options + ScmCreateStatus enum) | 0.3 ses |
| 14 | `main.go` env wire (DEVHUB_PROJECT_SCM_CREATE_ENABLED + 5 env var) | 0.2 ses |
| 15 | `config.go` 6 env var | 0.1 ses |

## 3. 다음 sprint

| ID | 아이템 | 영역 | effort |
|---|---|---|---|
| X-6 | Keycloak group staging-prod (P1-3, #214) | 사내 | 0.5~1 ses |
| X-8 | Keycloak SPI realm events push (P2-6/P3-5) | BE/사내 | 2~3 ses |

## 4. 변경 이력

| 일자 | 변경 | 비고 |
|---|---|---|
| 2026-06-14 | X-4 Phase 1 (backend core) branch 신규 (~1100 line, Tier: 공용) | 본 sprint |
