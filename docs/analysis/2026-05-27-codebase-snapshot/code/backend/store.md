# backend-core/internal/store — 데이터 계층 분석

- 문서 목적: `backend-core/internal/store/` 의 PostgreSQL 데이터 접근 계층(`PostgresStore`)을 파일·도메인·메서드·트랜잭션·에러 분기 관점에서 정밀 분석한다.
- 범위: 구현 25 파일 중 Go 소스 14개(`*.go`) + 통합/단위 테스트 11개(`*_test.go`). in-memory fake 는 `internal/httpapi/` 에 거주하므로 parity 비교 대상으로만 다룬다.
- 대상 독자: backend 유지보수자, 스키마 변경 작업자, 신규 store 메서드 추가자.
- 상태: snapshot (2026-05-27, main `cf19c94`)
- 관련 문서: `domain.md`, `migrations.md`

## 1. 파일 구성 개요

`PostgresStore` 는 단일 구조체(`postgres.go:34`)이며, 모든 메서드가 동일 `*pgxpool.Pool` 위에 receiver method 로 분산 정의된 partial-type 패턴이다. 도메인별로 파일이 쪼개져 있고 한 struct 에 메서드가 모인다.

| 파일 | 담당 도메인 | 핵심 책임 |
|------|-------------|-----------|
| `postgres.go` (1856 L) | webhook_events / gitea mirror(repositories·gitea_users·issues·pull_requests·ci_runs) / risks / commands / audit(command flow) | Sink upsert + Command 생성·승인·claim + Repository list. `ErrDuplicateEvent`/`ErrNotFound`/`ErrConflict` 정의처. |
| `applications.go` (1057 L) | applications / application_repositories / scm_providers / projects / project_repositories / repository draft | Application·Project CRUD + repo draft/publish + `CreateProjectWithRepositoryPayload` 단일 tx |
| `repository_ops.go` (553 L) | pr_activities / build_runs / quality_snapshots | Repository 운영 지표 read-only 조회 + `ComputeApplicationRollup` 가중치 집계 |
| `dev_requests.go` (279 L) | dev_requests | DREQ CRUD + status 전이(`Transition`/`Reassign`/`MarkRegistered`) |
| `dev_requests_promote.go` (155 L) | dev_requests → applications/projects | promote 단일 tx (`RegisterDevRequestWithNew{Application,Project}`) |
| `dev_request_intake_tokens.go` (365 L) | dev_request_intake_tokens | intake 토큰 lookup/CRUD + 만료·stale cron 조회 + CTE atomic update |
| `integration_registry.go` (605 L) | integration_providers / integration_bindings / integration_sync_jobs | 외부 연동 registry CRUD + SKIP LOCKED sync job 큐 |
| `integrations.go` (178 L) | project_integrations | Jira/Confluence 연동 CRUD (scope polymorphism) |
| `users_units.go` (1264 L) | users / org_units / unit_appointments | 사용자·조직 CRUD + 계층 집계 CTE + onboarding + leader single-source tx. `isUniqueViolation`/`isForeignKeyViolation` 정의처 |
| `postgres_rbac.go` (254 L) | rbac_policies | RBAC role CRUD + audit invariant + `isCheckViolation` 정의처. `ErrSystemRoleImmutable`/`ErrRoleInUse`/`ErrAuditInvariantViolation` 정의처 |
| `audit_logs.go` (246 L) | audit_logs (standalone) | source_event_id dedup INSERT + 목록 조회 |
| `event_cursors.go` (81 L) | event_cursors | Keycloak event polling 커서 upsert (`EventCursorStore` interface) |
| `realtime_tickets.go` (84 L) | realtime_tickets | WS ticket single-use consume (ADR-0024 §6 carve 6) |
| `infra_snapshots.go` (58 L) | infra_service_snapshots | infra agent 스냅샷 save/load-latest |

테스트 파일: `postgres_integration_test.go`, `applications_integration_test.go`, `integrations_integration_test.go`, `repository_ops_integration_test.go`, `audit_logs_dedup_integration_test.go`, `dev_request_intake_tokens_integration_test.go`, `integration_sync_jobs_integration_test.go`, `realtime_tickets_integration_test.go`, `projects_creation_integration_test.go` (실 DB 필요 통합 테스트) + `postgres_rbac_test.go`, `users_units_test.go` (단위 테스트).

## 2. 파일별 핵심 메서드 표

