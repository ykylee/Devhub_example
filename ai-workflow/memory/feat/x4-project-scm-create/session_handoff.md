# Session Handoff — feat/x4-project-scm-create

- 문서 목적: X-4 (Phase D — project 생성 flow ↔ SCM create 연계) sprint 의 session handoff
- 범위: GiteaClient + GiteaRepo + GiteaAPIError + SCMCretor + 4 metric + migration 000044 + ADR-0035 + 9 unit test (~1200 line, Tier: **공용**)
- 상태: branch `feat/x4-project-scm-create` 작업 완료, commit + push + PR 발행 pending
- 최종 수정일: 2026-06-14

## 0. 본 세션 핵심 결과

### X-4 정공법 결정

**문제**:
- 현재 `CreateProjectStandalone` 의 `req.RepositoryCreatePayload` 는 **DB tx 안에서 `repositories` 테이블에 row 박기만** — 실제 Gitea instance 에는 repo 없음 (clone_url = `scm+gitea://full_name.git` placeholder).
- 운영 가시성 false. 사용자 입장에서 admin catalog repo link 클릭 시 404.
- N-3 PR (closed) 의 SCM create adapter API 는 노출만, project flow 와 자동 연계 미정.

**결정** (ADR-0035):
- **opt-in feature flag** (`DEVHUB_PROJECT_SCM_CREATE_ENABLED=false`, prod-safe)
- **post-commit Gitea API call** (DB tx commit 후 best-effort, 30s timeout)
- **best-effort 보상** (4 state: pending | success | failed | retry_scheduled, 24h backoff cap)
- **GiteaClient 공유** (X-4 reference minimal + X-5 PR #592 머지 시 통합)
- 4 audit + 4 metric

### 변경 (10 file, +1100 line, Tier: 공용)

1. `backend-core/migrations/000044_repositories_scm_create_state.{up,down}.sql` (NEW)
   - repositories 테이블 4 신규 컬럼: sc_create_status + sc_create_error + sc_external_id + sc_create_at
   - 2 partial index: sc_create_failed queue + sc_create_pending visibility
2. `backend-core/internal/integrations/adapters/gitea_client.go` (NEW)
   - GiteaClient reference minimal (X-5 PR #592 머지 시 통합)
   - doJSON helper (POST + GET + bearer token + JSON decode)
3. `backend-core/internal/integrations/adapters/gitea_repo_create.go` (NEW)
   - GiteaRepo struct + GiteaRepoOptions
   - CreateUserRepo (POST /api/v1/user/repos)
   - CreateOrgRepo (POST /api/v1/orgs/{org}/repos)
   - GiteaAPIError typed error + giteaErrorClass
4. `backend-core/internal/integrations/adapters/scm_creator.go` (NEW)
   - SCMCretor + SCMCreateRequest/Result/Status/Store interface
   - post-commit best-effort + 4 state machine
   - OnSuccess/OnError/OnCompensation hook + recordOutcome
5. `backend-core/internal/integrations/adapters/metrics.go` (MODIFY)
   - 4 scm create metric + observeSCMSuccess + observeSCMError
   - init 을 initHomeLabMetrics 와 합쳐 sync.Once 충돌 회피 (race condition fix)
6. `backend-core/internal/integrations/adapters/gitea_repo_create_test.go` (NEW)
   - 9 unit test: GiteaClient.CreateUserRepo (4) + SCMCretor (4) + giteaErrorClass (1)
7. `docs/adr/0035-project-scm-create-integration.md` (NEW, ~10KB, 9 section)
8. `docs/planning/2026-06-14-x4-project-scm-create-sprint-plan.md` (NEW, ~10KB)
9. `docs/traceability/report.md` (§6 row 추가)
10. 메모리 4 file (state.json + session_handoff + work_backlog + backlog)

## 1. Tier 분류

- **공용** (코드 + ADR + openapi + migration 모두 사내 한정 정보 미포함)
- staging Gitea 실 검증 SOP = **사내** (별도 follow-up docs)

## 2. 신규 ID 발급 (8 row)

- REQ-FR-PROJECT-SCM-CREATE-01
- ARCH-PROJECT-SCM-CREATE-01
- API-110 (POST `/api/v1/projects` 의 `auto_create_scm` + `scm_options` 정합, 4 enum ScmCreateStatus)
- RM-PROJECT-SCM-CREATE-01 (운영자 수동 retry SOP)
- IMPL-PROJECT-SCM-CREATE-01 (SCMCretor + GiteaClient 확장 + handler post-commit hook)
- IMPL-SCM-CREATE-STATE-01 (repositories 테이블 4 신규 컬럼)
- UT-PROJECT-SCM-CREATE-01 (9 unit test)
- TC-PROJECT-SCM-CREATE-01 (e2e, 후속 sprint)

## 3. 검증

- `go build ./...` PASS
- `go test ./internal/integrations/adapters/...` **9 unit test PASS**
- `go test ./internal/domain/application-lifecycle/...` **회귀 0** (기존 CreateProjectStandalone test)
- `go test ./...` 30+ packages 회귀 0 (X-5 와 동일 사전 결함 외)
- e2e shard 1/2/3 skip (path-detect: backend only)

## 4. 잔여 follow-up (X-4 Phase 2)

- **handler post-commit wire**: `CreateProjectStandalone` 에 SCMCretor hook 추가 (tx commit 후 best-effort 호출)
- **openapi spec 확장**: createProjectRequest + ProjectResponse schema + 4 enum ScmCreateStatus + error code
- **main.go env wire**: DEVHUB_PROJECT_SCM_CREATE_ENABLED + SCMCretor wire
- **운영자 수동 retry API**: `POST /api/v1/repositories/{id}/scm-recreate` (follow-up PR)
- **staging Gitea 실 검증 SOP** (사내)
- **e2e spec** (사내)

## 5. 다음 세션 directive

- 본 PR commit + push + PR 발행 + 머지
- 위키 mirror 갱신: `bash scripts/wiki-sync-devhub.sh` 1회 실행 (PR 머지 후)
- X-5 PR #592 머지 후 GiteaClient 통합 (X-4 reference minimal → X-5 rich client)
- X-4 Phase 2: handler wire + openapi + main.go
- 다음 sprint: X-6 (Keycloak group staging-prod, 사내) 또는 X-8 (Keycloak SPI realm events)
- PR #591 + #592 + (X-4 PR) cron 모니터링
