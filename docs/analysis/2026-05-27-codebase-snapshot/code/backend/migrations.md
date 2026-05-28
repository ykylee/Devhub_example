# backend-core/migrations — 스키마 마이그레이션 분석

- 문서 목적: `backend-core/migrations/` 의 000001~000045 up/down 마이그레이션을 테이블·변경 1줄 요약 + 도메인 그룹핑 + up/down 정합성 관점에서 분석한다.
- 범위: golang-migrate sequence 45쌍(`*.up.sql`/`*.down.sql`).
- 대상 독자: 스키마 변경자, 마이그레이션 작성자, 운영 배포자.
- 상태: snapshot (2026-05-27, main `cf19c94`)
- 관련 문서: `store.md`, `domain.md`

## 1. 마이그레이션 1줄 요약 표

| # | 파일 | 다루는 테이블/변경 |
|---|------|--------------------|
| 000001 | create_webhook_events | `webhook_events` 생성 (dedupe_key UNIQUE, status CHECK 5종) |
| 000002 | create_domain_tables | `gitea_users`/`repositories`/`issues`/`pull_requests`/`ci_runs`/`risks` SCM mirror 6 테이블 생성 |
| 000003 | create_command_audit_tables | `commands`(status 6종 CHECK)/`audit_logs` 생성 |
| 000004 | create_users_units | `org_units`/`users`(role/status CHECK)/`unit_appointments` 생성 + seed(7 unit/3 user/4 appointment) |
| 000005 | create_rbac_policies | `rbac_policies` 생성 (role_id format + audit_invariant CHECK) + system role 3종 seed |
| 000006 | users_role_fkey | `users.role` CHECK → `rbac_policies(role_id)` FK 로 교체 (ADR-0002) |
| 000007 | add_user_type_to_users | `users.user_type` 컬럼 추가 (default 'human') |
| 000008 | audit_logs_actor_enrichment | `audit_logs` 에 source_ip/request_id/source_type 컬럼 + partial index |
| 000009 | add_kratos_identity_id_to_users | `users.kratos_identity_id` 컬럼 + partial UNIQUE index |
| 000010 | create_hrdb_persons | `hrdb` schema + `hrdb.persons` 테이블 (ADR-0008) |
| 000011 | create_org_units_total_count_mv | `org_units_total_count` Materialized View (WITH NO DATA + REFRESH) |
| 000012 | create_scm_providers | `scm_providers` catalog 생성 + 4 provider seed |
| 000013 | create_applications | `applications` 생성 (pgcrypto, key/status/visibility/archived/due_date CHECK) |
| 000014 | create_application_repositories | `application_repositories` composite-PK link (role/sync_status/sync_error CHECK) |
| 000015 | create_projects | `projects` 생성 (repository_id NOT NULL RESTRICT, (repository_id,key) UNIQUE) |
| 000016 | create_project_members_and_integrations | `project_members`/`project_integrations`(scope polymorphism, partial UNIQUE 2종) |
| 000017 | create_repo_ops_snapshots | `pr_activities`/`build_runs`/`quality_snapshots` 운영 지표 3 테이블 |
| 000018 | rbac_add_application_resources | rbac_policies.permissions 에 applications/application_repositories/projects/scm_providers 4 resource 추가 |
| 000019 | unit_single_leader | `unit_appointments` 에 partial UNIQUE index (unit 당 leader 1명) |
| 000020 | application_leader_department | `applications.leader_user_id`/`development_unit_id` 컬럼 + owner backfill |
| 000021 | rbac_pmo_manager | `pmo_manager` system role + role_id_format CHECK 확장 |
| 000022 | dev_requests | `dev_requests` 생성 (status/target/registered/rejected CHECK + idempotency partial UNIQUE) |
| 000023 | dev_request_intake_tokens | `dev_request_intake_tokens` 생성 (hashed_token UNIQUE, allowed_ips array CHECK) |
| 000024 | rbac_dev_requests_resource | rbac_policies.permissions 에 dev_requests resource 추가 |
| 000025 | dev_requests_assignee_nullable | `dev_requests.assignee_user_id` NOT NULL 해제 + active-status CHECK |
| 000026 | rbac_dev_request_intake_tokens | rbac_policies.permissions 에 dev_request_intake_tokens resource 추가 |
| 000027 | add_expires_at_to_intake_tokens | `dev_request_intake_tokens` 에 expires_at/updated_at 컬럼 |
| 000028 | integration_registry | `integration_providers`/`integration_bindings`/`integration_sync_jobs` 3 테이블 |
| 000029 | infra_service_snapshots | `infra_service_snapshots` 생성 (ingest_id PK, jsonb payload) |
| 000030 | rename_kratos_identity_to_idp_subject | `users.kratos_identity_id` → `idp_subject` rename + index 재생성 (ADR-0019) |
| 000031 | create_event_cursors | `event_cursors` 생성 (Keycloak event polling 커서) |
| 000032 | audit_logs_source_event_id | `audit_logs.source_event_id` 컬럼 + partial UNIQUE (source_type,source_event_id) dedup |
| 000033 | user_onboarding_state | `users.onboarding_completed_at`/`review_status` 컬럼 + bi-implication CHECK + partial index |
| 000034 | project_repositories | `project_repositories` N:M 조인 테이블 + legacy repository_id backfill |
| 000035 | create_realtime_tickets | `realtime_tickets` 생성 (ticket PK, single-use, ADR-0024 §6) |
| 000036 | relax_applications_key_format | `applications_key_format` CHECK 10자 고정 → 1~10자 완화 |
| 000037 | projects_repository_nullable | `projects.repository_id` NOT NULL 해제 (standalone project) |
| 000038 | integration_providers_base_url | `integration_providers.base_url` 컬럼 |
| 000039 | projects_standalone_key_unique | standalone(repository_id NULL) project key partial UNIQUE + preflight dedup |
| 000040 | integration_providers_api_token | `integration_providers.api_token` 컬럼 (outbound PAT) |
| 000041 | integration_providers_auth_credentials | `integration_providers` 에 auth_username/auth_client_id/auth_token_url/auth_secret 4 컬럼 |
| 000042 | repositories_source_provider | `repositories.source`/`provider_id`(FK)/`description` 컬럼 + provider_id index (소유권 분리) |
| 000043 | repositories_draft_status | `repositories.repository_status`(draft/active CHECK)/`scm_provider`/`publish_requested_at`/`published_at` 컬럼 |
| 000044 | projects_key_unique_active_only | projects key UNIQUE 범위를 active row(archived_at IS NULL) 로 제한 (구 000042, prefix 충돌로 재번호) |
| 000045 | repositories_drop_scm_provider | `repositories.scm_provider` 제거 + provider_id 로 backfill (provider_id 단일화) |