### postgres.go (Sink + Command + webhook)
| 메서드 | 동작 | 에러 분기 |
|--------|------|-----------|
| `SaveWebhookEvent` | `ON CONFLICT (dedupe_key) DO NOTHING RETURNING id` | `pgx.ErrNoRows`→`ErrDuplicateEvent` (`postgres.go:90`) |
| `UpsertRepository` | SCM mirror upsert (§3 참조) | raw err |
| `UpsertUser/Issue/PullRequest/CIRun/Risk` | `ON CONFLICT … DO UPDATE` mirror | raw err |
| `CreateRiskMitigationCommand` / `CreateServiceActionCommand` | command + audit 단일 tx, idempotency 선조회 | `IdempotencyKey` 충돌 시 기존 row 재조회 후 `found=true` |
| `GetCommand` / `ListRunnableDryRunCommands` / `ListRunnableLiveServiceActionCommands` | 조회 | `pgx.ErrNoRows`→`ErrNotFound` |
| `ClaimRunnableLiveServiceActionCommands` | `FOR UPDATE SKIP LOCKED` CTE → `status='running'` | — |
| `UpdateCommandStatus` | status UPDATE | `ErrNotFound` |
| `ApproveCommand`/`RejectCommand` → `reviewServiceActionCommand` | `FOR UPDATE` select + 상태 가드 + audit | type≠service_action ∨ status≠pending ∨ ¬requires_approval → `ErrConflict` (`postgres.go:1154`) |
| `ListRepositories` / `GetRepositoryByID` | `LEFT JOIN integration_providers` 로 provider_key derive | `ErrNotFound` |
| `ListRepositoriesByProvider` | provider_id 로 import 표식 조회 | — |

### applications.go
| 메서드 | 동작 | 에러 분기 |
|--------|------|-----------|
| `ListApplications` | count + list, `applicationsSearchPredicate` 공유 (key/name/owner/leader/unit/연결 repo·project ILIKE) | — |
| `CreateApplication` | `applicationsInsertQuery` (archived 시 archived_at 자동) | unique/FK violation → `ErrConflict` |
| `UpdateApplication`/`ArchiveApplication` | status='archived' 전이 시 archived_at = `COALESCE(archived_at, NOW())` | `ErrNotFound`, FK→`ErrConflict` |
| `CountActiveApplicationRepositories` | planning→active 가드용 active link 카운트 | — |
| `CreateApplicationRepository`/`UpdateApplicationRepositorySync`/`DeleteApplicationRepository` | link CRUD, sync_error CASE 명시 cast(`NULL::boolean`/`NULL::timestamptz`) | unique/FK→`ErrConflict`, affected=0→`ErrNotFound` |
| `ListSCMProviders`/`UpdateSCMProvider` | catalog. adapter_version 은 UPDATE 미포함(배포 파이프라인 전용) | `ErrNotFound` |
| `CreateProject`/`UpdateProject`/`ArchiveProject` | `projectsInsertQuery` 공유 | unique/FK→`ErrConflict` |
| `CreateProjectWithRepositoryPayload` | **단일 tx**(§3) | unique/FK/check→`ErrConflict` |
| `CreateRepositoryDraft` | `source='system', repository_status='draft'` row | unique→`ErrConflict` |
| `MarkRepositoryDraftPublishRequested` | `WHERE repository_status='draft'` UPDATE | `ErrNotFound` |
| `CreateProjectRepository`/`DeleteProjectRepository`/`ListProjectRepositories` | N:M link CRUD | unique/FK/check→`ErrConflict`, affected=0→`ErrNotFound` |

### repository_ops.go
| 메서드 | 동작 |
|--------|------|
| `ListRepositoryActivity` | pr_activities + build_runs 기간 집계 (commit 활동은 미구현) |
| `ListRepositoryPullRequests`/`ListRepositoryBuildRuns`/`ListRepositoryQualitySnapshots` | 페이지네이션 read-only |
| `ComputeApplicationRollup` | weight_policy(equal/repo_role/custom) 정규화 + data_gap/fallback 계산. custom 합계 검증(`±CustomWeightTolerance`), missing fallback 후 재정규화(`repository_ops.go:415`) |
| `CountApplicationCriticalWarnings` | rollup 재사용 critical 카운트 |

