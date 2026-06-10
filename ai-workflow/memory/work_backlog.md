# Integrated Work Backlog (main, post M1 RBAC)

- 문서 목적: main 브랜치 기준 상위 백로그 인덱스. 세부 sprint backlog 는 브랜치별 메모리 디렉터리 참조.
- 범위: 마일스톤 상태, 최근 머지, 잔여/후속 작업
- 대상 독자: 프로젝트 리드, 후속 에이전트, 트랙 담당자
- 상태: **2026-06-10 update — main HEAD `ad8d481` (PR #505 swagger UI 1차 bootstrap) + housekeeping (N-8/N-11 row close, ADR-0025/0026 row 보강, sprint -k 마킹).** v1.0 출시 직전, 자유 에이전트 정책. 다음 sprint: N-10 Manager RBAC E2E 6 TC (`maintenance/work_260610-c-N10-rbac-e2e-tcs`).
- 최종 수정일: 2026-06-10
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [세션 인계](./session_handoff.md), [상태 스냅샷](./state.json), [M1 PR 리뷰 actions](./M1-PR-review-actions.md), [ADR-0025](../../docs/adr/0025-envelope-encryption-key-management.md)

## 1. 마일스톤 진행 상황

| 마일스톤 | 상태 | 종료 일자 | 메모 |
| --- | --- | --- | --- |
| **M0** — 보안 게이트 통과 | ✅ done | 2026-05-08 | PR #14·15·16·17·18·19. SEC-1~4 resolved. T-M0-10 운영 검증 PASS. |
| **M1** — 핵심 기능 contract 정합성 | ✅ done | 2026-05-11 | PR-B+C (#56) + PR-D (#57) 머지로 envelope / types / WS / CommandStatus / audit actor enrichment / request_id 완결. PR-D 후속 (#80/#82) 으로 commands enrichment + DEVHUB_TRUSTED_PROXIES 보강. |
| **M2** — 사용자 경험 정합 | ✅ done (1차 완성) | 2026-05-12 | login_action + work_26_05_11 + work_26_05_11-d 완료 후 PR #85 (`claude/login_usermanagement_finish`) 가 1차 완성 sprint 로 닫음: 로드맵 정합 + UX hygiene (PR-UX1+2+3) + Kratos audit (PR-M2-AUDIT) + 30 TC e2e 게이트. |
| **CI** — GitHub Actions 도입 | ✅ done (1차) | 2026-05-13 | PR #86 (`gemini/prepare-github-action`) 가 backend-unit + frontend-unit + e2e (Playwright 40 TC) 3잡 도입. 리뷰어 모드 2-pass 에서 5 blocker + follow-on 5 발견 → 보강 commit 7개로 그린 도달. PR-T5 의 핵심 잡 묶음은 본 PR 으로 1차 완료. 후속: FU-CI-1 (no-docker policy 정합), FU-CI-2/3/4 는 `claude/work_260513-a` 처리 중. |
| **M3** — Realtime 확장 + 외부 연동 1차 | ✅ done (1차) | 2026-05-13 | RM-M3-01..03 (Sign Up + 인사 DB + 조직 polish) 완료. ADR-0008/0009/0010 신규. |
| **Application Domain (backend 1차)** | ✅ done | 2026-05-14 | API-01~58 전체 activated (본 세션 #104~#110). 마이그레이션 000012~000018 (7), ADR-0011 accepted, RBAC 4 신규 resource (system_admin 일임), CI 5 job (Backend Integration 신설). 23 integration test (P1/P2 회귀 guard 포함) 가 CI 에서 실 실행. |
| **DREQ Domain (closing 1차 + TC-DREQ-* 13건)** | ✅ closing | 2026-05-18 | API-59..68 + API-79 (PATCH allowed_ips) activated + ADR-0012/0013/0014/**0017**. P2 carve out 6/6 모두 해소 + codex hotfix #5 (store revoked guard + 회귀 가드 3 unit test). **TC-DREQ-* 13건 정식 발급** (sprint d) + `dev-requests.spec.ts` 6 step + `test_cases_m5_dreq.md` 신규. 잔여 carve out: ADR-0017 §6 atomicity 실제 구현 (CTE refactor) + 자동 cron revoke + last_used staleness alert. |
| **External Integration Domain (1차 종합 closing)** | ✅ closing | 2026-05-18 | concept staged (PR #135) → backend 1차 (PR #139) → **ADR-0015/0016/0017** (sprint c) → provider frontend (sprint g/h) + API-80 DELETE (sprint j) → **bindings 관리 UI** (sprint m, PR #154) + **topology v2 시각화** (sprint n, PR #155) + **ADR-0017 §6 atomicity** 실 구현 (sprint o, PR #156) + **ADR-0015 §6 (1)+(2)** (sprint p, PR #157). API-69..80 모두 activated. TC-INT-FRONTEND-* 10건 + TC-INT-FRONTEND-BIND-* 3건 + TC-INT-HOMELAB-03 + TC-INT-FRONTEND-TOPOLOGY-V2-* 2건. ADR carve out 추가 종결 (atomicity + size limit + token rotation SOP). 잔여 carve: ADR-0016 §6 (Alertmanager YAML / Grafana JSON / p95 alert / push 알림 / 임계), ADR-0017 §6 잔여 (cron revoke / PATCH expires_at / 만료 alert), ADR-0015 §6 (3)+(4) (dedicated worker / push-pull dedup), React Flow group sub-node + WebSocket 실시간. |
| **M4** — 운영 / SSO / MFA / 후속 ADR | planned | — | 통합 로드맵 §3.5. RM-M4-01..09 (9 항목). |

## 2. 최근 머지 (M1 RBAC track, 2026-05-08)

| PR | 제목 | 머지 |
| --- | --- | --- |
| #20 | feat(http): SEC-5 mask 5xx errors + M1 sprint plan (PR-A) | `ae8aca1` |
| #21 | docs(adr): ADR-0002 RBAC policy edit API (PR-F) | `950a11f` |
| #22 | docs(api): RBAC §12 rewrite + route mapping (PR-G1) | `1a090a3` |
| #23 | feat(rbac): domain + rbac_policies migration (PR-G2) | `5239a87` |
| #29 | feat(rbac): postgres store + users.role FK (PR-G3 + FIX-A) | `24815b8` |
| #30 | feat(rbac): RBAC handlers (PR-G4 + FIX-B + FIX-C) | `27b6817` |
| #31 | feat(rbac): permission cache + enforcement (PR-G5) | `02eef35` |
| #27 | feat(frontend): RBAC PermissionEditor ↔ backend (PR-G6 + FIX-D) | `e02ba67` |
| #28 | docs(memory): M1 PR review actions tracker | `9bc30c9` |

원본 PR #24, #25, #26 은 stack base 자동 삭제로 close 후 main 위에서 #29/#30/#31 로 재등록.

## 3. M1 잔여 + DEFER 후속 작업

### 3.1 M1 잔여 (P1, 진입 시 분해)

- **T-M1-02 (PR-B)** — API envelope/role wire/필드 정합 (`backend_api_contract.md` ↔ 코드 1:1 강제, 통합 테스트 매트릭스).
- **T-M1-03 (PR-C)** — command lifecycle 6 상태 일관 적용 + dry-run/live 경계 테스트.
- **T-M1-04 (PR-D)** — Audit actor 보강 (`source_ip`, `request_id`, `source_type` + 마이그레이션 + 미들웨어 + 응답 헤더 `X-Request-ID`).
- **T-M1-05** — `auth_test.go` prod 가드 + role 가드 통합 테스트 매트릭스. PR-C 또는 PR-D 에 흡수 가능.
- **T-M1-07 (frontend)** — `frontend/lib/services/types.ts` UI vs wire 분리, 표시 포맷 프론트 이전. PR-B 에 묶거나 단독.
- **T-M1-08** — WebSocket envelope `{schema_version, type, event_id, occurred_at, data}` 코드/문서 정합. PR-B 에 묶거나 단독.

### 3.2 DEFER (M1 PR 리뷰 후속, [상세](./M1-PR-review-actions.md#3-다음-개발로-넘김--defer))

- **M1-DEFER-A** — `rbac_policies` 의 `is_system` ↔ `role_id` 일관성 CHECK 제약 (P2 방어선)
- **M1-DEFER-B** — `requireMinRole`/`roleMeetsMin`/`roleRank` deadcode 정리 + 단위 테스트 제거
- **M1-DEFER-C** — `writeRBACServerError` 임시 helper → `writeServerError` 통합 (PR-G4 의 TODO)
- **M1-DEFER-D** — `DeleteRBACRole` row-lock (다중 인스턴스 race 강화)
- **M1-DEFER-E** — `PermissionCache` 다중 인스턴스 일관성 (pub/sub 또는 polling)
- **M1-DEFER-F** — API contract §12.4 / §12.5 응답 예시 추가
- **M1-DEFER-G** — `MemberTable` role display 회귀 사용자 환경 검증

### 3.2 P1~P2 should-have

- **frontend `/auth/callback`**: Hydra authorization_code → `/oauth2/token` 교환 후 세션 저장 흐름. PR-D 의 자연 후속.
- **frontend `account.service.ts`**: Kratos public flow 호출 helper.
- **frontend `types.ts` 분리**: UI 표시명 vs API wire 타입 분리, 표시 포맷팅을 프론트로 이전.
- **WebSocket envelope 표준화**: `{schema_version, type, event_id, occurred_at, data}` 코드/문서 정합.
- **RBAC policy 편집 API**: 12.x — write/audit 경계 + persistence 또는 *static-default 유지* 결정.

### 3.3 P3 nice-to-have

- ADR-0002 (Gitea SSO 통합) 작성.
- ADR (commits 정규화 테이블 도입 여부).
- ADR (OS service wrapper 결정).

## 4. 운영 / 환경 메모

- Hydra/Kratos PoC: 본 세션에서 e2e 검증 완료. binary 위치 `$env:USERPROFILE\go\bin\` (Windows). 다음 세션에서 재가동은 `hydra serve all --config infra/idp/hydra.yaml --dev` + `kratos serve --config infra/idp/kratos.yaml`.
- 검증용 임시 OIDC client (client_credentials, id `43aa4b74-...`) 는 Hydra 안에 잔존. 보안 위험은 없으나 cleanup 가능.
- backend-core `go test ./...` 는 사내 GoProxy mirror 환경에서 PASS 검증됨.

## 5. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-09 | **N-8 sign-out e2e deterministic race hotfix 4차 (issue #501)** — 3 commit / 2 PR (PR #502 + #503, 4 신규 TC). **PR #502** (`fix/work_260609-a-N8-logout-graceful-degrade`): backend logout handler 502 → 204 graceful degradation + audit `revoke_status=unreachable` + hotfix 식별자. **PR #503** (`fix/work_260609-b-N8-codex-p1-distinguishable`, 2 commit): (1) codex P1 응답 — response header `X-Keycloak-Likely-Down: true` + `X-Logout-Hotfix` marker; (2) codex P1 follow-up 응답 — typed error sentinel `authview.ErrOIDCConfigMissing` (config error, marker 미부착) + `authview.ErrOIDCNetworkUnreachable` (네트워크/5xx, marker 부착). **검증**: e2e shard 1..3 PASS (N-11 잔여 DoD 해소), backend 35 packages PASS, frontend 80 files / 1033 tests PASS (FE-08 신규 +1). **신규 ID 없음** (N-8 follow-up hotfix). main HEAD `897953c`. issue #501 close. |
| 2026-06-09 | **워커 분업 전면 취소 (사용자 결정, PR #500)** — Claude/Codex 자유 이용 불가. `worker_division.md` §0 + §1~§4 historical 標記 + §2.5 branch prefix 자유화 + §5 Owner 권한 명시. `AGENTS.md` 워커 일반 메모 갱신 + 워커별 전용 메모 4종 historical 標記. 유지 정책: §4.2 ADR reversal supersession 정공법 + §5 Owner 권한 + 우선순위 P0~P3 강제. main `f99fef7`. |
| 2026-06-09 | **N-11 메모리 sync (PR #499)** — 5 파일 (+36/-15) main `da7d57e` 머지. 충돌 해소 (release_v1_roadmap.md N-11 row + 워커 분업 row 둘 다 보존). release_v1_roadmap.md 의 PR #500 row 는 §9 변경 이력에 별도 정합. 신규 ID 없음 (cross-cutting infrastructure 운영 정합). |
| 2026-06-08 | **N-11 CI e2e + backend-integration 복원 운영 정합 (sprint 260608-a, issue #419)** — `&& false` 2건은 PR #407 cleanup-recovery 후속 4 squash merge (4a1942e / 5f5fdba / 9395cd9 / ce8ce7c) 로 코드 레벨 복원 완료. 본 sprint 의 1차 PR #498 은 ci.yml 코멘트만 갱신 (코드 변경 0줄), 2차 PR (메모리 4종 + traceability report.md §6 + release_v1_roadmap §3.5/§4.1/§9) 정합. **잔여 DoD**: main 첫 PR 에서 두 job 실 실행 PASS (state.json head + §3 잔여 표 N-11 잔여 row 추가). 신규 ID 없음 (cross-cutting infrastructure 운영 정합). |
| 2026-06-06 | **sprint -h 신규 carve 의 ID 발급 + 매트릭스 cross-ref** — §3.1 의 auth-session / integration-registry / repository-integration row 3 row cross-ref 갱신. **신규 ID**: REQ-FR-106/107/108 + ARCH-18/19/20 + API-98/99/100 + IMPL/UT/TC 관련 ID 발급. Codex 리뷰에 따라 `docs/traceability/report.md`의 `integration-registry` 도메인에 `IMPL-ci-runs-01`, `UT-ci-runs-01`, `TC-CI-RUN-01` 추가 보완 완료. PR #490 머지 완료. |
| 2026-06-02 | 중간 개발 보고 자료 준비 착수. `docs/presentations/2026-06-02-midterm-report-plan.md` 에 슬라이드 구조/데이터 수집 축/디자인 방향을 정리했고, `docs/analysis/2026-06-02-midterm-report-baseline.md` 에 현재 기능 범위, SDLC/추적성, 테스트, AI agent 활용 현황, 활동 통계를 베이스라인으로 기록. 다음 단계는 HTML/CSS/JS 슬라이드 초안 구현. |
| 2026-06-01 | CI 회귀 복구: (1) `frontend/app/(dashboard)/applications/[id]/page.tsx` 중복 import 제거로 `Build App` 타입 에러 해소, (2) `frontend/tests/e2e/admin-projects.spec.ts` TC-PROJ-UI-04를 환경 독립 검증(ComboBox/input 공용)으로 보강. 로컬 `npm run test`/`npm run build` 통과, CI run `26738464130` 성공. |
| 2026-06-10 | **v1.0 출시 직전 housekeeping + 메모리 비대한 문서 분리/간소화** (sprint `maintenance/work_260610-b-v1-pre-release-housekeeping`) — (1) `release_v1_roadmap.md` §3.5 N-8 race row status `✅ resolved` + N-11 row "잔여 DoD" → "✅ 해소" 명시 + §4.1 sprint -k 의 N-11 잔여 DoD 완료 마킹 + §3.5 N-10 follow-up 6 TC 명시 + §9 변경 이력 1 row. (2) `docs/traceability/report.md` §4 ADR-0025 (envelope encryption) + ADR-0026 (Keycloak role excluded) row 보강 + §6 변경 이력 1 row. (3) **메모리 비대한 문서 분리**: `ai-workflow/memory/state.json` 1515 → 150 line (90% 감소). 21 top-level key (merged_prs_2026_05_11~2026_05_20 + 5 progress 객체 + 1 milestone 객체) + next_actions 5 closed key → `_archive/state-2026-05-pr-tracker/` (state.json 85KB + next_actions_closed.json 3.5KB + work_backlog_2026-05_archive.md 80KB). 본 work_backlog.md §5 의 5월 135행 archive summary 1 line 으로 축약. 5월 136건 변경 이력의 canonical source 는 `git log --since=2026-05-01 --until=2026-06-01 --merges`. (4) PR 머지 후 main flat memory finalize. **신규 ID 발급 없음** (governance housekeeping). main HEAD `ad8d481` (PR #505). |
| 2026-06-10 | **v1.1 sprint -a follow-up PR #539** (sprint `feat/work_260610-v1-1-sprint-a-followup`) — sprint -a 본 PR #538 (port interface) 의 후속. (1) `backend-core/internal/sso-integrations/keycloak/saovae_stub.go` (NEW, 105 lines) — 4 port stub (BearerTokenVerifier + IdentityAdmin + OIDCLogoutClient + KeycloakEventPort) + webhook handler. 사외 build 시 default wiring, Keycloak 인프라 의존성 0. (2) `backend-core/main.go` — `DEVHUB_BUILD_TIER` env var 분기 추가. default (사외) = saovae_stub; `=internal` 시 real KeycloakAdminClient (현 path). event listener 의 type assertion (`*httpapi.KeycloakAdminClient`) 정리는 다음 PR (real adapter 이전) 에서. (3) `backend-core/internal/domain/auth-session/integration/ports.go` — `KeycloakUserEvent` / `KeycloakAdminEvent` mirror struct → `type X = httpapi.X` alias. `*KeycloakAdminClient` 가 `KeycloakEventPort` 충족 위해 필수. v1.0 의 mirror struct 자체 제거는 별도 PR. (4) `backend-core/internal/domain/auth-session/view/{auth,handler}.go` — `BearerTokenVerifier` + `IdentityAdmin` + `OIDCLogoutClient` 3 interface deprecation comment. canonical = `integration/`. **Tier**: ports.go / view/ = 공용, sso-integrations/keycloak/ = 사외 (saovae_stub), main.go 변경 = 공용 (runtime injection branch 만, 사내 한정 패턴 미도입). **CI 7/7 PASS** (Backend Integration 1m15s, Backend Unit 1m8s, E2E Build 1m49s, E2E Playwright 1/3 4m0s, 2/3 4m2s, 3/3 5m30s, Detect Changed Paths 11s, Migration Prefix 7s, OpenAPI Lint 9s, Workflow Lint 13s, Frontend Unit skip). **commit `a00793bc`** (5 file, +238 -97). **PR #539 squash merge** → main HEAD `87e6c1f5` (branch delete). **신규 ID**: IMPL-30 (saovae_stub), IMPL-31 (main.go runtime injection), IMPL-32 (view/ deprecation + ports.go alias) — `docs/traceability/report.md` 갱신은 후속 PR (C-g). |
| 2026-06-10 | **v1.1 sprint -a follow-up PR1 (PR #540) 의 carry-over C-g + C-h 정공법 PR (docs only)** (sprint `docs/work_260610-traceability-impl-sso-keycloak`) — PR #540 (`feat/work_260610-v1-1-sprint-a-real-adapter`, main `58d163f`) 머지 후의 후속 정공법 PR. **문서만 변경 (코드 0줄)**, 5 row IMPL ID 신규 발급 (`conventions.md §1` kebab-case 정합 — 메모리 출발점의 `IMPL-30/31/32` 표기는 형식 위반 → 정정). (1) `docs/traceability/report.md` §2.4 IMPL 개요 paragraph 갱신 + `IMPL-sso-keycloak-01` (real adapter — `sso-integrations/keycloak/verifier.go` + `admin_client.go` 의 `KeycloakJWKSVerifier` + `KeycloakAdminClient` 3 port 동시 충족) + `IMPL-sso-keycloak-stub-01` (사외 build stub — `sso-integrations/keycloak/saovae_stub.go` 4 port + webhook handler) + `IMPL-sso-keycloak-metrics-01` (JWKS stale-while-error metric — `sso-integrations/keycloak/metrics.go`) + `IMPL-auth-session-port-01` (canonical port interface + view/ deprecation + struct 직접 정의) + `IMPL-audit-ops-event-mirror-01` (mirror 통합 + `KeycloakEventLister` interface 통폐합) 5 row 신규 sub-table. (2) §3.1 auth-session / audit-ops + §3.3 keycloak-idp 매트릭스 row 갱신 (5 row cross-ref + ADR-0030 link). (3) §4 ADR 인덱스 ADR-0030 row 신규 + §6 변경 이력 row 신규. (4) `docs/adr/0030-sso-integrations-and-auth-session-port.md` §5 timeline 1.1a + 1.1b status = **accepted/done** + C-h row 신규 + §9 변경 이력 row. **신규 ID**: 5건 (모두 IMPL, REQ/UC/ARCH/API/RM/UT/TC 신규 발급 0건 — sprint -a follow-up PR #540 코드 정합의 문서 cell fill). **Tier**: 본 PR 의 모든 변경 = **공용** (사내 한정 정보 미포함, 문서만). **잔여 검증** (PR 발행 직전): `bash scripts/check-tier-separation.sh` PASS + `bash scripts/check-openapi-yaml-lint.sh` PASS + `pytest ai-workflow/tests/check_docs.py` PASS. **Refs**: PR #538 + #539 + #540, [ADR-0030](../../docs/adr/0030-sso-integrations-and-auth-session-port.md). |
| 2026-06-10 | **v1.1 sprint -a follow-up PR1 PR #540 의 carry-over C-i 정공법 PR** (sprint `feat/work_260610-c-i-e2e-internal-job`) — e2e shard 1/2/3 (saovae_stub default) + 별도 e2e-internal job 1개 (`DEVHUB_BUILD_TIER=internal`) 의 CI matrix 1쌍 정합. 옵션 A (1쌍) 채택, 옵션 B (unit test only) / C (6 matrix) / D (matrix shard 1/2/3 × 2) 모두 거부. (1) `.github/workflows/ci.yml` 에 `e2e-internal` job 신규 (+202 lines, 23 step: PG 15 native + Keycloak container port 8181 + apply migrations + validate E2E-CI sync contract + **Start Backend DEVHUB_BUILD_TIER=internal — real Keycloak adapter** + Start Frontend + Wait + Run E2E Tests shard 1/1 + Upload Report + Upload Logs). (2) `scripts/ci-e2e-sync-check.sh` 에 DEVHUB_BUILD_TIER 의도적 미포함 rationale comment (+5 lines). e2e shard 1/2/3 의 env block, start command, test invocation 모두 변경 0. **신규 ID 없음** (CI workflow + script 만, ADR-0030 §2.3 runtime injection 결정의 code-level 적용). **Tier**: 본 PR 의 모든 변경 = **공용** (`.github/workflows/*` + `scripts/*` 모두 사내 한정 정보 미포함). **검증** (run on this branch): `bash scripts/check-tier-separation.sh` PASS (no changes between origin/main and HEAD) + `bash scripts/check-openapi-yaml-lint.sh` PASS + `bash scripts/check-migration-uniqueness.sh` PASS + `bash scripts/ci-e2e-sync-check.sh` PASS (E2E-CI sync contract check). **CI 4/4 PASS** (Detect Changed Paths 6s + Migration Prefix 7s + OpenAPI YAML Lint 7s + Workflow Lint actionlint 11s, backend/e2e/frontend skip — workflow 변경만). **commit `9c2a3f8`** (7 file, +515/-0). **PR #542 squash merge** → main HEAD `24674b8` (branch delete). **Refs**: PR #538 + #539 + #540 + #541 (C-g/C-h 정합), [ADR-0030 §2.3 runtime injection 결정](../adr/0030-sso-integrations-and-auth-session-port.md), [release_v1_roadmap.md §3.5 N-13](../planning/release_v1_roadmap.md). |
| 2026-06-10 | **v1.1 sprint -a follow-up PR1 PR #540 의 carry-over C-j 정공법 PR (docs only)** (sprint `docs/work_260610-c-j-build-tag-review`) — PR #540 의 carry-over C-j (P3, build tag 정책 재검토) 의 정공법 PR. **문서만 변경 (코드 0줄)**. (1) `docs/adr/0031-build-tag-policy-review.md` **신규 (12KB, 9 section)** — §0 메타 + §1 배경 (ADR-0030 §2.3 결정) + §2 **정량 측정** (런타임 injection vs build tag) + §3 후보 옵션 3건 (build tag / 런타임 injection / hybrid) + §4 **결정 (런타임 injection confirmed, supersede X)** + §5 재검토 trigger 5건 (현시점 0건) + §6 cross-tier + §7 risks + §8 supersession (X) + §9 변경 이력. 정량 데이터: `sso-integrations/keycloak/saovae_stub.go` 105 lines 3,831 bytes + `metrics.go` 70 lines + 8 file 2,335 lines ~70KB, stub binary overhead < 5KB (전체 backend-core < 50MB 대비 0.01%). (2) `docs/adr/0030-...` §2.3 row 에 `**2026-06-10 confirmed (ADR-0031 §4 재평가)**` reference 추가 + §9 row 추가. (3) `docs/traceability/report.md` §4 ADR-0031 row 신규 + §6 row 신규. (4) `ai-workflow/memory/docs/work_260610-c-j-build-tag-review/{state.json,session_handoff.md,work_backlog.md,backlog/2026-06-10.md,pr_body.md}` sprint memory 5종 신규. **신규 ID**: 0건 (ADR-0031 신규 1건, IMPL/REQ/UC/ARCH/API/RM/UT/TC 미발급). **Tier**: 본 PR 의 모든 변경 = **공용** (docs/adr/ + docs/traceability/ + ai-workflow/memory/ 모두 사내 한정 정보 미포함). **검증** (run on this branch): `bash scripts/check-tier-separation.sh` PASS (no changes between origin/main and HEAD) + `bash scripts/check-openapi-yaml-lint.sh` PASS + `bash scripts/check-migration-uniqueness.sh` PASS + `python3.13 ai-workflow/tests/check_docs.py` 의 본 PR 5 file 정합 (metadata 6 field + cross-link + 제목 헤더). **CI 4/4 PASS** (Detect Changed Paths 5s + Migration Prefix 5s + OpenAPI YAML Lint 12s + Workflow Lint actionlint 15s, backend/e2e/frontend skip — docs only PR). **commit `9b4ee1c`** (8 file, +511/-2). **PR #543 squash merge** → main HEAD `d3488ca` (branch delete). **Refs**: [ADR-0031 §4 결정](../adr/0031-build-tag-policy-review.md), PR #539 + #540 + #541 + #542, [release_v1_roadmap.md §3.5 N-13](../planning/release_v1_roadmap.md). |