## 2. 도메인 그룹핑

- **SCM mirror / webhook 수집** (M0~M1): 000001(webhook_events), 000002(repositories·issues·PR·CI·risks·gitea_users), 000017(pr_activities·build_runs·quality_snapshots), 000042/000043/000045(repositories 소유권 분리 + draft/publish).
- **Command / Audit** (운영 명령): 000003(commands·audit_logs), 000008(audit enrichment), 000032(audit source_event_id dedup).
- **사용자 / 조직 / 신원** (Identity): 000004(users·org_units·unit_appointments + seed), 000007(user_type), 000009→000030(kratos_identity_id→idp_subject), 000011(total_count MV), 000019(single-leader), 000033(onboarding state), 000010(hrdb.persons).
- **RBAC**: 000005(rbac_policies + 3 role seed), 000006(role FK), 000018(application resources), 000021(pmo_manager), 000024(dev_requests resource), 000026(intake_tokens resource).
- **Application / Project / SCM catalog** (거버넌스): 000012(scm_providers), 000013(applications), 000014(application_repositories), 000015→000037(projects + nullable), 000016(project_members·integrations), 000020(application leader/dept), 000034(project_repositories N:M), 000036(key format), 000039/000044(project key UNIQUE 정책).
- **DREQ** (Dev Request): 000022(dev_requests), 000023→000027(intake_tokens + expires_at), 000025(assignee nullable).
- **External Integration registry**: 000028(providers·bindings·sync_jobs), 000038(base_url), 000040(api_token), 000041(auth credentials).
- **Infra / Realtime / Keycloak event**: 000029(infra_service_snapshots), 000031(event_cursors), 000035(realtime_tickets).