### dev_requests.go / dev_requests_promote.go
| 메서드 | 동작 | 에러 분기 |
|--------|------|-----------|
| `CreateDevRequest` | row insert (invalid_intake 도 audit 보존 목적 저장) | unique/FK→`ErrConflict` |
| `GetDevRequest`/`GetDevRequestByExternalRef`/`ListDevRequests` | 조회 (`ANY($1::text[])` status 필터) | `ErrNotFound` |
| `TransitionDevRequestStatus` | reason/target CASE 정리(`status='registered'` 외엔 target NULL) | `ErrNotFound` |
| `ReassignDevRequest` | assignee 변경 | FK→`ErrConflict` |
| `MarkDevRequestRegistered` | `WHERE status IN ('pending','in_review')`, rejected_reason=NULL | `ErrNotFound` |
| `RegisterDevRequestWithNewApplication`/`…Project` | **promote 단일 tx**(§3) | unique/FK/check→`ErrConflict`, `pgx.ErrNoRows`→`ErrNotFound` |

### integration_registry.go
| 메서드 | 동작 | 에러 분기 |
|--------|------|-----------|
| `List/Get…ByID/…ByKey/Create/UpdateIntegrationProvider` | provider CRUD, 평문 secret 컬럼 RETURNING 포함 | unique→`ErrConflict`, `ErrNotFound` |
| `DeleteIntegrationProvider` | binding count>0 면 차단 (FK guard) | bindingCount>0→`ErrConflict`, `ErrNotFound` |
| `CreateIntegrationSyncJob` | queued job insert | FK→`ErrNotFound` |
| `AcquireNextQueuedSyncJob` | `provider_type='scm'` gate + `FOR UPDATE OF j SKIP LOCKED` (§3) | `pgx.ErrNoRows`→`ErrNotFound` |
| `UpdateIntegrationSyncJobStatus` | status UPDATE | affected=0→`ErrNotFound` |
| `List/Create/Get/Update/DeleteIntegrationBinding` | binding CRUD | unique/FK/check→`ErrConflict`, `ErrNotFound` |

### users_units.go
| 메서드 | 동작 | 에러 분기 |
|--------|------|-----------|
| `ListUsers`/`GetUser`/`GetUserByIdPSubject`/`ListUnitMembers` | 사용자 조회 + appointments 합성 | `pgx.ErrNoRows`→`ErrNotFound` |
| `CreateUser`/`UpdateUser`/`DeleteUser` | 동적 SET 절 UPDATE | unique→`ErrConflict`, FK→`ErrNotFound`, affected=0→`ErrNotFound` |
| `SubmitOnboarding` | INSERT 또는 UPDATE 단일 tx + `FOR UPDATE` lock | 이미 완료→`ErrConflict`, unit FK→`ErrNotFound` |
| `ConfirmUserReview` | pending_review→reviewed | affected=0→`ErrNotFound` |
| `CountPendingReview` | Prometheus gauge용 카운트 | — |
| `SetIdPSubject` | idp_subject 캐시 | unique→`ErrConflict`, affected=0→`ErrNotFound` |
| `GetHierarchy` | RECURSIVE CTE (descendants/depths/ranked_appointments/canonical) 로 dedupe 카운트 | — |
| `SearchOrgUnits`/`GetOrgUnit` | typeahead/단건 | `ErrNotFound` |
| `CreateOrgUnit`/`UpdateOrgUnit`/`DeleteOrgUnit` | unit CRUD + leader appointment sync tx | unique→`ErrConflict`, FK→`ErrNotFound`/`ErrConflict` |
| `ReplaceUnitMembers`/`UpdateHierarchy` | 멤버 일괄 교체 / 좌표 일괄 갱신 tx | unit 없음→`ErrNotFound` |

### postgres_rbac.go / audit_logs.go / event_cursors.go / realtime_tickets.go / infra_snapshots.go
| 메서드 | 동작 | 에러 분기 |
|--------|------|-----------|
| `List/Get/Create/UpdateRBACRolePermissions/UpdateRBACRoleMetadata/DeleteRBACRole` | RBAC role CRUD | system role→`ErrSystemRoleImmutable`, in-use→`ErrRoleInUse`, audit invariant CHECK→`ErrAuditInvariantViolation`, unique→`ErrConflict`, `ErrNotFound` |
| `CreateAuditLog` | `ON CONFLICT (source_type, source_event_id) … DO NOTHING` + dedup 시 기존 row 재조회 | `pgx.ErrNoRows`→기존 row SELECT 반환 |
| `ListAuditLogs` | 필터 목록 | — |
| `GetEventCursor`/`UpsertEventCursor` | Keycloak 커서 upsert | `pgx.ErrNoRows`→`ErrNotFound` |
| `InsertRealtimeTicket`/`ConsumeRealtimeTicket`/`DeleteExpiredRealtimeTickets` | ticket single-use(§3) | miss→`(false, nil)` (NOT `ErrNotFound`) |
| `SaveInfraSnapshot`/`LoadLatestInfraSnapshot` | infra 스냅샷 | `pgx.ErrNoRows`→`ErrNotFound` |

