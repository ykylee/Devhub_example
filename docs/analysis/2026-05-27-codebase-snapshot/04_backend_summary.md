# 04. 백엔드 개발 사항 정리

- 문서 목적: 백엔드(backend-core Go + backend-ai Python) 개발 현황을 도메인/패키지별로 정리한다.
- 기준: main `cf19c94` / `backend-core/` + `backend-ai/`.
- 스택: Go + Gin(HTTP) + pgx(PostgreSQL) + JWT(Keycloak JWKS). Python + FastAPI(스켈레톤).

---

## 1. 영역별 완성 현황

| 영역 | 상태 | 비고 |
| --- | :-: | --- |
| Keycloak OIDC 토큰 검증 (JWKS + stale-while-error) | ✅ done | `internal/auth/keycloak_verifier.go` + MaxStaleDuration fallback. |
| RBAC enforcement (4-boolean matrix + 라우트 매핑 + deny-by-default) | ✅ done | `permissions.go` routePermissionTable + PermissionCache. |
| 계정/사용자/조직 CRUD + 계층 + 임명 | ✅ done | `organization.go` + `users_units.go`. single-leader invariant(SQL). |
| Onboarding (gate middleware + submit/search/review + migration) | ✅ done | `me_onboarding.go`·`onboarding_gate.go`·`organizations_search.go`·`users_admin_review.go`. lazy_auto_create 폐기. |
| Audit (Keycloak event polling → audit_logs + user sync) | ✅ done | `internal/audit/*` + event_cursors dedup + Prometheus. |
| Application/Repository/Project (CRUD + 상태전이 + rollup + row-scoping) | ✅ done | `applications.go`·`projects.go`·`repository_ops.go`·`application_rollup.go`. |
| Repository draft→publish lifecycle | ✅ done(무테스트) | `repository_ops.go` createRepositoryDraft/requestRepositoryPublish (#368, UT 미흡). |
| SCM↔시스템 repository 양방향 (소유권 분리 + import + create) | ✅ done | `integration_scm_repositories.go` (API-88/89/90) + UpsertRepository ON CONFLICT. |
| DREQ (intake auth + promote-tx + token admin + cron) | ✅ done | `dev_requests.go`·`dev_request_intake_*.go` + `internal/devrequest/`. |
| External Integration (provider/binding registry + auth_mode + webhook ingest) | ✅ done | `integration_registry.go`·`integrations.go`·`infra_integrations.go`. |
| HomeLab pull adapter (file + HTTP + health policy + metrics) | ✅ done | `internal/integrations/adapters/*`. |
| Gitea SCM 동기화 워커 (pull, sync job 큐) | ✅ done | `internal/gitea/{client,syncer,worker}.go` + integration_sync_jobs(SKIP LOCKED). |
| Realtime WS (ticket auth + command publish) | 🟡 부분 | ticket store PG/in-memory 완성. event publish 는 command 만(infra/ci/risk 미완 = RM-M4-01). |
| Command/Service Action (dry-run + live executor + approval) | ✅ done | `commandworker/*` + `serviceaction/*`. |
| Python AI gRPC (AnalysisService) | 🔴 미구현 | `backend-ai/main.py` 스켈레톤(`/health` 만). |

## 2. 패키지별 정리 (backend-core, 89 구현 + 72 테스트)

| 패키지 | 책임 | 핵심 파일 |
| --- | --- | --- |
| `httpapi` (37 test) | 라우터·핸들러·미들웨어 | router.go·permissions.go·auth.go·request_context.go·realtime*.go + 42 핸들러 |
| `store` (11) | PostgreSQL 영속화 | applications/integrations/dev_requests/users_units/audit_logs/postgres_rbac/realtime_tickets/infra_snapshots/event_cursors |
| `domain` (3) | 모델·enum·상태머신 | domain.go·application.go·dev_request.go·rbac.go |
| `gitea` (6) | webhook(push) + pull sync | signature.go(HMAC) + client/syncer/worker |
| `audit` (4) | Keycloak event → audit | keycloak_event_puller·user_sync·metrics |
| `auth` (1) | JWKS 검증 + stale fallback | keycloak_verifier.go |
| `commandworker` (2) | command 실행 워커 | dry-run + live executor |
| `serviceaction` (1) | service action executor | mock/compose/k8s |
| `integrations/adapters` (3) | HomeLab pull | contract·homelab·puller |
| `config`/`devrequest`/`normalize`/`hrdb` (각 1) | 설정·DREQ cron·정규화·HR 조회 | |

## 3. 데이터 계층 (45 migration)

§01 §2.5 참조. 핵심 신규(2026-05 후반): 000033(onboarding) · 000035(realtime_tickets) · 000038(base_url) · 000040(api_token) · 000041(auth_credentials) · 000042(repository source/provider_id) · 000043(draft status) · 000044(projects key active-only) · 000045(scm_provider DROP → provider_id 단일화).

## 4. 보안/운영 baseline

- **인증**: Keycloak 단일 IdP. `SetTrustedProxies(nil)` 로 X-Forwarded-For 위조 차단. JWKS cache TTL 5분 + stale-while-error(default 24h).
- **권한**: deny-by-default + row-level scoping(ADR-0011, enforceRowOwnership).
- **audit**: SourceType 분류 + SourceEventID partial UNIQUE dedup.
- **secret 처리**: 신규 secret(api_token/auth_secret)은 **write-only 응답 패턴**(`<field>_set` bool, raw 비노출). 단 **저장은 평문** — `credentials_ref`/`api_token`/`auth_secret` envelope 암호화는 미해소 carve(#6).
- **outbound 자격증명**: 명시 provider 는 env 토큰 fallback 금지(#359 — 잘못된 계정 유출 방지).
- **관측**: Prometheus metric(JWKS stale / intake token expiring·stale·revoked / onboarding pending / sync). Alertmanager + Grafana JSON.

## 5. 최근 주요 변화 (2026-05-26~27)

1. **Gitea SCM 동기화 워커** (#341) — REST pull(repos/issues/PRs) 정규화 upsert + sync job 큐.
2. **provider api_token 슬롯** (#355, 000040) — write-only.
3. **외부 연동 등록 깊이** (#352) — base_url(000038) + 연결 테스트(API-87) + vendor preset.
4. **webhook 헤더 alias + auth_mode full 모델** (#358, 000041) — X-Gitea/X-Gogs fallback + OutboundAuth(token/basic/app_password/oauth2/agent).
5. **env token 유출 방지 hotfix** (#359) — 명시 provider 의 자격증명만 사용.
6. **SCM↔시스템 repository 양방향** (#363 A+B / #366 Phase C / #373 provider_id 단일화) — 소유권 분리(source=scm|system) + import(API-89) + create(API-90, gitea CreateRepo) + capability gate.
7. **repository draft→publish lifecycle** (#368, 000043) — draft 생성 + publish 요청.
8. **전수 점검 hotfix** (#371) — migration 000042 prefix 충돌 정정 + disabled provider 거부(409).

## 6. 백엔드 부채

- **#368 draft→publish handler 무테스트 머지** — UT/통합테스트 보강 필요.
- **평문 secret 저장** — envelope 암호화(KMS/DEK) carve(#6) 미해소.
- **마이그레이션 prefix 충돌 위험** — CI bypass 머지가 동시 번호 충돌을 통과시킨 이력(000042). prefix uniqueness guard 강화 필요.
- **Realtime event publish 미완** — RM-M4-01 infra/ci/risk publish + RM-M4-02 replay/리소스 필터.
- **backend-ai 미구현** — gRPC AnalysisService(v2).
- **inbound webhook 정규화 깊이** — multi-provider sync 일반화 여지.