## 3. up/down 정합성 평가

대부분 down 은 표준 역연산(DROP TABLE/INDEX/COLUMN, ADD/DROP CONSTRAINT)으로 깔끔하다. 다음은 비대칭·데이터 파괴·재생성 패턴.

| # | up/down 정합성 메모 |
|---|---------------------|
| 000004 | up 이 seed INSERT 포함하나 down 은 DROP TABLE 로 일괄 제거 — 정합 (seed 별도 삭제 불요). |
| 000005 | down 이 system role seed 를 별도 DELETE 하지 않고 DROP TABLE — 정합. |
| 000006 | down 주석이 명시: 재생성 CHECK 가 custom role 보유 row 에서 위반 → rollback 전 system role 재할당 필요(운영 수동). |
| 000010 | down 이 `DROP SCHEMA hrdb` — schema 에 다른 객체 있으면 실패 가능(현 시점 persons 만). |
| 000011 | down 이 MV+index DROP — 정합. 단 up 의 `REFRESH` 는 cron 의존(운영 메모). |
| 000018/000024/000026 | RBAC resource 추가는 `permissions \|\| jsonb_build_object` UPDATE, down 은 `permissions - 'key'` — 정합한 JSONB 가감. |
| 000020/000033/000041/000042/000043 | 컬럼 ADD ↔ down DROP COLUMN IF EXISTS — 정합. |
| **000021** | **down 이 데이터 변형**: `UPDATE users SET role='developer' WHERE role='pmo_manager'` (FK 위반 회피용 강제 재할당, codex PR #119 P1). rollback 후 원 role 복원 불가. |
| **000025** | **down 이 데이터 삭제**: `DELETE FROM dev_requests WHERE assignee_user_id IS NULL` (NULL assignee row 가 재생성 NOT NULL 과 양립 불가). invalid_intake 보존 row 손실. |
| 000030 | up/down 모두 `DO $$ … information_schema 체크 … RENAME` 가드 + index 재생성 — 양방향 idempotent, 정합. |
| **000037** | down `ALTER COLUMN repository_id SET NOT NULL` 은 NULL row 존재 시 **실패**(주석 명시). standalone project 생성 후 rollback 불가. |
| **000039** | up 이 preflight dedup(`key \|\| '-dup-' \|\| id`)으로 중복 standalone key 를 재명명 — codex #354 P2. down 은 index DROP 만 (dedup suffix row 는 그대로 남음, 운영자 reconcile). |
| 000044 | up 이 `projects_repository_key_unique` 제약 + `projects_standalone_key_uniq`(000039) DROP 후 active-only partial UNIQUE 2종 생성. down 은 역으로 복원 — 정합. 단 down 복원 시 archived 포함 충돌 정책으로 되돌아감. |
| **000045** | up 이 scm_provider→provider_id backfill 후 `DROP COLUMN scm_provider`. down 은 컬럼 재추가 + provider_id→provider_key best-effort backfill — 비대칭(원 scm_provider 텍스트값이 integration_providers 에 매칭 안 되면 NULL 로 복원). |

전반 평가: golang-migrate 의 forward-only 운영을 전제로 작성됐고, **000021/000025/000037/000039/000045 의 down 은 데이터 손실/재생성 위험**이 있어 운영 rollback 시 사전 점검 필수(각 파일 down 주석에 경고 명시).

## 발견 사항 (불일치/stale/부채)

1. **migration prefix 충돌 이력 (000042 #363↔#368 → #371 000044 재번호)** — `000044_projects_key_unique_active_only.up.sql:1-3` 주석: "원래 000042 로 작성됐으나 #363 의 000042_repositories_source_provider 와 prefix 충돌 → 000044 로 재번호. projects 전용이라 repositories migration 과 순서 무관." 두 PR(#363 repositories, #368 projects-key)이 동시에 000042 를 선점했고, CI bypass 머지로 lint 가 못 잡아 점검 sprint(#371)가 나중 머지분을 000044 로 재번호. 현재 sequence 에 000042 는 단 1개(repositories_source_provider)로 정상.

2. **평문 secret 컬럼 (credentials_ref / api_token / auth_secret)** — 000028 의 `credentials_ref TEXT NOT NULL`, 000040 의 `api_token TEXT`, 000041 의 `auth_secret TEXT` 가 모두 평문 컬럼. 000041 주석은 "auth_secret is treated write-only at the API layer" 라고만 명시 — DB at-rest 암호화/KMS 없음. `credentials_ref`(inbound webhook 시크릿)·`api_token`(outbound PAT)·`auth_secret`(basic/oauth2 비밀)이 평문 저장 (#6 평문 secret 보안 carve).

3. **scm_provider 컬럼의 짧은 dead 라이프사이클 (000043 ADD → 000045 DROP)** — `scm_provider TEXT` 가 #368(000043)에서 추가됐다가 #373(000045)에서 제거. #368 의 `scm_provider`(provider_key TEXT) 와 #363 의 `provider_id`(FK UUID, 000042)가 동일 SCM 참조를 의미 중복(post-#371 soft note) → provider_id(FK)를 canonical 로 단일화. 000045 가 backfill(provider_key→provider_id) 후 DROP. 운영 DB 에 000043~000045 사이가 적용됐다면 backfill 로 데이터 보존되나, 짧게 존재한 dead 컬럼 흔적.

4. **`scm_providers`(000012) ↔ `integration_providers`(000028) 의 SCM 도메인 이중화** — 별개 두 catalog 가 공존: `scm_providers`(provider_key PK, application_repositories.repo_provider FK 대상)와 `integration_providers`(provider_id UUID PK, scm 포함 5종 type). repositories.provider_id(000042)는 `integration_providers` 를 FK 참조 — 즉 SCM provider 정보가 두 테이블에 분산. 어느 쪽이 canonical SCM catalog 인지 모호한 도메인 부채.

5. **MV 갱신/seed 의 런타임 의존** — 000011 `org_units_total_count` MV 는 `WITH NO DATA` + 1회 `REFRESH` 후 갱신을 cron 에 위임(주석). users_units.go 의 `GetHierarchy` 는 MV 를 안 쓰고 RECURSIVE CTE 로 매번 재계산하므로(store.md 참조) **MV 가 실 read 경로에서 사용되지 않는 dead 자산** 가능성. 000004 seed 의 user u1/u2/u3 + org-root 등은 dev/test fixture 라 prod 에서 의도치 않게 잔존할 위험.

6. **idempotency partial UNIQUE 의 NULL distinct 함정 패턴 반복** — 000022(dev_requests `WHERE external_ref IS NOT NULL`), 000032(audit `WHERE source_event_id IS NOT NULL AND source_type IS NOT NULL`), 000009(`WHERE kratos_identity_id IS NOT NULL`), 000039(`WHERE repository_id IS NULL`)가 모두 PG NULL distinct 규칙 회피용 partial UNIQUE. 000039 주석이 명시하듯 000037(repository_id nullable)이 기존 `UNIQUE(repository_id,key)` 의 NULL row dedup 무력화를 유발 → partial index 보강. 동일 함정이 여러 곳에서 반복돼 신규 nullable 컬럼 추가 시 주의 요함.

7. **000030 주석 헤더 번호 오기** — `000030_rename_kratos_identity_to_idp_subject.up.sql:1` 의 헤더 주석이 "-- 000021: Rename …" 로 잘못된 번호(실제 000030). down 도 동일("-- 000021 rollback"). 기능 무해하나 stale 주석. 마찬가지로 down 파일들의 헤더 번호는 대체로 정확하나 이 한 건이 copy-paste 오기.

8. **CHECK 어휘 ↔ 도메인 enum 수동 동기화 부채** — 000003 commands status(6종), 000014 sync_error_code(8종), 000028 provider_type(5종)/auth_mode(5종) 등 모든 enum CHECK 가 `domain/*.go` 상수와 손으로 맞춰진다. 한쪽만 늘리면 INSERT CHECK 위반 또는 도메인 default 분기로 갈림 (domain.md 발견 4/6 과 동일). golang-migrate 가 forward-only 라 enum 추가는 항상 새 ALTER 마이그레이션 필요.