## 3. 트랜잭션 패턴 정리

`PostgresStore` 는 다음 위치에서 `s.pool.Begin(ctx)` + `defer tx.Rollback(ctx)` + 명시 `tx.Commit(ctx)` 패턴을 쓴다.

1. **Command 생성** (`CreateRiskMitigationCommand` `postgres.go:514`, `CreateServiceActionCommand` `postgres.go:714`): commands INSERT + audit_logs INSERT 를 한 tx 로 묶어 "command 가 audit 없이 생기는" 부분 실패를 차단. idempotency_key 충돌 시 INSERT 실패 후 기존 row 를 재조회해 `found=true` 반환.

2. **Command 승인/거절** (`reviewServiceActionCommand` `postgres.go:1118`): `SELECT … FOR UPDATE` 로 row lock 후 `command_type='service_action' AND status='pending' AND requires_approval` 가드. 위반 시 `ErrConflict`. UPDATE + audit INSERT 동일 tx.

3. **DREQ promote** (`RegisterDevRequestWithNewApplication` `dev_requests_promote.go:40`, `…Project` `:114`): application/project INSERT + (옵션) primary repo link + `dreqMarkRegisteredUpdateQuery`(`WHERE status IN ('pending','in_review')`) 를 단일 tx 로. application/project insert 실패(중복 key 등) 시 dev_request 상태 갱신도 rollback. ADR-0013 §5.

4. **`CreateProjectWithRepositoryPayload`** (`applications.go:849`): codex #349 P2 atomicity. `repoPayload != nil` 이면 `createRepositoryTx`(`applications.go:921`, `source='system'` upsert) → project INSERT → `project_repositories` link INSERT 를 모두 한 tx 로. project insert 실패 시 동반 repository 생성도 rollback → "고아 repository 없음". `CreateRepositoryForProject` 를 tx-aware `createRepositoryTx` 로 전환한 결과.

5. **single-leader demote-then-promote** (`UpdateOrgUnit` `users_units.go:1167`): `SELECT id FROM org_units WHERE unit_id=$1 FOR UPDATE` 로 unit row 직렬화 → 본 UPDATE → `UPDATE unit_appointments SET role='member' WHERE role='leader'`(demote) → 신규 leader `INSERT … ON CONFLICT DO UPDATE SET role='leader'`(promote). 동시 admin 작업이 인터리브해 두 leader 가 남는 것을 방지. partial unique index `unit_single_leader_idx`(migration 000019)가 최종 보증. `CreateOrgUnit`(`users_units.go:1061`)도 leader appointment sync 를 tx 로.

6. **onboarding** (`SubmitOnboarding` `users_units.go:768`): `SELECT onboarding_completed_at … FOR UPDATE` lock 후 INSERT(미존재) 또는 UPDATE(pre-seeded 미완료). 이미 완료면 `ErrConflict`.

7. **RBAC role 삭제** (`DeleteRBACRole` `postgres_rbac.go:190`): is_system 확인 + users.role usage count + DELETE 를 tx 로. COUNT-then-DELETE window race 를 FK violation(23503)→`ErrRoleInUse` 로 흡수.

### UpsertRepository 의 SCM mirror / system-owned 보존 (소유권 분리, migration 000042)

`UpsertRepository`(`postgres.go:163`)의 `ON CONFLICT (full_name) DO UPDATE` 는 SCM mirror 필드(owner_login/name/clone_url/html_url/default_branch/private/gitea_repository_id)만 `EXCLUDED` 로 덮어쓰고, 다음은 보존한다:
- `source = COALESCE(repositories.source, EXCLUDED.source)` — 기존 값 우선
- `provider_id = COALESCE(repositories.provider_id, EXCLUDED.provider_id)` — 기존 값 우선
- `description` 컬럼은 UPDATE SET 절에 **아예 없음** → system-owned 메타라 sync 가 절대 덮어쓰지 않음 (`postgres.go:188-201`)

INSERT 분기는 `source=COALESCE(NULLIF($9,''),'scm')`, `repository_status='active'`, `published_at=NOW()` 로 채운다. import(`ListRepositoriesByProvider`)는 외부 재조회 값을 쓰고, draft 생성(`CreateRepositoryDraft`)은 `source='system'`.

