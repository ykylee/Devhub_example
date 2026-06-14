# X-4 Project ↔ SCM create 연계 (Phase D) — Sprint Plan (2026-06-14)

- 문서 목적: X-4 (Phase D — project 생성 flow ↔ SCM create 연계) 의 sprint plan.
- 범위: project 생성 시 SCM create (Gitea API) 자동 호출 + 4 audit + 4 metric + ADR-0035 + openapi + UT.
- 상태: draft (X-3 PR #591 + X-5 PR #592 머지 후 본 sprint 진입)
- 결정 근거 sprint: `feat/x4-project-scm-create`
- 관련 문서: [release_v0-1_roadmap.md §3.5 X-4](../release_v0-1_roadmap.md), [ADR-0015 HomeLab pull strategy](../adr/0015-homelab-adapter-pull-strategy.md) (정합 패턴), [ADR-0034 Gitea Hourly Pull Architecture (X-5)](../adr/0034-gitea-hourly-pull-architecture.md) (GiteaClient 재사용, X-5 PR #592), [ADR-0028 dev-requests-voc-external-ref](../adr/0028-dev-requests-voc-external-ref.md) (N-13 inbound source cross-ref), [`backend-core/internal/domain/application-lifecycle/view/projects.go` line 303 CreateProjectStandalone](../../backend-core/internal/domain/application-lifecycle/view/projects.go) (기존 standalone route, `req.RepositoryCreatePayload` 지원), [`backend-core/internal/domain/application-lifecycle/repository/projects.go` line 567 CreateProjectWithRepositoryPayload](../../backend-core/internal/domain/application-lifecycle/repository/projects.go) (현재 DB draft 만 — `createRepositoryTx` 가 DB row 만 생성, Gitea API 호출 없음).

## 0. 컨텍스트

### 0.1 현재 정공법 (X-4 진입 전)

`POST /api/v1/projects` (v2 standalone route) 가 `req.RepositoryCreatePayload {key, slug, scm_provider}` 옵션 받음 → store `CreateProjectWithRepositoryPayload` → **DB tx 안에서 `createRepositoryTx` 가 `repositories` 테이블에 row 박기만 함** (clone_url = `scm+gitea://full_name.git` 형식 placeholder).

**결과**: DevHub DB 에는 repo 가 존재하지만, **실제 Gitea instance 에는 repo 가 없음**. 사용자 입장에서 admin catalog 에서 repo link 클릭 시 404. **운영 가시성 false** — `scm_provider` 가 placeholder URL 일 뿐.

### 0.2 issue #231 / N-3 / N-2 정공법 정합

- N-2 (`repository draft→publish UT 보강`): backend draft→publish lifecycle 만 다룸. X-4 와 별도.
- N-3 (`SCM import/create + draft/publish happy-path E2E`): SCM **import** (existing repo link) + **create** (DevHub 가 SCM 에 create) 의 1차 PR. **이 1차 PR 은 SCM adapter 의 import/create API 자체를 노출**한 것이고, **project 생성 flow 와의 자동 연계** 는 X-4 의 정밀화.
- 본 X-4 = N-3 의 **Phase D 정밀화** = `CreateProjectStandalone` 의 `req.RepositoryCreatePayload` 활성화 (실제 Gitea API 호출 + 실패 보상 + audit + metric).

### 0.3 X-5 정합

X-5 (Gitea Hourly Pull) 의 `GiteaClient` (PR #592) 를 X-4 에서 재사용:
- `ListPullRequestsSince` / `ListBuilds` (X-5 전용)
- `CreateUserRepo(ctx, name, options)` (X-4 신규) — `POST /api/v1/user/repos` 또는 org repo
- `CreateOrgRepo(ctx, org, name, options)` (X-4 신규) — `POST /api/v1/orgs/{org}/repos`

## 1. 결정

### 1.1 opt-in default

| 옵션 | 채택 | 이유 |
|---|---|---|
| **opt-in (default off, `DEVHUB_PROJECT_SCM_CREATE_ENABLED=false`)** | ✅ | production 의 모든 project 생성 flow 가 SCM call 까지 가게 됨 → 안전 우선. staging 1주 관찰 후 prod 활성화 |
| opt-out (default on) | ❌ | Gitea outage / Gitea 정지 시 모든 project 생성 실패 → v1.0 출시 차단 |
| always on (no flag) | ❌ | 운영 가시성 부재. 사내 staging Gitea instance 미정 → 사내 staging PR 로 분리 |

### 1.2 SCM call 위치 (DB tx commit 후)

| 옵션 | 채택 | 이유 |
|---|---|---|
| **DB tx commit 후 (post-commit)** | ✅ | DB tx 안에서 SCM call 시 Gitea outage → DB rollback 시 SCM orphan (compensation logic 필요). post-commit 시 SCM failure = DB row 보존 + 보상 처리 명확 |
| DB tx 안 (pre-commit) | ❌ | Gitea timeout 시 DB rollback → SCM orphan. X-5 정공법 (post-commit cron) 와 일관 |

### 1.3 SCM failure 보상

| 시나리오 | 정공법 |
|---|---|
| **Gitea success** | `repositories.sc_create_status = 'success'` + `sc_create_at = now()` + `sc_external_id` (Gitea repo ID) 갱신 + audit `scm.create_success` + metric `success` |
| **Gitea 4xx (validation/permission)** | DB tx 이미 commit (project + repository row 존재). `repositories.sc_create_status = 'failed'` + `sc_create_error` (error_class + message) + audit `scm.create_failed` + metric `error`. **project 는 그대로 사용 가능** (DevHub DB row 만으론 repo reference 보존). 운영자가 별도 retry (follow-up API, `POST /api/v1/repositories/{id}/scm-recreate`) |
| **Gitea 5xx / network** | `sc_create_status = 'failed'` + audit + metric 동일. **단, backoff 적용** (X-5 와 동일 24h cap) — 즉 다음 project 생성 시 자동 retry 안 함. 운영자 수동 trigger 만 |
| **Gitea timeout (30s default)** | 5xx 와 동일 처리 |

### 1.4 4 audit event

- `scm.create_started` (per-request, payload: `project_id` + `repository_id` + `scm_provider` + `gitea_org` + `gitea_repo_name`)
- `scm.create_success` (per-request, payload: `project_id` + `repository_id` + `gitea_repo_id` + `gitea_repo_url` + `duration_ms`)
- `scm.create_failed` (per-request, payload: `project_id` + `repository_id` + `error_class` + `error_message` + `http_status` + `duration_ms`)
- `scm.create_compensation` (per-failure, payload: `repository_id` + `compensation_action` (none|retry_scheduled) + `next_retry_at`)

### 1.5 4 metric

- `devhub_scm_create_runs_total{result=success|error}` (Counter)
- `devhub_scm_create_duration_seconds{result}` (Histogram)
- `devhub_scm_create_repos_total` (Gauge) — 현재 success 상태 repo 수
- `devhub_scm_create_failures_total{scm_provider}` (Counter, label 분리 — Gitea vs others)

### 1.6 env config

| Env | Default | 의미 |
|---|---|---|
| `DEVHUB_PROJECT_SCM_CREATE_ENABLED` | `false` | opt-in |
| `DEVHUB_GITEA_API_BASE_URL` | (X-5 와 공유) | Gitea instance |
| `DEVHUB_GITEA_API_TOKEN` | (X-5 와 공유) | API token |
| `DEVHUB_SCM_CREATE_TIMEOUT` | `30s` | Gitea API call timeout |
| `DEVHUB_SCM_CREATE_RETRY_MAX` | `0` | 자동 retry 횟수 (default 0 = 보상 처리만, 운영자 수동 trigger) |
| `DEVHUB_SCM_CREATE_BACKOFF_CAP` | `24h` | 자동 backoff cap (X-5 와 동일) |

## 2. 변경 범위 (PR 1개, 코드 ~800 line)

### 2.1 backend (6 file)

1. `backend-core/internal/integrations/adapters/gitea_pull.go` (MODIFY, +60 line)
   - `CreateUserRepo(ctx, name, options) (*GiteaRepo, error)` 추가
   - `CreateOrgRepo(ctx, org, name, options) (*GiteaRepo, error)` 추가
   - `GiteaRepo` struct (ID, Name, FullName, CloneURL, HTMLURL, DefaultBranch, Private)

2. `backend-core/internal/integrations/adapters/gitea_scm_create.go` (NEW, ~200 line)
   - `SCMCreator` struct (GiteaClient + store + metric + audit + opts)
   - `CreateRepository(ctx, payload) (*SCMCreateResult, error)`
   - post-commit 호출 → success/failed 분기
   - 4 audit emit
   - 4 metric emit
   - `*SCMCreateError` typed error

3. `backend-core/internal/integrations/adapters/metrics.go` (MODIFY, +30 line)
   - 4 scm create metric + observe helper

4. `backend-core/internal/integrations/adapters/gitea_scm_create_test.go` (NEW, ~250 line)
   - 6 unit test: CreateUserRepo success / 401 invalid token / 422 duplicate / 5xx network / timeout / org repo not found
   - integration 시 mock Gitea server + post-commit verify

5. `backend-core/internal/domain/application-lifecycle/view/projects.go` (MODIFY, +50 line)
   - `CreateProjectStandalone` 에 SCM create post-commit hook 추가
   - req.RepositoryCreatePayload 확장 (`auto_create_scm: bool`, `scm_options: {private, description}`)
   - feature flag `DEVHUB_PROJECT_SCM_CREATE_ENABLED` gate
   - tx commit 후 `SCMCreator.CreateRepository` 호출 + best-effort (project response 200, scm_create status 별도 envelope data field)

6. `backend-core/internal/shared/config/config.go` (MODIFY, +15 line)
   - 6 env var 추가

7. `backend-core/main.go` (MODIFY, +20 line)
   - SCMCreator wire + DEVHUB_PROJECT_SCM_CREATE_ENABLED parse

### 2.2 docs (3 file)

1. `docs/adr/0035-project-scm-create-integration.md` (NEW, ~10KB)
   - §1 상태: Accepted
   - §2 컨텍스트: 현재 DB draft 만 + Gitea 호출 부재
   - §3 검토 옵션: 3가지 (opt-in vs opt-out vs always on) + tx 위치 (pre-commit vs post-commit)
   - §4 결정: opt-in + post-commit + best-effort 보상
   - §5 trade-off: tx 안 vs 밖 / backoff vs retry / 사내 staging Gitea
   - §6 cross-tier: 사외 PR, 사내 staging 검증 SOP 별도
   - §7 supersession: ADR-0015 / ADR-0034 와 cross-ref
   - §8 변경 이력

2. `docs/traceability/report.md` (MODIFY, +1 row)
   - §6 본 row 추가

3. `docs/openapi.yaml` (MODIFY, ~120 line)
   - `createProjectRequest` schema 확장 (`auto_create_scm` + `scm_options` field)
   - `ProjectResponse` schema 확장 (`scm_create_status` + `scm_create_error` + `gitea_repo_url`)
   - 4 신규 enum: `ScmCreateStatus` (pending | success | failed | retry_scheduled)
   - 신규 error code: `scm_create_failed` (5xx-like, 502 — Gitea outage 정공법)

### 2.3 메모리 (4 file)

1. `ai-workflow/memory/feat/x4-project-scm-create/state.json` (NEW)
2. `ai-workflow/memory/feat/x4-project-scm-create/session_handoff.md` (NEW)
3. `ai-workflow/memory/feat/x4-project-scm-create/work_backlog.md` (NEW)
4. `ai-workflow/memory/feat/x4-project-scm-create/backlog/2026-06-14.md` (NEW)

## 3. 신규 ID 발급 (8 row)

- `REQ-FR-PROJECT-SCM-CREATE-01` (project 생성 flow ↔ SCM create 자동 연계)
- `ARCH-PROJECT-SCM-CREATE-01` (post-commit + opt-in + best-effort 보상 결정)
- `API-110` (POST `/api/v1/projects` 의 `auto_create_scm` + `scm_options` 정합, 4 enum ScmCreateStatus)
- `RM-PROJECT-SCM-CREATE-01` (운영자 수동 retry SOP)
- `IMPL-PROJECT-SCM-CREATE-01` (SCMCreator + GiteaClient 확장 + handler post-commit hook)
- `IMPL-SCM-CREATE-STATE-01` (`repositories` 테이블 3 신규 컬럼: `sc_create_status` + `sc_create_error` + `sc_external_id` + `sc_create_at`)
- `UT-PROJECT-SCM-CREATE-01` (6 unit test)
- `TC-PROJECT-SCM-CREATE-01` (e2e: staging Gitea 1회 happy-path + 1회 failure path, 후속 sprint)

## 4. migration 추가

`000044_repositories_scm_create_state.up.sql` (NEW):
```sql
ALTER TABLE public.repositories
  ADD COLUMN sc_create_status text,
  ADD COLUMN sc_create_error text,
  ADD COLUMN sc_external_id bigint,
  ADD COLUMN sc_create_at timestamptz;
```

## 5. Tier

- **공용** (코드 + ADR + openapi + migration 모두 사내 한정 정보 미포함)
- staging Gitea 실 검증 SOP = **사내** (별도 follow-up docs)

## 6. 검증

- `go test ./internal/integrations/adapters/...` 6 신규 unit test PASS
- `go test ./internal/domain/application-lifecycle/...` 회귀 0 (기존 CreateProjectStandalone)
- `go test ./...` 30+ packages 회귀 0
- `go build ./...` silent PASS
- openapi lint PASS (schema 확장 정합)
- e2e shard 1/2/3 skip (path-detect: backend only 변경)

## 7. 잔여 (follow-up)

- **운영자 수동 retry API**: `POST /api/v1/repositories/{id}/scm-recreate` (ADR-0035 §1.3)
- **staging Gitea 실 검증 SOP**: 1회 happy-path + 1회 failure path
- **e2e spec**: staging Gitea mock + best-effort
- **multi-SCM 정밀화**: GitHub / GitLab / Bitbucket adapter (X-4 Phase E, 별도 sprint)

## 8. 다음 세션 directive

- 본 sprint 진입 시 X-3 PR #591 + X-5 PR #592 머지 확인 + main 정합
- 본 sprint = PR 1개 (backend + ADR + openapi + migration, ~800 line)
- PR 머지 후 main flat memory 3 file finalize + 위키 mirror 갱신
- X-6 / X-8 (사내 2건) follow-up 결정

## 9. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-14 | 본 sprint plan 초안 (X-4 Project ↔ SCM create 연계) | `feat/x4-project-scm-create` |
