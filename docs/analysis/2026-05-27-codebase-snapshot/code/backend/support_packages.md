# Backend 보조 패키지 상세 분석 — backend-core/internal/* + main.go + backend-ai

- 문서 목적: DevHub backend-core 의 도메인 핵심(httpapi/store/domain) 외곽에 위치한 보조 패키지(SCM 동기화, audit, auth, worker, config, cron, normalize, hrdb, adapters)와 백그라운드 워커 기동(main.go), backend-ai 스켈레톤을 항목별로 정리한다.
- 범위: `backend-core/internal/{gitea,audit,auth,commandworker,serviceaction,config,devrequest,normalize,hrdb,integrations/adapters}`, `backend-core/main.go`, `backend-ai/`.
- 기준 스냅샷: main `99d6edc` (2026-05-27, #373 직후). 본 분석은 코드 read-only.
- 작성일: 2026-05-27

---

## 1. `internal/gitea/` — Gitea webhook HMAC + SCM pull sync

### 1.1 signature.go — webhook HMAC 검증
- `VerifySignature(body, secret, signature) bool` (`signature.go:10`): HMAC-SHA256 검증. `sha256=` prefix 허용, `hex.DecodeString` 후 `hmac.Equal` 으로 상수시간 비교. secret 또는 signature 가 비면 `false` (fail-closed). 호출처는 전용 Gitea webhook 핸들러(`internal/httpapi/gitea_webhook.go:55`).
- `PayloadHash(body) string` (`signature.go:28`): dedupe key 용 SHA256 hex.

### 1.2 client.go — Gitea REST API client
- `Client` (`client.go:16`): `BaseURL` / `Token`(legacy PAT) / `AuthHeader`(전체 Authorization 값, Token 우선) / `HTTPClient`(10s timeout).
- 페이지네이션 조회 3종: `ListUserRepos`(`:91`), `ListIssues`(`:123`, open/closed), `ListPullRequests`(`:158`). 모두 `limit=50` page loop.
- `CreateRepo(ctx, owner, opts)` (`:203`): owner 비면 `POST /user/repos`, 있으면 `POST /orgs/{owner}/repos`. 409 는 `do()` 가 status error 로 전파.
- `newRequest`(`:219`): `AuthHeader` 가 있으면 그대로, 없고 `Token` 만 있으면 `token <pat>`. `do`(`:246`): 2xx 외 status 는 `gitea api returned status %d` 에러.

### 1.3 auth.go — outbound auth 해석 (auth_mode 5종)
- `AuthorizationHeader(ctx, auth, httpClient) (string, ok bool, err)` (`auth.go:25`): auth_mode 별 Authorization 헤더 산출.
  - `basic`/`app_password` → `Basic base64(user:secret)`, 자격 누락 시 `ok=false`.
  - `oauth2` → client-credentials grant(`fetchOAuth2Token`, `:61`)로 access token 교환 후 `Bearer`. client_id/token_url/secret 누락 시 `ok=false`, 교환 실패 시 error.
  - `agent` → 직접 API sync 불가, `ok=false`(skip).
  - `token`/unset → `token <pat>`, 비면 `ok=false`.
- `NewClientForAuth(ctx, baseURL, auth)` (`:97`): 헤더 해석 후 `Client.AuthHeader` 주입. 자격 없으면 `(nil, nil)` 반환 → caller 가 skip 판단.

### 1.4 syncer.go — repo 단위 deep sync
- `SyncStore` interface(`syncer.go:14`): `UpsertRepository/UpsertUser/UpsertIssue/UpsertPullRequest`.
- `Syncer.SyncRepository`(`:32`): issues(open/closed) + PRs(open/closed) 조회 → user upsert + issue/PR upsert. issue/PR upsert 실패는 warning log 만(non-fatal), 목록 fetch 실패만 error 전파.
- `normalizePullRequestState`(`:141`): `MergedAt != nil` → `merged`, 그 외 `closed`/`open`.

### 1.5 worker.go — 백그라운드 sync 워커 (per-provider, Phase 3)
- `SyncJobStore`(`worker.go:14`) = `SyncStore` + `AcquireNextQueuedSyncJob`(SKIP LOCKED 큐 acquire) + `UpdateIntegrationSyncJobStatus` + `GetIntegrationProviderByID`.
- `SyncWorker.Run`(`:40`): interval ≤0 면 30s ticker loop. ctx cancel 까지.
- `ProcessOnce`(`:63`) 동작 순서:
  1. queued sync job 우선 acquire. 비-SCM job 은 store layer(`provider_type='scm'` gate, codex #341 P1) 에서 차단되어 도달 안 함.
  2. `resolveSyncConfig`(`:114`) 로 provider 의 base_url + auth_mode 별 자격 해석. base_url 비면 job `failed`. `NewClientForAuth` 가 `nil` 반환(자격 미설정/agent 등) 시 job `failed`.
  3. `syncAllWith`(`:128`): `ListUserRepos` → repo 마다 `UpsertRepository`(`Source=SCM`, `ProviderID` 기록) + `SyncRepository`.
  4. 큐가 비면 env(`GITEA_URL`/`GITEA_TOKEN`) 기반 주기 sync(legacy, token mode). 둘 중 하나라도 비면 no-op.
- **보안 핵심**(`resolveSyncConfig` 주석 `:105-126`): 명시 provider 해석 시 env token 으로 **fallback 하지 않는다**(codex #358 P1). provider 고유 host 에 worker 전역 토큰 유출 회피. env fallback 은 provider 미명시(legacy)만.

### 1.6 외부 의존 / 주기
- 외부: Gitea/Forgejo/Gogs REST API(`/api/v1/...`). webhook 은 push(서버→DevHub), sync 워커는 pull(DevHub→Gitea).
- 주기: 30s tick(main.go 고정). Prometheus metric **없음**(gitea 패키지에 metrics 파일 부재 — sync 진행/실패는 log 만, 부채 §발견사항).

---

## 2. `internal/audit/` — Keycloak event 폴러 + user_sync + metrics

### 2.1 keycloak_event_puller.go — cron 폴러
- `KeycloakEventLister`(`:32`): `ListUserEvents` / `ListAdminEvents`(dateFrom inclusive, max page size). `KeycloakUserEvent`/`KeycloakAdminEvent` 는 httpapi struct 미러(circular import 회피).
- `RunKeycloakEventPuller`(`:104`): interval default 30s, max_events default 500. startup 즉시 1차 tick 후 ticker loop. ctx cancel 까지.
- `pullOnce`(`:151`): user + admin 두 branch 독립 처리(한쪽 error 가 다른 쪽 emit 막지 않음), 첫 error 만 반환.
- dedup / cursor(`pullUserEvents` `:173`, `pullAdminEvents` `:248`):
  - cursor 이전 event skip + 동일 ms boundary 는 hash dedup(Keycloak `dateFrom` inclusive 라 직전 event 재등장 — at-least-once).
  - cursor 는 skip type 까지 포함해 advance(skip-only page 에서 cursor stuck 회피, codex hotfix #10 P1-A).
  - same-ms 에 skip+emit 혼재 시 emit-able event 의 hash 를 cursor 에 저장(codex hotfix #11 P2 — 매 tick re-emit/counter 누적 side effect 해소, `:191-220`).
- `loadCursor`(`:318`): row 없으면 `now()` 로 seed + **즉시 UPSERT 영구화**(codex hotfix #10 P1-B — in-memory only 면 빈 첫 poll 후 reinit 으로 두 tick 사이 event 영구 누락).
- `hashUserEvent`(`:347`) / `hashAdminEvent`(`:356`): distinguishing 7-tuple SHA256(time/type/user/ip/client/realm/session 등). store partial UNIQUE INDEX(migration 000032) 의 `source_event_id` 로도 사용(at-least-once 가드, codex hotfix #10 P2-D).
- 매핑: `mapUserEventToAudit`(`:376`, 15 row, 미매핑 `keycloak.event.unknown:<TYPE>`), `mapAdminEventToAudit`(`:445`, USER/ROLE_MAPPING/GROUP_MEMBERSHIP/CLIENT/REALM, 미매핑 `keycloak.admin.<key>`).
- `classifyAdminEventForSync`(`:427`): USER:UPDATE→`profile`, USER:DELETE→`status`, GROUP_MEMBERSHIP:CREATE/DELETE→`membership`. `ParseIdentityIDFromResourcePath`(user_sync.go `:253`) 로 UUID 파싱.

### 2.2 keycloak_admin_adapter.go — httpapi → audit adapter
- `NewHTTPAPIEventListerAdapter`(`:53`): main.go 가 `httpapi.KeycloakAdminClient` 결과를 audit 미러 struct 로 named-type 변환(`audit ← httpapi` 단방향 의존만 유지). admin event 의 `AuthDetails` 평탄화는 main.go `keycloakAdminEventLister`(`main.go:415`) 에서 수행.

### 2.3 user_sync.go — DevHub `users` 컬럼 sync (ADR-0020 sub-carve C)
- `SyncUserProfile`(`:72`): `GetUserDetails` → email/display_name/status(`Enabled` 반영) UPDATE. DevHub row 없으면(`ErrNotFound`) noop, 그 외 DB error 는 propagate(PR #241 codex P1 — 이전엔 모든 error swallow → stale 위험).
- `SyncUserMembership`(`:120`): `GetUserGroups` → `pickHighestPriorityRole`(`:211`, `devhub-{role}s` group → system_admin>pmo_manager>manager>developer). role 빈 경우 default `developer`.
- `MarkUserDeactivated`(`:170`): USER:DELETE soft delete. Keycloak user 이미 gone 이라 `GetUserByIdPSubject`(idp_subject UNIQUE) 로 lookup(PR #241 codex P1 — 이전엔 identity UUID 를 user_id 로 직접 GetUser → silent noop).
- compile-time check `var _ UserSyncOrgStore = (*store.PostgresStore)(nil)`(`:268`).

### 2.4 metrics.go — Prometheus
| metric | type | label | 의미 |
|---|---|---|---|
| `devhub_keycloak_events_processed_total` | counter | kind, action | audit emit 1건마다. action 은 `normalizeMetricAction`(`:107`) 로 bounded(unknown suffix cardinality 폭발 회피) |
| `devhub_keycloak_event_cursor_lag_seconds` | gauge | cursor_key | now - cursor.LastEventAt. 매 tick |
| `devhub_keycloak_event_pull_errors_total` | counter | kind | tick pull error |
| `devhub_keycloak_user_sync_total` | counter | action | user_sync 성공(profile/membership/status) |
| `devhub_keycloak_user_sync_errors_total` | counter | action | user_sync 실패 |
| `devhub_keycloak_user_sync_lag_seconds` | histogram | - | event ts ↔ DevHub write 차(1s~~17m) |

- 외부 의존: Keycloak Admin REST(`/admin/realms/{realm}/events` + `/admin-events`). 인증은 KeycloakAdminClient(client-credentials).
- 주기: 30s polling. 인증/보안: webhook secret 없음(pull 방식), admin client OIDC 자격은 config.

---

## 3. `internal/auth/` — Keycloak JWKS verifier + stale-while-error + metrics

### 3.1 keycloak_verifier.go
- `KeycloakJWKSVerifier`(`:23`): `IssuerURL`/`JWKSURL`/`ClientID`/`HTTPClient` + `CacheTTL`(default 5m) + `MaxStaleDuration`(default 24h) + RWMutex 보호 캐시(`cachedKeys`/`cachedAt`/`cachedUntil`).
- `VerifyBearerToken`(`:58`): RS256/384/512 만 허용. issuer/audience(ClientID) validation. cached JWKS 1차 시도 → `errKidMismatch`(`:56`) 면 **cache invalidate + forced refetch + 1회 retry**(key rotation 직후 새 kid 가 TTL 만료까지 401 되던 문제 해소). signature/expired/issuer/audience error 는 retry 안 함(security).
- claim 추출: `extractKeycloakRole`(`:429`) 가 `role` → `roles` → `realm_access.roles` → `resource_access.*.roles` 순. multi-role 은 `selectHighestPriorityRole`(`:413`, `devhubRolePriority` `:403`). `extractDisplayName`(`:147`) name>given+family>"".
- **stale-while-error**(`fetchJWKS` `:179`): TTL 만료 → network fetch 시도 → 실패 시 `readStaleCachedKeys`(`:291`) 로 stale 검증 fallback(Keycloak unreachable 시 DevHub uptime 보장). cutoff = `cachedUntil + MaxStaleDuration`(PR #242 codex P1 — 이전 `cachedAt + maxStale` 는 stale window 를 TTL 만큼 축소). stale 만료 시 사용 안 함(revoked key 보호 한도).
- **보안 trade-off**(주석 `:33-39`, `:281-290`): stale 사용 중 revoked key 보호 깨짐. rotation 직후 운영 SOP(강제 재시작 / cache flush endpoint) 는 별도 carve.

### 3.2 metrics.go
| metric | type | label | 의미 |
|---|---|---|---|
| `devhub_jwks_stale_while_error_total` | counter | result(ok/fail) | stale fallback 진입(ok=stale 사용, fail=stale 없음/만료) |
| `devhub_jwks_stale_age_seconds` | histogram | - | stale 사용 시 cache age(1m~~68h) |

---

## 4. `internal/commandworker/` + `internal/serviceaction/` — command dry-run / live executor

### 4.1 commandworker/worker.go — dry-run 워커
- `Worker.ProcessOnce`(`:27`): `ListRunnableDryRunCommands(limit)`(default 25) → 각 command `running` → `succeeded` 로 상태 전이(`executor=dry_run`, external side effect 없음) + `PublishCommandStatus`(realtime hub) publish.
- `Run`(`:61`): interval ≤0 면 1s, 즉시 1차 + ticker. **tick error 시 loop 종료**(반환) — cron 류(audit/dreq/homelab)와 달리 fail-fast.

### 4.2 commandworker/live_worker.go — live service action 워커
- `LiveWorker.ProcessOnce`(`:26`): Store/Executor nil 이면 no-op. `ClaimRunnableLiveServiceActionCommands`(claim 시 running 마킹) → `ExecuteServiceAction`. 성공 시 result 에 `executor=service_action` + `succeeded`, 실패 시 `failed`(error 메시지 payload) — 개별 command 실패는 loop 계속(`continue`).

### 4.3 serviceaction/executor.go
- `NewExecutor(mode, services, actions)`(`:25`): mode 가 `simulation`(`SimulationMode` `:13`) 아니면 `ErrUnsupportedExecutorMode`. allowlist 는 CSV → set(`csvSet` `:68`).
- `ExecuteServiceAction`(`:40`): `command_type != "service_action"` 거부. `AllowedServices`(TargetID) + `AllowedActions`(ActionType) allowlist 검사 — **빈 allowlist 면 전부 거부**(`allowed` `:61`, fail-closed). 통과 시 `simulated:true` map 반환(외부 side effect 없음).
- live 모드 활성 조건: `SERVICE_ACTION_EXECUTOR_MODE` 설정(main.go `:68`). 현재 지원 모드는 `simulation` 뿐 — 실제 외부 실행 executor 미구현(부채).

---

## 5. `internal/config/` — env 로딩 / 검증

- `Config`(`config.go:10`): Port/DBURL/Gitea*/BackendAIURL/Env/IdPProvider/AuthDevFallback/OnboardingGateEnabled/ProjectModel/ServiceAction*/OIDC*/Keycloak*/InfraAgentToken/HomeLab*/DREQToken*/KeycloakEventListener*/KeycloakWebhookSecret.
- `Load()`(`:122`): 모든 필드를 env 에서 trim 후 매핑. `OnboardingGateEnabled` 만 `envBoolDefault(..., true)`(2026-05-21 default ON flip, `0`/`false` 로 rollback). 나머지 bool 은 `envBool`(default false).
- `Validate(hasVerifier)`(`:170`): IdP 가 `keycloak` 아니면 거부. `DEVHUB_ENV=prod` 일 때만 fail-fast — verifier 없으면 거부, `AuthDevFallback=1` 이면 거부. dev 모드는 무제약.
- `normalizeIDPProvider`(`:223`): 빈 값 → `keycloak`(Keycloak 단일화, ADR-0019). `normalizeProjectModel`(`:236`): legacy/v2/hybrid, default `hybrid`.
- helper: `envBool`/`envBoolDefault`/`envInt`/`envInt64`.

---

## 6. `internal/devrequest/` — DREQ intake token cron

### 6.1 intake_token_cron.go
- `IntakeTokenStore`(`:11`): `HardRevokeExpiredIntakeTokens(before) []string` / `CountExpiringSoonIntakeTokens` / `CountStaleIntakeTokens`.
- `RunIntakeTokenCron`(`:51`): interval default 10m, 즉시 1차 tick 후 ticker. tick error 는 log 만(blast radius 격리, ctx cancel 까지 계속).
- `runIntakeTokenCronTick`(`:77`):
  1. 만료 token hard-revoke → revoke 된 ID 마다 audit emit(`dev_request_intake_token.auto_revoked`) + counter.
  2. `ExpiringSoonThreshold>0` 면 `now+threshold` 카운트 gauge, ≤0 면 0 emit.
  3. `StaleThreshold>0` 면 `now-threshold` 카운트 gauge, ≤0 면 0 emit(disabled, "no data" 와 구분).

### 6.2 metrics.go
| metric | type | 의미 |
|---|---|---|
| `devhub_intake_token_expiring_soon` | gauge | 임박(default 24h) 활성 token 수 |
| `devhub_intake_token_stale` | gauge | last_used_at 노후(default 30d) token 수 |
| `devhub_intake_token_auto_revoked_total` | counter | cron 자동 revoke 누계 |

- 활성 조건: `DEVHUB_DREQ_TOKEN_CRON_ENABLED=1` + store 가 interface 만족(main.go `:247`). ADR-0017 §6 (a)+(c)+(d).

---

## 7. `internal/normalize/` — webhook payload 정규화

- `Processor.Process`(`gitea.go:22`): `Normalize` → changeSet → Sink 에 순차 upsert(Repository→Sender→Issue→PullRequest→CIRun→Risk) → `MarkWebhookEventProcessed`. 실패 시 `MarkWebhookEventFailed`, ignored 면 `MarkWebhookEventIgnored`. Sink nil 이면 검증만.
- `Normalize`(`:82`): event_type 별 분기.
  - `issues`/`issue`, `pull_request`, `action_run`/`workflow_run`(CIRun + 실패 시 `riskFromCIRun` `:134`), `push`(repo/sender metadata 만), 그 외 ignored.
  - 비-ignored 인데 repository payload 없으면 error.
- 정규화 helper: `normalizeIssueState`/`normalizePullRequestState`(merged 우선)/`normalizeCIStatus`(`:427`, completed+conclusion 매핑, failure/error→failed). `loginOf`(`:367`) login→username fallback, `firstInt64`/`firstTime`/`stringify`(`id`/`run_number` 가 `any` 타입 — 숫자/문자 혼재 흡수).
- 외부 의존 없음(순수 변환). metric 없음.

---

## 8. `internal/hrdb/` — HR DB Mock + Postgres adapter

- `Person`(`mock.go:10`) + `Client` interface(`:20`, `Lookup`).
- `MockClient`(`:25`): 3행 하드코딩 in-memory(yklee/akim/sjones). `Lookup`(`:39`) systemID+employeeID+name 정확 매치(EqualFold), email 은 `<systemid>@example.com` 합성. 미스 → `ErrPersonNotFound`.
  - 주의: interface 선언은 `(*Person, error)` 인데 Mock/Postgres 구현은 `(string,string,string,error)` 시그니처(`router.go::HRDBClient` 가 실제 contract). `Client` interface 와 실제 사용 시그니처 불일치(dead interface, 부채 §발견사항).
- `PostgresClient`(`postgres.go:19`): `hrdb.persons` 조회(ADR-0008). `EmailFallbackDomain`(`DEVHUB_HR_EMAIL_FALLBACK_DOMAIN`, default `example.com`) 로 NULL email COALESCE. `pgx.ErrNoRows` → `ErrPersonNotFound`.
- 현재 main.go 는 **Mock 만 wire**(`main.go:129`) — Postgres adapter 는 구현 존재하나 미배선(부채 §발견사항).

---

## 9. `internal/integrations/adapters/` — HomeLab puller (file/http) + health policy + metrics

### 9.1 contract.go / homelab.go
- `InfraSnapshotStore`(`contract.go:10`): `SaveInfraSnapshot` / `LoadLatestInfraSnapshot`.
- `HomeLabPuller`(`:27`): `PullSnapshot() HomeLabRawSnapshot`. `HomeLabRawSnapshot`(`:32`) agent_id/snapshot_at/trace_id/nodes/services.
- `HomeLabAdapter`(`homelab.go:14`): Store+Puller+HealthPolicy.
  - `IngestSnapshot`(`:36`): ingest_id/agent_id/snapshot_at/nodes·services 검증 → SaveInfraSnapshot → 성공 시 metric observe.
  - `PullAndIngest`(`:81`): Pull → `NormalizeSnapshot`(`:99`, RFC3339 파싱 + ingest_id 합성 + degraded providers 수집) → Ingest.
  - `collectDegradedProviders`(`:128`): service health_status 가 degraded set 에 있으면 ProviderKey 1건 반환. default policy = warning/degraded/down + `homelab-agent`.

### 9.2 homelab_file_puller.go / homelab_http_puller.go (ADR-0015 §6 (1) size guard)
- `HomeLabFilePuller`(`:19`): `os.Stat` 사전 size 검사 + `os.Open` + streaming decode. `MaxBytes>0` 면 `io.LimitReader(f, max+1)`(stat 후 grow race 가드). **trailing data 검증**(`dec.Token()` EOF, codex hotfix #8 P2 #3 — 후행 JSON object silent 무시 회귀 방지). 0 이면 unlimited(legacy).
- `HomeLabHTTPPuller`(`:19`): Bearer token(optional), retry(RetryMax/RetryBackoff, 5xx·429 retryable), Content-Length 사전 reject + `io.LimitedReader`. **oversized 는 `LimitedReader.N==0` 으로만 명시 감지**(codex hotfix #8 P1 — `ErrUnexpectedEOF` 는 transient transport 실패도 포함하므로 oversized 단정 금지, 그 외는 retryable).

### 9.3 homelab_pull_loop.go
- `RunHomeLabPullLoop`(`:9`): interval default 30s, 즉시 1차 + ticker. 각 run 은 `context.WithTimeout(ctx, interval)`. error 는 `onError` 콜백 + metric `error`, 성공 시 metric `success` + last_success.

### 9.4 metrics.go
| metric | type | label | 의미 |
|---|---|---|---|
| `devhub_homelab_pull_runs_total` | counter | result | pull 결과 |
| `devhub_homelab_pull_duration_seconds` | histogram | - | pull+ingest 소요 |
| `devhub_homelab_snapshot_services` | gauge | - | 최신 snapshot service 수 |
| `devhub_homelab_degraded_providers` | gauge | - | degraded provider 수 |
| `devhub_homelab_last_success_unixtime` | gauge | - | 마지막 성공 시각 |

- 외부 의존: HomeLab agent snapshot endpoint(HTTP) 또는 로컬 JSON fixture(file). 인증: HTTP Bearer token(`DEVHUB_HOMELAB_PULL_TOKEN`, optional). push ingest(API-77)는 별도 `InfraAgentToken`(`X-Infra-Agent-Token`) 경로.

---

## 10. `main.go` — 백그라운드 워커 기동

### 10.1 부팅 순서 요약
1. `config.Load` → `DB_URL` 없으면 `log.Fatalf`(startup 거부, `:81`).
2. PostgresStore 1개를 모든 store interface 에 공유 wire(`:53-65`).
3. JWKS verifier 구성(`MaxStaleDuration` env wire `:93`) → `cfg.Validate`(prod 면 verifier 필수).
4. KeycloakAdminClient(4개 env 모두 있을 때만, `:110`).
5. HRDB Mock wire(`:129`), Router 구성(`:132`).
6. 각 워커 goroutine 기동 후 `router.Run`.

### 10.2 워커 기동 조건 / 주기 표
| 워커 | 기동 조건 | 주기 | error 처리 | 근거 |
|---|---|---|---|---|
| command Worker (dry-run) | `pgStore != nil` (항상) | 2s | error 시 loop 종료 | `main.go:67,166` |
| LiveWorker (service action) | `SERVICE_ACTION_EXECUTOR_MODE` 설정 시 | 2s | 개별 실패는 continue, store error 시 종료 | `:68,173` |
| HomeLab pull loop | `DEVHUB_HOMELAB_PULL_ENABLED=1` + (FILE 또는 URL 설정) | `DEVHUB_HOMELAB_PULL_INTERVAL` (default 30s) | onError 콜백, loop 계속 | `:180` |
| DREQ intake token cron | `DEVHUB_DREQ_TOKEN_CRON_ENABLED=1` + store 만족 | `..._INTERVAL` (default 10m) | tick error log, 계속 | `:247` |
| Keycloak event listener | `DEVHUB_KEYCLOAK_EVENT_LISTENER_ENABLED=1` + KeycloakAdminClient + auditStore + eventCursorStore | `..._INTERVAL` (default 30s) | tick error log, 계속 | `:291` |
| Onboarding pending_review gauge | organizationStore 가 `OnboardingPendingReviewCounter` 만족 (항상) | 60s 고정 | error log | `:358` |
| Gitea sync worker | `pgStore != nil` (항상) | 30s 고정 | error log, 계속 | `:372` |

### 10.3 wire 어댑터
- `keycloakAdminEventLister`(`:390`): KeycloakAdminClient → audit.HTTPAPIEventLister(admin event 의 `AuthDetails` 평탄화).
- `buildKeycloakEventAuditEmitter`(`:445`): event → `audit_logs` INSERT(best-effort, `source_type=keycloak_event`, `sourceEventID`=dedup hash, partial UNIQUE INDEX 000032 이 중복 차단). actor_login 은 payload 의 user_id/auth_user_id.
- `buildIntakeTokenAuditEmitter`(`:472`): DREQ cron audit emit(best-effort).
- UserSync dispatcher(`:315`): organizationStore 있을 때만. profile/membership/status → audit user_sync helper + metric.

---

## 11. `backend-ai/` — Python FastAPI 스켈레톤

- `main.py`(15줄): FastAPI 앱 + `GET /health`(`{"status":"ok","service":"backend-ai"}`) 1개 엔드포인트. uvicorn 0.0.0.0:8000.
- `# TODO: gRPC Server for AnalysisService`, `# TODO: AI Logic for Log Analysis`(`main.py:10-11`) — 핵심 기능 전부 미구현.
- `requirements.txt`: fastapi/uvicorn/grpcio/grpcio-tools(grpc 의존 명시되나 .proto/서버 코드 부재).
- `Dockerfile`: python:3.12-slim, `.build/site-packages` vendored(사내 mirror build), `main.py` 만 COPY.
- backend-core 연동: `cfg.BackendAIURL`(`BACKEND_AI_URL`)이 `RuntimeSnapshotProvider` 에 주입되어 health snapshot 에만 사용. 실제 AI 분석/gRPC 호출 경로 없음.
- **현재 구현 수준**: health probe 만 있는 빈 스켈레톤. 미구현: gRPC AnalysisService, 로그 분석 AI 로직, backend-core ↔ backend-ai 간 실 호출(snapshot health 표시 외).

---

## 발견 사항 (불일치 / stale / 부채)

1. **env token fallback 유출 방지는 적용됨(#359)** — `gitea/worker.go:105-126` `resolveSyncConfig` 가 명시 provider 해석 시 env token fallback 을 차단(lookup 실패 시에만 env fallback). `auth.go:25-52` 도 자격 미설정 시 `ok=false` 로 skip. 회귀 가드로 동작 중이며 stale 아님(확인 목적).

2. **평문 secret 노출 (#6 carve, 미해소)** — `IntegrationProvider.CredentialsRef`(`domain/application.go:235`)는 webhook HMAC secret 으로 사용(`integration_registry.go:98,147`)되는데 GET 응답 직렬화(`integration_registry.go:36` `"credentials_ref": p.CredentialsRef`)에 **raw 그대로 노출**. 반면 신규 `api_token`/`auth_secret` 은 write-only(`api_token_set`/`auth_secret_set` bool 만, `:42,48`). 즉 신규 필드는 write-only 패턴인데 legacy `credentials_ref` 는 평문 노출 — 보안 gap 잔존(#6).

3. **polling latency 30s + SPI push 병행(미전환)** — audit 폴러 default 30s(`keycloak_event_puller.go:111`, main.go `:296`). 한편 Keycloak SPI **push** 엔드포인트 `keycloak_events_webhook.go`(`X-Webhook-Secret`)도 존재 — 즉 push 경로가 추가됐으나 audit_logs emit 의 정식 경로는 여전히 polling 워커이며, push→audit 통합/폴러 폐기는 미완(push 와 poll 이 동시 존재 시 중복은 dedup hash + 000032 partial UNIQUE 가 흡수). SPI push 단일화는 미전환 부채.

4. **backend-ai 미구현** — `backend-ai/main.py:10-11` TODO 2건(gRPC AnalysisService + Log Analysis AI). grpcio 의존만 선언, 실제 서버/proto/연동 부재(§11).

5. **Gitea webhook 헤더 alias 처리 — 두 핸들러 분기** — 전용 핸들러 `gitea_webhook.go:54` 는 `X-Gitea-Signature`/`X-Gogs-Signature` 만 수용. 범용 ingest `integration_registry.go:599` 는 `X-Integration-Signature`→`X-Gitea-Signature`→`X-Gogs-Signature` fallback(헤더 불일치 정정 #358). 두 경로가 헤더 수용 범위가 달라(전용은 X-Integration-* 미수용) 잠재적 혼선 여지. 의미 버그는 아니나 일관성 부채.

6. **hrdb dead interface + Postgres adapter 미배선** — `hrdb.Client` interface 선언(`mock.go:20`)은 `(*Person, error)` 인데 실제 구현/사용 시그니처는 `(string,string,string,error)`(`postgres.go:28`, `router.go::HRDBClient`). interface 가 실제로 만족되지 않는 dead 선언. 또한 `PostgresClient`(ADR-0008 production adapter)는 구현 존재하나 main.go 는 `NewMockClient` 만 wire(`main.go:129`) — production HR 조회 경로 미연결.

7. **dead config / 미사용 env** — `KeycloakAdminClientSecret` 등 admin 자격은 4개 모두 있어야 admin client 활성(`main.go:110`); 부분 설정 시 silent skip(log 만). `serviceaction` 은 `simulation` 모드만 구현 — `SERVICE_ACTION_EXECUTOR_MODE` 의 다른 값은 startup fatal(`executor.go:30`), 실 실행 executor 부재. live 워커는 실질적으로 simulation dry-run 의 변형(외부 side effect 없음).

8. **gitea sync 워커 metric 부재** — audit/auth/devrequest/adapters 는 Prometheus metric 을 갖췄으나 `gitea/` 패키지는 metrics 파일 없음. sync job 진행/실패/지연이 log(`[Gitea Sync Worker]`) 로만 노출 — 운영 dashboard 가시성 부족(부채). 30s 고정 주기도 env override 불가(`main.go:375` 하드코딩).

9. **ADR 참조 정합 (양호)** — 주석의 ADR 참조는 대체로 정합. auth/devrequest 헤더가 metrics Help 에서 `ADR-0020 §5.6` / `ADR-0017 §6` 인용. 단 `user_sync.go:144-147` 주석이 `lazy_auto_create.go` 가 issue #284 sprint 에서 삭제됐음을 명시 — 삭제된 파일 default role 정합을 코드 주석으로만 보존(코드 자체는 정합, 문서-코드 drift 위험 낮음). `pickHighestPriorityRole`(user_sync.go) 와 `selectHighestPriorityRole`(keycloak_verifier.go) 가 동일 priority 로직을 **중복 구현**(두 곳 모두 system_admin>pmo_manager>manager>developer) — DRY 부채.

10. **stale-while-error 보안 trade-off 명시 부채** — `keycloak_verifier.go:33-39` 가 revoked key 보호가 stale window(default 24h) 동안 깨질 수 있음을 인지하고, rotation 직후 강제 재시작 / cache flush endpoint 도입을 별도 carve 로 미룸(미구현). 운영 SOP 의존.
