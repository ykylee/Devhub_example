# ADR-0035: Project ↔ SCM Create Integration (X-4, Phase D)

- 문서 목적: X-4 (Phase D — project 생성 flow ↔ SCM create 연계) 의 architecture 결정. project 생성 시 Gitea API 자동 호출 + best-effort 보상.
- 범위: CreateProjectStandalone handler 의 post-commit Gitea API 호출 + 4 audit + 4 metric + opt-in feature flag + 사내 staging Gitea 검증 SOP.
- 상태: Accepted
- 작성일: 2026-06-14
- 결정 근거 sprint: `feat/x4-project-scm-create`
- 결정 근거 commit: (본 PR squash 예정)
- **Tier**: 공용
- 관련 문서: [release_v0-1_roadmap.md §3.5 X-4](../planning/release_v0-1_roadmap.md), [sprint plan](../planning/2026-06-14-x4-project-scm-create-sprint-plan.md), [ADR-0015 HomeLab pull strategy](./0015-homelab-adapter-pull-strategy.md), [ADR-0034 Gitea Hourly Pull Architecture (X-5)](./0034-gitea-hourly-pull-architecture.md) (GiteaClient 재사용, PR #592), [ADR-0028 dev-requests-voc-external-ref](./0028-dev-requests-voc-external-ref.md) (N-13 inbound source cross-ref).

---

## 1. 컨텍스트

### 1.1 현재 정공법 (X-4 진입 전)

`POST /api/v1/projects` (v2 standalone route) 가 `req.RepositoryCreatePayload {key, slug, scm_provider}` 옵션 받음 → store `CreateProjectWithRepositoryPayload` → **DB tx 안에서 `createRepositoryTx` 가 `repositories` 테이블에 row 박기만 함** (clone_url = `scm+gitea://full_name.git` 형식 placeholder).

**문제**:
1. **DevHub DB 에는 repo 가 존재하지만 실제 Gitea 에는 repo 없음** — 사용자 입장에서 admin catalog repo link 클릭 시 404.
2. **운영 가시성 false** — `scm_provider` 가 placeholder URL 일 뿐, 실제 Gitea instance 와 무관.
3. **N-3 PR** (closed PR #380) 의 SCM create 가 **adapter API 만 노출**한 상태 — project flow 와의 자동 연계 미정.

### 1.2 X-4 정공법

project 생성 flow 에서 **`req.RepositoryCreatePayload` 활성화** — 실제 Gitea API 호출 + 실패 시 best-effort 보상 + 4 audit + 4 metric.

### 1.3 X-5 정합

X-5 (Gitea Hourly Pull, PR #592) 의 `GiteaClient` 를 X-4 에서 재사용. **공유 client** (GiteaAPI base URL, token, page size, timeout) + `CreateUserRepo` / `CreateOrgRepo` 신규 method 2개 추가.

## 2. 검토 옵션

### 2.1 opt-in vs opt-out vs always on

| 옵션 | 채택 | 이유 |
|---|---|---|
| **opt-in (default off)** | ✅ | production 의 모든 project 생성 flow 가 SCM call 까지 가게 됨. staging 1주 관찰 후 prod 활성화. v1.0 출시 차단 회피 |
| opt-out (default on) | ❌ | Gitea outage / Gitea 정지 시 모든 project 생성 실패 → v1.0 출시 차단 |
| always on (no flag) | ❌ | 운영 가시성 부재. 사내 staging Gitea instance 미정 → 사내 staging PR 로 분리 |

### 2.2 SCM call 위치 (DB tx commit 전 vs 후)

| 옵션 | 채택 | 이유 |
|---|---|---|
| **DB tx commit 후 (post-commit)** | ✅ | DB tx 안에서 SCM call 시 Gitea outage → DB rollback 시 SCM orphan. post-commit 시 SCM failure = DB row 보존 + 보상 처리 명확. X-5 (post-commit cron) 와 일관 |
| DB tx 안 (pre-commit) | ❌ | Gitea timeout 시 DB rollback → SCM orphan. compensation logic 필요. CRITICAL 한 복잡도 |

## 3. 결정

### 3.1 SCM call 위치: post-commit

`CreateProjectStandalone` 가 다음 순서:

1. **DB tx commit** (project + repository row 생성)
2. **SCM create 호출** (best-effort, 30s timeout)
3. **Response 200** (DB state 기준) + envelope.data 에 `scm_create_status` (pending | success | failed | retry_scheduled) + `scm_create_error` (error_class + message)
4. **background metric + audit emit** (post-commit fire-and-forget)

### 3.2 4 audit event

- `scm.create_started` (per-request, payload: `project_id` + `repository_id` + `scm_provider` + `gitea_org` + `gitea_repo_name`)
- `scm.create_success` (per-request, payload: `project_id` + `repository_id` + `gitea_repo_id` + `gitea_repo_url` + `duration_ms`)
- `scm.create_failed` (per-request, payload: `project_id` + `repository_id` + `error_class` + `error_message` + `http_status` + `duration_ms`)
- `scm.create_compensation` (per-failure, payload: `repository_id` + `compensation_action` (none | retry_scheduled) + `next_retry_at`)

### 3.3 4 metric

- `devhub_scm_create_runs_total{result=success|error}` (Counter)
- `devhub_scm_create_duration_seconds{result}` (Histogram)
- `devhub_scm_create_repos_total` (Gauge) — 현재 success 상태 repo 수
- `devhub_scm_create_failures_total{error_class, scm_provider}` (Counter, 2 label)
  - `error_class`: validation | permission | not_found | rate_limit | server | network | config | unknown
  - `scm_provider`: gitea | github | gitlab | unknown (sourced from `SCMCreateRequest.SCMProvider`)
  - **2026-06-14 갱신**: prior 1 label `{scm_provider}` 의 `scm_provider` 에 ErrorClass (server/permission/...) 가 들어가던 버그 fix — CounterVec label 차원 2개 (error_class × scm_provider) 로 정합. dashboards/alerts 가 진짜 provider 별 grouping 가능. codex PR #595 re-review #6 (P2).

### 3.4 SCM failure 보상

| 시나리오 | 정공법 |
|---|---|
| Gitea success | `repositories.sc_create_status = 'success'` + `sc_create_at = now()` + `sc_external_id` (Gitea repo ID) 갱신 + audit `scm.create_success` + metric `success` |
| Gitea 4xx (validation/permission) | DB tx 이미 commit (project + repository row 존재). `sc_create_status = 'failed'` + `sc_create_error` (error_class + message) + audit `scm.create_failed` + metric `error`. **project 는 그대로 사용 가능** (DevHub DB row 만으론 repo reference 보존). 운영자 별도 retry (follow-up API, `POST /api/v1/repositories/{id}/scm-recreate`) |
| Gitea 5xx / network | `sc_create_status = 'failed'` + audit + metric 동일. **단, backoff 적용** (X-5 와 동일 24h cap) — 다음 project 생성 시 자동 retry 안 함. 운영자 수동 trigger 만 |
| Gitea timeout (30s default) | 5xx 와 동일 처리 |

### 3.5 GiteaClient 확장 (X-5 + X-4 공유)

`gitea_pull.go` (X-5 PR #592 의 신규 파일) 에 2 method 추가:
- `CreateUserRepo(ctx, name, options) (*GiteaRepo, error)` — `POST /api/v1/user/repos`
- `CreateOrgRepo(ctx, org, name, options) (*GiteaRepo, error)` — `POST /api/v1/orgs/{org}/repos`

`GiteaRepo` struct:
```go
type GiteaRepo struct {
    ID           int64
    Name         string
    FullName     string
    CloneURL     string
    HTMLURL      string
    DefaultBranch string
    Private      bool
}
```

### 3.6 schema 확장

신규 migration `000044_repositories_scm_create_state.up.sql`:
```sql
ALTER TABLE public.repositories
  ADD COLUMN sc_create_status text,
  ADD COLUMN sc_create_error text,
  ADD COLUMN sc_external_id bigint,
  ADD COLUMN sc_create_at timestamptz;
```

### 3.7 openapi 확장

`docs/openapi.yaml` 의 `createProjectRequest` schema 확장:
- `auto_create_scm: bool` (default false)
- `scm_options: { private: bool, description: string, auto_init: bool }` (optional)

`ProjectResponse` schema 확장:
- `scm_create_status: enum (pending | success | failed | retry_scheduled)`
- `scm_create_error: string` (optional)
- `gitea_repo_url: string` (optional)

신규 error code: `scm_create_failed` (5xx-like, 502 — Gitea outage 정공법).

### 3.8 env config

| Env | Default | 의미 |
|---|---|---|
| `DEVHUB_PROJECT_SCM_CREATE_ENABLED` | `false` | opt-in |
| `DEVHUB_GITEA_API_BASE_URL` | (X-5 와 공유) | Gitea instance |
| `DEVHUB_GITEA_API_TOKEN` | (X-5 와 공유) | API token |
| `DEVHUB_SCM_CREATE_TIMEOUT` | `30s` | Gitea API call timeout |
| `DEVHUB_SCM_CREATE_RETRY_MAX` | `0` | 자동 retry (default 0 = 보상만, 운영자 수동) |
| `DEVHUB_SCM_CREATE_BACKOFF_CAP` | `24h` | 자동 backoff cap (X-5 와 동일) |

### 3.6 변경이력 (2026-06-14)

| 일자 | commit | 변경 | trigger |
|---|---|---|---|
| 2026-06-14 | `3d810e5` | timeout 시 failure state write 가 parent ctx 사용 (callCtx expired 회피) | codex PR #593 review #2 (P2) |
| 2026-06-14 | `3d810e5` | recordOutcome success path: metric unconditional + hook optional (nil hook double-count 회피) | codex PR #593 review #3 (P2) |
| 2026-06-14 | `3e90cef` | observeSCMError 이중 호출 제거 (failure Counter 2배 fix) | codex PR #595 re-review #4 (P2) |
| 2026-06-14 | `3e90cef` | nil Client (config error) 시 store 에 'failed' write 추가 (pending stuck 회피) | codex PR #595 re-review #5 (P2) |
| 2026-06-14 | `3e90cef` | `devhub_scm_create_failures_total` label 1차원 → 2차원 (error_class × scm_provider) | codex PR #595 re-review #6 (P2) |

## 4. trade-off

### 4.1 post-commit SCM call vs in-tx

**in-tx 의 장점**: SCM failure 시 DB row 자동 rollback. 원자성.
**in-tx 의 단점**: Gitea 30s timeout 동안 DB connection 점유. 동시 project 생성 N건 시 connection pool 고갈.

**post-commit 의 장점**: DB connection 짧음. X-5 cron 와 일관. 보상 logic 명확.
**post-commit 의 단점**: SCM orphan 가능 (DB row 존재 + Gitea 없음). **best-effort + 운영자 수동 retry 로 보완** (ADR-0035 §3.4).

**결론**: post-commit 채택. orphan 가능성은 운영 SOP (slack alert + on-call ticket) + manual retry API (follow-up) 로 cover.

### 4.2 backoff vs retry

**retry (N=3) 의 장점**: 일시적 Gitea 5xx 자동 회복.
**retry 의 단점**: project 생성 flow latency 90s+ (30s × 3 retry). 사용자 경험 저하.

**backoff 의 장점**: project 생성 flow latency < 1s (즉시 best-effort). 24h cap 으로 운영자 manual trigger.

**결론**: backoff 채택. retry=0 (default). 운영자가 follow-up API (POST /api/v1/repositories/{id}/scm-recreate) 로 manual trigger.

### 4.3 사내 staging Gitea

본 PR 의 staging 검증은 staging Gitea instance 필요. **사내 staging Gitea = `git@homelab.ddn777.synology.me:3000` 또는 별도 staging** — **사내 tier PR** 로 분리 (PR body 의 self-check 에 명시).

## 5. cross-tier (사외 / 사내 정합)

| 영역 | Tier | 비고 |
|---|---|---|
| backend migration 000044 | 공용 | schema 만, 사내 한정 컬럼 없음 |
| backend gitea_pull.go (CreateUserRepo/CreateOrgRepo) | 공용 | Gitea API generic, 사내 instance 한정 없음 |
| backend gitea_scm_create.go | 공용 | post-commit + best-effort, 사내 instance 한정 없음 |
| backend handler post-commit hook | 공용 | feature flag gate, opt-in |
| backend metrics.go (scm create metric) | 공용 | Prometheus 표준 metric |
| ADR-0035 | 공용 | docs only |
| sprint plan | 공용 | docs only |
| openapi spec | 공용 | schema extension |
| staging Gitea 실 검증 SOP | **사내** | Gitea instance URL / API token / staging account 사내 한정 → 별도 docs (X-4 follow-up) |
| production Gitea 검증 SOP | **사내** | 동일 |

## 6. 검증

- `go test ./internal/integrations/adapters/...` 6 신규 unit test PASS
- `go test ./internal/domain/application-lifecycle/...` 회귀 0 (기존 CreateProjectStandalone test)
- `go test ./...` 30+ packages 회귀 0
- `go build ./...` silent PASS
- openapi lint PASS (schema extension)
- e2e shard 1/2/3 skip (path-detect: backend only 변경)

## 7. supersession

- ADR-0015 §5 (Homelab 1차 본) 와 cross-ref, **conflict 없음**. ADR-0015 = Homelab pull, ADR-0035 = Gitea create.
- ADR-0034 §3.5 (X-5 Gitea metric) 와 cross-ref, **conflict 없음**. ADR-0034 = pull cron, ADR-0035 = create hook. metric 4종 각각 별도 (`devhub_gitea_pull_*` vs `devhub_scm_create_*`).
- ADR-0028 §6 (a) (N-13 inbound source) 와 cross-ref. 본 X-4 = project 생성 시 SCM create, N-13 = dev-request intake 시 inbound source 자동 routing. **별도 시나리오, cross-ref 만**.

## 8. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-14 | 1차 발행 (Accepted). X-4 Project ↔ SCM create 연계 + opt-in feature flag + post-commit Gitea API call + best-effort 보상 + 4 audit + 4 metric. | `feat/x4-project-scm-create` |
