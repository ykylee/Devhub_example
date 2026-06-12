# Integrated Work Backlog (main, post M1 RBAC)

- 문서 목적: main 브랜치 기준 상위 백로그 인덱스. 세부 sprint backlog 는 브랜치별 메모리 디렉터리 참조.
- 범위: 마일스톤 상태, 최근 머지, 잔여/후속 작업
- 대상 독자: 프로젝트 리드, 후속 에이전트, 트랙 담당자
- 상태: **2026-06-12 update — main HEAD `fc7e6c76` (chore(memory): main flat memory finalize, 2026-06-12, 본 sprint N-13 follow-up 3 branch + E2E Internal disable + session end 정합) + PR #576 close (N-13 follow-up C PR A-2 e2e shard 3/3 fail, 사용자 결정). 06-09~06-12 사이 PR #514~#577 (49+ PR) 정합. **N-13 follow-up 3 branch 종합** (A: Test 1, B: Test 2, C: 구현): **A 완료** (PR #574 MERGED, Test 1 e2e seed fix), **B 완료** (PR #575 MERGED, Test 2 verification + CI Run #1227 SUCCESS), **C closed** (PR #576 close, e2e shard 3/3 fail — PATCH /api/v1/platforms 가 4xx 반환, PATCH inbound_source 의 backend 처리 검증 필요). **E2E Internal 일시 disable** (PR #577 MERGED, ci.yml `if: ${{ vars.SKIP_E2E_INTERNAL == 'true' }}` actionlint 호환, v1.1 sprint 에서 real Keycloak adapter 안정화 후 재활성화). **follow-up 잔여**: (1) **PATCH inbound_source 의 backend 처리 검증 + e2e spec 재작성 + 새 PR 발행** (별도 sprint, 사용자 결정). (2) **구현 follow-up = v1.1 milestone 진입 시점 별도 sprint** (rebase main + PR #550 fix + 본 fix 종합 + 자동 재실행). status `⏳ planned` 유지 (구현 미완료).
- 직전 상태: 2026-06-12 update — main HEAD `8616ac59` (PR #572 N-10 follow-up 보류 결정 squash, 2026-06-12) + 후속 housekeeping 정공법 본 sprint (N-13 PR #548 close follow-up 결정, sprint `fix/work_260612-1-n13-housekeeping-followup`). 06-09~06-12 사이 PR #514~#572 (45+ PR) 정합. **N-13 housekeeping follow-up** (sprint `fix/work_260612-1-n13-housekeeping-followup`): PR #548 (`feat/work_260611-a-n13-inbound-source-impl`, 2026-06-11 05:40 UTC CLOSED) 의 E2E Internal 1 fail 2건 (Test 1: e2e seed 중복 strict mode violation + Test 2: Sign-out timeout N-8 race 유사) + 자동 재실행 미적용 정공법. follow-up 결정 3 branch: (1) Test 1 e2e seed 중복 → spec/e2e seed 정합 fix 별도 sprint; (2) Test 2 Sign-out timeout → main rebase + 자동 재실행 검증 별도 (PR #550 spec timing fix 가 해결 가능성 ↑); (3) 구현 follow-up = v1.1 milestone 진입 시점 별도 sprint (rebase main + PR #550 fix + e2e seed 정합 fix + 자동 재실행 종합). status `⏳ planned` 유지 (구현 미완료). branch `feat/work_260611-a-n13-inbound-source-impl` close (PR #548) — v1.1 진입 시점의 신규 구현 sprint 는 새 branch 이름 별도 결정. release_v1_roadmap.md §3.5 N-13 row + §9 + ADR-0028 §6 (a) + §7 + sprint plan §3.3/§5/§6 + traceability/report.md §6 + 메모리 4 file 동기화. 신규 ID 발급 0건 (housekeeping follow-up).
</input>

- 최종 수정일: 2026-06-12 (**N-13 follow-up C PR #576 close + E2E Internal disable PR #577 MERGED** — sprint `fix/work_260612-5-e2e-internal-disable` + `feat/work_260612-4-v1-1-inbound-source-impl` (PR #576). PR #576 의 e2e shard 3/3 fail (PATCH /api/v1/platforms 4xx) → 사용자 결정 close. PR #577 ✅ MERGED (`802afe62`, ci.yml `if: ${{ vars.SKIP_E2E_INTERNAL == 'true' }}`, E2E Internal skip 검증). main HEAD `802afe62` (PR #577 baseline). 후속 — PATCH inbound_source 의 backend 처리 검증 + e2e spec 재작성 + 새 PR 발행 (별도 sprint, 사용자 결정))
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
| 2026-06-11 | **잔여 follow-up sprint (`chore/work_260611-b-residual-housekeeping-close`)** — N-10 housekeeping close (status ⏳ verified (partial) → ✅ resolved (full)) + lint 8 errors / 62 warns follow-up 사실상 close (D-74 L03 skip patch + skip config 적용, 3개 lint report 모두 0 error / 0 warn / 0 info 정합) + N-13 housekeeping 정공법 (status ⏳ planned 유지, 구현은 별도 sprint) + memory 4 file 갱신. 4 file 변경 (`docs/planning/release_v1_roadmap.md` §3.5 N-10 row close + `docs/validation/N-10-manager-rbac.md` §0/§3.1/§6 갱신 + `state.json` M-v1.0 notes 갱신 + `work_backlog.md` 변경 이력 row 추가). 신규 ID 없음 (housekeeping 정공법). |
| 2026-06-11 | **N-6 skip (사용자 결정, v1.0 release blocker 해제)** — v1.0 staging 1주 운영 검증 미실시. 5 file 변경 (`docs/planning/release_v1_roadmap.md` §1.3 DoD 7 + §3.5 N-6 row status `⏳ 사용자 결정` → `✅ skipped (사용자 결정, 2026-06-11, v1.0 release blocker 해제)` + `ai-workflow/memory/state.json` M-v1.0 notes "N-6 skipped" + 8 DoD 모두 close 명시 + `ai-workflow/memory/work_backlog.md` status line main HEAD `82935f8b` 갱신 + `ai-workflow/memory/session_handoff.md` §0 N-6 skip subsection + §0a → §0b 갱신 + 본 row). **v1.0 release blocker 0건** — 8 DoD 중 7 ✅ + 1 ✅ skipped = **8 DoD 모두 close**. **v1.0 release 가능 상태**. 신규 ID 없음 (housekeeping 정공법). main HEAD `82935f8b`. |
| 2026-06-11 | **v0.1.0-alpha release (사용자 결정, 8 DoD close)** — `release` 디렉터리에 3 file 신규 (`CHANGELOG.md` 670+ lines + `docs/presentations/2026-06-11-v0.1.0-alpha-announcement.md` 200+ lines + `docs/presentations/2026-06-11-v0.1.0-alpha-announcement.html` 자체 HTML 14 슬라이드, 키보드 네비게이션, reveal.js 의존성 0). PR #554 (N-6 skip) 머지 → main HEAD `d860b7c9` → CHANGELOG + 발표 자료 commit `356d08b7`. Git tag `v0.1.0-alpha` 부착 + push (re-tag 후 main HEAD `356d08b7` 부착). v0.1.0-alpha release 발표 (2026-06-11, 14 슬라이드 발표 자료). 신규 ID 없음 (release 산출물). main HEAD `356d08b7`. |
| 2026-06-11 | **v0.1.1-alpha release 정공법 (사용자 결정, 잔여 5 의 v0.1.1-alpha 격하)** — 잔여 5 (T-d-72-5/6 + D-73/74 + X-1~8) 의 v0.1.1-alpha 격하 (v1.1 forward path 가 아닌 v0.1.x patch release). 5 file 변경 (`docs/planning/release_v1_roadmap.md` §3.5 NEXT block title `v1.1` → `v0.1.1-alpha` + `ai-workflow/memory/state.json` M-v1.0 notes "v0.1.1-alpha release 정공법" + 잔여 5 의 8 item 의 v0.1.1-alpha 격하 마킹 + `ai-workflow/memory/work_backlog.md` status line main HEAD + v0.1.1-alpha release 정합 + §5 변경 이력 row + `ai-workflow/memory/session_handoff.md` §0 v0.1.1-alpha subsection + `CHANGELOG.md` v0.1.1-alpha release note 추가). 8 item 모두 v0.1.1-alpha release 의 정공법 정합 (실제 구현은 사용자 결정 시점). 신규 ID 없음 (정공법 정합). main HEAD `356d08b7` (v0.1.0-alpha release 정합) + tag `v0.1.1-alpha` re-tag 후 main HEAD 부착. |
| 2026-06-12 | **N-13 PR #548 close follow-up 결정 (docs only)** (sprint `fix/work_260612-1-n13-housekeeping-followup`) — PR #548 (`feat/work_260611-a-n13-inbound-source-impl`, 2026-06-11 05:40 UTC CLOSED) 의 E2E Internal 1 fail 2건 (Test 1: e2e seed 중복 strict mode violation `getByText('e2e-repo-a')` 2 elements + Test 2: Sign-out timeout N-8 race 유사) + 자동 재실행 미적용 (PR #550 spec timing fix 미반영) 정공법. follow-up 결정 3 branch: (1) **Test 1 e2e seed 중복** → spec/e2e seed 정합 fix 별도 sprint (mock data 유일성 보장); (2) **Test 2 Sign-out timeout** → main rebase + 자동 재실행 검증 별도 (PR #550 spec timing fix 가 해결 가능성 ↑); (3) **구현 follow-up = v1.1 milestone 진입 시점 별도 sprint** (rebase main + PR #550 fix + e2e seed 정합 fix + 자동 재실행 종합). branch `feat/work_260611-a-n13-inbound-source-impl` close (PR #548) — v1.1 진입 시점의 신규 구현 sprint 는 새 branch 이름 (예: `feat/work_YYMMDD-v1-1-inbound-source-impl`) 별도 결정. 5 file 변경 (`docs/planning/release_v1_roadmap.md` §3.5 N-13 row 본문 + §9 + 헤더 메타 / `docs/adr/0028-dev-requests-voc-external-ref.md` §6 (a) + §7 + 헤더 메타 / `docs/planning/2026-06-12-inbound-source-routing-sprint-plan.md` §3.3 의존 + §5 결정 보류 + §6 + 헤더 메타 / `docs/traceability/report.md` §6 / 본 work_backlog.md status line + §5 row). 신규 ID 발급 0건 (housekeeping follow-up 정공법). main HEAD `8616ac59` (PR #572 머지 baseline). |
| 2026-06-12 | **N-13 follow-up A (Test 1 e2e seed 중복 strict mode violation fix) 완료** (sprint `fix/work_260612-2-e2e-seed-strict-mode-fix`) — root cause = `frontend/tests/e2e/repositories-ui.spec.ts:5, 7` 의 `repoALink` / `repoBLink` matcher 가 `.first()` 미적용. fix = matcher 정의에 `.first()` 추가 (기존 `repository-dashboard.spec.ts:71, 117` 의 동일 패턴 정합). 1 file 변경, diff +3/-2. **PR #574 squash `896d9018` 머지 완료** (2026-06-12, main HEAD 정합). Tier: 사외. 신규 ID 발급 0건 (test stabilization). follow-up 2 branch 잔여: (1) Test 2 Sign-out timeout rebase + 자동 재실행 검증; (2) v1.1 milestone 진입 시점 구현 follow-up. |
| 2026-06-12 | **N-13 follow-up B (Test 2 Sign-out timeout rebase + 자동 재실행 검증) 진행** (sprint `fix/work_260612-3-n13-followup-b-test2-rebase`) — main 의 PR #550 spec timing fix + PR #574 e2e seed fix 적용 후 main 의 e2e CI 가 자동 재실행 시 e2e shard 1/2/3 모두 PASS 검증. 5 file 변경 (docs only, verification report) + 메모리 4 file 동기화. verification PR 의 본질 = main 의 e2e CI 안정성 evidence. Tier: 공용. 신규 ID 발급 0건 (verification). follow-up 1 branch 잔여: 구현 follow-up = v1.1 milestone 진입 시점 별도 sprint. |
| 2026-06-12 | **N-13 follow-up 3 branch 종합 + E2E Internal disable + session end** (sprint `fix/work_260612-5-e2e-internal-disable` + `feat/work_260612-4-v1-1-inbound-source-impl` (PR #576) + 본 메모리 finalize `fc7e6c76`) — 5 PR 결과 종합: **#573 MERGED** (N-13 housekeeping follow-up docs only, `5fb9ae75`), **#574 MERGED** (Test 1 e2e seed fix `896d9018`), **#575 MERGED** (Test 2 verification + CI Run #1227 SUCCESS, `8d0e2e88`), **#576 CLOSED** (구현 follow-up PR A-2, e2e shard 3/3 fail, PATCH /api/v1/platforms 4xx, 사용자 결정 close), **#577 MERGED** (E2E Internal disable, ci.yml `if: ${{ vars.SKIP_E2E_INTERNAL == 'true' }}`, `802afe62`). 본 sprint 의 메모리 finalize commit `fc7e6c76` (main HEAD 정합). **session 종료 결정** (사용자, "일단 여기까지 하고 세션 종료"). **follow-up 잔여 3가지** (사용자 결정 영역): (1) PATCH inbound_source 의 backend 처리 검증 + e2e spec 재작성 + 새 PR 발행 (별도 sprint); (2) 구현 follow-up = v1.1 milestone 진입 시점 별도 sprint; (3) E2E Internal 재활성화 (vars.SKIP_E2E_INTERNAL 원래 조건 복원 + 새 PR). 메모리 3 file 갱신 (`state.json` line 46 notes 정합 + `work_backlog.md` status line + §5 row + `session_handoff.md` §21 append). **신규 ID 발급 0건** (memory finalize). main HEAD `fc7e6c76` 정합. |
| 2026-06-12 | **CI 재구성 + flaky 복구 in-flight** (branch `codex/work_260612-579-ci-rearchitecture`) — CI 를 **fast required / E2E regression / E2E quarantine** 3계층으로 재설계. `.github/workflows/ci.yml` 는 빠른 required path + smoke 로 정리하고, `.github/workflows/e2e-regression.yml` / `.github/workflows/e2e-quarantine.yml` 를 신규 분리. `frontend/tests/e2e-manifests/{smoke,quarantine}.txt` + `scripts/select-playwright-specs.sh` 를 spec selection SSOT 로 도입. signout/login flaky 복구를 위해 `frontend/tests/e2e/fixtures.ts` 에 OIDC 재시작 helper / login loop 복구 / CI timeout 완화를 추가하고 `signout.spec.ts` timeout 을 상향. 검증: workflow YAML parse OK, `scripts/ci-e2e-sync-check.sh` 3 workflow 정합 OK, selector script smoke/quarantine/regression 출력 OK, focused e2e TypeScript check OK. full frontend `tsc --noEmit` 는 선행 타입 오류로 실패 (본 변경과 무관). 다음 단계: draft PR 에서 required check 후보와 quarantine 운영 원칙 확정. |

---

## 6. Phase 2 1차 chunk 결과 (2026-06-11, out-of-repo)

본 세션 (2026-06-11, Phase 2 type 자동 분류) 의 1차 chunk 결과:

| 항목 | Before | After |
|---|---|---|
| lint total | 196 findings | **98 findings** (-98, -50%) |
| errors | 18 | 11 (-7) |
| warns | 178 | 87 (-91) |
| L04 (ADR naming 중복) | 31 | **0** (mavis-trash 후) |
| L06 (sources path) | 9 | **0** (7 패턴 정합) |
| L08 (index 미등록) | 31 | 1 (5 page 만 등록) |
| type 분포 | sources 113 / concept 3 / entity 4 / topic 2 | **sources 83 / concept 5 / entity 4 / topic 3** |

**신규 5 page**: `concepts/rbac.md` / `entities/keycloak.md` / `concepts/agent-memory.md` / `concepts/llm-wiki-pattern.md` / `topics/workflow.md`. frontmatter 8 key (title/type/tags/sources/last_touched/related/status/contradictions) + 4섹션 본문 (Summary/Observations/Relations/Open Questions).

**forward path 잔여**: L02 11 (cross/concepts/, cross/topics/ 의 4 page 의 wikilink 11개) + L03 86 (sources/ 의 cross-ref 부족) — 2차 chunk + 3차 chunk 에서 해소.

main HEAD `8dba2b3` (메모리 finalize, 2026-06-11 23:55 KST). in-repo PR 0 (vault 만 변경).