### single-use ticket consume (ADR-0024 §6)

`ConsumeRealtimeTicket`(`realtime_tickets.go:53`)은 `DELETE … WHERE ticket=$1 AND expires_at > NOW() RETURNING …`. 동시 consume 시 한쪽만 row 를 DELETE 하므로 multi-instance 에서 ticket 이 최대 1회만 honor 된다. miss 는 `(RealtimeTicket{}, false, nil)` — `ErrNotFound` 가 **아니라** ok=false 반환 (다른 패턴).

## 4. ErrConflict / ErrNotFound 분기 규약

- 세 sentinel: `ErrDuplicateEvent`, `ErrNotFound`, `ErrConflict` (`postgres.go:16-18`). RBAC 전용: `ErrSystemRoleImmutable`/`ErrRoleInUse`/`ErrAuditInvariantViolation` (`postgres_rbac.go:17-26`).
- PG SQLSTATE 분류 helper: `isUniqueViolation`(23505, `users_units.go:15`), `isForeignKeyViolation`(23503, `users_units.go:23`), `isCheckViolation`(23514, `postgres_rbac.go:32` — `name=""` 이면 모든 CHECK 매치).
- 매핑 관례: unique violation → 대부분 `ErrConflict`. FK violation → 케이스별 `ErrConflict`(application/project/dev_request 생성) 또는 `ErrNotFound`(users 의 missing unit 참조). `pgx.ErrNoRows` → `ErrNotFound` (조회), 단 `CreateAuditLog`/`commandByIdempotencyKey` 는 dedup/idempotency 신호로 해석.
- handler 가 이 sentinel 을 `errors.Is` 로 받아 409/404/422 HTTP 상태로 변환 (store 는 HTTP 무관).

## 5. in-memory fake ↔ production parity

in-memory fake(`memoryApplicationStore` 외)는 `backend-core/internal/httpapi/` 에 있고 `internal/store/` 에는 없다. `applications_test.go:21` 의 `memoryApplicationStore` 가 대표 fake이며 다음 9 파일에서 공유·확장된다: `applications_test.go`, `integration_scm_repositories_test.go`, `organization_test.go`, `dev_requests_test.go`, `audit_test.go`, `application_rollup_test.go`, `commands_test.go`, `gitea_webhook_test.go`, `domain_test.go`.

parity 가 의도적으로 맞춰진 지점:
- `UpsertRepository`(`applications_test.go:640`): production 미러를 그대로 흉내 — 기존 row 면 SCM mirror 필드만 갱신, `existing.Source==""` / `existing.ProviderID==""` 일 때만 채우고 `Description` 은 보존(주석 명시 "production PostgresStore.UpsertRepository ON CONFLICT 미러"). 신규 row 는 `Source` 빈값을 `RepositorySourceSCM` 으로 기본 채움.
- `DeleteIntegrationProvider`(`applications_test.go:691`): binding count>0 → `ErrConflict` FK guard mirror.
- `UpdateIntegrationBinding`(`applications_test.go:784` 부근): provider_id 변경 시 신규 provider 존재 검증 — production FK guard mirror.

parity 한계(주의):
- fake 는 메모리 map 기반이라 PG CHECK 제약(예: `applications_archived_consistency`, `dev_requests_registered_consistency`, `rbac_policies_audit_invariant`)을 **직접 강제하지 않는다**. CHECK 위반 경로(`isCheckViolation`→`ErrConflict`)는 실 DB 통합 테스트에서만 커버된다.
- fake 의 unique 충돌은 명시 루프로 흉내내므로 production 의 partial unique index(예: standalone project key, source_event_id dedup) semantics 와 1:1 보장이 아니다.

## 발견 사항 (불일치/stale/부채)

1. **평문 secret 컬럼 raw 노출/저장** — `integration_registry.go` 의 모든 provider 쿼리가 `credentials_ref`, `api_token`, `auth_secret` 을 평문 컬럼에 저장하고 SELECT/RETURNING 에 그대로 포함한다 (`integration_registry.go:107-121`, `:240-248`, `:308-317`). `auth_secret`/`api_token` 은 핸들러 레이어에서 write-only(`<field>_set` bool)로 가려지지만 **store 레이어는 raw 값을 그대로 scan**(`domain.IntegrationProvider.APIToken`/`AuthSecret`). `credentials_ref` 는 raw 노출이 알려진 보안 gap(#6 평문 저장 carve, MEMORY `feedback_writeonly_secret_response_pattern`). DB 에 KMS/at-rest 암호화 부재.

2. **migration prefix 충돌 잔흔 (000042 #363↔#368 → #371 000044 재번호)** — `000044_projects_key_unique_active_only.up.sql:1-3` 주석이 "원래 000042 로 작성됐으나 #363 의 000042_repositories_source_provider 와 prefix 충돌 → 000044 로 재번호" 라고 명시. 점검(#371)으로 정정됐으나 store 코드 자체엔 흔적 없음(번호만 이동). 동일 sprint 동시 PR 의 migration 번호 선점 패턴(MEMORY `feedback_concurrent_migration_prefix_collision`).

3. **scm_provider → provider_id 단일화(000045) 흔적이 store 에 잔존** — `createRepositoryTx`(`applications.go:921`)의 시그니처/주석은 여전히 `scmProvider string` 을 받아 `clone_url/html_url` placeholder 만 만들고 `repositories.scm_provider` 컬럼에는 쓰지 않는다(컬럼이 000045 로 DROP 됨). `RepositoryCreatePayload.SCMProvider`(`applications.go:679`)도 placeholder 용도로만 남음. `CreateRepositoryDraft` 주석(`applications.go:685`)은 "migration 000045 — 구 scm_provider key 컬럼 통합" 을 명시하며 provider_key→provider_id 해석을 handler 에 위임. 즉 store 는 canonical(`provider_id` FK)로 이미 정리됐고 `scm_provider` 잔재는 네이밍에만 남음.

4. **`scm_provider` 컬럼의 짧은 수명 (000043 ADD → 000045 DROP)** — 컬럼이 #368(000043)에서 추가됐다가 2 마이그레이션 만에 #373(000045)에서 제거됐다. 운영 DB 에 이미 적용된 경우 000045 의 backfill(`UPDATE … SET provider_id = p.provider_id … WHERE p.provider_key = r.scm_provider`)이 데이터 보존. dead 컬럼 라이프사이클의 전형.

5. **draft→publish flow 무테스트 머지** — `CreateRepositoryDraft`/`MarkRepositoryDraftPublishRequested`(`applications.go:688`, `:744`)는 `internal/store/*_test.go` 어디에도 통합 테스트가 없다(grep 결과 구현 파일에만 존재). #368(codex)이 테스트 없이 머지된 결과로, MEMORY post-#373 directive "draft→publish flow 테스트 보강(무테스트 머지분)" 와 일치.

6. **column scan 비대칭 — `user_type` 누락** — `ListUsers`(`users_units.go:55`)는 `COALESCE(user_type,'human')` 를 scan 해 `user.Type` 을 채우지만, `GetUser`(`:138`)·`GetUserByIdPSubject`(`:209`)·`ListUnitMembers`(`:449`)는 user_type 컬럼을 SELECT 하지 않아 반환 `AppUser.Type` 이 항상 zero("") 다. 단건 조회 결과의 `Type` 필드가 목록 조회와 불일치 — human/system 구분이 단건 경로에서 소실되는 부채.

7. **`auditLogForCommand` 가 enrichment 컬럼 미조회** — command flow 의 기존 audit 재조회(`auditLogForCommand` `postgres.go:1762`)는 `source_ip/request_id/source_type/source_event_id` 를 SELECT 하지 않는다(`CreateAuditLog`/`ListAuditLogs` 는 조회). idempotency 재조회 경로에서 반환되는 `AuditLog` 의 actor enrichment 필드가 비는 미세 불일치.

8. **down migration 위험 — 데이터 파괴형** — `000025_dev_requests_assignee_nullable.down.sql:30` 은 `DELETE FROM dev_requests WHERE assignee_user_id IS NULL` (rollback 시 invalid_intake 보존 row 삭제). `000021_rbac_pmo_manager.down.sql:18` 은 `UPDATE users SET role='developer' WHERE role='pmo_manager'` (rollback 시 강제 role 재할당). `000039`/`000037` down 은 NULL row 존재 시 SET NOT NULL/CREATE UNIQUE INDEX 실패 가능. 자세한 정합은 `migrations.md` 참조.

9. **`activeLinkCounts`/`repositoryIDs` 등 fake 전용 필드 drift 위험** — `memoryApplicationStore` 가 production parity 를 주석으로만 약속하므로(map 기반), CHECK 제약·partial unique·tx 격리 수준 차이가 코드 리뷰로만 검증된다. 실 회귀는 통합 테스트(DB 필요) 에서만 잡힘 — MEMORY 의 "memoryStore production parity" 학습과 동일 부채.
